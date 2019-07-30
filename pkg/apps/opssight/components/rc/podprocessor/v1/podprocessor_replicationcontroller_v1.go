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
	"strings"

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
	store.Register(types.OpsSightPodProcessorRCV1, NewOpsSightReplicationController)
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
	if !o.opsSight.Spec.Perceiver.EnablePodPerceiver {
		return nil, nil
	}
	name := o.opsSight.Spec.Perceiver.PodPerceiver.Name
	image := o.opsSight.Spec.Perceiver.PodPerceiver.Image

	rc := o.processorReplicationController(name, 1)

	pod, err := o.processorPod(name, image, utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceiver.ServiceAccount))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create image processor pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (o *OpsSightReplicationController) processorReplicationController(name string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, name),
		Namespace: o.opsSight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": o.opsSight.Name})
	rc.AddLabels(map[string]string{"component": name, "app": "opssight", "name": o.opsSight.Name})
	return rc
}

func (o *OpsSightReplicationController) processorPod(name string, image string, account string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           utils.GetResourceName(o.opsSight.Name, util.OpsSightName, name),
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"component": name, "app": "opssight", "name": o.opsSight.Name})
	container, err := o.processorContainer(name, image)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := o.processorVolumes(name)

	if err != nil {
		return nil, errors.Annotate(err, "unable to create volumes")
	}

	for _, v := range vols {
		err = pod.AddVolume(v)
		if err != nil {
			return nil, errors.Annotate(err, "unable to add volume to pod")
		}
	}

	return pod, nil
}

func (o *OpsSightReplicationController) processorContainer(name string, image string) (*components.Container, error) {
	cmd := fmt.Sprintf("./%s", name)
	if strings.Contains(name, "processor") {
		cmd = fmt.Sprintf("./opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   image,
		Command: []string{cmd},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, o.opsSight.Spec.ConfigMapName)},
		MinCPU:  o.opsSight.Spec.DefaultCPU,
		MinMem:  o.opsSight.Spec.DefaultMem,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(o.opsSight.Spec.Perceiver.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})
	if err != nil {
		return nil, errors.Annotatef(err, "unable to add the volume mount to %s container", name)
	}

	return container, nil
}

func (o *OpsSightReplicationController) processorVolumes(name string) ([]*components.Volume, error) {
	vols := []*components.Volume{o.configMapVolume(name)}

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
