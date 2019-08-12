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

	alertsv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func patchAlert(alert *alertsv1.Alert, objects map[string]runtime.Object) map[string]runtime.Object {
	patcher := AlertPatcher{
		alert:   alert,
		objects: objects,
	}
	return patcher.patch()
}

type AlertBuilder struct {
	replicationController      *runtime.Object
	cfsslReplicationController *runtime.Object
	service                    *runtime.Object
	cfsslService               *runtime.Object
	nodeportService            *runtime.Object
	loadbalancerService        *runtime.Object
	route                      *runtime.Object
	configMap                  *runtime.Object
	secret                     *runtime.Object
}

type AlertPatcher struct {
	alert   *alertsv1.Alert
	objects map[string]runtime.Object
}

func (p *AlertPatcher) patch() map[string]runtime.Object {
	p.patchNamespace()
	p.patchEnvirons()
	p.patchSecrets()
	p.patchStandAlone()
	p.patchPersistentStorage()
	p.patchExposeService()
	p.patchAlertImage()
	p.patchAlertSize()

	return p.objects
}

func (p *AlertPatcher) patchNamespace() {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.objects {
		accessor.SetNamespace(runtimeObject, p.alert.Spec.Namespace)
	}
}

func (p *AlertPatcher) patchEnvirons() {
	ConfigMapUniqueID := "ConfigMap.default.demo-alert-blackduck-config"
	configMapRuntimeObject := p.objects[ConfigMapUniqueID]
	configMap := configMapRuntimeObject.(*k8scorev1.ConfigMap)
	for _, e := range p.alert.Spec.Environs {
		vals := strings.Split(e, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", e) // TODO change to log
			continue
		}
		environKey := strings.TrimSpace(vals[0])
		environVal := strings.TrimSpace(vals[1])
		configMap.Data[environKey] = environVal
	}
}

func (p *AlertPatcher) patchSecrets() {
	SecretUniqueID := "Secret.default.demo-alert-secret"
	secretRuntimeObject := p.objects[SecretUniqueID]
	secret := secretRuntimeObject.(*k8scorev1.Secret)
	for _, s := range p.alert.Spec.Environs {
		vals := strings.Split(s, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", s) // TODO change to log
			continue
		}
		secretKey := strings.TrimSpace(vals[0])
		secretVal := strings.TrimSpace(vals[1])
		secret.Data[secretKey] = []byte(secretVal)
	}
}

func (p *AlertPatcher) patchDesiredState() {
	accessor := meta.NewAccessor()
	if strings.EqualFold(p.alert.Spec.DesiredState, "STOP") {
		for uniqueID, runtimeObject := range p.objects {
			if k, _ := accessor.Kind(runtimeObject); k != "PersistentVolumeClaim" {
				delete(p.objects, uniqueID)
			}
		}
	}
}

func (p *AlertPatcher) patchPort() {
	port := *p.alert.Spec.Port
	ReplicationContollerUniqueID := "ReplicationController.default.demo-alert-alert"
	replicationControllerRuntimeObject := p.objects[ReplicationContollerUniqueID]
	replicationController := replicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = port
	replicationController.Spec.Template.Spec.Containers[0].Ports[0].Protocol = k8scorev1.ProtocolTCP

	ServiceUniqueID := "Service.default.demo-alert-alert"
	serviceRuntimeObject := p.objects[ServiceUniqueID]
	service := serviceRuntimeObject.(*k8scorev1.Service)
	service.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: port}
	service.Spec.Ports[0].Protocol = k8scorev1.ProtocolTCP

	ServiceExposedUniqueID := "Service.default.demo-alert-exposed"
	serviceExposedRuntimeObject := p.objects[ServiceExposedUniqueID]
	serviceExposed := serviceExposedRuntimeObject.(*k8scorev1.Service)
	serviceExposed.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: port}
	service.Spec.Ports[0].Protocol = k8scorev1.ProtocolTCP

	// TODO: Support Openshift Routes
	// RouteUniqueID := "Route.default.demo-alert-route"
	// routeRuntimeObject := p.objects[RouteUniqueID]
}

func (p *AlertPatcher) patchPersistentStorage() {
	if (p.alert.Spec.PersistentStorage == alertsv1.PersistentStorage{}) {
		PVCUniqueID := "PersistentVolumeClaim.default.demo-alert-pvc"
		delete(p.objects, PVCUniqueID)
		return
	}
	// Patch PVC Name
	PVCUniqueID := "PersistentVolumeClaim.default.demo-alert-pvc"
	PVCRuntimeObject := p.objects[PVCUniqueID]
	pvc := PVCRuntimeObject.(*k8scorev1.PersistentVolumeClaim)

	name := fmt.Sprintf("%s-%s-%s", p.alert.Name, "alert", p.alert.Spec.PersistentStorage.PVCName)
	if p.alert.Annotations["synopsys.com/created.by"] == "pre-2019.6.0" {
		name = p.alert.Spec.PersistentStorage.PVCName
	}
	pvc.Name = name
}

func (p *AlertPatcher) patchStandAlone() {
	if (p.alert.Spec.StandAlone == alertsv1.StandAlone{}) {
		// Remove Cfssl Resources
		uniqueID := "ReplicationController.default.demo-alert-cfssl"
		delete(p.objects, uniqueID)
		uniqueID = "Service.default.demo-alert-cfssl"
		delete(p.objects, uniqueID)

		// Add Environ to use BlackDuck Cfssl
		ConfigMapUniqueID := "ConfigMap.default.demo-alert-blackduck-config"
		configMapRuntimeObject := p.objects[ConfigMapUniqueID]
		configMap := configMapRuntimeObject.(*k8scorev1.ConfigMap)
		configMap.Data["HUB_CFSSL_HOST"] = fmt.Sprintf("%s-%s-%s", p.alert.Name, "alert", "cfssl")
	} else {
		uniqueID := "ReplicationController.default.demo-alert-cfssl"
		alertCfsslReplicationControllerRuntimeObject := p.objects[uniqueID]
		// patch Cfssl Image
		alertCfsslReplicationController := alertCfsslReplicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Image = p.alert.Spec.StandAlone.CfsslImage
		// patch Cfssl Memory
		minAndMaxMem, _ := resource.ParseQuantity(p.alert.Spec.StandAlone.CfsslMemory)
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Resources.Requests[k8scorev1.ResourceMemory] = minAndMaxMem
		alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Resources.Limits[k8scorev1.ResourceMemory] = minAndMaxMem
	}
}

func (p *AlertPatcher) patchExposeService() {
	switch p.alert.Spec.ExposeService {
	case "NODEPORT":
		uniqueID := "Service.default.demo-alert-exposed"
		delete(p.objects, uniqueID)
	case "LOADBALANCER":
		uniqueID := "Service.default.demo-alert-exposed"
		delete(p.objects, uniqueID)
	}
}

func (p *AlertPatcher) patchAlertImage() {
	uniqueID := "ReplicationController.default.demo-alert-alert"
	alertReplicationControllerRuntimeObject := p.objects[uniqueID]
	alertReplicationController := alertReplicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
	alertReplicationController.Spec.Template.Spec.Containers[0].Image = p.alert.Spec.AlertImage
}

func (p *AlertPatcher) patchAlertSize() {
	fmt.Printf("Alert Size Field is currently not implemented")
}

func (p *AlertPatcher) patchAlertMemory() {
	uniqueID := "ReplicationController.default.demo-alert-alert"
	alertReplicationControllerRuntimeObject := p.objects[uniqueID]
	alertReplicationController := alertReplicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
	minAndMaxMem, _ := resource.ParseQuantity(p.alert.Spec.AlertMemory)
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Requests[k8scorev1.ResourceMemory] = minAndMaxMem
	alertReplicationController.Spec.Template.Spec.Containers[0].Resources.Limits[k8scorev1.ResourceMemory] = minAndMaxMem
}

func (p *AlertPatcher) patchImageRegistries() {
	fmt.Printf("Alert ImageRegistires is currently not implemented")
}

func (p *AlertPatcher) patchRegistryConfiguration() {
	fmt.Printf("Alert RegistryConfiguration is currently not implemented")
}
