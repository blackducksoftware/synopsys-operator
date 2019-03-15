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

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ConfigMap stores the configuration to add or delete the config map object
type ConfigMap struct {
	kubeConfig    *rest.Config
	kubeClient    *kubernetes.Clientset
	deployer      *util.DeployerHelper
	namespace     string
	configMaps    []*components.ConfigMap
	labelSelector string
	oldConfigMaps map[string]*corev1.ConfigMap
	newConfigMaps map[string]*corev1.ConfigMap
}

// NewConfigMap returns the config map
func NewConfigMap(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, configMaps []*components.ConfigMap,
	namespace string, labelSelector string) (*ConfigMap, error) {
	deployer, err := util.NewDeployer(kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	return &ConfigMap{
		kubeConfig:    kubeConfig,
		kubeClient:    kubeClient,
		deployer:      deployer,
		namespace:     namespace,
		configMaps:    configMaps,
		labelSelector: labelSelector,
		oldConfigMaps: make(map[string]*corev1.ConfigMap, 0),
		newConfigMaps: make(map[string]*corev1.ConfigMap, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new config map
func (c *ConfigMap) buildNewAndOldObject() error {
	// build old config map
	oldConfigMaps, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get config maps for %s", c.namespace)
	}
	for _, oldConfigMap := range oldConfigMaps.(*corev1.ConfigMapList).Items {
		c.oldConfigMaps[oldConfigMap.GetName()] = &oldConfigMap
	}

	// build new config map
	for _, newConfigMap := range c.configMaps {
		newConfigMapKube, err := newConfigMap.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert config map %s to kube %s", newConfigMap.GetName(), c.namespace)
		}
		c.newConfigMaps[newConfigMap.GetName()] = newConfigMapKube.(*corev1.ConfigMap)
	}

	return nil
}

// add adds the config map
func (c *ConfigMap) add() error {
	isAdded := false
	for _, configMap := range c.configMaps {
		if _, ok := c.oldConfigMaps[configMap.GetName()]; !ok {
			c.deployer.Deployer.AddConfigMap(configMap)
			isAdded = true
		} else {
			err := c.patch(configMap)
			if err != nil {
				return errors.Annotatef(err, "patch config map:")
			}
		}
	}
	if isAdded {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return errors.Annotatef(err, "unable to deploy config map in %s", c.namespace)
		}
	}
	return nil
}

// list lists all the config maps
func (c *ConfigMap) list() (interface{}, error) {
	return util.ListConfigMaps(c.kubeClient, c.namespace, c.labelSelector)
}

// delete deletes the config map
func (c *ConfigMap) delete(name string) error {
	return util.DeleteConfigMap(c.kubeClient, c.namespace, name)
}

// remove removes the config map
func (c *ConfigMap) remove() error {
	// compare the old and new config map and delete if needed
	for _, oldConfigMap := range c.oldConfigMaps {
		if _, ok := c.newConfigMaps[oldConfigMap.GetName()]; !ok {
			err := c.delete(oldConfigMap.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete config map %s in namespace %s", oldConfigMap.GetName(), c.namespace)
			}
		}
	}
	return nil
}

// patch patches the config map
func (c *ConfigMap) patch(cm interface{}) error {
	configMap := cm.(*components.ConfigMap)
	configMapName := configMap.GetName()
	oldConfigMap := c.oldConfigMaps[configMapName]
	newConfigMap := c.newConfigMaps[configMapName]
	if !reflect.DeepEqual(newConfigMap.Data, oldConfigMap.Data) {
		oldConfigMap.Data = newConfigMap.Data
		err := util.UpdateConfigMap(c.kubeClient, c.namespace, oldConfigMap)
		if err != nil {
			return errors.Annotatef(err, "unable to update the config map %s in namespace %s", configMapName, c.namespace)
		}
	}
	return nil
}

// UpdateConfigMap updates the config map by comparing the old and new config map data
func UpdateConfigMap(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, configMapName string, newConfig *components.ConfigMap) (bool, error) {
	newConfigMapKube, err := newConfig.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert config map %s to kube in namespace %s", configMapName, namespace)
	}
	newConfigMap := newConfigMapKube.(*corev1.ConfigMap)
	newConfigMapData := newConfigMap.Data

	// getting old configmap data
	oldConfigMap, err := util.GetConfigMap(kubeClient, namespace, configMapName)
	if err != nil {
		// if configmap is not present, create the configmap
		deployer, err := util.NewDeployer(kubeConfig)
		deployer.Deployer.AddConfigMap(newConfig)
		err = deployer.Deployer.Run()
		return false, errors.Annotatef(err, "unable to create the config map %s in namespace %s", configMapName, namespace)
	}
	oldConfigMapData := oldConfigMap.Data

	// compare for difference between old and new configmap data, if changed update the configmap
	if !reflect.DeepEqual(newConfigMapData, oldConfigMapData) {
		oldConfigMap.Data = newConfigMapData
		err = util.UpdateConfigMap(kubeClient, namespace, oldConfigMap)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the config map %s in namespace %s", configMapName, namespace)
		}
		return true, nil
	}
	return false, nil
}
