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
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func fromYaml(content string, polaris Polaris) (map[string]runtime.Object, error) {
	// Basic
	content = strings.ReplaceAll(content, "${NAMESPACE}", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(JAEGER_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(EVENT_STORE_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "$(LEADER_ELECTOR_NAMESPACE)", polaris.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polaris.Namespace)
	content = strings.ReplaceAll(content, "${POLARIS_ROOT_DOMAIN}", polaris.EnvironmentDNS)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polaris.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${INGRESS_CLASS}", polaris.IngressClass)

	// PVC
	content = strings.ReplaceAll(content, "${DOWNLOAD_SERVER_PV_SIZE}", polaris.PolarisSpec.DownloadServerDetails.Storage.StorageSize)
	content = strings.ReplaceAll(content, "${EVENTSTORE_PV_SIZE}", polaris.PolarisDBSpec.EventstoreDetails.Storage.StorageSize)
	content = strings.ReplaceAll(content, "${MONGODB_PV_SIZE}", polaris.PolarisDBSpec.MongoDBDetails.Storage.StorageSize)
	content = strings.ReplaceAll(content, "${UPLOAD_SERVER_PV_SIZE}", polaris.PolarisDBSpec.UploadServerDetails.Storage.StorageSize)
	content = strings.ReplaceAll(content, "${POSTGRES_PV_SIZE}", polaris.PolarisDBSpec.PostgresDetails.Storage.StorageSize)
	content = strings.ReplaceAll(content, "${REPORT_STORAGE_PV_SIZE}", polaris.ReportingSpec.ReportStorageDetails.Storage.StorageSize)

	// SMTP
	content = strings.ReplaceAll(content, "${SMTP_SENDER_EMAIL}", polaris.PolarisDBSpec.SMTPDetails.SenderEmail)
	content = strings.ReplaceAll(content, "${SMTP_HOST}", polaris.PolarisDBSpec.SMTPDetails.Host)
	if polaris.PolarisDBSpec.SMTPDetails.Port != 2525 {
		// TODO this needs to be a placeholder
		content = strings.ReplaceAll(content, "2525", strconv.Itoa(polaris.PolarisDBSpec.SMTPDetails.Port))
	}
	if len(polaris.PolarisDBSpec.SMTPDetails.Username) != 0 {
		content = strings.ReplaceAll(content, "${SMTP_USERNAME}", EncodeStringToBase64(polaris.PolarisDBSpec.SMTPDetails.Username))
	} else {
		content = strings.ReplaceAll(content, "${SMTP_USERNAME}", "Cg==")
	}
	if len(polaris.PolarisDBSpec.SMTPDetails.Password) != 0 {
		content = strings.ReplaceAll(content, "${SMTP_PASSWORD}", fmt.Sprintf("\"%s\"", EncodeStringToBase64(polaris.PolarisDBSpec.SMTPDetails.Password)))
	} else {
		content = strings.ReplaceAll(content, "${SMTP_PASSWORD}", "Cg==")
	}

	// Postgres
	content = strings.ReplaceAll(content, "${POSTGRES_USERNAME}", polaris.PolarisDBSpec.PostgresDetails.Username)
	content = strings.ReplaceAll(content, "${POSTGRES_PASSWORD}", polaris.PolarisDBSpec.PostgresDetails.Password)
	content = strings.ReplaceAll(content, "${POSTGRES_HOST}", polaris.PolarisDBSpec.PostgresDetails.Host)
	if polaris.PolarisDBSpec.PostgresDetails.Port != 5432 {
		// TODO this needs to be a placeholder
		content = strings.ReplaceAll(content, "5432", strconv.Itoa(polaris.PolarisDBSpec.PostgresDetails.Port))
	}
	if polaris.PolarisDBSpec.PostgresInstanceType == "internal" {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "internal")
	} else {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "external")
	}

	// Reporting
	if polaris.EnableReporting {
		content = strings.ReplaceAll(content, "${REPORTING_URL}", fmt.Sprintf("https://%s/reporting", polaris.EnvironmentDNS))
	} else {
		content = strings.ReplaceAll(content, "${REPORTING_URL}", "")
	}

	// Org job
	content = strings.ReplaceAll(content, "${ORG_DESCRIPTION}", polaris.OrganizationDetails.OrganizationProvisionOrganizationDescription)
	content = strings.ReplaceAll(content, "${ORG_NAME}", polaris.OrganizationDetails.OrganizationProvisionOrganizationName)
	content = strings.ReplaceAll(content, "${ADMIN_NAME}", polaris.OrganizationDetails.OrganizationProvisionAdminName)
	content = strings.ReplaceAll(content, "${ADMIN_USERNAME}", polaris.OrganizationDetails.OrganizationProvisionAdminUsername)
	content = strings.ReplaceAll(content, "${ADMIN_EMAIL}", polaris.OrganizationDetails.OrganizationProvisionAdminEmail)
	content = strings.ReplaceAll(content, "${SEAT_COUNT}", polaris.OrganizationDetails.OrganizationProvisionLicenseSeatCount)
	content = strings.ReplaceAll(content, "${TYPE}", polaris.OrganizationDetails.OrganizationProvisionLicenseType)
	content = strings.ReplaceAll(content, "${RESULTS_START_DATE}", polaris.OrganizationDetails.OrganizationProvisionResultsStartDate)
	content = strings.ReplaceAll(content, "${RESULTS_END_DATE}", polaris.OrganizationDetails.OrganizationProvisionResultsEndDate)
	content = strings.ReplaceAll(content, "${RETENTION_START_DATE}", polaris.OrganizationDetails.OrganizationProvisionRetentionStartDate)
	content = strings.ReplaceAll(content, "${RETENTION_END_DATE}", polaris.OrganizationDetails.OrganizationProvisionRetentionEndDate)

	// TODO https://github.com/blackducksoftware/synopsys-operator/issues/943
	if len(polaris.Repository) != 0 {
		content = strings.ReplaceAll(content, "gcr.io/snps-swip-staging", polaris.Repository)
	}

	mapOfUniqueIDToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)
	mapOfUniqueIDToBaseRuntimeObject = removeTestManifests(mapOfUniqueIDToBaseRuntimeObject)

	patcher := Patcher{
		polaris:                          polaris,
		mapOfUniqueIDToBaseRuntimeObject: mapOfUniqueIDToBaseRuntimeObject,
	}

	return patcher.patch()
}

type Patcher struct {
	polaris                          Polaris
	mapOfUniqueIDToBaseRuntimeObject map[string]runtime.Object
}

func (p *Patcher) patch() (map[string]runtime.Object, error) {
	patches := []func() error{
		p.patchStorageClass,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			return nil, err
		}
	}
	return p.mapOfUniqueIDToBaseRuntimeObject, nil
}

// patchStorageClass will iterate over the runtime objects and update the storage class
func (p *Patcher) patchStorageClass() error {
	if len(p.polaris.StorageClass) > 0 {
		for k, v := range p.mapOfUniqueIDToBaseRuntimeObject {
			switch v.(type) {
			case *appsv1beta1.StatefulSet:
				for claimTemplateIndex := range p.mapOfUniqueIDToBaseRuntimeObject[k].(*appsv1beta1.StatefulSet).Spec.VolumeClaimTemplates {
					p.mapOfUniqueIDToBaseRuntimeObject[k].(*appsv1beta1.StatefulSet).Spec.VolumeClaimTemplates[claimTemplateIndex].Spec.StorageClassName = &p.polaris.StorageClass
				}
			case *appsv1.StatefulSet:
				for claimTemplateIndex := range p.mapOfUniqueIDToBaseRuntimeObject[k].(*appsv1.StatefulSet).Spec.VolumeClaimTemplates {
					p.mapOfUniqueIDToBaseRuntimeObject[k].(*appsv1.StatefulSet).Spec.VolumeClaimTemplates[claimTemplateIndex].Spec.StorageClassName = &p.polaris.StorageClass
				}
			case *corev1.PersistentVolumeClaim:
				p.mapOfUniqueIDToBaseRuntimeObject[k].(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &p.polaris.StorageClass
			}
		}
	}
	return nil
}
