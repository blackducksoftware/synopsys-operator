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

// GetWebserverDeployment will return the webserver deployment
func (c *Creater) GetWebserverDeployment() *components.ReplicationController {
	webServerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webserver", Image: c.getImageTag("blackduck-nginx"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.WebserverMemoryLimit,
			MaxMem: c.hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: c.getWebserverVolumeMounts(),
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: webserverPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		webServerContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/health-checks/liveness", "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE"}},
			Delay:           180,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	c.PostEditContainer(webServerContainerConfig)

	webserver := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{
		Namespace: c.hubSpec.Namespace, Name: "webserver", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{webServerContainerConfig}, c.getWebserverVolumes(), []*util.Container{}, []horizonapi.AffinityConfig{}, c.GetVersionLabel("webserver"), c.GetLabel("webserver"))
	// log.Infof("webserver : %v\n", webserver.GetObj())
	return webserver
}

// getWebserverVolumes will return the authentication volumes
func (c *Creater) getWebserverVolumes() []*components.Volume {
	webServerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webserver")
	webServerSecretVol, _ := util.CreateSecretVolume("certificate", "blackduck-certificate", 0777)

	volumes := []*components.Volume{webServerEmptyDir, webServerSecretVol}

	// Custom CA auth
	if len(c.hubSpec.AuthCustomCA) > 1 {
		authCustomCaVolume, _ := util.CreateSecretVolume("auth-custom-ca", "auth-custom-ca", 0777)
		volumes = append(volumes, authCustomCaVolume)
	}
	return volumes
}

// getWebserverVolumeMounts will return the authentication volume mounts
func (c *Creater) getWebserverVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security"},
		{Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE", SubPath: "WEBSERVER_CUSTOM_CERT_FILE"},
		{Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_KEY_FILE", SubPath: "WEBSERVER_CUSTOM_KEY_FILE"},
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

// GetWebServerService will return the webserver service
func (c *Creater) GetWebServerService() *components.Service {
	return util.CreateService("webserver", c.GetLabel("webserver"), c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("webserver"))
}

// GetWebServerNodePortService will return the webserver nodeport service
func (c *Creater) GetWebServerNodePortService() *components.Service {
	return util.CreateService("webserver-np", c.GetLabel("webserver"), c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeNodePort, c.GetLabel("webserver-np"))
}

// GetWebServerLoadBalancerService will return the webserver loadbalancer service
func (c *Creater) GetWebServerLoadBalancerService() *components.Service {
	return util.CreateService("webserver-lb", c.GetLabel("webserver"), c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeLoadBalancer, c.GetLabel("webserver-lb"))
}
