package model

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WARNING: If you add a config value, make sure to
// add it to the parameterize function as well !
type ProtoformConfig struct {
	// Dry run wont actually install, but will print the objects definitions out.
	DryRun bool

	// CONTAINER CONFIGS
	// These are sed replaced into the config maps for the containers.
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	HubHost                   string
	HubUser                   string
	HubUserPassword           string
	HubPort                   int
	ConcurrentScanLimit       int

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPasswordOrToken string
	DockerUsername        string

	ServiceAccounts map[string]string

	UseMockPerceptorMode bool
}

func (p *ProtoformConfig) fillInDefaultValues() {
	if p.PerceptorHost == "" {
		p.PerceptorHost = "perceptor"
	}
	if p.PerceptorPort == 0 {
		p.PerceptorPort = 3001
	}
	if p.AnnotationIntervalSeconds == 0 {
		p.AnnotationIntervalSeconds = 30
	}
	if p.DumpIntervalMinutes == 0 {
		p.DumpIntervalMinutes = 30
	}
	if p.HubHost == "" {
		// meaningless default unless your in same namespace as hub.
		p.HubHost = "nginx-webapp-logstash"
	}
	if p.HubUser == "" {
		p.HubUser = "sysadmin"
	}
	if p.HubUserPassword == "" {
		panic("config failing: cannot continue without a hub password!!!")
	}
	if p.HubPort == 0 {
		p.HubPort = 443
	}
	if p.DockerUsername == "" {
		p.DockerUsername = "admin"
	}
	if p.DockerPasswordOrToken == "" {
		panic("config failing: cannot continue without a Docker password!!!")
	}
	if p.ConcurrentScanLimit == 0 {
		p.ConcurrentScanLimit = 2
	}
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
		Dockerpassword: pc.DockerPasswordOrToken,
		Dockerusername: pc.DockerUsername,
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

func (pc *ProtoformConfig) ToConfigMap() []*v1.ConfigMap {
	pc.fillInDefaultValues()

	return []*v1.ConfigMap{
		makeConfigMap("prometheus", "prometheus.yml", `{"global":{"scrape_interval":"5s"},"scrape_configs":[{"job_name":"perceptor-scrape","scrape_interval":"5s","static_configs":[{"targets":["perceptor:3001","perceptor-scanner:3003","perceptor-imagefacade:4000"]}]}]}`),
		makeConfigMap("perceptor-scanner-config", "perceptor_scanner_conf.yaml", pc.PerceptorScannerConfig()),
		makeConfigMap("kube-generic-perceiver-config", "perceiver.yaml", pc.PodPerceiverConfig()),
		makeConfigMap("perceptor-config", "perceptor_conf.yaml", pc.PerceptorConfig()),
		makeConfigMap("openshift-perceiver-config", "perceiver.yaml", pc.ImagePerceiverConfig()),
		makeConfigMap("perceptor-imagefacade-config", "perceptor_imagefacade_conf.yaml", pc.PerceptorImagefacadeConfig()),
	}
}
