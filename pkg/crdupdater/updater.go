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
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// UpdateComponents consist of methods to add, patch or remove the components for update events
type UpdateComponents interface {
	buildNewAndOldObject() error
	add(bool) (bool, error)
	list() (interface{}, error)
	delete(name string) error
	remove() error
	patch(interface{}, bool) (bool, error)
}

// CommonConfig stores the common configuration for add, patch or remove the components for update events
type CommonConfig struct {
	kubeConfig    *rest.Config
	kubeClient    *kubernetes.Clientset
	dryRun        bool
	namespace     string
	labelSelector string
}

// Updater handles in updating the components
type Updater struct {
	updaters []UpdateComponents
	dryRun   bool
}

// NewUpdater will create the specification that is used for updating the components
func NewUpdater(dryRun bool) *Updater {
	updater := Updater{
		updaters: make([]UpdateComponents, 0),
		dryRun:   dryRun,
	}
	return &updater
}

// AddUpdater will add the updater to the list
func (u *Updater) AddUpdater(updater UpdateComponents) {
	u.updaters = append(u.updaters, updater)
}

// Update add or remove the components
func (u *Updater) Update() error {
	isPatched := false
	for _, updater := range u.updaters {
		if !u.dryRun {
			err := updater.buildNewAndOldObject()
			if err != nil {
				return errors.Annotatef(err, "build components:")
			}
		}
		isUpdated, err := updater.add(isPatched)
		isPatched = isPatched || isUpdated
		if err != nil {
			return errors.Annotatef(err, "add/patch components:")
		}
		if !u.dryRun {
			err = updater.remove()
			if err != nil {
				return errors.Annotatef(err, "remove components:")
			}
		}
	}
	return nil
}

// CRUDComponents will add, update or delete components
func CRUDComponents(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, dryRun bool, namespace string, components *api.ComponentList, labelSelector string) []error {
	var errors []error
	updater := NewUpdater(dryRun)

	commonConfig := &CommonConfig{kubeConfig: kubeConfig, kubeClient: kubeClient, dryRun: dryRun, namespace: namespace, labelSelector: labelSelector}

	// cluster role
	clusterRoles, err := NewClusterRole(commonConfig, components.ClusterRoles)
	errors = append(errors, fmt.Errorf("unable to create new cluster role updater due to %+v", err))
	updater.AddUpdater(clusterRoles)

	// cluster role binding
	clusterRoleBindings, err := NewClusterRoleBinding(commonConfig, components.ClusterRoleBindings)
	errors = append(errors, fmt.Errorf("unable to create new cluster role binding updater due to %+v", err))
	updater.AddUpdater(clusterRoleBindings)

	// service account
	serviceAccounts, err := NewServiceAccount(commonConfig, components.ServiceAccounts)
	errors = append(errors, fmt.Errorf("unable to create new service account updater due to %+v", err))
	updater.AddUpdater(serviceAccounts)

	// config map
	configMaps, err := NewConfigMap(commonConfig, components.ConfigMaps)
	errors = append(errors, fmt.Errorf("unable to create new config map updater due to %+v", err))
	updater.AddUpdater(configMaps)

	// secret
	secrets, err := NewSecret(commonConfig, components.Secrets)
	errors = append(errors, fmt.Errorf("unable to create new secret updater due to %+v", err))
	updater.AddUpdater(secrets)

	// service
	services, err := NewService(commonConfig, components.Services)
	errors = append(errors, fmt.Errorf("unable to create new service updater due to %+v", err))
	updater.AddUpdater(services)

	// replication controller
	rcs, err := NewReplicationController(commonConfig, components.ReplicationControllers)
	errors = append(errors, fmt.Errorf("unable to create new replication controller updater due to %+v", err))
	updater.AddUpdater(rcs)

	// deployment
	deployments, err := NewDeployment(commonConfig, components.Deployments)
	errors = append(errors, fmt.Errorf("unable to create new deployment updater due to %+v", err))
	updater.AddUpdater(deployments)

	// execute updates for all added components
	err = updater.Update()
	errors = append(errors, err)

	return errors
}
