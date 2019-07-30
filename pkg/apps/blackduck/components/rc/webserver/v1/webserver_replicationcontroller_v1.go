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
	store.Register(types.BlackDuckWebserverRCV1, NewBdReplicationController)
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
	containerConfig, ok := c.Containers[types.WebserverContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.WebserverContainerName)
	}

	webServerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webserver", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts:    c.getWebserverVolumeMounts(),
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(webServerContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
		webServerContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/health-checks/liveness", "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE"},
			},
			Delay:           180,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getWebserverVolumes(),
		Containers:          []*util.Container{webServerContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("webserver", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("webserver", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webserver"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("webserver", c.blackDuck.Name))
}

// getWebserverVolumes will return the authentication volumes
func (c *BdReplicationController) getWebserverVolumes() []*components.Volume {
	webServerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webserver")
	webServerSecretVol, _ := util.CreateSecretVolume("certificate", apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webserver-certificate"), 0444)

	volumes := []*components.Volume{webServerEmptyDir, webServerSecretVol}

	// Custom CA auth
	if len(c.blackDuck.Spec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "auth-custom-ca"), 0444)
		volumes = append(volumes, authCustomCaVolume)
	}
	return volumes
}

// getWebserverVolumeMounts will return the authentication volume mounts
func (c *BdReplicationController) getWebserverVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security"},
		{Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE", SubPath: "WEBSERVER_CUSTOM_CERT_FILE"},
		{Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_KEY_FILE", SubPath: "WEBSERVER_CUSTOM_KEY_FILE"},
	}

	if len(c.blackDuck.Spec.AuthCustomCA) > 1 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "auth-custom-ca",
			MountPath: "/tmp/secrets/AUTH_CUSTOM_CA",
			SubPath:   "AUTH_CUSTOM_CA",
		})
	}

	return volumesMounts
}
