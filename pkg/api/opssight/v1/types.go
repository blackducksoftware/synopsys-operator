/*
Copyright (C) 2018 Synopsys, Inc.

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
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpsSight is a specification for a OpsSight resource
type OpsSight struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpsSightSpec   `json:"spec"`
	Status OpsSightStatus `json:"status,omitempty"`
}

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

// Blackduck ...
type Blackduck struct {
	ExternalHosts                      []*Host `json:"externalHosts"`
	ConnectionsEnvironmentVariableName string  `json:"connectionsEnvironmentVariableName"`
	TLSVerification                    bool    `json:"tlsVerification"`

	// Auto scaling parameters
	InitialCount                       int               `json:"initialCount"`
	MaxCount                           int               `json:"maxCount"`
	DeleteBlackDuckThresholdPercentage int               `json:"deleteBlackDuckThresholdPercentage"`
	BlackduckSpec                      *v1.BlackduckSpec `json:"blackduckSpec"`
}

// Perceptor ...
type Perceptor struct {
	Name                           string `json:"name"`
	Image                          string `json:"image"`
	Port                           int    `json:"port"`
	CheckForStalledScansPauseHours int    `json:"checkForStalledScansPauseHours"`
	StalledScanClientTimeoutHours  int    `json:"stalledScanClientTimeoutHours"`
	ModelMetricsPauseSeconds       int    `json:"modelMetricsPauseSeconds"`
	UnknownImagePauseMilliseconds  int    `json:"unknownImagePauseMilliseconds"`
	ClientTimeoutMilliseconds      int    `json:"clientTimeoutMilliseconds"`
}

// ScannerPod ...
type ScannerPod struct {
	Name           string       `json:"name"`
	Scanner        *Scanner     `json:"scanner"`
	ImageFacade    *ImageFacade `json:"imageFacade"`
	ReplicaCount   int          `json:"scannerReplicaCount"`
	ImageDirectory string       `json:"imageDirectory"`
}

// Scanner ...
type Scanner struct {
	Name                 string `json:"name"`
	Image                string `json:"image"`
	Port                 int    `json:"port"`
	ClientTimeoutSeconds int    `json:"clientTimeoutSeconds"`
}

// ImageFacade ...
type ImageFacade struct {
	Name               string          `json:"name"`
	Image              string          `json:"image"`
	Port               int             `json:"port"`
	InternalRegistries []*RegistryAuth `json:"internalRegistries"`
	ImagePullerType    string          `json:"imagePullerType"`
	ServiceAccount     string          `json:"serviceAccount"`
}

// ImagePerceiver ...
type ImagePerceiver struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// PodPerceiver ...
type PodPerceiver struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	NamespaceFilter string `json:"namespaceFilter,omitempty"`
}

// Perceiver ...
type Perceiver struct {
	EnableImagePerceiver      bool            `json:"enableImagePerceiver"`
	EnablePodPerceiver        bool            `json:"enablePodPerceiver"`
	ImagePerceiver            *ImagePerceiver `json:"imagePerceiver,omitempty"`
	PodPerceiver              *PodPerceiver   `json:"podPerceiver,omitempty"`
	AnnotationIntervalSeconds int             `json:"annotationIntervalSeconds"`
	DumpIntervalMinutes       int             `json:"dumpIntervalMinutes"`
	ServiceAccount            string          `json:"serviceAccount"`
	Port                      int             `json:"port"`
}

// Skyfire ...
type Skyfire struct {
	Name           string `json:"name"`
	Image          string `json:"image"`
	Port           int    `json:"port"`
	PrometheusPort int    `json:"prometheusPort"`
	ServiceAccount string `json:"serviceAccount"`

	HubClientTimeoutSeconds      int `json:"hubClientTimeoutSeconds"`
	HubDumpPauseSeconds          int `json:"hubDumpPauseSeconds"`
	KubeDumpIntervalSeconds      int `json:"kubeDumpIntervalSeconds"`
	PerceptorDumpIntervalSeconds int `json:"perceptorDumpIntervalSeconds"`
}

// Prometheus container definition
type Prometheus struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Port  int    `json:"port"`
}

// OpsSightSpec is the spec for a OpsSight resource
type OpsSightSpec struct {
	// OpsSight
	Namespace     string      `json:"namespace"`
	Perceptor     *Perceptor  `json:"perceptor"`
	ScannerPod    *ScannerPod `json:"scannerPod"`
	Perceiver     *Perceiver  `json:"perceiver"`
	ConfigMapName string      `json:"configMapName"`
	SecretName    string      `json:"secretName"`

	// CPU and memory configurations
	DefaultCPU string `json:"defaultCpu,omitempty"` // Example: "300m"
	DefaultMem string `json:"defaultMem,omitempty"` // Example: "1300Mi"
	ScannerCPU string `json:"scannerCpu,omitempty"` // Example: "300m"
	ScannerMem string `json:"scannerMem,omitempty"` // Example: "1300Mi"

	// Log level
	LogLevel string `json:"logLevel,omitempty"`

	// Metrics
	EnableMetrics bool        `json:"enableMetrics"`
	Prometheus    *Prometheus `json:"prometheus"`

	// Skyfire
	EnableSkyfire bool     `json:"enableSkyfire"`
	Skyfire       *Skyfire `json:"skyfire"`

	// Black Duck
	Blackduck *Blackduck `json:"blackduck"`

	DesiredState string `json:"desiredState"`
}

// OpsSightStatus is the status for a OpsSight resource
type OpsSightStatus struct {
	State         string  `json:"state"`
	ErrorMessage  string  `json:"errorMessage"`
	InternalHosts []*Host `json:"internalHosts"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpsSightList is a list of OpsSight resources
type OpsSightList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []OpsSight `json:"items"`
}
