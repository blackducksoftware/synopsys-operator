/*
Copyright (C) 2019 Synopsys, Inc.

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

package util

import (
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	//samplev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
)

// GetSampleDefaultValue creates a sample crd configuration object with defaults
//func GetSampleDefaultValue() *samplev1.SampleSpec {
//	return &samplev1.SampleSpec{
//		Namespace:   "namesapce",
//		SampleValue: "Value",
//	}
//}

// GetBlackDuckTemplate returns the required fields for Black Duck
func GetBlackDuckTemplate() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Size:              "Small",
		DbPrototype:       "",
		CertificateName:   "default",
		Type:              "worker",
		Version:           "2019.4.0",
		LicenseKey:        "",
		PersistentStorage: false,
	}
}

// GetBlackDuckDefaultPersistentStorageLatest creates a Black Duck crd configuration object
// with defaults and persistent storage
func GetBlackDuckDefaultPersistentStorageLatest() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-pvc",
		LicenseKey:        "",
		CertificateName:   "default",
		Version:           "2019.4.0",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVCStorageClass:   "",
		Environs:          []string{},
		ImageRegistries:   []string{},
		PVC: []blackduckv1.PVC{
			{
				Name: "blackduck-postgres",
				Size: "150Gi",
			},
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
			{
				Name: "blackduck-uploadcache-data",
				Size: "100Gi",
			},
			{
				Name: "blackduck-uploadcache-key",
				Size: "2Gi",
			},
		},
	}
}

// GetBlackDuckDefaultExternalPersistentStorageLatest creates a BlackDuck crd configuration object
// with defaults and external persistent storage for latest BlackDuck
func GetBlackDuckDefaultExternalPersistentStorageLatest() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Version:           "2019.4.0",
		Size:              "small",
		PVCStorageClass:   "",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC: []blackduckv1.PVC{
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
			{
				Name: "blackduck-uploadcache-data",
				Size: "100Gi",
			},
			{
				Name: "blackduck-uploadcache-key",
				Size: "2Gi",
			},
		},
		CertificateName: "default",
		Type:            "Artifacts",
		Environs:        []string{},
		ImageRegistries: []string{},
		LicenseKey:      "",
	}
}

// GetBlackDuckDefaultPersistentStorageV1 creates a BlackDuck crd configuration object
// with defaults and persistent storage for V1 BlackDuck
func GetBlackDuckDefaultPersistentStorageV1() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Version:           "2019.2.2",
		Size:              "small",
		PVCStorageClass:   "",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC: []blackduckv1.PVC{
			{
				Name: "blackduck-postgres",
				Size: "150Gi",
			},
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
		},
		CertificateName: "default",
		Type:            "Artifacts",
		Environs:        []string{},
		ImageRegistries: []string{},
		LicenseKey:      "",
	}
}

// GetBlackDuckDefaultExternalPersistentStorageV1 creates a BlackDuck crd configuration object
// with defaults and external persistent storage for V1 BlackDuck
func GetBlackDuckDefaultExternalPersistentStorageV1() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Version:           "2019.2.2",
		Size:              "small",
		PVCStorageClass:   "",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC: []blackduckv1.PVC{
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
		},
		Type: "Artifacts",
	}
}

// GetBlackDuckDefaultBDBA returns a BlackDuck with BDBA
func GetBlackDuckDefaultBDBA() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:       "blackduck-bdba",
		LicenseKey:      "",
		CertificateName: "default",
		Version:         "2019.4.0",
		Environs: []string{
			"USE_BINARY_UPLOADS:1",
		},
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "small",
	}
}

// GetBlackDuckDefaultEphemeral returns a BlackDuck with ephemeral storage
func GetBlackDuckDefaultEphemeral() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-ephemeral",
		LicenseKey:        "",
		CertificateName:   "default",
		Version:           "2019.4.0",
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "small",
		Type:              "worker",
	}
}

// GetBlackDuckDefaultEphemeralCustomAuthCA returns a BlackDuck with ephemeral storage
// using custom auth CA
func GetBlackDuckDefaultEphemeralCustomAuthCA() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-auth-ca",
		LicenseKey:        "",
		CertificateName:   "default",
		Version:           "2019.4.0",
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "Small",
		AuthCustomCA:      "-----BEGIN CERTIFICATE-----\r\nMIIE1DCCArwCCQCuw9TgaoBKVDANBgkqhkiG9w0BAQsFADAsMQswCQYDVQQGEwJV\r\nUzELMAkGA1UECgwCYmQxEDAOBgNVBAMMB1JPT1QgQ0EwHhcNMTkwMjA2MDAzMjM3\r\nWhcNMjExMTI2MDAzMjM3WjAsMQswCQYDVQQGEwJVUzELMAkGA1UECgwCYmQxEDAO\r\nBgNVBAMMB1JPT1QgQ0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCr\r\nIctvPVoqRS3Ti38uFRVfJDovyi0p9PIaOmja3tMvkfecCsCVYHMo/vAy/fm9qiJI\r\nKutTwX9aLuiLO0tsDDUNwv0CrbXvuHpWvASOAdKyl6uxiYl0fq0cyBZSdKlsdDGk\r\nivENpN2gKHxDSUgAo74wUskfBrKvfKLhJhOmKCbN/NvxlsGMM5DgPgFGNegmw5r0\r\nZlDTXlWn3J/8C80dfGjT5hLr6Jtl0KTqxSREVTLT0fDk7bt9BHH/TCtNs9UwR1UI\r\nJVjjzW6pgS1DmGZ7Mfg2WBhhdDBuN0gxk/bcoiV2tfI0MLQyeVP+qWmdUXSNn9CT\r\nmpYdKezMfi5ieSy40fy23n+D1C+Xm5pnFErm3BwZYdN9gI633IBPQa0ELo28ZxhI\r\nIclGGyhUubZJ+ybNvGOIrgypTXYrZqvyWMV3qiMZb1EzpKdqAzGfsN1zmF+o4Rc3\r\ntBa2EF/lNSVCClUeFBA2UXvD/K9QA84cbLNJwpBZ9Bc6CZyvRTYGzXtAuZUVvNju\r\nMcWhsqXWzhVkChTyYicOdT8ZB+7/eC3tFyjAKSszIA5xuO8NtuIZBAc2AzRrkoE5\r\nCgHEUxNA3tbRUjYnH5HcgaQveFQtFwBWqIMxPeJixSLk2KYJSsWpTPC1x6s1IBLO\r\nITWhedDbtbs/FT9+cXd9K+/L+6UgR31oHaY/hYai1QIDAQABMA0GCSqGSIb3DQEB\r\nCwUAA4ICAQAz7aK5m9yPE/tTFQJfZRr35ug8ikBuGFvzb5s3fWYlQ1QbKUPBp9Q/\r\n1kUGJF2niOULUp5Gig6urz+E1m3wE5jgYRwZjgTmoEQEmN0/VQWTus72isWhTsZ5\r\nJKDSzcKGRJnHzO91gA3ZP1Cxoin5GX6w8eqEA2vh1hc7+GyKPTOsxu8hYMYI1yId\r\nfWAjqEUobLZZoijf+c3AqBVcf4tOpFMRTy4au3H+v7TNjc/fAeZUeAz7BswfqEV9\r\n0QNNTpezq5IS+pSPShRatL9k/BaE3MaF0Ossfnv3UPV80Yrup+9pRV8Lu6EXrdg5\r\n3L2+KK2Nz9A+iF2u9VqUw9lcJCIjgY+APf6Tf2AKQxNCA/pV1z0I8aQAlSLolgpx\r\nSMLwMecpjAcHPWF5ut3Re+8PfeyLGzeXCVyhZc9Aj9KaTNLRa/kb21KNVbcGGTu/\r\nuiGMEJXq1a1fKzMKTPnARz70XCS7nLJ7qEK3TuvrMhCqEEdFUf/S4yAmmWaEO9Fr\r\nUBk9ACW9UYBFtowqbJkbJm3KEXMMFP5cs33j/HEA1IkKDVT9Hi7NEK2/Y7e9afv7\r\no1UGNrGgU1rK8K+/2htOH9JhlPFWHQkk+wvGL6fFI7p+6TGes0KILN4WioOEKY0t\r\n0V1Zr8bejDW49cu1Awy443SrauhFLOInubZLA8S9ZvwTVIvpmTDjdQ==\r\n-----END CERTIFICATE-----",
	}
}

// GetBlackDuckDefaultExternalDB returns a BlackDuck with an external Data Base
func GetBlackDuckDefaultExternalDB() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-externaldb",
		LicenseKey:        "",
		CertificateName:   "default",
		DbPrototype:       "",
		Size:              "small",
		PersistentStorage: false,
		ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
			PostgresHost:          "<<IP/FQDN>>",
			PostgresPort:          5432,
			PostgresAdmin:         "blackduck",
			PostgresUser:          "blackduck_user",
			PostgresSsl:           false,
			PostgresAdminPassword: "<<PASSWORD>>",
			PostgresUserPassword:  "<<PASSWORD>>",
		},
		Type:    "worker",
		Version: "2019.4.0",
	}
}

// GetBlackDuckDefaultIPV6Disabled returns a BlackDuck with IPV6 Disabled
func GetBlackDuckDefaultIPV6Disabled() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:       "blackduck-ipv6disabled",
		LicenseKey:      "",
		CertificateName: "default",
		Environs: []string{
			"IPV4_ONLY:1",
			"BLACKDUCK_HUB_SERVER_ADDRESS:0.0.0.0",
		},
		DbPrototype:       "",
		Size:              "small",
		PersistentStorage: false,
		Type:              "worker",
		Version:           "2019.4.0",
	}
}

// GetOpsSightTemplate returns the required fields for OpsSight
func GetOpsSightTemplate() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Perceptor: &opssightv1.Perceptor{
			Name:                           "perceptor",
			Port:                           3001,
			Image:                          "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         "",
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver: false,
			EnablePodPerceiver:   true,
			Port:                 3002,
			ImagePerceiver: &opssightv1.ImagePerceiver{
				Name:  "image-perceiver",
				Image: "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
			},
			PodPerceiver: &opssightv1.PodPerceiver{
				Name:  "pod-perceiver",
				Image: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
			},
			ServiceAccount:            "perceiver",
			AnnotationIntervalSeconds: 30,
			DumpIntervalMinutes:       30,
		},
		ScannerPod: &opssightv1.ScannerPod{
			Name: "perceptor-scanner",
			ImageFacade: &opssightv1.ImageFacade{
				Port:           3004,
				Image:          "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
				ServiceAccount: "perceptor-scanner",
				Name:           "perceptor-imagefacade",
			},
			Scanner: &opssightv1.Scanner{
				Name:                 "perceptor-scanner",
				Port:                 3003,
				Image:                "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
				ClientTimeoutSeconds: 600,
			},
			ReplicaCount:   1,
			ImageDirectory: "/var/images",
		},
		Prometheus: &opssightv1.Prometheus{
			Name:   "prometheus",
			Image:  "docker.io/prom/prometheus:v2.1.0",
			Port:   9090,
			Expose: "",
		},
		Skyfire: &opssightv1.Skyfire{
			Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			Name:                         "skyfire",
			Port:                         3005,
			PrometheusPort:               3006,
			ServiceAccount:               "skyfire",
			HubClientTimeoutSeconds:      100,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec:                      GetBlackDuckTemplate(),
		},
		EnableMetrics: true,
		EnableSkyfire: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "perceptor",
		ConfigMapName: "opssight",
		DesiredState:  "START",
	}
}

// GetOpsSightDefault returns the required fields for OpsSight
func GetOpsSightDefault() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Namespace: "opssight-test",
		Perceptor: &opssightv1.Perceptor{
			Name:                           "opssight-core",
			Port:                           3001,
			Image:                          "docker.io/blackducksoftware/opssight-core:2.2.3-RC",
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         "",
		},
		ScannerPod: &opssightv1.ScannerPod{
			Name: "opssight-scanner",
			Scanner: &opssightv1.Scanner{
				Name:                 "opssight-scanner",
				Port:                 3003,
				Image:                "docker.io/blackducksoftware/opssight-scanner:2.2.3-RC",
				ClientTimeoutSeconds: 600,
			},
			ImageFacade: &opssightv1.ImageFacade{
				Name:               "opssight-image-getter",
				Port:               3004,
				InternalRegistries: []*opssightv1.RegistryAuth{},
				Image:              "docker.io/blackducksoftware/opssight-image-getter:2.2.3-RC",
				ServiceAccount:     "opssight-scanner",
				ImagePullerType:    "skopeo",
			},
			ReplicaCount: 1,
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver: false,
			EnablePodPerceiver:   true,
			Port:                 3002,
			ImagePerceiver: &opssightv1.ImagePerceiver{
				Name:  "opssight-image-processor",
				Image: "docker.io/blackducksoftware/opssight-image-processor:2.2.3-RC",
			},
			PodPerceiver: &opssightv1.PodPerceiver{
				Name:  "opssight-pod-processor",
				Image: "docker.io/blackducksoftware/opssight-pod-processor:2.2.3-RC",
			},
			ServiceAccount:            "opssight-processor",
			AnnotationIntervalSeconds: 30,
			DumpIntervalMinutes:       30,
		},
		Prometheus: &opssightv1.Prometheus{
			Name:   "prometheus",
			Port:   9090,
			Image:  "docker.io/prom/prometheus:v2.1.0",
			Expose: "",
		},
		EnableSkyfire: false,
		Skyfire: &opssightv1.Skyfire{
			Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			Name:                         "skyfire",
			Port:                         3005,
			PrometheusPort:               3006,
			ServiceAccount:               "skyfire",
			HubClientTimeoutSeconds:      120,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		EnableMetrics: true,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "blackduck",
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec:                      GetBlackDuckTemplate(),
		},
	}
}

// GetOpsSightDefaultWithIPV6DisabledBlackDuck retuns an OpsSight with a BlackDuck and
// IPV6 disabled
func GetOpsSightDefaultWithIPV6DisabledBlackDuck() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Namespace: "opssight-test",
		Perceptor: &opssightv1.Perceptor{
			Name:                           "opssight-core",
			Port:                           3001,
			Image:                          "docker.io/blackducksoftware/opssight-core:2.2.3-RC",
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         "",
		},
		ScannerPod: &opssightv1.ScannerPod{
			Name: "opssight-scanner",
			Scanner: &opssightv1.Scanner{
				Name:                 "opssight-scanner",
				Port:                 3003,
				Image:                "docker.io/blackducksoftware/opssight-scanner:2.2.3-RC",
				ClientTimeoutSeconds: 600,
			},
			ImageFacade: &opssightv1.ImageFacade{
				Name:               "opssight-image-getter",
				Port:               3004,
				InternalRegistries: []*opssightv1.RegistryAuth{},
				Image:              "docker.io/blackducksoftware/opssight-image-getter:2.2.3-RC",
				ServiceAccount:     "opssight-scanner",
				ImagePullerType:    "skopeo",
			},
			ReplicaCount: 1,
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver: false,
			EnablePodPerceiver:   true,
			ImagePerceiver: &opssightv1.ImagePerceiver{
				Name:  "opssight-image-processor",
				Image: "docker.io/blackducksoftware/opssight-image-processor:2.2.3-RC",
			},
			PodPerceiver: &opssightv1.PodPerceiver{
				Name:  "opssight-pod-processor",
				Image: "docker.io/blackducksoftware/opssight-pod-processor:2.2.3-RC",
			},
			ServiceAccount:            "opssight-processor",
			AnnotationIntervalSeconds: 30,
			DumpIntervalMinutes:       30,
			Port:                      3002,
		},
		Prometheus: &opssightv1.Prometheus{
			Name:   "prometheus",
			Port:   9090,
			Image:  "docker.io/prom/prometheus:v2.1.0",
			Expose: "",
		},
		EnableSkyfire: false,
		Skyfire: &opssightv1.Skyfire{
			Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			Name:                         "skyfire",
			Port:                         3005,
			PrometheusPort:               3006,
			ServiceAccount:               "skyfire",
			HubClientTimeoutSeconds:      120,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		EnableMetrics: true,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "blackduck",
		DesiredState:  "START",
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec: &blackduckv1.BlackduckSpec{
				LicenseKey:        "",
				PersistentStorage: false,
				CertificateName:   "default",
				Environs: []string{
					"IPV4_ONLY:1",
					"BLACKDUCK_HUB_SERVER_ADDRESS:0.0.0.0",
				},
				DbPrototype: "",
				Size:        "small",
				Type:        "worker",
				Version:     "2019.4.0",
			},
		},
	}
}

// GetAlertTemplate returns the required fields for Alert
func GetAlertTemplate() *alertv1.AlertSpec {
	return &alertv1.AlertSpec{}
}

// GetAlertDefault creates an Alert crd configuration object with defaults
func GetAlertDefault() *alertv1.AlertSpec {
	port := 8443
	standAlone := true

	return &alertv1.AlertSpec{
		Namespace:            "alert-test",
		Version:              "3.1.0",
		AlertImage:           "docker.io/blackducksoftware/blackduck-alert:3.1.0",
		CfsslImage:           "docker.io/blackducksoftware/blackduck-cfssl:1.0.0",
		ExposeService:        "NODEPORT",
		Port:                 &port,
		EncryptionPassword:   "",
		EncryptionGlobalSalt: "",
		PersistentStorage:    true,
		PVCName:              "alert-pvc",
		StandAlone:           &standAlone,
		PVCSize:              "5G",
		PVCStorageClass:      "",
		AlertMemory:          "2560M",
		CfsslMemory:          "640M",
		Environs: []string{
			"ALERT_SERVER_PORT:8443",
			"PUBLIC_HUB_WEBSERVER_HOST:localhost",
			"PUBLIC_HUB_WEBSERVER_PORT:443",
		},
	}
}
