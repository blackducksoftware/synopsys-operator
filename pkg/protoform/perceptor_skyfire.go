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

package protoform

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// PerceptorSkyfireReplicationController creates a replication controller for perceptor skyfire
func (i *Installer) PerceptorSkyfireReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      i.Config.SkyfireImageName,
		Namespace: i.Config.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": i.Config.SkyfireImageName})
	pod, err := i.perceptorSkyfirePod()
	if err != nil {
		return nil, fmt.Errorf("failed to create skyfire volumes: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (i *Installer) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           i.Config.SkyfireImageName,
		ServiceAccount: i.Config.ServiceAccounts["image-perceiver"],
	})
	pod.AddLabels(map[string]string{"name": i.Config.SkyfireImageName})

	pod.AddContainer(i.perceptorSkyfireContainer())

	vols, err := i.perceptorSkyfireVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating skyfire volumes: %v", err)
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (i *Installer) perceptorSkyfireContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    i.Config.SkyfireImageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", i.Config.Registry, i.Config.ImagePath, i.Config.SkyfireImageName, i.Config.SkyfireImageVersion),
		Command: []string{"./skyfire"},
		Args:    []string{"/etc/skyfire/skyfire.yaml"},
		MinCPU:  i.Config.DefaultCPU,
		MinMem:  i.Config.DefaultMem,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", i.Config.SkyfirePort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "skyfire",
		MountPath: "/etc/skyfire",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})

	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: i.Config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     i.Config.ViperSecret,
	})

	return container
}

func (i *Installer) perceptorSkyfireVolumes() ([]*components.Volume, error) {
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
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

// PerceptorSkyfireService creates a service for perceptor skyfire
func (i *Installer) PerceptorSkyfireService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      i.Config.SkyfireImageName,
		Namespace: i.Config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(i.Config.SkyfirePort),
		TargetPort: fmt.Sprintf("%d", i.Config.SkyfirePort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": i.Config.SkyfireImageName})

	return service
}

// PerceptorSkyfireConfigMap creates a config map for perceptor skyfire
func (i *Installer) PerceptorSkyfireConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "skyfire",
		Namespace: i.Config.Namespace,
	})
	configMap.AddData(map[string]string{"skyfire.yaml": fmt.Sprint(`{"UseInClusterConfig": "`, "true", `","Port": "`, "3005", `","HubHost": "`, i.Config.HubHost, `","HubPort": "`, i.Config.HubPort, `","HubUser": "`, i.Config.HubUser, `","HubUserPasswordEnvVar": "`, i.Config.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, i.Config.HubClientTimeoutScannerSeconds, `","PerceptorHost": "`, i.Config.PerceptorImageName, `","PerceptorPort": "`, i.Config.PerceptorPort, `","KubeDumpIntervalSeconds": "`, "15", `","PerceptorDumpIntervalSeconds": "`, "15", `","HubDumpPauseSeconds": "`, "30", `","ImageFacadePort": "`, i.Config.ImageFacadePort, `","LogLevel": "`, i.Config.LogLevel, `"}`)})

	return configMap
}

// PerceptorSkyfireServiceAccount creates a service account for perceptor skyfire
func (i *Installer) PerceptorSkyfireServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "skyfire",
		Namespace: i.Config.Namespace,
	})

	return serviceAccount
}

// PerceptorSkyfireClusterRole creates a cluster role for perceptor skyfire
func (i *Installer) PerceptorSkyfireClusterRole() *components.ClusterRole {
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
func (i *Installer) PerceptorSkyfireClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "skyfire",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "skyfire",
		Namespace: i.Config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}
