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

	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func createPerceptorResources(config *model.ProtoformConfig, clientset *kubernetes.Clientset) {
	perceptor := model.NewPerceptorCore()
	perceptor.Config = config.PerceptorConfig()

	podPerceiver := model.NewPodPerceiver(config.AuxConfig.PodPerceiverServiceAccountName)
	podPerceiver.Config = config.PodPerceiverConfig()

	imagePerceiverReplicaCount := int32(0)
	imagePerceiver := model.NewImagePerceiver(imagePerceiverReplicaCount, config.AuxConfig.ImagePerceiverServiceAccountName)
	imagePerceiver.Config = config.ImagePerceiverConfig()

	perceptorScanner := model.NewPerceptorScanner()
	perceptorScanner.Config = config.PerceptorScannerConfig()

	perceptorImagefacade := model.NewPerceptorImagefacade(config.AuxConfig.ImageFacadeServiceAccountName)
	perceptorImagefacade.Config = config.PerceptorImagefacadeConfig()

	prometheus := model.NewPrometheus()
	//	prometheus.Config = config.PrometheusConfig() // TODO ?

	scanner := model.NewScanner(perceptorScanner, perceptorImagefacade)

	replicationControllers := []*v1.ReplicationController{
		perceptor.ReplicationController(),
		podPerceiver.ReplicationController(),
		imagePerceiver.ReplicationController(),
		scanner.ReplicationController(),
	}
	services := []*v1.Service{
		perceptor.Service(),
		podPerceiver.Service(),
		imagePerceiver.Service(),
		perceptorScanner.Service(),
		perceptorImagefacade.Service(),
		prometheus.Service(),
	}
	configMaps := []*v1.ConfigMap{
		perceptor.ConfigMap(),
		podPerceiver.ConfigMap(),
		imagePerceiver.ConfigMap(),
		perceptorScanner.ConfigMap(),
		perceptorImagefacade.ConfigMap(),
		prometheus.ConfigMap(),
	}
	deployments := []*v1beta1.Deployment{
		prometheus.Deployment(),
	}

	namespace := config.AuxConfig.Namespace

	for _, configMap := range configMaps {
		_, err := clientset.Core().ConfigMaps(namespace).Create(configMap)
		if err != nil {
			panic(err)
		}
	}
	for _, service := range services {
		_, err := clientset.Core().Services(namespace).Create(service)
		if err != nil {
			panic(err)
		}
	}
	for _, rc := range replicationControllers {
		PrettyPrint(rc)
		_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
		if err != nil {
			panic(err)
		}
	}
	for _, dep := range deployments {
		_, err := clientset.Extensions().Deployments(namespace).Create(dep)
		if err != nil {
			panic(err)
		}
	}
}

func CreateConfigMapsFromInput(namespace string, clientset *kubernetes.Clientset, configMaps []*v1.ConfigMap) {
	for _, configMap := range configMaps {
		clientset.Core().ConfigMaps(namespace).Create(configMap)
	}
}

func main() {
	configPath := os.Args[1]
	auxConfigPath := os.Args[2]
	config := model.ReadProtoformConfig(configPath)
	if config == nil {
		panic("didn't find config")
	}
	auxConfig := model.ReadAuxiliaryConfig(auxConfigPath)
	if auxConfig == nil {
		panic("didn't find auxconfig")
	}
	config.AuxConfig = auxConfig
	fmt.Printf("config: %+v\n", config)
	runProtoform(config)
}

func runProtoform(config *model.ProtoformConfig) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(config.MasterURL, config.KubeConfigPath)
	//		kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	createPerceptorResources(config, clientset)
}
