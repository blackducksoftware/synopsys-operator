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

type SkyfireConfigMap struct {
	UseInClusterConfig bool
	MasterURL          string
	KubeConfigPath     string
	LogLevel           string

	KubeDumpIntervalSeconds      int
	PerceptorDumpIntervalSeconds int
	HubDumpPauseSeconds          int

	Port int32

	HubHost               string
	HubUser               string
	HubUserPasswordEnvVar string

	PerceptorHost string
	PerceptorPort int32
}

func NewSkyfireConfigMap(logLevel string, port int32, hubHost string, hubUser string, hubUserPasswordEnvVar string, perceptorHost string, perceptorPort int32) *SkyfireConfigMap {
	return &SkyfireConfigMap{
		UseInClusterConfig:    true,
		LogLevel:              logLevel,
		Port:                  port,
		HubHost:               hubHost,
		HubUser:               hubUser,
		HubUserPasswordEnvVar: hubUserPasswordEnvVar,
		PerceptorHost:         perceptorHost,
		PerceptorPort:         perceptorPort,
	}
}

type Skyfire struct {
	PodName string
	Image   string
	CPU     resource.Quantity
	Memory  resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         SkyfireConfigMap

	HubPasswordSecretName string
	HubPasswordSecretKey  string

	ReplicaCount int32
	ServiceName  string
}

func NewSkyfire(hubPasswordSecretName string, hubPasswordSecretKey string) *Skyfire {
	memory, err := resource.ParseQuantity("512Mi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("100m")
	if err != nil {
		panic(err)
	}

	return &Skyfire{
		PodName:               "skyfire",
		Image:                 "gcr.io/gke-verification/blackducksoftware/skyfire:master",
		CPU:                   cpu,
		Memory:                memory,
		ConfigMapName:         "skyfire-config",
		ConfigMapMount:        "/etc/skyfire",
		ConfigMapPath:         "skyfire_conf.yaml",
		HubPasswordSecretName: hubPasswordSecretName,
		HubPasswordSecretKey:  hubPasswordSecretKey,
		ReplicaCount:          1,
		ServiceName:           "skyfire",
	}
}

func (sf *Skyfire) FullConfigMapPath() string {
	return fmt.Sprintf("%s/%s", sf.ConfigMapMount, sf.ConfigMapPath)
}

func (sf *Skyfire) Container() *v1.Container {
	return &v1.Container{
		Name:            "skyfire",
		Image:           sf.Image,
		ImagePullPolicy: "Always",
		Env: []v1.EnvVar{
			v1.EnvVar{
				Name: sf.Config.HubUserPasswordEnvVar,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: sf.HubPasswordSecretName,
						},
						Key: sf.HubPasswordSecretKey,
					},
				},
			},
		},
		Command: []string{"./skyfire", sf.FullConfigMapPath()},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: sf.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    sf.CPU,
				v1.ResourceMemory: sf.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      sf.ConfigMapName,
				MountPath: sf.ConfigMapMount,
			},
		},
	}
}

func (sf *Skyfire) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: sf.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &sf.ReplicaCount,
			Selector: map[string]string{"name": sf.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": sf.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: sf.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sf.ConfigMapName},
								},
							},
						},
					},
					Containers: []v1.Container{*sf.Container()},
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (sf *Skyfire) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: sf.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: sf.ServiceName,
					Port: sf.Config.Port,
				},
			},
			Selector: map[string]string{"name": sf.ServiceName}}}
}

func (sf *Skyfire) ConfigMap() *v1.ConfigMap {
	return MakeConfigMap(sf.ConfigMapName, sf.ConfigMapPath, sf.Config)
}
