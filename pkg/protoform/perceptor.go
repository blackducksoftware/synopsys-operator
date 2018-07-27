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

package protoform

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// PerceptorReplicationController creates a replication controller for the perceptor
func (i *Installer) PerceptorReplicationController() *components.ReplicationController {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      i.Config.PerceptorImageName,
		Namespace: i.Config.Namespace,
	})
	rc.AddPod(i.perceptorPod())
	rc.AddLabelSelectors(map[string]string{"name": i.Config.PerceptorImageName})

	return rc
}

func (i *Installer) perceptorPod() *components.Pod {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: i.Config.PerceptorImageName,
	})
	pod.AddLabels(map[string]string{"name": i.Config.PerceptorImageName})
	pod.AddContainer(i.perceptorContainer())
	pod.AddVolume(i.perceptorVolume())

	return pod
}

func (i *Installer) perceptorContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    i.Config.PerceptorImageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", i.Config.Registry, i.Config.ImagePath, i.Config.PerceptorImageName, i.Config.PerceptorImageVersion),
		Command: []string{"./perceptor"},
		Args:    []string{"/etc/perceptor/perceptor.yaml"},
		MinCPU:  i.Config.DefaultCPU,
		MinMem:  i.Config.DefaultMem,
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", i.Config.PerceptorPort),
		Protocol:      horizonapi.ProtocolTCP,
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceptor",
		MountPath: "/etc/perceptor",
	})
	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: i.Config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     i.Config.ViperSecret,
	})

	return container
}

func (i *Installer) perceptorVolume() *components.Volume {
	volume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceptor",
		MapOrSecretName: "perceptor",
	})

	return volume
}

// PerceptorService creates a service for the perceptor
func (i *Installer) PerceptorService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      i.Config.PerceptorImageName,
		Namespace: i.Config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(i.Config.PerceptorPort),
		TargetPort: fmt.Sprintf("%d", i.Config.PerceptorPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": i.Config.PerceptorImageName})

	return service
}

// PerceptorConfigMap creates a config map for the perceptor
func (i *Installer) PerceptorConfigMap() *components.ConfigMap {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor",
		Namespace: i.Config.Namespace,
	})
	cm.AddData(map[string]string{"perceptor.yaml": fmt.Sprint(`{"HubHost": "`, i.Config.HubHost, `","HubPort": "`, i.Config.HubPort, `","HubUser": "`, i.Config.HubUser, `","HubUserPasswordEnvVar": "`, i.Config.HubUserPasswordEnvVar, `","HubClientTimeoutMilliseconds": "`, i.Config.HubClientTimeoutPerceptorMilliseconds, `","ConcurrentScanLimit": "`, i.Config.ConcurrentScanLimit, `","Port": "`, i.Config.PerceptorPort, `","LogLevel": "`, i.Config.LogLevel, `"}`)})

	return cm
}
