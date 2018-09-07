/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package hubfederator

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// ReplicationController will create a replication controller for the hub federator
func (a *App) ReplicationController() *components.ReplicationController {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      "hub-federator",
		Namespace: a.config.Namespace,
	})

	rc.AddPod(a.pod())

	return rc
}

func (a *App) pod() *components.Pod {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           a.config.ImageName,
		ServiceAccount: "hub-federator",
	})
	pod.AddLabels(map[string]string{"name": a.config.ImageName})
	pod.AddContainer(a.container())
	pod.AddVolume(a.volume())

	return pod
}

func (a *App) container() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    a.config.ImageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", a.config.Registry, a.config.ImagePath, a.config.ImageName, a.config.ImageVersion),
		Command: []string{"./hub"},
		Args:    []string{"/etc/hub-federator/config.json"},
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *a.config.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "hub-federator",
		MountPath: "/etc/hub-federator",
	})
	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "REGISTRATION_KEY",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     a.config.RegistrationKey,
	})

	return container
}

func (a *App) volume() *components.Volume {
	mode := int32(420)
	volume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hub-federator",
		MapOrSecretName: "hub-federator",
		DefaultMode:     &mode,
	})

	return volume
}

// Services creates services for the hub federator
func (a *App) Services() []*components.Service {
	services := []*components.Service{}
	services = append(services, a.service("", horizonapi.ClusterIPServiceTypeDefault))
	services = append(services, a.service("-np", horizonapi.ClusterIPServiceTypeNodePort))
	services = append(services, a.service("-lb", horizonapi.ClusterIPServiceTypeLoadBalancer))

	return services
}

func (a *App) service(postfix string, serviceType horizonapi.ClusterIPServiceType) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          fmt.Sprintf("hub-federator%s", postfix),
		Namespace:     a.config.Namespace,
		IPServiceType: serviceType,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*a.config.Port),
		TargetPort: fmt.Sprintf("%d", *a.config.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       a.config.ImageName,
	})

	service.AddSelectors(map[string]string{"name": a.config.ImageName})

	return service
}

// ConfigMap creates a config map for the hub federator
func (a *App) ConfigMap() *components.ConfigMap {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      a.config.ImageName,
		Namespace: a.config.Namespace,
	})
	cm.AddData(map[string]string{"config.json": fmt.Sprint(`{"DryRun": "`, a.config.DryRun, `","LogLevel": "`, a.config.LogLevel, `","Namespace": "`, a.config.Namespace, `","Threadiness": "`, a.config.NumberOfThreads, `","HubFederatorConfig": {"HubConfig": {"User": "sysadmin","PasswordEnvVar": "HUB_PASSWORD","ClientTimeoutMilliseconds": 5000,"Port": 443,"FetchAllProjectsPauseSeconds": 60},"UseMockMode": false,"Port": 3016}}"`)})

	return cm
}

// ServiceAccount creates a service account for the hub federator
func (a *App) ServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "hub-federator",
		Namespace: a.config.Namespace,
	})

	return serviceAccount
}

// ClusterRoleBinding creates a cluster role binding for the hub federator
func (a *App) ClusterRoleBinding() *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "hub-federator",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "hub-federator",
		Namespace: a.config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	return clusterRoleBinding
}
