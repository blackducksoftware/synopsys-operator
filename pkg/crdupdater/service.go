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

// Service stores the configuration to add or delete the service object
type Service struct {
	config      *CommonConfig
	deployer    *util.DeployerHelper
	services    []*components.Service
	oldServices map[string]*corev1.Service
	newServices map[string]*corev1.Service
}

// NewService returns the service
func NewService(config *CommonConfig, services []*components.Service) (*Service, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &Service{
		config:      config,
		deployer:    deployer,
		services:    services,
		oldServices: make(map[string]*corev1.Service, 0),
		newServices: make(map[string]*corev1.Service, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new service
func (s *Service) buildNewAndOldObject() error {
	// build old service
	oldSvcs, err := s.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get services for %s", s.config.namespace)
	}
	for _, oldSvc := range oldSvcs.(*corev1.ServiceList).Items {
		s.oldServices[oldSvc.GetName()] = &oldSvc
	}

	// build new service
	for _, newSvc := range s.services {
		newServiceKube, err := newSvc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert service %s to kube %s", newSvc.GetName(), s.config.namespace)
		}
		s.newServices[newSvc.GetName()] = newServiceKube.(*corev1.Service)
	}

	return nil
}

// add adds the service
func (s *Service) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, service := range s.services {
		if _, ok := s.oldServices[service.GetName()]; !ok {
			s.deployer.Deployer.AddService(service)
			isAdded = true
		}
	}
	if isAdded && !s.config.dryRun {
		err := s.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy service in %s", s.config.namespace)
		}
	}
	return isAdded, nil
}

// get gets the service
func (s *Service) get(name string) (interface{}, error) {
	return util.GetService(s.config.kubeClient, s.config.namespace, name)
}

// list lists all the services
func (s *Service) list() (interface{}, error) {
	return util.ListServices(s.config.kubeClient, s.config.namespace, s.config.labelSelector)
}

// delete deletes the service
func (s *Service) delete(name string) error {
	return util.DeleteService(s.config.kubeClient, s.config.namespace, name)
}

// remove removes the service
func (s *Service) remove() error {
	// compare the old and new service and delete if needed
	for _, oldService := range s.oldServices {
		if _, ok := s.newServices[oldService.GetName()]; !ok {
			err := s.delete(oldService.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete service %s in namespace %s", oldService.GetName(), s.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the service
func (s *Service) patch(svc interface{}, isPatched bool) (bool, error) {
	return false, nil
}
