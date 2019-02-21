/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package sample

import (
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sampleclientset "github.com/blackducksoftware/synopsys-operator/pkg/sample/client/clientset/versioned"
	sampleinformerv1 "github.com/blackducksoftware/synopsys-operator/pkg/sample/client/informers/externalversions/sample/v1"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification for the controller
type CRDInstaller struct {
	config       *protoform.Config
	kubeConfig   *rest.Config
	kubeClient   *kubernetes.Clientset
	defaults     interface{}
	resyncPeriod time.Duration
	indexers     cache.Indexers
	infomer      cache.SharedIndexInformer
	queue        workqueue.RateLimitingInterface
	handler      *Handler
	controller   *Controller
	sampleClient *sampleclientset.Clientset
	threadiness  int
	stopCh       <-chan struct{}
}

// NewCRDInstaller will create a installer configuration
func NewCRDInstaller(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, defaults: defaults, threadiness: config.Threadiness, stopCh: stopCh}
	crdInstaller.resyncPeriod = 0
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	sampleClient, err := sampleclientset.NewForConfig(c.kubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.sampleClient = sampleClient
	return nil
}

// Deploy will deploy the Sample's CRD into the Cluster
func (c *CRDInstaller) Deploy() error {
	// Deploy the Custom Defined Resource from Horizon
	deployer, err := horizon.NewDeployer(c.kubeConfig)
	if err != nil {
		return err
	}

	// Sample CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "samples.synopsys.com",
		Namespace:  c.config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Sample",
		Plural:     "samples",
		Singular:   "sample",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	err = deployer.Run()
	if err != nil {
		log.Errorf("Unable to create the Sample's CRD: %+v", err)
	}

	time.Sleep(5 * time.Second)

	return err
}

// PostDeploy will initialize before deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	c.infomer = sampleinformerv1.NewSampleInformer(
		c.sampleClient,
		c.config.Namespace,
		c.resyncPeriod,
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
			log.Infof("Add Sample Event: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update Sample Event: %s", key)
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
			log.Infof("Delete Sample Event: %s: %+v", key, obj)

			if err == nil {
				c.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	c.handler = &Handler{
		config:       c.config,
		kubeConfig:   c.kubeConfig,
		kubeClient:   c.kubeClient,
		sampleClient: c.sampleClient,
		defaults:     c.defaults.(*v1.SampleSpec),
	}
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.controller = NewController(log.NewEntry(log.New()), c.queue, c.infomer, c.handler)
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.controller.Run(c.config.Threadiness, c.stopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
}
