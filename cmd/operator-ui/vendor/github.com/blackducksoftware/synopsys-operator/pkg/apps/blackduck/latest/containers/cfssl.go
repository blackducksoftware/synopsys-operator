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
)

// GetCfsslDeployment will return the cfssl deployment
func (c *Creater) GetCfsslDeployment() *components.ReplicationController {
	cfsslVolumeMounts := c.getCfsslolumeMounts()
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: c.getImageTag("blackduck-cfssl"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.CfsslMemoryLimit, MaxMem: c.hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: cfsslVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: cfsslPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		cfsslContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"}},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-cfssl") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /etc/cfssl"}},
			VolumeMounts:    cfsslVolumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(cfsslContainerConfig)

	cfssl := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "cfssl", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{cfsslContainerConfig}, c.getCfsslVolumes(), initContainers,
		[]horizonapi.AffinityConfig{}, c.GetVersionLabel("cfssl"), c.GetLabel("cfssl"))

	return cfssl
}

// getCfsslVolumes will return the cfssl volumes
func (c *Creater) getCfsslVolumes() []*components.Volume {
	var cfsslVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-cfssl") {
		cfsslVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-cfssl", "blackduck-cfssl")
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
	return util.CreateService("cfssl", c.GetLabel("cfssl"), c.hubSpec.Namespace, cfsslPort, cfsslPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("cfssl"))
}
