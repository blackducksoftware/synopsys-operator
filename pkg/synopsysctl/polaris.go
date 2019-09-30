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

	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ensurePolaris esnures the Polaris instance in the cluster (creates/updates it)
// and esnures the Secret that stores the Polaris object specification/
// This function requires that the global 'namespace' variable is set
func ensurePolaris(polarisObj *polaris.Polaris, isUpdate bool) error {
	oldPolaris, err := getPolarisFromSecret()
	if err != nil {
		return err
	}

	if !isUpdate && oldPolaris != nil {
		return fmt.Errorf("the polaris instance already exist")
	}

	// Delete old polaris jobs
	jobList, err := kubeClient.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprint("polaris.synopsys.com/environment=", polarisObj.Namespace),
	})
	if err != nil {
		return err
	}

	for _, job := range jobList.Items {
		log.Infof("Deleting job %s", job.Name)
		if err := kubeClient.BatchV1().Jobs(namespace).Delete(job.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
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

	dbComponents, err := polaris.GetPolarisDBComponents(baseURL, *polarisObj)
	if err != nil {
		return err
	}
	deployments = append(deployments, deploy{name: "Polaris DB", obj: dbComponents})

	polarisComponents, err := polaris.GetPolarisComponents(baseURL, *polarisObj)
	if err != nil {
		return err
	}
	deployments = append(deployments, deploy{name: "Polaris Core", obj: polarisComponents})

	if polarisObj.EnableReporting {
		reportingComponents, err := polaris.GetPolarisReportingComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}
		deployments = append(deployments, deploy{name: "Polaris Reporting", obj: reportingComponents})
	}

	provisionComponents, err := polaris.GetPolarisProvisionComponents(baseURL, *polarisObj)
	if err != nil {
		return err
	}

	deployments = append(deployments, deploy{name: "Polaris Organization Provision", obj: provisionComponents})

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
		log.Infof("Deploying %s", v.name)
		var content []byte
		for _, v := range v.obj {
			polarisComponentsByte, err := json.Marshal(v)
			if err != nil {
				return err
			}
			content = append(content, polarisComponentsByte...)
		}

		out, err := RunKubeCmdWithStdin(restconfig, kubeClient, string(content), "apply", "--validate=false", "-f", "-")
		if err != nil {
			if oldPolaris == nil {
				kubeClient.CoreV1().Secrets(namespace).Delete("polaris", &metav1.DeleteOptions{})
			}
			return fmt.Errorf("couldn't deploy polaris |  %+v - %s", out, err)
		}
	}

	// Delete old resources if the version changed
	if oldPolaris != nil && (oldPolaris.Version != polarisObj.Version) {
		log.Info("Deleting old resources...")
		if err := cleanupByLabel(namespace, fmt.Sprintf("polaris.synopsys.com/version=%s", oldPolaris.Version)); err != nil {
			return err
		}
	}
	return nil
}
