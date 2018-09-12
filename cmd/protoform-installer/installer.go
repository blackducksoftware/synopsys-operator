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

package main

import (
	"fmt"
	"os"

	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/opssight"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
)

func main() {
	configPath := os.Args[1]
	fmt.Printf("Config path: %s", configPath)
	runProtoform(configPath)
}

func runProtoform(configPath string) {
	installer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	alertConfig, err := alert.NewController(&alert.ProtoformControllerConfig{
		Config:        installer.Config,
		KubeConfig:    installer.KubeConfig,
		KubeClientSet: installer.KubeClientSet,
		Threadiness:   installer.Config.Threadiness,
		StopCh:        stopCh,
	})
	installer.AddController(alertConfig)

	hubConfig, err := hub.NewController(&hub.ProtoformControllerConfig{
		Config:        installer.Config,
		KubeConfig:    installer.KubeConfig,
		KubeClientSet: installer.KubeClientSet,
		Threadiness:   installer.Config.Threadiness,
		StopCh:        stopCh,
	})
	installer.AddController(hubConfig)

	opssSightConfig, err := opssight.NewController(&opssight.ProtoformControllerConfig{
		Config:        installer.Config,
		KubeConfig:    installer.KubeConfig,
		KubeClientSet: installer.KubeClientSet,
		Threadiness:   installer.Config.Threadiness,
		StopCh:        stopCh,
	})
	installer.AddController(opssSightConfig)

	installer.Deploy()

	<-stopCh

}
