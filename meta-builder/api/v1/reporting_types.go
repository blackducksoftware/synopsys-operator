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

// ReportingSpec defines the desired state of Reporting
type ReportingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Namespace             string                   `json:"namespace,omitempty"`
	EnvironmentName       string                   `json:"environment"`
	EnvironmentDNS        string                   `json:"environment_address"`
	ImagePullSecrets      string                   `json:"image_pull_secrets"`
	Version               string                   `json:"version"`
	PostgresDetails       ReportingPostgresDetails `json:"postgres"`
	ReportServiceSpec     ReportStorageSpec        `json:"report-storage,omitempty"`
	IsReportingStandalone bool                     `json:"isReportingStandalone"`
}

type ReportingPostgresDetails struct {
	Hostname string `json:"hostname"`
	Port     int32  `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ReportStorageSpec struct {
	Volume VolumeSpec `json:"volume"`
}

type VolumeSpec struct {
	Size string `json:"size"`
}

// ReportingStatus defines the observed state of Reporting
type ReportingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +kubebuilder:object:root=true

// Reporting is the Schema for the reportings API
type Reporting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReportingSpec   `json:"spec,omitempty"`
	Status ReportingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReportingList contains a list of Reporting
type ReportingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Reporting `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Reporting{}, &ReportingList{})
}
