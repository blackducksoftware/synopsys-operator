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

package perceptor

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/blackducksoftware/horizon/pkg/components"
)

func TestGenerateStringFromStringArr(t *testing.T) {
	var strArr = []string{"example1", "example2"}
	a := App{}
	str := a.generateStringFromStringArr(strArr)

	if str != "[\"example1\",\"example2\"]" {
		fmt.Printf("The final string is %s", str)
		t.Fail()
	}
}

func TestPerceptorContainers(t *testing.T) {
	d := NewPerceptorAppDefaults()
	d.Registry = "gcr.io"
	d.ImagePath = "gke-verification/blackducksoftware"
	d.Namespace = "perceptor"
	d.DefaultVersion = "master"
	d.DefaultCPU = "300m"
	d.DefaultMem = "1300Mi"

	a, _ := NewApp(d)
	a.configServiceAccounts()
	a.substituteDefaultImageVersion()

	rcsArray := []*components.ReplicationController{}
	rcsArray = append(rcsArray, a.PerceptorReplicationController())
	rc, _ := a.PodPerceiverReplicationController()
	rcsArray = append(rcsArray, rc)
	rc, _ = a.ImagePerceiverReplicationController()
	rcsArray = append(rcsArray, rc)
	rc, _ = a.ScannerReplicationController()
	rcsArray = append(rcsArray, rc)
	rc, _ = a.PerceptorSkyfireReplicationController()
	rcsArray = append(rcsArray, rc)

	args := map[string]string{
		"perceptor":             "/etc/perceptor/perceptor.yaml",
		"pod-perceiver":         "/etc/perceiver/perceiver.yaml",
		"image-perceiver":       "/etc/perceiver/perceiver.yaml",
		"perceptor-scanner":     "/etc/perceptor_scanner/perceptor_scanner.yaml",
		"perceptor-imagefacade": "/etc/perceptor_imagefacade/perceptor_imagefacade.json",
		"skyfire":               "/etc/skyfire/skyfire.yaml",
	}

	var imageRegexp = regexp.MustCompile("(.+)/(.+):(.+)")
	for _, rcs := range rcsArray {
		for _, container := range rcs.GetObj().Containers {

			// verify the image expressions
			match := imageRegexp.FindStringSubmatch(container.Image)
			if len(match) != 4 {
				t.Errorf("%s is not matching to the regex %s", container.Image, imageRegexp.String())
			}

			// verify the args parameter in Replication Controller
			arg := container.Args

			if arg[0].String() != args[container.Name] {
				t.Errorf("Arguments not matched for %s, Expected: %s, Actual: %s", container.Name, args[container.Name], arg[0].String())
			}

			// verify the default cpu parameters
			if strings.Compare(d.DefaultCPU, container.CPU.Min) != 0 {
				t.Errorf("Default CPU is not configured for %s, Expected: %s, Actual: %s", container.Name, d.DefaultCPU, container.CPU.Min)
			}

			// verify the default memory parameters
			if strings.Compare(d.DefaultMem, container.Mem.Min) != 0 {
				t.Errorf("Default memory is not configured for %s, Expected: %s, Actual: %s", container.Name, d.DefaultMem, container.Mem.Min)
			}

		}
	}

}
