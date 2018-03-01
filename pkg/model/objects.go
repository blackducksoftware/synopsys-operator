package model

// Prometheus
type PromGlobal struct {
	ScrapeInterval string `yaml:"scrape_interval"`
}
type PromStaticConfig struct {
	Targets []string `yaml:"targets"`
}
type PromScrape struct {
	JobName           string `yaml:"job_name"`
	ScrapeInterval    string `yaml:"scrape_interval"`
	PromStaticConfigs []PromStaticConfig
}
type Prometheus struct {
	global        PromGlobal `yaml:"global"`
	scrapeConfigs PromScrape `yaml:"scrape_configs"`
}

type PerceptorScannerConfig struct {
	HubHost         string
	HubUser         string
	HubUserPassword string
}

// kube-generic-perceiver-config
type KubeGenericPerceiverConfig struct {
	PerceptorHost             string // "perceptor"
	PerceptorPort             int    //  3001
	AnnotationIntervalSeconds int    // 30
	DumpIntervalMinutes       int    //: 30
}

type OSPerceiverConfig struct {
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
}
