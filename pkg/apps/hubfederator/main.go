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

package hubfederator

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/apps"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
)

// App defines the hub federation application
type App struct {
	config *AppConfig
}

// NewApp creates a App object
func NewApp(defaults interface{}) (*App, error) {
	d, ok := defaults.(*AppConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert hub federation defaults: %v", defaults)
	}
	a := App{config: d}

	return &a, nil
}

// Configure will configure the hub federation app
func (a *App) Configure(config interface{}) error {
	return util.MergeConfig(config, a.config)
}

// GetNamespace returns the namespace for this hub federation app
func (a *App) GetNamespace() string {
	return a.config.Namespace
}

// GetComponents will return the list of components for hub federation
func (a *App) GetComponents() (*apps.ComponentList, error) {
	components := &apps.ComponentList{}

	// Add hub federation
	components.ReplicationControllers = append(components.ReplicationControllers, a.ReplicationController())
	for _, svc := range a.Services() {
		components.Services = append(components.Services, svc)
	}
	components.ConfigMaps = append(components.ConfigMaps, a.ConfigMap())
	components.ServiceAccounts = append(components.ServiceAccounts, a.ServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, a.ClusterRoleBinding())

	return components, nil
}
