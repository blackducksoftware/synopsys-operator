/*
Copyright (C) 2018 Synopsys, Inc.

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

package apps

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// Vault stores the vault configuration
type Vault struct {
	namespace   string
	vaultConfig string
	//vaultCaCert  string
	//vaultTLSCert string
	//vaultTLSKey  string
	vaultSecrets    map[string]string
	vaultCacertPath string
	//vaultTLSSecretName string
	//vaultTLSMountPath string
}

// NewVault returns the vault configuration
func NewVault(namespace string, vaultConfig string, vaultSecrets map[string]string, vaultCacertPath string) *Vault {
	return &Vault{namespace: namespace, vaultConfig: vaultConfig, vaultSecrets: vaultSecrets, vaultCacertPath: vaultCacertPath}
}

// GetVaultDeployment will return the vault deployment
func (v *Vault) GetVaultDeployment() *components.Deployment {
	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "vault",
		Namespace: v.namespace,
	}
	return util.CreateDeployment(deployConfig, v.GetPod())
}

// GetPod returns the vault pod
func (v *Vault) GetPod() *components.Pod {
	envs := v.getVaultEnvConfigs()

	volumeMounts := v.getVaultVolumeMounts()
	var containers []*util.Container

	var trueBool = true

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "vault",
			Image:      "vault:0.11.2",
			PullPolicy: horizonapi.PullAlways,
			MinMem:     "",
			MaxMem:     "",
			MinCPU:     "",
			MaxCPU:     "",
			Command: []string{
				"vault",
				"server",
				"-config",
				"/vault/config/config.json",
			},
			ReadOnlyFS: &trueBool,
		},
		Capabilities: []string{"IPC_LOCK"},
		EnvConfigs:   envs,
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "8200"},
			{ContainerPort: "8201"},
		},
		//ReadinessProbeConfigs: []*horizonapi.ProbeConfig{
		//	{
		//		ActionConfig: horizonapi.ActionConfig{
		//			URL:     "TCP://:8200",
		//		},
		//	},
		//},
		LivenessProbeConfigs: []*horizonapi.ProbeConfig{
			{
				Delay: 180,
				ActionConfig: horizonapi.ActionConfig{
					URL: "https://:8200/v1/sys/health?standbycode=204&uninitcode=204&",
				},
			},
		},
	}

	containers = append(containers, container)
	return util.CreatePod("vault", "", v.getVaultVolumes(), containers, nil, nil)
}

// GetVaultServices will return the vault service
func (v *Vault) GetVaultServices() *components.Service {
	// Consul service
	vault := components.NewService(horizonapi.ServiceConfig{
		Name:          "vault",
		Namespace:     v.namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	vault.AddSelectors(map[string]string{
		"app": "vault",
	})
	vault.AddPort(horizonapi.ServicePortConfig{Name: "api", Port: 8200, Protocol: horizonapi.ProtocolTCP})
	return vault
}

// getVaultVolumes will return the vault volumes
func (v *Vault) getVaultVolumes() []*components.Volume {
	var volumes []*components.Volume

	emptyDir, _ := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "vault-root",
	})
	volumes = append(volumes, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-config",
		MapOrSecretName: "vault-config",
	}))
	volumes = append(volumes, emptyDir)

	for k := range v.vaultSecrets {
		volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
			VolumeName:      k,
			MapOrSecretName: k,
		}))
	}
	return volumes
}

// getVaultVolumeMounts will return the vault volume mounts
func (v *Vault) getVaultVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-config", MountPath: "/vault/config/"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-root", MountPath: "/root/"})

	for k, v := range v.vaultSecrets {
		volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: k, MountPath: v})
	}

	return volumeMounts
}

// getVaultEnvConfigs will return the vault environment config maps
func (v *Vault) getVaultEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromPodIP, NameOrPrefix: "POD_IP"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLUSTER_ADDR", KeyOrVal: "https://$(POD_IP):8201"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_LOG_LEVEL", KeyOrVal: "info"})
	if len(v.vaultCacertPath) > 0 {
		envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: v.vaultCacertPath})

	}
	return envs
}

// GetVaultConfigConfigMap returns the vault config maps
func (v *Vault) GetVaultConfigConfigMap() *components.ConfigMap {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "vault-config",
		Namespace: v.namespace,
	})

	cm.AddData(map[string]string{
		"config.json": v.vaultConfig,
	})
	return cm
}
