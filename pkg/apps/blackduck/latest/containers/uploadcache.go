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
func (c *Creater) GetUploadCacheDeployment(imageName string) (*components.ReplicationController, error) {
	volumeMounts := c.getUploadCacheVolumeMounts()

	uploadCacheContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "uploadcache", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.UploadCacheMemoryLimit, MaxMem: c.hubContainerFlavor.UploadCacheMemoryLimit,
			MinCPU: "", MaxCPU: ""},
		EnvConfigs: []*horizonapi.EnvConfig{
			c.getHubConfigEnv(),
			// &horizonapi.EnvConfig{NameOrPrefix: "BLACKDUCK_UPLOAD_CACHE_HOST", Type: horizonapi.EnvVal, KeyOrVal: "0.0.0.0"},
		},
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: uploadCachePort1, Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: uploadCachePort2, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		uploadCacheContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"curl", "--insecure", "-X", "GET", "--verbose", "http://localhost:8086/live?full=1"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 5,
		}}
	}

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/blackduck-upload-cache/security /opt/blackduck/hub/blackduck-upload-cache/keys /opt/blackduck/hub/blackduck-upload-cache/uploads"}},
			VolumeMounts:    volumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(uploadCacheContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, "uploadcache", c.config.IsClusterScoped), Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			Volumes:             c.getUploadCacheVolumes(),
			Containers:          []*util.Container{uploadCacheContainerConfig},
			InitContainers:      initContainers,
			ImagePullSecrets:    c.hubSpec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("uploadcache"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("uploadcache"),
		}, c.GetLabel("uploadcache"))
}

// getUploadCacheVolumes will return the uploadCache volumes
func (c *Creater) getUploadCacheVolumes() []*components.Volume {
	uploadCacheSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-security")
	sealKeySecretVol, _ := util.CreateSecretVolume("dir-seal-key", util.GetResourceName(c.name, "upload-cache", c.config.IsClusterScoped), 0444)
	var uploadCacheDataDir *components.Volume
	var uploadCacheDataKey *components.Volume
	if c.hubSpec.PersistentStorage {
		uploadCacheDataDir, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-data", util.GetResourceName(c.name, "blackduck-uploadcache-data", c.config.IsClusterScoped))
		uploadCacheDataKey, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-key", util.GetResourceName(c.name, "blackduck-uploadcache-key", c.config.IsClusterScoped))
	} else {
		uploadCacheDataDir, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-data")
		uploadCacheDataKey, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-key")
	}
	volumes := []*components.Volume{uploadCacheSecurityEmptyDir, uploadCacheDataDir, uploadCacheDataKey, sealKeySecretVol}
	return volumes
}

// getUploadCacheVolumeMounts will return the uploadCache volume mounts
func (c *Creater) getUploadCacheVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-uploadcache-security", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/security"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/uploads"},
		{Name: "dir-uploadcache-key", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/keys"},
		{Name: "dir-seal-key", MountPath: "/tmp/secrets"},
	}
	return volumesMounts
}

// GetUploadCacheService will return the uploadCache service
func (c *Creater) GetUploadCacheService() *components.Service {
	// TODO: remove GetResourceName method until the HUB-20412 is fixed. once it if fixed, add them back
	return util.CreateServiceWithMultiplePort("uploadcache", c.GetLabel("uploadcache"), c.hubSpec.Namespace, []int32{uploadCachePort1, uploadCachePort2},
		horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("uploadcache"))
}
