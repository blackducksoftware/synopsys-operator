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
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/client/informers/externalversions/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	model "github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/webservice"
)

const perceptorPort = "3016"

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

	handler := HubHandler{
		config:       kubeConfig,
		clientset:    clientset,
		hubClientset: hubResourceClient,
		namespace:    config.Namespace,
	}

	informer := hubinformerv1.NewHubInformer(
		hubResourceClient,
		config.Namespace,
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handler.ObjectCreated(obj.(*hubv1.Hub))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			handler.ObjectUpdated(oldObj.(*hubv1.Hub), newObj.(*hubv1.Hub))
		},
		DeleteFunc: func(obj interface{}) {
			handler.ObjectDeleted(obj.(*hubv1.Hub))
		},
	})

	controller := Controller{
		logger:       log.NewEntry(log.New()),
		clientset:    clientset,
		informer:     informer,
		hubclientset: hubResourceClient,
		namespace:    config.Namespace,
	}

	stopCh := make(chan struct{})

	defer close(stopCh)

	go controller.Run(stopCh)

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
	perceptorConfig := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: config.Namespace, Name: "hubfederator"})
	perceptorConfig.AddData(map[string]string{"config.json": fmt.Sprint(`{"HubConfig": {"User": "sysadmin", "PasswordEnvVar": "HUB_PASSWORD", "ClientTimeoutMilliseconds": 5000, "Port": 443, "FetchAllProjectsPauseMinutes": 5}, "UseMockMode": false, "LogLevel": "debug", "Port": "3016"}`)})
	deployer.AddConfigMap(perceptorConfig)

	// Perceptor service
	deployer.AddService(hub.CreateService("hub-federator", "hub-federator", config.Namespace, perceptorPort, perceptorPort, false))

	// Perceptor deployment
	perceptorContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-federator", Image: "gcr.io/gke-verification/blackducksoftware/federator:hub", PullPolicy: kapi.PullAlways},
		EnvConfigs:      []*kapi.EnvConfig{{Type: kapi.EnvVal, NameOrPrefix: "HUB_PASSWORD", KeyOrVal: "blackduck"}},
		VolumeMounts:    []*kapi.VolumeMountConfig{{Name: "hubfederator", MountPath: "/etc/hubfederator", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:      &kapi.PortConfig{ContainerPort: perceptorPort, Protocol: kapi.ProtocolTCP},
	}
	perceptorVolume := components.NewConfigMapVolume(kapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hubfederator",
		MapOrSecretName: "hubfederator",
		DefaultMode:     hub.IntToInt32(420),
	})
	perceptor := hub.CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: config.Namespace, Name: "hub-federator", Replicas: hub.IntToInt32(1)},
		[]*api.Container{perceptorContainerConfig}, []*components.Volume{perceptorVolume}, []*api.Container{}, []kapi.AffinityConfig{})
	deployer.AddDeployment(perceptor)

	err = deployer.Run()

	if err != nil {
		log.Errorf("unable to create the perceptor resources due to %+v", err)
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
