/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package sample

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
)

// SampleSpecConfig represents your Sample component
type SampleSpecConfig struct {
	config *v1.SampleSpec
}

// NewSample will create a SampleSpecConfig
func NewSample(config *v1.SampleSpec) *SampleSpecConfig {
	return &SampleSpecConfig{config: config}
}

// GetComponents will return a ComponentList for the Sample
func (specConfig *SampleSpecConfig) GetComponents(config *v1.SampleSpec) (*api.ComponentList, error) {
	// TODO(user) Add your Horizon Components
	components := &api.ComponentList{
		Deployments: []*components.Deployment{
			specConfig.getSampleDeployment(),
		},
	}
	return components, nil
}
