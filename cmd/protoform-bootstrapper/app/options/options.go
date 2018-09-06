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

package options

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
)

// BootstrapperOptions defines all the options that can
// be used to configure a bootstrapper
type BootstrapperOptions struct {
	LogLevel            string
	Namespace           string
	DefaultCPU          string // Should be passed like: e.g. "300m"
	DefaultMem          string // Should be passed like: e.g "1300Mi"
	ClusterConfigFile   string
	DefaultImageVersion string
	DefaultRegistry     string
	DefaultImagePath    string

	// Perceptor
	AnnotateImages                        *bool
	AnnotatePods                          *bool
	AnnotationIntervalSeconds             *int
	DumpIntervalMinutes                   *int
	EnableMetrics                         *bool
	EnableSkyfire                         *bool
	HubClientTimeoutPerceptorMilliseconds *int
	HubClientTimeoutScannerSeconds        *int
	PerceptorImage                        string
	ScannerImage                          string
	ImagePerceiverImage                   string
	PodPerceiverImage                     string
	ImageFacadeImage                      string
	SkyfireImage                          string
	ProtoformImage                        string
	PerceptorImageVersion                 string
	ScannerImageVersion                   string
	PerceiverImageVersion                 string
	ImageFacadeImageVersion               string
	SkyfireImageVersion                   string
	ProtoformImageVersion                 string
	ConcurrentScanLimit                   *int
	InternalDockerRegistries              []string
	DockerUsername                        string
	DockerPasswordOrToken                 string
	PerceptorNamespace                    string

	// Hub
	HubHost         string
	HubUser         string
	HubUserPassword string
	HubPort         *int

	// Alert
	AlertEnabled      *bool
	AlertRegistry     string
	AlertImagePath    string
	AlertImageName    string
	AlertImageVersion string
	CfsslImageName    string
	CfsslImageVersion string
	AlertNamespace    string
}

// NewBootstrapperOptions creates a BootstrapperOptions object
// and sets configuation defaults
func NewBootstrapperOptions() *BootstrapperOptions {
	viper.SetDefault("AnnotatePods", false)
	viper.SetDefault("AnnotateImages", false)
	viper.SetDefault("EnableMetrics", true)
	viper.SetDefault("ClusterConfigFile", "$HOME/.kube/config")
	viper.SetDefault("Namespace", "protoform")
	viper.SetDefault("HubPort", 443)
	viper.SetDefault("HubHost", "webserver")
	viper.SetDefault("ConcurrentScanLimit", 7)
	viper.SetDefault("HubClientTimeoutPerceptorMilliseconds", 5000)
	viper.SetDefault("HubClientTimeoutScannerSeconds", 30)
	viper.SetDefault("ProtoformImage", "perceptor-protoform")
	viper.SetDefault("ProtoformImageVersion", "master")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("AlertEnabled", false)
	return &BootstrapperOptions{}
}

// ReadConfig will read the configuration file provided
func (o *BootstrapperOptions) ReadConfig(conf string) error {
	viper.SetConfigFile(conf)
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Failed to read option file: %v", err)
	}

	err = viper.Unmarshal(&o)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal options: %v", err)
	}
	return nil
}

// MergeOptions will merge 2 BootstrapperOptions
func (o *BootstrapperOptions) MergeOptions(new *BootstrapperOptions) error {
	return util.MergeConfig(new, o)
}
