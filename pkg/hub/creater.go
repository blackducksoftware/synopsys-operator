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

package hub

import (
	"fmt"
	"math"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/hub/v2"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/hub/containers"
	hubutils "github.com/blackducksoftware/synopsys-operator/pkg/hub/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	Config           *protoform.Config
	KubeConfig       *rest.Config
	KubeClient       *kubernetes.Clientset
	HubClient        *hubclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset,
	osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, HubClient: hubClient, osSecurityClient: osSecurityClient, routeClient: routeClient}
}

// DeleteHub will delete the Black Duck Hub
func (hc *Creater) DeleteHub(namespace string) error {

	logrus.Infof("Deleting hub: %s", namespace)

	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("unable to find the namespace %+v because %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(hc.KubeClient, namespace)
		if err != nil {
			log.Errorf("unable to delete the namespace %+v because %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(hc.KubeClient, namespace)
			log.Infof("namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("deleted the namespace %+v", namespace)
				break
			}
		}
	}

	// Delete a persistent volume
	err = util.DeletePersistentVolume(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("unable to delete the pv for %+v", namespace)
	}

	// Delete a Cluster Role Binding
	err = util.DeleteClusterRoleBinding(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("unable to delete the cluster role binding for %+v", namespace)
	}
	return nil
}

// CreateHub will create the Black Duck Hub
func (hc *Creater) CreateHub(createHub *v2.HubSpec) (string, string, bool, error) {
	log.Debugf("create Hub details for %s: %+v", createHub.Namespace, createHub)

	// Create a horizon deployer for each hub
	deployer, err := horizon.NewDeployer(hc.KubeConfig)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to create the horizon deployer because %+v", err)
	}

	// Get Containers Flavor
	hubContainerFlavor := containers.GetContainersFlavor(createHub.Size)
	log.Debugf("Hub Container Flavor: %+v", hubContainerFlavor)

	if hubContainerFlavor == nil {
		return "", "", true, fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) X-Large, Actual: %s", createHub.Size)
	}

	// All ConfigMap environment variables
	allConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config-granular"},
	}

	var adminPassword, userPassword, postgresPassword string

	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		adminPassword, userPassword, postgresPassword, err = hubutils.GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)
		if err == nil {
			break
		} else {
			log.Infof("[%s] wasn't able to init database, sleeping 5 seconds.  try = %v", createHub.Namespace, dbInitTry)
			time.Sleep(5 * time.Second)
		}
	}

	log.Debugf("before init: %+v", &createHub)
	// Create the config-maps, secrets and postgres container
	err = hc.init(deployer, createHub, hubContainerFlavor, allConfigEnv, adminPassword, userPassword)
	if err != nil {
		return "", "", true, err
	}
	// Deploy config-maps, secrets and postgres container
	err = deployer.Run()
	if err != nil {
		log.Errorf("init deployments failed for %s because %+v", createHub.Namespace, err)
	}
	// time.Sleep(20 * time.Second)

	if createHub.ExternalPostgres == (v2.PostgresExternalDBConfig{}) {
		// Validate postgres pod is cloned/backed up
		err = util.WaitForServiceEndpointReady(hc.KubeClient, createHub.Namespace, "postgres")
		if err != nil {
			return "", "", true, err
		}

		if len(createHub.DbPrototype) == 0 {
			err := InitDatabase(createHub, adminPassword, userPassword, postgresPassword)
			if err != nil {
				log.Errorf("%v: error: %+v", createHub.Namespace, err)
				return "", "", true, fmt.Errorf("%v: error: %+v", createHub.Namespace, err)
			}
		} else {
			_, fromPw, err := hubutils.GetHubDBPassword(hc.KubeClient, createHub.DbPrototype)
			if err != nil {
				return "", "", true, err
			}
			err = hubutils.CloneJob(hc.KubeClient, hc.Config.Namespace, createHub.DbPrototype, createHub.Namespace, fromPw)
			if err != nil {
				return "", "", true, err
			}
		}
	}

	err = hc.addAnyUIDToServiceAccount(createHub)
	if err != nil {
		log.Error(err)
	}

	// Create all hub deployments
	deployer, _ = horizon.NewDeployer(hc.KubeConfig)
	hc.AddToDeployer(deployer, createHub, hubContainerFlavor, allConfigEnv)
	log.Debugf("%+v", deployer)
	// Deploy all hub containers
	err = deployer.Run()
	if err != nil {
		log.Errorf("post deployments failed for %s because %+v", createHub.Namespace, err)
		return "", "", true, fmt.Errorf("unable to deploy the hub in %s because %+v", createHub.Namespace, err)
	}
	time.Sleep(10 * time.Second)

	// Validate all pods are in running state
	err = util.ValidatePodsAreRunningInNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		return "", "", true, err
	}

	// Retrieve the PVC volume name
	pvcVolumeName := ""
	if createHub.PersistentStorage && createHub.ExternalPostgres == (v2.PostgresExternalDBConfig{}) {
		pvcVolumeName, err = hc.getPVCVolumeName(createHub.Namespace, "blackduck-postgres")
		if err != nil {
			return "", "", false, err
		}
	}

	// OpenShift routes
	ipAddress := ""
	if hc.routeClient != nil {
		route, err := util.CreateOpenShiftRoutes(hc.routeClient, createHub.Namespace, createHub.Namespace, "Service", "webserver")
		if err != nil {
			return "", pvcVolumeName, false, err
		}
		log.Debugf("openshift route host: %s", route.Spec.Host)
		ipAddress = route.Spec.Host
	}

	if strings.EqualFold(ipAddress, "") {
		ipAddress, err = hc.getLoadBalancerIPAddress(createHub.Namespace, "webserver-lb")
		if err != nil {
			ipAddress, err = hc.getNodePortIPAddress(createHub.Namespace, "webserver-np")
			if err != nil {
				return "", pvcVolumeName, false, err
			}
		}
	}
	log.Infof("hub Ip address: %s", ipAddress)

	return ipAddress, pvcVolumeName, false, nil
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
