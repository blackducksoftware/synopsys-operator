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

package containers

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// GetServiceAccount will return the service account
func (c *Creater) GetServiceAccount() *components.ServiceAccount {
	svc := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      c.hubSpec.Namespace,
		Namespace: c.hubSpec.Namespace,
	})

	svc.AddLabels(c.GetVersionLabel("serviceAccount"))

	return svc
}

// GetClusterRoleBinding will return the cluster role binding
func (c *Creater) GetClusterRoleBinding() *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       c.hubSpec.Namespace,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      c.hubSpec.Namespace,
		Namespace: c.hubSpec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "synopsys-operator-admin",
	})

	clusterRoleBinding.AddLabels(c.GetVersionLabel("clusterRoleBinding"))

	return clusterRoleBinding
}
