/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.config. The ASF licenses this file
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

package perceptor

import (
	"testing"
)

func TestScannerContainers(t *testing.T) {
	d := NewPerceptorAppDefaults()
	a, _ := NewApp(d)

	ifObj := a.imageFacadeContainer()
	// Image facade needs to be privileged !
	if *ifObj.GetObj().Privileged == false {
		t.Errorf("%v %v", ifObj.GetObj().Name, ifObj.GetObj().Privileged)
	}

	sObj := a.scannerContainer()
	// The scanner needs to be UNPRIVILEGED
	if *sObj.GetObj().Privileged == true {
		t.Errorf("%v %v", sObj.GetObj().Name, sObj.GetObj().Privileged)
	}

	s0 := sObj.GetObj().Name
	s := sObj.GetObj().VolumeMounts[1].Store
	if s != "var-images" {
		t.Errorf("%v %v", s0, s)
	}
}

func TestScannerServiceAccount(t *testing.T) {
	d := NewPerceptorAppDefaults()
	a, _ := NewApp(d)
	a.configServiceAccounts()

	pod, _ := a.scannerPod()
	t.Logf("template: %v ", pod.GetObj())
	name := pod.GetObj().Account
	if name == "" {
		t.Errorf("scanner svc ==> ( %v ) EMPTY !", name)
	}
}
