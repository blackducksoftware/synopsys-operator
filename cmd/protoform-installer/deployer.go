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

	alertv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	opssightv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
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
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})

	alertController, err := alert.NewController(&alert.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetAlertDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(alertController)

	hubController, err := hub.NewController(&hub.ProtoformConfig{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetHubDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(hubController)

	opssSightController, err := opssight.NewController(&opssight.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetOpsSightDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(opssSightController)

	deployer.Deploy()

	<-stopCh
}

// GetAlertDefaultValue creates a alert crd configuration object with defaults
func GetAlertDefaultValue() *alertv1.AlertSpec {
	port := 8443
	hubPort := 443
	standAlone := true

	return &alertv1.AlertSpec{
		Port:           &port,
		HubPort:        &hubPort,
		StandAlone:     &standAlone,
		AlertMemory:    "512M",
		CfsslMemory:    "640M",
		AlertImageName: "blackduck-alert",
		CfsslImageName: "hub-cfssl",
	}
}

// GetHubDefaultValue creates a hub crd configuration object with defaults
func GetHubDefaultValue() *hubv1.HubSpec {
	return &hubv1.HubSpec{
		Flavor:          "small",
		DockerRegistry:  "docker.io",
		DockerRepo:      "blackducksoftware",
		HubVersion:      "5.0.0",
		DbPrototype:     "empty",
		CertificateName: "default",
		HubType:         "worker",
		Environs:        []hubv1.Environs{},
		ImagePrefix:     "hub",
	}
}

// GetOpsSightDefaultValue creates a perceptor crd configuration object with defaults
func GetOpsSightDefaultValue() *opssightv1.OpsSightSpec {
	defaultAnnotationInterval := 30
	defaultDumpInterval := 30
	defaultCheckForStalledScansPauseHours := 999999
	defaultStalledScanClientTimeoutHours := 999999
	defaultModelMetricsPauseSeconds := 15
	defaultUnknownImagePauseMilliseconds := 15000

	return &opssightv1.OpsSightSpec{
		PerceptorPort:             3001,
		PerceiverPort:             3002,
		ScannerPort:               3003,
		ImageFacadePort:           3004,
		SkyfirePort:               3005,
		InternalRegistries:        []opssightv1.RegistryAuth{},
		AnnotationIntervalSeconds: &defaultAnnotationInterval,
		DumpIntervalMinutes:       &defaultDumpInterval,
		Hub: &opssightv1.HubSpec{
			User:                               "sysadmin",
			Port:                               443,
			ClientTimeoutPerceptorMilliseconds: 100000,
			ClientTimeoutScannerSeconds:        600,
			ConcurrentScanLimit:                2,
			TotalScanLimit:                     1000,
			PasswordEnvVar:                     "PCP_HUBUSERPASSWORD",
			Password:                           "blackduck",
		},
		CheckForStalledScansPauseHours: &defaultCheckForStalledScansPauseHours,
		StalledScanClientTimeoutHours:  &defaultStalledScanClientTimeoutHours,
		ModelMetricsPauseSeconds:       &defaultModelMetricsPauseSeconds,
		UnknownImagePauseMilliseconds:  &defaultUnknownImagePauseMilliseconds,
		PerceptorImage:                 "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
		ScannerImage:                   "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
		ImagePerceiverImage:            "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
		PodPerceiverImage:              "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
		ImageFacadeImage:               "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
		SkyfireImage:                   "gcr.io/saas-hub-stg/blackducksoftware/skyfire:master",
		PodPerceiver:                   true,
		ImagePerceiver:                 false,
		Metrics:                        true,
		PerceptorSkyfire:               false,
		DefaultCPU:                     "300m",
		DefaultMem:                     "1300Mi",
		LogLevel:                       "debug",
		SecretName:                     "perceptor",
		ScannerReplicaCount:            1,
		ServiceAccounts: map[string]string{
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner",
			"skyfire":                "skyfire",
		},
		Names: &opssightv1.ResourceNames{
			PodPerceiver:   "pod-perceiver",
			ImagePerceiver: "image-perceiver",
			Perceptor:      "perceptor",
			Skyfire:        "skyfire",
			Scanner:        "perceptor-scanner",
			ImageFacade:    "perceptor-imagefacade",
		},
	}
}
