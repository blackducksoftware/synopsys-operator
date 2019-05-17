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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// GetPolarisDeployment returns the polaris deployment
func (g *RgpDeployer) GetPolarisDeployment() *components.Deployment {
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "polaris-service",
		Namespace: g.Grspec.Namespace,
	})

	deployment.AddPod(g.GetPolarisPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "polaris-service",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "polaris-service",
	})
	return deployment
}

// GetPolarisPod returns the polaris pod
func (g *RgpDeployer) GetPolarisPod() *components.Pod {

	pod := components.NewPod(horizonapi.PodConfig{
		Name: "polaris-service",
	})

	container, _ := g.getPolarisContainer()

	pod.AddContainer(container)
	for _, v := range g.getPolarisVolumes() {
		pod.AddVolume(v)
	}

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "polaris-service",
	})

	return pod
}

func (g *RgpDeployer) getPolarisContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "polaris-service",
		Image:      "gcr.io/snps-swip-staging/reporting-polaris-service:0.0.111",
		PullPolicy: horizonapi.PullAlways,
		MinCPU:     "500m",
		MinMem:     "1Gi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 7788,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getPolarisVolumeMounts() {
		err := container.AddVolumeMount(*v)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range g.getPolarisEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// GetPolarisService returns the polaris service
func (g *RgpDeployer) GetPolarisService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "polaris-service",
		Namespace: g.Grspec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "polaris-service",
	})
	service.AddSelectors(map[string]string{
		"name": "polaris-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "7788", Port: 7788, Protocol: horizonapi.ProtocolTCP, TargetPort: "7788"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *RgpDeployer) getPolarisVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "vault_cacrt",
				Path: "tls.crt",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-key",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "vault_client_key",
				Path: "tls.key",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-cert",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "vault_client_crt",
				Path: "tls.crt",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	return volumes
}

func (g *RgpDeployer) getPolarisVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *RgpDeployer) getPolarisEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_client_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_client_cert"})

	envs = append(envs, g.getCommonEnvConfigs()...)
	envs = append(envs, g.getSwipEnvConfigs()...)
	envs = append(envs, g.getPostgresEnvConfigs()...)

	return envs
}
