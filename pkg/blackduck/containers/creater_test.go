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
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
)

func getUID(s int) *int64 {
	x := int64(s)
	return &x
}

func TestC(t *testing.T) {
	hubVersion := "5.0.0"
	externalVersion := "1.0.0"
	hubSpec := &v1.BlackduckSpec{ImageRegistries: []string{
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-authentication:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-documentation:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-jobrunner:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-registration:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-scan:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-webapp:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-cfssl:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-logstash:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-nginx:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-solr:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-zookeeper:%s", externalVersion)},
		Environs: []string{"HUB_VERSION:5.0.0"}}

	myCont := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: fmt.Sprintf("docker.io/blackducksoftware/blackduck-documentation:%s", hubVersion), UID: getUID(100)},
	}

	creater := NewCreater(nil, hubSpec, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)
	creater.PostEditContainer(myCont)
	if myCont.ContainerConfig.Image != fmt.Sprintf("docker.io/blackducksoftware/blackduck-documentation:%s", hubVersion) {
		logrus.Infof("Got wrong tag %v", myCont.ContainerConfig.Image)
		t.Fail()
	}
	if *myCont.ContainerConfig.UID != 100 {
		t.Fail()
	}
}

func TestImageTag(t *testing.T) {
	hubVersion := "5.0.0"
	externalVersion := "1.0.0"
	hubSpec := &v1.BlackduckSpec{ImageRegistries: []string{
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-authentication:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-documentation:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-jobrunner:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-registration:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-scan:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-webapp:%s", hubVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-cfssl:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-logstash:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-nginx:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-solr:%s", externalVersion),
		fmt.Sprintf("docker.io/blackducksoftware/blackduck-zookeeper:%s", externalVersion)},
		Environs: []string{"HUB_VERSION:4.5.0"}}
	creater := NewCreater(nil, hubSpec, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)

	external100 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl"}
	internal50 := []string{"registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range external100 {
		containerName := creater.getFullContainerName(v)
		if containerName == fmt.Sprintf("docker.io/blackducksoftware/blackduck-%s:%s", v, externalVersion) {
			t.Logf("%s: %s", v, containerName)
		} else {
			t.Fail()
		}
	}
	for _, v := range internal50 {
		containerName := creater.getFullContainerName(v)
		if containerName == fmt.Sprintf("docker.io/blackducksoftware/blackduck-%s:%s", v, hubVersion) {
			t.Logf("%s: %s", v, containerName)
		} else {
			t.Fail()
		}
	}

	hubSpec1 := &v1.BlackduckSpec{Environs: []string{"HUB_VERSION:4.5.0"}}
	creater = NewCreater(nil, hubSpec1, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil)
	all50 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl", "registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range all50 {
		containerName := creater.getFullContainerName(v)
		if containerName == fmt.Sprintf("docker.io/blackducksoftware/hub-%s:%s", v, "4.5.0") {
			t.Logf("%s: %s", v, containerName)
		} else {
			t.Fail()
		}
	}
}
