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
	"strconv"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// getAlertDeployment returns a new deployment for an Alert
func (a *SpecConfig) getAlertDeployment() *components.Deployment {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "alert",
		Namespace: a.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "alert", "tier": "alert"})

	pod := a.getAlertPod()

	deployment.AddPod(pod)

	deployment.AddLabels(map[string]string{"app": "alert"})
	return deployment
}

// getAlertPod returns a new Pod for an Alert
func (a *SpecConfig) getAlertPod() *components.Pod {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "alert",
	})
	pod.AddLabels(map[string]string{"app": "alert", "tier": "alert"})

	pod.AddContainer(a.getAlertContainer())

	vol := a.getAlertVolume()

	pod.AddVolume(vol)

	pod.AddLabels(map[string]string{"app": "alert"})
	return pod
}

// getAlertContainer returns a new Container for an Alert
func (a *SpecConfig) getAlertContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "alert",
		Image:      fmt.Sprintf("%s/%s/%s:%s", a.config.Registry, a.config.ImagePath, a.config.AlertImageName, a.config.AlertImageVersion),
		PullPolicy: horizonapi.PullAlways,
		MinMem:     a.config.AlertMemory,
		MaxMem:     a.config.AlertMemory,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: strconv.Itoa(*a.config.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-alert",
		MountPath: "/opt/blackduck/alert/alert-config",
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: "blackduck-alert-config",
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/alert/api/about"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 5,
	})

	return container
}

// getAlertVolume returns a new Volume for an Alert
func (a *SpecConfig) getAlertVolume() *components.Volume {
	vol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		VolumeName: "dir-alert",
		PVCName:    "alert-pvc",
		ReadOnly:   false,
	})

	return vol
}
