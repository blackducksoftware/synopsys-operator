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

// AlertSpec defines the desired state of Alert
// Important: Run "make" to regenerate code after modifying this file
type AlertSpec struct {
	Namespace         string   `json:"namespace,omitempty"`
	Size              string   `json:"size"`
	AlertImage        string   `json:"alertImage,omitempty"`
	CfsslImage        string   `json:"cfsslImage,omitempty"`
	ExposeService     string   `json:"exposeService"`
	StandAlone        *bool    `json:"standAlone"`
	Port              *int32   `json:"port"`
	Environs          []string `json:"environs,omitempty"`
	Secrets           []string `json:"secrets,omitempty"`
	PersistentStorage bool     `json:"persistentStorage"`
	PVCName           string   `json:"pvcName"`
	PVCStorageClass   string   `json:"pvcStorageClass"`

	// Should be passed like: e.g "1300Mi"
	PVCSize               string                `json:"pvcSize"`
	AlertMemory           string                `json:"alertMemory,omitempty"`
	CfsslMemory           string                `json:"cfsslMemory,omitempty"`
	DesiredState          string                `json:"desiredState,omitempty"`
	ImageRegistries       []string              `json:"imageRegistries,omitempty"`
	RegistryConfiguration RegistryConfiguration `json:"registryConfiguration,omitempty"`
}

// RegistryConfiguration contains the registry configuration
type RegistryConfiguration struct {
	Registry    string   `json:"registry"`
	Namespace   string   `json:"namespace"`
	PullSecrets []string `json:"pullSecrets"`
}

// AlertStatus defines the observed state of Alert
// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// Important: Run "make" to regenerate code after modifying this file
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
