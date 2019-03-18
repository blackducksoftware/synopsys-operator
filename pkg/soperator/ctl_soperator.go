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
	// Get CRDs that need to be updated
	log.Debugf("Getting CRDs to update to Versions the new Operator can handle")
	newOperatorVersion := strings.Split(newSOperatorSpec.SynopsysOperatorImage, ":")[1]
	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorVersion)
	oldBlackducks, err := GetUpdatedBlackduckCRDs(blackduckClient, newCrdData.Blackduck.APIVersion)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	oldOpsSights, err := GetUpdatedOpsSightCRDs(opssightClient, newCrdData.OpsSight.APIVersion)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	oldAlerts, err := GetUpdatedAlertCRDs(alertClient, newCrdData.Alert.APIVersion)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	// Delete the current CRD definitions from the cluster
	log.Debugf("Deleting the CRD definitions from cluster")
	currImage, err := GetOperatorImage(kubeClient, namespace)
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}
	currOperatorVersion := strings.Split(currImage, ":")[1]
	for _, crd := range SOperatorCRDVersionMap.GetIterableCRDData(currOperatorVersion) {
		operatorutil.RunKubeCmd("delete", "crd", crd.CRDName)
	}

	// Update the Synopsys-Operator's Components
	log.Debugf("Updating Synopsys-Operator's Componenets")
	err = UpdateSynopsysOperatorComponents(restconfig, kubeClient, namespace, newSOperatorSpec)
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

// UpdateSynopsysOperatorComponents updates kubernete's resources for the Synopsys-Operator by comparing
// it's current componenets with the new componenets
func UpdateSynopsysOperatorComponents(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newSOperatorSpec SpecConfig) error {
	// Get Components of New Synopsys-Operator
	newSOperatorComponents, err := newSOperatorSpec.GetComponents()
	if err != nil {
		return fmt.Errorf("Failed to Create New Components: %s", err)
	}

	// Update S-O ConfigMap if necessary
	isConfigMapUpdated, err := crdupdater.UpdateConfigMap(kubeClient, namespace, "synopsys-operator", newSOperatorComponents.ConfigMaps[0])
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	// Update S-O Secret if necessary
	isSecretUpdated, err := crdupdater.UpdateSecret(kubeClient, namespace, "blackduck-secret", newSOperatorComponents.Secrets[0])
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	operatorUpdater := crdupdater.NewUpdater()

	// Update S-O ReplicationController if necessary
	replicationControllerUpdater, err := crdupdater.NewReplicationController(restconfig, kubeClient, newSOperatorComponents.ReplicationControllers, namespace, "app=opssight", isConfigMapUpdated || isSecretUpdated)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	operatorUpdater.AddUpdater(replicationControllerUpdater)

	// Update S-O Service if necessary
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newSOperatorComponents.Services, namespace, "app=opssight")
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	operatorUpdater.AddUpdater(serviceUpdater)

	// Update S-O ServiceAccount if necessary

	// Update S-O ClusterRoleBinding if necessary
	clusterRoleBindingUpdater, err := crdupdater.NewClusterRoleBinding(restconfig, kubeClient, newSOperatorComponents.ClusterRoleBindings, namespace, "app=opssight")
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	operatorUpdater.AddUpdater(clusterRoleBindingUpdater)

	err = operatorUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

// UpdatePrometheus updates kubernete's resources for Prometheus by comparing
// it's current componenets with the new componenets
func UpdatePrometheus(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, newPrometheusSpecConfig PrometheusSpecConfig) error {
	// Get Components of New Prometheus
	newPrometheusComponents, err := newPrometheusSpecConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	prometheusUpdater := crdupdater.NewUpdater()

	// Update Prometheus ConfigMap
	_, err = crdupdater.UpdateConfigMap(kubeClient, namespace, "prometheus", newPrometheusComponents.ConfigMaps[0])
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	// Update Prometheus Deployment
	deploymentUpdater, err := crdupdater.NewDeployment(restconfig, kubeClient, newPrometheusComponents.Deployments, namespace, "app=prometheus", false)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	prometheusUpdater.AddUpdater(deploymentUpdater)

	// Update Prometheus Service
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newPrometheusComponents.Services, namespace, "app=prometheus")
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	prometheusUpdater.AddUpdater(serviceUpdater)

	err = prometheusUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
