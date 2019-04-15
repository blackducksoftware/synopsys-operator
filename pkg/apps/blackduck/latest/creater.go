/*
Copyright (C) 2019 Synopsys, Inc.

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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	bdutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the Blackduck
type Creater struct {
	Config           *protoform.Config
	KubeConfig       *rest.Config
	KubeClient       *kubernetes.Clientset
	BlackduckClient  *blackduckclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *blackduckclientset.Clientset,
	osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, BlackduckClient: hubClient, osSecurityClient: osSecurityClient,
		routeClient: routeClient}
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (hc *Creater) Ensure(blackduck *blackduckapi.Blackduck) error {
	newBlackuck := blackduck.DeepCopy()

	pvcs := hc.GetPVC(blackduck)

	commonConfig := crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace,
		&api.ComponentList{PersistentVolumeClaims: pvcs}, "app=blackduck,component=pvc")
	errors := commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update postgres components due to %+v", errors)
	}

	// Get postgres components
	cpPostgresList, err := hc.getPostgresComponents(blackduck)
	if err != nil {
		return err
	}

	// install postgres
	commonConfig = crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace,
		cpPostgresList, "app=blackduck,component=postgres")
	errors = commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update postgres components due to %+v", errors)
	}
	// log.Debugf("created/updated postgres component for %s", blackduck.Spec.Namespace)

	// Check postgres and initialize if needed.
	if blackduck.Spec.ExternalPostgres == nil {
		// TODO return whether we re-initialized or not
		err = hc.initPostgres(&blackduck.Spec)
		if err != nil {
			return err
		}
	}

	// Get non postgres components
	cpList, err := hc.getComponents(blackduck)
	if err != nil {
		return err
	}

	// deploy non postgres and uploadcache component
	commonConfig = crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace,
		cpList, "app=blackduck,component notin (postgres,uploadcache)")
	errors = commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update non postgres and uploadcache components due to %+v", errors)
	}

	// log.Debugf("created/updated non postgres and upload cache component for %s", blackduck.Spec.Namespace)

	// deploy upload cache component
	commonConfig = crdupdater.NewCRUDComponents(hc.KubeConfig, hc.KubeClient, hc.Config.DryRun, blackduck.Spec.Namespace,
		cpList, "app=blackduck,component=uploadcache")
	errors = commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update upload cache components due to %+v", errors)
	}
	// log.Debugf("created/updated upload cache component for %s", blackduck.Spec.Namespace)

	if strings.ToUpper(blackduck.Spec.ExposeService) == "NODEPORT" {
		newBlackuck.Status.IP, err = bdutils.GetNodePortIPAddress(hc.KubeClient, blackduck.Spec.Namespace, "webserver-exposed")
	} else if strings.ToUpper(blackduck.Spec.ExposeService) == "LOADBALANCER" {
		newBlackuck.Status.IP, err = bdutils.GetLoadBalancerIPAddress(hc.KubeClient, blackduck.Spec.Namespace, "webserver-exposed")
	}

	// Create Route on Openshift
	if strings.ToUpper(blackduck.Spec.ExposeService) == "OPENSHIFT" && hc.routeClient != nil {
		route, err := util.GetOpenShiftRoutes(hc.routeClient, blackduck.Spec.Namespace, blackduck.Spec.Namespace)
		if err != nil {
			route, err = util.CreateOpenShiftRoutes(hc.routeClient, blackduck.Spec.Namespace, blackduck.Spec.Namespace, "Service", "webserver", routev1.TLSTerminationPassthrough)
			if err != nil {
				log.Errorf("unable to create the openshift route due to %+v", err)
			}
		}
		if route != nil {
			newBlackuck.Status.IP = route.Spec.Host
		}
	}

	if err := util.ValidatePodsAreRunningInNamespace(hc.KubeClient, blackduck.Spec.Namespace, 600); err != nil {
		return err
	}

	// TODO wait for webserver to be up before we register
	if len(blackduck.Spec.LicenseKey) > 0 {
		if err := hc.registerIfNeeded(blackduck); err != nil {
			log.Infof("couldn't register blackduck %s: %v", blackduck.Name, err)
		}
	}

	if blackduck.Spec.PersistentStorage {
		pvcVolumeNames := map[string]string{}
		for _, v := range blackduck.Spec.PVC {
			pvName, err := hc.getPVCVolumeName(blackduck.Spec.Namespace, v.Name)
			if err != nil {
				continue
			}
			pvcVolumeNames[v.Name] = pvName
		}
		newBlackuck.Status.PVCVolumeName = pvcVolumeNames
	}

	if !reflect.DeepEqual(blackduck, newBlackuck) {
		if _, err := hc.BlackduckClient.SynopsysV1().Blackducks(hc.Config.Namespace).Update(newBlackuck); err != nil {
			return err
		}
	}

	return nil
}

// Versions will return the versions supported
func (hc *Creater) Versions() []string {
	return containers.GetVersions()
}

// getContainersFlavor will get the Containers flavor
func (hc *Creater) getContainersFlavor(bd *blackduckapi.Blackduck) (*containers.ContainerFlavor, error) {
	// Get Containers Flavor
	hubContainerFlavor := containers.GetContainersFlavor(bd.Spec.Size)

	if hubContainerFlavor == nil {
		return nil, fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) X-Large, Actual: %s", bd.Spec.Size)
	}
	return hubContainerFlavor, nil
}

func (hc *Creater) initPostgres(bdspec *blackduckapi.BlackduckSpec) error {
	var adminPassword, userPassword, postgresPassword string
	var err error

	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		adminPassword, userPassword, postgresPassword, err = bdutils.GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)
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
	defer db.Connection.Close()

	// Wait for the DB to be up
	if !db.WaitForDatabase(10) {
		return fmt.Errorf("database %s is not accessible", bdspec.Namespace)
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
			_, fromPw, err := bdutils.GetHubDBPassword(hc.KubeClient, bdspec.DbPrototype)
			if err != nil {
				return err
			}
			err = bdutils.CloneJob(hc.KubeClient, hc.Config.Namespace, bdspec.DbPrototype, bdspec.Namespace, fromPw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (hc *Creater) getPVCVolumeName(namespace string, name string) (string, error) {
	pvc, err := util.GetPVC(hc.KubeClient, namespace, name)
	if err != nil {
		return "", fmt.Errorf("unable to get pvc in %s namespace because %s", namespace, err.Error())
	}

	return pvc.Spec.VolumeName, nil
}

func (hc *Creater) registerIfNeeded(bd *blackduckapi.Blackduck) error {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(fmt.Sprintf("https://webserver.%s.svc:443/api/v1/registrations?summary=true", bd.Spec.Namespace))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var objmap map[string]*json.RawMessage

	err = dec.Decode(&objmap)
	if err != nil {
		return err
	}

	// Check whether the registration is valid
	if val, ok := objmap["valid"]; ok {
		var r bool
		err := json.Unmarshal(*val, &r)
		if err != nil {
			return err
		}

		// We register if the registration is invalid
		if !r {
			if err := hc.autoRegisterHub(&bd.Spec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (hc *Creater) autoRegisterHub(bdspec *blackduckapi.BlackduckSpec) error {
	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod, err := util.FilterPodByNamePrefixInNamespace(hc.KubeClient, bdspec.Namespace, "registration")
	if err != nil {
		return err
	}

	registrationKey := bdspec.LicenseKey

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for i := 0; i < 20; i++ {
			registrationPod, err := util.GetPod(hc.KubeClient, bdspec.Namespace, registrationPod.Name)
			if err != nil {
				return err
			}

			// Create the exec into kubernetes pod request
			req := util.CreateExecContainerRequest(hc.KubeClient, registrationPod, "/bin/bash")
			// Exec into the kubernetes pod and execute the commands
			_, err = util.ExecContainer(hc.KubeConfig, req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate" -k --cert /opt/blackduck/hub/hub-registration/security/blackduck_system.crt --key /opt/blackduck/hub/hub-registration/security/blackduck_system.key`, registrationKey)})

			if err == nil {
				log.Infof("blackduck %s has been registered", bdspec.Namespace)
				return nil
			}
			time.Sleep(10 * time.Second)
		}
	}
	return fmt.Errorf("unable to register the blackduck %s", bdspec.Namespace)
}

func (hc *Creater) isBinaryAnalysisEnabled(bdspec *blackduckapi.BlackduckSpec) bool {
	for _, value := range bdspec.Environs {
		if strings.Contains(value, "USE_BINARY_UPLOADS") {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				mapValue := strings.Trim(values[1], " ")
				if strings.EqualFold(mapValue, "1") {
					return true
				}
			}
			return false
		}
	}
	return false
}
