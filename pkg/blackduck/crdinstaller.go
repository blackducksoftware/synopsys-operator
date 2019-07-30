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
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	hubinformerv2 "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/informers/externalversions/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification for the CRD
type CRDInstaller struct {
	protformDeployer *protoform.Deployer
	defaults         interface{}
	indexers         cache.Indexers
	infomer          cache.SharedIndexInformer
	queue            workqueue.RateLimitingInterface
	handler          *Handler
	controller       *Controller
	hubClient        *hubclientset.Clientset
	stopCh           <-chan struct{}
}

// NewCRDInstaller will create a CRD installer configuration
func NewCRDInstaller(protoformDeployer *protoform.Deployer, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{protformDeployer: protoformDeployer, defaults: defaults, stopCh: stopCh}
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	hubClient, err := hubclientset.NewForConfig(c.protformDeployer.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.hubClient = hubClient
	return nil
}

// Deploy will deploy the CRD and other relevant components
func (c *CRDInstaller) Deploy() error {
	deployer, err := horizon.NewDeployer(c.protformDeployer.KubeConfig)
	if err != nil {
		return err
	}

	// creation of default blackduck nginx certificate secret
	certificate, key := CreateSelfSignedCert()
	certificateSecret := components.NewSecret(horizonapi.SecretConfig{
		Namespace: c.protformDeployer.Config.Namespace,
		Name:      "blackduck-certificate",
		Type:      horizonapi.SecretTypeOpaque,
	})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})
	deployer.AddComponent(horizonapi.SecretComponent, certificateSecret)

	_, err = util.GetSecret(c.protformDeployer.KubeClient, c.protformDeployer.Config.Namespace, "blackduck-secret")
	if err != nil {
		sealKey, err := util.GetRandomString(32)
		if err != nil {
			log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
		}
		// creation of blackduck secret to store the seal key
		operatorSecret := components.NewSecret(horizonapi.SecretConfig{
			Name:      "blackduck-secret",
			Namespace: c.protformDeployer.Config.Namespace,
			Type:      horizonapi.SecretTypeOpaque,
		})
		operatorSecret.AddData(map[string][]byte{
			"SEAL_KEY": []byte(sealKey),
		})

		operatorSecret.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
		deployer.AddComponent(horizonapi.SecretComponent, operatorSecret)
	} else {
		log.Warn("blackduck-secret already exists")
	}

	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to create the black duck secrets due to %+v", err)
	}

	time.Sleep(5 * time.Second)

	return err
}

// PostDeploy will call after deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	resyncPeriod := time.Duration(c.protformDeployer.Config.ResyncIntervalInSeconds) * time.Second
	c.infomer = hubinformerv2.NewBlackduckInformer(
		c.hubClient,
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
			if err == nil {
				// add the key to the queue for the handler to get
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
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
			log.Infof("delete hub: %s: %+v", key, obj)

			if err == nil {
				c.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	c.handler = NewHandler(c.protformDeployer, c.hubClient, c.defaults.(*v1.BlackduckSpec))
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
