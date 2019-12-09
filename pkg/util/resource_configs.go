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

package util

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// Container defines the configuration for a container
type Container struct {
	ContainerConfig       *horizonapi.ContainerConfig
	EnvConfigs            []*horizonapi.EnvConfig
	VolumeMounts          []*horizonapi.VolumeMountConfig
	PortConfig            []*horizonapi.PortConfig
	ActionConfig          *horizonapi.ActionConfig
	ReadinessProbeConfigs []*horizonapi.ProbeConfig
	LivenessProbeConfigs  []*horizonapi.ProbeConfig
	PreStopConfig         *horizonapi.ActionConfig
}

// PodConfig used for configuring the pod
type PodConfig struct {
	Name                   string
	Labels                 map[string]string
	ServiceAccount         string
	Containers             []*Container
	Volumes                []*components.Volume
	InitContainers         []*Container
	PodAffinityConfigs     map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	PodAntiAffinityConfigs map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	NodeAffinityConfigs    map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig
	ImagePullSecrets       []string
	FSGID                  *int64
	RunAsUser              *int64
	RunAsGroup             *int64
}
