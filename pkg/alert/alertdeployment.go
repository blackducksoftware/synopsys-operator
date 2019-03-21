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
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// getAlertDeployment returns a new deployment for an Alert
func (a *SpecConfig) getAlertDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "alert",
		Namespace: a.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "alert", "tier": "alert"})

	pod, err := a.getAlertPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}
	deployment.AddPod(pod)

	return deployment, nil
}

// getAlertPod returns a new Pod for an Alert
func (a *SpecConfig) getAlertPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "alert",
	})
	pod.AddLabels(map[string]string{"app": "alert", "tier": "alert"})

	pod.AddContainer(a.getAlertContainer())

	vol, err := a.getAlertVolume()
	if err != nil {
		return nil, fmt.Errorf("error creating volumes: %v", err)
	}
	pod.AddVolume(vol)

	return pod, nil
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
func (a *SpecConfig) getAlertVolume() (*components.Volume, error) {
	vol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		VolumeName: "dir-alert",
		PVCName:    "alert-pvc",
		ReadOnly:   true,
	})

	return vol, nil
}

func (a *SpecConfig) getPersistentVolumeClaim() (*components.PersistentVolumeClaim, error) {
	size := "100Gi"
	storageClass := ""
	pvc, err := operatorutil.CreatePersistentVolumeClaim("alert-pvc", a.config.Namespace, size, storageClass, horizonapi.ReadWriteOnce)
	if err != nil {
		return nil, fmt.Errorf("failed to create the postgres PVC %s in namespace %s because %+v", "alert-pvc", a.config.Namespace, err)
	}
	return pvc, nil
}
