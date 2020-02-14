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

package containers

import "fmt"

var imageTags = map[string]map[string]string{
	"2019.10.3": {
		"blackduck-authentication": "2019.10.3",
		"blackduck-documentation":  "2019.10.3",
		"blackduck-jobrunner":      "2019.10.3",
		"blackduck-registration":   "2019.10.3",
		"blackduck-scan":           "2019.10.3",
		"blackduck-webapp":         "2019.10.3",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.9",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.09",
		"rabbitmq":                 "1.0.0",
	},
	"2019.10.2": {
		"blackduck-authentication": "2019.10.2",
		"blackduck-documentation":  "2019.10.2",
		"blackduck-jobrunner":      "2019.10.2",
		"blackduck-registration":   "2019.10.2",
		"blackduck-scan":           "2019.10.2",
		"blackduck-webapp":         "2019.10.2",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.9",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.09",
		"rabbitmq":                 "1.0.0",
	},
	"2019.10.1": {
		"blackduck-authentication": "2019.10.1",
		"blackduck-documentation":  "2019.10.1",
		"blackduck-jobrunner":      "2019.10.1",
		"blackduck-registration":   "2019.10.1",
		"blackduck-scan":           "2019.10.1",
		"blackduck-webapp":         "2019.10.1",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.9",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.09",
		"rabbitmq":                 "1.0.0",
	},
	"2019.10.0": {
		"blackduck-authentication": "2019.10.0",
		"blackduck-documentation":  "2019.10.0",
		"blackduck-jobrunner":      "2019.10.0",
		"blackduck-registration":   "2019.10.0",
		"blackduck-scan":           "2019.10.0",
		"blackduck-webapp":         "2019.10.0",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.9",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.09",
		"rabbitmq":                 "1.0.0",
	},
	"2019.8.1": {
		"blackduck-authentication": "2019.8.1",
		"blackduck-documentation":  "2019.8.1",
		"blackduck-jobrunner":      "2019.8.1",
		"blackduck-registration":   "2019.8.1",
		"blackduck-scan":           "2019.8.1",
		"blackduck-webapp":         "2019.8.1",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.8",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.06",
		"rabbitmq":                 "1.0.0",
	},
	"2019.8.0": {
		"blackduck-authentication": "2019.8.0",
		"blackduck-documentation":  "2019.8.0",
		"blackduck-jobrunner":      "2019.8.0",
		"blackduck-registration":   "2019.8.0",
		"blackduck-scan":           "2019.8.0",
		"blackduck-webapp":         "2019.8.0",
		"blackduck-cfssl":          "1.0.0",
		"blackduck-logstash":       "1.0.4",
		"blackduck-nginx":          "1.0.8",
		"blackduck-zookeeper":      "1.0.0",
		"blackduck-upload-cache":   "1.0.9",
		"appcheck-worker":          "2019.06",
		"rabbitmq":                 "1.0.0",
	},
}

// GetImageTag returns the image tag of the given container
func (c *Creater) GetImageTag(inputs ...string) string {
	var name string
	defaultRegistry := "docker.io/blackducksoftware"
	switch len(inputs) {
	case 1:
		name = inputs[0]
	case 2:
		name = inputs[0]
		defaultRegistry = inputs[1]
	default:
		return ""
	}

	if _, ok := imageTags[c.blackDuck.Spec.Version][name]; ok {
		confImageTag := c.GetFullContainerNameFromImageRegistryConf(name)
		if len(confImageTag) > 0 {
			return confImageTag
		}

		if c.blackDuck.Spec.RegistryConfiguration != nil && len(c.blackDuck.Spec.RegistryConfiguration.Registry) > 0 {
			return fmt.Sprintf("%s/%s:%s", c.blackDuck.Spec.RegistryConfiguration.Registry, name, imageTags[c.blackDuck.Spec.Version][name])
		}
		return fmt.Sprintf("%s/%s:%s", defaultRegistry, name, imageTags[c.blackDuck.Spec.Version][name])
	}
	return ""
}

// GetVersions returns the supported versions
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
