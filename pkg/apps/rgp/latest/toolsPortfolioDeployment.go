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

package rgp

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// GetToolsPortfolioDeployment returns the tools portfolio deployment
func (g *SpecConfig) GetToolsPortfolioDeployment() *components.Deployment {
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "tools-portfolio-service",
		Namespace: g.config.Namespace,
	})

	deployment.AddPod(g.getToolsPortfolioPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "tools-portfolio-service",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "tools-portfolio-service",
	})
	return deployment
}

func (g *SpecConfig) getToolsPortfolioPod() *components.Pod {

	// TODO: HELM CHART HAS serviceAccount: "auth-server"
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           "tools-portfolio-service",
		ServiceAccount: "auth-server",
	})

	container, _ := g.getToolPortfolioContainer()

	pod.AddContainer(container)
	for _, v := range g.getToolsPortfolioVolumes() {
		pod.AddVolume(v)
	}

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "tools-portfolio-service",
	})

	return pod
}

func (g *SpecConfig) getToolPortfolioContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "tools-portfolio-service",
		Image:      GetImageTag(g.config.Version, "reporting-tools-portfolio-service"),
		PullPolicy: horizonapi.PullIfNotPresent,
		MinCPU:     "250m",
		MinMem:     "1Gi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 60281,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getToolsPortfolioVolumeMounts() {
		err := container.AddVolumeMount(*v)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range g.getToolsPortfolioEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// GetToolsPortfolioService returns the tools portfolio service
func (g *SpecConfig) GetToolsPortfolioService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "tools-portfolio-service",
		Namespace: g.config.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "tools-portfolio-service",
	})
	service.AddSelectors(map[string]string{
		"name": "tools-portfolio-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "60289", Port: 60281, Protocol: horizonapi.ProtocolTCP, TargetPort: "60281"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *SpecConfig) getToolsPortfolioVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.crt",
				Path: "vault_cacrt",
				// TODO: 420 not specified as DefaultMode in HELM chart, not sure, if bug on their end or we should just assume
				// Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-server-key",
		MapOrSecretName: "auth-server-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.key",
				Path: "vault_server_key",
				// Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-server-cert",
		MapOrSecretName: "auth-server-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.crt",
				Path: "vault_server_cert",
				// Mode: util.IntToInt32(420),
			},
		},
	}))

	return volumes
}

func (g *SpecConfig) getToolsPortfolioVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *SpecConfig) getToolsPortfolioEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_server_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_server_cert"})

	envs = append(envs, g.getCommonEnvConfigs()...)
	envs = append(envs, g.getSwipEnvConfigs()...)
	envs = append(envs, g.getPostgresEnvConfigs()...)

	return envs
}
