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
	store.Register(types.BlackDuckAuthenticationRCV1, NewBdReplicationController)
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
	containerConfig, ok := c.Containers[types.AuthenticationContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.AuthenticationContainerName)
	}

	// hubMaxMemory is the amount of memory allocated to the JVM. We keep 512mb for alpine
	hubMaxMemory := 512
	if containerConfig.MaxMem != nil && *containerConfig.MaxMem > 512 {
		hubMaxMemory = int(*containerConfig.MaxMem - 512)
	}

	volumeMounts := c.getAuthenticationVolumeMounts()
	var authEnvs []*horizonapi.EnvConfig
	authEnvs = append(authEnvs, utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name))
	authEnvs = append(authEnvs, utils.GetBlackDuckConfigEnv(c.blackDuck.Name))
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dm", hubMaxMemory)})
	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "authentication", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      authEnvs,
		VolumeMounts:    volumeMounts,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(hubAuthContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
		hubAuthContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-authentication/security/root.crt",
					"/opt/blackduck/hub/hub-authentication/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-authentication/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getAuthenticationVolumes(),
		Containers:          []*util.Container{hubAuthContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("authentication", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("authentication", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "authentication"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("authentication", c.blackDuck.Name))
}

// getAuthenticationVolumes will return the authentication volumes
func (c *BdReplicationController) getAuthenticationVolumes() []*components.Volume {
	hubAuthSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication-security")

	var hubAuthVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		hubAuthVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-authentication", utils.GetPVCName("authentication", c.blackDuck))
	} else {
		hubAuthVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	}

	volumes := []*components.Volume{hubAuthVolume, utils.GetDBSecretVolume(c.blackDuck.Name), hubAuthSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, utils.GetProxyVolume(c.blackDuck.Name))
	}

	// Custom CA auth
	if len(c.blackDuck.Spec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", "auth-custom-ca", 0444)
		volumes = append(volumes, authCustomCaVolume)
	}
	return volumes
}

// getAuthenticationVolumeMounts will return the authentication volume mounts
func (c *BdReplicationController) getAuthenticationVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
		{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/ldap"},
		{Name: "dir-authentication-security", MountPath: "/opt/blackduck/hub/hub-authentication/security"},
	}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "blackduck-proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
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
