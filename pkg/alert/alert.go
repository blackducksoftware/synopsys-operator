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

package alert

import (
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
)

// SpecConfig will contain the specification of Alert
type SpecConfig struct {
	config *alertapi.AlertSpec
}

// NewAlert will create the Alert object
func NewAlert(config *alertapi.AlertSpec) *SpecConfig {
	return &SpecConfig{config: config}
}

// GetComponents will return the list of components for alert
func (a *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add alert components
	components.ConfigMaps = append(components.ConfigMaps, a.getAlertConfigMap())

	dep, err := a.getAlertDeployment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Deployment: %s", err)
	}
	components.Deployments = append(components.Deployments, dep)

	switch a.config.ExposeService {
	case "NODEPORT":
		components.Services = append(components.Services, a.getAlertServiceNodePort())
	case "LOADBALANCER":
		components.Services = append(components.Services, a.getAlertServiceLoadBalancer())
	}

	sec, err := a.GetAlertSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Secret: %s", err)
	}
	components.Secrets = append(components.Secrets, sec)

	if a.config.PersistentStorage {
		pvc, err := a.getAlertPersistentVolumeClaim()
		if err != nil {
			return nil, fmt.Errorf("failed to create Alert's PVC: %s", err)
		}
		components.PersistentVolumeClaims = append(components.PersistentVolumeClaims, pvc)
	}

	// Add cfssl if running in stand alone mode
	if *a.config.StandAlone {
		dep, err := a.getCfsslDeployment()
		if err != nil {
			return nil, fmt.Errorf("failed to create cfssl deployment: %v", err)
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, a.getCfsslService())
	}

	return components, nil
}
