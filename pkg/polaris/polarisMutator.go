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
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetPolarisComponents get Polaris components
func GetPolarisComponents(baseURL string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := GetBaseYaml(baseURL, "polaris", polaris.Version, "polaris_base.yaml")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(JAEGER_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(EVENT_STORE_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(LEADER_ELECTOR_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polaris.Namespace)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", polaris.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polaris.ImagePullSecrets)

	content = strings.ReplaceAll(content, "${DOWNLOAD_SERVER_PV_SIZE}", polaris.PolarisSpec.DownloadServerDetails.Storage.StorageSize)

	if polaris.EnableReporting {
		content = strings.ReplaceAll(content, "${REPORTING_URL}", fmt.Sprintf("https://%s/reporting", polaris.EnvironmentDNS))
	} else {
		content = strings.ReplaceAll(content, "${REPORTING_URL}", "")
	}

	if len(polaris.Repository) != 0 {
		content = strings.ReplaceAll(content, "gcr.io/snps-swip-staging", polaris.Repository)
	}

	mapOfUniqueIDToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)
	mapOfUniqueIDToBaseRuntimeObject = removeTestManifests(mapOfUniqueIDToBaseRuntimeObject)

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

	patcher := polarisPatcher{
		polaris:                          polaris,
		mapOfUniqueIDToBaseRuntimeObject: mapOfUniqueIDToBaseRuntimeObject,
	}
	return patcher.patch(), nil
}

type polarisPatcher struct {
	polaris                          Polaris
	mapOfUniqueIDToBaseRuntimeObject map[string]runtime.Object
}

func (p *polarisPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchVersionLabel,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIDToBaseRuntimeObject
}

func (p *polarisPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIDToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polaris.Namespace)
	}
	return nil
}

func (p *polarisPatcher) patchVersionLabel() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIDToBaseRuntimeObject {
		labels, err := accessor.Labels(runtimeObject)
		if err != nil {
			return err
		}

		if labels == nil {
			labels = make(map[string]string)
		}

		labels["polaris.synopsys.com/version"] = p.polaris.Version
		labels["polaris.synopsys.com/environment"] = p.polaris.Namespace

		if err := accessor.SetLabels(runtimeObject, labels); err != nil {
			return err
		}
	}
	return nil
}
