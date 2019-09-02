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
	Namespace              string                 `json:"namespace,omitempty"`
	EnvironmentName        string                 `json:"environment"`
	EnvironmentDNS         string                 `json:"environment_address"`
	ImagePullSecrets       string                 `json:"image_pull_secrets"`
	Version                string                 `json:"version"`
	SMTPDetails            SMTPDetails            `json:"smtp_details"`
	PostgresInstanceType   string                 `json:"postgres_instance_type"`
	PostgresStorageDetails PostgresStorageDetails `json:"postgres_storage_details,omitempty"`
	PostgresDetails        PostgresDetails        `json:"postgres"`
	EventstoreDetails      EventstoreDetails      `json:"eventstore_details,omitempty"`
	UploadServerDetails    UploadServerDetails    `json:"upload_server_details,omitempty"`
}

type EventstoreDetails struct {
	Replicas    *int32 `json:"replicas,omitempty"`
	StorageSize string `json:"storage_size,omitempty"`
}

type UploadServerDetails struct {
	Replicas      *int32        `json:"replicas,omitempty"`
	ResourcesSpec ResourcesSpec `json:"resources,omitempty"`
	Storage       Storage       `json:"storage,omitempty"`
}

type Storage struct {
	Type         string  `json:"type,omitempty"`
	StorageSize  string  `json:"size,omitempty"`
	StorageClass *string `json:"storage_class,omitempty"`
}

type SMTPDetails struct {
	Host     string `json:"host"`
	Port     *int32 `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type PostgresDetails struct {
	Host     string `json:"host"`
	Port     *int32 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PostgresStorageDetails struct {
	StorageSize  string  `json:"size"`
	StorageClass *string `json:"storage_class"`
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
