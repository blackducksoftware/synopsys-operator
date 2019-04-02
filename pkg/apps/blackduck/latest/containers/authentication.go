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

// GetAuthenticationDeployment will return the authentication deployment
func (c *Creater) GetAuthenticationDeployment() *components.ReplicationController {
	volumeMounts := c.getAuthenticationVolumeMounts()
	var authEnvs []*horizonapi.EnvConfig
	authEnvs = append(authEnvs, c.getHubDBConfigEnv())
	authEnvs = append(authEnvs, c.getHubConfigEnv())
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.AuthenticationHubMaxMemory})

	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "authentication", Image: c.getImageTag("blackduck-authentication"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   authEnvs,
		VolumeMounts: volumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: authenticationPort, Protocol: horizonapi.ProtocolTCP}},
	}
	if c.hubSpec.LivenessProbes {
		hubAuthContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
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

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-authentication") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/hub-authentication/ldap"}},
			VolumeMounts:    volumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(hubAuthContainerConfig)

	hubAuth := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "authentication", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{hubAuthContainerConfig}, c.getAuthenticationVolumes(), initContainers,
		[]horizonapi.AffinityConfig{}, c.GetVersionLabel("authentication"), c.GetLabel("authentication"))

	return hubAuth
}

// getAuthenticationVolumes will return the authentication volumes
func (c *Creater) getAuthenticationVolumes() []*components.Volume {
	hubAuthSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication-security")

	var hubAuthVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-authentication") {
		hubAuthVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-authentication", "blackduck-authentication")
	} else {
		hubAuthVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	}

	volumes := []*components.Volume{hubAuthVolume, c.getDBSecretVolume(), hubAuthSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.hubSpec.ProxyCertificate) > 0 {
		volumes = append(volumes, c.getProxyVolume())
	}

	// Custom CA auth
	if len(c.hubSpec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", "auth-custom-ca", 0777)
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
	if len(c.hubSpec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "blackduck-proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
	}

	if len(c.hubSpec.AuthCustomCA) > 1 {
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
	return util.CreateService("authentication", c.GetLabel("authentication"), c.hubSpec.Namespace, authenticationPort, authenticationPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("authentication"))
}
