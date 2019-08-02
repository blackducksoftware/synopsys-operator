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
	store.Register(types.OpsSightScannerRCV1, NewOpsSightReplicationController)
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
	replicas := int32(o.opsSight.Spec.ScannerPod.ReplicaCount)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ScannerPod.Name),
		Namespace: o.opsSight.Spec.Namespace,
	})

	rc.AddSelectors(map[string]string{"component": o.opsSight.Spec.ScannerPod.Name, "app": "opssight", "name": o.opsSight.Name})
	rc.AddLabels(map[string]string{"component": o.opsSight.Spec.ScannerPod.Name, "app": "opssight", "name": o.opsSight.Name})
	pod, err := o.scannerPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (o *OpsSightReplicationController) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ScannerPod.Name),
		ServiceAccount: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ScannerPod.ImageFacade.ServiceAccount),
	})
	pod.AddLabels(map[string]string{"component": o.opsSight.Spec.ScannerPod.Name, "app": "opssight", "name": o.opsSight.Name})

	scannerContainer, err := o.scannerContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(scannerContainer)

	imageGetterContainer, err := o.imageGetterContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(imageGetterContainer)

	scannerVols, err := o.scannerVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating scanner volumes")
	}

	imageGetterVols, err := o.imageGetterVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating image getter volumes")
	}
	for _, v := range append(scannerVols, imageGetterVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (o *OpsSightReplicationController) scannerContainer() (*components.Container, error) {
	priv := false
	name := o.opsSight.Spec.ScannerPod.Scanner.Name
	command := name
	if name == "scanner" {
		command = fmt.Sprintf("opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      o.opsSight.Spec.ScannerPod.Scanner.Image,
		Command:    []string{fmt.Sprintf("./%s", command)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, o.opsSight.Spec.ConfigMapName)},
		MinCPU:     o.opsSight.Spec.ScannerCPU,
		MinMem:     o.opsSight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(o.opsSight.Spec.ScannerPod.Scanner.Port),
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
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.SecretName)})

	return container, nil
}

func (o *OpsSightReplicationController) imageGetterContainer() (*components.Container, error) {
	priv := false
	if !strings.EqualFold(o.opsSight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		priv = true
	}

	name := o.opsSight.Spec.ScannerPod.ImageFacade.Name
	command := name
	if name == "image-getter" {
		command = fmt.Sprintf("opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      o.opsSight.Spec.ScannerPod.ImageFacade.Image,
		Command:    []string{fmt.Sprintf("./%s", command)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, o.opsSight.Spec.ConfigMapName)},
		MinCPU:     o.opsSight.Spec.ScannerCPU,
		MinMem:     o.opsSight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(o.opsSight.Spec.ScannerPod.ImageFacade.Port),
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
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !strings.EqualFold(o.opsSight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name:      "dir-docker-socket",
			MountPath: "/var/run/docker.sock",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.SecretName)})

	return container, nil
}

func (o *OpsSightReplicationController) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{o.configMapVolume(o.opsSight.Spec.ScannerPod.Scanner.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

func (o *OpsSightReplicationController) imageGetterVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{o.configMapVolume(o.opsSight.Spec.ScannerPod.ImageFacade.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	if !strings.EqualFold(o.opsSight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
			VolumeName: "dir-docker-socket",
			Path:       "/var/run/docker.sock",
		}))
	}

	return vols, nil
}

func (o *OpsSightReplicationController) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.ConfigMapName),
		DefaultMode:     util.IntToInt32(420),
	})
}
