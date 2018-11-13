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

// GetCfsslDeployment will return the cfssl deployment
func (c *Creater) GetCfsslDeployment() *components.ReplicationController {
	cfsslEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: c.getFullContainerName("cfssl"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.CfsslMemoryLimit, MaxMem: c.hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-cfssl", MountPath: "/etc/cfssl"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: cfsslPort, Protocol: horizonapi.ProtocolTCP},
	}
	c.PostEditContainer(cfsslContainerConfig)

	cfssl := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "cfssl", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})

	return cfssl
}

// GetCfsslService will return the cfssl service
func (c *Creater) GetCfsslService() *components.Service {
	return util.CreateService("cfssl", "cfssl", c.hubSpec.Namespace, cfsslPort, cfsslPort, horizonapi.ClusterIPServiceTypeDefault)
}
