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
	"2019.4.0": {
		"blackduck-authentication": "2019.4.0",
		"blackduck-documentation":  "2019.4.0",
		"blackduck-jobrunner":      "2019.4.0",
		"blackduck-registration":   "2019.4.0",
		"blackduck-scan":           "2019.4.0",
		"blackduck-webapp":         "2019.4.0",
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

// GetVersions returns the supported versions
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
