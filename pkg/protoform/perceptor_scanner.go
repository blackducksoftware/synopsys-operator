/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershii.Config. The ASF licenses this file
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

package protoform

import (
	"fmt"
	"math"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (i *Installer) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(math.Ceil(float64(i.Config.ConcurrentScanLimit) / 2.0))
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      i.Config.ScannerImageName,
		Namespace: i.Config.Namespace,
	})

	rc.AddLabelSelectors(map[string]string{"name": i.Config.ScannerImageName})

	pod, err := i.scannerPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (i *Installer) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           i.Config.ScannerImageName,
		ServiceAccount: i.Config.ServiceAccounts["perceptor-image-facade"],
	})
	pod.AddLabels(map[string]string{"name": i.Config.ScannerImageName})

	pod.AddContainer(i.scannerContainer())
	pod.AddContainer(i.imageFacadeContainer())

	vols, err := i.scannerVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating scanner volumes: %v", err)
	}

	newVols, err := i.imageFacadeVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating image facade volumes: %v", err)
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (i *Installer) scannerContainer() *components.Container {
	priv := false
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       i.Config.ScannerImageName,
		Image:      fmt.Sprintf("%s/%s/%s:%s", i.Config.Registry, i.Config.ImagePath, i.Config.ScannerImageName, i.Config.ScannerImageVersion),
		Command:    []string{"./perceptor-scanner"},
		Args:       []string{"/etc/perceptor_scanner/perceptor_scanner.yaml"},
		MinCPU:     i.Config.DefaultCPU,
		MinMem:     i.Config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", i.Config.ScannerPort),
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
		NameOrPrefix: i.Config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     i.Config.ViperSecret,
	})

	return container
}

func (i *Installer) imageFacadeContainer() *components.Container {
	priv := true
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       i.Config.ImageFacadeImageName,
		Image:      fmt.Sprintf("%s/%s/%s:%s", i.Config.Registry, i.Config.ImagePath, i.Config.ImageFacadeImageName, i.Config.ImageFacadeImageVersion),
		Command:    []string{"./perceptor-imagefacade"},
		Args:       []string{"/etc/perceptor_imagefacade/perceptor_imagefacade.yaml"},
		MinCPU:     i.Config.DefaultCPU,
		MinMem:     i.Config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", i.Config.ImageFacadePort),
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

func (i *Installer) scannerVolumes() ([]*components.Volume, error) {
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

func (i *Installer) imageFacadeVolumes() ([]*components.Volume, error) {
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
func (i *Installer) ScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      i.Config.ScannerImageName,
		Namespace: i.Config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": i.Config.ScannerImageName})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(i.Config.ScannerPort),
		TargetPort: fmt.Sprintf("%d", i.Config.ScannerPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (i *Installer) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      i.Config.ImageFacadeImageName,
		Namespace: i.Config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": i.Config.ScannerImageName})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(i.Config.ImageFacadePort),
		TargetPort: fmt.Sprintf("%d", i.Config.ImageFacadePort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ScannerConfigMap creates a config map for the perceptor scanner
func (i *Installer) ScannerConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor-scanner",
		Namespace: i.Config.Namespace,
	})
	configMap.AddData(map[string]string{"perceptor_scanner.yaml": fmt.Sprint(`{"HubHost": "`, i.Config.HubHost, `","HubPort": "`, i.Config.HubPort, `","HubUser": "`, i.Config.HubUser, `","HubUserPasswordEnvVar": "`, i.Config.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, i.Config.HubClientTimeoutScannerSeconds, `","Port": "`, i.Config.ScannerPort, `","PerceptorHost": "`, i.Config.PerceptorImageName, `","PerceptorPort": "`, i.Config.PerceptorPort, `","ImageFacadeHost": "`, i.Config.ImageFacadeImageName, `","ImageFacadePort": "`, i.Config.ImageFacadePort, `","LogLevel": "`, i.Config.LogLevel, `"}`)})

	return configMap
}

//ImageFacadeConfigMap creates a config map for the perceptor image-facade
func (i *Installer) ImageFacadeConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor-imagefacade",
		Namespace: i.Config.Namespace,
	})
	configMap.AddData(map[string]string{"perceptor_imagefacade.yaml": fmt.Sprint(`{"DockerUser": "`, i.Config.DockerUsername, `","DockerPassword": "`, i.Config.DockerPasswordOrToken, `","Port": "`, i.Config.ImageFacadePort, `","InternalDockerRegistries": `, generateStringFromStringArr(i.Config.InternalDockerRegistries), `,"LogLevel": "`, i.Config.LogLevel, `"}`)})

	return configMap
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (i *Installer) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      i.Config.ServiceAccounts["perceptor-image-facade"],
		Namespace: i.Config.Namespace,
	})

	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (i *Installer) ScannerClusterRoleBinding() *components.ClusterRoleBinding {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "perceptor-scanner",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      i.Config.ServiceAccounts["perceptor-image-facade"],
		Namespace: i.Config.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	return scannerCRB
}
