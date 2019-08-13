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

// GetScanDeployment will return the scan deployment
func (c *Creater) GetScanDeployment(imageName string) (*components.ReplicationController, error) {
	scannerEnvs := []*horizonapi.EnvConfig{c.getHubConfigEnv(), c.getHubDBConfigEnv()}
	scannerEnvs = append(scannerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.ScanHubMaxMemory})
	hubScanEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "scan", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.ScanMemoryLimit, MaxMem: c.hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
			},
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_USER_PASSWORD_FILE",
			},
			{
				Name:      "dir-scan",
				MountPath: "/opt/blackduck/hub/hub-scan/security",
			},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: scannerPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackDuck.Spec.LivenessProbes {
		hubScanContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-scan/security/root.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	hubScanVolumes := []*components.Volume{hubScanEmptyDir, c.getDBSecretVolume()}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		hubScanContainerConfig.VolumeMounts = append(hubScanContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		hubScanVolumes = append(hubScanVolumes, c.getProxyVolume())
	}

	podConfig := &util.PodConfig{
		Volumes:             hubScanVolumes,
		Containers:          []*util.Container{hubScanContainerConfig},
		Labels:              c.GetVersionLabel("scan"),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs("scan"),
	}

	if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.PullSecrets) > 0 {
		podConfig.ImagePullSecrets = c.blackDuck.Spec.RegistryConfiguration.PullSecrets
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "scan"), Replicas: c.hubContainerFlavor.ScanReplicas},
		podConfig, c.GetLabel("scan"))
}

// GetScanService will return the scan service
func (c *Creater) GetScanService() *components.Service {
	return util.CreateService(util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "scan"), c.GetLabel("scan"), c.blackDuck.Spec.Namespace, scannerPort, scannerPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("scan"))
}
