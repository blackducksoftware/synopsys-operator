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

package alert

// AppConfig defines the configuration options for alert
type AppConfig struct {
	Registry          string `json:"registry,omitempty"`
	ImagePath         string `json:"imagePath,omitempty"`
	AlertImageName    string `json:"alertImageName,omitempty"`
	AlertImageVersion string `json:"alertImageVersion,omitempty"`
	CfsslImageName    string `json:"cfsslImageName,omitempty"`
	CfsslImageVersion string `json:"cfsslImageVersion,omitempty"`
	HubHost           string `json:"hubHost,omitempty"`
	HubUser           string `json:"hubUser,omitempty"`
	HubPort           *int   `json:"hubPort,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	Port              *int   `json:"port"`
	StandAlone        *bool  `json:"standAlone"`

	// Should be passed like: e.g "1300Mi"
	AlertMemory string `json:"alertMemory.omitempty"`
	CfsslMemory string `json:"cfsslMemory.omitempty"`
}

// NewAppDefaults will return defaults for alert
func NewAppDefaults() *AppConfig {
	port := 8443
	hubPort := 443
	standAlone := true

	return &AppConfig{
		Port:           &port,
		HubPort:        &hubPort,
		StandAlone:     &standAlone,
		AlertMemory:    "512M",
		CfsslMemory:    "640M",
		AlertImageName: "blackduck-alert",
		CfsslImageName: "hub-cfssl",
	}
}
