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

// GetPolarisProvisionComponents get Polaris provision components
func GetPolarisProvisionComponents(baseURL string, polarisConf Polaris) (map[string]runtime.Object, error) {
	content, err := GetBaseYaml(baseURL, "polaris", polarisConf.Version, "organization-provision-job.yaml")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", polarisConf.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polarisConf.Namespace)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", polarisConf.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polarisConf.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${ORG_DESCRIPTION}", polarisConf.OrganizationDetails.OrganizationProvisionOrganizationDescription)
	content = strings.ReplaceAll(content, "${ORG_NAME}", polarisConf.OrganizationDetails.OrganizationProvisionOrganizationName)
	content = strings.ReplaceAll(content, "${ADMIN_NAME}", polarisConf.OrganizationDetails.OrganizationProvisionAdminName)
	content = strings.ReplaceAll(content, "${ADMIN_USERNAME}", polarisConf.OrganizationDetails.OrganizationProvisionAdminUsername)
	content = strings.ReplaceAll(content, "${ADMIN_EMAIL}", polarisConf.OrganizationDetails.OrganizationProvisionAdminEmail)
	content = strings.ReplaceAll(content, "${SEAT_COUNT}", polarisConf.OrganizationDetails.OrganizationProvisionLicenseSeatCount)
	content = strings.ReplaceAll(content, "${TYPE}", polarisConf.OrganizationDetails.OrganizationProvisionLicenseType)
	content = strings.ReplaceAll(content, "${RESULTS_START_DATE}", polarisConf.OrganizationDetails.OrganizationProvisionResultsStartDate)
	content = strings.ReplaceAll(content, "${RESULTS_END_DATE}", polarisConf.OrganizationDetails.OrganizationProvisionResultsEndDate)
	content = strings.ReplaceAll(content, "${RETENTION_START_DATE}", polarisConf.OrganizationDetails.OrganizationProvisionRetentionStartDate)
	content = strings.ReplaceAll(content, "${RETENTION_END_DATE}", polarisConf.OrganizationDetails.OrganizationProvisionRetentionEndDate)

	if len(polarisConf.Repository) != 0 {
		content = strings.ReplaceAll(content, "gcr.io/snps-swip-staging", polarisConf.Repository)
	}

	mapOfUniqueIDToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)

	patcher := polarisOrganizationJobPatcher{
		polaris:                          polarisConf,
		mapOfUniqueIDToBaseRuntimeObject: mapOfUniqueIDToBaseRuntimeObject,
	}

	patchStorageClass(mapOfUniqueIDToBaseRuntimeObject, polarisConf.StorageClass)

	return patcher.patch(), nil
}

type polarisOrganizationJobPatcher struct {
	polaris                          Polaris
	mapOfUniqueIDToBaseRuntimeObject map[string]runtime.Object
}

func (p *polarisOrganizationJobPatcher) patch() map[string]runtime.Object {
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

func (p *polarisOrganizationJobPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIDToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polaris.Namespace)
	}
	return nil
}

func (p *polarisOrganizationJobPatcher) patchVersionLabel() error {
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
