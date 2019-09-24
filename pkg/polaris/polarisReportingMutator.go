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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetPolarisReportingComponents(baseUrl string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := GetBaseYaml(baseUrl, "polaris", polaris.Version, "reporting_base.yaml")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", polaris.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polaris.Namespace)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", polaris.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polaris.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${REPORT_STORAGE_PV_SIZE}", polaris.ReportingSpec.ReportStorageDetails.Storage.StorageSize)

	if len(polaris.ReportingSpec.ReportStorageDetails.Storage.StorageSize) != 0 {
		content = strings.ReplaceAll(content, "${REPORT_STORAGE_PV_SIZE}", polaris.ReportingSpec.ReportStorageDetails.Storage.StorageSize)
	} else {
		content = strings.ReplaceAll(content, "${REPORT_STORAGE_PV_SIZE}", REPORT_STORAGE_PV_SIZE)
	}

	mapOfUniqueIdToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)

	patcher := polarisReportingPatcher{
		polaris:                          polaris,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
	}
	return patcher.patch(), nil
}

type polarisReportingPatcher struct {
	polaris                          Polaris
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
}

func (p *polarisReportingPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *polarisReportingPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polaris.Namespace)
	}
	return nil
}
