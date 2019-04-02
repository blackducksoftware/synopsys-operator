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
	"fmt"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	hubinformerv2 "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/informers/externalversions/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification for the CRD
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
	hubClient    *hubclientset.Clientset
	threadiness  int
	stopCh       <-chan struct{}
}

// NewCRDInstaller will create a CRD installer configuration
func NewCRDInstaller(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, defaults: defaults, threadiness: config.Threadiness, stopCh: stopCh}
	crdInstaller.resyncPeriod = 2 * time.Minute
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	hubClient, err := hubclientset.NewForConfig(c.kubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.hubClient = hubClient
	return nil
}

// Deploy will deploy the CRD and other relevant components
func (c *CRDInstaller) Deploy() error {
	deployer, err := horizon.NewDeployer(c.kubeConfig)
	if err != nil {
		return err
	}

	// Blackduck CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "blackducks.synopsys.com",
		Namespace:  c.config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Blackduck",
		Plural:     "blackducks",
		Singular:   "blackduck",
		ShortNames: []string{
			"hub",
			"hubs",
		},
		Scope: horizonapi.CRDClusterScoped,
	}))

	// // Perceptor configMap
	// hubFederatorConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.config.Namespace, Name: "federator"})

	// if c.config.HubFederatorConfig.HubConfig == nil {
	// 	panic("Cant start with nil federator configuration ! Set HubFederatorConfig with Port, User")
	// }

	// data := map[string]interface{}{
	// 	"HubConfig": map[string]interface{}{
	// 		"Port":                         c.config.HubFederatorConfig.HubConfig.Port,
	// 		"User":                         c.config.HubFederatorConfig.HubConfig.User,
	// 		"PasswordEnvVar":               c.config.HubFederatorConfig.HubConfig.PasswordEnvVar,
	// 		"ClientTimeoutMilliseconds":    c.config.HubFederatorConfig.HubConfig.ClientTimeoutMilliseconds,
	// 		"FetchAllProjectsPauseSeconds": c.config.HubFederatorConfig.HubConfig.FetchAllProjectsPauseSeconds,
	// 	},
	// 	"Port":        c.config.HubFederatorConfig.Port,
	// 	"LogLevel":    c.config.LogLevel,
	// 	"UseMockMode": c.config.HubFederatorConfig.UseMockMode,
	// }
	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return errors.Trace(err)
	// }
	// hubFederatorConfig.AddData(map[string]string{"config.json": string(bytes)})
	// deployer.AddConfigMap(hubFederatorConfig)

	// // Perceptor service
	// deployer.AddService(util.CreateService("federator", "federator", c.config.Namespace, fmt.Sprint(c.config.HubFederatorConfig.Port), fmt.Sprint(c.config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeDefault))
	// deployer.AddService(util.CreateService("federator-np", "federator", c.config.Namespace, fmt.Sprint(c.config.HubFederatorConfig.Port), fmt.Sprint(c.config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeNodePort))
	// deployer.AddService(util.CreateService("federator-lb", "federator", c.config.Namespace, fmt.Sprint(c.config.HubFederatorConfig.Port), fmt.Sprint(c.config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeLoadBalancer))

	// var hubPassword string
	// for {
	// 	blackduckSecret, err := util.GetSecret(c.kubeClient, c.config.Namespace, "blackduck-secret")
	// 	if err != nil {
	// 		log.Infof("Aborting: You need to first create a 'blackduck-secret' in the %v namespace with HUB_PASSWORD and retry", c.config.Namespace)
	// 	} else {
	// 		hubPassword = string(blackduckSecret.Data["HUB_PASSWORD"])
	// 		break
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }

	// Blackduck federator deployment
	// hubFederatorContainerConfig := &util.Container{
	// 	ContainerConfig: &horizonapi.ContainerConfig{Name: "federator", Image: fmt.Sprintf("%s/%s/%s:%s", c.config.HubFederatorConfig.Registry, c.config.HubFederatorConfig.ImagePath, c.config.HubFederatorConfig.ImageName, c.config.HubFederatorConfig.ImageVersion),
	// 		PullPolicy: horizonapi.PullAlways, Command: []string{"./federator"}, Args: []string{"/etc/federator/config.json"}},
	// 	EnvConfigs:   []*horizonapi.EnvConfig{{Type: horizonapi.EnvVal, NameOrPrefix: c.config.HubFederatorConfig.HubConfig.PasswordEnvVar, KeyOrVal: hubPassword}},
	// 	VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "federator", MountPath: "/etc/federator"}},
	// 	PortConfig:   &horizonapi.PortConfig{ContainerPort: fmt.Sprint(c.config.HubFederatorConfig.Port), Protocol: horizonapi.ProtocolTCP},
	// }
	// hubFederatorVolume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
	// 	VolumeName:      "federator",
	// 	MapOrSecretName: "federator",
	// 	DefaultMode:     util.IntToInt32(420),
	// })
	// hubFederator := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.config.Namespace, Name: "federator", Replicas: util.IntToInt32(1)}, "",
	// 	[]*util.Container{hubFederatorContainerConfig}, []*components.Volume{hubFederatorVolume}, []*util.Container{}, []horizonapi.AffinityConfig{})
	// deployer.AddReplicationController(hubFederator)

	certificate, key := CreateSelfSignedCert()

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.config.Namespace, Name: "blackduck-certificate", Type: horizonapi.SecretTypeOpaque})
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
func (c *CRDInstaller) PostDeploy() {
	protoform.SetupHTTPServer(c.kubeClient, c.hubClient, c.config.Namespace)
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	c.infomer = hubinformerv2.NewBlackduckInformer(
		c.hubClient,
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

	osClient, err := securityclient.NewForConfig(c.kubeConfig)
	if err != nil {
		osClient = nil
	} else {
		_, err := util.GetOpenShiftSecurityConstraint(osClient, "anyuid")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			log.Debugf("Ignoring scc privileged for kubernetes cluster")
			osClient = nil
		}
	}

	routeClient, err := routeclient.NewForConfig(c.kubeConfig)
	if err != nil {
		routeClient = nil
	} else {
		_, err := util.GetOpenShiftRoutes(routeClient, "default", "docker-registry")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			log.Debugf("Ignoring routes for kubernetes cluster")
			routeClient = nil
		}
	}

	c.handler = NewHandler(c.config, c.kubeConfig, c.kubeClient, c.hubClient, c.defaults.(*v1.BlackduckSpec), fmt.Sprint("http://federator:3016"), make(chan bool, 1), osClient, routeClient)
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.controller = NewController(log.NewEntry(log.New()), c.queue, c.infomer, c.handler)
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.controller.Run(c.threadiness, c.stopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
	//secretReplicator := plugins.NewSecretReplicator(c.kubeClient, c.hubClient, c.config.Namespace, c.resyncPeriod)
	//go secretReplicator.Run(c.stopCh)
}
