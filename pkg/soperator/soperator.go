/*
Copyright (C) 2019 Synopsys, Inc.

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
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SpecConfig represents the SOperator component
// Its methods include GetComponents() and any functions
// that create Kubernetes Resources for the Synopsys-Operator
type SpecConfig struct {
	Namespace                     string
	Image                         string
	Expose                        string
	SecretType                    horizonapi.SecretType
	AdminPassword                 string
	PostgresPassword              string
	UserPassword                  string
	BlackduckPassword             string
	OperatorTimeBombInSeconds     int64
	DryRun                        bool
	LogLevel                      string
	Threadiness                   int
	PostgresRestartInMins         int64
	PodWaitTimeoutSeconds         int64
	ResyncIntervalInSeconds       int64
	TerminationGracePeriodSeconds int64
	SealKey                       string
	RestConfig                    *rest.Config
	KubeClient                    *kubernetes.Clientset
}

// NewSOperator will create a SOperator type
func NewSOperator(namespace, synopsysOperatorImage, expose, adminPassword, postgresPassword, userPassword, blackduckpassword string,
	secretType horizonapi.SecretType, operatorTimeBombInSeconds int64, dryRun bool, logLevel string, threadiness int,
	postgresRestartInMins int64, podWaitTimeoutSeconds int64, resyncIntervalInSeconds int64, terminationGracePeriodSeconds int64, sealKey string, restConfig *rest.Config, kubeClient *kubernetes.Clientset) *SpecConfig {
	return &SpecConfig{
		Namespace:                     namespace,
		Image:                         synopsysOperatorImage,
		Expose:                        expose,
		SecretType:                    secretType,
		AdminPassword:                 adminPassword,
		PostgresPassword:              postgresPassword,
		UserPassword:                  userPassword,
		BlackduckPassword:             blackduckpassword,
		OperatorTimeBombInSeconds:     operatorTimeBombInSeconds,
		DryRun:                        dryRun,
		LogLevel:                      logLevel,
		Threadiness:                   threadiness,
		PostgresRestartInMins:         postgresRestartInMins,
		PodWaitTimeoutSeconds:         podWaitTimeoutSeconds,
		ResyncIntervalInSeconds:       resyncIntervalInSeconds,
		TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
		SealKey:                       sealKey,
		RestConfig:                    restConfig,
		KubeClient:                    kubeClient,
	}
}

// GetComponents will return a ComponentList representing all
// Kubernetes Resources for the Synopsys-Operator
func (specConfig *SpecConfig) GetComponents() (*api.ComponentList, error) {
	configMap, err := specConfig.GetOperatorConfigMap()
	if err != nil {
		return nil, err
	}
	components := &api.ComponentList{
		ReplicationControllers: []*components.ReplicationController{
			specConfig.GetOperatorReplicationController(),
		},
		Services:   specConfig.GetOperatorService(),
		ConfigMaps: []*components.ConfigMap{configMap},
		ServiceAccounts: []*components.ServiceAccount{
			specConfig.GetOperatorServiceAccount(),
		},
		ClusterRoleBindings: []*components.ClusterRoleBinding{
			specConfig.GetOperatorClusterRoleBinding(),
		},
		ClusterRoles: []*components.ClusterRole{
			specConfig.GetOperatorClusterRole(),
		},
		Secrets: []*components.Secret{
			specConfig.GetOperatorSecret(), specConfig.GetTLSCertificateSecret(),
		},
	}
	return components, nil
}
