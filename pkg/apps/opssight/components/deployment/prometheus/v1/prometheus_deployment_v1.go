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
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
)

// OpsSightDeployment holds the OpsSight deployment configuration
type OpsSightDeployment struct {
	*types.PodResource
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.OpsSightMetricsDeploymentV1, NewOpsSightDeployment)
}

// NewOpsSightDeployment returns the OpsSight deployment configuration
func NewOpsSightDeployment(podResource *types.PodResource, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.DeploymentInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightDeployment{PodResource: podResource, config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetDeployment returns the deployment
func (o *OpsSightDeployment) GetDeployment() (*components.Deployment, error) {
	if !o.opsSight.Spec.EnableMetrics {
		return nil, nil
	}

	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus"),
		Namespace: o.opsSight.Spec.Namespace,
	})
	deployment.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": o.opsSight.Name})
	deployment.AddMatchLabelsSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": o.opsSight.Name})

	pod, err := o.opsSightMetricsPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create metrics pod")
	}
	deployment.AddPod(pod)

	return deployment, nil
}

func (o *OpsSightDeployment) opsSightMetricsPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus"),
	})
	pod.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": o.opsSight.Name})
	container, err := o.opsSightMetricsContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := o.opsSightMetricsVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating metrics volumes")
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (o *OpsSightDeployment) opsSightMetricsContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:  o.opsSight.Spec.Prometheus.Name,
		Image: o.opsSight.Spec.Prometheus.Image,
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/", "--storage.tsdb.retention=120d"},
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(o.opsSight.Spec.Prometheus.Port),
		Protocol:      horizonapi.ProtocolTCP,
		Name:          "web",
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "data",
		MountPath: "/data",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "prometheus",
		MountPath: "/etc/prometheus",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func (o *OpsSightDeployment) opsSightMetricsVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}
	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "prometheus",
		MapOrSecretName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus"),
		DefaultMode:     util.IntToInt32(420),
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "data",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}
