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
)

// Deployment stores the configuration to add or delete the replication controller object
type Deployment struct {
	config         *CommonConfig
	deployer       *util.DeployerHelper
	deployments    []*components.Deployment
	oldDeployments map[string]*appsv1.Deployment
	newDeployments map[string]*appsv1.Deployment
}

// NewDeployment returns the replication controller
func NewDeployment(config *CommonConfig, deployments []*components.Deployment) (*Deployment, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &Deployment{
		config:         config,
		deployer:       deployer,
		deployments:    deployments,
		oldDeployments: make(map[string]*appsv1.Deployment, 0),
		newDeployments: make(map[string]*appsv1.Deployment, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new replication controller
func (d *Deployment) buildNewAndOldObject() error {
	// build old replication controller
	oldRCs, err := d.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get replication controllers for %s", d.config.namespace)
	}
	for _, oldRC := range oldRCs.(*appsv1.DeploymentList).Items {
		d.oldDeployments[oldRC.GetName()] = &oldRC
	}

	// build new replication controller
	for _, newRc := range d.deployments {
		newDeploymentKube, err := newRc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert replication controller %s to kube %s", newRc.GetName(), d.config.namespace)
		}
		d.newDeployments[newRc.GetName()] = newDeploymentKube.(*appsv1.Deployment)
	}

	return nil
}

// add adds the replication controller
func (d *Deployment) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, deployment := range d.deployments {
		if _, ok := d.oldDeployments[deployment.GetName()]; !ok {
			d.deployer.Deployer.AddDeployment(deployment)
			isAdded = true
		} else {
			_, err := d.patch(deployment, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch replication controller:")
			}
		}
	}
	if isAdded && !d.config.dryRun {
		err := d.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy replication controller in %s", d.config.namespace)
		}
	}
	return false, nil
}

// list lists all the replication controllers
func (d *Deployment) list() (interface{}, error) {
	return util.ListDeployments(d.config.kubeClient, d.config.namespace, d.config.labelSelector)
}

// delete deletes the replication controller
func (d *Deployment) delete(name string) error {
	return util.DeleteDeployment(d.config.kubeClient, d.config.namespace, name)
}

// remove removes the replication controller
func (d *Deployment) remove() error {
	// compare the old and new replication controller and delete if needed
	for _, oldDeployment := range d.oldDeployments {
		if _, ok := d.newDeployments[oldDeployment.GetName()]; !ok {
			err := d.delete(oldDeployment.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete replication controller %s in namespace %s", oldDeployment.GetName(), d.config.namespace)
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
func (d *Deployment) patch(rc interface{}, isPatched bool) (bool, error) {
	deployment := rc.(*components.Deployment)
	// check isPatched, why?
	// if there is any configuration change, irrespective of comparing any changes, patch the replication controller
	if isPatched && !d.config.dryRun {
		err := util.PatchDeployment(d.config.kubeClient, *d.newDeployments[deployment.GetName()], *d.oldDeployments[deployment.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch replication controller %s in namespace %s", deployment.GetName(), d.config.namespace)
		}
		return false, nil
	}

	// check whether the replication controller or its container got changed
	isChanged := false
	for _, oldContainer := range d.oldDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
		for _, newContainer := range d.newDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
			if strings.EqualFold(oldContainer.Name, newContainer.Name) && !d.config.dryRun &&
				!reflect.DeepEqual(
					deploymentComparator{
						Image:    oldContainer.Image,
						Replicas: d.oldDeployments[deployment.GetName()].Spec.Replicas,
						MinCPU:   oldContainer.Resources.Requests.Cpu(),
						MaxCPU:   oldContainer.Resources.Limits.Cpu(),
						MinMem:   oldContainer.Resources.Requests.Memory(),
						MaxMem:   oldContainer.Resources.Limits.Memory(),
					},
					deploymentComparator{
						Image:    newContainer.Image,
						Replicas: d.newDeployments[deployment.GetName()].Spec.Replicas,
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
		err := util.PatchDeployment(d.config.kubeClient, *d.newDeployments[deployment.GetName()], *d.oldDeployments[deployment.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch rc %s to kube in namespace %s", deployment.GetName(), d.config.namespace)
		}
	}
	return false, nil
}
