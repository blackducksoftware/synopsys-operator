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
)

type ProtoformConfig struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	// perceptor config
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	HubHost                   string
	HubUser                   string
	HubUserPassword           string
	HubPort                   int
	ConcurrentScanLimit       int

	UseMockPerceptorMode bool

	AuxConfig *AuxiliaryConfig
}

func (pc *ProtoformConfig) PerceptorConfig() string {
	jsonBytes, err := json.Marshal(PerceptorConfig{
		ConcurrentScanLimit: pc.ConcurrentScanLimit,
		HubHost:             pc.HubHost,
		HubUser:             pc.HubUser,
		HubUserPassword:     pc.HubUserPassword,
		UseMockMode:         pc.UseMockPerceptorMode,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PodPerceiverConfig() string {
	jsonBytes, err := json.Marshal(PodPerceiverConfig{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             pc.PerceptorHost,
		PerceptorPort:             pc.PerceptorPort,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) ImagePerceiverConfig() string {
	jsonBytes, err := json.Marshal(ImagePerceiverConfig{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             pc.PerceptorHost,
		PerceptorPort:             pc.PerceptorPort,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PerceptorScannerConfig() string {
	jsonBytes, err := json.Marshal(PerceptorScannerConfig{
		HubHost:         pc.HubHost,
		HubPort:         pc.HubPort,
		HubUser:         pc.HubUser,
		HubUserPassword: pc.HubUserPassword,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PerceptorImagefacadeConfig() string {
	jsonBytes, err := json.Marshal(PerceptorImagefacadeConfig{
		Dockerpassword: pc.AuxConfig.DockerPassword,
		Dockerusername: pc.AuxConfig.DockerUsername,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) ToConfigMaps() []*v1.ConfigMap {
	return []*v1.ConfigMap{
		MakeConfigMap("perceptor-scanner-config", "perceptor_scanner_conf.yaml", pc.PerceptorScannerConfig()),
		MakeConfigMap("kube-generic-perceiver-config", "perceiver.yaml", pc.PodPerceiverConfig()),
		MakeConfigMap("perceptor-config", "perceptor_conf.yaml", pc.PerceptorConfig()),
		MakeConfigMap("openshift-perceiver-config", "perceiver.yaml", pc.ImagePerceiverConfig()),
		MakeConfigMap("perceptor-imagefacade-config", "perceptor_imagefacade_conf.yaml", pc.PerceptorImagefacadeConfig()),
	}
}
