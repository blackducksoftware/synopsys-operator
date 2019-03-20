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

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// UpdateSynopsysOperator updates the Synopsys-Operator's kubernetes componenets and changes
// all CRDs to versions that the Operator can use
func UpdateSynopsysOperator(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newSOperatorSpec SpecConfig, blackduckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset, alertClient *alertclientset.Clientset) error {
	log.Debugf("Getting CRDs to update to Versions the new Operator can handle")
	currImage, err := GetOperatorImage(kubeClient, namespace)
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}
	// Get CRD Version Data
	newOperatorVersion := strings.Split(newSOperatorSpec.SynopsysOperatorImage, ":")[1]
	currOperatorVersion := strings.Split(currImage, ":")[1]
	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorVersion)
	currCrdData := SOperatorCRDVersionMap.GetCRDVersions(currOperatorVersion)
	// Get CRDs that need to be updated (specs have new version set)
	var oldBlackducks = []blackduckv1.Blackduck{}
	if newCrdData.Blackduck.APIVersion != currCrdData.Blackduck.APIVersion {
		oldBlackducks, err = GetBlackduckVersionsToRemove(blackduckClient, newCrdData.Blackduck.APIVersion, currCrdData.Blackduck.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		operatorutil.RunKubeCmd("delete", "crd", currCrdData.Blackduck.CRDName)
	}
	var oldOpsSights = []opssightv1.OpsSight{}
	if newCrdData.OpsSight.APIVersion != currCrdData.OpsSight.APIVersion {
		oldOpsSights, err = GetOpsSightVersionsToRemove(opssightClient, newCrdData.OpsSight.APIVersion, currCrdData.OpsSight.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		operatorutil.RunKubeCmd("delete", "crd", currCrdData.OpsSight.CRDName)
	}
	var oldAlerts = []alertv1.Alert{}
	if newCrdData.Alert.APIVersion != currCrdData.Alert.APIVersion {
		oldAlerts, err = GetAlertVersionsToRemove(alertClient, newCrdData.Alert.APIVersion, currCrdData.Alert.CRDName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		operatorutil.RunKubeCmd("delete", "crd", currCrdData.Alert.CRDName)
	}

	// Update the Synopsys-Operator's Components
	log.Debugf("Updating Synopsys-Operator's Componenets")
	err = UpdateSOperatorComponents(restconfig, kubeClient, namespace, newSOperatorSpec)
	if err != nil {
		return err
	}

	// Update the CRDs in the cluster with the new versions
	log.Debugf("Updating CRDs to Versions the new Operator can handle ")
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

// UpdateSOperatorComponentsByFlags updates kubernete's resources for the Synopsys-Operator by checking
// what flags were changed and updating the respective components
func UpdateSOperatorComponents(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newSOperatorSpec SpecConfig) error {
	newSOperatorComponents, err := newSOperatorSpec.GetComponents()
	if err != nil {
		return fmt.Errorf("Error creating new SOperator Components: %s", err)
	}
	var isConfigMapUpdated bool
	var isSecretUpdated bool
	// Update the Secret if the type or password changed
	isSecretUpdated, err = crdupdater.UpdateSecret(kubeClient, namespace, "blackduck-secret", newSOperatorComponents.Secrets[0])
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	// Update the Replication Controller if the image or reg key changed
	operatorUpdater := crdupdater.NewUpdater()
	replicationControllerUpdater, err := crdupdater.NewReplicationController(restconfig, kubeClient, newSOperatorComponents.ReplicationControllers, namespace, "app=opssight", isConfigMapUpdated || isSecretUpdated)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	operatorUpdater.AddUpdater(replicationControllerUpdater)
	err = operatorUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

// UpdatePrometheusByFlags updates kubernete's resources for Prometheus by checking
// what flags were changed and updating the respective components
func UpdatePrometheus(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newPrometheusSpecConfig PrometheusSpecConfig) error {
	// Get Components of New Prometheus
	newPrometheusComponents, err := newPrometheusSpecConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	prometheusUpdater := crdupdater.NewUpdater()
	deploymentUpdater, err := crdupdater.NewDeployment(restconfig, kubeClient, newPrometheusComponents.Deployments, namespace, "app=prometheus", false)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	prometheusUpdater.AddUpdater(deploymentUpdater)
	err = prometheusUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
