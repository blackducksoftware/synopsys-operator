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

package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizondeployer "github.com/blackducksoftware/horizon/pkg/deployer"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	"k8s.io/client-go/rest"
)

// RgpDeployer will contain the specification of RGP
type RgpDeployer struct {
	config     *v1.RgpSpec
	kubeConfig *rest.Config
	Grspec     *v1.RgpSpec
}

// NewRgpDeployer ...
func NewRgpDeployer(config *v1.RgpSpec, kubeConfig *rest.Config, rgpSpec *v1.RgpSpec) *RgpDeployer {
	return &RgpDeployer{config: config, kubeConfig: kubeConfig, Grspec: rgpSpec}
}

// GetDeployer will return the list of components for RGP
func (g *RgpDeployer) GetDeployer() (*horizondeployer.Deployer, error) {

	deployer, _ := horizondeployer.NewDeployer(g.kubeConfig)

	for _, deployment := range g.GetDeployments() {
		deployer.AddComponent(horizonapi.DeploymentComponent, deployment)
	}

	for _, service := range g.GetServices() {
		deployer.AddComponent(horizonapi.ServiceComponent, service)
	}

	return deployer, nil
}

// GetDeployments will return a list of Deployment
func (g *RgpDeployer) GetDeployments() []*components.Deployment {
	return []*components.Deployment{
		g.GetFrontendDeployment(),
		g.GetIssueManagerDeployment(),
		g.GetPolarisDeployment(),
		g.GetPortfolioDeployment(),
		g.GetReportDeployment(),
		g.GetToolsPortfolioDeployment(),
		g.GetAuthServerDeployment(),
	}
}

// GetServices will return a list of Service
func (g *RgpDeployer) GetServices() []*components.Service {
	return []*components.Service{
		g.GetFrontendService(),
		g.GetIssueManagerService(),
		g.GetPolarisService(),
		g.GetPortfolioService(),
		g.GetReportService(),
		g.GetToolsPortfolioService(),
		g.GetAuthServerService(),
	}
}
