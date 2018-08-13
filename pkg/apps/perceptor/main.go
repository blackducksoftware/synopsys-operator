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

package perceptor

import (
	"encoding/json"
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/apps"
)

// App defines the perceptor application
type App struct {
	config *AppConfig
}

// NewApp creates a App object
func NewApp(defaults interface{}) (*App, error) {
	d, ok := defaults.(*AppConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert defaults")
	}
	p := App{config: d}

	p.setInternalDefaults()

	return &p, nil
}

// Configure will configure the perceptor app
func (p *App) Configure(config interface{}) error {
	return apps.MergeConfig(config, p.config)
}

// GetNamespace returns the namespace for this perceptor app
func (p *App) GetNamespace() string {
	return p.config.Namespace
}

// GetComponents will return the list of components for perceptor
func (p *App) GetComponents() (*apps.ComponentList, error) {
	p.configServiceAccounts()
	err := p.sanityCheckServices()
	if err != nil {
		return nil, fmt.Errorf("Please set the service accounts correctly; %v", err)
	}

	p.substituteDefaultImageVersion()

	components := &apps.ComponentList{}

	// Add Perceptor
	components.ReplicationControllers = append(components.ReplicationControllers, p.PerceptorReplicationController())
	components.Services = append(components.Services, p.PerceptorService())
	components.ConfigMaps = append(components.ConfigMaps, p.PerceptorConfigMap())

	// Add Perceptor Scanner
	rc, err := p.ScannerReplicationController()
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner replication controller: %v", err)
	}
	components.ReplicationControllers = append(components.ReplicationControllers, rc)
	components.Services = append(components.Services, p.ScannerService())
	components.Services = append(components.Services, p.ImageFacadeService())
	components.ConfigMaps = append(components.ConfigMaps, p.ScannerConfigMap())
	components.ConfigMaps = append(components.ConfigMaps, p.ImageFacadeConfigMap())
	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ScannerClusterRoleBinding())

	if p.config.PodPerceiver != nil && *p.config.PodPerceiver {
		rc, err := p.PodPerceiverReplicationController()
		if err != nil {
			return nil, fmt.Errorf("failed to create pod perceiver: %v", err)
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceiverConfigMap())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		cr := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(cr))
	}

	if p.config.ImagePerceiver != nil && *p.config.ImagePerceiver {
		rc, err := p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, fmt.Errorf("failed to create image perceiver: %v", err)
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceiverConfigMap())
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		cr := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(cr))
	}

	if p.config.PerceptorSkyfire != nil && *p.config.PerceptorSkyfire {
		rc, err := p.PerceptorSkyfireReplicationController()
		if err != nil {
			return nil, fmt.Errorf("failed to create skyfire: %v", err)
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PerceptorSkyfireService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceptorSkyfireConfigMap())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PerceptorSkyfireServiceAccount())
		cr := p.PerceptorSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PerceptorSkyfireClusterRoleBinding(cr))
	}

	if p.config.Metrics != nil && *p.config.Metrics {
		dep, err := p.PerceptorMetricsDeployment()
		if err != nil {
			return nil, fmt.Errorf("failed to create metrics: %v", err)
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, p.PerceptorMetricsService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceptorMetricsConfigMap())
	}

	return components, nil
}

func (p *App) substituteDefaultImageVersion() {
	if len(p.config.PerceptorImageVersion) == 0 {
		p.config.PerceptorImageVersion = p.config.DefaultVersion
	}
	if len(p.config.ScannerImageVersion) == 0 {
		p.config.ScannerImageVersion = p.config.DefaultVersion
	}
	if len(p.config.PerceiverImageVersion) == 0 {
		p.config.PerceiverImageVersion = p.config.DefaultVersion
	}
	if len(p.config.ImageFacadeImageVersion) == 0 {
		p.config.ImageFacadeImageVersion = p.config.DefaultVersion
	}
	if len(p.config.SkyfireImageVersion) == 0 {
		p.config.SkyfireImageVersion = p.config.DefaultVersion
	}
}

func (p *App) configServiceAccounts() {
	// TODO Viperize these env vars.
	if len(p.config.ServiceAccounts) == 0 {
		svcAccounts := map[string]string{
			// WARNING: These service accounts need to exist !
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner",
			"skyfire":                "skyfire",
		}
		p.config.ServiceAccounts = svcAccounts
	}
}

// TODO programatically validate rather then sanity check.
func (p *App) sanityCheckServices() error {
	isValid := func(cn string) bool {
		for _, valid := range []string{"perceptor", "pod-perceiver", "image-perceiver", "perceptor-scanner", "perceptor-image-facade", "skyfire"} {
			if cn == valid {
				return true
			}
		}
		return false
	}
	for cn := range p.config.ServiceAccounts {
		if !isValid(cn) {
			return fmt.Errorf("failed at verifiying that the container name for a svc account was valid")
		}
	}
	return nil
}

func (p *App) setInternalDefaults() {
	if len(p.config.HubUserPasswordEnvVar) == 0 {
		p.config.HubUserPasswordEnvVar = "PCP_HUBUSERPASSWORD"
	}
}

func (p *App) generateStringFromStringArr(strArr []string) string {
	str, _ := json.Marshal(strArr)
	return string(str)
}
