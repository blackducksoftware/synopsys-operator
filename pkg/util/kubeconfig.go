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

package util

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GetKubeRestConfig gets the user's kubeconfig from their system
func GetKubeRestConfig() *rest.Config {
	log.Debugf("Getting Kube Rest Config\n")
	// Determine Config Paths
	var masterURL = ""
	var kubeconfigpath = ""
	if home := homeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfigpath = ""
	}
	// Get Rest Config using Paths
	var restconfig *rest.Config
	var err error
	if masterURL == "" && kubeconfigpath == "" {
		restconfig, err = rest.InClusterConfig() // if no paths then use in-cluster config
	} else {
		restconfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigpath,
			},
			&clientcmd.ConfigOverrides{
				ClusterInfo: clientcmdapi.Cluster{
					Server: masterURL,
				},
			}).ClientConfig()
	}
	if err != nil {
		panic(err)
	}
	return restconfig
}

// homeDir determines the user's home directory path
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
