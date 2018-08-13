/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.config. The ASF licenses this file
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

package perceptor

import (
	"fmt"
	"math"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (p *App) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(math.Ceil(float64(*p.config.ConcurrentScanLimit) / 2.0))
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.ScannerImageName,
		Namespace: p.config.Namespace,
	})

	rc.AddLabelSelectors(map[string]string{"name": p.config.ScannerImageName})

	pod, err := p.scannerPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *App) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.config.ScannerImageName,
		ServiceAccount: p.config.ServiceAccounts["perceptor-image-facade"],
	})
	pod.AddLabels(map[string]string{"name": p.config.ScannerImageName})

	pod.AddContainer(p.scannerContainer())
	pod.AddContainer(p.imageFacadeContainer())

	vols, err := p.scannerVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating scanner volumes: %v", err)
	}

	newVols, err := p.imageFacadeVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating image facade volumes: %v", err)
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *App) scannerContainer() *components.Container {
	priv := false
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       p.config.ScannerImageName,
		Image:      fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.ScannerImageName, p.config.ScannerImageVersion),
		Command:    []string{fmt.Sprintf("./%s", p.config.ScannerImageName)},
		Args:       []string{"/etc/perceptor_scanner/perceptor_scanner.yaml"},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.ScannerPort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceptor-scanner",
		MountPath: "/etc/perceptor_scanner",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})

	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})

	return container
}

func (p *App) imageFacadeContainer() *components.Container {
	priv := true
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       p.config.ImageFacadeImageName,
		Image:      fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.ImageFacadeImageName, p.config.ImageFacadeImageVersion),
		Command:    []string{"./perceptor-imagefacade"},
		Args:       []string{"/etc/perceptor_imagefacade/perceptor_imagefacade.yaml"},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.ImageFacadePort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceptor-imagefacade",
		MountPath: "/etc/perceptor_imagefacade",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-docker-socket",
		MountPath: "/var/run/docker.sock",
	})

	return container
}

func (p *App) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceptor-scanner",
		MapOrSecretName: "perceptor-scanner",
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

func (p *App) imageFacadeVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceptor-imagefacade",
		MapOrSecretName: "perceptor-imagefacade",
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
		VolumeName: "dir-docker-socket",
		Path:       "/var/run/docker.sock",
	}))

	return vols, nil
}

// ScannerService creates a service for perceptor scanner
func (p *App) ScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ScannerImageName,
		Namespace: p.config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": p.config.ScannerImageName})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.ScannerPort),
		TargetPort: fmt.Sprintf("%d", *p.config.ScannerPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (p *App) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ImageFacadeImageName,
		Namespace: p.config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": p.config.ScannerImageName})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.ImageFacadePort),
		TargetPort: fmt.Sprintf("%d", *p.config.ImageFacadePort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ScannerConfigMap creates a config map for the perceptor scanner
func (p *App) ScannerConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor-scanner",
		Namespace: p.config.Namespace,
	})
	configMap.AddData(map[string]string{"perceptor_scanner.yaml": fmt.Sprint(`{"HubHost": "`, p.config.HubHost, `","HubPort": "`, *p.config.HubPort, `","HubUser": "`, p.config.HubUser, `","HubUserPasswordEnvVar": "`, p.config.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, *p.config.HubClientTimeoutScannerSeconds, `","Port": "`, *p.config.ScannerPort, `","PerceptorHost": "`, p.config.PerceptorImageName, `","PerceptorPort": "`, *p.config.PerceptorPort, `","ImageFacadeHost": "`, p.config.ImageFacadeImageName, `","ImageFacadePort": "`, *p.config.ImageFacadePort, `","LogLevel": "`, p.config.LogLevel, `"}`)})

	return configMap
}

//ImageFacadeConfigMap creates a config map for the perceptor image-facade
func (p *App) ImageFacadeConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor-imagefacade",
		Namespace: p.config.Namespace,
	})
	configMap.AddData(map[string]string{"perceptor_imagefacade.yaml": fmt.Sprint(`{"DockerUser": "`, p.config.DockerUsername, `","DockerPassword": "`, p.config.DockerPasswordOrToken, `","Port": "`, *p.config.ImageFacadePort, `","InternalDockerRegistries": `, p.generateStringFromStringArr(p.config.InternalDockerRegistries), `,"LogLevel": "`, p.config.LogLevel, `"}`)})

	return configMap
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (p *App) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      p.config.ServiceAccounts["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *App) ScannerClusterRoleBinding() *components.ClusterRoleBinding {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "perceptor-scanner",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.config.ServiceAccounts["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	return scannerCRB
}
