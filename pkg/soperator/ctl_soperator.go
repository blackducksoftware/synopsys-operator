/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package soperator

import (
	"fmt"
	"strings"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// UpdateSynopsysOperator updates the Synopsys-Operator's kubernetes componenets and changes
// all CRDs to versions that the Operator can use
func UpdateSynopsysOperator(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newSOperatorSpec *SpecConfig, blackduckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset, alertClient *alertclientset.Clientset) error {
	currImage, err := GetOperatorImage(kubeClient, namespace)
	if err != nil {
		log.Errorf("Failed to Update the Synopsys Operator: %s", err)
		return nil
	}
	// Get CRD Version Data
	newOperatorVersion := strings.Split(newSOperatorSpec.SynopsysOperatorImage, ":")[1]
	currOperatorVersion := strings.Split(currImage, ":")[1]
	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorVersion)
	currCrdData := SOperatorCRDVersionMap.GetCRDVersions(currOperatorVersion)
	// Get CRDs that need to be updated (specs have new version set)
	log.Debugf("Getting CRDs that need new versions")
	var oldBlackducks = []blackduckv1.Blackduck{}
	kube, openshift := operatorutil.DetermineClusterClients(restconfig)
	if newCrdData.Blackduck.APIVersion != currCrdData.Blackduck.APIVersion {
		oldBlackducks, err = GetBlackduckVersionsToRemove(blackduckClient, newCrdData.Blackduck.APIVersion, currCrdData.Blackduck.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.Blackduck.CRDName)
		if err != nil {
			return fmt.Errorf("%s", out)
		}
		log.Debugf("Updating %d Black Ducks", len(oldBlackducks))
	}
	var oldOpsSights = []opssightv1.OpsSight{}
	if newCrdData.OpsSight.APIVersion != currCrdData.OpsSight.APIVersion {
		oldOpsSights, err = GetOpsSightVersionsToRemove(opssightClient, newCrdData.OpsSight.APIVersion, currCrdData.OpsSight.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.OpsSight.CRDName)
		if err != nil {
			return fmt.Errorf("%s", out)
		}
		log.Debugf("Updating %d OpsSights", len(oldOpsSights))
	}
	var oldAlerts = []alertv1.Alert{}
	if newCrdData.Alert.APIVersion != currCrdData.Alert.APIVersion {
		oldAlerts, err = GetAlertVersionsToRemove(alertClient, newCrdData.Alert.APIVersion, currCrdData.Alert.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.Alert.CRDName)
		if err != nil {
			return fmt.Errorf("%s", out)
		}
		log.Debugf("Updating %d Alerts", len(oldAlerts))
	}

	// Update the Synopsys-Operator's Components
	log.Debugf("Updating Synopsys-Operator's Components")
	err = UpdateSOperatorComponents(restconfig, kubeClient, namespace, newSOperatorSpec)
	if err != nil {
		return fmt.Errorf("Failed to Update Synopsys-Operator: %s", err)
	}

	// Update the CRDs in the cluster with the new versions
	log.Debugf("Updating CRDs to new Versions")
	err = operatorutil.UpdateBlackducks(blackduckClient, oldBlackducks)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	err = operatorutil.UpdateOpsSights(opssightClient, oldOpsSights)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	err = operatorutil.UpdateAlerts(alertClient, oldAlerts)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return nil
}

// UpdateSOperatorComponents updates kubernetes resources for the Synopsys-Operator
func UpdateSOperatorComponents(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newSOperatorSpecConfig *SpecConfig) error {
	sOperatorComponents, err := newSOperatorSpecConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("Failed to Update Operator Components: %s", err)
	}
	sOperatorCommonConfig := crdupdater.NewCRUDComponents(restconfig, kubeClient, false, namespace, sOperatorComponents, "app=synopsys-operator")
	errs := sOperatorCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("Failed to Update Operator Components: %+v", errs)
	}

	return nil
}

// UpdatePrometheus updates kubernetes resources for Prometheus
func UpdatePrometheus(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newPrometheusSpecConfig *PrometheusSpecConfig) error {
	prometheusComponents, err := newPrometheusSpecConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("Failed to Update Prometheus Components: %s", err)
	}
	prometheusCommonConfig := crdupdater.NewCRUDComponents(restconfig, kubeClient, false, namespace, prometheusComponents, "app=prometheus")
	errs := prometheusCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("Failed to Update Operator Components: %+v", errs)
	}
	return nil
}