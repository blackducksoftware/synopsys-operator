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

package alert

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/apps"
)

// App defines the alert application
type App struct {
	config *AppConfig
}

// NewApp creates a App object
func NewApp(defaults interface{}) (*App, error) {
	d, ok := defaults.(*AppConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert alert defaults: %v", defaults)
	}
	a := App{config: d}

	return &a, nil
}

// Configure will configure the alert app
func (a *App) Configure(config interface{}) error {
	return apps.MergeConfig(config, a.config)
}

// GetNamespace returns the namespace for this alert app
func (a *App) GetNamespace() string {
	return a.config.Namespace
}

// GetComponents will return the list of components for alert
func (a *App) GetComponents() (*apps.ComponentList, error) {
	components := &apps.ComponentList{}

	// Add alert
	dep, err := a.AlertDeployment()
	if err != nil {
		return nil, fmt.Errorf("failed to create alert deployment: %v", err)
	}
	components.Deployments = append(components.Deployments, dep)
	components.Services = append(components.Services, a.AlertService())
	components.ConfigMaps = append(components.ConfigMaps, a.ConfigMap())

	// Add cfssl if running in stand alone mode
	if *a.config.StandAlone {
		dep, err := a.CfsslDeployment()
		if err != nil {
			return nil, fmt.Errorf("failed to create cfssl deployment: %v", err)
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, a.CfsslService())
	}

	return components, nil
}
