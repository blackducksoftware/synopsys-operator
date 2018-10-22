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

// PerceptorSkyfireReplicationController creates a replication controller for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.Skyfire.Name,
		Namespace: p.config.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": p.config.Skyfire.Name})
	pod, err := p.perceptorSkyfirePod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create skyfire volumes")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.config.Skyfire.Name,
		ServiceAccount: p.config.Skyfire.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"name": p.config.Skyfire.Name})

	cont, err := p.perceptorSkyfireContainer()
	if err != nil {
		return nil, err
	}
	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Annotate(err, "unable to add skyfire container")
	}

	vols, err := p.perceptorSkyfireVolumes()
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

func (p *SpecConfig) perceptorSkyfireContainer() (*components.Container, error) {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    p.config.Skyfire.Name,
		Image:   p.config.Skyfire.Image,
		Command: []string{fmt.Sprintf("./%s", p.config.Skyfire.Name)},
		Args:    []string{"/etc/skyfire/skyfire.yaml"},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.config.Skyfire.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
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

	err = container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.Hub.PasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (p *SpecConfig) perceptorSkyfireVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "skyfire",
		MapOrSecretName: "skyfire",
	}))

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

// PerceptorSkyfireService creates a service for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.Skyfire.Name,
		Namespace: p.config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.config.Skyfire.Port),
		TargetPort: fmt.Sprintf("%d", p.config.Skyfire.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": p.config.Skyfire.Name})

	return service
}

// PerceptorSkyfireConfigMap creates a config map for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})
	configMap.AddData(map[string]string{"skyfire.yaml": fmt.Sprint(`{"UseInClusterConfig": "`, "true", `","Port": "`, p.config.Skyfire.Port, `","HubHost": "`, "TODO -- remove", `","HubPort": "`, p.config.Hub.Port, `","HubUser": "`, p.config.Hub.User, `","HubUserPasswordEnvVar": "`, p.config.Hub.PasswordEnvVar, `","HubClientTimeoutSeconds": "`, p.config.ScannerPod.Scanner.ClientTimeoutSeconds, `","PerceptorHost": "`, p.config.Perceptor.Name, `","PerceptorPort": "`, p.config.Perceptor.Port, `","KubeDumpIntervalSeconds": "`, "15", `","PerceptorDumpIntervalSeconds": "`, "15", `","HubDumpPauseSeconds": "`, "30", `","ImageFacadePort": "`, p.config.ScannerPod.ImageFacade.Port, `","LogLevel": "`, p.config.LogLevel, `"}`)})

	return configMap
}

// PerceptorSkyfireServiceAccount creates a service account for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// PerceptorSkyfireClusterRole creates a cluster role for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "skyfire",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"pods", "nodes"},
		Verbs:     []string{"get", "watch", "list"},
	})

	return clusterRole
}

// PerceptorSkyfireClusterRoleBinding creates a cluster role binding for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "skyfire",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}
