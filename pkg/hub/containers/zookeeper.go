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

// GetZookeeperDeployment will return the zookeeper deployment
func (c *Creater) GetZookeeperDeployment() *components.ReplicationController {
	zookeeperEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper")
	zookeeperContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/%s-zookeeper:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("zookeeper")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: c.hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: zookeeperMinCPUUsage, MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: zookeeperPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"zkServer.sh", "status", "/opt/blackduck/zookeeper/conf/zoo.cfg"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}
	c.PostEditContainer(zookeeperContainerConfig)

	zookeeper := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "zookeeper", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{zookeeperContainerConfig}, []*components.Volume{zookeeperEmptyDir}, []*util.Container{}, []horizonapi.AffinityConfig{})

	return zookeeper
}

// GetZookeeperService will return the zookeeper service
func (c *Creater) GetZookeeperService() *components.Service {
	return util.CreateService("zookeeper", "zookeeper", c.hubSpec.Namespace, zookeeperPort, zookeeperPort, horizonapi.ClusterIPServiceTypeDefault)
}
