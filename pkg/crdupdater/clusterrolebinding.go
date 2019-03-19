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
)

// ClusterRoleBinding stores the configuration to add or delete the cluster role binding
type ClusterRoleBinding struct {
	config                 *CommonConfig
	deployer               *util.DeployerHelper
	clusterRoleBindings    []*components.ClusterRoleBinding
	oldClusterRoleBindings map[string]*rbacv1.ClusterRoleBinding
	newClusterRoleBindings map[string]*rbacv1.ClusterRoleBinding
}

// NewClusterRoleBinding returns the cluster role binding
func NewClusterRoleBinding(config *CommonConfig, clusterRoleBindings []*components.ClusterRoleBinding) (*ClusterRoleBinding, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &ClusterRoleBinding{
		config:                 config,
		deployer:               deployer,
		clusterRoleBindings:    clusterRoleBindings,
		oldClusterRoleBindings: make(map[string]*rbacv1.ClusterRoleBinding, 0),
		newClusterRoleBindings: make(map[string]*rbacv1.ClusterRoleBinding, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new cluster role binding
func (c *ClusterRoleBinding) buildNewAndOldObject() error {
	// build old cluster role binding
	oldCrbs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get cluster role bindings for %s", c.config.namespace)
	}
	for _, oldCrb := range oldCrbs.(*rbacv1.ClusterRoleBindingList).Items {
		c.oldClusterRoleBindings[oldCrb.GetName()] = &oldCrb
	}

	// build new cluster role binding
	for _, newCrb := range c.clusterRoleBindings {
		newClusterRoleBindingKube, err := newCrb.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert cluster role binding %s to kube %s", newCrb.GetName(), c.config.namespace)
		}
		c.newClusterRoleBindings[newCrb.GetName()] = newClusterRoleBindingKube.(*rbacv1.ClusterRoleBinding)
	}

	return nil
}

// add adds the cluster role binding
func (c *ClusterRoleBinding) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, clusterRoleBinding := range c.clusterRoleBindings {
		if _, ok := c.oldClusterRoleBindings[clusterRoleBinding.GetName()]; !ok {
			c.deployer.Deployer.AddClusterRoleBinding(clusterRoleBinding)
			isAdded = true
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy cluster role binding in %s", c.config.namespace)
		}
	}
	return isAdded, nil
}

// get gets the cluster role binding
func (c *ClusterRoleBinding) get(name string) (interface{}, error) {
	return util.GetClusterRoleBinding(c.config.kubeClient, name)
}

// list lists all the cluster role bindings
func (c *ClusterRoleBinding) list() (interface{}, error) {
	return util.ListClusterRoleBindings(c.config.kubeClient, c.config.labelSelector)
}

// delete deletes the cluster role binding
func (c *ClusterRoleBinding) delete(name string) error {
	return util.DeleteClusterRoleBinding(c.config.kubeClient, name)
}

// remove removes the cluster role binding
func (c *ClusterRoleBinding) remove() error {
	// compare the old and new cluster role binding and delete if needed
	for _, oldClusterRoleBinding := range c.oldClusterRoleBindings {
		if _, ok := c.newClusterRoleBindings[oldClusterRoleBinding.GetName()]; !ok {
			err := c.delete(oldClusterRoleBinding.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete cluster role binding %s in namespace %s", oldClusterRoleBinding.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the cluster role binding
func (c *ClusterRoleBinding) patch(crb interface{}, isPatched bool) (bool, error) {
	return false, nil
}
