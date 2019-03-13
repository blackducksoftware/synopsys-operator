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

// Minio stores the minio configuration
type Minio struct {
	namespace string
	pvcName   string
	accessKey string
	secretKey string
}

// NewMinio returns the Minio configuration
func NewMinio(namespace string, pvcName string, accessKey string, secretKey string) *Minio {
	return &Minio{namespace: namespace, pvcName: pvcName, accessKey: accessKey, secretKey: secretKey}
}

// GetDeployment will return the deployment
func (c *Minio) GetDeployment() *components.Deployment {
	envs := c.getEnvConfigs()
	volumes := c.getVolumes()
	volumeMounts := c.getVolumeMounts()

	var containers []*util.Container

	container := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "minio",
			Image:      "gcr.io/snps-swip-staging/public/minio/minio:RELEASE.2018-12-27T18-33-08Z",
			PullPolicy: horizonapi.PullIfNotPresent,
			MinMem:     "",
			MaxMem:     "",
			MinCPU:     "",
			MaxCPU:     "",
			Args: []string{
				"server",
				"/storage",
			},
		},
		EnvConfigs:   envs,
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: "9000"},
		},
	}

	containers = append(containers, container)
	// TODO add capabalitiy, healthchecks, etc

	deployConfig := &horizonapi.DeploymentConfig{
		Name:      "minio",
		Namespace: c.namespace,
		Replicas:  util.IntToInt32(1),
	}

	return util.CreateDeploymentFromContainer(deployConfig, "", containers, volumes, nil, nil)
}

// GetServices will return the service
func (c *Minio) GetServices() *components.Service {
	// Consul service
	minio := components.NewService(horizonapi.ServiceConfig{
		Name:          "minio",
		Namespace:     c.namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeDefault,
	})
	minio.AddSelectors(map[string]string{
		"app": "minio",
	})
	minio.AddPort(horizonapi.ServicePortConfig{Name: "minio", Port: 9000, Protocol: horizonapi.ProtocolTCP})
	return minio
}

// getVolumes will return the  volumes
func (c *Minio) getVolumes() []*components.Volume {
	var volumes []*components.Volume
	var volume *components.Volume

	if len(c.pvcName) > 0 {
		volume, _ = util.CreatePersistentVolumeClaimVolume("storage", c.pvcName)
	} else {
		volume, _ = components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
			VolumeName: "storage",
		})

	}
	volumes = append(volumes, volume)

	volumes = append(volumes, components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "minio-keys",
		MapOrSecretName: "minio-keys",
	}))

	return volumes
}

// getVolumeMounts will return the volume mounts
func (c *Minio) getVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "storage", MountPath: "/storage"})

	return volumeMounts
}

// getEnvConfigs will return the environment variable configuration
func (c *Minio) getEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MINIO_ACCESS_KEY", KeyOrVal: "access_key", FromName: "minio-keys"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MINIO_SECRET_KEY", KeyOrVal: "secret_key", FromName: "minio-keys"})

	return envs
}

// GetSecret will return the secret
func (c *Minio) GetSecret() *components.Secret {
	secret := components.NewSecret(horizonapi.SecretConfig{
		Name:      "minio-keys",
		Namespace: c.namespace,
	})

	secret.AddStringData(map[string]string{
		"access_key": c.accessKey,
		"secret_key": c.secretKey,
	})

	return secret
}
