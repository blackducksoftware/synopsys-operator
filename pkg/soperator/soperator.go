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
)

// SpecConfig represents the SOperator component
// Its methods include GetComponents() and any functions
// that create Kubernetes Resources for the Synopsys-Operator
type SpecConfig struct {
	Namespace                     string
	SynopsysOperatorImage         string
	SecretType                    horizonapi.SecretType
	SecretAdminPassword           string
	SecretPostgresPassword        string
	SecretUserPassword            string
	SecretBlackduckPassword       string
	TerminationGracePeriodSeconds int64
	SealKey                       string
}

// NewSOperator will create a SOperator type
func NewSOperator(namespace, synopsysOperatorImage, blackduckRegistrationKey, secretName, adminPassword, postrgresPassword, userPassword,
	blackduckpassword string, secretType horizonapi.SecretType, terminationGracePeriodSeconds int64, sealKey string) *SpecConfig {
	return &SpecConfig{
		Namespace:                     namespace,
		SynopsysOperatorImage:         synopsysOperatorImage,
		SecretType:                    secretType,
		SecretAdminPassword:           adminPassword,
		SecretPostgresPassword:        postrgresPassword,
		SecretUserPassword:            userPassword,
		SecretBlackduckPassword:       blackduckpassword,
		TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
		SealKey:                       sealKey,
	}
}

// GetComponents will return a ComponentList representing all
// Kubernetes Resources for the Synopsys-Operator
func (specConfig *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{
		ReplicationControllers: []*components.ReplicationController{
			specConfig.GetOperatorReplicationController(),
		},
		Services: []*components.Service{
			specConfig.GetOperatorService(),
		},
		ConfigMaps: []*components.ConfigMap{
			specConfig.GetOperatorConfigMap(),
		},
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
