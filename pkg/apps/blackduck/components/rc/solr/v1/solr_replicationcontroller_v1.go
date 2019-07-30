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

package v1

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/rc/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdReplicationController holds the Black Duck RC configuration
type BdReplicationController struct {
	*types.PodResource
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.BlackDuckSolrRCV1, NewBdReplicationController)
}

// NewBdReplicationController returns the Black Duck RC configuration
func NewBdReplicationController(podResource *types.PodResource, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdReplicationController{PodResource: podResource, config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetRc returns the RC
func (c *BdReplicationController) GetRc() (*components.ReplicationController, error) {
	containerConfig, ok := c.Containers[types.SolrContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.SolrContainerName)
	}
	solrVolumeMount := c.getSolrVolumeMounts()
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		VolumeMounts:    solrVolumeMount,
		PortConfig:      []*horizonapi.PortConfig{{ContainerPort: int32(8983), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(solrContainerConfig.ContainerConfig, containerConfig)

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
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("solr", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("solr", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "solr"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("solr", c.blackDuck.Name))
}

// getSolrVolumes will return the solr volumes
func (c *BdReplicationController) getSolrVolumes() []*components.Volume {
	var solrVolume *components.Volume
	if c.blackDuck.Spec.PersistentStorage {
		solrVolume, _ = util.CreatePersistentVolumeClaimVolume("dir-solr", utils.GetPVCName("solr", c.blackDuck))
	} else {
		solrVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	}

	volumes := []*components.Volume{solrVolume}
	return volumes
}

// getSolrVolumeMounts will return the solr volume mounts
func (c *BdReplicationController) getSolrVolumeMounts() []*horizonapi.VolumeMountConfig {
	volumesMounts := []*horizonapi.VolumeMountConfig{
		{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data"},
	}
	return volumesMounts
}
