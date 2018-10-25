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

// GetAuthenticationDeployment will return the authentication deployment
func (c *Creater) GetAuthenticationDeployment() *components.ReplicationController {
	authEnvs := c.allConfigEnv
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: c.hubContainerFlavor.AuthenticationHubMaxMemory})
	// hubAuthGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-authentication", fmt.Sprintf("%s-%s", "authentication-disk", c.hubSpec.Namespace), "ext4")
	hubAuthEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/%s-authentication:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("authentication")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: c.hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets"},
			{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security"}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: authenticationPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://127.0.0.1:8443/api/health-checks/liveness", "/opt/blackduck/hub/hub-authentication/security/root.crt"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}
	hubAuth := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "hub-authentication", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, c.dbSecretVolume, c.dbEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	return hubAuth
}

// GetAuthenticationService will return the authentication service
func (c *Creater) GetAuthenticationService() *components.Service {
	return util.CreateService("authentication", "hub-authentication", c.hubSpec.Namespace, authenticationPort, authenticationPort, horizonapi.ClusterIPServiceTypeDefault)
}
