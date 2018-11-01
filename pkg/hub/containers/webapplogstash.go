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
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
)

// GetWebappLogstashDeployment will return the webapp and logstash deployment
func (c *Creater) GetWebappLogstashDeployment() *components.ReplicationController {
	webappEnvs := c.allConfigEnv
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", c.hubSpec.Namespace), "ext4")
	webappEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	webappContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/%s-webapp:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("webapp")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.WebappMemoryLimit, MaxMem: c.hubContainerFlavor.WebappMemoryLimit, MinCPU: c.hubContainerFlavor.WebappCPULimit,
			MaxCPU: c.hubContainerFlavor.WebappCPULimit},
		EnvConfigs: webappEnvs,
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
				Name:      "dir-webapp",
				MountPath: "/opt/blackduck/hub/hub-webapp/security",
			},
			{
				Name:      "dir-logstash",
				MountPath: "/opt/blackduck/hub/logs",
			},
		},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webappPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://127.0.0.1:8443/api/health-checks/liveness", "/opt/blackduck/hub/hub-webapp/security/root.crt"}},
		// 	Delay:           360,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 1000,
		// }},
	}
	c.PostEditContainer(webappContainerConfig)

	logstashEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	logstashContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/%s-logstash:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("logstash")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.LogstashMemoryLimit, MaxMem: c.hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-logstash", MountPath: "/var/lib/logstash/data"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: logstashPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:9600/"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 1000,
		// }},
	}

	webappLogstashVolumes := []*components.Volume{webappEmptyDir, logstashEmptyDir, c.dbSecretVolume, c.dbEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.hubSpec.ProxyCertificate) > 0 && c.proxySecretVolume != nil {
		webappContainerConfig.VolumeMounts = append(webappContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "blackduck-proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		webappLogstashVolumes = append(webappLogstashVolumes, c.proxySecretVolume)
	}
  
  c.PostEditContainer(logstashContainerConfig)

	webappLogstash := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "webapp-logstash", Replicas: util.IntToInt32(1)},
		"", []*util.Container{webappContainerConfig, logstashContainerConfig}, webappLogstashVolumes,
		[]*util.Container{}, []horizonapi.AffinityConfig{})
	return webappLogstash
}

// GetWebAppService will return the webapp service
func (c *Creater) GetWebAppService() *components.Service {
	return util.CreateService("webapp", "webapp-logstash", c.hubSpec.Namespace, webappPort, webappPort, horizonapi.ClusterIPServiceTypeDefault)
}

// GetLogStashService will return the logstash service
func (c *Creater) GetLogStashService() *components.Service {
	return util.CreateService("logstash", "webapp-logstash", c.hubSpec.Namespace, logstashPort, logstashPort, horizonapi.ClusterIPServiceTypeDefault)
}
