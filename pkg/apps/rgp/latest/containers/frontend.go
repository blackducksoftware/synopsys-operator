/*
Copyright (C) 2019 Synopsys, Inc.

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

package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// GetFrontendDeployment returns the front end deployment
func (g *RgpDeployer) GetFrontendDeployment() *components.Deployment {

	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "frontend-service",
		Namespace: g.Grspec.Namespace,
	})

	deployment.AddPod(g.GetFrontendPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "frontend-service",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "frontend-service",
	})

	return deployment
}

// GetFrontendPod returns the front end pod
func (g *RgpDeployer) GetFrontendPod() *components.Pod {

	pod := components.NewPod(horizonapi.PodConfig{
		Name: "frontend-servicer",
	})

	container, _ := g.GetFrontendContainer()

	pod.AddContainer(container)

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "frontend-service",
	})

	return pod
}

// GetFrontendContainer will return the container
func (g *RgpDeployer) GetFrontendContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "frontend-service",
		Image:      "gcr.io/snps-swip-staging/reporting-frontend-service:0.0.673",
		PullPolicy: horizonapi.PullIfNotPresent,
		MinCPU:     "250m",
		MinMem:     "500Mi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 8080,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getFrontendEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// GetFrontendService returns the front end service
func (g *RgpDeployer) GetFrontendService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "frontend-service",
		Namespace: g.Grspec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "frontend-service",
	})
	service.AddSelectors(map[string]string{
		"name": "frontend-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "80", Port: 80, Protocol: horizonapi.ProtocolTCP, TargetPort: "8080"})
	return service
}

func (g *RgpDeployer) getFrontendEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, g.getSwipEnvConfigs()...)
	return envs
}
