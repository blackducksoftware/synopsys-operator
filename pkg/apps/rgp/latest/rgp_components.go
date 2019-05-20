/*
Copyright (C) 2019 Synopsys, Inc.

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

package rgp

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	rgpapi "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	log "github.com/sirupsen/logrus"
)

// SpecConfig will contain the specification to create the components of Rgp
type SpecConfig struct {
	config *rgpapi.RgpSpec
}

// NewSpecConfig will create the Rgp SpecConfig
func NewSpecConfig(config *rgpapi.RgpSpec) *SpecConfig {
	return &SpecConfig{config: config}
}

// GetComponents will return the list of components for rgp
func (a *SpecConfig) GetComponents() (*api.ComponentList, error) {
	log.Infof("Getting Rgp Components")
	components := &api.ComponentList{}

	components.Deployments = append(components.Deployments, a.GetFrontendDeployment())
	components.Deployments = append(components.Deployments, a.GetPolarisDeployment())
	components.Deployments = append(components.Deployments, a.GetReportDeployment())
	components.Deployments = append(components.Deployments, a.GetIssueManagerDeployment())
	components.Deployments = append(components.Deployments, a.GetPortfolioDeployment())
	components.Deployments = append(components.Deployments, a.GetToolsPortfolioDeployment())

	components.Services = append(components.Services, a.GetFrontendService())
	components.Services = append(components.Services, a.GetPolarisService())
	components.Services = append(components.Services, a.GetReportService())
	components.Services = append(components.Services, a.GetIssueManagerService())
	components.Services = append(components.Services, a.GetPortfolioService())
	components.Services = append(components.Services, a.GetToolsPortfolioService())

	return components, nil
}
