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

	v1 "k8s.io/api/batch/v1"
)

func GetPolarisProvisionJob(baseUrl string, jobConfig ProvisionJob) (*v1.Job, error) {
	content, err := GetBaseYaml(baseUrl, "polaris", jobConfig.Version, "organization-provision-job.yaml")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", jobConfig.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", jobConfig.Namespace)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", jobConfig.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", jobConfig.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${ORG_DESCRIPTION}", jobConfig.OrganizationProvisionOrganizationDescription)
	content = strings.ReplaceAll(content, "${ORG_NAME}", jobConfig.OrganizationProvisionOrganizationName)
	content = strings.ReplaceAll(content, "${ADMIN_NAME}", jobConfig.OrganizationProvisionAdminName)
	content = strings.ReplaceAll(content, "${ADMIN_USERNAME}", jobConfig.OrganizationProvisionAdminUsername)
	content = strings.ReplaceAll(content, "${ADMIN_EMAIL}", jobConfig.OrganizationProvisionAdminEmail)
	content = strings.ReplaceAll(content, "${SEAT_COUNT}", jobConfig.OrganizationProvisionLicenseSeatCount)
	content = strings.ReplaceAll(content, "${TYPE}", jobConfig.OrganizationProvisionLicenseType)
	content = strings.ReplaceAll(content, "${RESULTS_START_DATE}", jobConfig.OrganizationProvisionResultsStartDate)
	content = strings.ReplaceAll(content, "${RESULTS_END_DATE}", jobConfig.OrganizationProvisionResultsEndDate)
	content = strings.ReplaceAll(content, "${RETENTION_START_DATE}", jobConfig.OrganizationProvisionRetentionStartDate)
	content = strings.ReplaceAll(content, "${RETENTION_END_DATE}", jobConfig.OrganizationProvisionRetentionEndDate)

	if len(jobConfig.Repository) != 0 {
		content = strings.ReplaceAll(content, "gcr.io/snps-swip-staging", jobConfig.Repository)
	}

	mapOfUniqueIdToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)
	jobRuntimeObject, ok := mapOfUniqueIdToBaseRuntimeObject["Job.organization-provision-job"]
	if !ok {
		return nil, fmt.Errorf("couldn't find organization-provision-job Job ")
	}

	job, ok := jobRuntimeObject.(*v1.Job)
	if !ok {
		return nil, fmt.Errorf("couldn't cast organization-provision-job ")
	}

	return job, nil
}
