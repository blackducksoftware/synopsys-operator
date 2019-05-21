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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// StatefulSet stores the configuration to add or delete the statefulset object
type StatefulSet struct {
	config          *CommonConfig
	deployer        *util.DeployerHelper
	statefulSets    []*components.StatefulSet
	oldStatefulSets map[string]appsv1.StatefulSet
	newStatefulSets map[string]*appsv1.StatefulSet
}

// NewStatefulSet returns the statefulSet
func NewStatefulSet(config *CommonConfig, statefulSets []*components.StatefulSet) (*StatefulSet, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newStatefulSets := append([]*components.StatefulSet{}, statefulSets...)
	for i := 0; i < len(newStatefulSets); i++ {
		if !isLabelsExist(config.expectedLabels, newStatefulSets[i].Labels) {
			newStatefulSets = append(newStatefulSets[:i], newStatefulSets[i+1:]...)
			i--
		}
	}
	return &StatefulSet{
		config:          config,
		deployer:        deployer,
		statefulSets:    newStatefulSets,
		oldStatefulSets: make(map[string]appsv1.StatefulSet, 0),
		newStatefulSets: make(map[string]*appsv1.StatefulSet, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new statefulSet
func (s *StatefulSet) buildNewAndOldObject() error {
	// build old statefulSet
	oldStatefulSets, err := s.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get statefulSets for %s", s.config.namespace)
	}
	for _, oldStatefulSet := range oldStatefulSets.(*appsv1.StatefulSetList).Items {
		s.oldStatefulSets[oldStatefulSet.GetName()] = oldStatefulSet
	}

	// build new statefulSet
	for _, newStatefulSet := range s.statefulSets {
		s.newStatefulSets[newStatefulSet.GetName()] = newStatefulSet.StatefulSet
	}

	return nil
}

// add adds the statefulSet
func (s *StatefulSet) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, statefulSet := range s.statefulSets {
		if _, ok := s.oldStatefulSets[statefulSet.GetName()]; !ok {
			s.deployer.Deployer.AddComponent(horizonapi.StatefulSetComponent, statefulSet)
			isAdded = true
		} else {
			_, err := s.patch(statefulSet, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch statefulSet")
			}
		}
	}
	if isAdded && !s.config.dryRun {
		err := s.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy statefulSet in %s", s.config.namespace)
		}
	}
	return false, nil
}

// get gets the statefulSet
func (s *StatefulSet) get(name string) (interface{}, error) {
	return util.GetStatefulSet(s.config.kubeClient, s.config.namespace, name)
}

// list lists all the statefulSets
func (s *StatefulSet) list() (interface{}, error) {
	return util.ListStatefulSets(s.config.kubeClient, s.config.namespace, s.config.labelSelector)
}

// delete deletes the statefulSet
func (s *StatefulSet) delete(name string) error {
	log.Infof("deleting the statefulSet %s in %s namespace", name, s.config.namespace)
	return util.DeleteStatefulSet(s.config.kubeClient, s.config.namespace, name)
}

// remove removes the statefulSet
func (s *StatefulSet) remove() error {
	// compare the old and new statefulSet and delete if needed
	for _, oldStatefulSet := range s.oldStatefulSets {
		if _, ok := s.newStatefulSets[oldStatefulSet.GetName()]; !ok {
			err := s.delete(oldStatefulSet.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete statefulSet %s in namespace %s", oldStatefulSet.GetName(), s.config.namespace)
			}
		}
	}
	return nil
}

// statefulSetComparator used to compare statefulSet attributes
type statefulSetComparator struct {
	Image    string
	Replicas *int32
	MinCPU   *resource.Quantity
	MaxCPU   *resource.Quantity
	MinMem   *resource.Quantity
	MaxMem   *resource.Quantity
	EnvFrom  []corev1.EnvFromSource
}

// patch patches the statefulSet
func (s *StatefulSet) patch(rc interface{}, isPatched bool) (bool, error) {
	statefulSet := rc.(*components.StatefulSet)
	// check isPatched, why?
	// if there is any configuration change, irrespective of comparing any changes, patch the statefulSet
	if isPatched && !s.config.dryRun {
		log.Infof("updating the statefulSet %s in %s namespace", statefulSet.GetName(), s.config.namespace)
		err := util.PatchStatefulSet(s.config.kubeClient, s.oldStatefulSets[statefulSet.GetName()], *s.newStatefulSets[statefulSet.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch statefulSet %s in namespace %s", statefulSet.GetName(), s.config.namespace)
		}
		return false, nil
	}

	// check whether the statefulSet or its container got changed
	isChanged := false
	for _, oldContainer := range s.oldStatefulSets[statefulSet.GetName()].Spec.Template.Spec.Containers {
		for _, newContainer := range s.newStatefulSets[statefulSet.GetName()].Spec.Template.Spec.Containers {
			if strings.EqualFold(oldContainer.Name, newContainer.Name) && !s.config.dryRun &&
				(!reflect.DeepEqual(
					deploymentComparator{
						Image:    oldContainer.Image,
						Replicas: s.oldStatefulSets[statefulSet.GetName()].Spec.Replicas,
						MinCPU:   oldContainer.Resources.Requests.Cpu(),
						MaxCPU:   oldContainer.Resources.Limits.Cpu(),
						MinMem:   oldContainer.Resources.Requests.Memory(),
						MaxMem:   oldContainer.Resources.Limits.Memory(),
						EnvFrom:  oldContainer.EnvFrom,
					},
					deploymentComparator{
						Image:    newContainer.Image,
						Replicas: s.newStatefulSets[statefulSet.GetName()].Spec.Replicas,
						MinCPU:   newContainer.Resources.Requests.Cpu(),
						MaxCPU:   newContainer.Resources.Limits.Cpu(),
						MinMem:   newContainer.Resources.Requests.Memory(),
						MaxMem:   newContainer.Resources.Limits.Memory(),
						EnvFrom:  newContainer.EnvFrom,
					}) ||
					!reflect.DeepEqual(sortEnvs(oldContainer.Env), sortEnvs(newContainer.Env)) ||
					!reflect.DeepEqual(sortVolumeMounts(oldContainer.VolumeMounts), sortVolumeMounts(newContainer.VolumeMounts)) ||
					!compareVolumes(sortVolumes(s.oldStatefulSets[statefulSet.GetName()].Spec.Template.Spec.Volumes), sortVolumes(s.newStatefulSets[statefulSet.GetName()].Spec.Template.Spec.Volumes))) {
				isChanged = true
				break
			}
		}
		if isChanged {
			break
		}
	}

	// if there is any change from the above step, patch the statefulSet
	if isChanged {
		log.Infof("updating the statefulSet %s in %s namespace", statefulSet.GetName(), s.config.namespace)
		err := util.PatchStatefulSet(s.config.kubeClient, s.oldStatefulSets[statefulSet.GetName()], *s.newStatefulSets[statefulSet.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch rc %s to kube in namespace %s", statefulSet.GetName(), s.config.namespace)
		}
	}
	return false, nil
}
