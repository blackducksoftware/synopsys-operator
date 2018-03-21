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

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scanner struct {
	PodName string

	ReplicaCount int32

	DockerSocketName string
	DockerSocketPath string

	ImagesMountName string
	ImagesMountPath string

	PerceptorScanner     *PerceptorScanner
	PerceptorImagefacade *PerceptorImagefacade
}

func NewScanner(perceptorScanner *PerceptorScanner, perceptorImagefacade *PerceptorImagefacade) *Scanner {
	scanner := &Scanner{
		PodName: "perceptor-scanner",

		ReplicaCount: 0,

		DockerSocketName: "dir-docker-socket",
		DockerSocketPath: "/var/run/docker.sock",

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",

		PerceptorScanner:     perceptorScanner,
		PerceptorImagefacade: perceptorImagefacade,
	}

	perceptorScanner.ImagesMountName = scanner.ImagesMountName
	perceptorScanner.ImagesMountPath = scanner.ImagesMountPath

	perceptorScanner.PodName = scanner.PodName

	perceptorImagefacade.ImagesMountName = scanner.ImagesMountName
	perceptorImagefacade.ImagesMountPath = scanner.ImagesMountPath

	perceptorImagefacade.DockerSocketName = scanner.DockerSocketName
	perceptorImagefacade.DockerSocketPath = scanner.DockerSocketPath

	perceptorImagefacade.PodName = scanner.PodName

	return scanner
}

func (sc *Scanner) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: sc.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &sc.ReplicaCount,
			Selector: map[string]string{"name": sc.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": sc.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: sc.PerceptorScanner.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sc.PerceptorScanner.ConfigMapName},
								},
							},
						},
						v1.Volume{
							Name: sc.PerceptorImagefacade.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sc.PerceptorImagefacade.ConfigMapName},
								},
							},
						},
						v1.Volume{
							Name:         sc.ImagesMountName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						v1.Volume{
							Name: sc.DockerSocketName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{Path: sc.DockerSocketPath},
							},
						},
					},
					Containers:         []v1.Container{*sc.PerceptorScanner.Container(), *sc.PerceptorImagefacade.Container()},
					ServiceAccountName: sc.PerceptorImagefacade.ServiceAccountName,
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}
