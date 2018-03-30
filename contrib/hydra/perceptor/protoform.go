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

	"k8s.io/api/core/v1"
	// v1beta1 "k8s.io/api/extensions/v1beta1"

	perceptor "github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/standardperceptor"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func createPerceptorResources(config *perceptor.Config, clientset *kubernetes.Clientset) {
	var configMaps []*v1.ConfigMap
	var services []*v1.Service
	var secrets []*v1.Secret
	var replicationControllers []*v1.ReplicationController
	if config.AuxConfig.IsOpenshift {
		os := perceptor.NewOpenshift(config)
		configMaps = os.ConfigMaps
		services = os.Services
		secrets = os.Secrets
		replicationControllers = os.ReplicationControllers
	} else {
		kube := perceptor.NewKube(config)
		configMaps = kube.ConfigMaps
		services = kube.Services
		secrets = kube.Secrets
		replicationControllers = kube.ReplicationControllers
	}

	namespace := config.AuxConfig.Namespace

	for _, configMap := range configMaps {
		PrettyPrint(configMap)
		_, err := clientset.Core().ConfigMaps(namespace).Create(configMap)
		if err != nil {
			panic(err)
		}
	}
	for _, secret := range secrets {
		PrettyPrint(secret)
		_, err := clientset.Core().Secrets(namespace).Create(secret)
		if err != nil {
			panic(err)
		}
	}
	for _, service := range services {
		PrettyPrint(service)
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

	// for _, dep := range deployments {
	// 	PrettyPrint(dep)
	// 	_, err := clientset.Extensions().Deployments(namespace).Create(dep)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
}

func main() {
	configPath := os.Args[1]
	auxConfigPath := os.Args[2]
	config := perceptor.ReadConfig(configPath)
	if config == nil {
		panic("didn't find config")
	}
	auxConfig := perceptor.ReadAuxiliaryConfig(auxConfigPath)
	if auxConfig == nil {
		panic("didn't find auxconfig")
	}
	config.AuxConfig = auxConfig
	fmt.Printf("config: %+v\n", config)
	runProtoform(config)
}

func runProtoform(config *perceptor.Config) {
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
