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

package scannertester

import (
	"k8s.io/api/core/v1"

	model "github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScannerTester struct {
	PodName string

	ReplicaCount int32

	ImagesMountName string
	ImagesMountPath string

	PerceptorScanner *model.PerceptorScanner
	Imagefacade      *model.MockImagefacade
}

func NewScannerTester(perceptorScanner *model.PerceptorScanner, imagefacade *model.MockImagefacade) *ScannerTester {
	scanner := &ScannerTester{
		PodName: "scanner-tester",

		ReplicaCount: 1,

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",

		PerceptorScanner: perceptorScanner,
		Imagefacade:      imagefacade,
	}

	perceptorScanner.ImagesMountName = scanner.ImagesMountName
	perceptorScanner.ImagesMountPath = scanner.ImagesMountPath

	perceptorScanner.PodName = scanner.PodName

	imagefacade.ImagesMountName = scanner.ImagesMountName
	imagefacade.ImagesMountPath = scanner.ImagesMountPath

	imagefacade.PodName = scanner.PodName

	return scanner
}

func (sc *ScannerTester) ReplicationController() *v1.ReplicationController {
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
							Name: sc.Imagefacade.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sc.Imagefacade.ConfigMapName},
								},
							},
						},
						v1.Volume{
							Name:         sc.ImagesMountName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
					},
					Containers: []v1.Container{*sc.PerceptorScanner.Container(), *sc.Imagefacade.Container()},
				}}}}
}
