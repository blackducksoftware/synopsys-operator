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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Size defines Size configuration
type Size struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SizeSpec `json:"spec"`
}

// SizeSpec is a specification for a template size
type SizeSpec struct {
	PodResources map[string]PodResource `json:"podResources"`
}

// PodResource defines the pod resource configuration
type PodResource struct {
	Replica        int                      `json:"replica"`
	ContainerLimit map[string]ContainerSize `json:"containerLimit"`
}

// ContainerSize refers to container size configuration
type ContainerSize struct {
	MinCPU *int32 `json:"minCpu"`
	MaxCPU *int32 `json:"maxCpu"`
	MinMem *int32 `json:"minMem"`
	MaxMem *int32 `json:"maxMem"`
}

// SizeList is a list of Size resources
type SizeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Size `json:"items"`
}
