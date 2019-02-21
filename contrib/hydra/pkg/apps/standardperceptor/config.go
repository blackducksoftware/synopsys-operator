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

package standardperceptor

import (
	model "github.com/blackducksoftware/synopsys-operator/contrib/hydra/pkg/model"
	"github.com/spf13/viper"
)

type Config struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	Hub struct {
		Host     string
		User     string
		Password string
		Port     int32
	}

	Perceptor struct {
		HubClientTimeoutMilliseconds int
		ConcurrentScanLimit          int
		TotalScanLimit               int
		ServiceName                  string
		Port                         int32
		UseMockMode                  bool
	}

	// Perceivers config
	ImagePerceiver struct {
		Port                      int32
		AnnotationIntervalSeconds int
		DumpIntervalMinutes       int
	}

	PodPerceiver struct {
		Port                      int32
		ReplicationCount          int
		AnnotationIntervalSeconds int
		DumpIntervalMinutes       int
	}

	ScannerPod struct {
		Name             string
		ReplicationCount int32
		ImageDirectory   string
	}

	Scanner struct {
		Port                   int32
		Memory                 string
		JavaInitialHeapSizeMBs int
		JavaMaxHeapSizeMBs     int
	}

	ImageFacade struct {
		Port int32
	}

	Skyfire struct {
		Port int32
	}

	// Secret config
	HubPasswordSecretName string
	HubPasswordSecretKey  string

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
		AnnotationIntervalSeconds: pc.PodPerceiver.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.PodPerceiver.DumpIntervalMinutes,
		PerceptorHost:             pc.Perceptor.ServiceName,
		PerceptorPort:             pc.Perceptor.Port,
		Port:                      pc.PodPerceiver.Port,
	}
}

func (pc *Config) ImagePerceiverConfig() model.ImagePerceiverConfigMap {
	return model.ImagePerceiverConfigMap{
		AnnotationIntervalSeconds: pc.ImagePerceiver.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.ImagePerceiver.DumpIntervalMinutes,
		PerceptorHost:             pc.Perceptor.ServiceName,
		PerceptorPort:             pc.Perceptor.Port,
		Port:                      pc.ImagePerceiver.Port,
	}
}

func (pc *Config) ScannerConfig() model.ScannerConfigMap {
	return model.ScannerConfigMap{
		Hub: &model.ScannerHubConfig{
			User:                 pc.Hub.User,
			PasswordEnvVar:       "SCANNER_HUBUSERPASSWORD",
			Port:                 pc.Hub.Port,
			ClientTimeoutSeconds: 600,
		},
		JavaInitialHeapSizeMBs: pc.Scanner.JavaInitialHeapSizeMBs,
		JavaMaxHeapSizeMBs:     pc.Scanner.JavaMaxHeapSizeMBs,
		LogLevel:               pc.LogLevel,
		Port:                   pc.Scanner.Port,
		Perceptor: &model.ScannerPerceptorConfig{
			Host: pc.Perceptor.ServiceName,
			Port: pc.Perceptor.Port,
		},
		ImageFacade: &model.ScannerImageFacadeConfig{
			Host: "localhost",
			Port: pc.ImageFacade.Port,
		},
		ImageDirectory: pc.ScannerPod.ImageDirectory,
	}
}

func (pc *Config) ImagefacadeConfig() model.ImagefacadeConfigMap {
	return model.ImagefacadeConfigMap{
		PrivateDockerRegistries: pc.AuxConfig.PrivateDockerRegistries,
		CreateImagesOnly:        false,
		Port:                    pc.ImageFacade.Port,
		LogLevel:                pc.LogLevel,
		ImageDirectory:          pc.ScannerPod.ImageDirectory,
	}
}

func (pc *Config) PerceptorConfig() model.PerceptorConfigMap {
	return model.PerceptorConfigMap{
		Hub: &model.PerceptorHubConfig{
			Host:                      pc.Hub.Host,
			User:                      pc.Hub.User,
			ClientTimeoutMilliseconds: pc.Perceptor.HubClientTimeoutMilliseconds,
			PasswordEnvVar:            "PERCEPTOR_HUBUSERPASSWORD",
			Port:                      int(pc.Hub.Port),
			ConcurrentScanLimit:       pc.Perceptor.ConcurrentScanLimit,
			TotalScanLimit:            pc.Perceptor.TotalScanLimit,
		},
		Timings: &model.PerceptorTimingsConfig{
			CheckForStalledScansPauseHours:  2,
			ModelMetricsPauseSeconds:        15,
			PruneOrphanedImagesPauseMinutes: 9999999,
			StalledScanClientTimeoutHours:   2,
			UnknownImagePauseMilliseconds:   15000,
		},
		UseMockMode: pc.Perceptor.UseMockMode,
		Port:        pc.Perceptor.Port,
		LogLevel:    pc.LogLevel,
	}
}

func (pc *Config) SkyfireConfig() model.SkyfireConfigMap {
	return model.SkyfireConfigMap{
		HubHost:               pc.Hub.Host,
		HubUser:               pc.Hub.User,
		HubUserPasswordEnvVar: "PERCEPTOR_HUBUSERPASSWORD",
		// TODO pc.HubPort ?
		KubeDumpIntervalSeconds:      15,
		PerceptorDumpIntervalSeconds: 15,
		HubDumpPauseSeconds:          30,
		LogLevel:                     pc.LogLevel,
		PerceptorHost:                pc.Perceptor.ServiceName,
		PerceptorPort:                pc.Perceptor.Port,
		Port:                         pc.Skyfire.Port,
		UseInClusterConfig:           true,
	}
}
