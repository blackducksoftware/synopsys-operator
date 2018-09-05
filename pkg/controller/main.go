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

package controller

import (
	"flag"
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/client/informers/externalversions/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	model "github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/webservice"
)

// RunHubController will initialize the input config file, create the hub informers, initiantiate all rest api
func RunHubController(configPath string) {
	config, err := model.GetConfig(configPath)
	if err != nil {
		log.Errorf("Failed to load configuration: %s", err.Error())
		panic(err)
	}
	if config == nil {
		err = fmt.Errorf("expected non-nil config, but got nil")
		log.Errorf(err.Error())
		panic(err)
	}

	level, err := config.GetLogLevel()
	if err != nil {
		log.Errorf(err.Error())
		panic(err)
	}
	log.SetLevel(level)

	log.Debugf("config: %+v", config)

	// creates the in-cluster config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("error getting in cluster config. Fallback to native config. Error message: %s", err)
		kubeConfig, err = newKubeClientFromOutsideCluster()
	}

	if err != nil {
		log.Panicf("error getting the default client config: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Panicf("unable to create kubernetes clientset: %s", err.Error())
	}

	hubResourceClient, err := hubclientset.NewForConfig(kubeConfig)
	if err != nil {
		log.Panicf("Unable to create Hub informer client: %s", err.Error())
	}

	deploy(kubeConfig, config)

	hc, err := hub.NewCreater(kubeConfig, clientset, hubResourceClient)
	webservice.SetupHTTPServer(hc, config)

	informer := hubinformerv1.NewHubInformer(
		hubResourceClient,
		config.Namespace,
		0,
		cache.Indexers{},
	)

	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add hub: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update hub: %s", key)
			if err == nil {
				queue.Add(key)
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
				queue.Add(key)
			}
		},
	})

	controller := Controller{
		logger:    log.NewEntry(log.New()),
		clientset: clientset,
		queue:     queue,
		informer:  informer,
		handler: &HubHandler{
			config:           kubeConfig,
			clientset:        clientset,
			hubClientset:     hubResourceClient,
			namespace:        config.Namespace,
			federatorBaseURL: fmt.Sprintf("http://hub-federator:%d", config.HubFederatorConfig.Port),
			cmMutex:          make(chan bool, 1),
		},
		hubclientset: hubResourceClient,
		namespace:    config.Namespace,
	}

	stopCh := make(chan struct{})

	defer close(stopCh)
	secretReplicator := NewSecretReplicator(clientset, hubResourceClient, config.Namespace, 0)

	go controller.Run(config.Threadiness, stopCh)
	go secretReplicator.Run(stopCh)

	<-stopCh
}

func deploy(kubeConfig *rest.Config, config *model.Config) {
	deployer, err := horizon.NewDeployer(kubeConfig)
	if err != nil {
		log.Errorf("unable to create the deployer object due to %+v", err)
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(kapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "hubs.synopsys.com",
		Namespace:  config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Hub",
		Plural:     "hubs",
		Singular:   "hub",
		Scope:      kapi.CRDClusterScoped,
	}))

	// Perceptor configMap
	hubFederatorConfig := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: config.Namespace, Name: "hubfederator"})
	hubFederatorConfig.AddData(map[string]string{"config.json": fmt.Sprint(`{"HubConfig": {"User": "`, config.HubFederatorConfig.HubConfig.User,
		`", "PasswordEnvVar": "`, config.HubFederatorConfig.HubConfig.PasswordEnvVar,
		`", "ClientTimeoutMilliseconds": `, config.HubFederatorConfig.HubConfig.ClientTimeoutMilliseconds,
		`, "Port": `, config.HubFederatorConfig.HubConfig.Port, `, "FetchAllProjectsPauseSeconds": `, config.HubFederatorConfig.HubConfig.FetchAllProjectsPauseSeconds,
		`}, "UseMockMode": `, config.HubFederatorConfig.UseMockMode, `, "LogLevel": "`, config.LogLevel, `", "Port": `, config.HubFederatorConfig.Port, `}`)})
	deployer.AddConfigMap(hubFederatorConfig)

	// Perceptor service
	deployer.AddService(hub.CreateService("hub-federator", "hub-federator", config.Namespace, fmt.Sprint(config.HubFederatorConfig.Port), fmt.Sprint(config.HubFederatorConfig.Port), kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(hub.CreateService("hub-federator-np", "hub-federator", config.Namespace, fmt.Sprint(config.HubFederatorConfig.Port), fmt.Sprint(config.HubFederatorConfig.Port), kapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(hub.CreateService("hub-federator-lb", "hub-federator", config.Namespace, fmt.Sprint(config.HubFederatorConfig.Port), fmt.Sprint(config.HubFederatorConfig.Port), kapi.ClusterIPServiceTypeLoadBalancer))

	// Hub federator deployment
	hubFederatorContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-federator", Image: "gcr.io/gke-verification/blackducksoftware/federator:master",
			PullPolicy: kapi.PullAlways, Command: []string{"./federator"}, Args: []string{"/etc/hubfederator/config.json"}},
		EnvConfigs:   []*kapi.EnvConfig{{Type: kapi.EnvVal, NameOrPrefix: config.HubFederatorConfig.HubConfig.PasswordEnvVar, KeyOrVal: "blackduck"}},
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "hubfederator", MountPath: "/etc/hubfederator", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: fmt.Sprint(config.HubFederatorConfig.Port), Protocol: kapi.ProtocolTCP},
	}
	hubFederatorVolume := components.NewConfigMapVolume(kapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hubfederator",
		MapOrSecretName: "hubfederator",
		DefaultMode:     hub.IntToInt32(420),
	})
	hubFederator := hub.CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: config.Namespace, Name: "hub-federator", Replicas: hub.IntToInt32(1)},
		[]*api.Container{hubFederatorContainerConfig}, []*components.Volume{hubFederatorVolume}, []*api.Container{}, []kapi.AffinityConfig{})
	deployer.AddDeployment(hubFederator)

	certificate, key := hub.CreateSelfSignedCert()

	certificateSecret := components.NewSecret(kapi.SecretConfig{Namespace: config.Namespace, Name: "hub-certificate", Type: kapi.SecretTypeOpaque})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})

	deployer.AddSecret(certificateSecret)

	err = deployer.Run()

	if err != nil {
		log.Errorf("unable to create the hub federator resources due to %+v", err)
	}
}

func newKubeClientFromOutsideCluster() (*rest.Config, error) {
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Errorf("error creating default client config: %s", err)
		return nil, err
	}
	return config, err
}
