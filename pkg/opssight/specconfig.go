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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/juju/errors"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config    *opssightapi.OpsSightSpec
	configMap *MainOpssightConfigMap
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *opssightapi.OpsSightSpec) *SpecConfig {
	configMap := &MainOpssightConfigMap{
		LogLevel: config.LogLevel,
		BlackDuck: &BlackDuckConfig{
			ConnectionsEnvironmentVariableName: config.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    config.Blackduck.TLSVerification,
		},
		ImageFacade: &ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             config.ScannerPod.ImageFacade.Port,
			ImagePullerType:  config.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &PerceiverConfig{
			Image: &ImagePerceiverConfig{},
			Pod: &PodPerceiverConfig{
				NamespaceFilter: config.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: config.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       config.Perceiver.DumpIntervalMinutes,
			Port:                      config.Perceiver.Port,
		},
		Perceptor: &PerceptorConfig{
			Timings: &PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: config.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      config.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       config.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  config.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  config.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        config.Perceptor.Name,
			Port:        config.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: &ScannerConfig{
			BlackDuckClientTimeoutSeconds: config.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                config.ScannerPod.ImageDirectory,
			Port:                          config.ScannerPod.Scanner.Port,
		},
		Skyfire: &SkyfireConfig{
			BlackDuckClientTimeoutSeconds: config.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     config.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       config.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  config.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          config.Skyfire.Port,
			PrometheusPort:                config.Skyfire.PrometheusPort,
			UseInClusterConfig:            true,
		},
	}
	return &SpecConfig{config: config, configMap: configMap}
}

func (p *SpecConfig) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: p.config.ConfigMapName,
	})
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add config map
	cm, err := p.configMap.horizonConfigMap(
		p.config.ConfigMapName,
		p.config.Namespace,
		fmt.Sprintf("%s.json", p.config.ConfigMapName))
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ConfigMaps = append(components.ConfigMaps, cm)

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
	components.Secrets = append(components.Secrets, p.PerceptorSecret())

	// Add Perceptor Scanner
	scannerRC, err := p.ScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.ScannerService(), p.ImageFacadeService())

	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ScannerClusterRoleBinding())

	// Add Pod Perceiver
	if p.config.Perceiver.EnablePodPerceiver {
		rc, err = p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		cr := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(cr))
	}

	// Add Image Perceiver
	if p.config.Perceiver.EnableImagePerceiver {
		rc, err = p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		cr := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(cr))
	}

	// Add skyfire
	if p.config.EnableSkyfire {
		skyfireRC, err := p.PerceptorSkyfireReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, skyfireRC)
		components.Services = append(components.Services, p.PerceptorSkyfireService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PerceptorSkyfireServiceAccount())
		skyfireClusterRole := p.PerceptorSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, skyfireClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PerceptorSkyfireClusterRoleBinding(skyfireClusterRole))
	}

	// Add Metrics
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
