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

package v2

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hub will be CRD hub definition
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Hub struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	View               HubView   `json:"view"`
	Spec               HubSpec   `json:"spec"`
	Status             HubStatus `json:"status,omitempty"`
}

// HubView will be used to populate information for the Hub UI.
type HubView struct {
	Clones           map[string]string `json:"clones"`
	StorageClasses   map[string]string `json:"storageClasses"`
	CertificateNames []string          `json:"certificateNames"`
	Environs         []string          `json:"environs"`
	ContainerTags    []string          `json:"containerTags"`
	Version          string            `json:"version"`
}

// HubSpec will be CRD Hub definition's Spec
type HubSpec struct {
	Namespace         string                   `json:"namespace"`
	Size              string                   `json:"size"`
	DbPrototype       string                   `json:"dbPrototype,omitempty"`
	ExternalPostgres  PostgresExternalDBConfig `json:"externalPostgres"`
	PVCStorageClass   string                   `json:"pvcStorageClass,omitempty"`
	LivenessProbes    bool                     `json:"livenessProbes"`
	ScanType          string                   `json:"scanType,omitempty"`
	PersistentStorage bool                     `json:"persistentStorage"`
	PVC               []PVC                    `json:"pvc,omitempty"`
	CertificateName   string                   `json:"certificateName"`
	Certificate       string                   `json:"certificate,omitempty"`
	CertificateKey    string                   `json:"certificateKey,omitempty"`
	ProxyCertificate  string                   `json:"proxyCertificate,omitempty"`
	HubType           string                   `json:"hubType,omitempty"`
	State             string                   `json:"state"`
	Environs          []string                 `json:"environs,omitempty"`
	ImageRegistries   []string                 `json:"imageRegistries,omitempty"`
	ImageUIDMap       map[string]int64         `json:"imageUidMap,omitempty"`
}

// Environs will hold the list of Environment variables
type Environs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PVC will contain the specifications of the different PVC.
// This will overwrite the default claim configuration
type PVC struct {
	Name         string `json:"name"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}

// PostgresExternalDBConfig contain the external database configuration
type PostgresExternalDBConfig struct {
	PostgresHost          string `json:"postgresHost"`
	PostgresPort          int    `json:"postgresPort"`
	PostgresAdmin         string `json:"postgresAdmin"`
	PostgresUser          string `json:"postgresUser"`
	PostgresSsl           bool   `json:"postgresSsl"`
	PostgresAdminPassword string `json:"postgresAdminPassword"`
	PostgresUserPassword  string `json:"postgresUserPassword"`
}

// HubStatus will be CRD Hub definition's Status
type HubStatus struct {
	State         string            `json:"state"`
	IP            string            `json:"ip"`
	PVCVolumeName map[string]string `json:"pvcVolumeName,omitempty"`
	Fqdn          string            `json:"fqdn,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
}

// HubList will store the list of Hubs
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HubList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []Hub `json:"items"`
}
