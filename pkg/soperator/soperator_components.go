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
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SpecConfig represents the SOperator component
// Its methods include GetComponents() and any functions
// that create Kubernetes Resources for Synopsys Operator
type SpecConfig struct {
	Namespace                     string
	Image                         string
	Expose                        string
	ClusterType                   ClusterType
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
	Certificate                   string
	CertificateKey                string
	IsClusterScoped               bool
	Crds                          map[string]string
}

// NewSOperator will create a SOperator type
func NewSOperator(namespace, synopsysOperatorImage, expose string, clusterType ClusterType, operatorTimeBombInSeconds int64, dryRun bool, logLevel string, threadiness int, postgresRestartInMins int64,
	podWaitTimeoutSeconds int64, resyncIntervalInSeconds int64, terminationGracePeriodSeconds int64, sealKey string, restConfig *rest.Config,
	kubeClient *kubernetes.Clientset, certificate string, certificateKey string, isClusterScoped bool, crds map[string]string) *SpecConfig {
	return &SpecConfig{
		Namespace:                     namespace,
		Image:                         synopsysOperatorImage,
		Expose:                        expose,
		ClusterType:                   clusterType,
		OperatorTimeBombInSeconds:     operatorTimeBombInSeconds,
		DryRun:                        dryRun,
		LogLevel:                      logLevel,
		Threadiness:                   threadiness,
		PostgresRestartInMins:         postgresRestartInMins,
		PodWaitTimeoutSeconds:         podWaitTimeoutSeconds,
		ResyncIntervalInSeconds:       resyncIntervalInSeconds,
		TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
		SealKey:                       sealKey,
		Certificate:                   certificate,
		CertificateKey:                certificateKey,
		IsClusterScoped:               isClusterScoped,
		Crds:                          crds,
	}
}

// ClusterType represents the cluster type
type ClusterType string

// Constants for the PrintFormats
const (
	KubernetesClusterType ClusterType = "KUBERNETES"
	OpenshiftClusterType  ClusterType = "OPENSHIFT"
)

// GetComponents will return a ComponentList representing all
// Kubernetes Resources for Synopsys Operator
func (specConfig *SpecConfig) GetComponents() (*api.ComponentList, error) {
	configMap, err := specConfig.GetOperatorConfigMap()
	if err != nil {
		return nil, errors.Trace(err)
	}

	deployment, err := specConfig.GetOperatorDeployment()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components := &api.ComponentList{
		CustomResourceDefinitions: specConfig.GetCrds(),
		Deployments:               []*components.Deployment{deployment},
		Services:                  specConfig.GetOperatorService(),
		ConfigMaps:                []*components.ConfigMap{configMap},
		ServiceAccounts:           []*components.ServiceAccount{specConfig.GetOperatorServiceAccount()},
		Secrets:                   []*components.Secret{specConfig.GetOperatorSecret(), specConfig.GetTLSCertificateSecret()},
	}

	if specConfig.IsClusterScoped {
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, specConfig.GetOperatorClusterRoleBinding())
		components.ClusterRoles = append(components.ClusterRoles, specConfig.GetOperatorClusterRole())
	} else {
		components.RoleBindings = append(components.RoleBindings, specConfig.GetOperatorRoleBinding())
		components.Roles = append(components.Roles, specConfig.GetOperatorRole())
	}

	// Add routes for OpenShift
	route := specConfig.GetOpenShiftRoute()
	log.Debugf("Route: %+v", route)
	if route != nil {
		components.Routes = []*api.Route{route}
	}
	return components, nil
}
