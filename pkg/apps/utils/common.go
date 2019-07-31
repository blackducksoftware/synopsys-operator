/*
Copyright (C) 2019 Synopsys, Inc.

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

package utils

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
)

// GetVersionLabel will return the label including the version
func GetVersionLabel(componentName string, name string, version string) map[string]string {
	m := GetLabel(componentName, name)
	m["version"] = version
	return m
}

// GetLabel will return the label
func GetLabel(componentName string, name string) map[string]string {
	return map[string]string{
		"app":       "blackduck",
		"name":      name,
		"component": componentName,
	}
}

// GetResourceName returns the name of the resource
func GetResourceName(name string, appName string, defaultName string) string {
	if len(appName) == 0 {
		return fmt.Sprintf("%s-%s", name, defaultName)
	}

	if len(defaultName) == 0 {
		return fmt.Sprintf("%s-%s", name, appName)
	}

	return fmt.Sprintf("%s-%s-%s", name, appName, defaultName)
}

// SetLimits set the container limits
func SetLimits(containerConfig *api.ContainerConfig, config types.Container) {
	if config.MinCPU != nil {
		containerConfig.MinCPU = fmt.Sprintf("%d", *config.MinCPU)
	}

	if config.MaxCPU != nil {
		containerConfig.MaxCPU = fmt.Sprintf("%d", *config.MaxCPU)
	}

	if config.MinMem != nil {
		containerConfig.MinMem = fmt.Sprintf("%dM", *config.MinMem)
	}

	if config.MaxMem != nil {
		containerConfig.MaxMem = fmt.Sprintf("%dM", *config.MaxMem)
	}
}
