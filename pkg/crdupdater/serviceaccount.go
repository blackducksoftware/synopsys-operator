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
	corev1 "k8s.io/api/core/v1"
)

// ServiceAccount stores the configuration to add or delete the service account
type ServiceAccount struct {
	config             *CommonConfig
	deployer           *util.DeployerHelper
	serviceAccounts    []*components.ServiceAccount
	oldServiceAccounts map[string]*corev1.ServiceAccount
	newServiceAccounts map[string]*corev1.ServiceAccount
}

// NewServiceAccount returns the service account
func NewServiceAccount(config *CommonConfig, serviceAccounts []*components.ServiceAccount) (*ServiceAccount, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &ServiceAccount{
		config:             config,
		deployer:           deployer,
		serviceAccounts:    serviceAccounts,
		oldServiceAccounts: make(map[string]*corev1.ServiceAccount, 0),
		newServiceAccounts: make(map[string]*corev1.ServiceAccount, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new service account
func (s *ServiceAccount) buildNewAndOldObject() error {
	// build old service account
	oldServiceAccounts, err := s.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get service accounts for %s", s.config.namespace)
	}
	for _, oldServiceAccount := range oldServiceAccounts.(*corev1.ServiceAccountList).Items {
		s.oldServiceAccounts[oldServiceAccount.GetName()] = &oldServiceAccount
	}

	// build new service account
	for _, newCr := range s.serviceAccounts {
		newServiceAccountKube, err := newCr.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert service account %s to kube %s", newCr.GetName(), s.config.namespace)
		}
		s.newServiceAccounts[newCr.GetName()] = newServiceAccountKube.(*corev1.ServiceAccount)
	}

	return nil
}

// add adds the service account
func (s *ServiceAccount) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, serviceAccount := range s.serviceAccounts {
		if _, ok := s.oldServiceAccounts[serviceAccount.GetName()]; !ok {
			s.deployer.Deployer.AddServiceAccount(serviceAccount)
			isAdded = true
		}
	}
	if isAdded && !s.config.dryRun {
		err := s.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy service account in %s", s.config.namespace)
		}
	}
	return isAdded, nil
}

// get gets the service account
func (s *ServiceAccount) get(name string) (interface{}, error) {
	return util.GetServiceAccount(s.config.kubeClient, s.config.namespace, name)
}

// list lists all the service accounts
func (s *ServiceAccount) list() (interface{}, error) {
	return util.ListServiceAccounts(s.config.kubeClient, s.config.namespace, s.config.labelSelector)
}

// delete deletes the service account
func (s *ServiceAccount) delete(name string) error {
	return util.DeleteServiceAccount(s.config.kubeClient, s.config.namespace, name)
}

// remove removes the service account
func (s *ServiceAccount) remove() error {
	// compare the old and new service account and delete if needed
	for _, oldServiceAccount := range s.oldServiceAccounts {
		if _, ok := s.newServiceAccounts[oldServiceAccount.GetName()]; !ok {
			err := s.delete(oldServiceAccount.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete service account %s in namespace %s", oldServiceAccount.GetName(), s.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the service account
func (s *ServiceAccount) patch(sa interface{}, isPatched bool) (bool, error) {
	return false, nil
}
