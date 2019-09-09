/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
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

package controllers

import (
	"fmt"
	"strings"

	controllers_utils "github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/util"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func patchAlert(alertCr *synopsysv1.Alert, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object) map[string]runtime.Object {
	patcher := AlertPatcher{
		alertCr:                          alertCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
	}
	return patcher.patch()
}

type AlertPatcher struct {
	alertCr                          *synopsysv1.Alert
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
}

func (p *AlertPatcher) patch() map[string]runtime.Object {
	// Patches have to be done in order
	err := p.patchEnvirons()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	err = p.patchSecrets()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	err = p.patchImages()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	err = p.patchStorage()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	err = p.patchDesiredState()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	err = p.patchStandAlone()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	return p.mapOfUniqueIdToBaseRuntimeObject
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchEnvirons() error {
	configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ConfigMap.%s-blackduck-alert-config", p.alertCr.Name)]
	if !ok {
		return nil
	}
	configMap := configMapRuntimeObject.(*corev1.ConfigMap)
	for _, e := range p.alertCr.Spec.Environs {
		vals := strings.Split(e, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", e) // TODO change to log
			continue
		}
		environKey := strings.TrimSpace(vals[0])
		environVal := strings.TrimSpace(vals[1])
		configMap.Data[environKey] = environVal
	}
	return nil
}

func (p *AlertPatcher) patchSecrets() error {
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Secret.%s-alert-secret", p.alertCr.Name)]
	if !ok {
		return nil
	}
	secret := secretRuntimeObject.(*corev1.Secret)
	for _, s := range p.alertCr.Spec.Secrets {
		vals := strings.Split(*s, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", *s) // TODO change to log
			continue
		}
		secretKey := strings.TrimSpace(vals[0])
		secretVal := strings.TrimSpace(vals[1])
		secret.Data[secretKey] = []byte(secretVal)
	}

	specEncryptionGlobalSalt := p.alertCr.Spec.EncryptionGlobalSalt
	specEncryptionPassword := p.alertCr.Spec.EncryptionPassword

	if len(specEncryptionGlobalSalt) > 0 {
		secret.Data["ALERT_ENCRYPTION_GLOBAL_SALT"] = []byte(specEncryptionGlobalSalt)
	}

	if len(specEncryptionPassword) > 0 {
		secret.Data["ALERT_ENCRYPTION_PASSWORD"] = []byte(specEncryptionPassword)
	}

	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchImages() error {
	if p.alertCr.Spec.RegistryConfiguration != nil && (len(p.alertCr.Spec.RegistryConfiguration.Registry) > 0 || len(p.alertCr.Spec.ImageRegistries) > 0) {
		for _, baseRuntimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch baseRuntimeObject.(type) {
			case *corev1.ReplicationController:
				baseReplicationControllerRuntimeObject := baseRuntimeObject.(*corev1.ReplicationController)
				for _, container := range baseReplicationControllerRuntimeObject.Spec.Template.Spec.Containers {
					container.Image = controllers_utils.GenerateImageTag(container.Image, p.alertCr.Spec.ImageRegistries, *p.alertCr.Spec.RegistryConfiguration)
				}
			}
		}
	}
	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchDesiredState() error {
	for _, baseRuntimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch baseRuntimeObject.(type) {
		case *corev1.ReplicationController:
			switch strings.ToUpper(p.alertCr.Spec.DesiredState) {
			case "STOP":
				baseRuntimeObject.(*corev1.ReplicationController).Spec.Replicas = func(i int32) *int32 { return &i }(0)
			}
		}
	}
	return nil
}

// TODO: make common with Black Duck
func (p *AlertPatcher) patchStorage() error {
	for uniqueId, baseRuntimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch baseRuntimeObject.(type) {
		case *corev1.PersistentVolumeClaim:
			if !p.alertCr.Spec.PersistentStorage {
				delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueId)
			} else {
				if len(p.alertCr.Spec.PVCStorageClass) > 0 {
					baseRuntimeObject.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &p.alertCr.Spec.PVCStorageClass
				}

				if strings.EqualFold(p.alertCr.Spec.PVCName, baseRuntimeObject.(*corev1.PersistentVolumeClaim).Name) {
					baseRuntimeObject.(*corev1.PersistentVolumeClaim).Spec.VolumeName = p.alertCr.Spec.PVCName // TODO
					baseRuntimeObject.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &p.alertCr.Spec.PVCStorageClass
					if quantity, err := resource.ParseQuantity(p.alertCr.Spec.PVCSize); err == nil {
						baseRuntimeObject.(*corev1.PersistentVolumeClaim).Spec.Resources.Requests[corev1.ResourceStorage] = quantity
					}
				}
			}
		case *corev1.ReplicationController:
			if !p.alertCr.Spec.PersistentStorage {
				for volume := range baseRuntimeObject.(*corev1.ReplicationController).Spec.Template.Spec.Volumes {
					// If no PersistentVolumeClaim then we change it to emptyDir in the replication controller
					if baseRuntimeObject.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[volume].VolumeSource.PersistentVolumeClaim != nil {
						baseRuntimeObject.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[volume].VolumeSource = corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumDefault,
								SizeLimit: nil,
							},
						}
					}
				}
			}
		}

	}
	return nil
}

func (p *AlertPatcher) patchStandAlone() error {
	// if we want to use Black Duck's cfssl
	if *p.alertCr.Spec.StandAlone == false {
		// Remove Cfssl Resources
		uniqueID := fmt.Sprintf("ReplicationController.%s-cfssl", p.alertCr.Name)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)
		uniqueID = fmt.Sprintf("Service.%s-cfssl", p.alertCr.Name)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)
		// TODO: NOTE: THIS WILL NOT WORK WITHOUT SETTING 'HUB_CFSSL_HOST' manually
		// See: https://synopsys.atlassian.net/wiki/spaces/BDLM/pages/153583626/Synopsys+Alert+Installation+Guide+for+Synopsys+Operator
		// TODO: this should really be implemented by removing standalone field, and reconciling on an environs add of 'HUB_CFSSL_HOST'
	}
	return nil
}
