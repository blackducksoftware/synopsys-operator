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

package soperator

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
)

// GetOperatorReplicationController creates a ReplicationController Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorReplicationController() *horizoncomponents.ReplicationController {
	// Add the Replication Controller to the Deployer
	var synopsysOperatorRCReplicas int32 = 1
	synopsysOperatorRC := horizoncomponents.NewReplicationController(horizonapi.ReplicationControllerConfig{
		APIVersion: "v1",
		//ClusterName:  "string",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
		Replicas:  &synopsysOperatorRCReplicas,
		//ReadySeconds: "int32",
	})

	synopsysOperatorRC.AddLabelSelectors(map[string]string{"name": "synopsys-operator"})

	synopsysOperatorPod := horizoncomponents.NewPod(horizonapi.PodConfig{
		APIVersion: "v1",
		//ClusterName:            "string",
		Name:           "synopsys-operator",
		Namespace:      specConfig.Namespace,
		ServiceAccount: "synopsys-operator",
		//RestartPolicy:          "RestartPolicyType",
		//TerminationGracePeriod: "*int64",
		//ActiveDeadline:         "*int64",
		//Node:                   "string",
		//FSGID:                  "*int64",
		//Hostname:               "string",
		//SchedulerName:          "string",
		//DNSPolicy:              "DNSPolicType",
		//PriorityValue:          "*int32",
		//PriorityClass:          "string",
		//SELinux:                "*SELinuxType",
		//RunAsUser:              "*int64",
		//RunAsGroup:             "*int64",
		//ForceNonRoot:           "*bool",
	})

	synopsysOperatorPodLabels := map[string]string{
		"name": "synopsys-operator",
		"app":  "synopsys-operator",
	}

	synopsysOperatorContainer := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name:       "synopsys-operator",
		Args:       []string{"/etc/synopsys-operator/config.json"},
		Command:    []string{"./operator"},
		Image:      specConfig.SynopsysOperatorImage,
		PullPolicy: horizonapi.PullAlways,
		//MinCPU:                   "string",
		//MaxCPU:                   "string",
		//MinMem:                   "string",
		//MaxMem:                   "string",
		//Privileged:               "*bool",
		//AllowPrivilegeEscalation: "*bool",
		//ReadOnlyFS:               "*bool",
		//ForceNonRoot:             "*bool",
		//SELinux:                  "*SELinuxType",
		//UID:                      "*int64",
		//AllocateStdin:            "bool",
		//StdinOnce:                "bool",
		//AllocateTTY:              "bool",
		//WorkingDirectory:         "string",
		//TerminationMsgPath:       "string",
		//TerminationMsgPolicy:     "TerminationMessagePolicyType",
	})
	synopsysOperatorContainer.AddPort(horizonapi.PortConfig{
		//Name:          "string",
		//Protocol:      "ProtocolType",
		//IP:            "string",
		//HostPort:      "string",
		ContainerPort: "8080",
	})
	synopsysOperatorContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/etc/synopsys-operator",
		//Propagation: "*MountPropagationType",
		Name: "synopsys-operator",
		//SubPath:     "string",
		//ReadOnly:    "*bool",
	})

	synopsysOperatorContainerUI := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name: "synopsys-operator-ui",
		//Args:                     "[]string",
		Command:    []string{"./app"},
		Image:      specConfig.SynopsysOperatorImage,
		PullPolicy: horizonapi.PullAlways,
		//MinCPU:                   "string",
		//MaxCPU:                   "string",
		//MinMem:                   "string",
		//MaxMem:                   "string",
		//Privileged:               "*bool",
		//AllowPrivilegeEscalation: "*bool",
		//ReadOnlyFS:               "*bool",
		//ForceNonRoot:             "*bool",
		//SELinux:                  "*SELinuxType",
		//UID:                      "*int64",
		//AllocateStdin:            "bool",
		//StdinOnce:                "bool",
		//AllocateTTY:              "bool",
		//WorkingDirectory:         "string",
		//TerminationMsgPath:       "string",
		//TerminationMsgPolicy:     "TerminationMessagePolicyType",
	})
	synopsysOperatorContainerUI.AddPort(horizonapi.PortConfig{
		//Name:          "string",
		//Protocol:      "ProtocolType",
		//IP:            "string",
		//HostPort:      "string",
		ContainerPort: "3000",
	})
	synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "ADDR",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "0.0.0.0",
		//FromName:     "string",
	})
	synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "PORT",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "3000",
		//FromName:     "string",
	})
	synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "GO_ENV",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "development",
		//FromName:     "string",
	})

	// Create config map volume
	var synopsysOperatorVolumeDefaultMode int32 = 420
	synopsysOperatorVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "synopsys-operator",
		MapOrSecretName: "synopsys-operator",
		//Items:           "map[string]KeyAndMode",
		DefaultMode: &synopsysOperatorVolumeDefaultMode,
		//Required:        "*bool",
	})

	synopsysOperatorPod.AddLabels(synopsysOperatorPodLabels)
	synopsysOperatorPod.AddContainer(synopsysOperatorContainer)
	synopsysOperatorPod.AddContainer(synopsysOperatorContainerUI)
	synopsysOperatorPod.AddVolume(synopsysOperatorVolume)
	synopsysOperatorRC.AddPod(synopsysOperatorPod)

	synopsysOperatorRC.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorRC
}

// GetOperatorService creates a Service Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorService() *horizoncomponents.Service {

	// Add the Service to the Deployer
	synopsysOperatorService := horizoncomponents.NewService(horizonapi.ServiceConfig{
		APIVersion: "v1",
		//ClusterName:              "string",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
		//ExternalName:             "string",
		//IPServiceType:            "ClusterIPServiceType",
		//ClusterIP:                "string",
		//PublishNotReadyAddresses: "bool",
		//TrafficPolicy:            "TrafficPolicyType",
		//Affinity:                 "string",
	})

	synopsysOperatorService.AddSelectors(map[string]string{"name": "synopsys-operator"})
	synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
		Name:       "synopsys-operator-ui",
		Port:       3000,
		TargetPort: "3000",
		//NodePort:   "int32",
		Protocol: horizonapi.ProtocolTCP,
	})
	synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
		Name:       "synopsys-operator-ui-standard-port",
		Port:       80,
		TargetPort: "3000",
		//NodePort:   "int32",
		Protocol: horizonapi.ProtocolTCP,
	})
	synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
		Name:       "synopsys-operator",
		Port:       8080,
		TargetPort: "8080",
		//NodePort:   "int32",
		Protocol: horizonapi.ProtocolTCP,
	})

	synopsysOperatorService.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorService
}

// GetOperatorConfigMap creates a ConfigMap Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorConfigMap() *horizoncomponents.ConfigMap {
	// Config Map
	synopsysOperatorConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		//ClusterName: "string",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
	})

	cmData := map[string]string{}
	cmData["config.json"] = fmt.Sprintf("{\"OperatorTimeBombInSeconds\":\"315576000\", \"DryRun\": false, \"LogLevel\": \"debug\", \"Namespace\": \"%s\", \"Threadiness\": 5, \"PostgresRestartInMins\": 10, \"PodWaitTimeoutSeconds\" : \"600\", \"ResyncIntervalInSeconds\" : \"30\", \"TerminationGracePeriodSeconds\" : \"%d\"}", specConfig.Namespace, specConfig.TerminationGracePeriodSeconds)
	cmData["image"] = specConfig.SynopsysOperatorImage
	synopsysOperatorConfigMap.AddData(cmData)

	synopsysOperatorConfigMap.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorConfigMap
}

// GetOperatorServiceAccount creates a ServiceAccount Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorServiceAccount() *horizoncomponents.ServiceAccount {
	// Service Account
	synopsysOperatorServiceAccount := horizoncomponents.NewServiceAccount(horizonapi.ServiceAccountConfig{
		APIVersion: "v1",
		//ClusterName:    "string",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
		//AutomountToken: "*bool",
	})

	synopsysOperatorServiceAccount.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorServiceAccount
}

// GetOperatorClusterRoleBinding creates a ClusterRoleBinding Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorClusterRoleBinding() *horizoncomponents.ClusterRoleBinding {
	// Cluster Role Binding
	synopsysOperatorClusterRoleBinding := horizoncomponents.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		//ClusterName: "string",
		Name:      "synopsys-operator-admin",
		Namespace: specConfig.Namespace,
	})
	synopsysOperatorClusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind: "ServiceAccount",
		//APIGroup:  "string",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
	})
	synopsysOperatorClusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	synopsysOperatorClusterRoleBinding.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorClusterRoleBinding
}

// GetOperatorClusterRole creates a ClusterRole Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorClusterRole() *horizoncomponents.ClusterRole {
	synopsysOperatorClusterRole := horizoncomponents.NewClusterRole(horizonapi.ClusterRoleConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		//ClusterName : "string,"
		Name:      "synopsys-operator-admin",
		Namespace: specConfig.Namespace,
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups:       []string{"apiextensions.k8s.io"},
		Resources:       []string{"customresourcedefinitions"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups:       []string{"rbac.authorization.k8s.io"},
		Resources:       []string{"clusterrolebindings", "clusterroles"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups:       []string{""},
		Resources:       []string{"namespaces", "pods", "configmaps", "persistentvolumeclaims", "services", "secrets", "replicationcontrollers", "deployments", "statefulsets", "serviceaccounts"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"get", "list", "watch"},
		APIGroups:       []string{""},
		Resources:       []string{"pods/log", "endpoints"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"create"},
		APIGroups:       []string{""},
		Resources:       []string{"pods/exec"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:           []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups:       []string{"synopsys.com"},
		Resources:       []string{"*"},
		ResourceNames:   []string{},
		NonResourceURLs: []string{},
	})

	// Add Openshift rules
	restConfig, err := protoform.GetKubeConfig()
	if err != nil {
		log.Errorf("Error getting Kube Rest Config: %s", err)
	}
	routeClient, err := routeclient.NewForConfig(restConfig) // kube doesn't have a routeclient
	if routeClient != nil && err == nil {                    // openshift: have a routeClient and no error
		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:           []string{"get", "update", "patch"},
			APIGroups:       []string{"security.openshift.io"},
			Resources:       []string{"securitycontextconstraints"},
			ResourceNames:   []string{},
			NonResourceURLs: []string{},
		})

		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:           []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			APIGroups:       []string{"route.openshift.io"},
			Resources:       []string{"routes"},
			ResourceNames:   []string{},
			NonResourceURLs: []string{},
		})
	} else if err != nil { // Kube or Error
		log.Warnf("Skipping Openshift Cluster Role Rules: %s", err)
	}

	synopsysOperatorClusterRole.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorClusterRole
}

// GetOperatorSecret creates a Secret Horizon component for the Synopsys-Operaotor
func (specConfig *SpecConfig) GetOperatorSecret() *horizoncomponents.Secret {
	// create a secret
	synopsysOperatorSecret := horizoncomponents.NewSecret(horizonapi.SecretConfig{
		APIVersion: "v1",
		// ClusterName : "cluster",
		Name:      "blackduck-secret",
		Namespace: specConfig.Namespace,
		Type:      specConfig.SecretType,
	})
	synopsysOperatorSecret.AddData(map[string][]byte{
		"ADMIN_PASSWORD":    []byte(specConfig.SecretAdminPassword),
		"POSTGRES_PASSWORD": []byte(specConfig.SecretPostgresPassword),
		"USER_PASSWORD":     []byte(specConfig.SecretUserPassword),
		"HUB_PASSWORD":      []byte(specConfig.SecretBlackduckPassword),
	})

	synopsysOperatorSecret.AddLabels(map[string]string{"app": "synopsys-operator"})
	return synopsysOperatorSecret
}
