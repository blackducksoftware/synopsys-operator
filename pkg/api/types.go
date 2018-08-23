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

package api

import (
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/perceptor"

	kapi "github.com/blackducksoftware/horizon/pkg/api"

	"github.com/koki/short/types"
)

// ReplicationControllerConfig defines the configuration for a
// replication controller
type ReplicationControllerConfig struct {
	Name           string
	Replicas       int32
	Selector       map[string]string
	Labels         map[string]string
	Vols           map[string]types.Volume
	Containers     []types.Container
	ServiceAccount string
}

// PodConfig defines the configuration for a pod
type PodConfig struct {
	Name           string
	Labels         map[string]string
	Vols           map[string]types.Volume
	Containers     []types.Container
	ServiceAccount string
}

// ServiceConfig defines the configuration for a service
type ServiceConfig struct {
	Name          string
	IPServiceType types.ClusterIPServiceType
	Ports         map[string]int32
	Selector      map[string]string
	Annotations   map[string]string
	Labels        map[string]string
}

// ConfigMapConfig defines the configuration for a config map
type ConfigMapConfig struct {
	Name      string
	Namespace string
	Data      map[string]string
}

// ServiceAccountConfig defines the configuration for a service account
type ServiceAccountConfig struct {
	Name      string
	Namespace string
}

// DeploymentConfig defines the configuration for a deployment
type DeploymentConfig struct {
	Name      string
	Namespace string
	Replicas  int32
	Selector  types.RSSelector
	Pod       PodConfig
}

// ClusterRoleConfig defines the configuration for a cluster role
type ClusterRoleConfig struct {
	Version string
	Name    string
	Rules   []types.PolicyRule
}

// ClusterRoleBindingConfig defines the configuration for a cluster role binding
type ClusterRoleBindingConfig struct {
	Version  string
	Name     string
	Subjects []types.Subject
	RoleRef  types.RoleRef
}

// Container defines the configuration for a container
type Container struct {
	ContainerConfig *kapi.ContainerConfig
	EnvConfigs      []*kapi.EnvConfig
	VolumeMounts    []*kapi.VolumeMountConfig
	PortConfig      *kapi.PortConfig
	ActionConfig    *kapi.ActionConfig
}

// ProtoformConfig defines the configuration for protoform
type ProtoformConfig struct {
	// Dry run wont actually install, but will print the objects definitions out.
	DryRun bool `json:"dryRun,omitempty"`

	HubUserPassword string `json:"hubUserPassword"`

	// Viper secrets
	ViperSecret string `json:"viperSecret,omitempty"`

	// Log level
	DefaultLogLevel string `json:"defaultLogLevel,omitempty"`

	Apps *ProtoformApps `json:"apps,omitempty"`
}

// ProtoformApps defines the configuration for supported apps
type ProtoformApps struct {
	PerceptorConfig *perceptor.AppConfig `json:"perceptorConfig,omitempty"`
	AlertConfig     *alert.AppConfig     `json:"alertConfig,omitempty"`
}
