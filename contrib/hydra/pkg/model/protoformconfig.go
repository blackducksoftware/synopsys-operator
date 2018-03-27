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
	"github.com/spf13/viper"
)

type ProtoformConfig struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	// perceptor config
	HubHost              string
	HubUser              string
	HubUserPassword      string
	HubPort              int32
	ConcurrentScanLimit  int
	PerceptorPort        int32
	UseMockPerceptorMode bool

	// Perceivers config
	ImagePerceiverPort        int32
	PodPerceiverPort          int32
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int

	// Scanner config
	ScannerReplicationCount int32
	ScannerPort             int32
	ImageFacadePort         int32
	JavaMaxHeapSizeMBs      int

	LogLevel string

	AuxConfig *AuxiliaryConfig
}

func ReadProtoformConfig(configPath string) *ProtoformConfig {
	viper.SetConfigFile(configPath)
	pc := &ProtoformConfig{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(pc)
	return pc
}

func (pc *ProtoformConfig) PodPerceiverConfig() PodPerceiverConfigMap {
	return PodPerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      pc.PodPerceiverPort,
	}
}

func (pc *ProtoformConfig) ImagePerceiverConfig() ImagePerceiverConfigMap {
	return ImagePerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      pc.ImagePerceiverPort,
	}
}

func (pc *ProtoformConfig) PerceptorScannerConfig() PerceptorScannerConfigMap {
	return PerceptorScannerConfigMap{
		HubHost:                 pc.HubHost,
		HubUser:                 pc.HubUser,
		HubUserPassword:         pc.HubUserPassword,
		HubPort:                 pc.HubPort,
		HubClientTimeoutSeconds: 60,
		JavaMaxHeapSizeMBs:      pc.JavaMaxHeapSizeMBs,
		LogLevel:                pc.LogLevel,
		Port:                    pc.ScannerPort,
		PerceptorHost:           "", // must be filled in elsewhere
		PerceptorPort:           pc.PerceptorPort,
		ImageFacadePort:         pc.ImageFacadePort,
	}
}

func (pc *ProtoformConfig) PerceptorImagefacadeConfig() PerceptorImagefacadeConfigMap {
	return PerceptorImagefacadeConfigMap{
		DockerPassword:           pc.AuxConfig.DockerPassword,
		DockerUser:               pc.AuxConfig.DockerUsername,
		InternalDockerRegistries: pc.AuxConfig.InternalDockerRegistries,
		CreateImagesOnly:         false,
		Port:                     pc.ImageFacadePort,
		LogLevel:                 pc.LogLevel,
	}
}

func (pc *ProtoformConfig) PerceptorConfig() PerceptorConfigMap {
	return PerceptorConfigMap{
		ConcurrentScanLimit: pc.ConcurrentScanLimit,
		HubHost:             pc.HubHost,
		HubUser:             pc.HubUser,
		HubUserPassword:     pc.HubUserPassword,
		UseMockMode:         pc.UseMockPerceptorMode,
		Port:                pc.PerceptorPort,
		LogLevel:            pc.LogLevel,
	}
}
