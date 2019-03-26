/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package blackduck

import (
	"fmt"
	"math"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"

	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database"
	postgres2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	hubutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1/containers"
)

// Creater will store the configuration to create the Blackduck
type Creater struct {
	Config                  *protoform.Config
	KubeConfig              *rest.Config
	KubeClient              *kubernetes.Clientset
	BlackduckClient         *blackduckclientset.Clientset
	osSecurityClient        *securityclient.SecurityV1Client
	routeClient             *routeclient.RouteV1Client
	isBinaryAnalysisEnabled bool
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *blackduckclientset.Clientset,
	osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client, isBinaryAnalysisEnabled bool) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, BlackduckClient: hubClient, osSecurityClient: osSecurityClient,
		routeClient: routeClient, isBinaryAnalysisEnabled: isBinaryAnalysisEnabled}
}

// Ensure will ensure the instance is correctly deployed
func (hc *Creater) Ensure(blackduck *v1.Blackduck) error {
	// Create namespace if it doesn't exist
	_, err := util.GetNamespace(hc.KubeClient, blackduck.Spec.Namespace)
	if err != nil {
		_, err = util.CreateNamespace(hc.KubeClient, blackduck.Spec.Namespace)
		if err != nil {
			return err
		}
	}

	// Get components
	cpList, err := hc.GetComponents(blackduck)
	if err != nil {
		return err
	}

	// Create PVC if they don't exist
	if blackduck.Spec.PersistentStorage {
		var missingPVCs []*components.PersistentVolumeClaim
		for _, v := range hc.GetPVC(blackduck) {
			_, err := hc.KubeClient.CoreV1().PersistentVolumeClaims(blackduck.Spec.Namespace).Get(v.GetName(), metav1.GetOptions{})
			if err != nil {
				missingPVCs = append(missingPVCs, v)
			}
		}

		if len(missingPVCs) > 0 {
			deploy, err := horizon.NewDeployer(hc.KubeConfig)
			if err != nil {
				return err
			}

			for _, v := range missingPVCs {
				deploy.AddPVC(v)
			}

			err = deploy.Run()
			if err != nil {
				return err
			}
		}
	}

	// Check postgres and initialize if needed.
	if blackduck.Spec.ExternalPostgres == nil {
		cpPostgresList, err := hc.getPostgresComponents(blackduck)
		if err != nil {
			return err
		}

		// CM
		commonConfig := crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace, &api.ComponentList{ConfigMaps: cpList.ConfigMaps}, "app=blackduck,component=configmap")
		errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("unable to update components due to %+v", errors)
		}

		// Secret
		commonConfig = crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace, &api.ComponentList{Secrets: cpList.Secrets}, "app=blackduck,component=secret")
		errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("unable to update components due to %+v", errors)
		}

		// Postgres
		commonConfig = crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace, cpPostgresList, "app=blackduck,component=postgres")
		errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("unable to update components due to %+v", errors)
		}

		// TODO return whether we re-initialized or not
		err = hc.initPostgres(&blackduck.Spec)
		if err != nil {
			return err
		}
	}

	// Ensure
	commonConfig := crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace, cpList, "app=blackduck,component!=postgres")
	errors := commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update components due to %+v", errors)
	}
	return nil
}

// Versions returns the supported version
func (hc *Creater) Versions() []string {
	return containers.GetVersions()
}

// getContainersFlavor will get the Containers flavor
func (hc *Creater) getContainersFlavor(bd *v1.Blackduck) (*containers.ContainerFlavor, error) {
	// Get Containers Flavor
	hubContainerFlavor := containers.GetContainersFlavor(bd.Spec.Size)

	if hubContainerFlavor == nil {
		return nil, fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) X-Large, Actual: %s", bd.Spec.Size)
	}
	return hubContainerFlavor, nil
}

func (hc *Creater) initPostgres(bdspec *v1.BlackduckSpec) error {
	var adminPassword, userPassword, postgresPassword string
	var err error

	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		adminPassword, userPassword, postgresPassword, err = hubutils.GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)
		if err == nil {
			break
		} else {
			log.Infof("[%s] wasn't able to init database, sleeping 5 seconds.  try = %v", bdspec.Namespace, dbInitTry)
			time.Sleep(5 * time.Second)
		}
	}

	// Validate postgres pod is cloned/backed up
	err = util.WaitForServiceEndpointReady(hc.KubeClient, bdspec.Namespace, "postgres")
	if err != nil {
		return err
	}

	// Validate the postgres container is running
	err = util.ValidatePodsAreRunningInNamespace(hc.KubeClient, bdspec.Namespace, hc.Config.PodWaitTimeoutSeconds)
	if err != nil {
		return err
	}

	// Check if initialization is required.
	db, err := database.NewDatabase(fmt.Sprintf("postgres.%s.svc.cluster.local", bdspec.Namespace), "postgres", "postgres", postgresPassword, "postgres")
	if err != nil {
		return err
	}

	result, err := db.Connection.Exec("SELECT datname FROM pg_catalog.pg_database WHERE datname='bds_hub';")
	if err != nil {
		return err
	}
	nbRow, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// We initialize the DB if the bds_hub database doesn't exist
	if nbRow == 0 {
		log.Infof("postres instance %s requires to be re-initialized", bdspec.Namespace)
		if len(bdspec.DbPrototype) == 0 {
			err := InitDatabase(bdspec, adminPassword, userPassword, postgresPassword)
			if err != nil {
				log.Errorf("%v: error: %+v", bdspec.Namespace, err)
				return fmt.Errorf("%v: error: %+v", bdspec.Namespace, err)
			}
		} else {
			_, fromPw, err := hubutils.GetHubDBPassword(hc.KubeClient, bdspec.DbPrototype)
			if err != nil {
				return err
			}
			err = hubutils.CloneJob(hc.KubeClient, hc.Config.Namespace, bdspec.DbPrototype, bdspec.Namespace, fromPw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (hc *Creater) getPostgresComponents(bd *v1.Blackduck) (*api.ComponentList, error) {
	componentList := &api.ComponentList{}

	// Get Containers Flavor
	hubContainerFlavor, err := hc.getContainersFlavor(bd)
	if err != nil {
		return nil, err
	}

	containerCreater := containers.NewCreater(hc.Config, &bd.Spec, hubContainerFlavor)
	postgresImage := containerCreater.GetFullContainerNameFromImageRegistryConf("postgres")
	if len(postgresImage) == 0 {
		postgresImage = "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1"
	}
	var pvcName string
	if bd.Spec.PersistentStorage {
		pvcName = "blackduck-postgres"
	}
	postgres := postgres2.Postgres{
		Namespace:              bd.Spec.Namespace,
		PVCName:                pvcName,
		Port:                   containers.PostgresPort,
		Image:                  postgresImage,
		MinCPU:                 hubContainerFlavor.PostgresCPULimit,
		MaxCPU:                 "",
		MinMemory:              hubContainerFlavor.PostgresMemoryLimit,
		MaxMemory:              "",
		Database:               "blackduck",
		User:                   "blackduck",
		PasswordSecretName:     "db-creds",
		UserPasswordSecretKey:  "HUB_POSTGRES_USER_PASSWORD_FILE",
		AdminPasswordSecretKey: "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
		EnvConfigMapRefs:       []string{"hub-db-config"},
		Labels:                 containerCreater.GetVersionLabel("postgres"),
	}

	componentList.ReplicationControllers = append(componentList.ReplicationControllers, postgres.GetPostgresReplicationController())
	componentList.Services = append(componentList.Services, postgres.GetPostgresService())

	return componentList, nil
}

func (hc *Creater) getPVCVolumeName(namespace string, name string) (string, error) {
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		pvc, err := util.GetPVC(hc.KubeClient, namespace, name)
		if err != nil {
			return "", fmt.Errorf("unable to get pvc in %s namespace because %s", namespace, err.Error())
		}

		log.Debugf("pvc: %v", pvc)

		if strings.EqualFold(pvc.Spec.VolumeName, "") {
			continue
		} else {
			return pvc.Spec.VolumeName, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get pvc %s in %s namespace", namespace, namespace)
}

func (hc *Creater) getLoadBalancerIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		service, err := util.GetService(hc.KubeClient, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace because %s", serviceName, namespace, err.Error())
		}

		log.Debugf("[%s] service: %v", serviceName, service.Status.LoadBalancer.Ingress)

		if len(service.Status.LoadBalancer.Ingress) > 0 {
			ipAddress := service.Status.LoadBalancer.Ingress[0].IP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}

func (hc *Creater) getNodePortIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		service, err := util.GetService(hc.KubeClient, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace because %s", serviceName, namespace, err.Error())
		}

		log.Debugf("[%s] service: %v", serviceName, service.Spec.ClusterIP)

		if !strings.EqualFold(service.Spec.ClusterIP, "") {
			ipAddress := service.Spec.ClusterIP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}
