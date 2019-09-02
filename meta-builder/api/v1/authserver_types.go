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

// AuthServerSpec defines the desired state of AuthServer
type AuthServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Namespace        string `json:"namespace,omitempty"`
	EnvironmentName  string `json:"environment"`
	EnvironmentDNS   string `json:"environment_address"`
	ImagePullSecrets string `json:"image_pull_secrets"`
	Version          string `json:"version"`
}

// AuthServerStatus defines the observed state of AuthServer
type AuthServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +kubebuilder:object:root=true

// AuthServer is the Schema for the authservers API
type AuthServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthServerSpec   `json:"spec,omitempty"`
	Status AuthServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AuthServerList contains a list of AuthServer
type AuthServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthServer{}, &AuthServerList{})
}
