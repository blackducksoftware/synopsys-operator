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

package protoform

import (
	"encoding/json"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
)

type protoformConfig struct {
	// Dry run wont actually install, but will print the objects definitions out.
	DryRun bool

	// CONTAINER CONFIGS
	// These are sed replaced into the config maps for the containers.
	PerceptorPort                         int
	ScannerPort                           int
	PerceiverPort                         int
	ImageFacadePort                       int
	SkyfirePort                           int
	InternalDockerRegistries              []string
	AnnotationIntervalSeconds             int
	DumpIntervalMinutes                   int
	HubHost                               string
	HubUser                               string
	HubUserPassword                       string
	HubPort                               int
	HubClientTimeoutPerceptorMilliseconds int
	HubClientTimeoutScannerSeconds        int
	ConcurrentScanLimit                   int
	Namespace                             string
	DefaultVersion                        string

	// CONTAINER PULL CONFIG
	// These are for defining docker registry and image location and versions
	Registry  string
	ImagePath string

	PerceptorImageName      string
	ScannerImageName        string
	PodPerceiverImageName   string
	ImagePerceiverImageName string
	ImageFacadeImageName    string
	SkyfireImageName        string

	PerceptorImageVersion   string
	ScannerImageVersion     string
	PerceiverImageVersion   string
	ImageFacadeImageVersion string
	SkyfireImageVersion     string

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPasswordOrToken string
	DockerUsername        string

	ServiceAccounts map[string]string
	ImagePerceiver  bool
	PodPerceiver    bool
	Metrics         bool

	// CPU and memory configurations
	DefaultCPU string // Should be passed like: e.g. "300m"
	DefaultMem string // Should be passed like: e.g "1300Mi"

	// Log level
	LogLevel string

	// Viper secrets
	ViperSecret string

	// Environment Variables
	HubUserPasswordEnvVar string

	// Automate test
	PerceptorSkyfire bool
}

// NewDefaultsObj returns a ProtoformDefaults object with sane default for most options
func NewDefaultsObj() *api.ProtoformDefaults {
	return &api.ProtoformDefaults{
		PerceptorPort:                         3001,
		PerceiverPort:                         3002,
		ScannerPort:                           3003,
		ImageFacadePort:                       3004,
		SkyfirePort:                           3005,
		AnnotationIntervalSeconds:             30,
		DumpIntervalMinutes:                   30,
		HubClientTimeoutPerceptorMilliseconds: 5000,
		HubClientTimeoutScannerSeconds:        30,
		HubHost:                 "nginx-webapp-logstash",
		HubPort:                 443,
		DockerUsername:          "admin",
		ConcurrentScanLimit:     7,
		DefaultVersion:          "master",
		PerceptorImageName:      "perceptor",
		ScannerImageName:        "perceptor-scanner",
		ImagePerceiverImageName: "image-perceiver",
		PodPerceiverImageName:   "pod-perceiver",
		ImageFacadeImageName:    "perceptor-imagefacade",
		SkyfireImageName:        "skyfire",
		LogLevel:                "debug",
		PodPerceiver:            true,
	}
}

func generateStringFromStringArr(strArr []string) string {
	str, _ := json.Marshal(strArr)
	return string(str)
}
