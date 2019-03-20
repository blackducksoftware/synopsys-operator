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

	"k8s.io/client-go/kubernetes"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// GetBlackduckVersionsToRemove finds all Blackducks with a different version, returns their specs with the new version
func GetBlackduckVersionsToRemove(blackduckClient *blackduckclientset.Clientset, newVersion string, oldCRDName string) ([]blackduckv1.Blackduck, error) {
	log.Debugf("Collecting all Blackducks of version: %s", newVersion)
	currCRDs, err := operatorutil.GetBlackducks(blackduckClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []blackduckv1.Blackduck{}
	for _, crd := range currCRDs.Items {
		log.Debugf("Found Blackduck version '%s': %s", crd.TypeMeta.APIVersion, crd.Name)
		if crd.TypeMeta.APIVersion != newVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// GetOpsSightVersionsToRemove finds all OpsSights with a different version, returns their specs with the new version
func GetOpsSightVersionsToRemove(opssightClient *opssightclientset.Clientset, newVersion string, oldCRDName string) ([]opssightv1.OpsSight, error) {
	log.Debugf("Collecting all OpsSights of version: %s", newVersion)
	currCRDs, err := operatorutil.GetOpsSights(opssightClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []opssightv1.OpsSight{}
	for _, crd := range currCRDs.Items {
		log.Debugf("Found OpsSight version '%s': %s", crd.TypeMeta.APIVersion, crd.Name)
		if crd.TypeMeta.APIVersion != newVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// GetAlertVersionsToRemove finds all Alerts with a different version, returns their specs with the new version
func GetAlertVersionsToRemove(alertClient *alertclientset.Clientset, newVersion string, oldCRDName string) ([]alertv1.Alert, error) {
	log.Debugf("Collecting all Alerts of version: %s", newVersion)
	currCRDs, err := operatorutil.GetAlerts(alertClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []alertv1.Alert{}
	for _, crd := range currCRDs.Items {
		log.Debugf("Found Alert version '%s': %s", crd.TypeMeta.APIVersion, crd.Name)
		if crd.TypeMeta.APIVersion != newVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// GetOperatorNamespace returns the namespace of the Synopsys-Operator by
// looking at its cluster role binding
func GetOperatorNamespace() (string, error) {
	namespace, err := operatorutil.RunKubeCmd("get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
	if err != nil {
		return "", fmt.Errorf("%s", namespace)
	}
	return namespace, nil
}

// GetOperatorImage returns the image for the synopsys-operator from
// the cluster
func GetOperatorImage(kubeClient *kubernetes.Clientset, namespace string) (string, error) {
	currPod, err := operatorutil.GetPod(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return "", fmt.Errorf("Failed to get Synopsys-Operator Pod: %s", err)
	}
	var currImage string
	for _, container := range currPod.Spec.Containers {
		if container.Name != "synopsys-operator" {
			continue
		}
		currImage = container.Image
	}
	return currImage, nil
}

// GetCurrentComponentsSpec returns a spec that respesents the current Synopsys-Operator in
// the cluster
func GetCurrentComponentsSpec(kubeClient *kubernetes.Clientset, namespace string) (*SpecConfig, error) {
	sOperatorSpec := SpecConfig{}
	// Set the Namespace
	sOperatorSpec.Namespace = namespace
	// Set the image and reg key
	currPod, err := operatorutil.GetPod(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return nil, fmt.Errorf("Failed to get Synopsys-Operator Pod: %s", err)
	}
	for _, container := range currPod.Spec.Containers {
		if container.Name != "synopsys-operator" {
			continue
		}
		sOperatorSpec.SynopsysOperatorImage = container.Image
		for _, env := range container.Env {
			if env.Name != "REGISTRATION_KEY" {
				continue
			}
			sOperatorSpec.BlackduckRegistrationKey = env.Value
		}
	}
	// Set the secretType and secret data
	currSecret, err := operatorutil.GetSecret(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return nil, fmt.Errorf("Failed to get Synopsys-Operator secret: %s", err)
	}
	currKubeSecret := currSecret.Type
	currHorizonSecretType, err := operatorutil.KubeSecretTypeToHorizon(currKubeSecret)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Current Synopsys-Operator Spec: %s", err)
	}
	sOperatorSpec.SecretType = currHorizonSecretType
	currKubeSecretData := currSecret.Data
	sOperatorSpec.SecretAdminPassword = string(currKubeSecretData["ADMIN_PASSWORD"])
	sOperatorSpec.SecretPostgresPassword = string(currKubeSecretData["POSTGRES_PASSWORD"])
	sOperatorSpec.SecretUserPassword = string(currKubeSecretData["USER_PASSWORD"])
	sOperatorSpec.SecretBlackduckPassword = string(currKubeSecretData["HUB_PASSWORD"])

	return &sOperatorSpec, nil
}
