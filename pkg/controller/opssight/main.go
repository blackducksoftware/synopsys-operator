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

	"github.com/blackducksoftware/horizon/pkg/components"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	opssightinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/informers/externalversions/opssight/v1"
	opssightcontroller "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/controller"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	log "github.com/sirupsen/logrus"
)

// Controller defines the specification for the controller
type ControllerConfig struct {
	config *ProtoformControllerConfig
}

// NewController will create a controller configuration
func NewController(config interface{}) (*ControllerConfig, error) {
	dependentConfig, ok := config.(*ProtoformControllerConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert opssight defaults: %v", config)
	}
	d := &ControllerConfig{config: dependentConfig}

	d.config.resyncPeriod = 0
	d.config.indexers = cache.Indexers{}
	d.config.threadiness = 5

	return d, nil
}

// CreateClientSet will create the CRD client
func (c *ControllerConfig) CreateClientSet() {
	opssightClient, err := opssightclientset.NewForConfig(c.config.KubeConfig)
	if err != nil {
		log.Panicf("Unable to create OpsSight informer client: %s", err.Error())
	}
	c.config.customClientSet = opssightClient
}

// Deploy will deploy the CRD
func (c *ControllerConfig) Deploy() error {
	deployer, err := horizon.NewDeployer(c.config.KubeConfig)
	if err != nil {
		return err
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "opssights.synopsys.com",
		Namespace:  c.config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "OpsSight",
		Plural:     "opssights",
		Singular:   "opssight",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	err = deployer.Run()
	return err
}

// PostDeploy will initialize before deploying the CRD
func (c *ControllerConfig) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *ControllerConfig) CreateInformer() {
	c.config.infomer = opssightinformerv1.NewOpsSightInformer(
		c.config.customClientSet,
		c.config.Namespace,
		c.config.resyncPeriod,
		c.config.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *ControllerConfig) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.config.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *ControllerConfig) AddInformerEventHandler() {
	c.config.infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add opssight: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.config.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update opssight: %s", key)
			if err == nil {
				c.config.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete opssight: %s: %+v", key, obj)

			if err == nil {
				c.config.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *ControllerConfig) CreateHandler() {
	c.config.handler = &opssightcontroller.OpsSightHandler{
		Config:            c.config.KubeConfig,
		Clientset:         c.config.KubeClientSet,
		OpsSightClientset: c.config.customClientSet,
		Namespace:         c.config.Namespace,
		CmMutex:           make(chan bool, 1),
	}
}

// CreateController will create a CRD controller
func (c *ControllerConfig) CreateController() {
	c.config.controller = opssightcontroller.NewController(
		&opssightcontroller.Controller{
			Logger:            log.NewEntry(log.New()),
			Clientset:         c.config.KubeClientSet,
			Queue:             c.config.queue,
			Informer:          c.config.infomer,
			Handler:           c.config.handler,
			OpsSightClientset: c.config.customClientSet,
			Namespace:         c.config.Namespace,
		})
}

// Run will run the CRD controller
func (c *ControllerConfig) Run() {
	go c.config.controller.Run(c.config.threadiness, c.config.StopCh)
	<-c.config.StopCh
}

// PostRun will run post CRD controller execution
func (c *ControllerConfig) PostRun() {
}
