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

// GetZookeeperDeployment will return the zookeeper deployment
func (c *Creater) GetZookeeperDeployment(imageName string) (*components.ReplicationController, error) {

	volumeMounts := c.getZookeeperVolumeMounts()

	zookeeperContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "zookeeper", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: c.hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: zookeeperMinCPUUsage, MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: volumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: zookeeperPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		zookeeperContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"zkServer.sh", "status", "/opt/blackduck/zookeeper/conf/zoo.cfg"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/zookeeper/data && chmod -cR 777 /opt/blackduck/zookeeper/datalog"}},
			VolumeMounts:    volumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(zookeeperContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, "zookeeper", c.config.IsClusterScoped), Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			Volumes:             c.getZookeeperVolumes(),
			Containers:          []*util.Container{zookeeperContainerConfig},
			InitContainers:      initContainers,
			ImagePullSecrets:    c.hubSpec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("zookeeper"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("zookeeper"),
		}, c.GetLabel("zookeeper"))
}

// getZookeeperVolumes will return the zookeeper volumes
func (c *Creater) getZookeeperVolumes() []*components.Volume {
	var zookeeperDataVolume *components.Volume
	var zookeeperDatalogVolume *components.Volume

	if c.hubSpec.PersistentStorage {
		zookeeperDataVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-zookeeper-data", util.GetResourceName(c.name, "blackduck-zookeeper-data", c.config.IsClusterScoped))
	} else {
		zookeeperDataVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper-data")
	}

	if c.hubSpec.PersistentStorage {
		zookeeperDatalogVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-zookeeper-datalog", util.GetResourceName(c.name, "blackduck-zookeeper-datalog", c.config.IsClusterScoped))
	} else {
		zookeeperDatalogVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper-datalog")
	}

	volumes := []*components.Volume{zookeeperDataVolume, zookeeperDatalogVolume}
	return volumes
}

// getZookeeperVolumeMounts will return the zookeeper volume mounts
func (c *Creater) getZookeeperVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-zookeeper-data", MountPath: "/opt/blackduck/zookeeper/data"},
		{Name: "dir-zookeeper-datalog", MountPath: "/opt/blackduck/zookeeper/datalog"},
	}
	return volumesMounts
}

// GetZookeeperService will return the zookeeper service
func (c *Creater) GetZookeeperService() *components.Service {
	return util.CreateService(util.GetResourceName(c.name, "zookeeper", c.config.IsClusterScoped), c.GetLabel("zookeeper"), c.hubSpec.Namespace, zookeeperPort, zookeeperPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("zookeeper"))
}
