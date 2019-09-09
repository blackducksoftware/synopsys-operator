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

package util

import (
	"fmt"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GenResourceRequirementsFromContainerSize converts ContainerSize to  ResourceRequirements
func GenResourceRequirementsFromContainerSize(containerSize synopsysv1.ContainerSize) (*corev1.ResourceRequirements, error) {
	req := &corev1.ResourceRequirements{}
	if containerSize.MinCPU != nil {
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%d", *containerSize.MinCPU))
		if err != nil {
			return nil, err
		}
		if req.Requests == nil {
			req.Requests = make(map[corev1.ResourceName]resource.Quantity)
		}
		req.Requests[corev1.ResourceCPU] = quantity
	}

	if containerSize.MaxCPU != nil {
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%d", *containerSize.MaxCPU))
		if err != nil {
			return nil, err
		}
		if req.Limits == nil {
			req.Limits = make(map[corev1.ResourceName]resource.Quantity)
		}
		req.Limits[corev1.ResourceCPU] = quantity
	}

	if containerSize.MinMem != nil {
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%dM", *containerSize.MinMem))
		if err != nil {
			return nil, err
		}
		if req.Requests == nil {
			req.Requests = make(map[corev1.ResourceName]resource.Quantity)
		}
		req.Requests[corev1.ResourceMemory] = quantity
	}

	if containerSize.MaxMem != nil {
		quantity, err := resource.ParseQuantity(fmt.Sprintf("%dM", *containerSize.MaxMem))
		if err != nil {
			return nil, err
		}
		if req.Limits == nil {
			req.Limits = make(map[corev1.ResourceName]resource.Quantity)
		}
		req.Limits[corev1.ResourceMemory] = quantity
	}
	return req, nil
}
