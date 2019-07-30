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
	store.Register(types.BlackDuckBinaryScannerRCV1, NewBdReplicationController)
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
	containerConfig, ok := c.Containers[types.BinaryScannerContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.BinaryScannerContainerName)
	}

	binaryScannerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "binaryscanner", Image: containerConfig.Image,
			PullPolicy: horizonapi.PullAlways, Command: []string{"/docker-entrypoint.sh"}},
		EnvConfigs: []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name)},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(3001), Protocol: horizonapi.ProtocolTCP}},
	}
	apputils.SetLimits(binaryScannerContainerConfig.ContainerConfig, containerConfig)

	podConfig := &util.PodConfig{
		Containers:          []*util.Container{binaryScannerContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("binaryscanner", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("binaryscanner", &c.blackDuck.Spec),
	}

	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "binaryscanner"), Replicas: util.IntToInt32(1)},
		podConfig, apputils.GetLabel("binaryscanner", c.blackDuck.Name))
}
