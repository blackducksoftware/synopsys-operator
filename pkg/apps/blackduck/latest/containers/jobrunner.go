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
	apputil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetJobRunnerDeployment will return the job runner deployment
func (c *Creater) GetJobRunnerDeployment(imageName string) (*components.Deployment, error) {
	podName := "jobrunner"

	jobRunnerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-jobrunner")
	jobRunnerEnvs := []*horizonapi.EnvConfig{c.getHubConfigEnv(), c.getHubDBConfigEnv()}
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.JobRunnerHubMaxMemory})
	jobRunnerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: podName, Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: c.hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: jobRunnerMinCPUUsage, MaxCPU: jobRunnerMaxCPUUsage},
		EnvConfigs: jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
			{Name: "dir-jobrunner", MountPath: "/opt/blackduck/hub/jobrunner/security"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: jobRunnerPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackDuck.Spec.LivenessProbes {
		jobRunnerContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	jobRunnerVolumes := []*components.Volume{c.getDBSecretVolume(), jobRunnerEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		jobRunnerContainerConfig.VolumeMounts = append(jobRunnerContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		jobRunnerVolumes = append(jobRunnerVolumes, c.getProxyVolume())
	}

	podConfig := &util.PodConfig{
		Volumes:             jobRunnerVolumes,
		Containers:          []*util.Container{jobRunnerContainerConfig},
		Labels:              c.GetVersionLabel(podName),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs(podName),
		ServiceAccount:      util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "service-account"),
	}

	if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.PullSecrets) > 0 {
		podConfig.ImagePullSecrets = c.blackDuck.Spec.RegistryConfiguration.PullSecrets
	}

	apputil.ConfigurePodConfigSecurityContext(podConfig, c.blackDuck.Spec.SecurityContexts, "blackduck-jobrunner", c.config.IsOpenshift)

	return util.CreateDeploymentFromContainer(
		&horizonapi.DeploymentConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, podName), Replicas: c.hubContainerFlavor.JobRunnerReplicas},
		podConfig, c.GetLabel(podName))
}
