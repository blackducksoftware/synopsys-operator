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
	"encoding/json"
	"github.com/blackducksoftware/synopsys-operator/utils"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	"github.com/juju/errors"
	//routev1 "github.com/openshift/api/route/v1"
	//log "github.com/sirupsen/logrus"
)

// getOperatorDeployment creates a deployment for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorDeployment() (*appv1.Deployment, error) {
	// Add the Replication Controller to the Deployer
	var synopsysOperatorReplicas int32 = 1
	var synopsysOperatorVolumeDefaultMode int32 = 420

	synopsysOperator := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &synopsysOperatorReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "synopsys-operator",
					"component": "operator",
				},
			},
			Template: v1.PodTemplateSpec{},
		},
	}

	synopsysOperator.Spec.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "synopsys-operator",
			Volumes: []v1.Volume{
				{
					Name: "synopsys-operator",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "synopsys-operator",
							},
							DefaultMode: &synopsysOperatorVolumeDefaultMode,
						},
					},
				},
				{
					Name: "synopsys-operator-tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName:  "synopsys-operator-tls",
							DefaultMode: &synopsysOperatorVolumeDefaultMode,
						},
					},
				},
				{
					Name: "tmp-logs",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium: v1.StorageMediumDefault,
						},
					},
				},
			},
		},
	}

	//synopsysOperatorContainer, err := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
	//	Name:       "synopsys-operator",
	//	Args:       []string{"/etc/synopsys-operator/config.json"},
	//	Command:    []string{"./operator"},
	//	Image:      specConfig.Image,
	//	PullPolicy: horizonapi.PullAlways,
	//})
	//if err != nil {
	//	return nil, errors.Trace(err)
	//}

	synopsysOperatorContainer := v1.Container{
		Name:       "synopsys-operator",
		Image:      specConfig.Image,
		Command:    []string{"./operator"},
		Args:       []string{"/etc/synopsys-operator/config.json"},
		WorkingDir: "",
		Ports: []v1.ContainerPort{
			{
				Name:          "8080-tcp",
				ContainerPort: 8080,
				Protocol:      v1.ProtocolTCP,
			},
		},
		EnvFrom: nil,
		Env: []v1.EnvVar{
			{
				Name: "SEAL_KEY",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "blackduck-secret",
						},
						Key:      "SEAL_KEY",
						Optional: nil,
					},
				},
			},
		},
		Resources: v1.ResourceRequirements{},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "synopsys-operator",
				MountPath: "/etc/synopsys-operator",
			},
			{
				Name:      "synopsys-operator-tls",
				MountPath: "/opt/synopsys-operator/tls",
			},
			{
				Name:      "tmp-logs",
				MountPath: "/tmp",
			},
		},
		TerminationMessagePath:   "",
		TerminationMessagePolicy: "",
		ImagePullPolicy:          v1.PullAlways,
	}

	synopsysOperatorUIContainer := v1.Container{
		Name:       "synopsys-operator-ui",
		Image:      specConfig.Image,
		Command:    []string{"./app"},
		Args:       nil,
		WorkingDir: "",
		Ports: []v1.ContainerPort{{
			Name:          "3000-TCP",
			ContainerPort: 3000,
			Protocol:      v1.ProtocolTCP,
		}},
		EnvFrom: nil,
		Env: []v1.EnvVar{
			{
				Name:  "CONFIG_FILE_PATH",
				Value: "/etc/synopsys-operator/config.json",
			},
			{
				Name:  "ADDR",
				Value: "0.0.0.0",
			},
			{
				Name:  "PORT",
				Value: "3000",
			},
			{
				Name:  "GO_ENV",
				Value: "development",
			},
		},
		Resources: v1.ResourceRequirements{},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "synopsys-operator",
				MountPath: "/etc/synopsys-operator",
			},
		},
		ImagePullPolicy: v1.PullAlways,
	}

	synopsysOperator.Spec.Template.Spec.Containers = append(synopsysOperator.Spec.Template.Spec.Containers, synopsysOperatorContainer)

	if specConfig.Expose != utils.NONE && len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), utils.BlackDuckCRDName) {
		synopsysOperator.Spec.Template.Spec.Containers = append(synopsysOperator.Spec.Template.Spec.Containers, synopsysOperatorUIContainer)
	}

	return synopsysOperator, nil
}

// getOperatorService creates a Service Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorService() []*v1.Service {

	services := []*v1.Service{}
	// Add the Service to the Deployer
	//synopsysOperatorService := horizoncomponents.NewService(horizonapi.ServiceConfig{
	//	APIVersion: "v1",
	//	Name:       "synopsys-operator",
	//	Namespace:  specConfig.Namespace,
	//})

	synopsysOperatorService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "synopsys-operator-ui",
					Protocol:   v1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
				{
					Name:       "synopsys-operator-ui-standard-port",
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(3000),
				},
				{
					Name:       "synopsys-operator",
					Protocol:   v1.ProtocolTCP,
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
				},
				{
					Name:       "synopsys-operator-tls",
					Protocol:   v1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
			},
			Selector: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
			ClusterIP:                "",
			Type:                     "",
			ExternalIPs:              nil,
			SessionAffinity:          "",
			LoadBalancerIP:           "",
			LoadBalancerSourceRanges: nil,
			ExternalName:             "",
			ExternalTrafficPolicy:    "",
			HealthCheckNodePort:      0,
			PublishNotReadyAddresses: false,
			SessionAffinityConfig:    nil,
		},
	}

	services = append(services, synopsysOperatorService)

	if strings.EqualFold(specConfig.Expose, utils.NODEPORT) || strings.EqualFold(specConfig.Expose, utils.LOADBALANCER) {

		var exposedServiceType v1.ServiceType
		if strings.EqualFold(specConfig.Expose, utils.NODEPORT) {
			exposedServiceType = v1.ServiceTypeNodePort
		} else {
			exposedServiceType = v1.ServiceTypeLoadBalancer
		}

		// Synopsys Operator UI exposed service
		//synopsysOperatorExposedService := horizoncomponents.NewService(horizonapi.ServiceConfig{
		//	APIVersion: "v1",
		//	Name:       "synopsys-operator-exposed",
		//	Namespace:  specConfig.Namespace,
		//	Type:       exposedServiceType,
		//})

		synopsysOperatorExposedService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "synopsys-operator-exposed",
				Namespace: specConfig.Namespace,
				Labels: map[string]string{
					"app":       "synopsys-operator",
					"component": "operator",
				},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"app":       "synopsys-operator",
					"component": "operator",
				},
				Type: exposedServiceType,
				Ports: []v1.ServicePort{
					{
						Name:       "synopsys-operator-ui",
						Protocol:   v1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt(3000),
					},
				},
			},
		}

		services = append(services, synopsysOperatorExposedService)
	}

	return services
}

// GetOperatorConfigMap creates a ConfigMap Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) GetOperatorConfigMap() (*v1.ConfigMap, error) {
	// Config Map
	cmData := map[string]string{}
	configData := map[string]interface{}{
		"Namespace":                     specConfig.Namespace,
		"Image":                         specConfig.Image,
		"Expose":                        specConfig.Expose,
		"ClusterType":                   specConfig.ClusterType,
		"DryRun":                        specConfig.DryRun,
		"LogLevel":                      specConfig.LogLevel,
		"Threadiness":                   specConfig.Threadiness,
		"PostgresRestartInMins":         specConfig.PostgresRestartInMins,
		"PodWaitTimeoutSeconds":         specConfig.PodWaitTimeoutSeconds,
		"ResyncIntervalInSeconds":       specConfig.ResyncIntervalInSeconds,
		"TerminationGracePeriodSeconds": specConfig.TerminationGracePeriodSeconds,
		"AdmissionWebhookListener":      specConfig.AdmissionWebhookListener,
		"CrdNames":                      strings.Join(specConfig.Crds, ","),
		"IsClusterScoped":               specConfig.IsClusterScoped,
	}
	bytes, err := json.Marshal(configData)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cmData["config.json"] = string(bytes)

	synopsysOperatorConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Data: cmData,
	}
	return synopsysOperatorConfigMap, nil
}

//getOperatorServiceAccount creates a ServiceAccount Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorServiceAccount() *v1.ServiceAccount {
	synopsysOperatorServiceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
	}
	return synopsysOperatorServiceAccount
}

//// getOperatorClusterRoleBinding creates a ClusterRoleBinding Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	synopsysOperatorClusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator-admin",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "synopsys-operator",
				Namespace: specConfig.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "",
			Kind:     "ClusterRole",
			Name:     "synopsys-operator-admin",
		},
	}
	return synopsysOperatorClusterRoleBinding
}

//
//// getOperatorRoleBinding creates a RoleBinding Horizon component for Synopsys Operator
func (specConfig *SpecConfig) getOperatorRoleBinding() *rbacv1.RoleBinding {
	// Role Binding
	//synopsysOperatorRoleBinding := horizoncomponents.NewRoleBinding(horizonapi.RoleBindingConfig{
	//	APIVersion: "rbac.authorization.k8s.io/v1beta1",
	//	Name:       "synopsys-operator-admin",
	//	Namespace:  specConfig.Namespace,
	//})

	synopsysOperatorRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator-admin",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				APIGroup:  "",
				Name:      "synopsys-operator",
				Namespace: specConfig.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "",
			Kind:     "Role",
			Name:     "synopsys-operator-admin",
		},
	}
	return synopsysOperatorRoleBinding
}

//
//// getOperatorClusterRole creates a ClusterRole Horizon component for the Synopsys Operator
func (specConfig *SpecConfig) getOperatorClusterRole() *rbacv1.ClusterRole {
	synopsysOperatorClusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator-admin",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"clusterrolebindings", "clusterroles"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"rolebindings", "roles"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"batch", "extensions"},
				Resources: []string{"jobs", "cronjobs", "ingresses", "endpoints", "namespaces"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"autoscaling"},
				Resources: []string{"horizontalpodautoscalers"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"extensions", "apps"},
				Resources: []string{"deployments", "deployments/scale", "deployments/rollback", "statefulsets", "statefulsets/scale", "replicasets", "replicasets/scale", "daemonsets"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{""},
				Resources: []string{"namespaces", "configmaps", "persistentvolumeclaims", "services", "secrets", "replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "update"},
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log", "endpoints"},
			},
			{
				Verbs:     []string{"create"},
				APIGroups: []string{""},
				Resources: []string{"pods/exec"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"synopsys.com"},
				Resources: []string{"*"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses", "volumeattachments"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"policy"},
				Resources: []string{"poddisruptionbudgets"},
			},
			{
				Verbs:     []string{"create", "delete", "patch"},
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
			},
		},
		AggregationRule: nil,
	}

	// Add Openshift rules
	if specConfig.ClusterType == OpenshiftClusterType {
		synopsysOperatorClusterRole.Rules = append(synopsysOperatorClusterRole.Rules, rbacv1.PolicyRule{
			Verbs:     []string{"get", "update", "patch"},
			APIGroups: []string{"security.openshift.io"},
			Resources: []string{"securitycontextconstraints"},
		})

		synopsysOperatorClusterRole.Rules = append(synopsysOperatorClusterRole.Rules, rbacv1.PolicyRule{
			Verbs:     []string{"get", "list", "create", "delete", "deletecollection"},
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes"},
		})

		synopsysOperatorClusterRole.Rules = append(synopsysOperatorClusterRole.Rules, rbacv1.PolicyRule{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"image.openshift.io"},
			Resources: []string{"images"},
		})

	}

	return synopsysOperatorClusterRole
}

// getOperatorRole creates a Role Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorRole() *rbacv1.Role {
	synopsysOperatorRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator-admin",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"rolebindings", "roles"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"batch", "extensions"},
				Resources: []string{"jobs", "cronjobs", "ingresses", "endpoints", "namespaces"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"autoscaling"},
				Resources: []string{"horizontalpodautoscalers"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"extensions", "apps"},
				Resources: []string{"deployments", "deployments/scale", "deployments/rollback", "statefulsets", "statefulsets/scale", "replicasets", "replicasets/scale", "daemonsets"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{""},
				Resources: []string{"namespaces", "configmaps", "persistentvolumeclaims", "services", "secrets", "replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "update"},
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log", "endpoints"},
			},
			{
				Verbs:     []string{"create"},
				APIGroups: []string{""},
				Resources: []string{"pods/exec"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"synopsys.com"},
				Resources: []string{"*"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses", "volumeattachments"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
				APIGroups: []string{"policy"},
				Resources: []string{"poddisruptionbudgets"},
			},
			{
				Verbs:     []string{"create", "delete", "patch"},
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
			},
		},
	}

	// Add Openshift rules
	if specConfig.ClusterType == OpenshiftClusterType {

		synopsysOperatorRole.Rules = append(synopsysOperatorRole.Rules, rbacv1.PolicyRule{
			Verbs:     []string{"get", "list", "create", "delete", "deletecollection"},
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes"},
		})

	}
	return synopsysOperatorRole
}

// getTLSCertificateSecret creates a TLS certificate in horizon format
func (specConfig *SpecConfig) getTLSCertificateSecret() *v1.Secret {
	tlsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "synopsys-operator-tls",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Data: map[string][]byte{
			"cert.crt": []byte(specConfig.Certificate),
			"cert.key": []byte(specConfig.CertificateKey),
		},
		Type: v1.SecretTypeOpaque,
	}
	return tlsSecret
}

// getOperatorSecret creates a Secret Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorSecret() *v1.Secret {
	//// create a secret
	synopsysOperatorSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "blackduck-secret",
			Namespace: specConfig.Namespace,
			Labels: map[string]string{
				"app":       "synopsys-operator",
				"component": "operator",
			},
		},
		Data: map[string][]byte{
			"SEAL_KEY": []byte(specConfig.SealKey),
		},
		Type: v1.SecretTypeOpaque,
	}

	return synopsysOperatorSecret
}

//// getOpenShiftRoute creates the OpenShift route component for Synopsys Operator
//func (specConfig *SpecConfig) getOpenShiftRoute() *api.Route {
//	if strings.ToUpper(specConfig.Expose) == util.OPENSHIFT {
//		return &api.Route{
//			Name:               "synopsys-operator-ui",
//			Namespace:          specConfig.Namespace,
//			Kind:               "Service",
//			ServiceName:        "synopsys-operator",
//			PortName:           "synopsys-operator-ui",
//			Labels:             map[string]string{"app": "synopsys-operator", "component": "operator"},
//			TLSTerminationType: routev1.TLSTerminationEdge,
//		}
//	}
//	return nil
//}
