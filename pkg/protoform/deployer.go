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
	controller "github.com/blackducksoftware/perceptor-protoform/pkg/controllers"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Deployer handles deploying configured components to a cluster
type Deployer struct {
	Config        *model.Config
	KubeConfig    *rest.Config
	KubeClientSet *kubernetes.Clientset
	Namespace     string
	StopCh        <-chan struct{}
	controllerMap map[controller.ControllerType]interface{}
	controllers   []controller.ProtoformControllerInterface
}

// NewDeployer will create the specification that is used for deploying controllers
func NewDeployer(config *model.Config, kubeConfig *rest.Config, kubeClientSet *kubernetes.Clientset, namespace string, stopCh <-chan struct{}) *Deployer {
	deployer := Deployer{
		Config:        config,
		controllerMap: make(map[controller.ControllerType]interface{}),
		controllers:   make([]controller.ProtoformControllerInterface, 0),
	}
	return &deployer
}

// LoadController will store the defaults for the provided controller
func (d *Deployer) LoadController(controller controller.ControllerType, defaults interface{}) {
	d.controllerMap[controller] = defaults
}

// AddController will add the controllers to the list
func (d *Deployer) AddController(controller controller.ProtoformControllerInterface) {
	d.controllers = append(d.controllers, controller)
}

// Deploy will deploy the controllers
func (d *Deployer) Deploy() {
	for _, controller := range d.controllers {
		controller.CreateClientSet()
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
}
