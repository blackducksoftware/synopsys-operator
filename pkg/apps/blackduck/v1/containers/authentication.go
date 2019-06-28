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

// GetAuthenticationDeployment will return the authentication deployment
func (c *Creater) GetAuthenticationDeployment(imageName string) (*components.ReplicationController, error) {
	volumeMounts := c.getAuthenticationVolumeMounts()
	var authEnvs []*horizonapi.EnvConfig
	authEnvs = append(authEnvs, c.getHubDBConfigEnv())
	authEnvs = append(authEnvs, c.getHubConfigEnv())
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.AuthenticationHubMaxMemory})

	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "authentication", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   authEnvs,
		VolumeMounts: volumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: authenticationPort, Protocol: horizonapi.ProtocolTCP}},
	}
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
		Labels:              c.GetVersionLabel("authentication"),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs("authentication"),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "authentication"), Replicas: util.IntToInt32(1)},
		podConfig, c.GetLabel("authentication"))
}

// getAuthenticationVolumes will return the authentication volumes
func (c *Creater) getAuthenticationVolumes() []*components.Volume {
	hubAuthSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication-security")

	var hubAuthVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		hubAuthVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-authentication", c.getPVCName("authentication"))
	} else {
		hubAuthVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	}

	volumes := []*components.Volume{hubAuthVolume, c.getDBSecretVolume(), hubAuthSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, c.getProxyVolume())
	}

	// Custom CA auth
	if len(c.blackDuck.Spec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "auth-custom-ca"), 0444)
		volumes = append(volumes, authCustomCaVolume)
	}
	return volumes
}

// getAuthenticationVolumeMounts will return the authentication volume mounts
func (c *Creater) getAuthenticationVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
		{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/ldap"},
		{Name: "dir-authentication-security", MountPath: "/opt/blackduck/hub/hub-authentication/security"},
	}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
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

// GetAuthenticationService will return the authentication service
func (c *Creater) GetAuthenticationService() *components.Service {
	return util.CreateService(util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "authentication"), c.GetLabel("authentication"), c.blackDuck.Spec.Namespace, authenticationPort, authenticationPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("authentication"))
}
