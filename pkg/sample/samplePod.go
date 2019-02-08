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

package sample

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
)

type SpecConfig struct {
	config *v1.SampleSpec
}

// NewSample will create the Sample object
func NewSample(config *v1.SampleSpec) *SpecConfig {
	return &SpecConfig{config: config}
}

// GetComponents will return the list of components for sample
func (a *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}
	return components, nil
}

// Create a Sample Pod using Horizon API
func (sampleSpecConfig *SpecConfig) samplePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "sample",
	})

	pod.AddContainer(sampleSpecConfig.sampleContainer())

	return pod, nil
}

// Create a Sample Container using Horizon API
func (sampleSpecConfig *SpecConfig) sampleContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:  "sample",
		Image: "registry.hub.docker.com/vasiliys/public-test:latest",
	})

	return container
}
