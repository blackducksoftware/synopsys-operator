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

// GetDocumentationDeployment will return the documentation deployment
func (c *Creater) GetDocumentationDeployment(imageName string) (*components.ReplicationController, error) {
	documentationEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-documentation")
	documentationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.DocumentationMemoryLimit, MaxMem: c.hubContainerFlavor.DocumentationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "dir-documentation", MountPath: "/opt/blackduck/hub/hub-documentation/security"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: documentationPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		documentationContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://127.0.0.1:8443/hubdoc/health-checks/liveness", "/opt/blackduck/hub/hub-documentation/security/root.crt"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}
	c.PostEditContainer(documentationContainerConfig)

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, util.BlackDuckName, "documentation", c.config.IsClusterScoped), Replicas: util.IntToInt32(1)},
		&util.PodConfig{
			Volumes:             []*components.Volume{documentationEmptyDir},
			Containers:          []*util.Container{documentationContainerConfig},
			ImagePullSecrets:    c.hubSpec.RegistryConfiguration.PullSecrets,
			Labels:              c.GetVersionLabel("documentation"),
			NodeAffinityConfigs: c.GetNodeAffinityConfigs("documentation"),
		}, c.GetLabel("documentation"))
}

// GetDocumentationService will return the cfssl service
func (c *Creater) GetDocumentationService() *components.Service {
	return util.CreateService(util.GetResourceName(c.name, util.BlackDuckName, "documentation", c.config.IsClusterScoped), c.GetLabel("documentation"), c.hubSpec.Namespace, documentationPort, documentationPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("documentation"))
}
