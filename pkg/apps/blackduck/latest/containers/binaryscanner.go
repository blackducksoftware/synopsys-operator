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

// GetBinaryScannerDeployment will return the binary scanner deployment
func (c *Creater) GetBinaryScannerDeployment(imageName string) (*components.ReplicationController, error) {
	binaryScannerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "binaryscanner", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.BinaryScannerMemoryLimit,
			MaxMem: c.hubContainerFlavor.BinaryScannerMemoryLimit, MinCPU: binaryScannerMinCPUUsage, MaxCPU: binaryScannerMaxCPUUsage,
			Command: []string{"/docker-entrypoint.sh"}},
		EnvConfigs: []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: binaryScannerPort, Protocol: horizonapi.ProtocolTCP}},
	}

	c.PostEditContainer(binaryScannerContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "binaryscanner", Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			ServiceAccount:      c.hubSpec.Namespace,
			Containers:          []*util.Container{binaryScannerContainerConfig},
			ImagePullSecrets:    c.hubSpec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("binaryscanner"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("binaryscanner"),
		}, c.GetLabel("binaryscanner"))
}
