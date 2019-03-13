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

package crdupdater

import (
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClusterRole stores the configuration to add or delete the cluster role
type ClusterRole struct {
	kubeConfig    *rest.Config
	kubeClient    *kubernetes.Clientset
	deployer      *util.DeployerHelper
	namespace     string
	clusterRoles  []*components.ClusterRole
	labelSelector string
}

// NewClusterRole returns the cluster role
func NewClusterRole(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, clusterRoles []*components.ClusterRole, namespace string, labelSelector string) (*ClusterRole, error) {
	deployer, err := util.NewDeployer(kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	return &ClusterRole{kubeConfig: kubeConfig, kubeClient: kubeClient, deployer: deployer, namespace: namespace, clusterRoles: clusterRoles, labelSelector: labelSelector}, nil
}

// get get the cluster role
func (c *ClusterRole) get(name string) (interface{}, error) {
	return util.GetClusterRole(c.kubeClient, name)
}

// add adds the cluster role
func (c *ClusterRole) add() error {
	isAdded := false
	for _, clusterRole := range c.clusterRoles {
		_, err := c.get(clusterRole.GetName())
		if err != nil {
			c.deployer.Deployer.AddClusterRole(clusterRole)
			isAdded = true
		}
	}
	if isAdded {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return errors.Annotatef(err, "unable to deploy cluster role in %s", c.namespace)
		}
	}
	return nil
}

// list lists all the cluster roles
func (c *ClusterRole) list() (interface{}, error) {
	return util.ListClusterRoles(c.kubeClient, c.labelSelector)
}

// delete deletes the cluster role
func (c *ClusterRole) delete(name string) error {
	return util.DeleteClusterRole(c.kubeClient, name)
}

// remove removes the cluster role
func (c *ClusterRole) remove() error {
	oldCrs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to list the cluster role for %s", c.namespace)
	}

	oldClusterRoles := oldCrs.(*rbacv1.ClusterRoleList)

	// construct the new cluster role using horizon to kube method
	newClusterRoles := make(map[string]*rbacv1.ClusterRole)
	for _, newCr := range c.clusterRoles {
		newClusterRoleKube, err := newCr.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert cluster role %s to kube in namespace %s", newCr.GetName(), c.namespace)
		}
		newClusterRoles[newCr.GetName()] = newClusterRoleKube.(*rbacv1.ClusterRole)
	}

	// compare the old and new cluster role and delete if needed
	for _, oldClusterRole := range oldClusterRoles.Items {
		if _, ok := newClusterRoles[oldClusterRole.GetName()]; !ok {
			err = c.delete(oldClusterRole.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete cluster role %s in namespace %s", oldClusterRole.GetName(), c.namespace)
			}
		}
	}
	return nil
}
