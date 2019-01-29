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

package blackduck

import (
	"testing"
)

func TestAddHub(t *testing.T) {
	// deployer, _ := NewHubInstaller()
	// createHub := &Blackduck{
	// 	Namespace:      "blackduck-hub",
	// 	DockerRegistry: "docker.io",
	// 	DockerRepo:     "blackducksoftware",
	// 	HubVersion:     "4.6.2",
	// 	// GcloudProject:  "gke-verification",
	// 	// InstanceName:   "test-senthil-2018-06-13-02-12-54",
	// 	Flavor: "small",
	// 	// RegionName:     "us-central1",
	// 	AdminPassword: "blackduck",
	// 	UserPassword:  "blackduck",
	// }
	// deployer.CreateHubDeployer(createHub)
	// fmt.Printf("%+v", deployer.Deployer)
	// err := deployer.Deployer.Run()
	//
	// fmt.Printf("Deployments failed because %+v", err)

	// req := deployer.Deployer.Client.RESTClient().Post().
	// 	Resource("pods").
	// 	Name("postgres-84bdfc6469-hkpgp").
	// 	Namespace("blackduck-hub").
	// 	SubResource("exec")
	// scheme := runtime.NewScheme()
	//
	// var stdin io.Reader
	//
	// stdin = nil
	//
	// parameterCodec := runtime.NewParameterCodec(scheme)
	// req.VersionedParams(&core_v1.PodExecOptions{
	// 	// Command: strings.Fields("-- bash -c \"curl https://raw.githubusercontent.com/blackducksoftware/opssight-connector/master/install/hub/external-postgres-init.pgsql > /tmp/external-postgres-init.pgsql\""),
	// 	Command:   []string{"/bin/bash", "ls"},
	// 	Container: "postgres",
	// 	Stdin:     stdin != nil,
	// 	Stdout:    true,
	// 	Stderr:    true,
	// 	TTY:       false,
	// }, parameterCodec)
	//
	// fmt.Printf("Request URL: %+v, request: %+v \n", req.URL().String(), req)
	//
	// exec, err := remotecommand.NewSPDYExecutor(deployer.Config, "POST", req.URL())
	// fmt.Printf("exec: %+v, error: %+v\n", exec, err)
	// if err != nil {
	// 	fmt.Errorf("error while creating Executor: %v", err)
	// }
	//
	// var stdout, stderr bytes.Buffer
	// err = exec.Stream(remotecommand.StreamOptions{
	// 	Stdin:  stdin,
	// 	Stdout: &stdout,
	// 	Stderr: &stderr,
	// 	Tty:    false,
	// })
	//
	// fmt.Printf("error: %+v\n", err)
	// if err != nil {
	// 	fmt.Errorf("error in Stream: %v", err)
	// }
	//
	// fmt.Printf("%s, %s", stdout.String(), stderr.String())
	// InitDatabase("blackduck-hub")
}
