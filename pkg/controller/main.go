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

	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/client/informers/externalversions/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/webservice"
)

// RunHubController will initialize the input config file, create the hub informers, initiantiate all rest api
func RunHubController(configPath string) {
	config, err := GetConfig(configPath)
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

	hc, err := hub.NewCreater(kubeConfig, clientset, hubResourceClient)
	webservice.SetupHTTPServer(hc)

	handler := HubHandler{
		config:       kubeConfig,
		clientset:    clientset,
		hubClientset: hubResourceClient,
		crdNamespace: config.CrdNamespace,
	}

	informer := hubinformerv1.NewHubInformer(
		hubResourceClient,
		config.CrdNamespace,
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
		crdNamespace: config.CrdNamespace,
	}

	stopCh := make(chan struct{})

	defer close(stopCh)

	go controller.Run(stopCh)

	<-stopCh
}

func newKubeClientFromOutsideCluster() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Errorf("error creating default client config: %s", err)
		return nil, err
	}
	return config, err
}
