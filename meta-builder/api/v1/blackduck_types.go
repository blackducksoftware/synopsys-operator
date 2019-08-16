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

// BlackduckSpec defines the desired state of Blackduck
type BlackduckSpec struct {
	Namespace             string                    `json:"namespace"`
	Size                  string                    `json:"size"`
	Version               string                    `json:"version"`
	ExposeService         string                    `json:"exposeService"`
	DbPrototype           string                    `json:"dbPrototype,omitempty"`
	ExternalPostgres      *PostgresExternalDBConfig `json:"externalPostgres,omitempty"`
	PVCStorageClass       string                    `json:"pvcStorageClass,omitempty"`
	LivenessProbes        bool                      `json:"livenessProbes"`
	ScanType              string                    `json:"scanType,omitempty"`
	PersistentStorage     bool                      `json:"persistentStorage"`
	PVC                   []PVC                     `json:"pvc,omitempty"`
	CertificateName       string                    `json:"certificateName"`
	Certificate           string                    `json:"certificate,omitempty"`
	CertificateKey        string                    `json:"certificateKey,omitempty"`
	ProxyCertificate      string                    `json:"proxyCertificate,omitempty"`
	AuthCustomCA          string                    `json:"authCustomCa"`
	Type                  string                    `json:"type,omitempty"`
	DesiredState          string                    `json:"desiredState"`
	Environs              []string                  `json:"environs,omitempty"`
	ImageRegistries       []string                  `json:"imageRegistries,omitempty"`
	LicenseKey            string                    `json:"licenseKey,omitempty"`
	RegistryConfiguration RegistryConfiguration     `json:"registryConfiguration,omitempty"`
	AdminPassword         string                    `json:"adminPassword"`
	UserPassword          string                    `json:"userPassword"`
	PostgresPassword      string                    `json:"postgresPassword"`
	//NodeAffinities        map[string][]NodeAffinity `json:"nodeAffinities,omitempty"`
}

type RegistryConfiguration struct {
	Registry    string   `json:"registry"`
	Namespace   string   `json:"namespace"`
	PullSecrets []string `json:"pullSecrets"`
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
	VolumeName   string `json:"volumeName,omitempty"`
}

// NodeAffinity will contain the specifications of a node affinity
// TODO: currently, keeping it simple, but can be modified in the future to take in complex scenarios
type NodeAffinity struct {
	AffinityType string   `json:"affinityType"`
	Key          string   `json:"key"`
	Op           string   `json:"op"`
	Values       []string `json:"values"`
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

// BlackduckStatus defines the observed state of Blackduck
type BlackduckStatus struct {
	State         string            `json:"state"`
	IP            string            `json:"ip"`
	PVCVolumeName map[string]string `json:"pvcVolumeName,omitempty"`
	Fqdn          string            `json:"fqdn,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true

// Blackduck is the Schema for the blackducks API
type Blackduck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlackduckSpec   `json:"spec,omitempty"`
	Status BlackduckStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BlackduckList contains a list of Blackduck
type BlackduckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Blackduck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Blackduck{}, &BlackduckList{})
}
