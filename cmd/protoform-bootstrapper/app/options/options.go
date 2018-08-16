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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/viper"

	"golang.org/x/crypto/ssh/terminal"
)

// BootstrapperOptions defines all the options that can
// be used to configure a bootstrapper
type BootstrapperOptions struct {
	LogLevel            string
	Interactive         bool
	Namespace           string
	DefaultCPU          string // Should be passed like: e.g. "300m"
	DefaultMem          string // Should be passed like: e.g "1300Mi"
	ClusterConfigFile   string
	DefaultImageVersion string
	DefaultRegistry     string
	DefaultImagePath    string

	// Perceptor
	AnnotateImages                        bool
	AnnotatePods                          bool
	AnnotationIntervalSeconds             int
	DumpIntervalMinutes                   int
	EnableMetrics                         bool
	EnableSkyfire                         bool
	HubClientTimeoutPerceptorMilliseconds int
	HubClientTimeoutScannerSeconds        int
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
	ConcurrentScanLimit                   int
	InternalDockerRegistries              []string
	DockerUsername                        string
	DockerPasswordOrToken                 string
	PerceptorNamespace                    string

	// Hub
	HubHost         string
	HubUser         string
	HubUserPassword string
	HubPort         int

	// Alert
	AlertEnabled      bool
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
	viper.SetDefault("Interactive", false)
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

// InteractiveConfig will prompt the user for information
func (o *BootstrapperOptions) InteractiveConfig() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Hub server host [%s]: ", o.HubHost)
	host, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("invalid hub server host: %v", err)
	}
	host = strings.TrimSpace(host)
	if len(host) > 0 {
		o.HubHost = host
	}

	fmt.Printf("Hub server port [%d]: ", o.HubPort)
	portStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("invalid hub server port: %v", err)
	}
	portStr = strings.TrimSpace(portStr)
	if len(portStr) > 0 {
		port, convErr := strconv.Atoi(portStr)
		if convErr != nil {
			return fmt.Errorf("hub server port isn't a number: %v", convErr)
		}
		o.HubPort = port
	}

	fmt.Printf("Hub user name [%s]: ", o.HubUser)
	user, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("invalid hub user name: %v", err)
	}
	user = strings.TrimSpace(user)
	if len(user) > 0 {
		o.HubUser = user
	}

	fmt.Printf("Hub user password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("unable to read password: %v", err)
	}
	o.HubUserPassword = string(bytePassword)
	fmt.Println()

	fmt.Printf("Maximum concurrent scans [%d]: ", o.ConcurrentScanLimit)
	limitStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("invalid maximum concurrent scans: %v", err)
	}
	limitStr = strings.TrimSpace(limitStr)
	if len(limitStr) > 0 {
		limit, convErr := strconv.Atoi(limitStr)
		if convErr != nil {
			return fmt.Errorf("maximum concurrent scans isn't a number: %v", convErr)
		}
		o.ConcurrentScanLimit = limit
	}

	return nil
}
