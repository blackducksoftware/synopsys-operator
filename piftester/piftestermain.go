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
	"fmt"
	"os"

	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/pif"
	"k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CreatePerceptorResources(config *pif.PifConfig, clientset *kubernetes.Clientset, serviceAccounts map[string]string) {
	imagePerceiverReplicaCount := int32(1)

	podPerceiver := model.NewPodPerceiver(serviceAccounts["pod-perceiver"])
	podPerceiver.Config = config.PodPerceiverConfig()

	imagePerceiver := model.NewImagePerceiver(imagePerceiverReplicaCount, serviceAccounts["image-perceiver"])
	imagePerceiver.Config = config.ImagePerceiverConfig()

	pifTester := pif.NewPifTester()
	pifTester.Config = config.PifTesterConfig()

	perceptorImagefacade := model.NewPerceptorImagefacade(serviceAccounts["perceptor-image-facade"])
	perceptorImagefacade.Config = config.PerceptorImagefacadeConfig()
	perceptorImagefacade.PodName = "perceptor-imagefacade"

	replicationControllers := []*v1.ReplicationController{
		podPerceiver.ReplicationController(),
		imagePerceiver.ReplicationController(),
		pifTester.ReplicationController(),
		perceptorImagefacade.ReplicationController(),
	}
	services := []*v1.Service{
		podPerceiver.Service(),
		imagePerceiver.Service(),
		pifTester.Service(),
		perceptorImagefacade.Service(),
	}
	configMaps := []*v1.ConfigMap{
		podPerceiver.ConfigMap(),
		imagePerceiver.ConfigMap(),
		pifTester.ConfigMap(),
		perceptorImagefacade.ConfigMap(),
	}

	namespace := config.AuxConfig.Namespace
	for _, rc := range replicationControllers {
		_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
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
	for _, configMap := range configMaps {
		_, err := clientset.Core().ConfigMaps(namespace).Create(configMap)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	configPath := os.Args[1]
	auxConfigPath := os.Args[2]
	config := pif.ReadPifConfig(configPath)
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

func runProtoform(config *pif.PifConfig) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(config.MasterURL, config.KubeConfigPath)
	//		kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	// TODO do something intelligent with service account names -- inject from install.sh or something
	serviceAccounts := map[string]string{
		// WARNINNG: These service accounts need to exist !
		"pod-perceiver":          "openshift-perceiver",
		"image-perceiver":        "openshift-perceiver",
		"perceptor-image-facade": "perceptor-scanner-sa",
	}

	CreatePerceptorResources(config, clientset, serviceAccounts)
}
