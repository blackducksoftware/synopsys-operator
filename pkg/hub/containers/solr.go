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
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
)

// GetSolrDeployment will return the solr deployment
func (c *Creater) GetSolrDeployment() *components.ReplicationController {
	solrEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: c.getFullContainerName("solr"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.SolrMemoryLimit, MaxMem: c.hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   c.hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data"}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: solrPort, Protocol: horizonapi.ProtocolTCP},
		// LivenessProbeConfigs: []*horizonapi.ProbeConfig{{
		// 	ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8983/solr/project/admin/ping?wt=json"}},
		// 	Delay:           240,
		// 	Interval:        30,
		// 	Timeout:         10,
		// 	MinCountFailure: 10,
		// }},
	}
	c.PostEditContainer(solrContainerConfig)

	solr := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "solr", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})

	return solr
}

// GetSolrService will return the solr service
func (c *Creater) GetSolrService() *components.Service {
	return util.CreateService("solr", "solr", c.hubSpec.Namespace, solrPort, solrPort, horizonapi.ClusterIPServiceTypeDefault)
}
