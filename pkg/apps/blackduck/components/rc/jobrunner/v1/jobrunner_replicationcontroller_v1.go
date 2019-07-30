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
	store.Register(types.BlackDuckJobRunnerRCV1, NewBdReplicationController)
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
	containerConfig, ok := c.Containers[types.JobrunnerContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.JobrunnerContainerName)
	}

	// hubMaxMemory is the amount of memory allocated to the JVM. We keep 512mb for alpine
	hubMaxMemory := 4096
	if containerConfig.MaxMem != nil && *containerConfig.MaxMem > 512 {
		hubMaxMemory = int(*containerConfig.MaxMem - 512)
	}

	jobRunnerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-jobrunner")
	jobRunnerEnvs := []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name), utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name)}
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dm", hubMaxMemory)})
	jobRunnerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "jobrunner", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE", SubPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE"},
			{Name: "db-passwords", MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE", SubPath: "HUB_POSTGRES_USER_PASSWORD_FILE"},
			{Name: "dir-jobrunner", MountPath: "/opt/blackduck/hub/jobrunner/security"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(3001), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(jobRunnerContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
		jobRunnerContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type:    horizonapi.ActionTypeCommand,
				Command: []string{"/usr/local/bin/docker-healthcheck.sh"},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	jobRunnerVolumes := []*components.Volume{utils.GetDBSecretVolume(c.blackDuck.Name), jobRunnerEmptyDir}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		jobRunnerContainerConfig.VolumeMounts = append(jobRunnerContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		jobRunnerVolumes = append(jobRunnerVolumes, utils.GetProxyVolume(c.blackDuck.Name))
	}

	podConfig := &util.PodConfig{
		Volumes:             jobRunnerVolumes,
		Containers:          []*util.Container{jobRunnerContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("jobrunner", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("jobrunner", &c.blackDuck.Spec),
	}
	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}
	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "jobrunner"), Replicas: util.IntToInt32(c.Replicas)},
		podConfig, apputils.GetLabel("jobrunner", c.blackDuck.Name))
}
