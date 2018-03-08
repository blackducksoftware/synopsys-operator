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

package model

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PerceptorConfigMap struct {
	HubHost             string
	HubUser             string
	HubUserPassword     string
	ConcurrentScanLimit int
	UseMockMode         bool
}

type PerceptorCore struct {
	PodName string
	Image   string
	Port    int32
	CPU     resource.Quantity
	Memory  resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         PerceptorConfigMap

	ReplicaCount int32
	ServiceName  string
}

func NewPerceptorCore() *PerceptorCore {
	memory, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}

	return &PerceptorCore{
		PodName:        "perceptor",
		Image:          "gcr.io/gke-verification/blackducksoftware/perceptor:master",
		Port:           3001,
		CPU:            cpu,
		Memory:         memory,
		ConfigMapName:  "perceptor-config",
		ConfigMapMount: "/etc/perceptor",
		ConfigMapPath:  "perceptor_conf.yaml",
		ReplicaCount:   1,
		ServiceName:    "perceptor",
	}
}

func (pc *PerceptorCore) Container() *v1.Container {
	return &v1.Container{
		Name:            "perceptor",
		Image:           pc.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: pc.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    pc.CPU,
				v1.ResourceMemory: pc.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      pc.ConfigMapName,
				MountPath: pc.ConfigMapMount,
			},
		},
	}
}

func (pc *PerceptorCore) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: pc.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &pc.ReplicaCount,
			Selector: map[string]string{"name": pc.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": pc.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: pc.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: pc.ConfigMapName},
								},
							},
						},
					},
					Containers: []v1.Container{*pc.Container()},
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (pc *PerceptorCore) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: pc.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: pc.ServiceName,
					Port: pc.Port,
				},
			},
			Selector: map[string]string{"name": pc.ServiceName}}}
}

func (pc *PerceptorCore) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(pc.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(pc.ConfigMapName, pc.ConfigMapPath, string(jsonBytes))
}
