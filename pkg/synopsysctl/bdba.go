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

package synopsysctl

import (
	"encoding/json"
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/bdba"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ensureBDBA ensures the BDBA instance in the cluster (creates/updates it)
// and ensures the Secret that stores the BDBA object specification/
// This function requires that the global 'namespace' variable is set
func ensureBDBA(bdbaObj *bdba.BDBA, isUpdate bool, createOrganization bool) error {
	oldBDBA, err := getBDBAFromSecret()
	if err != nil {
		return err
	}

	if !isUpdate && oldBDBA != nil {
		return fmt.Errorf("the BDBA instance already exist")
	}

	// Get Runtime Objects for BDBA Components
	type runtimeobjectsPerComponent struct {
		componentName  string
		rutnimeobjects map[string]runtime.Object
	}

	var bdbaComponentRuntimeObjects []runtimeobjectsPerComponent

	// add the Core BDBA Runtime Objects
	bdbaRuntimeObjects, err := bdba.GetComponents(baseURL, *bdbaObj)
	if err != nil {
		return err
	}
	bdbaComponentRuntimeObjects = append(bdbaComponentRuntimeObjects, runtimeobjectsPerComponent{componentName: "BDBA Core", rutnimeobjects: bdbaRuntimeObjects})

	// Marshal BDBA and write for the Secret
	bdbaByte, err := json.Marshal(bdbaObj)
	if err != nil {
		return err
	}
	bdbaSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bdba",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"bdba": bdbaByte,
		},
	}

	// Create/update the BDBA Secret
	if oldBDBA == nil {
		bdbaSecret, err = kubeClient.CoreV1().Secrets(namespace).Create(bdbaSecret)
		if err != nil {
			return err
		}
	} else {
		_, err = kubeClient.CoreV1().Secrets(namespace).Update(bdbaSecret)
		if err != nil {
			return err
		}
	}

	// Get PVCs that are already deployed
	existingPVCList, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Get Jobs that are already deployed
	existingJobList, err := kubeClient.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	accessor := meta.NewAccessor()

	// Deploy the RuntimeObjects for each BDBA deployment
	for _, bdbaComponent := range bdbaComponentRuntimeObjects {
		if len(bdbaComponent.rutnimeobjects) == 0 {
			log.Infof("Skipping %s", bdbaComponent.componentName)
			continue
		}

		log.Infof("Deploying %s", bdbaComponent.componentName)
		var polarisComponentBytes []byte

	NEXTOBJECT: // convert each runtimeObject to bytes and add to polarisComponentBytes
		for _, runtimeobject := range bdbaComponent.rutnimeobjects {
			// Skip PVCs / Jobs that already exist - they cannot be patched / updated
			switch runtimeobject.(type) {
			case *corev1.PersistentVolumeClaim:
				for _, pvc := range existingPVCList.Items {
					name, _ := accessor.Name(runtimeobject)
					if pvc.Name == name {
						log.Debugf("PVC %s already exists", name)
						continue NEXTOBJECT
					}
				}
			case *batchv1.Job:
				for _, job := range existingJobList.Items {
					name, _ := accessor.Name(runtimeobject)
					if job.Name == name {
						log.Debugf("Job %s already exists", name)
						continue NEXTOBJECT
					}
				}
			}

			polarisRuntimeObjectBytes, err := json.Marshal(runtimeobject)
			if err != nil {
				return err
			}
			polarisComponentBytes = append(polarisComponentBytes, polarisRuntimeObjectBytes...)
		}

		// deploy the bdba component
		out, err := RunKubeCmdWithStdin(restconfig, kubeClient, string(polarisComponentBytes), "apply", "--validate=false", "-f", "-")
		if err != nil {
			if oldBDBA == nil {
				kubeClient.CoreV1().Secrets(namespace).Delete("bdba", &metav1.DeleteOptions{})
			}
			return fmt.Errorf("couldn't deploy bdba |  %+v - %s", out, err)
		}
	}

	return nil
}
