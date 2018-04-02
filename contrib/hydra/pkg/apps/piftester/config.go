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

package pif

import (
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"github.com/spf13/viper"
)

type Config struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	// perceivers config
	PerceptorPort             int32
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int

	// imagefacade config
	CreateImagesOnly bool
	ImageFacadePort  int32

	LogLevel string

	AuxConfig *AuxiliaryConfig
}

func ReadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	pc := &Config{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(pc)
	return pc
}

func (pc *Config) PodPerceiverConfig() model.PodPerceiverConfigMap {
	return model.PodPerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      4000,
	}
}

func (pc *Config) ImagePerceiverConfig() model.ImagePerceiverConfigMap {
	return model.ImagePerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      4000,
	}
}

func (pc *Config) ImagefacadeConfig() model.ImagefacadeConfigMap {
	return model.ImagefacadeConfigMap{
		DockerPassword:           pc.AuxConfig.DockerPassword,
		DockerUser:               pc.AuxConfig.DockerUsername,
		InternalDockerRegistries: pc.AuxConfig.InternalDockerRegistries,
		CreateImagesOnly:         pc.CreateImagesOnly,
		Port:                     pc.ImageFacadePort,
		LogLevel:                 pc.LogLevel,
	}
}

func (pc *Config) PifTesterConfig() model.PifTesterConfigMap {
	return model.PifTesterConfigMap{
		Port:            pc.PerceptorPort,
		ImageFacadePort: pc.ImageFacadePort,
	}
}
