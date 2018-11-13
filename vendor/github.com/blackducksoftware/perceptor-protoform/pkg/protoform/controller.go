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

package protoform

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/juju/errors"

	log "github.com/sirupsen/logrus"
)

// NewController will initialize the input config file, create the hub informers, initiantiate all rest api
func NewController(configPath string) (*Deployer, error) {
	config, err := GetConfig(configPath)
	if err != nil {
		return nil, errors.Annotate(err, "Failed to load configuration")
	}
	if config == nil {
		return nil, errors.Errorf("expected non-nil config, but got nil")
	}

	level, err := config.GetLogLevel()
	if err != nil {
		return nil, errors.Annotate(err, "unable to get log level")
	}
	log.SetLevel(level)

	log.Debugf("config: %+v", config)

	// creates the in-cluster config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("error getting in cluster config. Fallback to native config. Error message: %s\n", err)
		kubeConfig, err = newKubeClientFromOutsideCluster()
	}

	if err != nil {
		return nil, errors.Annotate(err, "unable to create config for both in-cluster and external to cluster")
	}

	kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create kubernetes clientset")
	}

	return NewDeployer(config, kubeConfig, kubeClientSet), nil
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
	return config, errors.Annotate(err, "error creating default client config")
}
