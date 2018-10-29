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

package opssight

import (
	"encoding/json"

	"github.com/blackducksoftware/perceptor-scanner/pkg/docker"
	"github.com/juju/errors"
)

// PerceiverConfig ...
type PerceiverConfig struct {
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int
	Pod                       PodPerceiverConfig
	Image                     ImagePerceiverConfig
}

// ImagePerceiverConfig ...
type ImagePerceiverConfig struct{}

// PodPerceiverConfig ...
type PodPerceiverConfig struct {
	NamespaceFilter string
}

// HubConfig ...
type HubConfig struct {
	Hosts               []string
	User                string
	PasswordEnvVar      string
	Port                int
	ConcurrentScanLimit int
	TotalScanLimit      int
}

// PerceptorTimingsConfig ...
type PerceptorTimingsConfig struct {
	CheckForStalledScansPauseHours int
	StalledScanClientTimeoutHours  int
	ModelMetricsPauseSeconds       int
	UnknownImagePauseMilliseconds  int
	HubClientTimeoutMilliseconds   int
}

// PerceptorConfig ...
type PerceptorConfig struct {
	Timings     PerceptorTimingsConfig
	UseMockMode bool
	Host        string
	Port        int
}

// ScannerConfig ...
type ScannerConfig struct {
	Port                    int
	ImageDirectory          string
	HubClientTimeoutSeconds int
}

// ImageFacadeConfig ...
type ImageFacadeConfig struct {
	Host                    string
	Port                    int
	PrivateDockerRegistries []docker.RegistryAuth
	CreateImagesOnly        bool
}

// ConfigMap ...
type ConfigMap struct {
	Perceiver   PerceiverConfig
	Hub         HubConfig
	Perceptor   PerceptorConfig
	Scanner     ScannerConfig
	ImageFacade ImageFacadeConfig
	LogLevel    string
}

func (cm *ConfigMap) jsonString() (string, error) {
	bytes, err := json.Marshal(cm)
	if err != nil {
		return "", errors.Annotate(err, "unable to serialize to json")
	}
	return string(bytes), nil
}
