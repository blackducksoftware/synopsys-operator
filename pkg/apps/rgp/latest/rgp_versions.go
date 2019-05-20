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

package rgp

import "fmt"

// imageTags is a map of the Rgp versions to it's images and tags
var imageTags = map[string]map[string]string{
	"2019.04": {
		"reporting-frontend-service":        "0.0.673",
		"reporting-polaris-service":         "0.0.111",
		"reporting-report-service":          "0.0.450",
		"reporting-rp-issue-manager":        "0.0.487",
		"reporting-rp-portfolio-service":    "0.0.663",
		"reporting-tools-portfolio-service": "0.0.974",
	},
	"2019.03": {
		"reporting-frontend-service":        "0.0.658",
		"reporting-polaris-service":         "0.0.83",
		"reporting-report-service":          "0.0.419",
		"reporting-rp-issue-manager":        "0.0.445",
		"reporting-rp-portfolio-service":    "0.0.637",
		"reporting-tools-portfolio-service": "0.0.911",
	},
}

// GetImageTag returns the url for an image
func GetImageTag(version, name string) string {
	return fmt.Sprintf("gcr.io/snps-swip-staging/%s:%s", name, imageTags[version][name])
}

// GetVersions returns the supported versions for this Creater
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
