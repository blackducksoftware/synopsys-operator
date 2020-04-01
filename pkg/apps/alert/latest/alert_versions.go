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

package alert

import "fmt"

// imageTags is a map of the Alert versions to it's images and tags
var imageTags = map[string]map[string]string{
	"3.1.0": {
		"blackduck-alert": "3.1.0",
		"blackduck-cfssl": "1.0.0",
	},
	"4.0.0": {
		"blackduck-alert": "4.0.0",
		"blackduck-cfssl": "1.0.0",
	},
	"4.1.0": {
		"blackduck-alert": "4.1.0",
		"blackduck-cfssl": "1.0.0",
	},
	"4.2.0": {
		"blackduck-alert": "4.2.0",
		"blackduck-cfssl": "1.0.0",
	},
	"5.0.0": {
		"blackduck-alert": "5.0.0",
		"blackduck-cfssl": "1.0.0",
	},
	"5.0.1": {
		"blackduck-alert": "5.0.1",
		"blackduck-cfssl": "1.0.0",
	},
	"5.0.2": {
		"blackduck-alert": "5.0.2",
		"blackduck-cfssl": "1.0.0",
	},
	"5.0.3": {
		"blackduck-alert": "5.0.3",
		"blackduck-cfssl": "1.0.0",
	},
	"5.1.0": {
		"blackduck-alert": "5.1.0",
		"blackduck-cfssl": "1.0.0",
	},
	"5.2.0": {
		"blackduck-alert": "5.2.0",
		"blackduck-cfssl": "1.0.0",
	},
	"5.2.1": {
		"blackduck-alert": "5.2.1",
		"blackduck-cfssl": "1.0.0",
	},
	"5.2.2": {
		"blackduck-alert": "5.2.2",
		"blackduck-cfssl": "1.0.0",
	},
	"5.2.3": {
		"blackduck-alert": "5.2.3",
		"blackduck-cfssl": "1.0.0",
	},
	"5.3.0": {
		"blackduck-alert": "5.3.0",
		"blackduck-cfssl": "1.0.0",
	},
}

// GetImageTag returns the url for an image
func GetImageTag(version, name string) string {
	return fmt.Sprintf("docker.io/blackducksoftware/%s:%s", name, imageTags[version][name])
}

// GetVersions returns the supported versions for this Creater
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
