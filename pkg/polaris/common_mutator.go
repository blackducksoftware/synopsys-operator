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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func updateServiceCoreContainerImage(containers *[]corev1.Container, serviceName *string, imageDetails *ImageDetails) *[]corev1.Container {
	// pop the service core container
	for index, container := range *containers {
		if container.Name == *serviceName {
			// Update image details in container
			container.Image = strings.ReplaceAll(container.Image, "gcr.io/snps-swip-staging", imageDetails.Repository)

			// Replace the container object in list
			(*containers)[index] = container
		}
	}
	return containers
}

func PatchImageForService(imageDetails *ImageDetails, deployment *appsv1.Deployment) *appsv1.Deployment {
	deployment.Spec.Template.Spec.Containers = *updateServiceCoreContainerImage(
		&deployment.Spec.Template.Spec.Containers,
		&deployment.ObjectMeta.Name,
		imageDetails,
	)
	return deployment
}

func UpdateImagePullSecretsForDeployment(objects map[string]runtime.Object, deployments []string, imagePullSecret string) error {
	for _, deployment := range deployments {
		DeploymentUniqueID := "Deployment." + deployment
		deploymentRuntimeObject, ok := objects[DeploymentUniqueID]
		if !ok {
			return nil
		}
		deploymentInstance := deploymentRuntimeObject.(*appsv1.Deployment)
		deploymentInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdateImagePullSecretsForStatefulSets(objects map[string]runtime.Object, statefulsets []string, imagePullSecret string) error {
	for _, statefulset := range statefulsets {
		StatefulSetUniqueID := "StatefulSet." + statefulset
		statefulsetRuntimeObject, ok := objects[StatefulSetUniqueID]
		if !ok {
			return nil
		}
		statefulsetInstance := statefulsetRuntimeObject.(*appsv1.StatefulSet)
		statefulsetInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdateImagePullSecretsForJobs(objects map[string]runtime.Object, jobs []string, imagePullSecret string) error {
	for _, job := range jobs {
		JobUniqueID := "Job." + job
		jobRuntimeObject, ok := objects[JobUniqueID]
		if !ok {
			return nil
		}
		jobInstance := jobRuntimeObject.(*batchv1.Job)
		jobInstance.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}

func UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, size string) error {
	if size, err := resource.ParseQuantity(size); err == nil {
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = size
	}
	return nil
}
