/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package polaris

import (
	"encoding/json"
	"fmt"
	flying_dutchman "github.com/blackducksoftware/synopsys-operator/flying-dutchman"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PolarisReconciler reconciles a Polaris object
type PolarisReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Log         logr.Logger
	IsOpenShift bool
	IsDryRun    bool
	BaseUrl     string
}

func (r *PolarisReconciler) GetClient() client.Client {
	return r.Client
}

func (r *PolarisReconciler) GetRuntimeObjects(cr interface{}) (map[string]runtime.Object, error) {
	polarisSecret := cr.(*corev1.Secret)
	polarisSecretBytes, ok := polarisSecret.Data["polaris"]
	if !ok {
		return nil, fmt.Errorf("polaris entry is missing in the secret")
	}

	var p *Polaris
	if err := json.Unmarshal(polarisSecretBytes, &p); err != nil {
		return nil, err
	}

	mapOfUniqueIdToDesiredRuntimeObject := GetComponents(r.BaseUrl, *p)
	if !r.IsDryRun {
		for _, desiredRuntimeObject := range mapOfUniqueIdToDesiredRuntimeObject {
			// set an owner reference
			if err := ctrl.SetControllerReference(polarisSecret, desiredRuntimeObject.(metav1.Object), r.Scheme); err != nil {
				// requeue if we cannot set owner on the object
				// TODO: change this to requeue, and only not requeue when we get "newAlreadyOwnedError", i.e: if it's already owned by our CR
				//return ctrl.Result{}, err
				return mapOfUniqueIdToDesiredRuntimeObject, nil
			}
		}
	}

	return mapOfUniqueIdToDesiredRuntimeObject, nil
}

func (r *PolarisReconciler) GetInstructionManual(cr interface{}) (*flying_dutchman.RuntimeObjectDependencyYaml, error) {
	polarisSecret := cr.(*corev1.Secret)
	polarisSecretBytes, ok := polarisSecret.Data["polaris"]
	if !ok {
		return nil, fmt.Errorf("polaris entry is missing in the secret")
	}

	var p *Polaris
	if err := json.Unmarshal(polarisSecretBytes, &p); err != nil {
		return nil, err
	}

	mamual, err := GetBaseYaml(r.BaseUrl, "polaris", p.Version, "manual.yaml")
	if err != nil {
		return nil, err
	}

	dependencyYamlStruct := &flying_dutchman.RuntimeObjectDependencyYaml{}

	err = yaml.Unmarshal([]byte(mamual), dependencyYamlStruct)
	if err != nil {
		return nil, err
	}
	return dependencyYamlStruct, nil
}
