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
)

// PerceptorReplicationController creates a replication controller for perceptor
func (p *OpsSightConfig) PerceptorReplicationController() *components.ReplicationController {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.PerceptorImageName,
		Namespace: p.config.Namespace,
	})
	rc.AddPod(p.perceptorPod())
	rc.AddLabelSelectors(map[string]string{"name": p.config.PerceptorImageName})

	return rc
}

func (p *OpsSightConfig) perceptorPod() *components.Pod {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: p.config.PerceptorImageName,
	})
	pod.AddLabels(map[string]string{"name": p.config.PerceptorImageName})
	pod.AddContainer(p.perceptorContainer())
	pod.AddVolume(p.perceptorVolume())

	return pod
}

func (p *OpsSightConfig) perceptorContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    p.config.PerceptorImageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.PerceptorImageName, p.config.PerceptorImageVersion),
		Command: []string{fmt.Sprintf("./%s", p.config.PerceptorImageName)},
		Args:    []string{"/etc/perceptor/perceptor.yaml"},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.PerceptorPort),
		Protocol:      horizonapi.ProtocolTCP,
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceptor",
		MountPath: "/etc/perceptor",
	})
	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})

	return container
}

func (p *OpsSightConfig) perceptorVolume() *components.Volume {
	volume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceptor",
		MapOrSecretName: "perceptor",
	})

	return volume
}

// PerceptorService creates a service for perceptor
func (p *OpsSightConfig) PerceptorService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.PerceptorImageName,
		Namespace: p.config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.PerceptorPort),
		TargetPort: fmt.Sprintf("%d", *p.config.PerceptorPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": p.config.PerceptorImageName})

	return service
}

// PerceptorConfigMap creates a config map for perceptor
func (p *OpsSightConfig) PerceptorConfigMap() *components.ConfigMap {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceptor",
		Namespace: p.config.Namespace,
	})
	cm.AddData(map[string]string{"perceptor.yaml": fmt.Sprint(`{"HubHost": "`, p.config.HubHost, `","HubPort": "`, *p.config.HubPort, `","HubUser": "`, p.config.HubUser, `","HubUserPasswordEnvVar": "`, p.config.HubUserPasswordEnvVar, `","HubClientTimeoutMilliseconds": "`, *p.config.HubClientTimeoutPerceptorMilliseconds, `","ConcurrentScanLimit": "`, *p.config.ConcurrentScanLimit, `","Port": "`, *p.config.PerceptorPort, `","LogLevel": "`, p.config.LogLevel, `"}`)})

	return cm
}

// PerceptorSecret create a secret for perceptor
func (p *OpsSightConfig) PerceptorSecret() *components.Secret {
	secretConfig := horizonapi.SecretConfig{
		Name:      p.config.SecretName,
		Namespace: p.config.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)
	secret.AddData(map[string][]byte{"HubUserPassword": []byte(p.config.HubUserPassword)})

	return secret
}
