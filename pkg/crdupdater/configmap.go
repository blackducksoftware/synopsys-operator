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
)

// UpdateConfigMap updates the config map by comparing the old and new config map data
func UpdateConfigMap(kubeClient *kubernetes.Clientset, namespace string, configMapName string, newConfig *components.ConfigMap) (bool, error) {
	newConfigMapKube, err := newConfig.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert configmap %s to kube in namespace %s", configMapName, namespace)
	}
	newConfigMap := newConfigMapKube.(*corev1.ConfigMap)
	newConfigMapData := newConfigMap.Data

	// getting old configmap data
	oldConfigMap, err := util.GetConfigMap(kubeClient, namespace, configMapName)
	if err != nil {
		return false, errors.Annotatef(err, "unable to find the configmap %s in namespace %s", configMapName, namespace)
	}
	oldConfigMapData := oldConfigMap.Data

	// compare for difference between old and new configmap data, if changed update the configmap
	if !reflect.DeepEqual(newConfigMapData, oldConfigMapData) {
		oldConfigMap.Data = newConfigMapData
		err = util.UpdateConfigMap(kubeClient, namespace, oldConfigMap)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the configmap %s in namespace %s", configMapName, namespace)
		}
		return true, nil
	}
	return false, nil
}
