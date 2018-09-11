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

package hub

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/components"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/informers/externalversions/hub/v1"
	hubcontroller "github.com/blackducksoftware/perceptor-protoform/pkg/hub/controller"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub/webservice"

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
		return nil, fmt.Errorf("failed to convert hub defaults: %v", config)
	}
	d := &ControllerConfig{config: dependentConfig}

	d.config.resyncPeriod = 0
	d.config.indexers = cache.Indexers{}
	d.config.threadiness = 5

	return d, nil
}

// CreateClientSet will create the CRD client
func (c *ControllerConfig) CreateClientSet() {
	hubClient, err := hubclientset.NewForConfig(c.config.KubeConfig)
	if err != nil {
		log.Panicf("Unable to create Hub informer client: %s", err.Error())
	}
	c.config.customClientSet = hubClient
}

// Deploy will deploy the CRD and other relevant components
func (c *ControllerConfig) Deploy() error {
	deployer, err := horizon.NewDeployer(c.config.KubeConfig)
	if err != nil {
		return err
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "hubs.synopsys.com",
		Namespace:  c.config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Hub",
		Plural:     "hubs",
		Singular:   "hub",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	// Perceptor configMap
	hubFederatorConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.config.Namespace, Name: "hubfederator"})
	hubFederatorConfig.AddData(map[string]string{"config.json": fmt.Sprint(`{"HubConfig": {"User": "`, "sysadmin",
		`", "PasswordEnvVar": "`, "HUB_PASSWORD",
		`", "ClientTimeoutMilliseconds": `, 5000,
		`, "Port": `, "443", `, "FetchAllProjectsPauseSeconds": `, 60,
		`}, "UseMockMode": `, false, `, "LogLevel": "`, "debug", `", "Port": `, "3016", `}`)})
	deployer.AddConfigMap(hubFederatorConfig)

	// Perceptor service
	deployer.AddService(hub.CreateService("hub-federator", "hub-federator", c.config.Namespace, "3016", "3016", horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(hub.CreateService("hub-federator-np", "hub-federator", c.config.Namespace, "3016", "3016", horizonapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(hub.CreateService("hub-federator-lb", "hub-federator", c.config.Namespace, "3016", "3016", horizonapi.ClusterIPServiceTypeLoadBalancer))

	// Hub federator deployment
	hubFederatorContainerConfig := &hub.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-federator", Image: "gcr.io/gke-verification/blackducksoftware/federator:master",
			PullPolicy: horizonapi.PullAlways, Command: []string{"./federator"}, Args: []string{"/etc/hubfederator/config.json"}},
		EnvConfigs:   []*horizonapi.EnvConfig{{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_PASSWORD", KeyOrVal: "blackduck"}},
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "hubfederator", MountPath: "/etc/hubfederator", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: "3016", Protocol: horizonapi.ProtocolTCP},
	}
	hubFederatorVolume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hubfederator",
		MapOrSecretName: "hubfederator",
		DefaultMode:     hub.IntToInt32(420),
	})
	hubFederator := hub.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: c.config.Namespace, Name: "hub-federator", Replicas: hub.IntToInt32(1)},
		[]*hub.Container{hubFederatorContainerConfig}, []*components.Volume{hubFederatorVolume}, []*hub.Container{}, []horizonapi.AffinityConfig{})
	deployer.AddDeployment(hubFederator)

	certificate, key := hub.CreateSelfSignedCert()

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.config.Namespace, Name: "hub-certificate", Type: horizonapi.SecretTypeOpaque})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})

	deployer.AddSecret(certificateSecret)

	err = deployer.Run()

	if err != nil {
		log.Errorf("unable to create the hub federator resources due to %+v", err)
	}
	return err
}

// PostDeploy will call after deploying the CRD
func (c *ControllerConfig) PostDeploy() {
	hc := hub.NewCreater(c.config.KubeConfig, c.config.KubeClientSet, c.config.customClientSet)
	webservice.SetupHTTPServer(hc, c.config.Namespace)
}

// CreateInformer will create a informer for the CRD
func (c *ControllerConfig) CreateInformer() {
	c.config.infomer = hubinformerv1.NewHubInformer(
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
			log.Infof("add hub: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.config.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update hub: %s", key)
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
			log.Infof("delete hub: %s: %+v", key, obj)

			if err == nil {
				c.config.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *ControllerConfig) CreateHandler() {
	c.config.handler = &hubcontroller.HubHandler{
		Config:           c.config.KubeConfig,
		Clientset:        c.config.KubeClientSet,
		HubClientset:     c.config.customClientSet,
		Namespace:        c.config.Namespace,
		FederatorBaseURL: fmt.Sprintf("http://hub-federator:%d", 3016),
		CmMutex:          make(chan bool, 1),
	}
}

// CreateController will create a CRD controller
func (c *ControllerConfig) CreateController() {
	c.config.controller = hubcontroller.NewController(
		&hubcontroller.Controller{
			Logger:   log.NewEntry(log.New()),
			Queue:    c.config.queue,
			Informer: c.config.infomer,
			Handler:  c.config.handler,
		})
}

// Run will run the CRD controller
func (c *ControllerConfig) Run() {
	go c.config.controller.Run(c.config.threadiness, c.config.StopCh)
}

// PostRun will run post CRD controller execution
func (c *ControllerConfig) PostRun() {
	secretReplicator := hubcontroller.NewSecretReplicator(c.config.KubeClientSet, c.config.customClientSet, c.config.Namespace, c.config.resyncPeriod)
	go secretReplicator.Run(c.config.StopCh)
	<-c.config.StopCh
}
