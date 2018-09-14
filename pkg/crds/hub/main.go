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
	"time"

	"github.com/blackducksoftware/horizon/pkg/components"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/informers/externalversions/hub/v1"
	hubcontroller "github.com/blackducksoftware/perceptor-protoform/pkg/hub/controller"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub/webservice"

	log "github.com/sirupsen/logrus"
)

// ControllerConfig defines the specification for the controller
type ControllerConfig struct {
	protoformConfig *ProtoformControllerConfig
}

// NewController will create a controller configuration
func NewController(config interface{}) (*ControllerConfig, error) {
	dependentConfig, ok := config.(*ProtoformControllerConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert hub defaults: %v", config)
	}
	d := &ControllerConfig{protoformConfig: dependentConfig}

	d.protoformConfig.resyncPeriod = 0
	d.protoformConfig.indexers = cache.Indexers{}

	return d, nil
}

// CreateClientSet will create the CRD client
func (c *ControllerConfig) CreateClientSet() {
	hubClient, err := hubclientset.NewForConfig(c.protoformConfig.KubeConfig)
	if err != nil {
		log.Panicf("Unable to create Hub informer client: %s", err.Error())
	}
	c.protoformConfig.customClientSet = hubClient
}

// Deploy will deploy the CRD and other relevant components
func (c *ControllerConfig) Deploy() error {
	deployer, err := horizon.NewDeployer(c.protoformConfig.KubeConfig)
	if err != nil {
		return err
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "hubs.synopsys.com",
		Namespace:  c.protoformConfig.Config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Hub",
		Plural:     "hubs",
		Singular:   "hub",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	// Perceptor configMap
	hubFederatorConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.protoformConfig.Config.Namespace, Name: "hubfederator"})
	hubFederatorConfig.AddData(map[string]string{"config.json": fmt.Sprint(`{"HubConfig": {"User": "`, c.protoformConfig.Config.HubFederatorConfig.HubConfig.User,
		`", "PasswordEnvVar": "`, c.protoformConfig.Config.HubFederatorConfig.HubConfig.PasswordEnvVar,
		`", "ClientTimeoutMilliseconds": `, c.protoformConfig.Config.HubFederatorConfig.HubConfig.ClientTimeoutMilliseconds,
		`, "Port": `, c.protoformConfig.Config.HubFederatorConfig.HubConfig.Port,
		`, "FetchAllProjectsPauseSeconds": `, c.protoformConfig.Config.HubFederatorConfig.HubConfig.FetchAllProjectsPauseSeconds,
		`}, "UseMockMode": `, c.protoformConfig.Config.HubFederatorConfig.UseMockMode, `, "LogLevel": "`, c.protoformConfig.Config.LogLevel,
		`", "Port": `, c.protoformConfig.Config.HubFederatorConfig.Port, `}`)})
	deployer.AddConfigMap(hubFederatorConfig)

	// Perceptor service
	deployer.AddService(util.CreateService("hub-federator", "hub-federator", c.protoformConfig.Config.Namespace, fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("hub-federator-np", "hub-federator", c.protoformConfig.Config.Namespace, fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(util.CreateService("hub-federator-lb", "hub-federator", c.protoformConfig.Config.Namespace, fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeLoadBalancer))

	// Hub federator deployment
	hubFederatorContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-federator", Image: "gcr.io/gke-verification/blackducksoftware/federator:master",
			PullPolicy: horizonapi.PullAlways, Command: []string{"./federator"}, Args: []string{"/etc/hubfederator/config.json"}},
		EnvConfigs:   []*horizonapi.EnvConfig{{Type: horizonapi.EnvVal, NameOrPrefix: c.protoformConfig.Config.HubFederatorConfig.HubConfig.PasswordEnvVar, KeyOrVal: "blackduck"}},
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "hubfederator", MountPath: "/etc/hubfederator", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: fmt.Sprint(c.protoformConfig.Config.HubFederatorConfig.Port), Protocol: horizonapi.ProtocolTCP},
	}
	hubFederatorVolume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hubfederator",
		MapOrSecretName: "hubfederator",
		DefaultMode:     util.IntToInt32(420),
	})
	hubFederator := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: c.protoformConfig.Config.Namespace, Name: "hub-federator", Replicas: util.IntToInt32(1)},
		[]*util.Container{hubFederatorContainerConfig}, []*components.Volume{hubFederatorVolume}, []*util.Container{}, []horizonapi.AffinityConfig{})
	deployer.AddDeployment(hubFederator)

	certificate, key := hub.CreateSelfSignedCert()

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.protoformConfig.Config.Namespace, Name: "hub-certificate", Type: horizonapi.SecretTypeOpaque})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})

	deployer.AddSecret(certificateSecret)

	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to create the hub federator resources due to %+v", err)
	}

	time.Sleep(5 * time.Second)

	return err
}

// PostDeploy will call after deploying the CRD
func (c *ControllerConfig) PostDeploy() {
	hc := hub.NewCreater(c.protoformConfig.KubeConfig, c.protoformConfig.KubeClientSet, c.protoformConfig.customClientSet)
	webservice.SetupHTTPServer(hc, c.protoformConfig.Config.Namespace)
}

// CreateInformer will create a informer for the CRD
func (c *ControllerConfig) CreateInformer() {
	c.protoformConfig.infomer = hubinformerv1.NewHubInformer(
		c.protoformConfig.customClientSet,
		c.protoformConfig.Config.Namespace,
		c.protoformConfig.resyncPeriod,
		c.protoformConfig.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *ControllerConfig) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.protoformConfig.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *ControllerConfig) AddInformerEventHandler() {
	c.protoformConfig.infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add hub: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.protoformConfig.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update hub: %s", key)
			if err == nil {
				c.protoformConfig.queue.Add(key)
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
				c.protoformConfig.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *ControllerConfig) CreateHandler() {
	c.protoformConfig.handler = &hubcontroller.HubHandler{
		Config:           c.protoformConfig.KubeConfig,
		Clientset:        c.protoformConfig.KubeClientSet,
		HubClientset:     c.protoformConfig.customClientSet,
		Namespace:        c.protoformConfig.Config.Namespace,
		FederatorBaseURL: fmt.Sprintf("http://hub-federator:%d", c.protoformConfig.Config.HubFederatorConfig.Port),
		CmMutex:          make(chan bool, 1),
	}
}

// CreateController will create a CRD controller
func (c *ControllerConfig) CreateController() {
	c.protoformConfig.controller = hubcontroller.NewController(
		&hubcontroller.Controller{
			Logger:   log.NewEntry(log.New()),
			Queue:    c.protoformConfig.queue,
			Informer: c.protoformConfig.infomer,
			Handler:  c.protoformConfig.handler,
		})
}

// Run will run the CRD controller
func (c *ControllerConfig) Run() {
	go c.protoformConfig.controller.Run(c.protoformConfig.Threadiness, c.protoformConfig.StopCh)
}

// PostRun will run post CRD controller execution
func (c *ControllerConfig) PostRun() {
	secretReplicator := hubcontroller.NewSecretReplicator(c.protoformConfig.KubeClientSet, c.protoformConfig.customClientSet, c.protoformConfig.Config.Namespace, c.protoformConfig.resyncPeriod)
	go secretReplicator.Run(c.protoformConfig.StopCh)
}
