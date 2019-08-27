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

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchOpsSight(client client.Client, opsSight *synopsysv1.OpsSight, runtimeObjects map[string]runtime.Object) map[string]runtime.Object {
	patcher := OpsSightPatcher{
		Client:         client,
		opsSight:       opsSight,
		runtimeObjects: runtimeObjects,
	}
	return patcher.patch()
}

// OpsSightPatcher applies the patches to OpsSight
type OpsSightPatcher struct {
	client.Client
	opsSight       *synopsysv1.OpsSight
	runtimeObjects map[string]runtime.Object
}

func (p *OpsSightPatcher) patch() map[string]runtime.Object {
	// TODO JD: Patching this way is costly. Consider iterating over the objects only once
	// and apply the necessary changes

	patches := []func() error{
		p.patchImages,
		p.patchOpsSightCoreModelExposeService,
		p.patchOpsSightPrometheusExposeService,
		p.patchImageProcessor,
		p.patchPodProcessor,
		p.patchPrometheusMetrics,
		p.patchProcessor,
		p.patchCoreModelUI,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return p.runtimeObjects
}

func (p *OpsSightPatcher) patchOpsSightCoreModelExposeService() error {
	// TODO use constants
	id := fmt.Sprintf("Service.%s-opssight-core-exposed", p.opsSight.Name)
	runtimeObject, ok := p.runtimeObjects[id]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.opsSight.Spec.Perceptor.Expose) {
	case "LOADBALANCER":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
	case "NODEPORT":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
	default:
		delete(p.runtimeObjects, id)
	}

	return nil
}

func (p *OpsSightPatcher) patchOpsSightPrometheusExposeService() error {
	// TODO use constants
	id := fmt.Sprintf("Service.%s-opssight-prometheus-exposed", p.opsSight.Name)
	runtimeObject, ok := p.runtimeObjects[id]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.opsSight.Spec.Prometheus.Expose) {
	case "LOADBALANCER":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
	case "NODEPORT":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
	}

	return nil
}

func (p *OpsSightPatcher) patchImages() error {
	if len(p.opsSight.Spec.RegistryConfiguration.Registry) > 0 || len(p.opsSight.Spec.ImageRegistries) > 0 {
		for _, v := range p.runtimeObjects {
			switch v.(type) {
			case *corev1.ReplicationController:
				for i := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image = controllers_utils.GenerateImageTag(v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[i].Image, p.opsSight.Spec.ImageRegistries, p.opsSight.Spec.RegistryConfiguration)
				}
			}
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchProcessor() error {
	if !p.opsSight.Spec.Perceiver.EnableImagePerceiver && !p.opsSight.Spec.Perceiver.EnablePodPerceiver {
		delete(p.runtimeObjects, fmt.Sprintf("ServiceAccount.%s-opssight-processor", p.opsSight.Name))
	}
	return nil
}

func (p *OpsSightPatcher) patchImageProcessor() error {
	// TODO need to check if the cluster is OpenShift
	if !p.opsSight.Spec.Perceiver.EnableImagePerceiver {
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRole.%s-opssight-image-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRoleBinding.%s-opssight-image-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Service.%s-opssight-image-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("ReplicationController.%s-opssight-image-processor", p.opsSight.Name))
	}
	return nil
}

func (p *OpsSightPatcher) patchPodProcessor() error {
	if !p.opsSight.Spec.Perceiver.EnablePodPerceiver {
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRole.%s-opssight-pod-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRoleBinding.%s-opssight-pod-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Service.%s-opssight-pod-processor", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("ReplicationController.%s-opssight-pod-processor", p.opsSight.Name))
	}
	return nil
}

func (p *OpsSightPatcher) patchPrometheusMetrics() error {
	if !p.opsSight.Spec.EnableMetrics {
		delete(p.runtimeObjects, fmt.Sprintf("ConfigMap.%s-opssight-prometheus", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Service.%s-opssight-prometheus", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Service.%s-opssight-prometheus-exposed", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Deployment.%s-opssight-prometheus", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-prometheus-metrics", p.opsSight.Name))
	} else {
		// TODO need to check if the cluster is OpenShift
		if p.opsSight.Spec.Prometheus.Expose != "OPENSHIFT" {
			delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-prometheus-metrics", p.opsSight.Name))
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchCoreModelUI() error {
	// TODO need to check if the cluster is OpenShift
	if p.opsSight.Spec.Perceptor.Expose != "OPENSHIFT" {
		delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-core", p.opsSight.Name))
	}
	return nil
}
