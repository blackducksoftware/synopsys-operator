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

// SizeSpec defines the desired state of Size
type SizeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodResources map[string]PodResource `json:"podResources"`
}

// SizeStatus defines the observed state of Size
type SizeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Size is the Schema for the sizes API
type Size struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SizeSpec   `json:"spec,omitempty"`
	Status SizeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SizeList contains a list of Size
type SizeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Size `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Size{}, &SizeList{})
}
