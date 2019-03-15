/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package soperator

import (
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func UpdateSynopsysOperator(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, oldSOperatorSpec SOperatorSpecConfig, newSOperatorSpec SOperatorSpecConfig) error {
	currSOperatorComponents, err := oldSOperatorSpec.GetComponents()
	fmt.Printf("%+v\n", currSOperatorComponents)

	// Get Components of New Synopsys-Operator

	newSOperatorComponents, err := newSOperatorSpec.GetComponents()
	fmt.Printf("%+v\n", newSOperatorComponents)

	// Update S-O ConfigMap if necessary
	isConfigMapUpdated, err := crdupdater.UpdateConfigMap(kubeClient, namespace, "synopsys-operator", newSOperatorComponents.ConfigMaps[0])

	// Update S-O Secret if necessary
	isSecretUpdated, err := crdupdater.UpdateSecret(kubeClient, namespace, "blackduck-secret", newSOperatorComponents.Secrets[0])

	operatorUpdater := crdupdater.NewUpdater()

	// Update S-O ReplicationController if necessary
	replicationControllerUpdater, err := crdupdater.NewReplicationController(restconfig, kubeClient, newSOperatorComponents.ReplicationControllers, namespace, "app=opssight", isConfigMapUpdated || isSecretUpdated)
	operatorUpdater.AddUpdater(replicationControllerUpdater)

	// Update S-O Service if necessary
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newSOperatorComponents.Services, namespace, "app=opssight")
	operatorUpdater.AddUpdater(serviceUpdater)

	// Update S-O ServiceAccount if necessary

	// Update S-O ClusterRoleBinding if necessary
	clusterRoleBindingUpdater, err := crdupdater.NewClusterRoleBinding(restconfig, kubeClient, newSOperatorComponents.ClusterRoleBindings, namespace, "app=opssight")
	operatorUpdater.AddUpdater(clusterRoleBindingUpdater)

	err = operatorUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func UpdatePrometheus(restconfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, currPrometheusSpecConfig PrometheusSpecConfig, newPrometheusSpecConfig PrometheusSpecConfig) error {
	currPrometheusComponents, err := currPrometheusSpecConfig.GetComponents()
	fmt.Printf("%+v\n", currPrometheusComponents)

	// Get Components of New Prometheus

	newPrometheusComponents, err := newPrometheusSpecConfig.GetComponents()
	fmt.Printf("%+v\n", newPrometheusComponents)

	prometheusUpdater := crdupdater.NewUpdater()

	// Update Prometheus ConfigMap
	_, err = crdupdater.UpdateConfigMap(kubeClient, namespace, "prometheus", newPrometheusComponents.ConfigMaps[0])

	// Update Prometheus Deployment
	deploymentUpdater, err := crdupdater.NewDeployment(restconfig, kubeClient, newPrometheusComponents.Deployments, namespace, "app=prometheus", false)
	prometheusUpdater.AddUpdater(deploymentUpdater)

	// Update Prometheus Service
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newPrometheusComponents.Services, namespace, "app=prometheus")
	prometheusUpdater.AddUpdater(serviceUpdater)

	err = prometheusUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
