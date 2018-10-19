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
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type staticConfig struct {
	perceptorContainerName string
}

var (
	defaultStaticConfig = staticConfig{
		perceptorContainerName: "perceptor"}
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config *v1.OpsSightSpec
	//	staticConfig
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *v1.OpsSightSpec) *SpecConfig {
	return &SpecConfig{config: config} //, staticConfig: defaultStaticConfig}
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
	cm, err := p.PerceptorConfigMap()
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
	scannerCM, err := p.ScannerConfigMap()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	ifCM, err := p.ImageFacadeConfigMap()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
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
		perceiverConfigMap, err := p.PerceiverConfigMap(p.config.Perceiver.PodPerceiver.Name)
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, perceiverConfigMap)
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
		perceiverConfigMap, err := p.PerceiverConfigMap(p.config.Perceiver.ImagePerceiver.Name)
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, perceiverConfigMap)
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
		components.ConfigMaps = append(components.ConfigMaps, p.PerceptorSkyfireConfigMap())
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
