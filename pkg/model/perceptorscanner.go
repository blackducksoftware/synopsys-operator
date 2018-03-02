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
	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PerceptorScannerPod struct {
	PodName               string
	ScannerImage          string
	ScannerPort           int32
	ScannerCPU            resource.Quantity
	ScannerMemory         resource.Quantity
	ScannerConfigMapName  string
	ScannerConfigMapMount string
	ScannerServiceName    string
	ScannerReplicaCount   int32

	ImageFacadeImage              string
	ImageFacadePort               int32
	ImageFacadeCPU                resource.Quantity
	ImageFacadeMemory             resource.Quantity
	ImageFacadeConfigMapName      string
	ImageFacadeConfigMapMount     string
	ImageFacadeServiceAccountName string
	ImageFacadeServiceName        string

	DockerSocketName string
	DockerSocketPath string
}

func NewPerceptorScannerPod(serviceAccountName string) *PerceptorScannerPod {
	defaultMem, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}
	return &PerceptorScannerPod{
		PodName:               "perceptor-scanner",
		ScannerImage:          "gcr.io/gke-verification/blackducksoftware/perceptor-scanner:latest",
		ScannerPort:           3003,
		ScannerCPU:            defaultCPU,
		ScannerMemory:         defaultMem,
		ScannerConfigMapName:  "perceptor-scanner-config",
		ScannerConfigMapMount: "/etc/perceptor_scanner",
		ScannerServiceName:    "perceptor-scanner",
		ScannerReplicaCount:   2,

		ImageFacadeImage:              "gcr.io/gke-verification/blackducksoftware/perceptor-imagefacade:latest",
		ImageFacadePort:               4000,
		ImageFacadeCPU:                defaultCPU,
		ImageFacadeMemory:             defaultMem,
		ImageFacadeConfigMapName:      "perceptor-imagefacade-config",
		ImageFacadeConfigMapMount:     "/etc/perceptor_imagefacade",
		ImageFacadeServiceAccountName: serviceAccountName,
		ImageFacadeServiceName:        "perceptor-imagefacade",

		DockerSocketName: "dir-docker-socket",
		DockerSocketPath: "/var/run/docker.sock",
	}
}

func (psp *PerceptorScannerPod) scannerContainer() *v1.Container {
	return &v1.Container{
		Name:            "perceptor-scanner",
		Image:           psp.ScannerImage,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: psp.ScannerPort,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    psp.ScannerCPU,
				v1.ResourceMemory: psp.ScannerMemory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "var-images",
				MountPath: "/var/images",
			},
			v1.VolumeMount{
				Name:      psp.ScannerConfigMapName,
				MountPath: psp.ScannerConfigMapMount,
			},
		},
	}
}

func (psp *PerceptorScannerPod) imageFacadeContainer() *v1.Container {
	privileged := true
	return &v1.Container{
		Name:            "perceptor-imagefacade",
		Image:           psp.ImageFacadeImage,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: psp.ImageFacadePort,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    psp.ImageFacadeCPU,
				v1.ResourceMemory: psp.ImageFacadeMemory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "var-images",
				MountPath: "/var/images",
			},
			v1.VolumeMount{
				Name:      psp.ImageFacadeConfigMapName,
				MountPath: psp.ImageFacadeConfigMapMount,
			},
			v1.VolumeMount{
				Name:      psp.DockerSocketName,
				MountPath: psp.DockerSocketPath,
			},
		},
		SecurityContext: &v1.SecurityContext{Privileged: &privileged},
	}
}

func (psp *PerceptorScannerPod) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: psp.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &psp.ScannerReplicaCount,
			Selector: map[string]string{"name": psp.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": psp.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: psp.ScannerConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: psp.ScannerConfigMapName},
								},
							},
						},
						v1.Volume{
							Name: psp.ImageFacadeConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: psp.ImageFacadeConfigMapName},
								},
							},
						},
						v1.Volume{
							Name: psp.ImageFacadeConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: psp.ImageFacadeConfigMapName},
								},
							},
						},
						v1.Volume{
							Name:         "var-images",
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						v1.Volume{
							Name: psp.DockerSocketName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{Path: psp.DockerSocketPath},
							},
						},
					},
					Containers:         []v1.Container{*psp.scannerContainer(), *psp.imageFacadeContainer()},
					ServiceAccountName: psp.ImageFacadeServiceAccountName,
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (psp *PerceptorScannerPod) ScannerService() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: psp.ScannerServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: psp.ScannerServiceName,
					Port: psp.ScannerPort,
				},
			},
			Selector: map[string]string{"name": psp.ScannerServiceName}}}
}

func (psp *PerceptorScannerPod) ImageFacadeService() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: psp.ImageFacadeServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: psp.ImageFacadeServiceName,
					Port: psp.ImageFacadePort,
				},
			},
			Selector: map[string]string{"name": psp.ImageFacadeServiceName}}}
}
