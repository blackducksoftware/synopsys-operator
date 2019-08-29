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

// RegistryAuth will store the Secured Registries
type RegistryAuth struct {
	URL      string `json:"Url"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Host configures the Black Duck hosts
type Host struct {
	Scheme              string `json:"scheme"`
	Domain              string `json:"domain"` // it can be domain name or ip address
	Port                int    `json:"port"`
	User                string `json:"user"`
	Password            string `json:"password"`
	ConcurrentScanLimit int    `json:"concurrentScanLimit"`
}

// Blackducks stores the Black Duck instances
type Blackducks struct {
	ExternalHosts                      []*Host `json:"externalHosts,omitempty"`
	ConnectionsEnvironmentVariableName string  `json:"connectionsEnvironmentVariableName"`
	BlackduckPassword                  string  `json:"blackduckPassword"`
	TLSVerification                    bool    `json:"tlsVerification"`

	// Auto scaling parameters
	InitialCount                       int            `json:"initialCount"`
	MaxCount                           int            `json:"maxCount"`
	DeleteBlackduckThresholdPercentage int            `json:"deleteBlackduckThresholdPercentage"`
	BlackduckSpec                      *BlackduckSpec `json:"blackduckSpec"`
}

// Perceptor stores the Perceptor configuration
type Perceptor struct {
	CheckForStalledScansPauseHours int    `json:"checkForStalledScansPauseHours"`
	StalledScanClientTimeoutHours  int    `json:"stalledScanClientTimeoutHours"`
	ModelMetricsPauseSeconds       int    `json:"modelMetricsPauseSeconds"`
	UnknownImagePauseMilliseconds  int    `json:"unknownImagePauseMilliseconds"`
	ClientTimeoutMilliseconds      int    `json:"clientTimeoutMilliseconds"`
	Expose                         string `json:"expose"`
}

// ScannerPod stores the Perceptor scanner and Image Facade configuration
type ScannerPod struct {
	Scanner        *Scanner     `json:"scanner"`
	ImageFacade    *ImageFacade `json:"imageFacade"`
	ReplicaCount   int          `json:"scannerReplicaCount"`
	ImageDirectory string       `json:"imageDirectory"`
}

// Scanner stores the Perceptor scanner configuration
type Scanner struct {
	ClientTimeoutSeconds int `json:"clientTimeoutSeconds"`
}

// ImageFacade stores the Image Facade configuration
type ImageFacade struct {
	InternalRegistries []*RegistryAuth `json:"internalRegistries"`
	ImagePullerType    string          `json:"imagePullerType"`
}

// PodPerceiver stores the Pod Perceiver configuration
type PodPerceiver struct {
	NamespaceFilter string `json:"namespaceFilter,omitempty"`
}

// Perceiver stores the Perceiver configuration
type Perceiver struct {
	EnableImagePerceiver      bool          `json:"enableImagePerceiver"`
	EnablePodPerceiver        bool          `json:"enablePodPerceiver"`
	PodPerceiver              *PodPerceiver `json:"podPerceiver,omitempty"`
	AnnotationIntervalSeconds int           `json:"annotationIntervalSeconds"`
	DumpIntervalMinutes       int           `json:"dumpIntervalMinutes"`
}

// Prometheus container definition
type Prometheus struct {
	Expose string `json:"expose"`
}

// OpsSightSpec defines the desired state of OpsSight
type OpsSightSpec struct {
	// OpsSight
	Namespace  string      `json:"namespace"`
	Size       string      `json:"size"`
	Version    string      `json:"version,omitempty"`
	IsUpstream bool        `json:"isUpstream"`
	Perceptor  *Perceptor  `json:"perceptor"`
	ScannerPod *ScannerPod `json:"scannerPod"`
	Perceiver  *Perceiver  `json:"perceiver"`

	// CPU and memory configurations
	DefaultCPU string `json:"defaultCpu,omitempty"` // Example: "300m"
	DefaultMem string `json:"defaultMem,omitempty"` // Example: "1300Mi"
	ScannerCPU string `json:"scannerCpu,omitempty"` // Example: "300m"
	ScannerMem string `json:"scannerMem,omitempty"` // Example: "1300Mi"

	// Log level
	LogLevel string `json:"logLevel,omitempty"`

	// Metrics
	EnableMetrics bool        `json:"enableMetrics"`
	Prometheus    *Prometheus `json:"prometheus,omitempty"`

	// Black Duck
	Blackduck *Blackducks `json:"blackduck"`

	DesiredState string `json:"desiredState"`

	// Image handler
	ImageRegistries       []string              `json:"imageRegistries,omitempty"`
	RegistryConfiguration RegistryConfiguration `json:"registryConfiguration,omitempty"`
}

// OpsSightStatus defines the observed state of OpsSight
type OpsSightStatus struct {
	State         string  `json:"state"`
	ErrorMessage  string  `json:"errorMessage"`
	InternalHosts []*Host `json:"internalHosts"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// OpsSight is the Schema for the opssights API
type OpsSight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpsSightSpec   `json:"spec,omitempty"`
	Status OpsSightStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpsSightList contains a list of OpsSight
type OpsSightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpsSight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpsSight{}, &OpsSightList{})
}
