package containers

import "fmt"

var imageTags = map[string]map[string]string{
	"2019.4.0": {
		"blackduck-authentication": "2019.4.0",
		"blackduck-documentation":  "2019.4.0",
		"blackduck-jobrunner":      "2019.4.0",
		"blackduck-registration":   "2019.4.0",
		"blackduck-scan":           "2019.4.0",
		"blackduck-webapp":         "2019.4.0",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.7",
		"blackduck-solr":           "1.0.0",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.8",
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

// GetVersions returns the supported versions
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
