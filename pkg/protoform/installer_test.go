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

package protoform

import (
	"os"
	"testing"
)

func TestProto(t *testing.T) {
	os.Setenv("PCP_HUBUSERPASSWORD", "example")

	d := NewDefaultsObj()
	d.DefaultCPU = "300m"
	d.DefaultMem = "1300Mi"

	installer := NewInstaller(d, "/etc/protoform")
	installer.init()
	rcsArray := installer.replicationControllers

	for i, rcs := range rcsArray {
		t.Logf("%v : %v", i, rcs.Name)
	}

	// Image facade needs to be privileged !
	if *rcsArray[2].PodTemplate.Containers[1].Privileged == false {
		t.Errorf("%v %v", rcsArray[2].PodTemplate.Containers[1].Name, *rcsArray[2].PodTemplate.Containers[1].Privileged)
	}

	// The scanner needs to be UNPRIVILEGED
	if *rcsArray[2].PodTemplate.Containers[0].Privileged == true {
		t.Errorf("%v %v", rcsArray[2].PodTemplate.Containers[0].Name, *rcsArray[2].PodTemplate.Containers[0].Privileged)
	}

	t.Logf("template: %v ", rcsArray[2].PodTemplate)
	scannerSvc := rcsArray[2].PodTemplate.Account
	if scannerSvc == "" {
		t.Errorf("scanner svc ==> ( %v ) EMPTY !", scannerSvc)
	}

	s0 := rcsArray[2].PodTemplate.Containers[0].Name
	s := rcsArray[2].PodTemplate.Containers[0].VolumeMounts[1].Store
	if s != "var-images" {
		t.Errorf("%v %v", s0, s)
	}
}
