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

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/pif"
	"k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func prettyPrint(v interface{}) {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	println(string(bytes))
}

func createResources(config *pif.Config, clientset *kubernetes.Clientset) {
	pifTester := pif.NewPifTester()
	pifTester.Config = config.PifTesterConfig()

	podPerceiver := model.NewPodPerceiver(config.AuxConfig.PodPerceiverServiceAccountName)
	podPerceiver.Config = config.PodPerceiverConfig()
	podPerceiver.Config.PerceptorHost = pifTester.ServiceName

	imagePerceiverReplicaCount := int32(1)
	imagePerceiver := model.NewImagePerceiver(imagePerceiverReplicaCount, config.AuxConfig.ImagePerceiverServiceAccountName)
	imagePerceiver.Config = config.ImagePerceiverConfig()
	imagePerceiver.Config.PerceptorHost = pifTester.ServiceName

	imageFacade := model.NewImagefacade(config.AuxConfig.ImageFacadeServiceAccountName)
	imageFacade.Config = config.ImagefacadeConfig()
	imageFacade.PodName = "perceptor-imagefacade"

	replicationControllers := []*v1.ReplicationController{
		podPerceiver.ReplicationController(),
		imagePerceiver.ReplicationController(),
		pifTester.ReplicationController(),
		imageFacade.ReplicationController(),
	}
	services := []*v1.Service{
		podPerceiver.Service(),
		imagePerceiver.Service(),
		pifTester.Service(),
		imageFacade.Service(),
	}
	configMaps := []*v1.ConfigMap{
		podPerceiver.ConfigMap(),
		imagePerceiver.ConfigMap(),
		pifTester.ConfigMap(),
		imageFacade.ConfigMap(),
	}

	namespace := config.AuxConfig.Namespace
	for _, configMap := range configMaps {
		_, err := clientset.Core().ConfigMaps(namespace).Create(configMap)
		if err != nil {
			panic(err)
		}
		prettyPrint(configMap)
	}
	for _, rc := range replicationControllers {
		_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
		if err != nil {
			panic(err)
		}
		prettyPrint(rc)
	}
	for _, service := range services {
		_, err := clientset.Core().Services(namespace).Create(service)
		if err != nil {
			panic(err)
		}
		prettyPrint(service)
	}
}

func main() {
	configPath := os.Args[1]
	auxConfigPath := os.Args[2]
	config := pif.ReadConfig(configPath)
	if config == nil {
		panic("didn't find config")
	}
	auxConfig := pif.ReadAuxiliaryConfig(auxConfigPath)
	if auxConfig == nil {
		panic("didn't find auxconfig")
	}
	config.AuxConfig = auxConfig
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	fmt.Printf("config: %s\n", string(jsonBytes))
	runProtoform(config)
}

func runProtoform(config *pif.Config) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(config.MasterURL, config.KubeConfigPath)
	//		kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	createResources(config, clientset)
}
