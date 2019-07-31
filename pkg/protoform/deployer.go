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

package protoform

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"strings"

	crd "github.com/blackducksoftware/synopsys-operator/pkg/crds"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Deployer handles deploying configured components to a cluster
type Deployer struct {
	Config              *Config
	KubeConfig          *rest.Config
	KubeClient          *kubernetes.Clientset
	APIExtensionsClient *apiextensionsclient.Clientset
	RouteClient         *routeclient.RouteV1Client
	SecurityClient      *securityclient.SecurityV1Client
	controllers         []crd.ProtoformControllerInterface
}

// NewDeployer will create the specification that is used for deploying controllers
func NewDeployer(config *Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset) (*Deployer, error) {
	apiExtensionsClient, err := apiextensionsclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating API Extensions Client: %s", err)
	}

	// Make OpenShift clients
	var routeClient *routeclient.RouteV1Client
	var securityClient *securityclient.SecurityV1Client
	if config.IsOpenshift {
		routeClient, err = routeclient.NewForConfig(kubeConfig)
		if err != nil {
			// If we can't make a routeClient, then a customer should still be able to deploy the application
			fmt.Printf("on OpenShift, still an error creating Route Client: %s", err)
		}
		// We only need securityClient for OpsSight
		if strings.Contains(config.CrdNames, util.OpsSightCRDName) {
			securityClient, err = securityclient.NewForConfig(kubeConfig)
			if err != nil {
				return nil, fmt.Errorf("on OpenShift, still an error creating Security Client: %s", err)
			}
		}
	}
	deployer := Deployer{
		Config:              config,
		KubeConfig:          kubeConfig,
		KubeClient:          kubeClient,
		APIExtensionsClient: apiExtensionsClient,
		RouteClient:         routeClient,
		SecurityClient:      securityClient,
		controllers:         make([]crd.ProtoformControllerInterface, 0),
	}
	return &deployer, nil
}

// AddController will add the controllers to the list
func (d *Deployer) AddController(controller crd.ProtoformControllerInterface) {
	d.controllers = append(d.controllers, controller)
}

// Cleanup makes an empty controller list
func (d *Deployer) Cleanup() {
	d.controllers = make([]crd.ProtoformControllerInterface, 0)
}

// Deploy will deploy the controllers
func (d *Deployer) Deploy() error {
	for _, controller := range d.controllers {
		err := controller.CreateClientSet()
		if err != nil {
			return errors.Annotate(err, "unable to create clientset for controller")
		}
		controller.Deploy()
		controller.PostDeploy()
		controller.CreateInformer()
		controller.CreateQueue()
		controller.AddInformerEventHandler()
		controller.CreateHandler()
		controller.CreateController()
		controller.Run()
		controller.PostRun()
	}
	return nil
}
