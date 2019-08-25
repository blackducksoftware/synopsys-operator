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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlertSpec defines the desired state of Alert
type AlertSpec struct {
	// Set the namespace where you want to deploy alert. [dev-note]: this is strictly needed to handle cluster scope
	Namespace string `json:"namespace,omitempty"`
	// Set the version of the alert
	Version string `json:"version,omitempty"` // TODO:
	// Set the type for the service. [dev-note]: LOADBALANCER, NODEPORT allowed so far
	ExposeService string `json:"exposeService"`
	// Deploy alert in standalone mode. [dev-note]: this does not need to be a pointer
	StandAlone *bool `json:"standAlone"` // TODO: check with mphammer
	// Set Port for alert rc and service. [dev-note]: this does not need to be a pointer
	Port *int32 `json:"port"`
	// Base64Encoded string for ALERT_ENCRYPTION_PASSWORD. [dev-note]: this should be a pointer, also json should not be capitalized.
	// Also previously this was plain text, so TODO: migrate, either take this out from the api, or convert from plain to encoded
	EncryptionPassword string `json:"EncryptionPassword"`
	// Base64Encoded string for ALERT_ENCRYPTION_GLOBAL_SALT. [dev-note]: this should be a pointer, also json should not be capitalized.
	// Also previously this was plain text, so TODO: migrate, either take this out from the api, or convert from plain to encoded
	EncryptionGlobalSalt string `json:"EncryptionGlobalSalt"`
	// add data to secret as a slice of "key:base64encodedvalue". [dev-note]: another implementation to consider is a map or a field per variable
	Secrets []*string `json:"secrets,omitempty"` // TODO: not in previous API, make an api-change note
	// add data to environment variables config map as a slice of "key:value". [dev-note]: another implementation to consider is a map or a field per variable
	Environs []string `json:"environs,omitempty"`
	// enable or disable persistent storage. [dev-note]: this is a different implementation than Black Duck, for example missing volumeName
	PersistentStorage bool   `json:"persistentStorage"`
	PVCName           string `json:"pvcName"`
	PVCStorageClass   string `json:"pvcStorageClass"`
	PVCSize           string `json:"pvcSize"`
	// set min and max memory for alert rc. [dev-note]: again, different implementation than Black Duck, also why min and max set to be the same?
	AlertMemory string `json:"alertMemory,omitempty"` // TODO: make this consistent with Black Duck, how "sizes" are handled
	// set min and max memory for cfssl rc. [dev-note]: again, different implementation than Black Duck, also why min and max set to be the same?
	CfsslMemory string `json:"cfsslMemory,omitempty"` // TODO: make this consistent with Black Duck, how "sizes" are handled
	// set the desired state of the alert. [dev-note]: currently, only "STOP"
	DesiredState string `json:"desiredState,omitempty"`
	// slice of "key:value" for images, takes precedence over registryConfiguration [dev-note]: make explicit precedence over registryConfiguration
	ImageRegistries []string `json:"imageRegistries,omitempty"`
	// [dev-note]: this does not need to be a pointer
	RegistryConfiguration *RegistryConfiguration `json:"registryConfiguration,omitempty"` // TODO:
}

// AlertStatus defines the observed state of Alert
type AlertStatus struct {
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +kubebuilder:object:root=true

// Alert is the Schema for the alerts API
type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertSpec   `json:"spec,omitempty"`
	Status AlertStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertList contains a list of Alert
type AlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alert{}, &AlertList{})
}
