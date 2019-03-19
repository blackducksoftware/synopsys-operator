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

// ClusterRole stores the configuration to add or delete the cluster role
type ClusterRole struct {
	config          *CommonConfig
	deployer        *util.DeployerHelper
	clusterRoles    []*components.ClusterRole
	oldClusterRoles map[string]*rbacv1.ClusterRole
	newClusterRoles map[string]*rbacv1.ClusterRole
}

// NewClusterRole returns the cluster role
func NewClusterRole(config *CommonConfig, clusterRoles []*components.ClusterRole) (*ClusterRole, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &ClusterRole{
		config:          config,
		deployer:        deployer,
		clusterRoles:    clusterRoles,
		oldClusterRoles: make(map[string]*rbacv1.ClusterRole, 0),
		newClusterRoles: make(map[string]*rbacv1.ClusterRole, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new cluster role
func (c *ClusterRole) buildNewAndOldObject() error {
	// build old cluster role
	oldCrs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get cluster roles for %s", c.config.namespace)
	}
	for _, oldCr := range oldCrs.(*rbacv1.ClusterRoleList).Items {
		c.oldClusterRoles[oldCr.GetName()] = &oldCr
	}

	// build new cluster role
	for _, newCr := range c.clusterRoles {
		newClusterRoleKube, err := newCr.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert cluster role %s to kube %s", newCr.GetName(), c.config.namespace)
		}
		c.newClusterRoles[newCr.GetName()] = newClusterRoleKube.(*rbacv1.ClusterRole)
	}

	return nil
}

// add adds the cluster role
func (c *ClusterRole) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, clusterRole := range c.clusterRoles {
		if _, ok := c.oldClusterRoles[clusterRole.GetName()]; !ok {
			c.deployer.Deployer.AddClusterRole(clusterRole)
			isAdded = true
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy cluster role in %s", c.config.namespace)
		}
	}
	return isAdded, nil
}

// get gets the cluster role
func (c *ClusterRole) get(name string) (interface{}, error) {
	return util.GetClusterRole(c.config.kubeClient, name)
}

// list lists all the cluster roles
func (c *ClusterRole) list() (interface{}, error) {
	return util.ListClusterRoles(c.config.kubeClient, c.config.labelSelector)
}

// delete deletes the cluster role
func (c *ClusterRole) delete(name string) error {
	return util.DeleteClusterRole(c.config.kubeClient, name)
}

// remove removes the cluster role
func (c *ClusterRole) remove() error {
	// compare the old and new cluster role and delete if needed
	for _, oldClusterRole := range c.oldClusterRoles {
		if _, ok := c.newClusterRoles[oldClusterRole.GetName()]; !ok {
			err := c.delete(oldClusterRole.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete cluster role %s in namespace %s", oldClusterRole.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the cluster role
func (c *ClusterRole) patch(cr interface{}, isPatched bool) (bool, error) {
	return false, nil
}
