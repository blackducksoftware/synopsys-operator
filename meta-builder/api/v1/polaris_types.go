/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisSpec defines the desired state of Polaris
type PolarisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Namespace        string         `json:"namespace,omitempty"`
	EnvironmentName  string         `json:"environment"`
	EnvironmentDNS   string         `json:"environment_address"`
	ImagePullSecrets string         `json:"image_pull_secrets"`
	AuthServerSpec   AuthServerSpec `json:"auth_server,omitempty"`
}

type AuthServerSpec struct {
	Replicas      *int32        `json:"replicas,omitempty"`
	ResourcesSpec ResourcesSpec `json:"resources,omitempty"`
}

type ResourcesSpec struct {
	RequestsSpec RequestsSpec `json:"requests,omitempty"`
	LimitsSpec   LimitsSpec   `json:"limits,omitempty"`
}

type RequestsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type LimitsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// PolarisStatus defines the observed state of Polaris
type PolarisStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +kubebuilder:object:root=true

// Polaris is the Schema for the polaris API
type Polaris struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisSpec   `json:"spec,omitempty"`
	Status PolarisStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolarisList contains a list of Polaris
type PolarisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Polaris `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Polaris{}, &PolarisList{})
}
