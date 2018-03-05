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

type PerceptorScannerConfigMap struct {
	HubHost         string
	HubPort         int
	HubUser         string
	HubUserPassword string
}

type PerceptorImagefacadeConfigMap struct {
	Dockerusername string
	Dockerpassword string
}

type PerceptorScanner struct {
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

	ImagesMountName string
	ImagesMountPath string
}

func NewPerceptorScanner(serviceAccountName string) *PerceptorScanner {
	defaultMem, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}
	return &PerceptorScanner{
		PodName:               "perceptor-scanner",
		ScannerImage:          "gcr.io/gke-verification/blackducksoftware/perceptor-scanner:master",
		ScannerPort:           3003,
		ScannerCPU:            defaultCPU,
		ScannerMemory:         defaultMem,
		ScannerConfigMapName:  "perceptor-scanner-config",
		ScannerConfigMapMount: "/etc/perceptor_scanner",
		ScannerServiceName:    "perceptor-scanner",
		ScannerReplicaCount:   2,

		ImageFacadeImage:              "gcr.io/gke-verification/blackducksoftware/perceptor-imagefacade:master",
		ImageFacadePort:               3004,
		ImageFacadeCPU:                defaultCPU,
		ImageFacadeMemory:             defaultMem,
		ImageFacadeConfigMapName:      "perceptor-imagefacade-config",
		ImageFacadeConfigMapMount:     "/etc/perceptor_imagefacade",
		ImageFacadeServiceAccountName: serviceAccountName,
		ImageFacadeServiceName:        "perceptor-imagefacade",

		DockerSocketName: "dir-docker-socket",
		DockerSocketPath: "/var/run/docker.sock",

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",
	}
}

func (psp *PerceptorScanner) scannerContainer() *v1.Container {
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
				Name:      psp.ImagesMountName,
				MountPath: psp.ImagesMountPath,
			},
			v1.VolumeMount{
				Name:      psp.ScannerConfigMapName,
				MountPath: psp.ScannerConfigMapMount,
			},
		},
	}
}

func (psp *PerceptorScanner) imageFacadeContainer() *v1.Container {
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
				Name:      psp.ImagesMountName,
				MountPath: psp.ImagesMountPath,
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

func (psp *PerceptorScanner) ReplicationController() *v1.ReplicationController {
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
							Name:         psp.ImagesMountName,
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

func (psp *PerceptorScanner) ScannerService() *v1.Service {
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

func (psp *PerceptorScanner) ImageFacadeService() *v1.Service {
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

// just for testing:

func (psp *PerceptorScanner) ImageFacadeReplicationController() *v1.ReplicationController {
	replicaCount := int32(1)
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: psp.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &replicaCount,
			Selector: map[string]string{"name": "perceptor-imagefacade"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": "perceptor-imagefacade"}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: psp.ImageFacadeConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: psp.ImageFacadeConfigMapName},
								},
							},
						},
						v1.Volume{
							Name:         psp.ImagesMountName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						v1.Volume{
							Name: psp.DockerSocketName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{Path: psp.DockerSocketPath},
							},
						},
					},
					Containers:         []v1.Container{*psp.imageFacadeContainer()},
					ServiceAccountName: psp.ImageFacadeServiceAccountName,
				}}}}
}
