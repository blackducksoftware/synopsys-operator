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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/util"

	"github.com/go-logr/logr"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchOpsSight(client client.Client, scheme *runtime.Scheme, opsSight *synopsysv1.OpsSight, runtimeObjects map[string]runtime.Object, log logr.Logger, isOpenShift bool) map[string]runtime.Object {
	log.Info("patching")
	patcher := OpsSightPatcher{
		Client:         client,
		scheme:         scheme,
		opsSight:       opsSight,
		runtimeObjects: runtimeObjects,
		log:            log,
		isOpenShift:    isOpenShift,
	}
	return patcher.patch()
}

// OpsSightPatcher applies the patches to OpsSight
type OpsSightPatcher struct {
	client.Client
	scheme         *runtime.Scheme
	opsSight       *synopsysv1.OpsSight
	runtimeObjects map[string]runtime.Object
	log            logr.Logger
	isOpenShift    bool
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
		// p.patchWithSize,
		p.patchReplicas,
		p.patchAddRegistryAuth,
		p.patchSecretData,
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
	default:
		delete(p.runtimeObjects, id)
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
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRole.%s-opssight-scanner", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("ClusterRoleBinding.%s-opssight-scanner", p.opsSight.Name))
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
		delete(p.runtimeObjects, fmt.Sprintf("Deployment.%s-opssight-prometheus", p.opsSight.Name))
		delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-prometheus-metrics", p.opsSight.Name))
	} else {
		// TODO need to check if the cluster is OpenShift
		if !p.isOpenShift || p.opsSight.Spec.Prometheus.Expose != "OPENSHIFT" {
			delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-prometheus-metrics", p.opsSight.Name))
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchCoreModelUI() error {
	// TODO need to check if the cluster is OpenShift
	if !p.isOpenShift || p.opsSight.Spec.Perceptor.Expose != "OPENSHIFT" {
		delete(p.runtimeObjects, fmt.Sprintf("Route.%s-opssight-core", p.opsSight.Name))
	}
	return nil
}

// TODO: common with Alert
func (p *OpsSightPatcher) patchWithSize() error {
	var size synopsysv1.Size
	if len(p.opsSight.Spec.Size) > 0 {
		if err := p.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: p.opsSight.Namespace,
			Name:      strings.ToLower(p.opsSight.Spec.Size),
		}, &size); err != nil {

			if !apierrs.IsNotFound(err) {
				return err
			}
			if apierrs.IsNotFound(err) {
				return fmt.Errorf("opsSight instance [%s] is configured to use a Size [%s] that doesn't exist", p.opsSight.Spec.Namespace, p.opsSight.Spec.Size)
			}
		}

		for _, v := range p.runtimeObjects {
			switch v.(type) {
			case *corev1.ReplicationController:
				componentName, ok := v.(*corev1.ReplicationController).GetLabels()["component"]
				if !ok {
					return fmt.Errorf("component name is missing in %s", v.(*corev1.ReplicationController).Name)
				}

				sizeConf, ok := size.Spec.PodResources[componentName]
				if !ok {
					return fmt.Errorf("opsSight instance [%s] is configured to use a Size [%s] but the size doesn't contain an entry for [%s]", p.opsSight.Spec.Namespace, p.opsSight.Spec.Size, v.(*corev1.ReplicationController).Name)
				}
				v.(*corev1.ReplicationController).Spec.Replicas = func(i int) *int32 { j := int32(i); return &j }(sizeConf.Replica)
				for containerIndex, container := range v.(*corev1.ReplicationController).Spec.Template.Spec.Containers {
					containerConf, ok := sizeConf.ContainerLimit[container.Name]
					if !ok {
						return fmt.Errorf("opsSight instance [%s] is configured to use a Size [%s]. The size oesn't contain an entry for pod [%s] container [%s]", p.opsSight.Spec.Namespace, p.opsSight.Spec.Size, v.(*corev1.ReplicationController).Name, container.Name)
					}
					resourceRequirements, err := controllers_utils.GenResourceRequirementsFromContainerSize(containerConf)
					if err != nil {
						return err
					}
					v.(*corev1.ReplicationController).Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements
				}
			case *appsv1.Deployment:
				componentName, ok := v.(*appsv1.Deployment).GetLabels()["component"]
				if !ok {
					return fmt.Errorf("component name is missing in %s", v.(*appsv1.Deployment).Name)
				}

				sizeConf, ok := size.Spec.PodResources[componentName]
				if !ok {
					return fmt.Errorf("opsSight instance [%s] is configured to use a Size [%s] but the size doesn't contain an entry for [%s]", p.opsSight.Spec.Namespace, p.opsSight.Spec.Size, v.(*corev1.ReplicationController).Name)
				}
				v.(*appsv1.Deployment).Spec.Replicas = func(i int) *int32 { j := int32(i); return &j }(sizeConf.Replica)
				for containerIndex, container := range v.(*appsv1.Deployment).Spec.Template.Spec.Containers {
					containerConf, ok := sizeConf.ContainerLimit[container.Name]
					if !ok {
						return fmt.Errorf("opsSight instance [%s] is configured to use a Size [%s]. The size oesn't contain an entry for pod [%s] container [%s]", p.opsSight.Spec.Namespace, p.opsSight.Spec.Size, v.(*corev1.ReplicationController).Name, container.Name)
					}
					resourceRequirements, err := controllers_utils.GenResourceRequirementsFromContainerSize(containerConf)
					if err != nil {
						return err
					}
					v.(*appsv1.Deployment).Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements
				}
			}
		}
	}
	return nil
}

// TODO: common with Alert
func (p *OpsSightPatcher) patchReplicas() error {
	for _, v := range p.runtimeObjects {
		switch v.(type) {
		case *corev1.ReplicationController:
			switch strings.ToUpper(p.opsSight.Spec.DesiredState) {
			case "STOP":
				v.(*corev1.ReplicationController).Spec.Replicas = func(i int32) *int32 { return &i }(0)
			}
		case *appsv1.Deployment:
			switch strings.ToUpper(p.opsSight.Spec.DesiredState) {
			case "STOP":
				v.(*appsv1.Deployment).Spec.Replicas = func(i int32) *int32 { return &i }(0)
			}
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchAddRegistryAuth() error {
	// if OpenShift, get the registry auth informations
	if !p.isOpenShift {
		return nil
	}

	internalRegistries := []*string{}

	// Adding default image registry routes
	routes := map[string]string{"default": "docker-registry", "openshift-image-registry": "image-registry"}
	for namespace, name := range routes {
		route := &routev1.Route{}
		err := p.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, route)
		if err != nil {
			continue
		}
		internalRegistries = append(internalRegistries, &route.Spec.Host)
		routeHostPort := fmt.Sprintf("%s:443", route.Spec.Host)
		internalRegistries = append(internalRegistries, &routeHostPort)
	}

	// Adding default OpenShift internal Docker/image registry service
	labelSelectors := []string{"docker-registry=default", "router in (router,router-default)"}
	for _, labelSelector := range labelSelectors {
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			p.log.Error(err, "label selector", "selector", labelSelector)
			continue
		}
		registrySvcs := &corev1.ServiceList{}
		err = p.Client.List(context.TODO(), registrySvcs, MatchingLabels{LabelSelector: selector})
		if err != nil {
			continue
		}
		for _, registrySvc := range registrySvcs.Items {
			if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
				for _, port := range registrySvc.Spec.Ports {
					clusterIPSvc := fmt.Sprintf("%s:%d", registrySvc.Spec.ClusterIP, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvc)
					clusterIPSvcPort := fmt.Sprintf("%s.%s.svc:%d", registrySvc.Name, registrySvc.Namespace, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvcPort)
				}
			}
		}
	}

	// read the operator service account token to set it as password to pull the image from an OpenShift internal Docker registry
	file, err := controllers_utils.ReadFileData("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		p.log.Error(err, "unable to read the service account token file")
	} else {
		for _, internalRegistry := range internalRegistries {
			p.opsSight.Spec.ScannerPod.ImageFacade.InternalRegistries = append(p.opsSight.Spec.ScannerPod.ImageFacade.InternalRegistries, &synopsysv1.RegistryAuth{URL: *internalRegistry, User: "admin", Password: string(file)})
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchSecretData() error {
	blackDuckHosts := make(map[string]*synopsysv1.Host)
	// adding External Black Duck credentials
	for _, host := range p.opsSight.Spec.Blackduck.ExternalHosts {
		blackDuckHosts[host.Domain] = host
	}

	// adding Internal Black Duck credentials
	opsSightUpdater := NewOpsSightBlackDuckReconciler(p.Client, p.scheme, p.log)
	blackDuckType := p.opsSight.Spec.Blackduck.BlackduckSpec.Type
	blackDuckPassword, err := controllers_utils.Base64Decode(p.opsSight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return fmt.Errorf("unable to decode Black Duck Password due to %+v", err)
	}

	allBlackDucks := opsSightUpdater.GetAllBlackDucks(blackDuckType, blackDuckPassword)
	blackDuckMergedHosts := opsSightUpdater.AppendBlackDuckSecrets(blackDuckHosts, p.opsSight.Status.InternalHosts, allBlackDucks)

	// add internal hosts to status
	p.opsSight.Status.InternalHosts = opsSightUpdater.AppendBlackDuckHosts(p.opsSight.Status.InternalHosts, allBlackDucks)

	for _, v := range p.runtimeObjects {
		switch v.(type) {
		case *corev1.Secret:
			// marshal the blackduck credentials to bytes
			bytes, err := json.Marshal(blackDuckMergedHosts)
			if err != nil {
				return fmt.Errorf("unable to marshal Black Duck passwords due to %+v", err)
			}
			v.(*corev1.Secret).Data[p.opsSight.Spec.Blackduck.ConnectionsEnvironmentVariableName] = bytes

			// adding Secured registries credentials
			securedRegistries := make(map[string]*synopsysv1.RegistryAuth)
			for _, internalRegistry := range p.opsSight.Spec.ScannerPod.ImageFacade.InternalRegistries {
				securedRegistries[internalRegistry.URL] = internalRegistry
			}
			// marshal the Secured registries credentials to bytes
			bytes, err = json.Marshal(securedRegistries)
			if err != nil {
				return fmt.Errorf("unable to marshal secured registries due to %+v", err)
			}
			v.(*corev1.Secret).Data["securedRegistries.json"] = bytes
		}
	}

	return nil
}

func (p *OpsSightPatcher) patchPodProcessorSCC() error {
	if !p.isOpenShift {
		return nil
	}

	for _, v := range p.runtimeObjects {
		switch v.(type) {
		case *rbacv1.ClusterRole:
			if v.(*rbacv1.ClusterRole).GetName() == fmt.Sprintf("%s-opssight-pod-processor", p.opsSight.Name) {
				rule := rbacv1.PolicyRule{
					Verbs:         []string{"use"},
					APIGroups:     []string{"security.openshift.io"},
					Resources:     []string{"securitycontextconstraints"},
					ResourceNames: []string{"privileged"},
				}
				v.(*rbacv1.ClusterRole).Rules = append(v.(*rbacv1.ClusterRole).Rules, rule)
			}
		}
	}
	return nil
}

func (p *OpsSightPatcher) patchScannerSCC() error {
	if !p.isOpenShift || p.opsSight.Spec.ScannerPod.ImageFacade.ImagePullerType == "skopeo" || !p.opsSight.Spec.Perceiver.EnableImagePerceiver {
		return nil
	}

	for _, v := range p.runtimeObjects {
		switch v.(type) {
		case *rbacv1.ClusterRole:
			if v.(*rbacv1.ClusterRole).GetName() == fmt.Sprintf("%s-opssight-scanner", p.opsSight.Name) {
				rule := rbacv1.PolicyRule{
					Verbs:         []string{"use"},
					APIGroups:     []string{"security.openshift.io"},
					Resources:     []string{"securitycontextconstraints"},
					ResourceNames: []string{"privileged"},
				}
				v.(*rbacv1.ClusterRole).Rules = append(v.(*rbacv1.ClusterRole).Rules, rule)
			}
		}
	}
	return nil
}
