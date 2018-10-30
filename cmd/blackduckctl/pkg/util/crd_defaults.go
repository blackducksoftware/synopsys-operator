package util

import alertv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
import opssightv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
import hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"

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
	return &opssightv1.OpsSightSpec{
		Perceptor: &opssightv1.Perceptor{
			Name:  "perceptor",
			Port:  3001,
			Image: "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
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
				Port:               3004,
				InternalRegistries: []opssightv1.RegistryAuth{},
				Image:              "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
				ServiceAccount:     "perceptor-scanner",
				Name:               "perceptor-imagefacade",
			},
			Scanner: &opssightv1.Scanner{
				Name:                 "perceptor-scanner",
				Port:                 3003,
				Image:                "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
				ClientTimeoutSeconds: 600,
			},
			ReplicaCount: 1,
		},
		Skyfire: &opssightv1.Skyfire{
			Image:          "gcr.io/saas-hub-stg/blackducksoftware/skyfire:master",
			Name:           "skyfire",
			Port:           3005,
			ServiceAccount: "skyfire",
		},
		Hub: &opssightv1.Hub{
			User:                         "sysadmin",
			Port:                         443,
			ConcurrentScanLimit:          2,
			TotalScanLimit:               1000,
			PasswordEnvVar:               "PCP_HUBUSERPASSWORD",
			Password:                     "blackduck",
			InitialCount:                 1,
			MaxCount:                     1,
			DeleteHubThresholdPercentage: 50,
			HubSpec: GetHubDefaultValue(),
		},
		EnableMetrics: true,
		EnableSkyfire: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "perceptor",
	}
}
