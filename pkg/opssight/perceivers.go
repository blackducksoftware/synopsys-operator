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

package opssight

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// PodPerceiverReplicationController creates a replication controller for the pod perceiver
func (p *SpecConfig) PodPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.opssight.Spec.Perceiver.PodPerceiver.Name
	image := p.opssight.Spec.Perceiver.PodPerceiver.Image

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ServiceAccount))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create pod perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// ImagePerceiverReplicationController creates a replication controller for the image perceiver
func (p *SpecConfig) ImagePerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.opssight.Spec.Perceiver.ImagePerceiver.Name
	image := p.opssight.Spec.Perceiver.ImagePerceiver.Image

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ServiceAccount))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create image perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceiverReplicationController(name string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      utils.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	rc.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	return rc
}

func (p *SpecConfig) perceiverPod(name string, image string, account string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           utils.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	container, err := p.perceiverContainer(name, image)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := p.perceiverVolumes(name)

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

func (p *SpecConfig) perceiverContainer(name string, image string) (*components.Container, error) {
	cmd := fmt.Sprintf("./%s", name)
	if strings.Contains(name, "processor") {
		cmd = fmt.Sprintf("./opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   image,
		Command: []string{cmd},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.Spec.ConfigMapName)},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(p.opssight.Spec.Perceiver.Port),
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

func (p *SpecConfig) perceiverVolumes(name string) ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(name)}

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

func (p *SpecConfig) perceiverService(name string) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      utils.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceiver.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceiver.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})

	service.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})

	return service
}

// PodPerceiverService creates a service for the pod perceiver
func (p *SpecConfig) PodPerceiverService() *components.Service {
	return p.perceiverService(p.opssight.Spec.Perceiver.PodPerceiver.Name)
}

// ImagePerceiverService creates a service for the image perceiver
func (p *SpecConfig) ImagePerceiverService() *components.Service {
	return p.perceiverService(p.opssight.Spec.Perceiver.ImagePerceiver.Name)
}

func (p *SpecConfig) perceiverServiceAccount(name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      utils.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	return serviceAccount
}

// PodPerceiverServiceAccount creates a service account for the pod perceiver
func (p *SpecConfig) PodPerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.opssight.Spec.Perceiver.ServiceAccount)
}

// ImagePerceiverServiceAccount creates a service account for the image perceiver
func (p *SpecConfig) ImagePerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.opssight.Spec.Perceiver.ServiceAccount)
}

// PodPerceiverClusterRole creates a cluster role for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.PodPerceiver.Name),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{""},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"component": p.opssight.Spec.Perceiver.PodPerceiver.Name, "app": "opssight", "name": p.opssight.Name})

	return clusterRole
}

// ImagePerceiverClusterRole creates a cluster role for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ImagePerceiver.Name),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"image.openshift.io"},
		Resources: []string{"images"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"component": p.opssight.Spec.Perceiver.ImagePerceiver.Name, "app": "opssight", "name": p.opssight.Name})

	return clusterRole
}

// PodPerceiverClusterRoleBinding creates a cluster role binding for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.PodPerceiver.Name),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ServiceAccount),
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"component": p.opssight.Spec.Perceiver.PodPerceiver.Name, "app": "opssight", "name": p.opssight.Name})

	return clusterRoleBinding
}

// ImagePerceiverClusterRoleBinding creates a cluster role binding for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ImagePerceiver.Name),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      utils.GetResourceName(p.opssight.Name, util.OpsSightName, p.opssight.Spec.Perceiver.ServiceAccount),
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"component": p.opssight.Spec.Perceiver.ImagePerceiver.Name, "app": "opssight", "name": p.opssight.Name})

	return clusterRoleBinding
}
