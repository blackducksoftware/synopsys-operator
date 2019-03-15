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

// ClusterRoleBinding stores the configuration to add or delete the cluster role binding
type ClusterRoleBinding struct {
	kubeConfig             *rest.Config
	kubeClient             *kubernetes.Clientset
	deployer               *util.DeployerHelper
	namespace              string
	clusterRoleBindings    []*components.ClusterRoleBinding
	labelSelector          string
	oldClusterRoleBindings map[string]*rbacv1.ClusterRoleBinding
	newClusterRoleBindings map[string]*rbacv1.ClusterRoleBinding
}

// NewClusterRoleBinding returns the cluster role binding
func NewClusterRoleBinding(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, clusterRoleBindings []*components.ClusterRoleBinding,
	namespace string, labelSelector string) (*ClusterRoleBinding, error) {
	deployer, err := util.NewDeployer(kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	return &ClusterRoleBinding{
		kubeConfig:             kubeConfig,
		kubeClient:             kubeClient,
		deployer:               deployer,
		namespace:              namespace,
		clusterRoleBindings:    clusterRoleBindings,
		labelSelector:          labelSelector,
		oldClusterRoleBindings: make(map[string]*rbacv1.ClusterRoleBinding, 0),
		newClusterRoleBindings: make(map[string]*rbacv1.ClusterRoleBinding, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new cluster role binding
func (c *ClusterRoleBinding) buildNewAndOldObject() error {
	// build old cluster role binding
	oldCrbs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get cluster role bindings for %s", c.namespace)
	}
	for _, oldCrb := range oldCrbs.(*rbacv1.ClusterRoleBindingList).Items {
		c.oldClusterRoleBindings[oldCrb.GetName()] = &oldCrb
	}

	// build new cluster role binding
	for _, newCrb := range c.clusterRoleBindings {
		newClusterRoleBindingKube, err := newCrb.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert cluster role binding %s to kube %s", newCrb.GetName(), c.namespace)
		}
		c.newClusterRoleBindings[newCrb.GetName()] = newClusterRoleBindingKube.(*rbacv1.ClusterRoleBinding)
	}

	return nil
}

// add adds the cluster role binding
func (c *ClusterRoleBinding) add() error {
	isAdded := false
	for _, clusterRoleBinding := range c.clusterRoleBindings {
		if _, ok := c.oldClusterRoleBindings[clusterRoleBinding.GetName()]; !ok {
			c.deployer.Deployer.AddClusterRoleBinding(clusterRoleBinding)
			isAdded = true
		}
	}
	if isAdded {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return errors.Annotatef(err, "unable to deploy cluster role binding in %s", c.namespace)
		}
	}
	return nil
}

// list lists all the cluster role bindings
func (c *ClusterRoleBinding) list() (interface{}, error) {
	return util.ListClusterRoleBindings(c.kubeClient, c.labelSelector)
}

// delete deletes the cluster role binding
func (c *ClusterRoleBinding) delete(name string) error {
	return util.DeleteClusterRoleBinding(c.kubeClient, name)
}

// remove removes the cluster role binding
func (c *ClusterRoleBinding) remove() error {
	// compare the old and new cluster role binding and delete if needed
	for _, oldClusterRoleBinding := range c.oldClusterRoleBindings {
		if _, ok := c.newClusterRoleBindings[oldClusterRoleBinding.GetName()]; !ok {
			err := c.delete(oldClusterRoleBinding.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete cluster role binding %s in namespace %s", oldClusterRoleBinding.GetName(), c.namespace)
			}
		}
	}
	return nil
}

// patch patches the cluster role binding
func (c *ClusterRoleBinding) patch(crb interface{}) error {
	return nil
}
