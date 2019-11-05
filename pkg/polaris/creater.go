/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package polaris

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetComponents get all Polaris related components
func GetComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	components := make(map[string]runtime.Object)

	baseSecrets, err := GetPolarisBaseSecrets(polaris)
	if err != nil {
		return nil, err
	}
	dbComponents, err := GetPolarisDBComponents(baseURL, polaris)
	if err != nil {
		return nil, err
	}
	polarisComponents, err := GetPolarisComponents(baseURL, polaris)
	if err != nil {
		return nil, err
	}

	for k, v := range baseSecrets {
		components[k] = v
	}
	for k, v := range dbComponents {
		components[k] = v
	}
	for k, v := range polarisComponents {
		components[k] = v
	}

	if polaris.PolarisDBSpec.PostgresDetails.IsInternal {
		postgresComponents, err := GetPolarisPostgresComponents(baseURL, polaris)
		if err != nil {
			return nil, err
		}

		for k, v := range postgresComponents {
			components[k] = v
		}
	}

	if polaris.EnableReporting {
		reportingComponents, err := GetPolarisReportingComponents(baseURL, polaris)
		if err != nil {
			return nil, err
		}
		for k, v := range reportingComponents {
			components[k] = v
		}
	}

	provisionComponents, err := GetPolarisProvisionComponents(baseURL, polaris)
	if err != nil {
		return nil, err
	}
	for k, v := range provisionComponents {
		components[k] = v
	}

	return components, nil
}

// GetPolarisBaseSecrets get polaris base secrets
func GetPolarisBaseSecrets(polaris Polaris) (map[string]runtime.Object, error) {
	mapOfUniqueIDToBaseRuntimeObject := make(map[string]runtime.Object, 0)

	var plaformLicense *PlatformLicense
	if err := json.Unmarshal([]byte(polaris.Licenses.Polaris), &plaformLicense); err != nil {
		return nil, err
	}

	if strings.Compare(plaformLicense.License.IssuedTo, polaris.OrganizationDetails.OrganizationProvisionOrganizationName) != 0 {
		return nil, fmt.Errorf("the Polaris license is only valid for the following organization: %s", plaformLicense.License.IssuedTo)
	}

	// TODO store all the licenses inside a single secret
	// Coverity license
	mapOfUniqueIDToBaseRuntimeObject["Secret.coverity-license"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coverity-license",
			Namespace: polaris.Namespace,
			Labels: map[string]string{
				"environment": polaris.Namespace,
			},
		},
		Data: map[string][]byte{
			"license": []byte(polaris.Licenses.Coverity),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Org license
	mapOfUniqueIDToBaseRuntimeObject["Secret.organization-license"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "organization-license",
			Namespace: polaris.Namespace,
			Labels: map[string]string{
				"environment": polaris.Namespace,
			},
		},
		Data: map[string][]byte{
			"polaris": []byte(util.EncodeStringToBase64(polaris.Licenses.Polaris)),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Tool store sync service account
	mapOfUniqueIDToBaseRuntimeObject["Secret.tools-store-sync"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tools-store-sync",
			Namespace: polaris.Namespace,
			Labels: map[string]string{
				"environment": polaris.Namespace,
			},
		},
		Data: map[string][]byte{
			"credentials.json": []byte(polaris.GCPServiceAccount),
			"instance-name":    []byte("tools-store-sync"),
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Pull secret gcr-json-key
	type dockerAuthConfig struct {
		Username string
		Password string
		Email    string
	}

	dockerCfg := struct {
		Auths map[string]dockerAuthConfig `json:"auths"`
	}{
		Auths: map[string]dockerAuthConfig{
			"https://gcr.io": {
				Username: "_json_key",
				Password: polaris.GCPServiceAccount,
			},
		},
	}

	dockerCfgByte, err := json.Marshal(dockerCfg)
	if err != nil {
		return nil, err
	}

	mapOfUniqueIDToBaseRuntimeObject["Secret.gcr-json-key"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcr-json-key",
			Namespace: polaris.Namespace,
			Labels: map[string]string{
				"environment": polaris.Namespace,
			},
		},
		Data: map[string][]byte{
			".dockerconfigjson": dockerCfgByte,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	return mapOfUniqueIDToBaseRuntimeObject, nil
}

// GetPolarisReportingComponents get Polaris reporting components
func GetPolarisReportingComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := util.GetBaseYaml(baseURL, "polaris", polaris.Version, "reporting_base.yaml")
	if err != nil {
		return nil, err
	}

	return fromYaml(content, polaris)
}

// GetPolarisDBComponents get Polaris DB components
func GetPolarisDBComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := util.GetBaseYaml(baseURL, "polaris", polaris.Version, "polarisdb_base.yaml")
	if err != nil {
		return nil, err
	}

	return fromYaml(content, polaris)
}

// GetPolarisComponents get Polaris components
func GetPolarisComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := util.GetBaseYaml(baseURL, "polaris", polaris.Version, "polaris_base.yaml")
	if err != nil {
		return nil, err
	}

	return fromYaml(content, polaris)
}

// GetPolarisProvisionComponents get Polaris provision components
func GetPolarisProvisionComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := util.GetBaseYaml(baseURL, "polaris", polaris.Version, "organization-provision-job.yaml")
	if err != nil {
		return nil, err
	}

	return fromYaml(content, polaris)
}

// GetPolarisPostgresComponents get Polaris postgres components
func GetPolarisPostgresComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := util.GetBaseYaml(baseURL, "polaris", polaris.Version, "postgres_base.yaml")
	if err != nil {
		return nil, err
	}

	return fromYaml(content, polaris)
}

func removeTestManifests(objects map[string]runtime.Object) map[string]runtime.Object {
	objectsToBeRemoved := []string{
		"Pod.swip-db-ui-test-off9s",
		"Pod.swip-db-vault-status-test",
		"ConfigMap.polaris-db-consul-tests",
	}
	for _, object := range objectsToBeRemoved {
		delete(objects, object)
	}
	return objects
}
