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

// GetWebserverDeployment will return the webserver deployment
func (c *Creater) GetWebserverDeployment() *components.ReplicationController {
	webServerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webserver")
	webServerSecretVol, _ := util.CreateSecretVolume("certificate", "blackduck-certificate", 0777)
	webServerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/%s-nginx:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("nginx")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.WebserverMemoryLimit, MaxMem: c.hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security"},
			{Name: "certificate", MountPath: "/tmp/secrets"},
		},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webserverPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/health-checks/liveness", "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE"}},
		// 	Delay:           180,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}

	c.PostEditContainer(webServerContainerConfig)

	webserver := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "webserver",
		Replicas: util.IntToInt32(1)}, c.hubSpec.Namespace, []*util.Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir, webServerSecretVol},
		[]*util.Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("webserver : %v\n", webserver.GetObj())
	return webserver
}

// GetWebServerService will return the webserver service
func (c *Creater) GetWebServerService() *components.Service {
	return util.CreateService("webserver", "webserver", c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeDefault)
}

// GetWebServerNodePortService will return the webserver nodeport service
func (c *Creater) GetWebServerNodePortService() *components.Service {
	return util.CreateService("webserver-np", "webserver", c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeNodePort)
}

// GetWebServerLoadBalancerService will return the webserver loadbalancer service
func (c *Creater) GetWebServerLoadBalancerService() *components.Service {
	return util.CreateService("webserver-lb", "webserver", c.hubSpec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeLoadBalancer)
}
