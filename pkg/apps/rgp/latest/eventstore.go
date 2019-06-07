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

package rgp

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Eventstore stores the event store configuration
type Eventstore struct {
	Namespace     string
	StorageClass  string
	DiskSizeInGiB int
}

// NewEventstore returns the event store
func NewEventstore(namespace string, storageClass string, diskSizeInGiB int) *Eventstore {
	// Set the disk size to 100Gb if the size is not provided.
	if diskSizeInGiB == 0 {
		diskSizeInGiB = 100
	}
	return &Eventstore{Namespace: namespace, StorageClass: storageClass, DiskSizeInGiB: diskSizeInGiB}
}

// GetEventStoreStatefulSet will return the postgres deployment
func (e *Eventstore) GetEventStoreStatefulSet() *components.StatefulSet {
	envs := e.getEventStoreEnvConfigs()
	volumeMounts := e.getEventStoreVolumeMounts()

	var containers []*util.Container

	containers = append(containers, &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "eventstore",
			Image:      "gcr.io/snps-swip-staging/swip_eventstore:0.0.8",
			PullPolicy: horizonapi.PullIfNotPresent,
			MinCPU:     "1000m",
			MinMem:     "8Gi",
		},
		EnvConfigs:   envs,
		VolumeMounts: volumeMounts,
		PortConfig: []*horizonapi.PortConfig{
			{ContainerPort: 1112, Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: 1113, Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: 2112, Protocol: horizonapi.ProtocolTCP},
			{ContainerPort: 2113, Protocol: horizonapi.ProtocolTCP},
		},
	})

	stateFulSetConfig := &horizonapi.StatefulSetConfig{
		Name:      "eventstore",
		Namespace: e.Namespace,
		Replicas:  util.IntToInt32(3),
		Service:   "eventstore",
	}

	labelSelector := map[string]string{
		"app":       "rgp",
		"component": "eventstore",
	}

	// TODO add service account
	stateFulSet := util.CreateStateFulSetFromContainer(stateFulSetConfig, "", containers, nil, nil, nil, labelSelector)
	stateFulSet.AddLabels(labelSelector)
	stateFulSet.AddMatchLabelsSelectors(labelSelector)

	claim, _ := util.CreatePersistentVolumeClaim("data", e.Namespace, fmt.Sprintf("%dGi", e.DiskSizeInGiB), e.StorageClass, horizonapi.ReadWriteOnce)
	stateFulSet.AddVolumeClaimTemplate(*claim)
	return stateFulSet
}

// GetEventStoreService will return the event store service
func (e *Eventstore) GetEventStoreService() *components.Service {
	// Consul service
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      "eventstore",
		Namespace: e.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{
		"app":       "rgp",
		"component": "eventstore",
	})
	service.AddSelectors(map[string]string{
		"component": "eventstore",
	})
	service.AddPort(horizonapi.ServicePortConfig{Name: "int-tcp", Port: 1112})
	service.AddPort(horizonapi.ServicePortConfig{Name: "int-http", Port: 1113})
	service.AddPort(horizonapi.ServicePortConfig{Name: "ext-tcp", Port: 2112})
	service.AddPort(horizonapi.ServicePortConfig{Name: "ext-http", Port: 2113})

	return service
}

// getConsulVolumeMounts will return the postgres volume mounts
func (e *Eventstore) getEventStoreVolumeMounts() []*horizonapi.VolumeMountConfig {
	var volumeMounts []*horizonapi.VolumeMountConfig
	volumeMounts = append(volumeMounts, &horizonapi.VolumeMountConfig{Name: "data", MountPath: "/var/lib/eventstore"})
	return volumeMounts
}

// getConsulEnvConfigs will return the postgres environment config maps
func (e *Eventstore) getEventStoreEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENTSTORE_CLUSTER_SIZE", KeyOrVal: "3"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENTSTORE_CLUSTER_DNS", KeyOrVal: "eventstore"})

	return envs
}

// GetInitJob ...
func (e *Eventstore) GetInitJob() *v1.Job {

	job := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eventstore-init",
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "eventstore-init",
					Containers: []corev1.Container{
						{
							Name:            "eventstore-init",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           "gcr.io/snps-swip-staging/eventstore-util:0.0.14",
							Command:         []string{"eventstore-init"},
							Env: []corev1.EnvVar{
								{
									Name:  "EVENTSTORE_KUBERNETES_NAMESPACE",
									Value: e.Namespace,
								},
								{
									Name:  "EVENTSTORE_SECRET_NAME",
									Value: "swip-eventstore-creds",
								},
								{
									Name:  "EVENTSTORE_ADDR",
									Value: "http://eventstore:2113",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	return job
}
