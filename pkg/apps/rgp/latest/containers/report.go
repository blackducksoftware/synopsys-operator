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

// GetReportDeployment returns the report deployment
func (g *RgpDeployer) GetReportDeployment() *components.Deployment {
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "report-service",
		Namespace: g.Grspec.Namespace,
	})

	deployment.AddPod(g.GetReportPod())
	deployment.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "report-service",
	})

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  "rgp",
		"name": "report-service",
	})
	log.Debugf("report: %+v", deployment)

	return deployment
}

// GetReportPod returns the report pod
func (g *RgpDeployer) GetReportPod() *components.Pod {

	pod := components.NewPod(horizonapi.PodConfig{
		Name: "report-service",
	})

	reportContainer, _ := g.getReportContainer()
	clamavContainer, _ := g.getClamavContainer()

	pod.AddContainer(reportContainer)
	pod.AddContainer(clamavContainer)

	for _, v := range g.getReportVolumes() {
		pod.AddVolume(v)
	}

	pod.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "report-service",
	})

	return pod
}

// getReportContainer returns the rp-report-service container
func (g *RgpDeployer) getReportContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "report-service",
		Image:      "gcr.io/snps-swip-staging/reporting-report-service:0.0.450",
		PullPolicy: horizonapi.PullIfNotPresent,
		MinCPU:     "250m",
		MinMem:     "1Gi",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 7979,
		Protocol:      horizonapi.ProtocolTCP,
	})

	for _, v := range g.getReportVolumeMounts() {
		err := container.AddVolumeMount(*v)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range g.getReportEnvConfigs() {
		container.AddEnv(*v)
	}

	return container, nil
}

// getClamavContainer returns the clamav container
func (g *RgpDeployer) getClamavContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "clamav",
		Image:      "gcr.io/snps-swip-staging/reporting-clamav:latest",
		PullPolicy: horizonapi.PullIfNotPresent,
		// TODO: RESTART POLICY: ALWAYS, horizon doesn't have it
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: 3310,
		Protocol:      horizonapi.ProtocolTCP,
	})

	return container, nil
}

// GetReportService returns the report service
func (g *RgpDeployer) GetReportService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "report-service",
		Namespace: g.Grspec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":  "rgp",
		"name": "report-service",
	})
	service.AddSelectors(map[string]string{
		"name": "report-service",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "7979", Port: 7979, Protocol: horizonapi.ProtocolTCP, TargetPort: "7979"})
	service.AddPort(horizonapi.ServicePortConfig{Name: "admin", Port: 8081, Protocol: horizonapi.ProtocolTCP, TargetPort: "8081"})
	return service
}

func (g *RgpDeployer) getReportVolumes() []*components.Volume {
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

func (g *RgpDeployer) getReportVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-cacert", MountPath: "/mnt/vault/ca"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-key", MountPath: "/mnt/vault/key"})
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "vault-client-cert", MountPath: "/mnt/vault/cert"})

	return volumeMounts
}

func (g *RgpDeployer) getReportEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	//envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromPodIP, NameOrPrefix: "POD_IP"})
	//envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLUSTER_ADDR", KeyOrVal: "https://$(POD_IP):8201"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_VAULT_ADDRESS", KeyOrVal: "https://vault:8200"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CACERT", KeyOrVal: "/mnt/vault/ca/vault_cacrt"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_KEY", KeyOrVal: "/mnt/vault/key/vault_client_key"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "VAULT_CLIENT_CERT", KeyOrVal: "/mnt/vault/cert/vault_client_cert"})

	envs = append(envs, g.getCommonEnvConfigs()...)
	envs = append(envs, g.getSwipEnvConfigs()...)
	envs = append(envs, g.getPostgresEnvConfigs()...)

	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MINIO_HOST", KeyOrVal: "minio"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MINIO_PORT", KeyOrVal: "9000"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MINIO_BUCKET", KeyOrVal: "reports"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MINIO_REGION", KeyOrVal: "us-central1"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MINIO_ACCESS_KEY", KeyOrVal: "access_key", FromName: "minio-keys"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MINIO_SECRET_KEY", KeyOrVal: "secret_key", FromName: "minio-keys"})

	return envs
}
