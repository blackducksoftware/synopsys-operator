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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// GetUpdatedBlackduckCRDs finds all Blackducks with different versions and returns their CRDs
// with the new version
func GetUpdatedBlackduckCRDs(blackduckClient *blackduckclientset.Clientset, newVersion string) ([]blackduckv1.Blackduck, error) {
	currCRDs, err := operatorutil.GetBlackducks(blackduckClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []blackduckv1.Blackduck{}
	for _, crd := range currCRDs.Items {
		if newVersion != crd.TypeMeta.APIVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// GetUpdatedOpsSightCRDs finds all OpsSights with different versions and returns their CRDs
// with the new version
func GetUpdatedOpsSightCRDs(opssightClient *opssightclientset.Clientset, newVersion string) ([]opssightv1.OpsSight, error) {
	currCRDs, err := operatorutil.GetOpsSights(opssightClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []opssightv1.OpsSight{}
	for _, crd := range currCRDs.Items {
		if newVersion != crd.TypeMeta.APIVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// GetUpdatedAlertCRDs finds all Alerts with different versions and returns their CRDs
// with the new version
func GetUpdatedAlertCRDs(alertClient *alertclientset.Clientset, newVersion string) ([]alertv1.Alert, error) {
	curCRDs, err := operatorutil.GetAlerts(alertClient)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	newCRDs := []alertv1.Alert{}
	for _, crd := range curCRDs.Items {
		if newVersion != crd.TypeMeta.APIVersion {
			crd.TypeMeta.APIVersion = newVersion
			newCRDs = append(newCRDs, crd)
		}
	}
	return newCRDs, nil
}

// TODO
func GetOperatorSpecConfig(kubeClient *kubernetes.Clientset, namespace string) {
	namespace := GetOperatorNamespace()
	image := GetOperatorImage(kubeClient, namespace)
	newSOperatorSpec := soperator.SpecConfig{
		Namespace:                namespace,
		SynopsysOperatorImage:    image,
		BlackduckRegistrationKey: deployBlackduckRegistrationKey,
		SecretType:               secretType,
		SecretAdminPassword:      deploySecretAdminPassword,
		SecretPostgresPassword:   deploySecretPostgresPassword,
		SecretUserPassword:       deploySecretUserPassword,
		SecretBlackduckPassword:  deploySecretBlackduckPassword,
	}

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
		if container.Name == "synopsys-operator" {
			continue
		}
		currImage = container.Image
	}
	return currImage, nil
}

// KubeSecretTypeToHorizon converts a kubernetes SecretType to Horizon's SecretType
func KubeSecretTypeToHorizon(secretType corev1.SecretType) (horizonapi.SecretType, error) {
	switch secretType {
	case corev1.SecretTypeOpaque:
		return horizonapi.SecretTypeOpaque, nil
	case corev1.SecretTypeServiceAccountToken:
		return horizonapi.SecretTypeServiceAccountToken, nil
	case corev1.SecretTypeDockercfg:
		return horizonapi.SecretTypeDockercfg, nil
	case corev1.SecretTypeDockerConfigJson:
		return horizonapi.SecretTypeDockerConfigJSON, nil
	case corev1.SecretTypeBasicAuth:
		return horizonapi.SecretTypeBasicAuth, nil
	case corev1.SecretTypeSSHAuth:
		return horizonapi.SecretTypeSSHAuth, nil
	case corev1.SecretTypeTLS:
		return horizonapi.SecretTypeTLS, nil
	default:
		return horizonapi.SecretTypeOpaque, fmt.Errorf("Invalid Secret Type: %+v", secretType)
	}
}
