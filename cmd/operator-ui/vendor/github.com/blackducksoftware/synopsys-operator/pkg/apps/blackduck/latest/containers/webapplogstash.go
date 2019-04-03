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

// GetWebappLogstashDeployment will return the webapp and logstash deployment
func (c *Creater) GetWebappLogstashDeployment() *components.ReplicationController {
	webappEnvs := []*horizonapi.EnvConfig{c.getHubConfigEnv(), c.getHubDBConfigEnv()}
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.WebappHubMaxMemory})

	webappVolumeMounts := c.getWebappVolumeMounts()

	webappContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: c.getImageTag("blackduck-webapp"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.WebappMemoryLimit, MaxMem: c.hubContainerFlavor.WebappMemoryLimit, MinCPU: c.hubContainerFlavor.WebappCPULimit,
			MaxCPU: c.hubContainerFlavor.WebappCPULimit},
		EnvConfigs:   webappEnvs,
		VolumeMounts: webappVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: webappPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		webappContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-webapp/security/root.crt",
					"/opt/blackduck/hub/hub-webapp/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-webapp/security/blackduck_system.key",
				},
			},
			Delay:           360,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 1000,
		}}
	}

	c.PostEditContainer(webappContainerConfig)

	logstashVolumeMounts := c.getLogstashVolumeMounts()

	logstashContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: c.getImageTag("blackduck-logstash"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.LogstashMemoryLimit, MaxMem: c.hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: logstashVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: logstashPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		logstashContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:9600/"}},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 1000,
		}}
	}

	c.PostEditContainer(logstashContainerConfig)

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-webapp") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine-webapp", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/hub-webapp/ldap"}},
			VolumeMounts:    webappVolumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-logstash") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine-logstash", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /var/lib/logstash/data"}},
			VolumeMounts:    logstashVolumeMounts,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	webappLogstash := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "webapp-logstash", Replicas: util.IntToInt32(1)},
		"", []*util.Container{webappContainerConfig, logstashContainerConfig}, c.getWebappLogtashVolumes(),
		initContainers, []horizonapi.AffinityConfig{}, c.GetVersionLabel("webapp-logstash"), c.GetLabel("webapp-logstash"))
	return webappLogstash
}

// getWebappLogtashVolumes will return the webapp and logstash volumes
func (c *Creater) getWebappLogtashVolumes() []*components.Volume {
	webappSecurityEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp-security")
	var webappVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-webapp") {
		webappVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-webapp", "blackduck-webapp")
	} else {
		webappVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	}

	var logstashVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-logstash") {
		logstashVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-logstash", "blackduck-logstash")
	} else {
		logstashVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	}

	volumes := []*components.Volume{webappSecurityEmptyDir, webappVolume, logstashVolume, c.getDBSecretVolume()}
	// Mount the HTTPS proxy certificate if provided
	if len(c.hubSpec.ProxyCertificate) > 0 {
		volumes = append(volumes, c.getProxyVolume())
	}

	return volumes
}

// getLogstashVolumeMounts will return the Logstash volume mounts
func (c *Creater) getLogstashVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-logstash", MountPath: "/var/lib/logstash/data"},
	}
	return volumesMounts
}

// getWebappVolumeMounts will return the Webapp volume mounts
func (c *Creater) getWebappVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
		{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
		{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/ldap"},
		{Name: "dir-webapp-security", MountPath: "/opt/blackduck/hub/hub-webapp/security"},
		{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs"},
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

// GetWebAppService will return the webapp service
func (c *Creater) GetWebAppService() *components.Service {
	return util.CreateService("webapp", c.GetLabel("webapp-logstash"), c.hubSpec.Namespace, webappPort, webappPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("webapp-logstash"))
}

// GetLogStashService will return the logstash service
func (c *Creater) GetLogStashService() *components.Service {
	return util.CreateService("logstash", c.GetLabel("webapp-logstash"), c.hubSpec.Namespace, logstashPort, logstashPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("webapp-logstash"))
}
