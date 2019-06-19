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
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	bdutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the Blackduck
type Creater struct {
	config          *protoform.Config
	kubeConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	blackduckClient *blackduckclientset.Clientset
	routeClient     *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *blackduckclientset.Clientset,
	routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, blackduckClient: hubClient, routeClient: routeClient}
}

// Ensure will ensure the instance is correctly deployed
func (hc *Creater) Ensure(blackduck *blackduckapi.Blackduck) error {
	newBlackuck := blackduck.DeepCopy()

	pvcs := hc.GetPVC(blackduck)

	if strings.EqualFold(blackduck.Spec.DesiredState, "STOP") {
		commonConfig := crdupdater.NewCRUDComponents(hc.kubeConfig, hc.kubeClient, hc.config.DryRun, false, blackduck.Spec.Namespace,
			&api.ComponentList{PersistentVolumeClaims: pvcs}, fmt.Sprintf("app=%s,name=%s", util.BlackDuckName, blackduck.Name), false)
		_, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("stop blackduck: %+v", errors)
		}
	} else {
		commonConfig := crdupdater.NewCRUDComponents(hc.kubeConfig, hc.kubeClient, hc.config.DryRun, false, blackduck.Spec.Namespace,
			&api.ComponentList{PersistentVolumeClaims: pvcs}, fmt.Sprintf("app=%s,name=%s,component=pvc", util.BlackDuckName, blackduck.Name), false)
		isPatched, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update pvc: %+v", errors)
		}

		// Get postgres components
		cpPostgresList, err := hc.getPostgresComponents(blackduck)
		if err != nil {
			return err
		}

		// install postgres
		commonConfig = crdupdater.NewCRUDComponents(hc.kubeConfig, hc.kubeClient, hc.config.DryRun, isPatched, blackduck.Spec.Namespace,
			cpPostgresList, fmt.Sprintf("app=%s,name=%s,component=postgres", util.BlackDuckName, blackduck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update postgres component: %+v", errors)
		}
		// log.Debugf("created/updated postgres component for %s", blackduck.Spec.Namespace)

		// Check postgres and initialize if needed.
		if blackduck.Spec.ExternalPostgres == nil {
			// TODO return whether we re-initialized or not
			err = hc.initPostgres(blackduck.Name, &blackduck.Spec)
			if err != nil {
				return err
			}
		}

		// Get non postgres components
		cpList, err := hc.GetComponents(blackduck)
		if err != nil {
			return err
		}

		// install cfssl
		commonConfig = crdupdater.NewCRUDComponents(hc.kubeConfig, hc.kubeClient, hc.config.DryRun, isPatched, blackduck.Spec.Namespace,
			cpList, fmt.Sprintf("app=%s,name=%s,component in (configmap,serviceAccount,cfssl)", util.BlackDuckName, blackduck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update cfssl component: %+v", errors)
		}

		if err := util.ValidatePodsAreRunningInNamespace(hc.kubeClient, blackduck.Spec.Namespace, hc.config.PodWaitTimeoutSeconds); err != nil {
			return err
		}

		// deploy non postgres and uploadcache component
		commonConfig = crdupdater.NewCRUDComponents(hc.kubeConfig, hc.kubeClient, hc.config.DryRun, isPatched, blackduck.Spec.Namespace,
			cpList, fmt.Sprintf("app=%s,name=%s,component notin (postgres,configmap,serviceAccount,cfssl)", util.BlackDuckName, blackduck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update non postgres and cfssl component: %+v", errors)
		}
		// log.Debugf("created/updated non postgres and upload cache component for %s", blackduck.Spec.Namespace)

		if strings.ToUpper(blackduck.Spec.ExposeService) == "NODEPORT" {
			newBlackuck.Status.IP, err = bdutils.GetNodePortIPAddress(hc.kubeClient, blackduck.Spec.Namespace, util.GetResourceName(blackduck.Name, util.BlackDuckName, "webserver-exposed"))
		} else if strings.ToUpper(blackduck.Spec.ExposeService) == "LOADBALANCER" {
			newBlackuck.Status.IP, err = bdutils.GetLoadBalancerIPAddress(hc.kubeClient, blackduck.Spec.Namespace, util.GetResourceName(blackduck.Name, util.BlackDuckName, "webserver-exposed"))
		}

		// Create Route on Openshift
		if strings.ToUpper(blackduck.Spec.ExposeService) == util.OPENSHIFT && hc.routeClient != nil {
			route, _ := util.GetRoute(hc.routeClient, blackduck.Spec.Namespace, util.GetResourceName(blackduck.Name, util.BlackDuckName, ""))
			if route != nil {
				newBlackuck.Status.IP = route.Spec.Host
			}
		}

		if err := util.ValidatePodsAreRunningInNamespace(hc.kubeClient, blackduck.Spec.Namespace, 600); err != nil {
			return err
		}

		// TODO wait for webserver to be up before we register
		if len(blackduck.Spec.LicenseKey) > 0 {
			if err := hc.registerIfNeeded(blackduck); err != nil {
				log.Infof("couldn't register blackduck %s: %v", blackduck.Name, err)
			}
		}
	}

	// Commented to verify where it is used
	// if blackduck.Spec.PersistentStorage {
	// 	pvcVolumeNames := map[string]string{}
	// 	pvcList, err := hc.kubeClient.CoreV1().PersistentVolumeClaims(blackduck.Spec.Namespace).List(metav1.ListOptions{
	// 		LabelSelector: fmt.Sprintf("app=%s,name=%s,component=pvc", util.BlackDuckName, blackduck.Name),
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for _, v := range pvcList.Items {
	// 		pvName, err := hc.getPVCVolumeName(blackduck.Spec.Namespace, v.Name)
	// 		if err != nil {
	// 			continue
	// 		}
	// 		pvcVolumeNames[v.Name] = pvName
	// 	}
	// 	newBlackuck.Status.PVCVolumeName = pvcVolumeNames
	// }

	if !reflect.DeepEqual(blackduck.Status, newBlackuck.Status) {
		bd, err := util.GetHub(hc.blackduckClient, blackduck.Spec.Namespace, blackduck.Name)
		if err != nil {
			return err
		}
		bd.Status = newBlackuck.Status
		if _, err := hc.blackduckClient.SynopsysV1().Blackducks(blackduck.Spec.Namespace).Update(bd); err != nil {
			return err
		}
	}

	return nil
}

// Versions returns the supported version
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

func (hc *Creater) initPostgres(name string, bdspec *blackduckapi.BlackduckSpec) error {
	adminPassword, err := util.Base64Decode(bdspec.AdminPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode adminPassword due to: %+v", bdspec.Namespace, err)
	}
	userPassword, err := util.Base64Decode(bdspec.UserPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode userPassword due to: %+v", bdspec.Namespace, err)
	}
	postgresPassword, err := util.Base64Decode(bdspec.PostgresPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode postgresPassword due to: %+v", bdspec.Namespace, err)
	}

	ready, err := util.WaitUntilPodsAreReady(hc.kubeClient, bdspec.Namespace, fmt.Sprintf("app=%s,name=%s,component=postgres", util.BlackDuckName, name), hc.config.PodWaitTimeoutSeconds)
	if err != nil {
		return err
	}

	if !ready {
		return errors.New("the postgres pod is not yet ready")
	}

	// Check if initialization is required.
	db, err := database.NewDatabase(fmt.Sprintf("%s.%s.svc.cluster.local", util.GetResourceName(name, util.BlackDuckName, "postgres"), bdspec.Namespace), "postgres", "postgres", postgresPassword, "postgres")
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
			err := InitDatabase(name, bdspec, hc.config.IsClusterScoped, adminPassword, userPassword, postgresPassword)
			if err != nil {
				log.Errorf("%v: error: %+v", bdspec.Namespace, err)
				return fmt.Errorf("%v: error: %+v", bdspec.Namespace, err)
			}
		} else {
			fromNamespaces, err := util.ListNamespaces(hc.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.BlackDuckName, bdspec.DbPrototype))
			if len(fromNamespaces.Items) == 0 {
				return fmt.Errorf("unable to find the %s Black Duck instance", bdspec.DbPrototype)
			}
			fromNamespace := fromNamespaces.Items[0].Name
			_, fromPw, err := bdutils.GetHubDBPassword(hc.kubeClient, fromNamespace, bdspec.DbPrototype)
			if err != nil {
				return err
			}
			err = bdutils.CloneJob(hc.kubeClient, fromNamespace, bdspec.DbPrototype, bdspec.Namespace, name, fromPw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (hc *Creater) getPVCVolumeName(namespace string, name string) (string, error) {
	pvc, err := util.GetPVC(hc.kubeClient, namespace, name)
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

	resp, err := client.Get(fmt.Sprintf("https://%s.%s.svc:443/api/v1/registrations?summary=true", util.GetResourceName(bd.Name, util.BlackDuckName, "webserver"), bd.Spec.Namespace))
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
			if err := hc.autoRegisterHub(bd.Name, &bd.Spec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (hc *Creater) autoRegisterHub(name string, bdspec *blackduckapi.BlackduckSpec) error {
	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod, err := util.FilterPodByNamePrefixInNamespace(hc.kubeClient, bdspec.Namespace, util.GetResourceName(name, util.BlackDuckName, "registration"))
	if err != nil {
		return err
	}

	registrationKey := bdspec.LicenseKey

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for i := 0; i < 20; i++ {
			registrationPod, err := util.GetPod(hc.kubeClient, bdspec.Namespace, registrationPod.Name)
			if err != nil {
				return err
			}

			// Create the exec into Kubernetes pod request
			req := util.CreateExecContainerRequest(hc.kubeClient, registrationPod, "/bin/bash")
			// Exec into the Kubernetes pod and execute the commands
			_, err = util.ExecContainer(hc.kubeConfig, req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate" -k --cert /opt/blackduck/hub/hub-registration/security/blackduck_system.crt --key /opt/blackduck/hub/hub-registration/security/blackduck_system.key`, registrationKey)})

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
				mapValue := strings.TrimSpace(values[1])
				if strings.EqualFold(mapValue, "1") {
					return true
				}
			}
			return false
		}
	}
	return false
}
