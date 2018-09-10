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

package api

import (
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/hubfederator"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/perceptor"
)

// ProtoformConfig defines the configuration for protoform
type ProtoformConfig struct {
	// Dry run wont actually install, but will print the objects definitions out.
	DryRun bool `json:"dryRun,omitempty"`

	HubUserPassword string `json:"hubUserPassword"`

	// Viper secrets
	ViperSecret string `json:"viperSecret,omitempty"`

	// Log level
	DefaultLogLevel string `json:"defaultLogLevel,omitempty"`

	Apps *ProtoformApps `json:"apps,omitempty"`
}

// ProtoformApps defines the configuration for supported apps
type ProtoformApps struct {
	PerceptorConfig    *perceptor.AppConfig    `json:"perceptorConfig,omitempty"`
	AlertConfig        *alert.AppConfig        `json:"alertConfig,omitempty"`
	HubFederatorConfig *hubfederator.AppConfig `json:"hubFederatorConfig,omitempty"`
}
