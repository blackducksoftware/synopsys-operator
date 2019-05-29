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

// GetAuthServerDeployment return the auth server deployment
func (g *SpecConfig) GetAuthServerDeployment() *components.Deployment {
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "auth-server",
		Namespace: g.config.Namespace,
	})

	deployment.AddPod(g.getAuthServersPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "auth-server",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "auth-server",
	})

	return deployment
}

func (g *SpecConfig) getAuthServersPod() *components.Pod {

	pod := components.NewPod(horizonapi.PodConfig{
		Name: "auth-server",
	})

	container, _ := g.getAuthServerContainer()

	pod.AddContainer(container)
	for _, v := range g.getAuthServerVolumes() {
		pod.AddVolume(v)
	}

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "auth-server",
	})

	return pod
}

// getAuthServersContainer returns the auth server pod
func (g *SpecConfig) getAuthServerContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "auth-server",
		Image:      "gcr.io/snps-swip-staging/swip_auth-server:latest",
		PullPolicy: horizonapi.PullIfNotPresent,
		MinCPU:     "250m",
		MinMem:     "2Gi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 8080,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getAuthServerVolumeMounts() {
		err := container.AddVolumeMount(*v)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range g.getAuthServerEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// GetAuthServerService returns the auth server service
func (g *SpecConfig) GetAuthServerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "auth-server",
		Namespace: g.config.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "auth-server",
	})
	service.AddSelectors(map[string]string{
		"name": "auth-server",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "http", Port: 80, Protocol: horizonapi.ProtocolTCP, TargetPort: "8080"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *SpecConfig) getAuthServerVolumes() []*components.Volume {
	var volumes []*components.Volume

	// swip.vault.server.volume
	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-cacert",
		MapOrSecretName: "vault-ca-certificate",
		Items: []horizonapi.KeyPath{
			{
				Key:  "tls.crt",
				Path: "vault_cacrt",
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

func (g *SpecConfig) getAuthServerVolumeMounts() []*horizonapi.VolumeMountConfig {
	// swip.vault.server.volumemount
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca", ReadOnly: true})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-key", MountPath: "/mnt/vault/key", ReadOnly: true})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-server-cert", MountPath: "/mnt/vault/cert", ReadOnly: true})

	return volumeMounts
}

func (g *SpecConfig) getAuthServerEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig

	// swip.common.environment
	envs = append(envs, g.getSwipEnvConfigs()...)

	// swip.mongodb.root.environment
	envs = append(envs, g.getMongoEnvConfigs()...)

	envs = append(envs, g.getEventStoreLegacyEnvConfigs()...)
	envs = append(envs, g.getEventStoreEnvConfigs("admin")...)
	envs = append(envs, g.getEventStoreEnvConfigs("writer")...)
	envs = append(envs, g.getEventStoreEnvConfigs("reader")...)

	// swip.vault.server.environment
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_server_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_server_cert"})

	// smtp stuff
	// envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "SMTP_HOST", KeyOrVal: "host", FromName: "smtp"})
	// envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "SMTP_PORT", KeyOrVal: "port", FromName: "smtp"})
	// envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "SMTP_PASSWORD", KeyOrVal: "passwd", FromName: "smtp"})
	// envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "SMTP_USERNAME", KeyOrVal: "username", FromName: "smtp"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SMTP_HOST", KeyOrVal: "mailhost.internal.synopsys.com"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SMTP_PORT", KeyOrVal: "25"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SMTP_PASSWORD", KeyOrVal: ""})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SMTP_USERNAME", KeyOrVal: ""})

	// TODO: this was previously, make sure it is not needed
	// envs = append(envs, g.getPostgresEnvConfigs()...)

	return envs
}
