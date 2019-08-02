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

// OpsSightReplicationController holds the OpsSight RC configuration
type OpsSightReplicationController struct {
	*types.PodResource
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.SkyfireRCV1, NewOpsSightReplicationController)
}

// NewOpsSightReplicationController returns the OpsSight RC configuration
func NewOpsSightReplicationController(podResource *types.PodResource, config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ReplicationControllerInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightReplicationController{PodResource: podResource, config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetRc returns the RC
func (o *OpsSightReplicationController) GetRc() (*components.ReplicationController, error) {
	if !o.opsSight.Spec.EnableSkyfire {
		return nil, nil
	}

	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Skyfire.Name),
		Namespace: o.opsSight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"component": o.opsSight.Spec.Skyfire.Name, "app": "opssight", "name": o.opsSight.Name})
	rc.AddLabels(map[string]string{"component": o.opsSight.Spec.Skyfire.Name, "app": "opssight", "name": o.opsSight.Name})
	pod, err := o.perceptorSkyfirePod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create skyfire volumes")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (o *OpsSightReplicationController) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Skyfire.Name),
		ServiceAccount: o.opsSight.Spec.Skyfire.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"component": o.opsSight.Spec.Skyfire.Name, "app": "opssight", "name": o.opsSight.Name})

	cont, err := o.perceptorSkyfireContainer()
	if err != nil {
		return nil, err
	}
	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Annotate(err, "unable to add skyfire container")
	}

	vols, err := o.perceptorSkyfireVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating skyfire volumes")
	}
	for _, v := range vols {
		err = pod.AddVolume(v)
		if err != nil {
			return nil, errors.Annotate(err, "error add pod volume")
		}
	}

	return pod, nil
}

func (o *OpsSightReplicationController) pyfireContainer() (*components.Container, error) {
	return components.NewContainer(horizonapi.ContainerConfig{
		Name:    o.opsSight.Spec.Skyfire.Name,
		Image:   o.opsSight.Spec.Skyfire.Image,
		Command: []string{"python3"},
		Args: []string{
			"src/main.py",
			fmt.Sprintf("/etc/skyfire/%s.json", o.opsSight.Spec.ConfigMapName),
		},
		MinCPU: o.opsSight.Spec.DefaultCPU,
		MinMem: o.opsSight.Spec.DefaultMem,
	})
}

func (o *OpsSightReplicationController) golangSkyfireContainer() (*components.Container, error) {
	return components.NewContainer(horizonapi.ContainerConfig{
		Name:    o.opsSight.Spec.Skyfire.Name,
		Image:   o.opsSight.Spec.Skyfire.Image,
		Command: []string{fmt.Sprintf("./%s", o.opsSight.Spec.Skyfire.Name)},
		Args:    []string{fmt.Sprintf("/etc/skyfire/%s.json", o.opsSight.Spec.ConfigMapName)},
		MinCPU:  o.opsSight.Spec.DefaultCPU,
		MinMem:  o.opsSight.Spec.DefaultMem,
	})
}

func (o *OpsSightReplicationController) perceptorSkyfireContainer() (*components.Container, error) {
	container, err := o.pyfireContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(o.opsSight.Spec.Skyfire.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "skyfire",
		MountPath: "/etc/skyfire",
	})
	if err != nil {
		return nil, err
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})
	if err != nil {
		return nil, err
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.SecretName)})

	return container, nil
}

func (o *OpsSightReplicationController) perceptorSkyfireVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{o.configMapVolume("skyfire")}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "logs",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

func (o *OpsSightReplicationController) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ConfigMapName),
		DefaultMode:     util.IntToInt32(420),
	})
}
