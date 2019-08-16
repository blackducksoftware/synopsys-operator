/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
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
		p.patchPersistentStorage,
		p.patchExposeUserInterface,
		p.patchAlertImage,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return p.mapOfUniqueIdToBaseRuntimeObject
}

func (p *AlertPatcher) patchNamespace() error {
	for _, runtimeObject := range p.mapOfUniqueIdToBaseRuntimeObject {
		p.accessor.SetNamespace(runtimeObject, p.alertCr.Spec.Namespace)
	}
	return nil
}

func (p *AlertPatcher) patchEnvirons() error {
	ConfigMapUniqueID := "ConfigMap.demo-alert-blackduck-config"
	configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
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
	SecretUniqueID := "Secret.demo-alert-secret"
	secretRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[SecretUniqueID]
	if !ok {
		return nil
	}
	secret := secretRuntimeObject.(*corev1.Secret)
	for _, s := range p.alertCr.Spec.Environs {
		vals := strings.Split(s, ":") // TODO - doesn't handle multiple colons
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
	ReplicationControllerUniqueID := "ReplicationController.demo-alert-alert"
	replicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ReplicationControllerUniqueID]
	if !ok {
		return nil
	}
	replicationController := replicationControllerRuntimeObject.(*corev1.ReplicationController)
	port := *p.alertCr.Spec.Port
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = port
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].Protocol = corev1.ProtocolTCP

	ServiceUniqueID := "Service.default.demo-alertCr-alertCr"
	serviceRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ServiceUniqueID]
	if !ok {
		return nil
	}
	service := serviceRuntimeObject.(*corev1.Service)
	service.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: port}
	service.Spec.Ports[0].Protocol = corev1.ProtocolTCP

	ServiceExposedUniqueID := "Service.default.demo-alertCr-exposed"
	serviceExposedRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ServiceExposedUniqueID]
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

func (p *AlertPatcher) patchAlertImage() error {
	uniqueID := "ReplicationController.demo-alert-alert"
	alertReplicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[uniqueID]
	if !ok {
		return nil
	}
	alertReplicationController := alertReplicationControllerRuntimeObject.(*corev1.ReplicationController)
	alertReplicationController.Spec.Template.Spec.Containers[0].Image = p.alertCr.Spec.AlertImage
	return nil
}

func (p *AlertPatcher) patchAlertMemory() error {
	uniqueID := "ReplicationController.demo-alert-alert"
	alertReplicationControllerRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[uniqueID]
	if !ok {
		return nil
	}
	alertReplicationController := alertReplicationControllerRuntimeObject.(*corev1.ReplicationController)
	minAndMaxMem, _ := resource.ParseQuantity(p.alertCr.Spec.AlertMemory)
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = minAndMaxMem
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = minAndMaxMem
	return nil
}

func (p *AlertPatcher) patchPersistentStorage() error {
	if (p.alertCr.Spec.PersistentStorage == synopsysv1.PersistentStorage{}) {
		PVCUniqueID := "PersistentVolumeClaim.demo-alert-pvc"
		delete(p.mapOfUniqueIdToBaseRuntimeObject, PVCUniqueID)
		return nil
	}
	// Patch PVC Name
	PVCUniqueID := "PersistentVolumeClaim.demo-alert-pvc"
	PVCRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[PVCUniqueID]
	if !ok {
		return nil
	}
	pvc := PVCRuntimeObject.(*corev1.PersistentVolumeClaim)

	name := fmt.Sprintf("%s-%s-%s", p.alertCr.Name, "alertCr", p.alertCr.Spec.PersistentStorage.PVCName)
	if p.alertCr.Annotations["synopsys.com/created.by"] == "pre-2019.6.0" {
		name = p.alertCr.Spec.PersistentStorage.PVCName
	}
	pvc.Name = name

	return nil
}

func (p *AlertPatcher) patchStandAlone() error {
	if (p.alertCr.Spec.StandAlone == synopsysv1.StandAlone{}) {
		// Remove Cfssl Resources
		uniqueID := "ReplicationController.demo-alert-cfssl"
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)
		uniqueID = "Service.demo-alert-cfssl"
		delete(p.mapOfUniqueIdToBaseRuntimeObject, uniqueID)

		// Add Environ to use BlackDuck Cfssl
		ConfigMapUniqueID := "ConfigMap.demo-alert-blackduck-config"
		configMapRuntimeObject, ok := p.mapOfUniqueIdToBaseRuntimeObject[ConfigMapUniqueID]
		if !ok {
			return nil
		}
		configMap := configMapRuntimeObject.(*corev1.ConfigMap)
		configMap.Data["HUB_CFSSL_HOST"] = fmt.Sprintf("%s-%s-%s", p.alertCr.Name, "alertCr", "cfssl")
	} else {
		uniqueID := "ReplicationController.demo-alert-cfssl"
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

func (p *AlertPatcher) patchExposeUserInterface() error {
	nodePortUniqueID := "Service.demo-alert-exposed"
	loadbalancerUniqueID := "Service.demo-alert-exposed"
	routeUniqueID := "Service.demo-alert-exposed"
	switch p.alertCr.Spec.ExposeService {
	case "NODEPORT":
		delete(p.mapOfUniqueIdToBaseRuntimeObject, loadbalancerUniqueID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeUniqueID)
	case "LOADBALANCER":
		delete(p.mapOfUniqueIdToBaseRuntimeObject, nodePortUniqueID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeUniqueID)
	case "OPENSHIFT":
		delete(p.mapOfUniqueIdToBaseRuntimeObject, nodePortUniqueID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, loadbalancerUniqueID)
	default:
		delete(p.mapOfUniqueIdToBaseRuntimeObject, nodePortUniqueID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, loadbalancerUniqueID)
		delete(p.mapOfUniqueIdToBaseRuntimeObject, routeUniqueID)
	}
	return nil
}
