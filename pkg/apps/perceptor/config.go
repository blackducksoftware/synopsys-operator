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

package perceptor

// RegistryAuth will store the Openshift Internal Registries
type RegistryAuth struct {
	URL      string `json:"Url"`
	User     string
	Password string
}

// AppConfig defines the configuration for the perceptor app
type AppConfig struct {
	// CONTAINER CONFIGS
	// These are sed replaced into the config maps for the containers.
	PerceptorPort                         *int           `json:"perceptorPort,omitempty"`
	ScannerPort                           *int           `json:"scannerPort,omitempty"`
	PerceiverPort                         *int           `json:"perceiverPort,omitempty"`
	ImageFacadePort                       *int           `json:"imageFacadePort,omitempty"`
	SkyfirePort                           *int           `json:"skyfirePort,omitempty"`
	InternalRegistries                    []RegistryAuth `json:"internalRegistries,omitempty"`
	AnnotationIntervalSeconds             *int           `json:"annotationIntervalSeconds,omitempty"`
	DumpIntervalMinutes                   *int           `json:"dumpIntervalMinutes,omitempty"`
	HubHost                               string         `json:"hubHost,omitempty"`
	HubUser                               string         `json:"hubUser,omitempty"`
	HubPort                               *int           `json:"hubPort,omitempty"`
	HubUserPassword                       string         `json:"hubUserPassword,omitempty"`
	HubClientTimeoutPerceptorMilliseconds *int           `json:"hubClientTimeoutPerceptorMilliseconds,omitempty"`
	HubClientTimeoutScannerSeconds        *int           `json:"hubClientTimeoutScannerSeconds,omitempty"`
	ConcurrentScanLimit                   *int           `json:"concurrentScanLimit,omitempty"`
	Namespace                             string         `json:"namespace,omitempty"`
	DefaultVersion                        string         `json:"defaultVersion,omitempty"`

	// CONTAINER PULL CONFIG
	// These are for defining docker registry and image location and versions
	Registry  string `json:"registry,omitempty"`
	ImagePath string `json:"imagePath,omitempty"`

	PerceptorImageName      string `json:"perceptorImageName,omitempty"`
	ScannerImageName        string `json:"scannerImageName,omitempty"`
	PodPerceiverImageName   string `json:"podPerceiverImageName,omitempty"`
	ImagePerceiverImageName string `json:"imagePerceiverImageName,omitempty"`
	ImageFacadeImageName    string `json:"imageFacadeImageName,omitempty"`
	SkyfireImageName        string `json:"skyfireImageName,omitempty"`

	PerceptorImageVersion   string `json:"perceptorImageVersion,omitempty"`
	ScannerImageVersion     string `json:"scannerImageVersion,omitempty"`
	PerceiverImageVersion   string `json:"perceiverImageVersion,omitempty"`
	ImageFacadeImageVersion string `json:"imageFacadeImageVersion,omitempty"`
	SkyfireImageVersion     string `json:"skyfireImageVersion,omitempty"`

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPasswordOrToken string `json:"dockerPasswordOrToken,omitempty"`
	DockerUsername        string `json:"dockerUsername,omitempty"`

	ServiceAccounts  map[string]string `json:"serviceAccounts,omitempty"`
	ImagePerceiver   *bool             `json:"imagePerceiver,omitempty"`
	PodPerceiver     *bool             `json:"podPerceiver,omitempty"`
	Metrics          *bool             `json:"metrics,omitempty"`
	PerceptorSkyfire *bool             `json:"perceptorSkyfire,omitempty"`

	// CPU and memory configurations
	// Should be passed like: e.g. "300m"
	DefaultCPU string `json:"defaultCpu,omitempty"`
	// Should be passed like: e.g "1300Mi"
	DefaultMem string `json:"defaultMem.omitempty"`

	// Log level
	LogLevel string `json:"logLevel,omitempty"`

	// Environment Variables
	HubUserPasswordEnvVar string `json:"hubuserPasswordEnvVar"`

	// Configuration secret
	SecretName string `json:"secretName"`
}

// NewPerceptorAppDefaults creates a perceptor app configuration object
// with defaults
func NewPerceptorAppDefaults() *AppConfig {
	defaultPerceptorPort := 3001
	defaultPerceiverPort := 3002
	defaultScannerPort := 3003
	defaultIFPort := 3004
	defaultSkyfirePort := 3005
	defaultAnnotationInterval := 30
	defaultDumpInterval := 30
	defaultPerceptorHubClientTimeout := 5000
	defaultScannerHubClientTimeout := 30
	defaultHubPort := 443
	defaultScanLimit := 7
	defaultPodPerceiverEnabled := true

	return &AppConfig{
		PerceptorPort:                         &defaultPerceptorPort,
		PerceiverPort:                         &defaultPerceiverPort,
		ScannerPort:                           &defaultScannerPort,
		ImageFacadePort:                       &defaultIFPort,
		SkyfirePort:                           &defaultSkyfirePort,
		AnnotationIntervalSeconds:             &defaultAnnotationInterval,
		DumpIntervalMinutes:                   &defaultDumpInterval,
		HubClientTimeoutPerceptorMilliseconds: &defaultPerceptorHubClientTimeout,
		HubClientTimeoutScannerSeconds:        &defaultScannerHubClientTimeout,
		HubHost:                 "nginx-webapp-logstash",
		HubPort:                 &defaultHubPort,
		DockerUsername:          "admin",
		ConcurrentScanLimit:     &defaultScanLimit,
		DefaultVersion:          "master",
		PerceptorImageName:      "perceptor",
		ScannerImageName:        "perceptor-scanner",
		ImagePerceiverImageName: "image-perceiver",
		PodPerceiverImageName:   "pod-perceiver",
		ImageFacadeImageName:    "perceptor-imagefacade",
		SkyfireImageName:        "skyfire",
		LogLevel:                "debug",
		PodPerceiver:            &defaultPodPerceiverEnabled,
		SecretName:              "perceptor",
	}
}
