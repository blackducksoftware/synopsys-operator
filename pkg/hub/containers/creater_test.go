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
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/sirupsen/logrus"
)

type TG struct {
}

func (t *TG) getUID(s string) *int64 {
	x := int64(100)
	return &x
}

func (t *TG) getTag(s string) string {
	return "CORRECT"
}

func TestC(t *testing.T) {
	c := &TG{}
	myCont := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: "a/b/c:tag"},
	}
	if myCont.ContainerConfig.Image == "a/b/c:CORRECT" {
		logrus.Infof("test setup isnt right")
		t.Fail()
	}
	PostEdit(myCont, c)
	if myCont.ContainerConfig.Image != "a/b/c:CORRECT" {
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
	hubSpec := &v1.HubSpec{ImageTagMap: []string{fmt.Sprintf("authentication:%s", hubVersion), fmt.Sprintf("documentation:%s", hubVersion),
		fmt.Sprintf("jobrunner:%s", hubVersion), fmt.Sprintf("registration:%s", hubVersion), fmt.Sprintf("scan:%s", hubVersion),
		fmt.Sprintf("webapp:%s", hubVersion), fmt.Sprintf("cfssl:%s", externalVersion), fmt.Sprintf("logstash:%s", externalVersion),
		fmt.Sprintf("nginx:%s", externalVersion), fmt.Sprintf("solr:%s", externalVersion), fmt.Sprintf("zookeeper:%s", externalVersion)},
		HubVersion: "4.5.0"}
	creater := NewCreater(nil, hubSpec, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil, nil)

	external100 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl"}
	internal50 := []string{"registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range external100 {
		if creater.getTag(v) == externalVersion {
			t.Logf("%s: %s", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
	for _, v := range internal50 {
		if creater.getTag(v) == hubVersion {
			t.Logf("%s: %s", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}

	hubSpec1 := &v1.HubSpec{HubVersion: "4.5.0"}
	creater = NewCreater(nil, hubSpec1, nil, []*horizonapi.EnvConfig{}, []*horizonapi.EnvConfig{}, nil, nil, nil)
	all50 := []string{"zookeeper", "nginx", "solr", "logstash", "cfssl", "registration", "webapp", "jobrunner", "documentation", "scan", "authentication"}
	for _, v := range all50 {
		if creater.getTag(v) == "4.5.0" {
			t.Logf("%s: %s", v, creater.getTag(v))
		} else {
			t.Fail()
		}
	}
}
