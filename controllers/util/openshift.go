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

package util

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// IsOpenShift returns whether the cluster is an OpenShift
func IsOpenShift(kubeConfig *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return false
	}

	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false
	}
	apiGroups := apiList.Groups
	for i := 0; i < len(apiGroups); i++ {
		if apiGroups[i].Name == "route.openshift.io" {
			return true
		}
	}
	return false
}
