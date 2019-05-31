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
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// UpdateSynopsysOperator updates Synopsys Operator's kubernetes componenets and changes
// all CRDs to versions that the Operator can use
func (specConfig *SpecConfig) UpdateSynopsysOperator(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string,
	blackduckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset, alertClient *alertclientset.Clientset,
	oldOperatorSpec *SpecConfig) error {

	// Get CRD Version Data
	newOperatorVersion := strings.Split(specConfig.Image, ":")[1]
	oldOperatorVersion := strings.Split(oldOperatorSpec.Image, ":")[1]
	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorVersion)
	oldCrdData := SOperatorCRDVersionMap.GetCRDVersions(oldOperatorVersion)

	// Get CRDs that need to be updated (specs have new version set)
	log.Debugf("Getting CRDs that need new versions")
	var oldBlackducks = []blackduckv1.Blackduck{}

	apiExtensionClient, err := apiextensionsclient.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("error creating the api extension client due to %+v", err)
	}

	if newCrdData.Blackduck.APIVersion != oldCrdData.Blackduck.APIVersion {
		oldBlackducks, err = GetBlackduckVersionsToRemove(blackduckClient, newCrdData.Blackduck.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get Blackduck's to update: %s", err)
		}
		err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.Blackduck.CRDName)
		if err != nil {
			return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.Blackduck.CRDName, err)
		}
		log.Debugf("updating %d Black Ducks", len(oldBlackducks))
	}
	var oldOpsSights = []opssightv1.OpsSight{}
	if newCrdData.OpsSight.APIVersion != oldCrdData.OpsSight.APIVersion {
		oldOpsSights, err = GetOpsSightVersionsToRemove(opssightClient, newCrdData.OpsSight.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get OpsSights to update: %s", err)
		}
		err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.OpsSight.CRDName)
		if err != nil {
			return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.OpsSight.CRDName, err)
		}
		log.Debugf("updating %d OpsSights", len(oldOpsSights))
	}
	var oldAlerts = []alertv1.Alert{}
	if newCrdData.Alert.APIVersion != oldCrdData.Alert.APIVersion {
		oldAlerts, err = GetAlertVersionsToRemove(alertClient, newCrdData.Alert.APIVersion)
		if err != nil {
			return fmt.Errorf("failed to get Alerts to update%s", err)
		}
		err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.Alert.CRDName)
		if err != nil {
			return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.Alert.CRDName, err)
		}
		log.Debugf("updating %d Alerts", len(oldAlerts))
	}

	// Update Synopsys Operator's Components
	log.Debugf("updating Synopsys Operator's Components")
	err = specConfig.UpdateSOperatorComponents()
	if err != nil {
		return fmt.Errorf("failed to update Synopsys Operator components: %s", err)
	}

	// Update the CRDs in the cluster with the new versions
	// loop to wait for kuberentes to register new CRDs
	log.Debugf("updating CRDs to new Versions")
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
		log.Debugf("Attempt %d to update Alerts", i)
		time.Sleep(1 * time.Second)
	}

	return nil
}

// UpdateSOperatorComponents updates kubernetes resources for Synopsys Operator
func (specConfig *SpecConfig) UpdateSOperatorComponents() error {
	sOperatorComponents, err := specConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("failed to get Synopsys Operator components: %s", err)
	}
	sOperatorCommonConfig := crdupdater.NewCRUDComponents(specConfig.RestConfig, specConfig.KubeClient, false, false, specConfig.Namespace, sOperatorComponents, "app=synopsys-operator,component=operator")
	_, errs := sOperatorCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("failed to update Synopsys Operator components: %+v", errs)
	}

	return nil
}

// UpdatePrometheus updates kubernetes resources for Prometheus
func (specConfig *PrometheusSpecConfig) UpdatePrometheus() error {
	prometheusComponents, err := specConfig.GetComponents()
	if err != nil {
		return fmt.Errorf("failed to get Prometheus components: %s", err)
	}
	prometheusCommonConfig := crdupdater.NewCRUDComponents(specConfig.RestConfig, specConfig.KubeClient, false, false, specConfig.Namespace, prometheusComponents, "app=synopsys-operator,component=prometheus")
	_, errs := prometheusCommonConfig.CRUDComponents()
	if errs != nil {
		return fmt.Errorf("failed to update Prometheus components: %+v", errs)
	}
	return nil
}
