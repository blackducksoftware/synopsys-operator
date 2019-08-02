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
	"reflect"
	"strings"
	"time"

	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	opssightinformer "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/informers/externalversions/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification
type CRDInstaller struct {
	protoformDeployer *protoform.Deployer
	defaults          interface{}
	indexers          cache.Indexers
	informer          cache.SharedIndexInformer
	queue             workqueue.RateLimitingInterface
	handler           *Handler
	controller        *Controller
	opssightclient    *opssightclientset.Clientset
	stopCh            <-chan struct{}
}

// NewCRDInstaller will create a controller configuration
func NewCRDInstaller(protoformDeployer *protoform.Deployer, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{protoformDeployer: protoformDeployer, defaults: defaults, stopCh: stopCh}
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	opssightClient, err := opssightclientset.NewForConfig(c.protoformDeployer.KubeConfig)
	if err != nil {
		return errors.Annotate(err, "Unable to create OpsSight informer client")
	}
	c.opssightclient = opssightClient
	return nil
}

// Deploy will deploy the CRD
func (c *CRDInstaller) Deploy() error {
	// Any new, pluggable maintainance stuff should go in here...
	blackDuckClient, err := blackduckclientset.NewForConfig(c.protoformDeployer.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	crdUpdater := opssight.NewUpdater(c.protoformDeployer.Config, c.protoformDeployer.KubeClient, blackDuckClient, c.opssightclient)
	go crdUpdater.Run(c.stopCh)
	return nil
}

// PostDeploy will initialize before deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	resyncPeriod := time.Duration(c.protoformDeployer.Config.ResyncIntervalInSeconds) * time.Second
	c.informer = opssightinformer.NewOpsSightInformer(
		c.opssightclient,
		c.protoformDeployer.Config.Namespace,
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
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add opssight: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.queue.Add(key)
			} else {
				log.Errorf("unable to add OpsSight: %v", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*opssightapi.OpsSight)
			new := newObj.(*opssightapi.OpsSight)
			if strings.EqualFold(old.Status.State, string(Running)) || !reflect.DeepEqual(old.Spec, new.Spec) || !reflect.DeepEqual(old.Status.InternalHosts, new.Status.InternalHosts) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				log.Infof("update opssight: %s", key)
				if err == nil {
					c.queue.Add(key)
				} else {
					log.Errorf("unable to update OpsSight: %v", err)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete opssight: %s: %+v", key, obj)

			if err == nil {
				c.queue.Add(key)
			} else {
				log.Errorf("unable to delete OpsSight: %v", err)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	securityClient := c.protoformDeployer.SecurityClient
	if securityClient != nil {
		_, err := util.GetOpenShiftSecurityConstraint(securityClient, "privileged")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			log.Debugf("ignoring scc privileged for Kubernetes cluster")
			securityClient = nil
		}
	}

	blackDuckClient, err := blackduckclientset.NewForConfig(c.protoformDeployer.KubeConfig)
	if err != nil {
		log.Errorf("unable to create the hub client for opssight: %+v", err)
		return
	}

	c.handler = &Handler{
		protoformDeployer: c.protoformDeployer,
		opsSightClient:    c.opssightclient,
		defaults:          c.defaults.(*opssightapi.OpsSightSpec),
		blackDuckClient:   blackDuckClient,
	}

	if util.IsOpenshift(c.protoformDeployer.KubeClient) {
		c.handler.protoformDeployer.RouteClient = util.GetRouteClient(c.protoformDeployer.KubeConfig)
		c.handler.protoformDeployer.SecurityClient = securityClient
	}
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.controller = NewController(
		&Controller{
			logger:   log.NewEntry(log.New()),
			queue:    c.queue,
			informer: c.informer,
			handler:  c.handler,
		})
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.controller.Run(c.protoformDeployer.Config.Threadiness, c.stopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
}
