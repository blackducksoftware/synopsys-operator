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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	kubeClient *kubernetes.Clientset
	config     *opssightapi.OpsSightSpec
	configMap  *MainOpssightConfigMap
	dryRun     bool
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(kubeClient *kubernetes.Clientset, config *opssightapi.OpsSightSpec, dryRun bool) *SpecConfig {
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
	return &SpecConfig{kubeClient: kubeClient, config: config, configMap: configMap, dryRun: dryRun}
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
	scannerClusterRoleBinding := p.ScannerClusterRoleBinding()
	if p.dryRun {
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, scannerClusterRoleBinding)
	} else {
		clusterRoleBinding, err := util.GetClusterRoleBinding(p.kubeClient, scannerClusterRoleBinding.GetName())
		if err != nil {
			log.Debugf("%s cluster role binding not exist!!!", scannerClusterRoleBinding.GetName())
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, scannerClusterRoleBinding)
		} else {
			if !isClusterRoleBindingSubjectExist(clusterRoleBinding.Subjects, p.config.Namespace) {
				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{Name: scannerClusterRoleBinding.GetName(), Namespace: p.config.Namespace, Kind: "ServiceAccount"})
				_, err = util.UpdateClusterRoleBinding(p.kubeClient, clusterRoleBinding)
				if err != nil {
					return nil, errors.Annotate(err, fmt.Sprintf("failed to update the %s cluster role binding", scannerClusterRoleBinding.GetName()))
				}
			}
		}
	}

	// Add Pod Perceiver
	if p.config.Perceiver.EnablePodPerceiver {
		rc, err = p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		podClusterRole := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, podClusterRole)

		podClusterRoleBinding := p.PodPerceiverClusterRoleBinding(podClusterRole)
		if p.dryRun {
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, podClusterRoleBinding)
		} else {
			clusterRoleBinding, err := util.GetClusterRoleBinding(p.kubeClient, podClusterRoleBinding.GetName())
			if err != nil {
				log.Debugf("%s cluster role binding not exist!!!", podClusterRoleBinding.GetName())
				components.ClusterRoleBindings = append(components.ClusterRoleBindings, podClusterRoleBinding)
			} else {
				if !isClusterRoleBindingSubjectExist(clusterRoleBinding.Subjects, p.config.Namespace) {
					clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{Name: podClusterRoleBinding.GetName(), Namespace: p.config.Namespace, Kind: "ServiceAccount"})
					_, err = util.UpdateClusterRoleBinding(p.kubeClient, clusterRoleBinding)
					if err != nil {
						return nil, errors.Annotate(err, fmt.Sprintf("failed to update the %s cluster role binding", podClusterRoleBinding.GetName()))
					}
				}
			}
		}
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
		imageClusterRole := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, imageClusterRole)

		imageClusterRoleBinding := p.ImagePerceiverClusterRoleBinding(imageClusterRole)
		if p.dryRun {
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, imageClusterRoleBinding)
		} else {
			clusterRoleBinding, err := util.GetClusterRoleBinding(p.kubeClient, imageClusterRoleBinding.GetName())
			if err != nil {
				log.Debugf("%s cluster role binding not exist!!!", imageClusterRoleBinding.GetName())
				components.ClusterRoleBindings = append(components.ClusterRoleBindings, imageClusterRoleBinding)
			} else {
				if !isClusterRoleBindingSubjectExist(clusterRoleBinding.Subjects, p.config.Namespace) {
					clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{Name: imageClusterRoleBinding.GetName(), Namespace: p.config.Namespace, Kind: "ServiceAccount"})
					_, err = util.UpdateClusterRoleBinding(p.kubeClient, clusterRoleBinding)
					if err != nil {
						return nil, errors.Annotate(err, fmt.Sprintf("failed to update the %s cluster role binding", imageClusterRoleBinding.GetName()))
					}
				}
			}
		}
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

		skyfireClusterRoleBinding := p.PerceptorSkyfireClusterRoleBinding(skyfireClusterRole)
		if p.dryRun {
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, skyfireClusterRoleBinding)
		} else {
			clusterRoleBinding, err := util.GetClusterRoleBinding(p.kubeClient, skyfireClusterRoleBinding.GetName())
			if err != nil {
				log.Debugf("%s cluster role binding not exist!!!", skyfireClusterRoleBinding.GetName())
				components.ClusterRoleBindings = append(components.ClusterRoleBindings, skyfireClusterRoleBinding)
			} else {
				if !isClusterRoleBindingSubjectExist(clusterRoleBinding.Subjects, p.config.Namespace) {
					clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{Name: skyfireClusterRoleBinding.GetName(), Namespace: p.config.Namespace, Kind: "ServiceAccount"})
					_, err = util.UpdateClusterRoleBinding(p.kubeClient, clusterRoleBinding)
					if err != nil {
						return nil, errors.Annotate(err, fmt.Sprintf("failed to update the %s cluster role binding", skyfireClusterRoleBinding.GetName()))
					}
				}
			}
		}
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

func isClusterRoleBindingSubjectExist(subjects []rbacv1.Subject, namespace string) bool {
	for _, subject := range subjects {
		if strings.EqualFold(subject.Namespace, namespace) {
			return true
		}
	}
	return false
}
