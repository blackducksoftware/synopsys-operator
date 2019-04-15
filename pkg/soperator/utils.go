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

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetBlackduckVersionsToRemove finds all Blackducks with a different version, returns their specs with the new version
func GetBlackduckVersionsToRemove(blackduckClient *blackduckclientset.Clientset, newVersion string) ([]blackduckv1.Blackduck, error) {
	log.Debugf("Collecting all Blackducks that are not version: %s", newVersion)
	currBlackDucks, err := operatorutil.GetBlackducks(blackduckClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get BlackDucks: %s", err)
	}
	newBlackDucks := []blackduckv1.Blackduck{}
	for _, blackDuck := range currBlackDucks.Items {
		log.Debugf("Found Blackduck version '%s': %s", blackDuck.TypeMeta.APIVersion, blackDuck.Name)
		if blackDuck.TypeMeta.APIVersion != newVersion {
			blackDuck.TypeMeta.APIVersion = newVersion
			newBlackDucks = append(newBlackDucks, blackDuck)
		}
	}
	return newBlackDucks, nil
}

// GetOpsSightVersionsToRemove finds all OpsSights with a different version, returns their specs with the new version
func GetOpsSightVersionsToRemove(opssightClient *opssightclientset.Clientset, newVersion string) ([]opssightv1.OpsSight, error) {
	log.Debugf("Collecting all OpsSights that are not version: %s", newVersion)
	currOpsSights, err := operatorutil.GetOpsSights(opssightClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpsSights: %s", err)
	}
	newOpsSights := []opssightv1.OpsSight{}
	for _, opsSight := range currOpsSights.Items {
		log.Debugf("Found OpsSight version '%s': %s", opsSight.TypeMeta.APIVersion, opsSight.Name)
		if opsSight.TypeMeta.APIVersion != newVersion {
			opsSight.TypeMeta.APIVersion = newVersion
			newOpsSights = append(newOpsSights, opsSight)
		}
	}
	return newOpsSights, nil
}

// GetAlertVersionsToRemove finds all Alerts with a different version, returns their specs with the new version
func GetAlertVersionsToRemove(alertClient *alertclientset.Clientset, newVersion string) ([]alertv1.Alert, error) {
	log.Debugf("Collecting all Alerts that are not version: %s", newVersion)
	currAlerts, err := operatorutil.GetAlerts(alertClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get Alerts: %s", err)
	}
	newAlerts := []alertv1.Alert{}
	for _, alert := range currAlerts.Items {
		log.Debugf("Found Alert version '%s': %s", alert.TypeMeta.APIVersion, alert.Name)
		if alert.TypeMeta.APIVersion != newVersion {
			alert.TypeMeta.APIVersion = newVersion
			newAlerts = append(newAlerts, alert)
		}
	}
	return newAlerts, nil
}

// GetOperatorNamespace returns the namespace of the Synopsys-Operator by
// looking at its cluster role binding
func GetOperatorNamespace(restConfig *rest.Config) (string, error) {
	kube, openshift := operatorutil.DetermineClusterClients(restConfig)
	namespace, err := operatorutil.RunKubeCmd(restConfig, kube, openshift, "get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
	if err != nil {
		return "", fmt.Errorf("failed to get Synopsys-Operator Namespace: %s", err)
	}
	return strings.Trim(namespace, "'"), nil
}

// GetOperatorImage returns the image for the synopsys-operator from
// the cluster
func GetOperatorImage(kubeClient *kubernetes.Clientset, namespace string) (string, error) {
	currCM, err := operatorutil.GetConfigMap(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return "", fmt.Errorf("failed to get Synopsys-Operator Image: %s", err)
	}
	return currCM.Data["image"], nil
}

// GetSpecConfigForCurrentComponents returns a spec that respesents the current Synopsys-Operator in the cluster
func GetSpecConfigForCurrentComponents(restConfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string) (*SpecConfig, error) {
	log.Debugf("creating new synopsys operator spec")
	sOperatorSpec := SpecConfig{}
	// Set the Namespace
	sOperatorSpec.Namespace = namespace
	// Set the image
	currCM, err := operatorutil.GetConfigMap(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return nil, fmt.Errorf("failed to get Synopsys-Operator ConfigMap: %s", err)
	}
	sOperatorSpec.Image = currCM.Data["image"]
	log.Debugf("got current synopsys operator image from cluster: %s", sOperatorSpec.Image)

	// Set the secretType and secret data
	currSecret, err := operatorutil.GetSecret(kubeClient, namespace, "blackduck-secret")
	if err != nil {
		return nil, fmt.Errorf("failed to get synopsys operator secret: %s", err)
	}
	currKubeSecret := currSecret.Type
	currHorizonSecretType, err := operatorutil.KubeSecretTypeToHorizon(currKubeSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create synopsys operator spec: %s", err)
	}
	sOperatorSpec.SecretType = currHorizonSecretType
	currKubeSecretData := currSecret.Data
	sOperatorSpec.AdminPassword = string(currKubeSecretData["ADMIN_PASSWORD"])
	sOperatorSpec.PostgresPassword = string(currKubeSecretData["POSTGRES_PASSWORD"])
	sOperatorSpec.UserPassword = string(currKubeSecretData["USER_PASSWORD"])
	sOperatorSpec.BlackduckPassword = string(currKubeSecretData["HUB_PASSWORD"])
	sealKey := string(currKubeSecretData["SEAL_KEY"])
	if len(sealKey) == 0 {
		sealKey, err = operatorutil.GetRandomString(32)
		if err != nil {
			log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
		}
	}
	sOperatorSpec.SealKey = sealKey
	sOperatorSpec.RestConfig = restConfig
	sOperatorSpec.KubeClient = kubeClient

	log.Debugf("got current synopsys operator secret data from Cluster")

	return &sOperatorSpec, nil
}

// GetSpecConfigForCurrentPrometheusComponents returns a spec that respesents the current prometheus in the cluster
func GetSpecConfigForCurrentPrometheusComponents(restConfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string) (*PrometheusSpecConfig, error) {
	log.Debugf("creating New Prometheus SpecConfig")
	prometheusSpec := PrometheusSpecConfig{}
	// Set Namespace
	prometheusSpec.Namespace = namespace
	// Set Image
	currCM, err := operatorutil.GetConfigMap(kubeClient, namespace, "prometheus")
	if err != nil {
		return nil, fmt.Errorf("Failed to get Prometheus ConfigMap: %s", err)
	}
	prometheusSpec.Image = currCM.Data["image"]
	prometheusSpec.RestConfig = restConfig
	prometheusSpec.KubeClient = kubeClient
	log.Debugf("added image %s to Prometheus SpecConfig", prometheusSpec.Image)

	return &prometheusSpec, nil

}
