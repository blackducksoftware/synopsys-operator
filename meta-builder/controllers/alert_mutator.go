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

	alertsv1 "github.com/yashbhutwala/kb-synopsys-operator/api/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
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
	if p.alert.Spec.Namespace != "" {
		p.patchNamespace()
	}
	if p.alert.Spec.Size != "" {
		fmt.Printf("Spec Size field is currently not implemented\n")
	}
	if p.alert.Spec.AlertImage != "" {
		p.patchAlertImage()
	}
	if p.alert.Spec.CfsslImage != "" {
		p.patchAlertCfsslImage()
	}
	if p.alert.Spec.ExposeService != "" {
		p.patchExposeService()
	}
	if p.alert.Spec.StandAlone != nil {
		p.patchStandAlone()
	}
	return p.objects
}

func (p *AlertPatcher) patchNamespace() {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.objects {
		accessor.SetNamespace(runtimeObject, p.alert.Spec.Namespace)
	}
}

func (p *AlertPatcher) patchAlertImage() {
	//uniqueID := fmt.Sprintf("ReplicationController.%s.%s-alert-alert", p.alert.Namespace, p.alert.Name)
	uniqueID := "ReplicationController.default.demo-alert-alert"
	alertReplicationControllerRuntimeObject := p.objects[uniqueID]
	alertReplicationController := alertReplicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
	alertReplicationController.Spec.Template.Spec.Containers[0].Image = p.alert.Spec.AlertImage
}

func (p *AlertPatcher) patchAlertCfsslImage() {
	//uniqueID := fmt.Sprintf("ReplicationController.%s.%s-alert-cfssl", p.alert.Namespace, p.alert.Name)
	uniqueID := "ReplicationController.default.demo-alert-cfssl"
	alertCfsslReplicationControllerRuntimeObject := p.objects[uniqueID]
	alertCfsslReplicationController := alertCfsslReplicationControllerRuntimeObject.(*k8scorev1.ReplicationController)
	alertCfsslReplicationController.Spec.Template.Spec.Containers[0].Image = p.alert.Spec.CfsslImage
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

func (p *AlertPatcher) patchStandAlone() {
	if *p.alert.Spec.StandAlone == false {
		uniqueID := "ReplicationController.default.demo-alert-cfssl"
		delete(p.objects, uniqueID)
		uniqueID = "Service.default.demo-alert-cfssl"
		delete(p.objects, uniqueID)
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
	//TODO service.Spec.Ports[0].TargetPort = fmt.Sprintf("%d", port)
	service.Spec.Ports[0].Protocol = k8scorev1.ProtocolTCP

	ServiceExposedUniqueID := "Service.default.demo-alert-exposed"
	serviceExposedRuntimeObject := p.objects[ServiceExposedUniqueID]
	serviceExposed := serviceExposedRuntimeObject.(*k8scorev1.Service)
	serviceExposed.Spec.Ports[0].Name = fmt.Sprintf("port-%d", port)
	service.Spec.Ports[0].Port = port
	//TODO service.Spec.Ports[0].TargetPort = fmt.Sprintf("%d", port)
	service.Spec.Ports[0].Protocol = k8scorev1.ProtocolTCP

	// TODO: Support Openshift Routes
	// RouteUniqueID := "Route.default.demo-alert-route"
	// routeRuntimeObject := p.objects[RouteUniqueID]
}

func (p *AlertPatcher) patchEnvirons() {}

func (p *AlertPatcher) patchSecrets() {}

// TODO: Create functions to patch the remaining spec fields
