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

// GetRegistrationDeployment will return the registration deployment
func (c *Creater) GetRegistrationDeployment(imageName string) (*components.Deployment, error) {
	registrationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "registration", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.RegistrationMemoryLimit, MaxMem: c.hubContainerFlavor.RegistrationMemoryLimit, MinCPU: registrationMinCPUUsage, MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: c.getRegistrationVolumeMounts(),
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: registrationPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackDuck.Spec.LivenessProbes {
		registrationContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://localhost:8443/registration/health-checks/liveness",
					"/opt/blackduck/hub/hub-registration/security/root.crt",
					"/opt/blackduck/hub/hub-registration/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-registration/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getRegistrationVolumes(),
		Containers:          []*util.Container{registrationContainerConfig},
		Labels:              c.GetVersionLabel("registration"),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs("registration"),
	}

	if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.PullSecrets) > 0 {
		podConfig.ImagePullSecrets = c.blackDuck.Spec.RegistryConfiguration.PullSecrets
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateDeploymentFromContainer(
		&horizonapi.DeploymentConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "registration"), Replicas: util.IntToInt32(1)},
		podConfig, c.GetLabel("registration"))
}

// getRegistrationVolumes will return the registration volumes
func (c *Creater) getRegistrationVolumes() []*components.Volume {
	var registrationVolume *components.Volume
	registrationSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration-security")

	if c.blackDuck.Spec.PersistentStorage {
		registrationVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-registration", c.getPVCName("registration"))
	} else {
		registrationVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	}

	volumes := []*components.Volume{registrationVolume, registrationSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumes = append(volumes, c.getProxyVolume())
	}
	return volumes
}

// getRegistrationVolumeMounts will return the registration volume mounts
func (c *Creater) getRegistrationVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config"},
		{Name: "dir-registration-security", MountPath: "/opt/blackduck/hub/hub-registration/security"},
	}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
	}

	return volumesMounts
}

// GetRegistrationService will return the registration service
func (c *Creater) GetRegistrationService() *components.Service {
	return util.CreateService(util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "registration"), c.GetLabel("registration"), c.blackDuck.Spec.Namespace, registrationPort, registrationPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("registration"))
}
