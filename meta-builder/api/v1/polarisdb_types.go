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

// PolarisDBSpec defines the desired state of PolarisDB
type PolarisDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Namespace string `json:"namespace,omitempty"`
	EnvironmentName string `json:"environment"`
	EnvironmentDNS string `json:"environment_address"`
	ImagePullSecrets string `json:"image_pull_secrets"`
}

// PolarisDBStatus defines the observed state of PolarisDB
type PolarisDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +kubebuilder:object:root=true

// PolarisDB is the Schema for the polarisdbs API
type PolarisDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisDBSpec   `json:"spec,omitempty"`
	Status PolarisDBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolarisDBList contains a list of PolarisDB
type PolarisDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisDB{}, &PolarisDBList{})
}
