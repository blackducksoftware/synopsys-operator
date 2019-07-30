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
	store.Register(types.BlackDuckScanRCV1, NewBdReplicationController)
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
	containerConfig, ok := c.Containers[types.ScanContainerName]
	if !ok {
		return nil, fmt.Errorf("couldn't find container %s", types.ScanContainerName)
	}

	// hubMaxMemory is the amount of memory allocated to the JVM. We keep 512mb for alpine
	hubMaxMemory := 2048
	if containerConfig.MaxMem != nil && *containerConfig.MaxMem > 512 {
		hubMaxMemory = int(*containerConfig.MaxMem - 512)
	}

	scannerEnvs := []*horizonapi.EnvConfig{utils.GetBlackDuckConfigEnv(c.blackDuck.Name), utils.GetBlackDuckDBConfigEnv(c.blackDuck.Name)}
	scannerEnvs = append(scannerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: fmt.Sprintf("%dm", hubMaxMemory)})
	hubScanEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "scan", Image: containerConfig.Image, PullPolicy: horizonapi.PullAlways},
		EnvConfigs:      scannerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_ADMIN_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
			},
			{
				Name:      "db-passwords",
				MountPath: "/tmp/secrets/HUB_POSTGRES_USER_PASSWORD_FILE",
				SubPath:   "HUB_POSTGRES_USER_PASSWORD_FILE",
			},
			{
				Name:      "dir-scan",
				MountPath: "/opt/blackduck/hub/hub-scan/security",
			},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: int32(8443), Protocol: horizonapi.ProtocolTCP}},
	}

	apputils.SetLimits(hubScanContainerConfig.ContainerConfig, containerConfig)

	if c.blackDuck.Spec.LivenessProbes {
		hubScanContainerConfig.LivenessProbeConfigs = []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/usr/local/bin/docker-healthcheck.sh",
					"https://127.0.0.1:8443/api/health-checks/liveness",
					"/opt/blackduck/hub/hub-scan/security/root.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.crt",
					"/opt/blackduck/hub/hub-scan/security/blackduck_system.key",
				},
			},
			Delay:           240,
			Interval:        30,
			Timeout:         10,
			MinCountFailure: 10,
		}}
	}

	hubScanVolumes := []*components.Volume{hubScanEmptyDir, utils.GetDBSecretVolume(c.blackDuck.Name)}

	// Mount the HTTPS proxy certificate if provided
	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		hubScanContainerConfig.VolumeMounts = append(hubScanContainerConfig.VolumeMounts, &horizonapi.VolumeMountConfig{
			Name:      "proxy-certificate",
			MountPath: "/tmp/secrets/HUB_PROXY_CERT_FILE",
			SubPath:   "HUB_PROXY_CERT_FILE",
		})
		hubScanVolumes = append(hubScanVolumes, utils.GetProxyVolume(c.blackDuck.Name))
	}

	podConfig := &util.PodConfig{
		Volumes:             hubScanVolumes,
		Containers:          []*util.Container{hubScanContainerConfig},
		ImagePullSecrets:    c.blackDuck.Spec.RegistryConfiguration.PullSecrets,
		Labels:              apputils.GetVersionLabel("scan", c.blackDuck.Name, c.blackDuck.Spec.Version),
		NodeAffinityConfigs: utils.GetNodeAffinityConfigs("scan", &c.blackDuck.Spec),
	}
	if !c.config.IsOpenshift {
		podConfig.FSGID = util.IntToInt64(0)
	}

	return util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{Namespace: c.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "scan"), Replicas: util.IntToInt32(c.Replicas)},
		podConfig, apputils.GetLabel("scan", c.blackDuck.Name))
}
