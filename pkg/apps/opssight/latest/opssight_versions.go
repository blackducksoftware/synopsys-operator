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

package opssight

import "fmt"

// imageTags is a map of the OpsSights versions to it's images and tags
var imageTags = map[string]map[string]RepoTag{
	"2.2.3": {
		"opssight-core":            RepoTag{Repo: "docker.io/blackducksoftware", Tag: "2.2.3"},
		"opssight-scanner":         RepoTag{Repo: "docker.io/blackducksoftware", Tag: "2.2.3"},
		"opssight-image-getter":    RepoTag{Repo: "docker.io/blackducksoftware", Tag: "2.2.3"},
		"opssight-image-processor": RepoTag{Repo: "docker.io/blackducksoftware", Tag: "2.2.3"},
		"opssight-pod-processor":   RepoTag{Repo: "docker.io/blackducksoftware", Tag: "2.2.3"},
		"prometheus":               RepoTag{Repo: "docker.io/prom", Tag: "v2.1.0"},
		"pyfire":                   RepoTag{Repo: "gcr.io/saas-hub-stg/blackducksoftware", Tag: "master"},
		"skyfire":                  RepoTag{Repo: "gcr.io/saas-hub-stg/blackducksoftware", Tag: "master"},
	},
}

// RepoTag is a struct for the repository and tag of images for OpsSight versions
type RepoTag struct {
	Repo string
	Tag  string
}

// GetImageTag returns the url for an image
func GetImageTag(version, name string) string {
	return fmt.Sprintf("%s/%s:%s", imageTags[version][name].Repo, name, imageTags[version][name].Tag)
}

// GetVersions returns the supported versions for this Creater
func GetVersions() []string {
	var versions []string
	for k := range imageTags {
		versions = append(versions, k)
	}
	return versions
}
