/*
Copyright (C) 2018 Synopsys, Inc.

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
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Gr will be CRD gr definition
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Gr struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	Spec               GrSpec   `json:"spec"`
	Status             GrStatus `json:"status,omitempty"`
}

// GrSpec will be CRD Gr definition's Spec
type GrSpec struct {
	Namespace    string `json:"namespace"`
	StorageClass string `json:"storageClass"`
	IngressClass string `json:"ingressClass"`
	IngressHost  string `json:"ingressHost"`
}

// GrStatus will be CRD Gr definition's Status
type GrStatus struct {
	State         string            `json:"state"`
	IP            string            `json:"ip"`
	PVCVolumeName map[string]string `json:"pvcVolumeName,omitempty"`
	Fqdn          string            `json:"fqdn,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
}

// GrList will store the list of Grs
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []Gr `json:"items"`
}
