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

// GetJobRunnerDeployment will return the job runner deployment
func (c *Creater) GetJobRunnerDeployment() *components.ReplicationController {
	jobRunnerEnvs := c.allConfigEnv
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "jobrunner-mem", FromName: "hub-config-resources"})
	jobRunnerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "jobrunner", Image: fmt.Sprintf("%s/%s/%s-jobrunner:%s", c.hubSpec.DockerRegistry, c.hubSpec.DockerRepo, c.hubSpec.ImagePrefix, c.getTag("jobrunner")),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: c.hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: jonRunnerMinCPUUsage, MaxCPU: jonRunnerMaxCPUUsage},
		EnvConfigs:   jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "db-passwords", MountPath: "/tmp/secrets"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: jobRunnerPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}
	c.PostEditContainer(jobRunnerContainerConfig)

	jobRunner := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "jobrunner", Replicas: c.hubContainerFlavor.JobRunnerReplicas}, "",
		[]*util.Container{jobRunnerContainerConfig}, []*components.Volume{c.dbSecretVolume, c.dbEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	return jobRunner
}
