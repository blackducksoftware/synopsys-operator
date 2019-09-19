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

// GetBinaryScannerDeployment will return the binary scanner deployment
func (c *Creater) GetBinaryScannerDeployment(imageName string) (*components.Deployment, error) {
	binaryScannerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "binaryscanner", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.BinaryScannerMemoryLimit,
			MaxMem: c.hubContainerFlavor.BinaryScannerMemoryLimit, MinCPU: binaryScannerMinCPUUsage, MaxCPU: binaryScannerMaxCPUUsage,
			Command: []string{"/docker-entrypoint.sh"}},
		EnvConfigs: []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: binaryScannerPort, Protocol: horizonapi.ProtocolTCP}},
	}

	podConfig := &util.PodConfig{
		Containers:          []*util.Container{binaryScannerContainerConfig},
		Labels:              c.GetVersionLabel("binaryscanner"),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs("binaryscanner"),
	}

	if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.PullSecrets) > 0 {
		podConfig.ImagePullSecrets = c.blackDuck.Spec.RegistryConfiguration.PullSecrets
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateDeploymentFromContainer(
		&horizonapi.DeploymentConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "binaryscanner"), Replicas: util.IntToInt32(1)},
		podConfig, c.GetLabel("binaryscanner"))
}
