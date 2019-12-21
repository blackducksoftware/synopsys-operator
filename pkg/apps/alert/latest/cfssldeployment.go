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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	appsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// getCfsslDeployment returns a new Deployment for a Cffsl
func (a *SpecConfig) getCfsslDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deploymentConfig := horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "cfssl"),
		Namespace: a.alert.Spec.Namespace,
	}
	labels := map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "cfssl"}

	pod, err := a.getCfsslPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	return util.CreateDeployment(&deploymentConfig, pod, pod.GetLabels(), labels), nil
}

// getCfsslPod returns a new Pod for a Cffsl
func (a *SpecConfig) getCfsslPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: util.GetResourceName(a.alert.Name, util.AlertName, "cfssl"),
	})
	pod.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "cfssl"})

	container, err := a.getCfsslContainer()
	if err != nil {
		return nil, err
	}
	pod.AddContainer(container)

	vol, err := a.getCfsslVolume()
	if err != nil {
		return nil, fmt.Errorf("error creating volumes: %v", err)
	}
	pod.AddVolume(vol)

	if a.alert.Spec.RegistryConfiguration != nil && len(a.alert.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(a.alert.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

// getCfsslContainer returns a new Container for a Cffsl
func (a *SpecConfig) getCfsslContainer() (*components.Container, error) {
	image := appsutil.GenerateImageTag(GetImageTag(a.alert.Spec.Version, "blackduck-cfssl"), a.alert.Spec.ImageRegistries, a.alert.Spec.RegistryConfiguration)
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:   "blackduck-cfssl",
		Image:  image,
		MinMem: a.alert.Spec.CfsslMemory,
		MaxMem: a.alert.Spec.CfsslMemory,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(8888),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-cfssl",
		MountPath: "/etc/cfssl",
	})

	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: util.GetResourceName(a.alert.Name, util.AlertName, "blackduck-config"),
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Type:    horizonapi.ActionTypeCommand,
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 10,
	})

	return container, nil
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
