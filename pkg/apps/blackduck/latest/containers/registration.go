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

// GetRegistrationDeployment will return the registration deployment
func (c *Creater) GetRegistrationDeployment(imageName string) (*components.ReplicationController, error) {

	volumeMounts := c.getRegistrationVolumeMounts()

	registrationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "registration", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.RegistrationMemoryLimit, MaxMem: c.hubContainerFlavor.RegistrationMemoryLimit, MinCPU: registrationMinCPUUsage, MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: c.getRegistrationVolumeMounts(),
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: registrationPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
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

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/hub-registration/config"}},
			VolumeMounts:    volumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(registrationContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "registration", Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			Volumes:             c.getRegistrationVolumes(),
			Containers:          []*util.Container{registrationContainerConfig},
			InitContainers:      initContainers,
			ImagePullSecrets:    c.hubSpec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("registration"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("registration"),
		}, c.GetLabel("registration"))
}

// getRegistrationVolumes will return the registration volumes
func (c *Creater) getRegistrationVolumes() []*components.Volume {
	var registrationVolume *components.Volume
	registrationSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration-security")

	if c.hubSpec.PersistentStorage {
		registrationVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-registration", "blackduck-registration")
	} else {
		registrationVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	}

	volumes := []*components.Volume{registrationVolume, registrationSecurityEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.hubSpec.ProxyCertificate) > 0 {
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
	if len(c.hubSpec.ProxyCertificate) > 0 {
		volumesMounts = append(volumesMounts, &horizonapi.VolumeMountConfig{
			Name:      "blackduck-proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
	}

	return volumesMounts
}

// GetRegistrationService will return the registration service
func (c *Creater) GetRegistrationService() *components.Service {
	return util.CreateService("registration", c.GetLabel("registration"), c.hubSpec.Namespace, registrationPort, registrationPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("registration"))
}
