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

package hubfederator

// AppConfig defines the configuration options for the hub federator
type AppConfig struct {
	DryRun          bool   `json:"dryRun,omitempty"`
	Registry        string `json:"registry,omitempty"`
	ImagePath       string `json:"imagePath,omitempty"`
	ImageName       string `json:"alertImageName,omitempty"`
	ImageVersion    string `json:"alertImageVersion,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	RegistrationKey string `json:"registrationKey,omitempty"`
	Port            *int   `json:"port,omitempty"`
	LogLevel        string `json:"logLevel,omitempty"`
	NumberOfThreads *int   `json:"numberOfThreads"`
}

// NewAppDefaults will return defaults for the hub federator
func NewAppDefaults() *AppConfig {
	port := 8080
	threads := 5

	return &AppConfig{
		ImageName:       "hub-federator",
		Port:            &port,
		NumberOfThreads: &threads,
	}
}
