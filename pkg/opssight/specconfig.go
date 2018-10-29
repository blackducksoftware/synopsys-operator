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

package opssight

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config    *v1.OpsSightSpec
	configMap *ConfigMap
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *v1.OpsSightSpec) *SpecConfig {
	privateRegistries := []RegistryAuth{}
	for _, reg := range config.ScannerPod.ImageFacade.InternalRegistries {
		privateRegistries = append(privateRegistries, RegistryAuth{
			Password: reg.Password,
			URL:      reg.URL,
			User:     reg.User,
		})
	}
	configMap := &ConfigMap{
		LogLevel: config.LogLevel,
		Hub: HubConfig{
			Hosts:               []string{},
			PasswordEnvVar:      config.Hub.PasswordEnvVar,
			ConcurrentScanLimit: config.Hub.ConcurrentScanLimit,
			Port:                config.Hub.Port,
			TotalScanLimit:      config.Hub.TotalScanLimit,
			User:                config.Hub.User,
		},
		ImageFacade: ImageFacadeConfig{
			CreateImagesOnly:        false,
			Host:                    "localhost",
			Port:                    config.ScannerPod.ImageFacade.Port,
			PrivateDockerRegistries: privateRegistries,
		},
		Perceiver: PerceiverConfig{
			Image: ImagePerceiverConfig{},
			Pod: PodPerceiverConfig{
				NamespaceFilter: config.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: config.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       config.Perceiver.DumpIntervalMinutes,
			Port:                      config.Perceiver.Port,
		},
		Perceptor: PerceptorConfig{
			Timings: PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: config.Perceptor.CheckForStalledScansPauseHours,
				HubClientTimeoutMilliseconds:   config.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       config.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  config.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  config.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        config.Perceptor.Name,
			Port:        config.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: ScannerConfig{
			HubClientTimeoutSeconds: config.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:          config.ScannerPod.ImageDirectory,
			Port:                    config.ScannerPod.Scanner.Port,
		},
		Skyfire: SkyfireConfig{
			HubClientTimeoutSeconds:      config.Skyfire.HubClientTimeoutSeconds,
			HubDumpPauseSeconds:          config.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:      config.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds: config.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                         config.Skyfire.Port,
			UseInClusterConfig:           true,
		},
	}
	return &SpecConfig{config: config, configMap: configMap}
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add Perceptor
	rc, err := p.PerceptorReplicationController()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ReplicationControllers = append(components.ReplicationControllers, rc)
	service, err := p.PerceptorService()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.Services = append(components.Services, service)
	cm, err := p.configMap.horizonConfigMap(p.config.Perceptor.Name, p.config.Namespace, fmt.Sprintf("%s.yaml", p.config.Perceptor.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ConfigMaps = append(components.ConfigMaps, cm)
	components.Secrets = append(components.Secrets, p.PerceptorSecret())

	// Add Perceptor Scanner
	scannerRC, err := p.ScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.ScannerService(), p.ImageFacadeService())

	scannerCM, err := p.configMap.horizonConfigMap(
		p.config.ScannerPod.Scanner.Name,
		p.config.Namespace,
		fmt.Sprintf("%s.yaml", p.config.ScannerPod.Scanner.Name))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner config map")
	}

	ifCM, err := p.configMap.horizonConfigMap(
		p.config.ScannerPod.ImageFacade.Name,
		p.config.Namespace,
		fmt.Sprintf("%s.json", p.config.ScannerPod.ImageFacade.Name))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create image facade config map")
	}
	components.ConfigMaps = append(components.ConfigMaps, scannerCM, ifCM)
	log.Debugf("image facade configmap: %+v", ifCM.GetObj())
	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ScannerClusterRoleBinding())

	if p.config.Perceiver.EnablePodPerceiver {
		rc, err := p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		podPerceiverConfigMap, err := p.configMap.horizonConfigMap(
			p.config.Perceiver.PodPerceiver.Name,
			p.config.Namespace,
			fmt.Sprintf("%s.yaml", p.config.Perceiver.PodPerceiver.Name))
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, podPerceiverConfigMap)
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		cr := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(cr))
	}

	if p.config.Perceiver.EnableImagePerceiver {
		rc, err := p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		imagePerceiverConfigMap, err := p.configMap.horizonConfigMap(
			p.config.Perceiver.ImagePerceiver.Name,
			p.config.Namespace,
			fmt.Sprintf("%s.yaml", p.config.Perceiver.ImagePerceiver.Name))
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, imagePerceiverConfigMap)
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		cr := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(cr))
	}

	if p.config.EnableSkyfire {
		rc, err := p.PerceptorSkyfireReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PerceptorSkyfireService())
		configMap, err := p.configMap.horizonConfigMap(
			p.config.Skyfire.Name,
			p.config.Namespace,
			fmt.Sprintf("%s.yaml", p.config.Skyfire.Name))
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire configmap")
		}
		components.ConfigMaps = append(components.ConfigMaps, configMap)
		components.ServiceAccounts = append(components.ServiceAccounts, p.PerceptorSkyfireServiceAccount())
		cr := p.PerceptorSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PerceptorSkyfireClusterRoleBinding(cr))
	}

	if p.config.EnableMetrics {
		dep, err := p.PerceptorMetricsDeployment()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create metrics")
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, p.PerceptorMetricsService())
		perceptorCm, err := p.PerceptorMetricsConfigMap()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create perceptor config map")
		}
		components.ConfigMaps = append(components.ConfigMaps, perceptorCm)
	}

	return components, nil
}
