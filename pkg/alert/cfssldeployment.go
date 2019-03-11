/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package alert

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// getCfsslDeployment returns a new Deployment for a Cffsl
func (a *SpecConfig) getCfsslDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "cfssl",
		Namespace: a.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "cfssl", "tier": "cfssl"})

	pod, err := a.getCfsslPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}
	deployment.AddPod(pod)

	return deployment, nil
}

// getCfsslPod returns a new Pod for a Cffsl
func (a *SpecConfig) getCfsslPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "cfssl",
	})
	pod.AddLabels(map[string]string{"app": "cfssl", "tier": "cfssl"})

	pod.AddContainer(a.getCfsslContainer())

	vol, err := a.getCfsslVolume()
	if err != nil {
		return nil, fmt.Errorf("error creating volumes: %v", err)
	}
	pod.AddVolume(vol)

	return pod, nil
}

// getCfsslContainer returns a new Container for a Cffsl
func (a *SpecConfig) getCfsslContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:   "hub-cfssl",
		Image:  fmt.Sprintf("%s/%s/%s:%s", a.config.Registry, a.config.ImagePath, a.config.CfsslImageName, a.config.CfsslImageVersion),
		MinMem: a.config.CfsslMemory,
		MaxMem: a.config.CfsslMemory,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: "8888",
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-cfssl",
		MountPath: "/etc/cfssl",
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: "blackduck-alert-config",
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 10,
	})

	return container
}

// getCfsslVolume returns a new Volume for a Cffsl
func (a *SpecConfig) getCfsslVolume() (*components.Volume, error) {
	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "dir-cfssl",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}

	return vol, nil
}
