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

// GetSolrDeployment will return the solr deployment
func (c *Creater) GetSolrDeployment(imageName string) (*components.Deployment, error) {
	solrVolumeMount := c.getSolrVolumeMounts()
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: imageName,
			PullPolicy: horizonapi.PullAlways, MinMem: c.hubContainerFlavor.SolrMemoryLimit, MaxMem: c.hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   []*horizonapi.EnvConfig{c.getHubConfigEnv()},
		VolumeMounts: solrVolumeMount,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: solrPort, Protocol: horizonapi.ProtocolTCP}},
	}

	if c.blackDuck.Spec.LivenessProbes {
		solrContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8983/solr/project/admin/ping?wt=json"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	podConfig := &util.PodConfig{
		Volumes:             c.getSolrVolumes(),
		Containers:          []*util.Container{solrContainerConfig},
		Labels:              c.GetVersionLabel("solr"),
		NodeAffinityConfigs: c.GetNodeAffinityConfigs("solr"),
	}

	if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.PullSecrets) > 0 {
		podConfig.ImagePullSecrets = c.blackDuck.Spec.RegistryConfiguration.PullSecrets
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateDeploymentFromContainer(
		&horizonapi.DeploymentConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "solr"), Replicas: util.IntToInt32(1)},
		podConfig, c.GetLabel("solr"))
}

// getSolrVolumes will return the solr volumes
func (c *Creater) getSolrVolumes() []*components.Volume {
	var solrVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		solrVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-solr", c.getPVCName("solr"))
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
	return util.CreateService(util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "solr"), c.GetLabel("solr"), c.blackDuck.Spec.Namespace, solrPort, solrPort, horizonapi.ServiceTypeServiceIP, c.GetVersionLabel("solr"))
}
