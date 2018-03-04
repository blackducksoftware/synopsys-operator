package model

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProtoformConfig struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	// perceptor config
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	HubHost                   string
	HubUser                   string
	HubUserPassword           string
	HubPort                   int
	ConcurrentScanLimit       int

	UseMockPerceptorMode bool

	AuxConfig *AuxiliaryConfig
}

func (pc *ProtoformConfig) PerceptorConfig() string {
	jsonBytes, err := json.Marshal(PerceptorConfig{
		ConcurrentScanLimit: pc.ConcurrentScanLimit,
		HubHost:             pc.HubHost,
		HubUser:             pc.HubUser,
		HubUserPassword:     pc.HubUserPassword,
		UseMockMode:         pc.UseMockPerceptorMode,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PodPerceiverConfig() string {
	jsonBytes, err := json.Marshal(PodPerceiverConfig{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             pc.PerceptorHost,
		PerceptorPort:             pc.PerceptorPort,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) ImagePerceiverConfig() string {
	jsonBytes, err := json.Marshal(ImagePerceiverConfig{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             pc.PerceptorHost,
		PerceptorPort:             pc.PerceptorPort,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PerceptorScannerConfig() string {
	jsonBytes, err := json.Marshal(PerceptorScannerConfig{
		HubHost:         pc.HubHost,
		HubPort:         pc.HubPort,
		HubUser:         pc.HubUser,
		HubUserPassword: pc.HubUserPassword,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (pc *ProtoformConfig) PerceptorImagefacadeConfig() string {
	jsonBytes, err := json.Marshal(PerceptorImagefacadeConfig{
		Dockerpassword: pc.AuxConfig.DockerPassword,
		Dockerusername: pc.AuxConfig.DockerUsername,
	})
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func makeConfigMap(name string, filename string, contents string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			filename: contents,
		},
	}
}

func (pc *ProtoformConfig) ToConfigMaps() []*v1.ConfigMap {
	return []*v1.ConfigMap{
		makeConfigMap("prometheus", "prometheus.yml", `{"global":{"scrape_interval":"5s"},"scrape_configs":[{"job_name":"perceptor-scrape","scrape_interval":"5s","static_configs":[{"targets":["perceptor:3001","perceptor-scanner:3003","perceptor-imagefacade:3004"]}]}]}`),
		makeConfigMap("perceptor-scanner-config", "perceptor_scanner_conf.yaml", pc.PerceptorScannerConfig()),
		makeConfigMap("kube-generic-perceiver-config", "perceiver.yaml", pc.PodPerceiverConfig()),
		makeConfigMap("perceptor-config", "perceptor_conf.yaml", pc.PerceptorConfig()),
		makeConfigMap("openshift-perceiver-config", "perceiver.yaml", pc.ImagePerceiverConfig()),
		makeConfigMap("perceptor-imagefacade-config", "perceptor_imagefacade_conf.yaml", pc.PerceptorImagefacadeConfig()),
	}
}
