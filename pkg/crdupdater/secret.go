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

// UpdateSecret updates the config map by comparing the old and new secret data
func UpdateSecret(kubeClient *kubernetes.Clientset, namespace string, secretName string, newConfig *components.Secret) (bool, error) {
	newSecretKube, err := newConfig.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert secret %s to kube in namespace %s", secretName, namespace)
	}
	newSecret := newSecretKube.(*corev1.Secret)
	newSecretData := newSecret.Data
	newSecretStringData := newSecret.StringData

	// getting old secret data
	oldSecret, err := util.GetSecret(kubeClient, namespace, secretName)
	if err != nil {
		return false, errors.Annotatef(err, "unable to find the secret %s in namespace %s", secretName, namespace)
	}
	oldSecretData := oldSecret.Data
	oldSecretStringData := oldSecret.StringData

	// compare for difference between old and new secret data, if changed update the secret
	if !reflect.DeepEqual(newSecretData, oldSecretData) || !reflect.DeepEqual(newSecretStringData, oldSecretStringData) {
		oldSecret.Data = newSecretData
		oldSecret.StringData = newSecretStringData
		err = util.UpdateSecret(kubeClient, namespace, oldSecret)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the secret %s in namespace %s", secretName, namespace)
		}
		return true, nil
	}
	return false, nil
}
