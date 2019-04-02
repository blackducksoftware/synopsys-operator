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

// GetUploadCacheDeployment will return the uploadCache deployment
func (c *Creater) GetUploadCacheDeployment() *components.ReplicationController {
	volumeMounts := c.getUploadCacheVolumeMounts()

	uploadCacheContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "uploadcache", Image: c.getImageTag("blackduck-upload-cache"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.UploadCacheMemoryLimit, MaxMem: c.hubContainerFlavor.UploadCacheMemoryLimit,
			MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: uploadCachePort1, Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: uploadCachePort2, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		uploadCacheContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"curl", "--insecure", "-X", "GET", "--verbose", "http://localhost:8086/live?full=1"}},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 5,
		}}
	}

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-uploadcache") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/hub-upload-cache/uploads"}},
			VolumeMounts:    volumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(uploadCacheContainerConfig)

	uploadCache := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace,
		Name: "uploadcache", Replicas: util.IntToInt32(1)}, "", []*util.Container{uploadCacheContainerConfig}, c.getUploadCacheVolumes(),
		initContainers, []horizonapi.AffinityConfig{}, c.GetVersionLabel("uploadcache"), c.GetLabel("uploadcache"))

	return uploadCache
}

// getUploadCacheVolumes will return the uploadCache volumes
func (c *Creater) getUploadCacheVolumes() []*components.Volume {
	uploadCacheSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-security")
	var uploadCacheDataEmptyDir *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-uploadcache") {
		uploadCacheDataEmptyDir, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-data", "blackduck-uploadcache")
	} else {
		uploadCacheDataEmptyDir, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-data")
	}
	volumes := []*components.Volume{uploadCacheSecurityEmptyDir, uploadCacheDataEmptyDir}
	return volumes
}

// getUploadCacheVolumeMounts will return the uploadCache volume mounts
func (c *Creater) getUploadCacheVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-uploadcache-security", MountPath: "/opt/blackduck/hub/hub-upload-cache/security"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/hub-upload-cache/uploads"},
	}
	return volumesMounts
}

// GetUploadCacheService will return the uploadCache service
func (c *Creater) GetUploadCacheService() *components.Service {
	return util.CreateServiceWithMultiplePort("uploadcache", c.GetVersionLabel("uploadcache"), c.hubSpec.Namespace, []string{uploadCachePort1, uploadCachePort2},
		horizonapi.ClusterIPServiceTypeDefault, c.GetLabel("uploadcache"))
}
