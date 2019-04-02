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

// GetSolrDeployment will return the solr deployment
func (c *Creater) GetSolrDeployment() *components.ReplicationController {
	solrVolumeMount := c.getSolrVolumeMounts()
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: c.getImageTag("blackduck-solr"),
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.SolrMemoryLimit, MaxMem: c.hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: solrVolumeMount,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: solrPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.hubSpec.LivenessProbes {
		solrContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8983/solr/project/admin/ping?wt=json"}},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-solr") {
		initContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/solr/cores.data"}},
			VolumeMounts:    solrVolumeMount,
		}
		initContainers = append(initContainers, initContainerConfig)
	}

	c.PostEditContainer(solrContainerConfig)

	solr := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "solr", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{solrContainerConfig}, c.getSolrVolumes(), initContainers,
		[]horizonapi.AffinityConfig{}, c.GetVersionLabel("solr"), c.GetLabel("solr"))

	return solr
}

// getSolrVolumes will return the solr volumes
func (c *Creater) getSolrVolumes() []*components.Volume {
	var solrVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-solr") {
		solrVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-solr", "blackduck-solr")
	} else {
		solrVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	}

	volumes := []*components.Volume{solrVolume}
	return volumes
}

// getSolrVolumeMounts will return the solr volume mounts
func (c *Creater) getSolrVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data"},
	}
	return volumesMounts
}

// GetSolrService will return the solr service
func (c *Creater) GetSolrService() *components.Service {
	return util.CreateService("solr", c.GetLabel("solr"), c.hubSpec.Namespace, solrPort, solrPort, horizonapi.ClusterIPServiceTypeDefault, c.GetVersionLabel("solr"))
}
