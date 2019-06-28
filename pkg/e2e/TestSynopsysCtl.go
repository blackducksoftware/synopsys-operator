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
package e2e

import (
	"os/exec"

	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Synopsysctl TODO
type Synopsysctl struct{}

func NewSynopsysctl() *Synopsysctl {
	return &Synopsysctl{}
}

// Exec TODO
func (*Synopsysctl) Exec(args ...string) (string, error) {
	var cmd *exec.Cmd
	path := "synopsysctl"
	cmd = exec.Command(path, args...)
	log.Printf("[Exec] cmd.Args: %s \n\n", cmd.Args)
	stdoutErr, err := cmd.CombinedOutput()
	log.Printf("[Exec] stdoutErr: %s, err: %v \n\n", stdoutErr, err)
	if err != nil {
		return string(stdoutErr), err
	}
	return string(stdoutErr), nil
}

func GetRestConfig() (*rest.Config, error) {
	kubeconfig := ""
	restconfig, err := protoform.GetKubeConfig(kubeconfig, false)
	log.Debugf("rest config: %+v", restconfig)
	if err != nil {
		return nil, err
	}
	log.Printf("[getRestConfig] restconfig: %v \n\n", restconfig)
	return restconfig, nil
}

// getKubeClient gets the kubernetes client
func GetKubeClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	log.Printf("[getKubeClient] client: %v \n\n", client)
	return client, nil
}
