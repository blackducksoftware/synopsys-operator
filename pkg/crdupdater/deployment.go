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
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Deployment stores the configuration to add or delete the replication controller object
type Deployment struct {
	kubeConfig     *rest.Config
	kubeClient     *kubernetes.Clientset
	deployer       *util.DeployerHelper
	namespace      string
	deployments    []*components.Deployment
	labelSelector  string
	isPatched      bool
	oldDeployments map[string]*appsv1.Deployment
	newDeployments map[string]*appsv1.Deployment
}

// NewDeployment returns the replication controller
func NewDeployment(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, deployments []*components.Deployment,
	namespace string, labelSelector string, isPatched bool) (*Deployment, error) {
	deployer, err := util.NewDeployer(kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	return &Deployment{
		kubeConfig:     kubeConfig,
		kubeClient:     kubeClient,
		deployer:       deployer,
		namespace:      namespace,
		deployments:    deployments,
		labelSelector:  labelSelector,
		isPatched:      isPatched,
		oldDeployments: make(map[string]*appsv1.Deployment, 0),
		newDeployments: make(map[string]*appsv1.Deployment, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new replication controller
func (r *Deployment) buildNewAndOldObject() error {
	// build old replication controller
	oldRCs, err := r.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get replication controllers for %s", r.namespace)
	}
	for _, oldRC := range oldRCs.(*appsv1.DeploymentList).Items {
		r.oldDeployments[oldRC.GetName()] = &oldRC
	}

	// build new replication controller
	for _, newRc := range r.deployments {
		newDeploymentKube, err := newRc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert replication controller %s to kube %s", newRc.GetName(), r.namespace)
		}
		r.newDeployments[newRc.GetName()] = newDeploymentKube.(*appsv1.Deployment)
	}

	return nil
}

// add adds the replication controller
func (r *Deployment) add() error {
	isAdded := false
	for _, deployment := range r.deployments {
		if _, ok := r.oldDeployments[deployment.GetName()]; !ok {
			r.deployer.Deployer.AddDeployment(deployment)
			isAdded = true
		} else {
			err := r.patch(deployment)
			if err != nil {
				return errors.Annotatef(err, "patch replication controller:")
			}
		}
	}
	if isAdded {
		err := r.deployer.Deployer.Run()
		if err != nil {
			return errors.Annotatef(err, "unable to deploy replication controller in %s", r.namespace)
		}
	}
	return nil
}

// list lists all the replication controllers
func (r *Deployment) list() (interface{}, error) {
	return util.ListDeployments(r.kubeClient, r.namespace, r.labelSelector)
}

// delete deletes the replication controller
func (r *Deployment) delete(name string) error {
	return util.DeleteDeployment(r.kubeClient, r.namespace, name)
}

// remove removes the replication controller
func (r *Deployment) remove() error {
	// compare the old and new replication controller and delete if needed
	for _, oldDeployment := range r.oldDeployments {
		if _, ok := r.newDeployments[oldDeployment.GetName()]; !ok {
			err := r.delete(oldDeployment.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete replication controller %s in namespace %s", oldDeployment.GetName(), r.namespace)
			}
		}
	}
	return nil
}

// deploymentComparator used to compare Replication controller attributes
type deploymentComparator struct {
	Image    string
	Replicas *int32
	MinCPU   *resource.Quantity
	MaxCPU   *resource.Quantity
	MinMem   *resource.Quantity
	MaxMem   *resource.Quantity
}

// patch patches the replication controller
func (r *Deployment) patch(rc interface{}) error {
	deployment := rc.(*components.Deployment)
	// check isPatched, why?
	// if there is any configuration change, irrespective of comparing any changes, patch the replication controller
	if r.isPatched {
		err := util.PatchDeployment(r.kubeClient, *r.newDeployments[deployment.GetName()], *r.oldDeployments[deployment.GetName()])
		if err != nil {
			return errors.Annotatef(err, "unable to patch replication controller %s in namespace %s", deployment.GetName(), r.namespace)
		}
		return nil
	}

	// check whether the replication controller or its container got changed
	isChanged := false
	for _, oldContainer := range r.oldDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
		for _, newContainer := range r.newDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
			if strings.EqualFold(oldContainer.Name, newContainer.Name) &&
				!reflect.DeepEqual(
					deploymentComparator{
						Image:    oldContainer.Image,
						Replicas: r.oldDeployments[deployment.GetName()].Spec.Replicas,
						MinCPU:   oldContainer.Resources.Requests.Cpu(),
						MaxCPU:   oldContainer.Resources.Limits.Cpu(),
						MinMem:   oldContainer.Resources.Requests.Memory(),
						MaxMem:   oldContainer.Resources.Limits.Memory(),
					},
					deploymentComparator{
						Image:    newContainer.Image,
						Replicas: r.newDeployments[deployment.GetName()].Spec.Replicas,
						MinCPU:   newContainer.Resources.Requests.Cpu(),
						MaxCPU:   newContainer.Resources.Limits.Cpu(),
						MinMem:   newContainer.Resources.Requests.Memory(),
						MaxMem:   newContainer.Resources.Limits.Memory(),
					}) {
				isChanged = true
			}
		}
	}

	// if there is any change from the above step, patch the replication controller
	if isChanged {
		err := util.PatchDeployment(r.kubeClient, *r.newDeployments[deployment.GetName()], *r.oldDeployments[deployment.GetName()])
		if err != nil {
			return errors.Annotatef(err, "unable to patch rc %s to kube in namespace %s", deployment.GetName(), r.namespace)
		}
	}
	return nil
}
