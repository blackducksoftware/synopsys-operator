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
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/kubernetes/scheme"
	//apijson "k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"

	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ensurePolaris ensures the Polaris instance in the cluster (creates/updates it)
// and ensures the Secret that stores the Polaris object specification/
// This function requires that the global 'namespace' variable is set
func ensurePolaris(polarisObj *polaris.Polaris, isUpdate bool, createOrganization bool) error {
	oldPolaris, err := getPolarisFromSecret()
	if err != nil {
		return err
	}

	if !isUpdate && oldPolaris != nil {
		return fmt.Errorf("the polaris instance already exist")
	}

	// Delete reporting components if it is being disabled
	if oldPolaris != nil && (oldPolaris.EnableReporting && !polarisObj.EnableReporting) {
		log.Info("Deleting Polaris Reporting components...")
		reportingComponents, err := polaris.GetPolarisReportingComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}
		var content []byte
		for _, v := range reportingComponents {
			reportingComponentsByte, err := json.Marshal(v)
			if err != nil {
				return err
			}
			content = append(content, reportingComponentsByte...)
		}
		out, err := RunKubeCmdWithStdin(restconfig, kubeClient, string(content), "delete", "-f", "-")
		if err != nil {
			return fmt.Errorf("couldn't disable reporting |  %+v - %s", out, err)
		}
	}

	// Get components
	type deploy struct {
		name string
		obj  map[string]runtime.Object
	}

	var deployments []deploy

	//Polaris pull secret, license
	mainSecretComponents, err := polaris.GetPolarisBaseSecrets(*polarisObj)
	if err != nil {
		return err
	}
	deployments = append(deployments, deploy{name: "Polaris Licences", obj: mainSecretComponents})

	// Polaris DB
	dbComponents, err := polaris.GetPolarisDBComponents(baseURL, *polarisObj)
	if err != nil {
		return err
	}
	deployments = append(deployments, deploy{name: "Polaris DB", obj: dbComponents})

	// Polaris Core
	polarisComponents, err := polaris.GetPolarisComponents(baseURL, *polarisObj)
	if err != nil {
		return err
	}
	deployments = append(deployments, deploy{name: "Polaris Core", obj: polarisComponents})

	// Reporting
	if polarisObj.EnableReporting {
		reportingComponents, err := polaris.GetPolarisReportingComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}
		deployments = append(deployments, deploy{name: "Polaris Reporting", obj: reportingComponents})
	}

	// Organization provision
	// Only deploy provision components during create operation. This job shouldn't run during update.
	if createOrganization {
		if err := kubeClient.BatchV1().Jobs(namespace).Delete("organization-provision-job", &metav1.DeleteOptions{}); err != nil && !apierrs.IsNotFound(err) {
			return err
		}
		provisionComponents, err := polaris.GetPolarisProvisionComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}

		deployments = append(deployments, deploy{name: "Polaris Organization Provision", obj: provisionComponents})
	}

	// Jobs cannot be patched / updated
	// Remove existing jobs from deployment components
	jobList, err := kubeClient.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprint("polaris.synopsys.com/environment=", polarisObj.Namespace),
	})
	if err != nil {
		return err
	}

	for _, job := range jobList.Items {
		for _, d := range deployments {
			delete(d.obj, fmt.Sprintf("Job.%s", job.Name))
		}
	}

	// Marshal Polaris
	polarisByte, err := json.Marshal(polarisObj)
	if err != nil {
		return err
	}

	polarisSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "polaris",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"polaris": polarisByte,
		},
	}

	// Create/update secret
	if oldPolaris == nil {
		polarisSecret, err = kubeClient.CoreV1().Secrets(namespace).Create(polarisSecret)
		if err != nil {
			return err
		}
	} else {
		_, err = kubeClient.CoreV1().Secrets(namespace).Update(polarisSecret)
		if err != nil {
			return err
		}
	}

	// Start deployment
	for _, v := range deployments {
		if len(v.obj) == 0 {
			log.Infof("Skipping %s", v.name)
			continue
		}

		log.Infof("Deploying %s", v.name)
		if err := deployObj(v.obj); err != nil {
			return err
		}
	}

	// Delete old resources if the version changed
	// TODO we need to find a better solution
	//if oldPolaris != nil && (oldPolaris.Version != polarisObj.Version) {
	//	log.Info("Deleting old resources...")
	//	if err := cleanupByLabel(namespace, fmt.Sprintf("polaris.synopsys.com/version=%s", oldPolaris.Version)); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func deployObj(objs map[string]runtime.Object) error {
	cl, err := client.New(restconfig, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}

	ctx := context.TODO()
	accessor := meta.NewAccessor()
	for _, v := range objs {
		kind, _ := accessor.Kind(v)
		name, _ := accessor.Name(v)
		gvk := v.GetObjectKind().GroupVersionKind()

		key, err := client.ObjectKeyFromObject(v)
		if err != nil {
			return err
		}

		currentRuntimeObject := v.DeepCopyObject()
		if err := cl.Get(ctx, key, currentRuntimeObject); err != nil {
			if !apierrs.IsNotFound(err) {
				return err
			}
			log.Infof("\tDeploying %s %s", kind, name)
			if err := cl.Create(ctx, v); err != nil {
				return err
			}
		} else {
			currentRuntimeObject.GetObjectKind().SetGroupVersionKind(gvk)
			// Skip PVC updates
			switch currentRuntimeObject.(type) {
			case *corev1.PersistentVolumeClaim, *v1.Job:
				continue
			default:
				if !equality.Semantic.DeepEqual(currentRuntimeObject, v) {
					rawDesiredRuntimeObjectInBytes, _ := yaml.Marshal(v)
					log.Infof("\tPatching %s %s", kind, name)
					if err := cl.Patch(ctx, currentRuntimeObject, client.ConstantPatch(types.ApplyPatchType, rawDesiredRuntimeObjectInBytes)); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func deleteObjs(objs map[string]runtime.Object) error {
	cl, err := client.New(restconfig, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}

	ctx := context.TODO()
	accessor := meta.NewAccessor()
	for _, v := range objs {
		kind, _ := accessor.Kind(v)
		name, _ := accessor.Name(v)

		if err := cl.Delete(ctx, v, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			if !apierrs.IsNotFound(err) {
				return err
			}
		} else {
			log.Infof("Deleted %s %s", kind, name)
		}
	}

	return nil
}
