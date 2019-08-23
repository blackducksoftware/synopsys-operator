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
	"github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/controllers_utils"
	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func patchAlert(alertCr *synopsysv1.Alert, mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object, accessor meta.MetadataAccessor) map[string]runtime.Object {
	patcher := AlertPatcher{
		alertCr:                          alertCr,
		mapOfUniqueIdToBaseRuntimeObject: mapOfUniqueIdToBaseRuntimeObject,
		accessor:                         accessor,
	}
	return patcher.patch()
}

type AlertPatcher struct {
	alertCr                          *synopsysv1.Alert
	mapOfUniqueIdToBaseRuntimeObject map[string]runtime.Object
	accessor                         meta.MetadataAccessor
}

func (p *AlertPatcher) patch() map[string]runtime.Object {
	patches := []func() error{
		p.patchNamespace,
		p.patchEnvirons,
		p.patchSecrets,
		p.patchStandAlone,
		p.patchStorage,
		p.patchExposeService,
		//p.patchAlertImage,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return p.mapOfUniqueIdToBaseRuntimeObject
}

// TODO: patchName based on cr name [CR_NAME-]
//func (p *AlertPatcher) patchName() error {
//	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
//		p.accessor.SetName(runtimeObject, p.alertCr.Name)
//	}
//	return nil
//}

func (p *AlertPatcher) patchExposeService() error {
	uniqueId := fmt.Sprintf("Service.%s-alert-exposed", p.alertCr.Name)
	runtimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[uniqueId]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.alertCr.Spec.ExposeService) {
	case "NODEPORT":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
	case "LOADBALANCER":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
	default:
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueId)
	}
	return nil
}

// TODO: remove alert image and cfssl image and instead use image registries to make it consistent with black duck
//func (p *AlertPatcher) patchAlertImage() error {
//	alertReplicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ReplicationController.%s-alert", p.alertCr.Name)]
//	if !ok {
//		return nil
//	}
//	alertReplicationController := alertReplicationControllerRuntimeObject.(*corev1.ReplicationController)
//	alertReplicationController.Spec.Template.Spec.Containers[0].Image = p.alertCr.Spec.AlertImage
//	return nil
//}

// TODO: common with Black Duck
func (p *AlertPatcher) patchNamespace() error {
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		p.accessor.SetNamespace(runtimeObject, p.alertCr.Spec.Namespace)
	}
	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchEnvirons() error {
	configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ConfigMap.%s-blackDuckCr-alert-config", p.alertCr.Name)]
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
		vals := strings.Split(s, ":") // TODO - doesn't handle multiple colons
		// TODO: Base 64 decode
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", s) // TODO change to log
			continue
		}
		secretKey := strings.TrimSpace(vals[0])
		secretVal := strings.TrimSpace(vals[1])
		secret.Data[secretKey] = []byte(secretVal)
	}
	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchImages() error {
	if len(p.alertCr.Spec.RegistryConfiguration.Registry) > 0 || len(p.alertCr.Spec.ImageRegistries) > 0 {
		for _, v := range p.mapOfUniqueIdToBaseRuntimeObject {
			switch v.(type) {
			case *corev1.ReplicationController:
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image = controllers_utils.GenerateImageTag(v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image, p.alertCr.Spec.ImageRegistries, p.alertCr.Spec.RegistryConfiguration)
				}
			}
		}
	}
	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchDesiredState() error {
	if strings.EqualFold(p.alertCr.Spec.DesiredState, "STOP") {
		for uniqueID, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
			if k, _ := p.accessor.Kind(runtimeObject); k != "PersistentVolumeClaim" {
				delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)
			}
		}
	}
	return nil
}

func (p *AlertPatcher) patchPort() error {
	replicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ReplicationController.%s-alert", p.alertCr.Name)]
	if !ok {
		return nil
	}
	replicationController := replicationControllerRuntimeObject.(*corev1.ReplicationController)
	port := *p.alertCr.Spec.Port
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = port
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].Protocol = corev1.ProtocolTCP

	serviceRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Service.%s-alert", p.alertCr.Name)]
	if !ok {
		return nil
	}
	service := serviceRuntimeObject.(*corev1.Service)
	service.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: port}
	service.Spec.Ports[0].Protocol = corev1.ProtocolTCP

	serviceExposedRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("Service.%s-alert-exposed", p.alertCr.Name)]
	if !ok {
		return nil
	}
	serviceExposed := serviceExposedRuntimeObject.(*corev1.Service)
	serviceExposed.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: port}
	service.Spec.Ports[0].Protocol = corev1.ProtocolTCP

	// TODO: Support OpenShift Routes
	// RouteUniqueID := "Route.default.demo-alertCr-route"
	// routeRuntimeObject := p.mapOfUniqueIdToBaseRuntimeObject[RouteUniqueID]

	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchAlertMemory() error {
	alertReplicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[fmt.Sprintf("ReplicationController.%s-alert", p.alertCr.Name)]
	if !ok {
		return nil
	}
	alertReplicationController := alertReplicationControllerRuntimeObject.(*corev1.ReplicationController)
	minAndMaxMem, _ := resource.ParseQuantity(p.alertCr.Spec.AlertMemory)
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = minAndMaxMem
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = minAndMaxMem
	return nil
}

// TODO: common with Black Duck
func (p *AlertPatcher) patchStorage() error {
	for k, v := range p.mapOfUniqueIdToBaseRuntimeObject {
		switch v.(type) {
		case *corev1.PersistentVolumeClaim:
			if !p.alertCr.Spec.PersistentStorage {
				delete(p.mapOfUniqueIdToBaseRuntimeObject, k)
			} else {
				if len(p.alertCr.Spec.PVCStorageClass) > 0 {
					v.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &p.alertCr.Spec.PVCStorageClass
				}
				for _, pvc := range p.alertCr.Spec.PVC {
					if strings.EqualFold(pvc.Name, v.(*corev1.PersistentVolumeClaim).Name) {
						v.(*corev1.PersistentVolumeClaim).Spec.VolumeName = pvc.VolumeName
						v.(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &pvc.StorageClass
						if quantity, err := resource.ParseQuantity(pvc.Size); err == nil {
							v.(*corev1.PersistentVolumeClaim).Spec.Resources.Requests[corev1.ResourceStorage] = quantity
						}
					}
				}
			}
		case *corev1.ReplicationController:
			if !p.alertCr.Spec.PersistentStorage {
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes {
					// If PersistentVolumeClaim then we change it to emptyDir
					if v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource.PersistentVolumeClaim != nil {
						v.(*corev1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource = corev1.VolumeSource{
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
	if (p.alertCr.Spec.StandAlone == synopsysv1.StandAlone{}) {
		// Remove Cfssl Resources
		uniqueID := fmt.Sprintf("ReplicationController.%s-cfssl", p.alertCr.Name)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)
		uniqueID = fmt.Sprintf("Service.%s-cfssl", p.alertCr.Name)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)

		// Add Environ to use BlackDuck Cfssl
		ConfigMapUniqueID := fmt.Sprintf("ConfigMap.%s-blackDuckCr-alert-config", p.alertCr.Name)
		configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
		if !ok {
			return nil
		}
		configMap := configMapRuntimeObject.(*corev1.ConfigMap)
		configMap.Data["HUB_CFSSL_HOST"] = fmt.Sprintf("%s-%s-%s", p.alertCr.Name, "alert", "cfssl")
	} else {
		uniqueID := fmt.Sprintf("ReplicationController.%s-cfssl", p.alertCr.Name)
		alertCfsslReplicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[uniqueID]
		if !ok {
			return nil
		}
		// patch Cfssl Image
		alertCfsslReplicationController := alertCfsslReplicationControllerRuntimeObject.(*corev1.ReplicationController)
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Image = p.alertCr.Spec.StandAlone.CfsslImage
		// patch Cfssl Memory
		minAndMaxMem, _ := resource.ParseQuantity(p.alertCr.Spec.StandAlone.CfsslMemory)
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = minAndMaxMem
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = minAndMaxMem
	}
	return nil
}
