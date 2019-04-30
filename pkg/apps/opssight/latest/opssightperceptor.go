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
	"encoding/json"
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
)

// GetPerceptorReplicationController creates a replication controller for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.opssight.Spec.Perceptor.Name,
		Namespace: p.opssight.Spec.Namespace,
	})
	pod, err := p.getPerceptorPod()
	if err != nil {
		return nil, errors.Trace(err)
	}
	rc.AddPod(pod)
	rc.AddSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	rc.AddLabels(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	return rc, nil
}

// getPerceptorPod returns a Pod for OpsSight's Perceptor
func (p *SpecConfig) getPerceptorPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: p.opssight.Spec.Perceptor.Name,
	})
	pod.AddLabels(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	cont, err := p.getPerceptorContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = pod.AddVolume(p.configMapVolume(p.opssight.Spec.Perceptor.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return pod, nil
}

// getPerceptorContainer returns a Container for OpsSight's Perceptor
func (p *SpecConfig) getPerceptorContainer() (*components.Container, error) {
	name := p.opssight.Spec.Perceptor.Name
	image := p.opssight.Spec.Perceptor.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "opssight-core")
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   image,
		Command: []string{fmt.Sprintf("./%s", name)},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.Spec.ConfigMapName)},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(p.opssight.Spec.Perceptor.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.Spec.SecretName})

	return container, nil
}

// GetPerceptorService creates a service for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.Spec.Perceptor.Name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceptor.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceptor.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.opssight.Spec.Perceptor.Name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	service.AddLabels(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})

	return service, nil
}

// GetPerceptorExposeService returns the correct Service type for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorExposeService() (*components.Service, error) {
	var svc *components.Service
	var err error
	switch strings.ToUpper(p.opssight.Spec.Perceptor.Expose) {
	case "NODEPORT":
		svc, err = p.GetPerceptorNodePortService()
		break
	case "LOADBALANCER":
		svc, err = p.GetPerceptorLoadBalancerService()
		break
	default:
	}
	return svc, err
}

// GetPerceptorNodePortService creates a nodeport service for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorNodePortService() (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.opssight.Spec.Perceptor.Name)
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceptor.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceptor.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"name": name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})

	return service, nil
}

// GetPerceptorLoadBalancerService creates a loadbalancer service for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorLoadBalancerService() (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.opssight.Spec.Perceptor.Name)
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceptor.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceptor.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"name": name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})

	return service, nil
}

// GetPerceptorSecret create a secret for OpsSight's Perceptor
func (p *SpecConfig) GetPerceptorSecret() *components.Secret {
	secretConfig := horizonapi.SecretConfig{
		Name:      p.opssight.Spec.SecretName,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)

	// empty data fields that will be overwritten
	emptyHosts := make(map[string]*opssightapi.Host)
	bytes, err := json.Marshal(emptyHosts)
	if err != nil {
		log.Errorf("unable to marshal Black Duck passwords: %+v", err)
	}
	secret.AddData(map[string][]byte{p.opssight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	emptySecuredRegistries := make(map[string]*opssightapi.RegistryAuth)
	bytes, err = json.Marshal(emptySecuredRegistries)
	if err != nil {
		log.Errorf("unable to marshal secured registries: %+v", err)
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	secret.AddLabels(map[string]string{"name": p.opssight.Spec.SecretName, "app": "opssight"})
	return secret
}

// GetPerceptorOpenShiftRoute creates the OpenShift route component for the perceptor model
func (p *SpecConfig) GetPerceptorOpenShiftRoute() *api.Route {
	namespace := p.opssight.Spec.Namespace
	if strings.ToUpper(p.opssight.Spec.Perceptor.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               fmt.Sprintf("%s-%s", p.opssight.Spec.Perceptor.Name, namespace),
			Namespace:          namespace,
			Kind:               "Service",
			ServiceName:        p.opssight.Spec.Perceptor.Name,
			PortName:           fmt.Sprintf("port-%s", p.opssight.Spec.Perceptor.Name),
			Labels:             map[string]string{"app": "opssight"},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
