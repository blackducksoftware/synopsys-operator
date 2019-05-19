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

// GetIssueManagerDeployment returns the issue manager deployment
func (g *RgpDeployer) GetIssueManagerDeployment() *components.Deployment {
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "rp-issue-manager",
		Namespace: g.Grspec.Namespace,
	})

	deployment.AddPod(g.GetIssueManagerPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "rp-issue-manager",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "rp-issue-manager",
	})
	return deployment
}

// GetIssueManagerPod returns the issue manager pod
func (g *RgpDeployer) GetIssueManagerPod() *components.Pod {

	pod := components.NewPod(horizonapi.PodConfig{
		Name: "rp-issue-manager",
	})

	container, _ := g.GetIssueManageContainer()

	pod.AddContainer(container)
	for _, v := range g.getIssueManagerVolumes() {
		pod.AddVolume(v)
	}

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "rp-issue-manager",
	})

	return pod
}

// GetIssueManageContainer will return the container
func (g *RgpDeployer) GetIssueManageContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "rp-issue-manager",
		Image:      "gcr.io/snps-swip-staging/reporting-rp-issue-manager:0.0.487",
		PullPolicy: horizonapi.PullAlways,
		MinCPU:     "500m",
		MinMem:     "1Gi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 6888,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getIssueManagerVolumeMounts() {
		err := container.AddVolumeMount(*v)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range g.getIssueManagerEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// GetIssueManagerService returns the issue manager service
func (g *RgpDeployer) GetIssueManagerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "rp-issue-manager",
		Namespace: g.Grspec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "rp-issue-manager",
	})
	service.AddSelectors(map[string]string{
		"name": "rp-issue-manager",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "6888", Port: 6888, Protocol: horizonapi.ProtocolTCP, TargetPort: "6888"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *RgpDeployer) getIssueManagerVolumes() []*components.Volume {
	var volumes []*components.Volume

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.crt",
				Path: "vault_cacrt",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-key",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.key",
				Path: "vault_client_key",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-client-cert",
		MapOrSecretName: "auth-client-tls-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.crt",
				Path: "vault_client_cert",
				Mode: util.IntToInt32(420),
			},
		},
	}))

	return volumes
}

func (g *RgpDeployer) getIssueManagerVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *RgpDeployer) getIssueManagerEnvConfigs() []*horizonapi.EnvConfig {
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
