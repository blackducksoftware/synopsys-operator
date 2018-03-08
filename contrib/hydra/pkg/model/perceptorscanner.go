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

type PerceptorScannerConfigMap struct {
	HubHost         string
	HubPort         int
	HubUser         string
	HubUserPassword string
}

type PerceptorScanner struct {
	Image  string
	Port   int32
	CPU    resource.Quantity
	Memory resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         PerceptorScannerConfigMap

	ServiceName string

	PodName string

	ImagesMountName string
	ImagesMountPath string
}

func NewPerceptorScanner() *PerceptorScanner {
	memory, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}
	return &PerceptorScanner{
		Image:          "gcr.io/gke-verification/blackducksoftware/perceptor-scanner:master",
		Port:           3003,
		CPU:            cpu,
		Memory:         memory,
		ConfigMapName:  "perceptor-scanner-config",
		ConfigMapMount: "/etc/perceptor_scanner",
		ConfigMapPath:  "perceptor_scanner_conf.yaml",
		ServiceName:    "perceptor-scanner",

		// Must fill these out before use
		PodName: "",

		ImagesMountName: "",
		ImagesMountPath: "",
	}
}

func (psp *PerceptorScanner) Container() *v1.Container {
	return &v1.Container{
		Name:            "perceptor-scanner",
		Image:           psp.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: psp.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    psp.CPU,
				v1.ResourceMemory: psp.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      psp.ImagesMountName,
				MountPath: psp.ImagesMountPath,
			},
			v1.VolumeMount{
				Name:      psp.ConfigMapName,
				MountPath: psp.ConfigMapMount,
			},
		},
	}
}

func (psp *PerceptorScanner) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: psp.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: psp.ServiceName,
					Port: psp.Port,
				},
			},
			Selector: map[string]string{"name": psp.ServiceName}}}
}

func (psp *PerceptorScanner) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(psp.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(psp.ConfigMapName, psp.ConfigMapPath, string(jsonBytes))
}
