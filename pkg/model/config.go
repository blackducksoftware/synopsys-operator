package model

import (
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WARNING: If you add a config value, make sure to
// add it to the parameterize function as well !
type ProtoformConfig struct {

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

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPasswordOrToken string
	DockerUsername        string

	ServiceAccounts map[string]string
}

func (p *ProtoformConfig) parameterize(json string) string {
	n := 1000
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
	json = strings.Replace(json, "_1", p.PerceptorHost, n)
	json = strings.Replace(json, "_2", string(p.PerceptorPort), n)
	json = strings.Replace(json, "_3", string(p.AnnotationIntervalSeconds), n)
	json = strings.Replace(json, "_4", string(p.DumpIntervalMinutes), n)
	json = strings.Replace(json, "_5", p.HubHost, n)
	json = strings.Replace(json, "_6", p.HubUser, n)
	json = strings.Replace(json, "_7", p.HubUserPassword, n)
	json = strings.Replace(json, "_8", string(p.HubPort), n)

	return json
}

// prometheus.yml
func (p *ProtoformConfig) ToConfigMap() []*v1.ConfigMap {

	// TODO, parameterize prometheus
	// strings.Replace(prometheus_t,
	configs := map[string]string{
		"prometheus":                    "prometheus.yml",
		"perceptor-scanner-config":      "perceptor_scanner_conf.yaml",
		"kube-generic-perceiver-config": "perceiver.yaml",
		"perceptor-config":              "perceptor_conf.yaml",
		"openshift-perceiver-config":    "perceiver.yaml",
	}

	// Sed replace these.  Fine grained control over the json default format
	// makes this easier then actually modelling / mutating nested json in golang.
	// (I think)? Due to the fct that nested anonymous []string's seem to not be
	// "a thing".
	defaults := map[string]string{
		"prometheus": `{
			"global":{"scrape_interval":"5s"},
			"scrape_configs":[
				{"job_name":"perceptor-scrape",
					"scrape_interval":"5s",
					"static_configs":[{"targets":["perceptor:3001","perceptor-scanner:3003"]}
					]}]
		}`,
		"perceptor-scanner-config": `{
		  "HubHost": "_5",
			"HubPort": "_8"
		  "HubUser": "_6",
		  "HubUserPassword": "_7"
		}`,
		"kube-generic-perceiver-config": `{
			"PerceptorHost": "_1",
			"PerceptorPort": "_2",
			"AnnotationIntervalSeconds": "_3",
			"DumpIntervalMinutes": "_4"
		}`,
		"openshift_perceiver_config_t": `{
			"PerceptorHost": "_1",
			"PerceptorPort": "_2",
			"AnnotationIntervalSeconds": "_3",
			"DumpIntervalMinutes": "_4"
		}`,
	}

	maps := make([]*v1.ConfigMap, len(configs))
	x := 0
	for config, filename := range configs {
		contents := p.parameterize(defaults[config])
		maps[x] = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prometheus",
			},
			Data: map[string]string{
				filename: contents,
			},
		}
		x = x + 1
	}
	return maps
}
