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
	"encoding/json"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	kubeClient *kubernetes.Clientset
	opssight   *opssightapi.OpsSightSpec
	configMap  *MainOpssightConfigMap
	dryRun     bool
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(kubeClient *kubernetes.Clientset, opssight *opssightapi.OpsSightSpec, dryRun bool) *SpecConfig {
	configMap := &MainOpssightConfigMap{
		LogLevel: opssight.LogLevel,
		BlackDuck: &BlackDuckConfig{
			ConnectionsEnvironmentVariableName: opssight.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    opssight.Blackduck.TLSVerification,
		},
		ImageFacade: &ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             opssight.ScannerPod.ImageFacade.Port,
			ImagePullerType:  opssight.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &PerceiverConfig{
			Image: &ImagePerceiverConfig{},
			Pod: &PodPerceiverConfig{
				NamespaceFilter: opssight.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: opssight.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       opssight.Perceiver.DumpIntervalMinutes,
			Port:                      opssight.Perceiver.Port,
		},
		Perceptor: &PerceptorConfig{
			Timings: &PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: opssight.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      opssight.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       opssight.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  opssight.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  opssight.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        opssight.Perceptor.Name,
			Port:        opssight.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: &ScannerConfig{
			BlackDuckClientTimeoutSeconds: opssight.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                opssight.ScannerPod.ImageDirectory,
			Port:                          opssight.ScannerPod.Scanner.Port,
		},
		Skyfire: &SkyfireConfig{
			BlackDuckClientTimeoutSeconds: opssight.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     opssight.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       opssight.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  opssight.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          opssight.Skyfire.Port,
			PrometheusPort:                opssight.Skyfire.PrometheusPort,
			UseInClusterConfig:            true,
		},
	}
	return &SpecConfig{kubeClient: kubeClient, opssight: opssight, configMap: configMap, dryRun: dryRun}
}

func (p *SpecConfig) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: p.opssight.ConfigMapName,
	})
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add config map
	cm, err := p.configMap.horizonConfigMap(
		p.opssight.ConfigMapName,
		p.opssight.Namespace,
		fmt.Sprintf("%s.json", p.opssight.ConfigMapName))
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
	secret := p.PerceptorSecret()
	p.addSecretData(secret)
	components.Secrets = append(components.Secrets, secret)

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
	if p.opssight.Perceiver.EnablePodPerceiver {
		rc, err = p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		podClusterRole := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, podClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(podClusterRole))
	}

	// Add Image Perceiver
	if p.opssight.Perceiver.EnableImagePerceiver {
		rc, err = p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		imageClusterRole := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, imageClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(imageClusterRole))
	}

	// Add skyfire
	if p.opssight.EnableSkyfire {
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
	if p.opssight.EnableMetrics {
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

func (p *SpecConfig) addSecretData(secret *components.Secret) error {
	blackduckPasswords := make(map[string]interface{})
	// adding External Black Duck passwords
	for _, host := range p.opssight.Blackduck.ExternalHosts {
		blackduckPasswords[host.Domain] = &host
	}
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return errors.Trace(err)
	}
	secret.AddData(map[string][]byte{p.opssight.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credential
	securedRegistries := make(map[string]interface{})
	for _, internalRegistry := range p.opssight.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = &internalRegistry
	}
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return errors.Trace(err)
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})
	return nil
}
