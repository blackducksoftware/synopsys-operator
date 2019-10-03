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
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"regexp"
)

// UpdatePersistentVolumeClaim updates the size for pvc
func UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, size string) error {
	if size, err := resource.ParseQuantity(size); err == nil {
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = size
	}
	return nil
}

func updateRegistry(obj map[string]runtime.Object, registry string) (map[string]runtime.Object, error) {
	for k, v := range obj {
		switch v.(type) {
		case *extensionv1beta1.Deployment:
			if err := updateContainersImage(obj[k].(*extensionv1beta1.Deployment).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *appsv1beta1.StatefulSet:
			if err := updateContainersImage(obj[k].(*appsv1beta1.StatefulSet).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *appsv1.DaemonSet:
			if err := updateContainersImage(obj[k].(*appsv1.DaemonSet).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *appsv1.Deployment:
			if err := updateContainersImage(obj[k].(*appsv1.Deployment).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *appsv1.StatefulSet:
			if err := updateContainersImage(obj[k].(*appsv1.StatefulSet).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *batchv1.Job:
			if err := updateContainersImage(obj[k].(*batchv1.Job).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *corev1.ReplicationController:
			if err := updateContainersImage(obj[k].(*corev1.ReplicationController).Spec.Template.Spec, registry); err != nil {
				return nil, err
			}
		case *corev1.Pod:
			if err := updateContainersImage(obj[k].(*corev1.Pod).Spec, registry); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func updateContainersImage(podSpec corev1.PodSpec, registry string) error {
	for containerIndex, container := range podSpec.Containers {
		newImage, err := generateNewImage(container.Image, registry)
		if err != nil {
			return err
		}
		podSpec.Containers[containerIndex].Image = newImage
	}

	for initContainerIndex, initContainer := range podSpec.InitContainers {
		newImage, err := generateNewImage(initContainer.Image, registry)
		if err != nil {
			return err
		}
		podSpec.InitContainers[initContainerIndex].Image = newImage
	}
	return nil
}

func generateNewImage(currentImage string, registry string) (string, error) {
	imageTag, err := getImageAndTag(currentImage)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", registry, imageTag), nil
}

func getImageAndTag(image string) (string, error) {
	r := regexp.MustCompile(`^(|.*/)([a-zA-Z_0-9-.:]+)$`)
	groups := r.FindStringSubmatch(image)
	if len(groups) < 3 && len(groups[2]) == 0 {
		return "", fmt.Errorf("couldn't find image and tags in [%s]", image)
	}
	return groups[2], nil
}
