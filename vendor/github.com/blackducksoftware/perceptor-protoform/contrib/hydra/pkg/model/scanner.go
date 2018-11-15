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
	"fmt"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScannerHubConfig struct {
	User                 string
	PasswordEnvVar       string
	Port                 int32
	ClientTimeoutSeconds int
}

type ScannerImageFacadeConfig struct {
	Host string
	Port int32
}

type ScannerPerceptorConfig struct {
	Host string
	Port int32
}

type ScannerConfigMap struct {
	Hub         *ScannerHubConfig
	ImageFacade *ScannerImageFacadeConfig
	Perceptor   *ScannerPerceptorConfig

	JavaInitialHeapSizeMBs int
	JavaMaxHeapSizeMBs     int

	ImageDirectory string

	LogLevel string
	Port     int32
}

type Scanner struct {
	Image  string
	Memory resource.Quantity
	CPU    resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         ScannerConfigMap

	ServiceName string

	PodName string

	HubPasswordSecretName string
	HubPasswordSecretKey  string

	ImagesMountName string
}

func NewScanner(memoryString string, podName string, hubPasswordSecretName string, hubPasswordSecretKey string) *Scanner {
	memory, err := resource.ParseQuantity(memoryString)
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}

	return &Scanner{
		Image:          "gcr.io/gke-verification/blackducksoftware/perceptor-scanner:master",
		Memory:         memory,
		CPU:            cpu,
		ConfigMapName:  "perceptor-scanner-config",
		ConfigMapMount: "/etc/perceptor_scanner",
		ConfigMapPath:  "perceptor_scanner_conf.yaml",
		ServiceName:    "perceptor-scanner",

		PodName: podName,

		HubPasswordSecretName: hubPasswordSecretName,
		HubPasswordSecretKey:  hubPasswordSecretKey,

		// Must fill these out before use
		ImagesMountName: "",
	}
}

func (psp *Scanner) FullConfigMapPath() string {
	return fmt.Sprintf("%s/%s", psp.ConfigMapMount, psp.ConfigMapPath)
}

func (psp *Scanner) Container() *v1.Container {
	return &v1.Container{
		Name:            "perceptor-scanner",
		Image:           psp.Image,
		ImagePullPolicy: "Always",
		Command:         []string{"./perceptor-scanner", psp.FullConfigMapPath()},
		Env: []v1.EnvVar{
			v1.EnvVar{
				Name: psp.Config.Hub.PasswordEnvVar,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: psp.HubPasswordSecretName,
						},
						Key: psp.HubPasswordSecretKey,
					},
				},
			},
		},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: psp.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    psp.CPU,
				v1.ResourceMemory: psp.Memory,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    psp.CPU,
				v1.ResourceMemory: psp.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      psp.ImagesMountName,
				MountPath: psp.Config.ImageDirectory,
			},
			v1.VolumeMount{
				Name:      psp.ConfigMapName,
				MountPath: psp.ConfigMapMount,
			},
		},
	}
}

func (psp *Scanner) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: psp.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: psp.ServiceName,
					Port: psp.Config.Port,
				},
			},
			Selector: map[string]string{"name": psp.PodName}}}
}

func (psp *Scanner) ConfigMap() *v1.ConfigMap {
	return MakeConfigMap(psp.ConfigMapName, psp.ConfigMapPath, psp.Config)
}
