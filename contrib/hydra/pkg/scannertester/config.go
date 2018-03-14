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
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"github.com/spf13/viper"
)

type Config struct {
	MasterURL      string
	KubeConfigPath string

	HubHost         string
	HubUser         string
	HubUserPassword string

	PerceptorPort   int32
	ImageFacadePort int32
	ScannerPort     int32

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

func (config *Config) PerceptorScannerConfig() model.PerceptorScannerConfigMap {
	return model.PerceptorScannerConfigMap{
		HubClientTimeoutSeconds: 60,
		HubHost:                 config.HubHost,
		HubUser:                 config.HubUser,
		HubUserPassword:         config.HubUserPassword,
		ImageFacadePort:         config.ImageFacadePort,
		PerceptorPort:           config.PerceptorPort,
		Port:                    config.ScannerPort,
	}
}

func (pc *Config) MockImagefacadeConfig() model.MockImagefacadeConfigMap {
	return model.MockImagefacadeConfigMap{
		Port: pc.ImageFacadePort,
	}
}

func (pc *Config) PerceptorConfig() model.PerceptorConfigMap {
	return model.PerceptorConfigMap{
		Port:        pc.PerceptorPort,
		UseMockMode: true,
	}
}
