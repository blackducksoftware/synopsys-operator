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
	"math"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config         *protoform.Config
	kubeClient     *kubernetes.Clientset
	opssightClient *opssightclientset.Clientset
	hubClient      *hubclientset.Clientset
	opssight       *opssightapi.OpsSight
	configMap      *MainOpssightConfigMap
	dryRun         bool
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *protoform.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, hubClient *hubclientset.Clientset, opssight *opssightapi.OpsSight, dryRun bool) *SpecConfig {
	opssightSpec := &opssight.Spec
	configMap := &MainOpssightConfigMap{
		LogLevel: opssightSpec.LogLevel,
		BlackDuck: &BlackDuckConfig{
			ConnectionsEnvironmentVariableName: opssightSpec.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    opssightSpec.Blackduck.TLSVerification,
		},
		ImageFacade: &ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             opssightSpec.ScannerPod.ImageFacade.Port,
			ImagePullerType:  opssightSpec.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &PerceiverConfig{
			Image: &ImagePerceiverConfig{},
			Pod: &PodPerceiverConfig{
				NamespaceFilter: opssightSpec.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: opssightSpec.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       opssightSpec.Perceiver.DumpIntervalMinutes,
			Port:                      opssightSpec.Perceiver.Port,
		},
		Perceptor: &PerceptorConfig{
			Timings: &PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: opssightSpec.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      opssightSpec.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       opssightSpec.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  opssightSpec.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  opssightSpec.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        opssightSpec.Perceptor.Name,
			Port:        opssightSpec.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: &ScannerConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                opssightSpec.ScannerPod.ImageDirectory,
			Port:                          opssightSpec.ScannerPod.Scanner.Port,
		},
		Skyfire: &SkyfireConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     opssightSpec.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       opssightSpec.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  opssightSpec.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          opssightSpec.Skyfire.Port,
			PrometheusPort:                opssightSpec.Skyfire.PrometheusPort,
			UseInClusterConfig:            true,
		},
	}
	return &SpecConfig{config: config, kubeClient: kubeClient, opssightClient: opssightClient, hubClient: hubClient, opssight: opssight, configMap: configMap, dryRun: dryRun}
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add config map
	cm, err := p.configMap.horizonConfigMap(
		p.opssight.Spec.ConfigMapName,
		p.opssight.Spec.Namespace,
		fmt.Sprintf("%s.json", p.opssight.Spec.ConfigMapName))
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ConfigMaps = append(components.ConfigMaps, cm)

	// Add Perceptor
	rc, err := p.GetPerceptorReplicationController()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ReplicationControllers = append(components.ReplicationControllers, rc)
	service, err := p.GetPerceptorService()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.Services = append(components.Services, service)
	perceptorSvc, err := p.GetPerceptorExposeService()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create perceptor service")
	}
	if perceptorSvc != nil {
		components.Services = append(components.Services, perceptorSvc)
	}
	secret := p.GetPerceptorSecret()
	if !p.dryRun {
		p.addSecretData(secret)
	}
	components.Secrets = append(components.Secrets, secret)

	route := p.GetPerceptorOpenShiftRoute()
	if route != nil {
		components.Routes = append(components.Routes, route)
	}

	// Add Perceptor Scanner
	scannerRC, err := p.GetScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.GetScannerService(), p.GetImageFacadeService())

	components.ServiceAccounts = append(components.ServiceAccounts, p.GetScannerServiceAccount())
	clusterRoleBinding, err := p.GetScannerClusterRoleBinding()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner cluster role binding")
	}
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, clusterRoleBinding)

	// Add Pod Perceiver
	if p.opssight.Spec.Perceiver.EnablePodPerceiver {
		rc, err = p.GetPodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.GetPodPerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.GetPodPerceiverServiceAccount())
		podClusterRole := p.GetPodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, podClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.GetPodPerceiverClusterRoleBinding(podClusterRole))
	}

	// Add Image Perceiver
	if p.opssight.Spec.Perceiver.EnableImagePerceiver {
		rc, err = p.GetImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.GetImagePerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.GetImagePerceiverServiceAccount())
		imageClusterRole := p.GetImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, imageClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.GetImagePerceiverClusterRoleBinding(imageClusterRole))
	}

	// Add skyfire
	if p.opssight.Spec.EnableSkyfire {
		skyfireRC, err := p.GetSkyfireReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, skyfireRC)
		components.Services = append(components.Services, p.GetSkyfireService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.GetSkyfireServiceAccount())
		skyfireClusterRole := p.GetSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, skyfireClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.GetSkyfireClusterRoleBinding(skyfireClusterRole))
	}

	// Add Metrics
	if p.opssight.Spec.EnableMetrics {
		// deployments
		dep, err := p.GetPrometheusDeployment()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create metrics")
		}
		components.Deployments = append(components.Deployments, dep)

		// services
		prometheusService, err := p.GetPrometheusService()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create prometheus metrics service")
		}
		components.Services = append(components.Services, prometheusService)
		prometheusSvc, err := p.GetPrometheusExposeService()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create prometheus metrics exposed service")
		}
		if prometheusSvc != nil {
			components.Services = append(components.Services, prometheusSvc)
		}

		// config map
		perceptorCm, err := p.GetPrometheusConfigMap()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create perceptor config map")
		}
		components.ConfigMaps = append(components.ConfigMaps, perceptorCm)

		route := p.GetPrometheusOpenShiftRoute()
		if route != nil {
			components.Routes = append(components.Routes, route)
		}
	}

	return components, nil
}

func (p *SpecConfig) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: p.opssight.Spec.ConfigMapName,
		DefaultMode:     util.IntToInt32(420),
	})
}

func (p *SpecConfig) addSecretData(secret *components.Secret) error {
	blackduckHosts := make(map[string]*opssightapi.Host)
	// adding External Black Duck credentials
	for _, host := range p.opssight.Spec.Blackduck.ExternalHosts {
		blackduckHosts[host.Domain] = host
	}

	// adding Internal Black Duck credentials
	hubType := p.opssight.Spec.Blackduck.BlackduckSpec.Type
	blackduckPassword, err := util.Base64Decode(p.opssight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return errors.Annotatef(err, "unable to decode blackduckPassword")
	}

	allHubs := p.getAllHubs(hubType, blackduckPassword)
	blackduckPasswords := util.AppendBlackDuckSecrets(blackduckHosts, p.opssight.Status.InternalHosts, allHubs)

	// marshal the blackduck credentials to bytes
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal blackduck passwords")
	}
	secret.AddData(map[string][]byte{p.opssight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credentials
	securedRegistries := make(map[string]*opssightapi.RegistryAuth)
	for _, internalRegistry := range p.opssight.Spec.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = internalRegistry
	}
	// marshal the Secured registries credentials to bytes
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal secured registries")
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	// add internal hosts to status
	p.opssight.Status.InternalHosts = util.AppendBlackDuckHosts(p.opssight.Status.InternalHosts, allHubs)
	return nil
}

// getAllHubs get only the internal Black Duck instances from the cluster
func (p *SpecConfig) getAllHubs(hubType string, blackduckPassword string) []*opssightapi.Host {
	hosts := []*opssightapi.Host{}
	hubsList, err := util.ListHubs(p.hubClient, p.config.Namespace)
	if err != nil {
		log.Errorf("unable to list blackducks due to %+v", err)
	}
	for _, hub := range hubsList.Items {
		if strings.EqualFold(hub.Spec.Type, hubType) {
			var concurrentScanLimit int
			switch strings.ToUpper(hub.Spec.Size) {
			case "MEDIUM":
				concurrentScanLimit = 3
			case "LARGE":
				concurrentScanLimit = 4
			case "X-LARGE":
				concurrentScanLimit = 6
			default:
				concurrentScanLimit = 2
			}
			host := &opssightapi.Host{Domain: fmt.Sprintf("webserver.%s.svc", hub.Name), ConcurrentScanLimit: concurrentScanLimit, Scheme: "https", User: "sysadmin", Port: 443, Password: blackduckPassword}
			hosts = append(hosts, host)
		}
	}
	log.Debugf("total no of Black Duck's for type %s is %d", hubType, len(hosts))
	return hosts
}

// getDefaultPassword get the default password for the hub
func (p *SpecConfig) getDefaultPassword() string {
	var hubPassword string
	var err error
	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		_, _, _, hubPassword, err = util.GetDefaultPasswords(p.kubeClient, p.config.Namespace)
		if err == nil {
			break
		} else {
			log.Infof("wasn't able to get hub password, sleeping 5 seconds.  try = %v", dbInitTry)
			time.Sleep(5 * time.Second)
		}
	}
	return hubPassword
}
