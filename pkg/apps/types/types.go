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
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
)

// ComponentName denotes the component/resource name
type ComponentName string

// ContainerName denotes the container name
type ContainerName string

// ReplicationControllerCreater refers to Replication Controller creater
type ReplicationControllerCreater func(*ReplicationController, *protoform.Config, *kubernetes.Clientset, interface{}) (ReplicationControllerInterface, error)

// ServiceCreater refers to Service creater
type ServiceCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ServiceInterface, error)

// ConfigmapCreater refers to Replication Controller creater
type ConfigmapCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (ConfigMapInterface, error)

// PvcCreater refers to Persistent Volume Claim creater
type PvcCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (PVCInterface, error)

// SecretCreater refers to Secret creater
type SecretCreater func(*protoform.Config, *kubernetes.Clientset, interface{}) (SecretInterface, error)

// TagOrImage refers to Image and Tag
type TagOrImage struct {
	Tag   string
	Image string
}

// ConfigMapInterface refers to Config Map related interface
type ConfigMapInterface interface {
	GetCM() []*components.ConfigMap
}

// PVCInterface refers to PVC related interface
type PVCInterface interface {
	GetPVCs() ([]*components.PersistentVolumeClaim, error)
	// TODO add deployment, rc
}

// ReplicationController refers to replication controller configuration
type ReplicationController struct {
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

// ReplicationControllerInterface refers to replication controller related interface
type ReplicationControllerInterface interface {
	GetRc() (*components.ReplicationController, error)
	// TODO add deployment, rc
}

// SecretInterface refers to secret related interface
type SecretInterface interface {
	GetSecrets() []*components.Secret
}

// ServiceInterface refers to service related interface
type ServiceInterface interface {
	GetService() *components.Service
	// TODO add deployment, rc
}
