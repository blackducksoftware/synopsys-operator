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

// GetRegistrationDeployment will return the registration deployment
func (c *Creater) GetRegistrationDeployment() *components.ReplicationController {
	registrationEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	registrationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/%s-registration:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("registration")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.RegistrationMemoryLimit, MaxMem: c.hubContainerFlavor.RegistrationMemoryLimit, MinCPU: registrationMinCPUUsage, MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: registrationPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/registration/health-checks/liveness", "/opt/blackduck/hub/hub-registration/security/root.crt"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}
	registration := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "registration", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	return registration
}

// GetRegistrationService will return the registration service
func (c *Creater) GetRegistrationService() *components.Service {
	return util.CreateService("registration", "registration", c.hubSpec.Namespace, registrationPort, registrationPort, horizonapi.ClusterIPServiceTypeDefault)
}
