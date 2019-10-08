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
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

// UpdatePersistentVolumeClaim updates the size for pvc
func UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, size string) error {
	if size, err := resource.ParseQuantity(size); err == nil {
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = size
	}
	return nil
}

// patchStorageClass will iterate over the runtime objects and update the storage class
func patchStorageClass(obj map[string]runtime.Object, storageClass string) {
	if len(storageClass) > 0 {
		for k, v := range obj {
			switch v.(type) {
			case *appsv1beta1.StatefulSet:
				for claimTemplateIndex := range obj[k].(*appsv1beta1.StatefulSet).Spec.VolumeClaimTemplates {
					obj[k].(*appsv1beta1.StatefulSet).Spec.VolumeClaimTemplates[claimTemplateIndex].Spec.StorageClassName = &storageClass
				}
			case *appsv1.StatefulSet:
				for claimTemplateIndex := range obj[k].(*appsv1.StatefulSet).Spec.VolumeClaimTemplates {
					obj[k].(*appsv1.StatefulSet).Spec.VolumeClaimTemplates[claimTemplateIndex].Spec.StorageClassName = &storageClass
				}
			case *corev1.PersistentVolumeClaim:
				obj[k].(*corev1.PersistentVolumeClaim).Spec.StorageClassName = &storageClass
			}
		}
	}
}
