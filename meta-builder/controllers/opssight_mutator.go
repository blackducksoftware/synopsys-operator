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
	"github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/controllers_utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchOpsSight(client client.Client, opsSight *synopsysv1.OpsSight, objects map[string]runtime.Object) map[string]runtime.Object {
	patcher := OpsSightPatcher{
		Client:   client,
		opsSight: opsSight,
		objects:  objects,
	}
	return patcher.patch()
}

// OpsSightPatcher applies the patches to OpsSight
type OpsSightPatcher struct {
	client.Client
	opsSight *synopsysv1.OpsSight
	objects  map[string]runtime.Object
}

func (p *OpsSightPatcher) patch() map[string]runtime.Object {
	// TODO JD: Patching this way is costly. Consider iterating over the objects only once
	// and apply the necessary changes

	patches := []func() error{
		p.patchNamespace,
		p.patchImages,
		p.patchOpsSightCoreModelExposeService,
		p.patchOpsSightPrometheusExposeService,
	}
	for _, f := range patches {
		err := f()
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	return p.objects
}

func (p *OpsSightPatcher) patchOpsSightCoreModelExposeService() error {
	// TODO use constants
	id := fmt.Sprintf("Service.%s-opssight-core-exposed", p.opsSight.Name)
	runtimeObject, ok := p.objects[id]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.opsSight.Spec.Perceptor.Expose) {
	case "LOADBALANCER":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
	case "NODEPORT":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
	default:
		delete(p.objects, id)
	}

	return nil
}

func (p *OpsSightPatcher) patchOpsSightPrometheusExposeService() error {
	// TODO use constants
	id := fmt.Sprintf("Service.%s-opssight-prometheus-exposed", p.opsSight.Name)
	runtimeObject, ok := p.objects[id]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.opsSight.Spec.Prometheus.Expose) {
	case "LOADBALANCER":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeLoadBalancer
	case "NODEPORT":
		runtimeObject.(*corev1.Service).Spec.Type = corev1.ServiceTypeNodePort
	default:
		delete(p.objects, id)
	}

	return nil
}

func (p *OpsSightPatcher) patchImages() error {
	if len(p.opsSight.Spec.RegistryConfiguration.Registry) > 0 || len(p.opsSight.Spec.ImageRegistries) > 0 {
		for _, v := range p.objects {
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

func (p *OpsSightPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.objects {
		if kind, err := accessor.Kind(runtimeObject); err == nil && !(kind == "ClusterRole" || kind == "ClusterRoleBinding") {
			accessor.SetNamespace(runtimeObject, p.opsSight.Spec.Namespace)
		}
	}
	return nil
}
