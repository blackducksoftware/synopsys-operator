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

package synopsysctl

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/deployer"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// getBlackduckFromCluster returns the CRD for a blackduck in namespace
func getBlackduckFromCluster(namespace string) (*blackduckv1.Blackduck, error) {
	blackduck, err := blackduckClient.SynopsysV1().Blackducks(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return blackduck, fmt.Errorf("error Editing Blackduck: %+v", err)
	}
	return blackduck, nil
}

// updateBlackduckInCluster updates the CRD for a blackduck
func updateBlackduckInCluster(namespace string, crd *blackduckv1.Blackduck) error {
	_, err := blackduckClient.SynopsysV1().Blackducks(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("error Editing Blackduck: %+v", err)
	}
	return nil
}

// getOpsSightFromCluster returns the CRD for an OpsSight in namespace
func getOpsSightFromCluster(namespace string) (*opssightv1.OpsSight, error) {
	opssight, err := opssightClient.SynopsysV1().OpsSights(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return opssight, fmt.Errorf("error Editing OpsSight: %+v", err)
	}
	return opssight, nil
}

// updateOpsSightInCluster updates the CRD for an OpsSight
func updateOpsSightInCluster(namespace string, crd *opssightv1.OpsSight) error {
	_, err := opssightClient.SynopsysV1().OpsSights(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("error Editing OpsSight: %+v", err)
	}
	return nil
}

// getAlertFromCluster returns the CRD for an Alert in namespace
func getAlertFromCluster(namespace string) (*alertv1.Alert, error) {
	alert, err := alertClient.SynopsysV1().Alerts(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return alert, fmt.Errorf("error Editing Alert: %+v", err)
	}
	return alert, nil
}

// updateAlertInCluster updates the CRD for an Alert
func updateAlertInCluster(namespace string, crd *alertv1.Alert) error {
	_, err := alertClient.SynopsysV1().Alerts(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("error Editing Alert: %+v", err)
	}
	return nil
}

// setResourceClients sets the global variables for the kuberentes rest config
// and the resource clients
func setResourceClients() {
	var err error
	restconfig, err = protoform.GetKubeConfig()
	if err != nil {
		log.Errorf("error getting Kube Rest Config: %s", err)
	}
	kubeClient, err = getKubeClient(restconfig)
	if err != nil {
		log.Errorf("error getting Kube Client: %s", err)
	}
	bClient, err := blackduckclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Blackduck Clientset: %s", err)
	}
	blackduckClient = bClient
	oClient, err := opssightclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the OpsSight Clientset: %s", err)
	}
	opssightClient = oClient
	aClient, err := alertclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Alert Clientset: %s", err)
	}
	alertClient = aClient
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
		return fmt.Errorf("error deploying namespace with Horizon : %s", err)
	}
	return nil
}
