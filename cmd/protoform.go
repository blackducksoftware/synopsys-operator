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
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func readConfig(configPath string) *model.ProtoformConfig {
	viper.SetConfigFile(configPath)
	pc := &model.ProtoformConfig{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(pc)
	PrettyPrint(pc)
	return pc
}

func readAuxiliaryConfig(auxConfigPath string) *model.AuxiliaryConfig {
	viper.SetConfigFile(auxConfigPath)
	aux := &model.AuxiliaryConfig{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(aux)
	PrettyPrint(aux)
	return aux
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func CreatePerceptorResources(namespace string, clientset *kubernetes.Clientset, serviceAccounts map[string]string) {

	imagePerceiverReplicaCount := int32(0)

	perceptor := model.NewPerceptorCore()
	podPerceiver := model.NewPodPerceiver(serviceAccounts["pod-perceiver"])
	imagePerceiver := model.NewImagePerceiver(imagePerceiverReplicaCount, serviceAccounts["image-perceiver"])
	perceptorScanner := model.NewPerceptorScanner(serviceAccounts["perceptor-image-facade"])

	replicationControllers := []*v1.ReplicationController{
		perceptor.ReplicationController(),
		podPerceiver.ReplicationController(),
		imagePerceiver.ReplicationController(),
		perceptorScanner.ReplicationController()}
	services := []*v1.Service{
		perceptor.Service(),
		podPerceiver.Service(),
		imagePerceiver.Service(),
		perceptorScanner.ScannerService(),
		perceptorScanner.ImageFacadeService()}

	for _, rc := range replicationControllers {
		PrettyPrint(rc)
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
}

func CreateConfigMapsFromInput(namespace string, clientset *kubernetes.Clientset, configMaps []*v1.ConfigMap) {
	for _, configMap := range configMaps {
		clientset.Core().ConfigMaps(namespace).Create(configMap)
	}
}

func main() {
	configPath := os.Args[1]
	auxConfigPath := os.Args[2]
	config := readConfig(configPath)
	if config == nil {
		panic("didn't find config")
	}
	auxConfig := readAuxiliaryConfig(auxConfigPath)
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

	// TODO do something intelligent with service account names -- inject from install.sh or something
	serviceAccounts := map[string]string{
		// WARNINNG: These service accounts need to exist !
		"pod-perceiver":          "openshift-perceiver",
		"image-perceiver":        "openshift-perceiver",
		"perceptor-image-facade": "perceptor-scanner-sa",
	}

	CreateConfigMapsFromInput(config.AuxConfig.Namespace, clientset, config.ToConfigMaps())
	CreatePerceptorResources(config.AuxConfig.Namespace, clientset, serviceAccounts)
}
