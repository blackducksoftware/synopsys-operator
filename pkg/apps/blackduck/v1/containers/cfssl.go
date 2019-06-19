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

package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetCfsslDeployment will return the cfssl deployment
func (c *Creater) GetCfsslDeployment(imageName string) (*components.ReplicationController, error) {
	cfsslVolumeMounts := c.getCfsslolumeMounts()
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.CfsslMemoryLimit, MaxMem: c.hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: cfsslVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: cfsslPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackDuck.Spec.LivenessProbes {
		cfsslContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	var initContainers []*util.Container
	if c.blackDuck.Spec.PersistentStorage {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -c 777 /etc/cfssl"}},
			VolumeMounts:    cfsslVolumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(cfsslContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"), Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			Volumes:             c.getCfsslVolumes(),
			Containers:          []*util.Container{cfsslContainerConfig},
			InitContainers:      initContainers,
			ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("cfssl"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("cfssl"),
		}, c.GetLabel("cfssl"))
}

// getCfsslVolumes will return the cfssl volumes
func (c *Creater) getCfsslVolumes() []*components.Volume {
	var cfsslVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		cfsslVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-cfssl", util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"))
	} else {
		cfsslVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	}

	volumes := []*components.Volume{cfsslVolume}
	return volumes
}

// getCfsslolumeMounts will return the cfssl volume mounts
func (c *Creater) getCfsslolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-cfssl", MountPath: "/etc/cfssl"},
	}
	return volumesMounts
}

// GetCfsslService will return the cfssl service
func (c *Creater) GetCfsslService() *components.Service {
	return util.CreateService(util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"), c.GetLabel("cfssl"), c.blackDuck.Spec.Namespace, cfsslPort, cfsslPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("cfssl"))
}
