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
	"fmt"
	"strconv"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	hubutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetUploadCacheDeployment will return the uploadCache deployment
func (c *Creater) GetUploadCacheDeployment() *components.ReplicationController {
	volumeMounts := c.getUploadCacheVolumeMounts()

	image := c.GetFullContainerName("upload")
	if strings.EqualFold(image, "") {
		return nil
	}

	uploadCacheContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "uploadcache", Image: c.GetFullContainerName("upload"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.UploadCacheMemoryLimit, MaxMem: c.hubContainerFlavor.UploadCacheMemoryLimit,
			MinCPU: "", MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
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
		initContainers, []horizonapi.AffinityConfig{})

	return uploadCache
}

// getUploadCacheVolumes will return the uploadCache volumes
func (c *Creater) getUploadCacheVolumes() []*components.Volume {
	uploadCacheSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-security")
	var uploadCacheDataDir *components.Volume
	var uploadCacheDataKey *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-uploadcache-data") {
		uploadCacheDataDir, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-data", "blackduck-uploadcache-data")
		uploadCacheDataKey, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-key", "blackduck-uploadcache-key")
	} else {
		uploadCacheDataDir, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-data")
		uploadCacheDataKey, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-key")
	}
	volumes := []*components.Volume{uploadCacheSecurityEmptyDir, uploadCacheDataDir, uploadCacheDataKey}
	return volumes
}

// getUploadCacheVolumeMounts will return the uploadCache volume mounts
func (c *Creater) getUploadCacheVolumeMounts() []*horizonapi.VolumeMountConfig {
	prefix := c.getMountPrefixPath()
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-uploadcache-security", MountPath: fmt.Sprintf("%s/security", prefix)},
		{Name: "dir-uploadcache-data", MountPath: fmt.Sprintf("%s/uploads", prefix)},
		{Name: "dir-uploadcache-key", MountPath: fmt.Sprintf("%s/keys", prefix)},
	}
	return volumesMounts
}

func (c *Creater) getMountPrefixPath() string {
	blackduckVersion := hubutils.GetHubVersion(c.hubSpec.Environs)
	versions, err := hubutils.ParseImageVersion(blackduckVersion)
	if len(versions) == 4 {
		version1, _ := strconv.Atoi(versions[1])
		version3, _ := strconv.Atoi(versions[3])
		if err == nil && version1 >= 1 && version3 > 3 {
			return "/opt/blackduck/hub/blackduck-upload-cache"
		}
	}
	return "/opt/blackduck/hub/hub-upload-cache"
}

// GetUploadCacheService will return the uploadCache service
func (c *Creater) GetUploadCacheService() *components.Service {
	return util.CreateServiceWithMultiplePort("uploadcache", "uploadcache", c.hubSpec.Namespace, []string{uploadCachePort1, uploadCachePort2},
		horizonapi.ClusterIPServiceTypeDefault)
}
