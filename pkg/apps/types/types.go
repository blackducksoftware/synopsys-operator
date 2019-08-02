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

package types

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

// ComponentName denotes the component/resource name
type ComponentName string

// ContainerName denotes the container name
type ContainerName string

// DeploymentCreater refers to Deployment creater
type DeploymentCreater func(*PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (DeploymentInterface, error)

// ReplicationControllerCreater refers to Replication Controller creater
type ReplicationControllerCreater func(*PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (ReplicationControllerInterface, error)

// ServiceCreater refers to Service creater
type ServiceCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ServiceInterface, error)

// ConfigmapCreater refers to Replication Controller creater
type ConfigmapCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ConfigMapInterface, error)

// PvcCreater refers to Persistent Volume Claim creater
type PvcCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (PVCInterface, error)

// SecretCreater refers to Secret creater
type SecretCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (SecretInterface, error)

// ClusterRoleCreater refers to Cluster role creater
type ClusterRoleCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ClusterRoleInterface, error)

// ClusterRoleBindingCreater refers to Cluster role binding creater
type ClusterRoleBindingCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ClusterRoleBindingInterface, error)

// ServiceAccountCreater refers to Service account creater
type ServiceAccountCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ServiceAccountInterface, error)

// RouteCreater refers to Route creater
type RouteCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (RouteInterface, error)

// TagOrImage refers to Image and Tag
type TagOrImage struct {
	Tag   string
	Image string
}

// ConfigMapInterface refers to Config Map related interface
type ConfigMapInterface interface {
	GetCM() (*components.ConfigMap, error)
}

// PVCInterface refers to PVC related interface
type PVCInterface interface {
	GetPVCs() ([]*components.PersistentVolumeClaim, error)
}

// PodResource refers to resource configuration that shares between RC and deployment
type PodResource struct {
	Namespace  string
	Replicas   int
	Containers map[ContainerName]Container
}

// Container refers to container configuration
type Container struct {
	Image  string
	MinCPU *int32
	MaxCPU *int32
	MinMem *int32
	MaxMem *int32
}

// DeploymentInterface refers to deployment related interface
type DeploymentInterface interface {
	GetDeployment() (*components.Deployment, error)
}

// ReplicationControllerInterface refers to replication controller related interface
type ReplicationControllerInterface interface {
	GetRc() (*components.ReplicationController, error)
}

// SecretInterface refers to secret related interface
type SecretInterface interface {
	GetSecret() (*components.Secret, error)
}

// ServiceInterface refers to service related interface
type ServiceInterface interface {
	GetService() (*components.Service, error)
}

// ClusterRoleInterface refers to cluster role related interface
type ClusterRoleInterface interface {
	GetClusterRole() (*components.ClusterRole, error)
}

// ClusterRoleBindingInterface refers to cluster role bindings related interface
type ClusterRoleBindingInterface interface {
	GetClusterRoleBinding() (*components.ClusterRoleBinding, error)
}

// ServiceAccountInterface refers to service account related interface
type ServiceAccountInterface interface {
	GetServiceAccount() (*components.ServiceAccount, error)
}

// RouteInterface refers to route related interface
type RouteInterface interface {
	GetRoute() (*api.Route, error)
}
