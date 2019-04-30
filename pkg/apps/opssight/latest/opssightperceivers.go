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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// GetPodPerceiverReplicationController creates a replication controller for the OpsSight's Pod Perceiver
func (p *SpecConfig) GetPodPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.opssight.Spec.Perceiver.PodPerceiver.Name
	image := p.opssight.Spec.Perceiver.PodPerceiver.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "opssight-pod-processor")
	}

	rc := p.createPerceiverReplicationController(name, 1)

	pod, err := p.createPerceiverPod(name, image, p.opssight.Spec.Perceiver.ServiceAccount)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create pod perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// GetImagePerceiverReplicationController creates a replication controller for OpsSight's Image Perceiver
func (p *SpecConfig) GetImagePerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.opssight.Spec.Perceiver.ImagePerceiver.Name
	image := p.opssight.Spec.Perceiver.ImagePerceiver.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "opssight-image-processor")
	}

	rc := p.createPerceiverReplicationController(name, 1)

	pod, err := p.createPerceiverPod(name, image, p.opssight.Spec.Perceiver.ServiceAccount)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create image perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// createPerceiverReplicationController returns a Replication Controller for a Perceiver
func (p *SpecConfig) createPerceiverReplicationController(name string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      name,
		Namespace: p.opssight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"name": name, "app": "opssight"})
	rc.AddLabels(map[string]string{"name": name, "app": "opssight"})
	return rc
}

// createPerceiverPod returns a Pod for a Perceiver
func (p *SpecConfig) createPerceiverPod(name string, image string, account string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           name,
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"name": name, "app": "opssight"})
	container, err := p.createPerceiverContainer(name, image)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := p.createPerceiverVolumes(name)

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

// createPerceiverContainer returns a Container for a Perciever
func (p *SpecConfig) createPerceiverContainer(name string, image string) (*components.Container, error) {
	cmd := fmt.Sprintf("./%s", name)
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

// createPerceiverVolumes returns a Volume for a Perceiver
func (p *SpecConfig) createPerceiverVolumes(name string) ([]*components.Volume, error) {
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

// GetPodPerceiverService creates a service for OpsSight's Pod Perceiver
func (p *SpecConfig) GetPodPerceiverService() *components.Service {
	return p.createPerceiverService(p.opssight.Spec.Perceiver.PodPerceiver.Name)
}

// GetImagePerceiverService creates a service for OpsSight's Image Perceiver
func (p *SpecConfig) GetImagePerceiverService() *components.Service {
	return p.createPerceiverService(p.opssight.Spec.Perceiver.ImagePerceiver.Name)
}

// createPerceiverService returns a Service for a Perceiver
func (p *SpecConfig) createPerceiverService(name string) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceiver.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceiver.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})

	service.AddLabels(map[string]string{"name": name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": name, "app": "opssight"})

	return service
}

// GetPodPerceiverServiceAccount creates a service account for OpsSight's Pod Perceiver
func (p *SpecConfig) GetPodPerceiverServiceAccount() *components.ServiceAccount {
	return p.createPerceiverServiceAccount(p.opssight.Spec.Perceiver.ServiceAccount)
}

// GetImagePerceiverServiceAccount creates a service account for OopsSight's Image Perceiver
func (p *SpecConfig) GetImagePerceiverServiceAccount() *components.ServiceAccount {
	return p.createPerceiverServiceAccount(p.opssight.Spec.Perceiver.ServiceAccount)
}

// createPerceiverServiceAccount returns a Service Account for a Perceiver
func (p *SpecConfig) createPerceiverServiceAccount(name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      name,
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"name": name, "app": "opssight"})
	return serviceAccount
}

// GetPodPerceiverClusterRole creates a cluster role for OpsSight's Pod Perceiver
func (p *SpecConfig) GetPodPerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       p.opssight.Spec.Perceiver.PodPerceiver.Name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{""},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"name": p.opssight.Spec.Perceiver.PodPerceiver.Name, "app": "opssight"})

	return clusterRole
}

// GetImagePerceiverClusterRole creates a cluster role for OpsSight's Image Perceiver
func (p *SpecConfig) GetImagePerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       p.opssight.Spec.Perceiver.ImagePerceiver.Name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"image.openshift.io"},
		Resources: []string{"images"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"name": p.opssight.Spec.Perceiver.ImagePerceiver.Name, "app": "opssight"})

	return clusterRole
}

// GetPodPerceiverClusterRoleBinding creates a cluster role binding for OpsSight's Pod Perceiver
func (p *SpecConfig) GetPodPerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.opssight.Spec.Perceiver.PodPerceiver.Name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.opssight.Spec.Perceiver.ServiceAccount,
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"name": p.opssight.Spec.Perceiver.PodPerceiver.Name, "app": "opssight"})

	return clusterRoleBinding
}

// GetImagePerceiverClusterRoleBinding creates a cluster role binding for OpsSight's Image Perceiver
func (p *SpecConfig) GetImagePerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.opssight.Spec.Perceiver.ImagePerceiver.Name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.opssight.Spec.Perceiver.ServiceAccount,
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"name": p.opssight.Spec.Perceiver.ImagePerceiver.Name, "app": "opssight"})

	return clusterRoleBinding
}
