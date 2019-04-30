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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
)

// GetPrometheusDeployment returns a Deployment for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "prometheus",
		Namespace: p.opssight.Spec.Namespace,
	})
	deployment.AddLabels(map[string]string{"name": "prometheus", "app": "opssight"})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "opssight"})

	pod, err := p.getPrometheusPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create metrics pod")
	}
	deployment.AddPod(pod)

	return deployment, nil
}

// getPrometheusPod returns a Pod for OpsSight's Prometheus
func (p *SpecConfig) getPrometheusPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "prometheus",
	})
	pod.AddLabels(map[string]string{"name": "prometheus", "app": "opssight"})
	container, err := p.getPrometheusContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := p.getPrometheusVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating metrics volumes")
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

// getPrometheusContainer returns a Container for OpsSight's Prometheus
func (p *SpecConfig) getPrometheusContainer() (*components.Container, error) {
	image := p.opssight.Spec.Prometheus.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "prometheus")
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:  p.opssight.Spec.Prometheus.Name,
		Image: image,
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/", "--storage.tsdb.retention=120d"},
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(p.opssight.Spec.Prometheus.Port),
		Protocol:      horizonapi.ProtocolTCP,
		Name:          "web",
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "data",
		MountPath: "/data",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "prometheus",
		MountPath: "/etc/prometheus",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

// getPrometheusVolumes returns a list of Volumes for OpsSight's Prometheus
func (p *SpecConfig) getPrometheusVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}
	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "prometheus",
		MapOrSecretName: "prometheus",
		DefaultMode:     util.IntToInt32(420),
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "data",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

// GetPrometheusService returns a service for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "prometheus",
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"name": "prometheus", "app": "opssight"})
	service.AddSelectors(map[string]string{"name": "prometheus", "app": "opssight"})

	return service, err
}

// GetPrometheusExposeService returns the correct service type for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusExposeService() (*components.Service, error) {
	var svc *components.Service
	var err error
	switch strings.ToUpper(p.opssight.Spec.Prometheus.Expose) {
	case "NODEPORT":
		svc, err = p.GetPrometheusNodePortService()
		break
	case "LOADBALANCER":
		svc, err = p.GetPrometheusLoadBalancerService()
		break
	default:
	}
	return svc, err
}

// GetPrometheusNodePortService returns a Nodeport Service for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusNodePortService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "prometheus-exposed",
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"name": "prometheus", "app": "opssight"})
	service.AddSelectors(map[string]string{"name": "prometheus", "app": "opssight"})

	return service, err
}

// GetPrometheusLoadBalancerService returns a Loadbalancer service for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusLoadBalancerService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "prometheus-exposed",
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"name": "prometheus", "app": "opssight"})
	service.AddSelectors(map[string]string{"name": "prometheus", "app": "opssight"})

	return service, err
}

// GetPrometheusConfigMap returns a config map for OpsSight's Prometheus
func (p *SpecConfig) GetPrometheusConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "prometheus",
		Namespace: p.opssight.Spec.Namespace,
	})

	/* EXAMPLE:
	{
	  "global": {
	    "scrape_interval": "5s"
	  },
	  "scrape_configs": [
	    {
	      "job_name": "perceptor-scrape",
	      "scrape_interval": "5s",
	      "static_configs": [
	        {
	          "targets": [
	            "perceptor:3001",
	            "perceptor-scanner:3003",
	            "perceptor-imagefacade:3004",
	            "pod-perceiver:3002"
	          ]
	        }
	      ]
	    }
	  ]
	}
	*/
	targets := []string{
		fmt.Sprintf("%s:%d", p.opssight.Spec.Perceptor.Name, p.opssight.Spec.Perceptor.Port),
		fmt.Sprintf("%s:%d", p.opssight.Spec.ScannerPod.Scanner.Name, p.opssight.Spec.ScannerPod.Scanner.Port),
		fmt.Sprintf("%s:%d", p.opssight.Spec.ScannerPod.ImageFacade.Name, p.opssight.Spec.ScannerPod.ImageFacade.Port),
	}
	if p.opssight.Spec.Perceiver.EnableImagePerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", p.opssight.Spec.Perceiver.ImagePerceiver.Name, p.opssight.Spec.Perceiver.Port))
	}
	if p.opssight.Spec.Perceiver.EnablePodPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", p.opssight.Spec.Perceiver.PodPerceiver.Name, p.opssight.Spec.Perceiver.Port))
	}
	if p.opssight.Spec.EnableSkyfire {
		targets = append(targets, fmt.Sprintf("%s:%d", p.opssight.Spec.Skyfire.Name, p.opssight.Spec.Skyfire.PrometheusPort))
	}
	data := map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "5s",
		},
		"scrape_configs": []interface{}{
			map[string]interface{}{
				"job_name":        "perceptor-scrape",
				"scrape_interval": "5s",
				"static_configs": []interface{}{
					map[string]interface{}{
						"targets": targets,
					},
				},
			},
		},
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddLabels(map[string]string{"app": "opssight"})
	configMap.AddData(map[string]string{"prometheus.yml": string(bytes)})

	return configMap, nil
}

// GetPrometheusOpenShiftRoute creates the OpenShift route component for the prometheus metrics
func (p *SpecConfig) GetPrometheusOpenShiftRoute() *api.Route {
	namespace := p.opssight.Spec.Namespace
	if strings.ToUpper(p.opssight.Spec.Perceptor.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               fmt.Sprintf("%s-%s", p.opssight.Spec.Prometheus.Name, namespace),
			Namespace:          namespace,
			Kind:               "Service",
			ServiceName:        p.opssight.Spec.Prometheus.Name,
			PortName:           fmt.Sprintf("port-%s", p.opssight.Spec.Prometheus.Name),
			Labels:             map[string]string{"app": "opssight"},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
