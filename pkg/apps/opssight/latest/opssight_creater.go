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

package opssight

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create OpsSight
type Creater struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	opsSightClient   *opssightclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	hubClient        *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client, hubClient *hubclientset.Clientset) *Creater {
	return &Creater{
		config:           config,
		kubeConfig:       kubeConfig,
		kubeClient:       kubeClient,
		opsSightClient:   opssightClient,
		osSecurityClient: osSecurityClient,
		routeClient:      routeClient,
		hubClient:        hubClient,
	}
}

// GetComponents returns the resource components for an OpsSight
func (oc *Creater) GetComponents(opsSight *opssightapi.OpsSight) (*api.ComponentList, error) {
	sc := NewSpecConfig(oc.config, oc.kubeClient, oc.opsSightClient, oc.hubClient, opsSight, oc.config.DryRun)
	return sc.GetComponents()
}

// Versions is an Interface function that returns the versions supported by this Creater
func (oc *Creater) Versions() []string {
	return GetVersions()
}

// Ensure is an Interface function that will make sure the instance is correctly deployed or deploy it if needed
func (oc *Creater) Ensure(opssight *opssightapi.OpsSight) error {
	return oc.UpdateOpsSight(opssight)
}

// UpdateOpsSight will create the OpsSight or update its components
func (oc *Creater) UpdateOpsSight(opssight *opssightapi.OpsSight) error {
	// log.Debugf("update OpsSight details for %s: %+v", opssight.Namespace, opssight)
	opssightSpec := &opssight.Spec
	// get the registry auth credentials for default OpenShift internal docker registries
	if !oc.config.DryRun {
		oc.addRegistryAuth(opssightSpec)
	}

	spec := NewSpecConfig(oc.config, oc.kubeClient, oc.opsSightClient, oc.hubClient, opssight, oc.config.DryRun)

	components, err := spec.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opssight.Spec.Namespace)
	}

	if !oc.config.DryRun {
		// call the CRUD updater to create or update opssight
		commonConfig := crdupdater.NewCRUDComponents(oc.kubeConfig, oc.kubeClient, oc.config.DryRun, false, opssightSpec.Namespace, components, "app=opssight")
		_, errs := commonConfig.CRUDComponents()

		if len(errs) > 0 {
			return fmt.Errorf("update components errors: %+v", errs)
		}

		// if OpenShift, add a privileged role to scanner account
		err = oc.postDeploy(spec, opssightSpec.Namespace)
		if err != nil {
			return errors.Annotatef(err, "post deploy")
		}

		err = oc.deployHub(opssightSpec)
		if err != nil {
			return errors.Annotatef(err, "deploy hub")
		}
	}

	return nil
}

func (oc *Creater) addRegistryAuth(opsSightSpec *opssightapi.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if oc.routeClient == nil {
		return
	}

	internalRegistries := []*string{}

	// Adding default image registry routes
	routes := map[string]string{"default": "docker-registry", "openshift-image-registry": "image-registry"}
	for namespace, name := range routes {
		route, err := util.GetRoute(oc.routeClient, namespace, name)
		if err != nil {
			continue
		}
		internalRegistries = append(internalRegistries, &route.Spec.Host)
		routeHostPort := fmt.Sprintf("%s:443", route.Spec.Host)
		internalRegistries = append(internalRegistries, &routeHostPort)
	}

	// Adding default OpenShift internal Docker/image registry service
	labelSelectors := []string{"docker-registry=default", "router in (router,router-default)"}
	for _, labelSelector := range labelSelectors {
		registrySvcs, err := util.ListServices(oc.kubeClient, "", labelSelector)
		if err != nil {
			continue
		}
		for _, registrySvc := range registrySvcs.Items {
			if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
				for _, port := range registrySvc.Spec.Ports {
					clusterIPSvc := fmt.Sprintf("%s:%d", registrySvc.Spec.ClusterIP, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvc)
					clusterIPSvcPort := fmt.Sprintf("%s.%s.svc:%d", registrySvc.Name, registrySvc.Namespace, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvcPort)
				}
			}
		}
	}

	file, err := util.ReadFromFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Errorf("unable to read the service account token file due to %+v", err)
	} else {
		for _, internalRegistry := range internalRegistries {
			opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightSpec.ScannerPod.ImageFacade.InternalRegistries, &opssightapi.RegistryAuth{URL: *internalRegistry, User: "admin", Password: string(file)})
		}
	}
}

func (oc *Creater) postDeploy(spec *SpecConfig, namespace string) error {
	// Need to add the perceptor-scanner service account to the privileged scc
	if oc.osSecurityClient != nil {
		scannerServiceAccount := spec.GetScannerServiceAccount()
		perceiverServiceAccount := spec.GetPodPerceiverServiceAccount()
		serviceAccounts := []string{fmt.Sprintf("system:serviceaccount:%s:%s", namespace, perceiverServiceAccount.GetName())}
		if !strings.EqualFold(spec.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
			serviceAccounts = append(serviceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, scannerServiceAccount.GetName()))
		}
		return util.UpdateOpenShiftSecurityConstraint(oc.osSecurityClient, serviceAccounts, "privileged")
	}
	return nil
}

func (oc *Creater) deployHub(createOpsSight *opssightapi.OpsSightSpec) error {
	if createOpsSight.Blackduck.InitialCount > createOpsSight.Blackduck.MaxCount {
		createOpsSight.Blackduck.InitialCount = createOpsSight.Blackduck.MaxCount
	}

	hubErrs := map[string]error{}
	for i := 0; i < createOpsSight.Blackduck.InitialCount; i++ {
		name := fmt.Sprintf("%s-%v", createOpsSight.Namespace, i)

		_, err := util.GetNamespace(oc.kubeClient, name)
		if err == nil {
			continue
		}

		ns, err := util.CreateNamespace(oc.kubeClient, name)
		log.Debugf("created namespace: %+v", ns)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the namespace due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the namespace due to %+v", err)
		}

		hubSpec := createOpsSight.Blackduck.BlackduckSpec
		hubSpec.Namespace = name
		createHub := &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *hubSpec}
		log.Debugf("hub[%d]: %+v", i, createHub)
		_, err = util.CreateHub(oc.hubClient, name, createHub)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the hub due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the hub due to %+v", err)
		}
	}

	return util.NewMapErrors(hubErrs)
}
