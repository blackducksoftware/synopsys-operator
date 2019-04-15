/*
Copyright (C) 2019 Synopsys, Inc.

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

package synopsysctl

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// These vars set by setResourceClients() in root command's init()
var restconfig *rest.Config
var kubeClient *kubernetes.Clientset
var blackduckClient *blackduckclientset.Clientset
var opssightClient *opssightclientset.Clientset
var alertClient *alertclientset.Clientset

// These vars used by KubeCmd
var openshift bool
var kube bool

// setResourceClients sets the global variables for the kuberentes rest config
// and the resource clients
func setResourceClients() {
	var err error
	restconfig, err = protoform.GetKubeConfig(kubeconfig)
	if err != nil {
		log.Errorf("error getting Kube Rest Config: %s", err)
	}
	kubeClient, err = getKubeClient(restconfig)
	if err != nil {
		log.Errorf("error getting Kube Client: %s", err)
	}
	blackduckClient, err = blackduckclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Blackduck Clientset: %s", err)
	}
	opssightClient, err = opssightclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the OpsSight Clientset: %s", err)
	}
	alertClient, err = alertclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Alert Clientset: %s", err)
	}
	kube, openshift = operatorutil.DetermineClusterClients(restconfig)
}

// getKubeClient gets the kubernetes client
func getKubeClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// DeployCRDNamespace creates an empty Horizon namespace
func DeployCRDNamespace(restconfig *rest.Config, namespace string) error {
	namespaceDeployer, err := deployer.NewDeployer(restconfig)
	ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
		Name:      namespace,
		Namespace: namespace,
	})
	namespaceDeployer.AddNamespace(ns)
	err = namespaceDeployer.Run()
	if err != nil {
		return fmt.Errorf("error in creating the namespace due to %+v", err)
	}
	return nil
}
