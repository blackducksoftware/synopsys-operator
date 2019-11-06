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

// SecurityContext will contain the specifications of a security context
type SecurityContext struct {
	FsGroup    *int64 `json:"fsGroup"`
	RunAsUser  *int64 `json:"runAsUser"`
	RunAsGroup *int64 `json:"runAsGroup"`
}

// SetSecurityContextInPodConfig sets the Security Context fields in the PodConfig
func SetSecurityContextInPodConfig(podConfig *PodConfig, securityContext *SecurityContext, isOpenshift bool) {
	if securityContext != nil {
		podConfig.RunAsUser = securityContext.RunAsUser
		podConfig.RunAsGroup = securityContext.RunAsGroup

		if !isOpenshift {
			podConfig.FSGID = securityContext.FsGroup
		}
	} else {
		// if not openshift and the user doesn't specify a securityContext, then FSGID still needs to be set to 0
		if !isOpenshift {
			podConfig.FSGID = IntToInt64(0)
		}
	}
}
