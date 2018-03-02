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
	"log"

	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func readConfig(configPath string) *model.ProtoformConfig {
	log.Print("*************** [protoform] initializing viper ****************")
	viper.SetConfigName("protoform")
	viper.AddConfigPath(configPath)
	pc := &model.ProtoformConfig{}
	log.Print(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Print("Didn't see a config file!  Using reasonable defaults")
		return nil
	}
	viper.Unmarshal(pc)
	PrettyPrint(pc)
	log.Print("*************** [protoform] done reading in viper ****************")
	return pc
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func CreatePerceptorResources(namespace string, clientset *kubernetes.Clientset, serviceAccounts map[string]string, dryRun bool) {

	perceptor := model.NewPerceptorCore()
	podPerceiver := model.NewPodPerceiver(serviceAccounts["pod-perceiver"])
	imagePerceiver := model.NewImagePerceiver(serviceAccounts["image-perceiver"])
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
		if !dryRun {
			_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, service := range services {
		if dryRun {
			// service dont really need much debug...
			//PrettyPrint(svc)
		} else {
			_, err := clientset.Core().Services(namespace).Create(service)
			if err != nil {
				panic(err)
			}
		}
	}
}

func CreateConfigMapsFromInput(namespace string, clientset *kubernetes.Clientset, configMaps []*v1.ConfigMap, dryRun bool) {
	for _, configMap := range configMaps {
		log.Println("*********************************************")
		log.Println("Creating config maps:", configMap)
		if !dryRun {
			log.Println("creating config map.")
			clientset.Core().ConfigMaps(namespace).Create(configMap)
		} else {
			PrettyPrint(configMap)
		}
	}
}

// protoform is an experimental installer which bootstraps perceptor and the other
// autobots.

// main installs prime
func main() {
	//configPath := os.Args[1]
	runProtoform("/etc/protoform/")
}

func runProtoform(configPath string) {
	namespace := "bds-perceptor"
	var clientset *kubernetes.Clientset
	pc := readConfig(configPath)
	if pc == nil {
		log.Println("didn't find a config")
	}
	if !pc.DryRun {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
	}

	// TODO Viperize these env vars.
	if pc.ServiceAccounts == nil {
		log.Println("[viper] NO SERVICE ACCOUNTS FOUND.  USING DEFAULTS: MAKE SURE THESE EXIST!")

		svcAccounts := map[string]string{
			// WARNINNG: These service accounts need to exist !
			"pod-perceiver":          "openshift-perceiver",
			"image-perceiver":        "openshift-perceiver",
			"perceptor-image-facade": "perceptor-scanner-sa",
		}
		// TODO programatically validate rather then sanity check.
		PrettyPrint(svcAccounts)
		pc.ServiceAccounts = svcAccounts
	}

	log.Println("Creating config maps : Dry Run ")

	CreateConfigMapsFromInput(namespace, clientset, pc.ToConfigMap(), pc.DryRun)
	CreatePerceptorResources(namespace, clientset, pc.ServiceAccounts, pc.DryRun)
}
