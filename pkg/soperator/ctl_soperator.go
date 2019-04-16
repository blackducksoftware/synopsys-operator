/*
Copyright (C) 2019 Synopsys, Inc.

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
	"time"

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
func (specConfig *SpecConfig) UpdateSynopsysOperator(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, blackduckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset, alertClient *alertclientset.Clientset) error {
	currImage, err := GetOperatorImage(kubeClient, namespace)
	if err != nil {
		log.Errorf("Failed to Update the Synopsys Operator: %s", err)
		return nil
	}

	// Get CRD Version Data
	newOperatorVersion := strings.Split(specConfig.Image, ":")[1]
	currOperatorVersion := strings.Split(currImage, ":")[1]
	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorVersion)
	currCrdData := SOperatorCRDVersionMap.GetCRDVersions(currOperatorVersion)

	// Get CRDs that need to be updated (specs have new version set)
	log.Debugf("Getting CRDs that need new versions")
	var oldBlackducks = []blackduckv1.Blackduck{}
	kube, openshift := operatorutil.DetermineClusterClients(restconfig)
	if newCrdData.Blackduck.APIVersion != currCrdData.Blackduck.APIVersion {
		oldBlackducks, err = GetBlackduckVersionsToRemove(blackduckClient, newCrdData.Blackduck.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get Blackduck's to update: %s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.Blackduck.CRDName)
		if err != nil {
			return fmt.Errorf("failed to delete the crd %s: %s", currCrdData.Blackduck.CRDName, out)
		}
		log.Debugf("Updating %d Black Ducks", len(oldBlackducks))
	}
	var oldOpsSights = []opssightv1.OpsSight{}
	if newCrdData.OpsSight.APIVersion != currCrdData.OpsSight.APIVersion {
		oldOpsSights, err = GetOpsSightVersionsToRemove(opssightClient, newCrdData.OpsSight.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get OpsSights to update: %s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.OpsSight.CRDName)
		if err != nil {
			return fmt.Errorf("failed to delete the crd %s: %s", currCrdData.OpsSight.CRDName, out)
		}
		log.Debugf("Updating %d OpsSights", len(oldOpsSights))
	}
	var oldAlerts = []alertv1.Alert{}
	if newCrdData.Alert.APIVersion != currCrdData.Alert.APIVersion {
		oldAlerts, err = GetAlertVersionsToRemove(alertClient, newCrdData.Alert.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get Alerts to update%s", err)
		}
		out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, "delete", "crd", currCrdData.Alert.CRDName)
		if err != nil {
			return fmt.Errorf("failed to delete the crd %s: %s", currCrdData.Alert.CRDName, out)
		}
		log.Debugf("Updating %d Alerts", len(oldAlerts))
	}

	// Update the Synopsys-Operator's Components
	log.Debugf("Updating Synopsys-Operator's Components")
	err = specConfig.UpdateSOperatorComponents()
	if err != nil {
		return fmt.Errorf("failed to update Synopsys-Operator components: %s", err)
	}

	// Update the CRDs in the cluster with the new versions
	// loop to wait for kuberentes to register new CRDs
	log.Debugf("Updating CRDs to new Versions")
	for i := 1; i <= 10; i++ {
		if err = operatorutil.UpdateBlackducks(blackduckClient, oldBlackducks); err == nil {
			break
		}
		if i >= 10 {
			return fmt.Errorf("failed to update Black Ducks: %s", err)
		}
		time.Sleep(1 * time.Second)
	}
	for i := 1; i <= 10; i++ {
		if err = operatorutil.UpdateOpsSights(opssightClient, oldOpsSights); err == nil {
			break
		}
		if i >= 10 {
			return fmt.Errorf("failed to update OpsSights: %s", err)
		}
		time.Sleep(1 * time.Second)
	}
	for i := 1; i <= 10; i++ {
		if err = operatorutil.UpdateAlerts(alertClient, oldAlerts); err == nil {
			break
		}
		if i >= 10 {
			return fmt.Errorf("failed to update Alerts: %s", err)
		}
		log.Debugf("Attempt %d to update Alerts\n", i)
		time.Sleep(1 * time.Second)
	}

	return nil
}

// UpdateSOperatorComponents updates kubernetes resources for the Synopsys-Operator
func (specConfig *SpecConfig) UpdateSOperatorComponents() error {
	sOperatorComponents, err := specConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("failed to get Synopsys-Operator components: %s", err)
	}
	sOperatorCommonConfig := crdupdater.NewCRUDComponents(specConfig.RestConfig, specConfig.KubeClient, false, specConfig.Namespace, sOperatorComponents, "app=synopsys-operator")
	errs := sOperatorCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("failed to update Synopsys-Operator components: %+v", errs)
	}

	return nil
}

// UpdatePrometheus updates kubernetes resources for Prometheus
func (specConfig *PrometheusSpecConfig) UpdatePrometheus() error {
	prometheusComponents, err := specConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("failed to get Prometheus components: %s", err)
	}
	prometheusCommonConfig := crdupdater.NewCRUDComponents(specConfig.RestConfig, specConfig.KubeClient, false, specConfig.Namespace, prometheusComponents, "app=prometheus")
	errs := prometheusCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("failed to update Prometheus components: %+v", errs)
	}
	return nil
}
