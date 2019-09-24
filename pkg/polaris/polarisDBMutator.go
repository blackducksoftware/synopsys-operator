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

	b64 "encoding/base64"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetPolarisDBComponents(baseUrl string, polaris Polaris) (map[string]runtime.Object, error) {
	content, err := GetBaseYaml(baseUrl, "polaris", polaris.Version, "polarisdb_base.yaml")
	if err != nil {
		return nil, err
	}

	// regex patching
	content = strings.ReplaceAll(content, "${NAMESPACE}", polaris.Namespace)
	content = strings.ReplaceAll(content, "${ENVIRONMENT_NAME}", polaris.EnvironmentName)
	content = strings.ReplaceAll(content, "${IMAGE_PULL_SECRETS}", polaris.ImagePullSecrets)
	content = strings.ReplaceAll(content, "${POSTGRES_USERNAME}", polaris.PolarisDBSpec.PostgresDetails.Username)
	content = strings.ReplaceAll(content, "${POSTGRES_PASSWORD}", polaris.PolarisDBSpec.PostgresDetails.Password)
	content = strings.ReplaceAll(content, "${SMTP_HOST}", polaris.PolarisDBSpec.SMTPDetails.Host)
	if polaris.PolarisDBSpec.SMTPDetails.Port != 2525 {
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
	content = strings.ReplaceAll(content, "${POSTGRES_HOST}", polaris.PolarisDBSpec.PostgresDetails.Host)
	if polaris.PolarisDBSpec.PostgresDetails.Port != 5432 {
		content = strings.ReplaceAll(content, "5432", strconv.Itoa(polaris.PolarisDBSpec.PostgresDetails.Port))
	}
	if polaris.PolarisDBSpec.PostgresInstanceType == "internal" {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "internal")
	} else {
		content = strings.ReplaceAll(content, "${POSTGRES_TYPE}", "external")
	}

	if len(polaris.PolarisDBSpec.EventstoreDetails.Storage.StorageSize) > 0 {
		content = strings.ReplaceAll(content, "${EVENTSTORE_PV_SIZE}", polaris.PolarisDBSpec.EventstoreDetails.Storage.StorageSize)
	} else {
		content = strings.ReplaceAll(content, "${EVENTSTORE_PV_SIZE}", EVENTSTORE_PV_SIZE)
	}

	if len(polaris.PolarisDBSpec.MongoDBDetails.Storage.StorageSize) > 0 {
		content = strings.ReplaceAll(content, "${MONGODB_PV_SIZE}", polaris.PolarisDBSpec.MongoDBDetails.Storage.StorageSize)
	} else {
		content = strings.ReplaceAll(content, "${MONGODB_PV_SIZE}", MONGODB_PV_SIZE)
	}

	if len(polaris.PolarisDBSpec.UploadServerDetails.Storage.StorageSize) > 0 {
		content = strings.ReplaceAll(content, "${UPLOAD_SERVER_PV_SIZE}", polaris.PolarisDBSpec.UploadServerDetails.Storage.StorageSize)
	} else {
		content = strings.ReplaceAll(content, "${UPLOAD_SERVER_PV_SIZE}", UPLOAD_SERVER_PV_SIZE)
	}

	if len(polaris.PolarisDBSpec.PostgresDetails.Storage.StorageSize) > 0 {
		content = strings.ReplaceAll(content, "${POSTGRES_PV_SIZE}", polaris.PolarisDBSpec.PostgresDetails.Storage.StorageSize)
	} else {
		content = strings.ReplaceAll(content, "${POSTGRES_PV_SIZE}", POSTGRES_PV_SIZE)
	}

	if len(polaris.Repository) != 0 {
		content = strings.ReplaceAll(content, "gcr.io/snps-swip-staging", polaris.Repository)
	}

	mapOfUniqueIdToBaseRuntimeObject := ConvertYamlFileToRuntimeObjects(content)
	mapOfUniqueIdToBaseRuntimeObject = removeTestManifests(mapOfUniqueIdToBaseRuntimeObject)

	patcher := polarisDBPatcher{
		polaris:                          polaris,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
	}
	return patcher.patch(), nil
}

func removeTestManifests(objects map[string]runtime.Object) map[string]runtime.Object {
	objectsToBeRemoved := []string{
		"Pod.swip-db-ui-test-off9s",
		"Pod.swip-db-vault-status-test",
		"ConfigMap.polaris-db-consul-tests",
		"Job.organization-provision-job",
	}
	for _, object := range objectsToBeRemoved {
		delete(objects, object)
	}
	return objects
}

type polarisDBPatcher struct {
	polaris                          Polaris
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
}

func (p *polarisDBPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchSMTPSecretDetails,
		p.patchSMTPConfigMapDetails,
		p.patchPostgresDetails,
		p.patchEventstoreDetails,
		p.patchUploadServerDetails,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *polarisDBPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.polaris.Namespace)
	}
	return nil
}

func (p *polarisDBPatcher) patchSMTPSecretDetails() error {
	SecretUniqueID := "Secret." + "smtp"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secretInstance := secretRuntimeObject.(*corev1.Secret)
	secretInstance.Data = map[string][]byte{
		"username": []byte(b64.StdEncoding.EncodeToString([]byte(p.polaris.PolarisDBSpec.SMTPDetails.Username))),
		"passwd":   []byte(b64.StdEncoding.EncodeToString([]byte(p.polaris.PolarisDBSpec.SMTPDetails.Password))),
	}
	return nil
}

func (p *polarisDBPatcher) patchSMTPConfigMapDetails() error {
	ConfigMapUniqueID := "ConfigMap." + "smtp"
	configmapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
	if !ok {
		return nil
	}
	configMapInstance := configmapRuntimeObject.(*corev1.ConfigMap)
	configMapInstance.Data = map[string]string{
		"host": p.polaris.PolarisDBSpec.SMTPDetails.Host,
		"port": strconv.Itoa(p.polaris.PolarisDBSpec.SMTPDetails.Port),
	}
	return nil
}

func (p *polarisDBPatcher) patchSMTPDetails() error {
	SecretUniqueID := "Secret." + "smtp"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secretInstance := secretRuntimeObject.(*corev1.Secret)
	secretInstance.Data = map[string][]byte{
		"username": []byte(b64.StdEncoding.EncodeToString([]byte(p.polaris.PolarisDBSpec.SMTPDetails.Username))),
		"passwd":   []byte(b64.StdEncoding.EncodeToString([]byte(p.polaris.PolarisDBSpec.SMTPDetails.Password))),
		"host":     []byte(b64.StdEncoding.EncodeToString([]byte(p.polaris.PolarisDBSpec.SMTPDetails.Host))),
		"port":     []byte(b64.StdEncoding.EncodeToString([]byte(string(p.polaris.PolarisDBSpec.SMTPDetails.Port)))),
	}
	return nil
}

func (p *polarisDBPatcher) patchPostgresDetails() error {
	// patch postgresql-config secret
	ConfigMapUniqueID := "ConfigMap." + "postgresql-config"
	configmapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
	if !ok {
		return nil
	}
	configMapInstance := configmapRuntimeObject.(*corev1.ConfigMap)
	configMapInstance.Data = map[string]string{
		"POSTGRESQL_ADMIN_PASSWORD": p.polaris.PolarisDBSpec.PostgresDetails.Password,
		"POSTGRESQL_DATABASE":       p.polaris.PolarisDBSpec.PostgresDetails.Username,
		"POSTGRESQL_PASSWORD":       p.polaris.PolarisDBSpec.PostgresDetails.Password,
		"POSTGRESQL_USER":           p.polaris.PolarisDBSpec.PostgresDetails.Username,
		"POSTGRESQL_HOST":           p.polaris.PolarisDBSpec.PostgresDetails.Host,
		"POSTGRESQL_PORT":           strconv.Itoa(p.polaris.PolarisDBSpec.PostgresDetails.Port),
	}
	if p.polaris.PolarisDBSpec.PostgresInstanceType == "internal" {
		// patch storage
		PostgresPVCUniqueID := "PersistentVolumeClaim." + "postgresql-pv-claim"
		PostgresPVCRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[PostgresPVCUniqueID]
		if !ok {
			return nil
		}
		PostgresPVCInstance := PostgresPVCRuntimeObject.(*corev1.PersistentVolumeClaim)
		UpdatePersistentVolumeClaim(PostgresPVCInstance, p.polaris.PolarisDBSpec.PostgresDetails.Storage.StorageSize)
	}
	return nil
}

func (p *polarisDBPatcher) patchEventstoreDetails() error {
	StatefulSetUniqueID := "StatefulSet." + "eventstore"
	statefulSetRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[StatefulSetUniqueID]
	if !ok {
		return nil
	}
	statefulsetInstance := statefulSetRuntimeObject.(*appsv1.StatefulSet)
	if size, err := resource.ParseQuantity(p.polaris.PolarisDBSpec.EventstoreDetails.Storage.StorageSize); err == nil {
		statefulsetInstance.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[v1.ResourceStorage] = size
	}
	return nil
}

func (p *polarisDBPatcher) patchUploadServerDetails() error {
	UploadServerPVCUniqueID := "PersistentVolumeClaim." + "upload-server-pv-claim"
	UploadServerPVCRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[UploadServerPVCUniqueID]
	if !ok {
		return nil
	}
	UploadServerPVCInstance := UploadServerPVCRuntimeObject.(*corev1.PersistentVolumeClaim)
	UpdatePersistentVolumeClaim(UploadServerPVCInstance, p.polaris.PolarisDBSpec.UploadServerDetails.Storage.StorageSize)
	return nil
}
