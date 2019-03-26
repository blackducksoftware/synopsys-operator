package containers

import "fmt"

var imageTags = map[string]map[string]string{
	"2018.12.0": {
		"blackduck-authentication": "2018.12.0",
		"blackduck-documentation":  "2018.12.0",
		"blackduck-jobrunner":      "2018.12.0",
		"blackduck-registration":   "2018.12.0",
		"blackduck-scan":           "2018.12.0",
		"blackduck-webapp":         "2018.12.0",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2018.12.1": {
		"blackduck-authentication": "2018.12.1",
		"blackduck-documentation":  "2018.12.1",
		"blackduck-jobrunner":      "2018.12.1",
		"blackduck-registration":   "2018.12.1",
		"blackduck-scan":           "2018.12.1",
		"blackduck-webapp":         "2018.12.1",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2018.12.2": {
		"blackduck-authentication": "2018.12.2",
		"blackduck-documentation":  "2018.12.2",
		"blackduck-jobrunner":      "2018.12.2",
		"blackduck-registration":   "2018.12.2",
		"blackduck-scan":           "2018.12.2",
		"blackduck-webapp":         "2018.12.2",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2018.12.3": {
		"blackduck-authentication": "2018.12.3",
		"blackduck-documentation":  "2018.12.3",
		"blackduck-jobrunner":      "2018.12.3",
		"blackduck-registration":   "2018.12.3",
		"blackduck-scan":           "2018.12.3",
		"blackduck-webapp":         "2018.12.3",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2018.12.4": {
		"blackduck-authentication": "2018.12.4",
		"blackduck-documentation":  "2018.12.4",
		"blackduck-jobrunner":      "2018.12.4",
		"blackduck-registration":   "2018.12.4",
		"blackduck-scan":           "2018.12.4",
		"blackduck-webapp":         "2018.12.4",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2019.2.0": {
		"blackduck-authentication": "2019.2.0",
		"blackduck-documentation":  "2019.2.0",
		"blackduck-jobrunner":      "2019.2.0",
		"blackduck-registration":   "2019.2.0",
		"blackduck-scan":           "2019.2.0",
		"blackduck-webapp":         "2019.2.0",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2019.2.1": {
		"blackduck-authentication": "2019.2.1",
		"blackduck-documentation":  "2019.2.1",
		"blackduck-jobrunner":      "2019.2.1",
		"blackduck-registration":   "2019.2.1",
		"blackduck-scan":           "2019.2.1",
		"blackduck-webapp":         "2019.2.1",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
	"2019.2.2": {
		"blackduck-authentication": "2019.2.2",
		"blackduck-documentation":  "2019.2.2",
		"blackduck-jobrunner":      "2019.2.2",
		"blackduck-registration":   "2019.2.2",
		"blackduck-scan":           "2019.2.2",
		"blackduck-webapp":         "2019.2.2",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.2",
		"blackduck-nginx":          "1.0.2",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.3",
		"appcheck-worker":          "1.0.1",
		"rabbitmq":                 "1.0.0",
	},
}

func (c *Creater) getImageTag(name string) string {
	confImageTag := c.GetFullContainerNameFromImageRegistryConf(name)
	if len(confImageTag) > 0 {
		return confImageTag
	}
	return fmt.Sprintf("docker.io/blackducksoftware/%s:%s", name, imageTags[c.hubSpec.Version][name])
}

// GetVersions will return the supported versions
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
