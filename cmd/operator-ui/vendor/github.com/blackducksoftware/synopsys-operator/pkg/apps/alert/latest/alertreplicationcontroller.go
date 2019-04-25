/*
Copyright (C) 2019 Synopsys, Inc.

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
	log "github.com/sirupsen/logrus"
)

// getAlertReplicationController returns a new replication controller for an Alert
func (a *SpecConfig) getAlertReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	replicationController := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      "alert",
		Namespace: a.config.Namespace,
	})
	replicationController.AddLabelSelectors(map[string]string{"app": "alert", "component": "alert"})

	pod, err := a.getAlertPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Pod: %s", err)
	}

	replicationController.AddPod(pod)
	replicationController.AddLabels(map[string]string{"app": "alert", "component": "alert"})
	return replicationController, nil
}

// getAlertPod returns a new Pod for an Alert
func (a *SpecConfig) getAlertPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "alert",
	})
	pod.AddLabels(map[string]string{"app": "alert", "component": "alert"})

	pod.AddContainer(a.getAlertContainer())

	if a.config.PersistentStorage {
		log.Debugf("Adding a PersistentVolumeClaim Volume to the Alert's Pod")
		pod.AddVolume(a.getAlertPVCVolume())
	} else {
		log.Debugf("Adding an EmptyDir Volume to the Alert's Pod")
		vol, err := a.getAlertEmptyDirVolume()
		if err != nil {
			return nil, fmt.Errorf("failed to Add Volume to Alert Pod: %s", err)
		}
		pod.AddVolume(vol)
	}

	pod.AddLabels(map[string]string{"app": "alert", "component": "alert"})
	return pod, nil
}

// getAlertContainer returns a new Container for an Alert
func (a *SpecConfig) getAlertContainer() *components.Container {
	image := a.config.AlertImage
	if image == "" {
		image = GetImageTag(a.config.Version, "blackduck-alert")
	}
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "alert",
		Image:      image,
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

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromSecret,
		FromName: "alert-secret",
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

// getAlertEmptyDirVolume returns a new EmptyDirVolume for an Alert
func (a *SpecConfig) getAlertEmptyDirVolume() (*components.Volume, error) {
	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "dir-alert",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert EmptyDir: %s", err)
	}

	return vol, err
}

// getAlertPVCVolume returns a new PVCVolume for an Alert
func (a *SpecConfig) getAlertPVCVolume() *components.Volume {
	vol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		VolumeName: "dir-alert",
		PVCName:    a.config.PVCName,
		ReadOnly:   false,
	})

	return vol
}
