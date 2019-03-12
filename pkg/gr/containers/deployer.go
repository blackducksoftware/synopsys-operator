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

package containers

import (
	horizondeployer "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/gr/v1"
	"k8s.io/client-go/rest"
)

// SpecConfig will contain the specification of Alert
type GrDeployer struct {
	config     *v1.GrSpec
	kubeConfig *rest.Config
	Grspec     *v1.GrSpec
}

func NewGrDeployer(config *v1.GrSpec, kubeConfig *rest.Config, grspec *v1.GrSpec) *GrDeployer {
	return &GrDeployer{config: config, kubeConfig: kubeConfig, Grspec: grspec}
}

// GetComponents will return the list of components for alert
func (g *GrDeployer) GetDeployer() (*horizondeployer.Deployer, error) {

	deployer, _ := horizondeployer.NewDeployer(g.kubeConfig)

	deployer.AddDeployment(g.GetFrontendDeployment())
	deployer.AddDeployment(g.GetIssueManagerDeployment())
	deployer.AddDeployment(g.GetPolarisDeployment())
	deployer.AddDeployment(g.GetPortfolioDeployment())
	deployer.AddDeployment(g.GetReportDeployment())
	deployer.AddDeployment(g.GetToolsPortfolioDeployment())
	deployer.AddDeployment(g.GetAuthServerDeployment())

	deployer.AddService(g.GetFrontendService())
	deployer.AddService(g.GetIssueManagerService())
	deployer.AddService(g.GetPolarisService())
	deployer.AddService(g.GetPortfolioService())
	deployer.AddService(g.GetReportService())
	deployer.AddService(g.GetToolsPortfolioService())
	deployer.AddService(g.GetAuthServerService())

	return deployer, nil
}
