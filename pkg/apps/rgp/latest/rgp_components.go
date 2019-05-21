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

// AddComponents will return the list of components for rgp
func (a *SpecConfig) AddComponents(componentList *api.ComponentList) {
	log.Infof("Adding Rgp Components")

	componentList.Deployments = append(componentList.Deployments, a.GetFrontendDeployment())
	componentList.Deployments = append(componentList.Deployments, a.GetPolarisDeployment())
	componentList.Deployments = append(componentList.Deployments, a.GetReportDeployment())
	componentList.Deployments = append(componentList.Deployments, a.GetIssueManagerDeployment())
	componentList.Deployments = append(componentList.Deployments, a.GetPortfolioDeployment())
	componentList.Deployments = append(componentList.Deployments, a.GetToolsPortfolioDeployment())

	componentList.Services = append(componentList.Services, a.GetFrontendService())
	componentList.Services = append(componentList.Services, a.GetPolarisService())
	componentList.Services = append(componentList.Services, a.GetReportService())
	componentList.Services = append(componentList.Services, a.GetIssueManagerService())
	componentList.Services = append(componentList.Services, a.GetPortfolioService())
	componentList.Services = append(componentList.Services, a.GetToolsPortfolioService())

}
