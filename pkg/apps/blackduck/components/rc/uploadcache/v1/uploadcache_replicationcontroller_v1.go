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

package v1

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdReplicationController holds the Black Duck RC configuration
type BdReplicationController struct {
	*types.PodResource
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.BlackDuckUploadCacheRCV1, NewBdReplicationController)
}

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(podResource *types.PodResource, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{PodResource: podResource, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[types.UploadCacheContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.UploadCacheContainerName)
	}
	volumeMounts := c.getUploadCacheVolumeMounts()

	uploadCacheContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "uploadcache", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts:    volumeMounts,
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(9443), Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: int32(9444), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(uploadCacheContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
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
	podConfig := &util.PodConfig{
		Volumes:             c.getUploadCacheVolumes(),
		Containers:          []*util.Container{uploadCacheContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("uploadcache", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("uploadcache", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "uploadcache"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("uploadcache", c.blackDuck.Name))
}

// getUploadCacheVolumes will return the uploadCache volumes
func (c *BdReplicationController) getUploadCacheVolumes() []*components.Volume {
	uploadCacheSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-security")
	sealKeySecretVol, _ := util.CreateSecretVolume("dir-seal-key", apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "upload-cache"), 0444)
	var uploadCacheDataDir *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		uploadCacheDataDir, _ = util.CreatePersistentVolumeClaimVolume("dir-uploadcache-data", utils.GetPVCName("uploadcache-data", c.blackDuck))
	} else {
		uploadCacheDataDir, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-uploadcache-data")
	}
	volumes := []*components.Volume{uploadCacheSecurityEmptyDir, uploadCacheDataDir, sealKeySecretVol}
	return volumes
}

// getUploadCacheVolumeMounts will return the uploadCache volume mounts
func (c *BdReplicationController) getUploadCacheVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-uploadcache-security", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/security"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/uploads", SubPath: "uploads"},
		{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache/keys", SubPath: "keys"},
		{Name: "dir-seal-key", MountPath: "/tmp/secrets"},
	}
	return volumesMounts
}
