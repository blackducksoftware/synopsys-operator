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
	"flag"
	"path/filepath"

	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	runHubInstaller()
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

func runHubInstaller() {
	log.Infof("Started Hub installer")
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Infof("unable to get in cluster config due to %v", err)
		log.Infof("trying to use local config")
		config, err = newKubeClientFromOutsideCluster()
		if err != nil {
			log.Errorf("unable to retrive the local config due to %v", err)
			log.Panicf("failed to find a valid cluster config")
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("unable to get the kubernetes client due to %v", err)
	}

	hubCreator := hub.NewHubInstaller(config, client)

	createHub := &hub.Hub{
		Namespace:      "blackduck-hub",
		DockerRegistry: "docker.io",
		DockerRepo:     "blackducksoftware",
		HubVersion:     "4.7.1",
		Flavor:         "small",
		AdminPassword:  "blackduck",
		UserPassword:   "blackduck",
	}

	hubCreator.CreateHub(createHub)
	return
}
