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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Service stores the configuration to add or delete the service object
type Service struct {
	kubeConfig    *rest.Config
	kubeClient    *kubernetes.Clientset
	deployer      *util.DeployerHelper
	namespace     string
	services      []*components.Service
	labelSelector string
}

// NewService returns the service
func NewService(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, services []*components.Service, namespace string, labelSelector string) (*Service, error) {
	deployer, err := util.NewDeployer(kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	return &Service{kubeConfig: kubeConfig, kubeClient: kubeClient, deployer: deployer, namespace: namespace, services: services, labelSelector: labelSelector}, nil
}

// get get the service
func (s *Service) get(name string) (interface{}, error) {
	return util.GetService(s.kubeClient, s.namespace, name)
}

// add adds the service
func (s *Service) add() error {
	isAdded := false
	for _, service := range s.services {
		_, err := s.get(service.GetName())
		if err != nil {
			s.deployer.Deployer.AddService(service)
			isAdded = true
		}
	}
	if isAdded {
		err := s.deployer.Deployer.Run()
		if err != nil {
			return errors.Annotatef(err, "unable to deploy service in %s", s.namespace)
		}
	}
	return nil
}

// list lists all the services
func (s *Service) list() (interface{}, error) {
	return util.ListServices(s.kubeClient, s.namespace, s.labelSelector)
}

// delete deletes the serive
func (s *Service) delete(name string) error {
	return util.DeleteService(s.kubeClient, s.namespace, name)
}

// remove removes the service
func (s *Service) remove() error {
	oldSvcs, err := s.list()
	if err != nil {
		return errors.Annotatef(err, "unable to list the services for %s", s.namespace)
	}
	oldServices := oldSvcs.(*corev1.ServiceList)

	// construct the new services using horizon to kube method
	newServices := make(map[string]*corev1.Service)
	for _, newSvc := range s.services {
		newServiceKube, err := newSvc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert service %s to kube in opssight %s", newSvc.GetName(), s.namespace)
		}
		newServices[newSvc.GetName()] = newServiceKube.(*corev1.Service)
	}

	// compare the old and new service and delete if needed
	for _, oldService := range oldServices.Items {
		if _, ok := newServices[oldService.GetName()]; !ok {
			err = s.delete(oldService.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete service %s in namespace %s", oldService.GetName(), s.namespace)
			}
		}
	}
	return nil
}
