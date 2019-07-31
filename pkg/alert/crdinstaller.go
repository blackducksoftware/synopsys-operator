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

package alert

import (
	"time"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertinformerv1 "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/informers/externalversions/alert/v1"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification for the controller
type CRDInstaller struct {
	protformDeployer *protoform.Deployer
	defaults         interface{}
	indexers         cache.Indexers
	infomer          cache.SharedIndexInformer
	queue            workqueue.RateLimitingInterface
	handler          *Handler
	controller       *Controller
	alertClient      *alertclientset.Clientset
	stopCh           <-chan struct{}
}

// NewCRDInstaller will create a installer configuration
func NewCRDInstaller(protoformDeployer *protoform.Deployer, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{protformDeployer: protoformDeployer, defaults: defaults, stopCh: stopCh}
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	alertClient, err := alertclientset.NewForConfig(c.protformDeployer.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.alertClient = alertClient
	return nil
}

// Deploy will deploy the CRD
func (c *CRDInstaller) Deploy() error {
	return nil
}

// PostDeploy will initialize before deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	resyncPeriod := time.Duration(c.protformDeployer.Config.ResyncIntervalInSeconds) * time.Second
	c.infomer = alertinformerv1.NewAlertInformer(
		c.alertClient,
		c.protformDeployer.Config.Namespace,
		resyncPeriod,
		c.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *CRDInstaller) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *CRDInstaller) AddInformerEventHandler() {
	c.infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add alert: %s, err: %+v", key, err)
			if err == nil {
				// add the key to the queue for the handler to get
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update alert: %s, err: %+v", key, err)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete alert: %s, err: %+v", key, err)

			if err == nil {
				c.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	c.handler = &Handler{
		protoformDeployer: c.protformDeployer,
		alertClient:       c.alertClient,
		defaults:          c.defaults.(*alertapi.AlertSpec),
	}
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.controller = NewController(log.NewEntry(log.New()), c.queue, c.infomer, c.handler)
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.controller.Run(c.protformDeployer.Config.Threadiness, c.stopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
}
