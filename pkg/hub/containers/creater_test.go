/*
Copyright (C) 2018 Synopsys, Inc.

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

import (
	"fmt"
	"testing"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
)

func TestImageTag(t *testing.T) {
	hubVersion := "5.0.0"
	externalVersion := "1.0.0"
	hubSpec := &v1.HubSpec{ImageTagMap: map[string]string{"authentication": hubVersion, "documentation": hubVersion, "jobrunner": hubVersion,
		"registration": hubVersion, "scan": hubVersion, "webapp": hubVersion, "cfssl": externalVersion, "logstash": externalVersion,
		"nginx": externalVersion, "solr": externalVersion, "zookeeper": externalVersion}, HubVersion: "4.5.0"}
	creater := NewCreater(nil, hubSpec, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)

	external_1_0_0 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl"}
	internal_50 := []string{"registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range external_1_0_0 {
		if creater.getTag(v) == externalVersion {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
	for _, v := range internal_50 {
		if creater.getTag(v) == hubVersion {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}

	hubSpec1 := &v1.HubSpec{HubVersion: "4.5.0"}
	creater = NewCreater(nil, hubSpec1, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)
	all_50 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl", "registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range all_50 {
		if creater.getTag(v) == "4.5.0" {
			fmt.Printf("%s: %s\n", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
}
