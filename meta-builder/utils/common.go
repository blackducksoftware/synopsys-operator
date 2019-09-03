/*
Copyright (C) 2018 Synopsys, Inc.

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

package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"strings"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/api/storage/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CreateSecretFromFile will create the secret from file
func CreateSecretFromFile(clientset *kubernetes.Clientset, jsonFile string, namespace string, name string, dataKey string) (*corev1.Secret, error) {
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Panicf("Unable to read the secret file %s due to error: %v\n", jsonFile, err)
	}

	return clientset.CoreV1().Secrets(namespace).Create(&corev1.Secret{
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{dataKey: string(file)},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// CreateSecret will create the secret
func CreateSecret(clientset *kubernetes.Clientset, namespace string, name string, stringData map[string]string) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Create(&corev1.Secret{
		Type:       corev1.SecretTypeOpaque,
		StringData: stringData,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// GetSecret will create the secret
func GetSecret(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

// ListSecrets will list the secret
func ListSecrets(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.SecretList, error) {
	return clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateSecret updates a secret
func UpdateSecret(clientset *kubernetes.Clientset, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Update(secret)
}

// DeleteSecret will delete the secret
func DeleteSecret(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{})
}

// ReadFromFile will read the file
func ReadFromFile(filePath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filePath)
	return file, err
}

// GetConfigMap will get the config map
func GetConfigMap(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.ConfigMap, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

// ListConfigMaps will list the config map
func ListConfigMaps(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.ConfigMapList, error) {
	return clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateConfigMap updates a config map
func UpdateConfigMap(clientset *kubernetes.Clientset, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Update(configMap)
}

// DeleteConfigMap will delete the config map
func DeleteConfigMap(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

// // CreateSecret will create the secret
// func CreateNewSecret(secretConfig *horizonapi.SecretConfig) *components.Secret {
//
// 	secret := components.NewSecret(horizonapi.SecretConfig{Namespace: secretConfig.Namespace, Name: secretConfig.Name, Type: secretConfig.Type})
//
// 	secret.AddData(secretConfig.Data)
// 	secret.AddStringData(secretConfig.StringData)
// 	secret.AddLabels(secretConfig.Labels)
// 	secret.AddAnnotations(secretConfig.Annotations)
//
// 	return secret
// }
//
// // CreateConfigMap will create the configMap
// func CreateConfigMap(configMapConfig *horizonapi.ConfigMapConfig) *components.ConfigMap {
//
// 	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: configMapConfig.Namespace, Name: configMapConfig.Name})
//
// 	configMap.AddData(configMapConfig.Data)
// 	configMap.AddLabels(configMapConfig.Labels)
// 	configMap.AddAnnotations(configMapConfig.Annotations)
//
// 	return configMap
// }

// CreateNamespace will create the namespace
func CreateNamespace(clientset *kubernetes.Clientset, namespace string) (*corev1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      namespace,
		},
	})
}

// GetNamespace will get the namespace
func GetNamespace(clientset *kubernetes.Clientset, namespace string) (*corev1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
}

// ListNamespaces will list the namespace
func ListNamespaces(clientset *kubernetes.Clientset, labelSelector string) (*corev1.NamespaceList, error) {
	return clientset.CoreV1().Namespaces().List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateNamespace updates a namespace
func UpdateNamespace(clientset *kubernetes.Clientset, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Update(namespace)
}

// DeleteNamespace will delete the namespace
func DeleteNamespace(clientset *kubernetes.Clientset, namespace string) error {
	return clientset.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
}

// GetPod will get the input pods corresponding to a namespace
func GetPod(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.Pod, error) {
	return clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

// ListPods will get all the pods corresponding to a namespace
func ListPods(clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
	return clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

// ListPodsWithLabels will get all the pods corresponding to a namespace and labels
func ListPodsWithLabels(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.PodList, error) {
	return clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// DeletePod will delete the input pods corresponding to a namespace
func DeletePod(clientset *kubernetes.Clientset, namespace string, name string) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return clientset.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

// GetReplicationController will get the replication controller corresponding to a namespace and name
func GetReplicationController(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.ReplicationController, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).Get(name, metav1.GetOptions{})
}

// ListReplicationControllers will get the replication controllers corresponding to a namespace
func ListReplicationControllers(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.ReplicationControllerList, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateReplicationController updates the replication controller
func UpdateReplicationController(clientset *kubernetes.Clientset, namespace string, rc *corev1.ReplicationController) (*corev1.ReplicationController, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).Update(rc)
}

// DeleteReplicationController will delete the replication controller corresponding to a namespace and name
func DeleteReplicationController(clientset *kubernetes.Clientset, namespace string, name string) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return clientset.CoreV1().ReplicationControllers(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

// GetDeployment will get the deployment corresponding to a namespace and name
func GetDeployment(clientset *kubernetes.Clientset, namespace string, name string) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

// ListDeployments will get all the deployments corresponding to a namespace
func ListDeployments(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*appsv1.DeploymentList, error) {
	return clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateDeployment updates the deployment
func UpdateDeployment(clientset *kubernetes.Clientset, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(namespace).Update(deployment)
}

// DeleteDeployment will delete the deployment corresponding to a namespace and name
func DeleteDeployment(clientset *kubernetes.Clientset, namespace string, name string) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return clientset.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

// CreatePersistentVolume will create the persistent volume
func CreatePersistentVolume(clientset *kubernetes.Clientset, name string, storageClass string, claimSize string, nfsPath string, nfsServer string) (*corev1.PersistentVolume, error) {
	pvQuantity, _ := resource.ParseQuantity(claimSize)
	return clientset.CoreV1().PersistentVolumes().Create(&corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name,
			Name:      name,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity:         map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: pvQuantity},
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: storageClass,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Path:   nfsPath,
					Server: nfsServer,
				},
			},
		},
	})
}

// DeletePersistentVolume will delete the persistent volume
func DeletePersistentVolume(clientset *kubernetes.Clientset, name string) error {
	return clientset.CoreV1().PersistentVolumes().Delete(name, &metav1.DeleteOptions{})
}

// ValidateServiceEndpoint will validate whether the service endpoint is ready to serve
func ValidateServiceEndpoint(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.Endpoints, error) {
	var endpoint *corev1.Endpoints
	var err error
	for i := 0; i < 20; i++ {
		endpoint, err = GetServiceEndPoint(clientset, namespace, name)
		if err != nil {
			log.Infof("waiting for %s endpoint in %s", name, namespace)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	return endpoint, err
}

// WaitForServiceEndpointReady will wait for the service endpoint to start the service
func WaitForServiceEndpointReady(clientset *kubernetes.Clientset, namespace string, name string) error {
	endpoint, err := ValidateServiceEndpoint(clientset, namespace, name)
	if err != nil {
		return fmt.Errorf("unable to get service endpoint %s in %s because %+v", name, namespace, err)
	}
	for _, subset := range endpoint.Subsets {
		if len(subset.NotReadyAddresses) > 0 {
			for {
				log.Infof("waiting for %s in %s to be cloned/backed up", name, namespace)
				svc, err := GetServiceEndPoint(clientset, namespace, name)
				if err != nil {
					return fmt.Errorf("unable to get service endpoint %s in %s because %+v", name, namespace, err)
				}

				for _, subset := range svc.Subsets {
					if len(subset.Addresses) > 0 {
						return nil
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}
	return nil
}

// FilterPodByNamePrefixInNamespace will filter the pod based on pod name prefix from a list a pods in a given namespace
func FilterPodByNamePrefixInNamespace(clientset *kubernetes.Clientset, namespace string, prefix string) (*corev1.Pod, error) {
	pods, err := ListPods(clientset, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to list the pods in namespace %s due to %+v", namespace, err)
	}

	pod := FilterPodByNamePrefix(pods, prefix)
	if pod != nil {
		return pod, nil
	}
	return nil, fmt.Errorf("unable to find the pod with prefix %s", prefix)
}

// FilterPodByNamePrefix will filter the pod based on pod name prefix from a list a pods
func FilterPodByNamePrefix(pods *corev1.PodList, prefix string) *corev1.Pod {
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			return &pod
		}
	}
	return nil
}

// NewStringReader will convert string array to string reader object
func NewStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}

// GetService will get the service information for the input service name inside the input namespace
func GetService(clientset *kubernetes.Clientset, namespace string, serviceName string) (*corev1.Service, error) {
	return clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
}

// ListServices will list the service information for the input service name inside the input namespace
func ListServices(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.ServiceList, error) {
	return clientset.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateService will update the service information for the input service name inside the input namespace
func UpdateService(clientset *kubernetes.Clientset, namespace string, service *corev1.Service) (*corev1.Service, error) {
	return clientset.CoreV1().Services(namespace).Update(service)
}

// DeleteService will delete the service information for the input service name inside the input namespace
func DeleteService(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().Services(namespace).Delete(name, &metav1.DeleteOptions{})
}

// GetServiceEndPoint will get the service endpoint information for the input service name inside the input namespace
func GetServiceEndPoint(clientset *kubernetes.Clientset, namespace string, serviceName string) (*corev1.Endpoints, error) {
	return clientset.CoreV1().Endpoints(namespace).Get(serviceName, metav1.GetOptions{})
}

// ListStorageClasses will list all the storageClass in the cluster
func ListStorageClasses(clientset *kubernetes.Clientset) (*v1beta1.StorageClassList, error) {
	return clientset.StorageV1beta1().StorageClasses().List(metav1.ListOptions{})
}

// GetPVC will get the PVC for the given name
func GetPVC(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

// ListPVCs will list the PVC for the given label selector
func ListPVCs(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.PersistentVolumeClaimList, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdatePVC will update the pvc information for the input pvc name inside the input namespace
func UpdatePVC(clientset *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Update(pvc)
}

// DeletePVC will delete the PVC information for the input pvc name inside the input namespace
func DeletePVC(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}

// IntToPtr will convert int to pointer
func IntToPtr(i int) *int {
	return &i
}

// BoolToPtr will convert bool to pointer
func BoolToPtr(b bool) *bool {
	return &b
}

// Int32ToInt will convert from int32 to int
func Int32ToInt(i *int32) int {
	return int(*i)
}

// IntToInt32 will convert from int to int32
func IntToInt32(i int) *int32 {
	j := int32(i)
	return &j
}

// IntToInt64 will convert from int to int64
func IntToInt64(i int) *int64 {
	j := int64(i)
	return &j
}

// IntToUInt32 will convert from int to uint32
func IntToUInt32(i int) uint32 {
	return uint32(i)
}

func getBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Base64Encode will return an encoded string using a URL-compatible base64 format
func Base64Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64Decode will return a decoded string using a URL-compatible base64 format;
// decoding may return an error, which you can check if you don’t already know the input to be well-formed.
func Base64Decode(data string) (string, error) {
	uDec, err := base64.URLEncoding.DecodeString(data)
	return string(uDec), err
}

// RandomString will generate the random string
func RandomString(n int) (string, error) {
	b, err := getBytes(n)
	return Base64Encode(b), err
}

// GetServiceAccount get a service account
func GetServiceAccount(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.ServiceAccount, error) {
	return clientset.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
}

// ListServiceAccounts list a service account
func ListServiceAccounts(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.ServiceAccountList, error) {
	return clientset.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateServiceAccount updates a service account
func UpdateServiceAccount(clientset *kubernetes.Clientset, namespace string, serviceAccount *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	return clientset.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
}

// DeleteServiceAccount delete a service account
func DeleteServiceAccount(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.CoreV1().ServiceAccounts(namespace).Delete(name, &metav1.DeleteOptions{})
}

// GetClusterRoleBinding get a cluster role
func GetClusterRoleBinding(clientset *kubernetes.Clientset, name string) (*rbacv1.ClusterRoleBinding, error) {
	return clientset.RbacV1().ClusterRoleBindings().Get(name, metav1.GetOptions{})
}

// ListClusterRoleBindings list a cluster role binding
func ListClusterRoleBindings(clientset *kubernetes.Clientset, labelSelector string) (*rbacv1.ClusterRoleBindingList, error) {
	return clientset.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateClusterRoleBinding updates the cluster role binding
func UpdateClusterRoleBinding(clientset *kubernetes.Clientset, clusterRoleBinding *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	return clientset.RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
}

// DeleteClusterRoleBinding delete a cluster role binding
func DeleteClusterRoleBinding(clientset *kubernetes.Clientset, name string) error {
	return clientset.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
}

// IsClusterRoleBindingSubjectNamespaceExist checks whether the namespace is already exist in the subject of cluster role binding
func IsClusterRoleBindingSubjectNamespaceExist(subjects []rbacv1.Subject, namespace string) bool {
	for _, subject := range subjects {
		if strings.EqualFold(subject.Namespace, namespace) {
			return true
		}
	}
	return false
}

// IsClusterRoleRefExistForOtherNamespace checks whether the cluster role exist for any cluster role bindings present in other namespace
func IsClusterRoleRefExistForOtherNamespace(roleRef rbacv1.RoleRef, roleName string, namespace string, subjects []rbacv1.Subject) bool {
	for _, subject := range subjects {
		if "clusterrole" == strings.ToLower(roleRef.Kind) && strings.EqualFold(roleRef.Name, roleName) && !strings.EqualFold(namespace, subject.Namespace) {
			return true
		}
	}
	return false
}

// IsSubjectExistForOtherNamespace checks whether anyother namespace is exist in the subject of cluster role binding
func IsSubjectExistForOtherNamespace(subject rbacv1.Subject, namespace string) bool {
	if !strings.EqualFold(subject.Namespace, namespace) {
		return true
	}
	return false
}

// IsSubjectExist checks whether the namespace is already exist in the subject of cluster role binding
func IsSubjectExist(subjects []rbacv1.Subject, namespace string, name string) bool {
	for _, subject := range subjects {
		if strings.EqualFold(subject.Namespace, namespace) && strings.EqualFold(subject.Name, name) {
			return true
		}
	}
	return false
}

// GetClusterRole get a cluster role
func GetClusterRole(clientset *kubernetes.Clientset, name string) (*rbacv1.ClusterRole, error) {
	return clientset.RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
}

// ListClusterRoles list a cluster role
func ListClusterRoles(clientset *kubernetes.Clientset, labelSelector string) (*rbacv1.ClusterRoleList, error) {
	return clientset.RbacV1().ClusterRoles().List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateClusterRole updates the cluster role
func UpdateClusterRole(clientset *kubernetes.Clientset, clusterRole *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	return clientset.RbacV1().ClusterRoles().Update(clusterRole)
}

// DeleteClusterRole delete a cluster role
func DeleteClusterRole(clientset *kubernetes.Clientset, name string) error {
	return clientset.RbacV1().ClusterRoles().Delete(name, &metav1.DeleteOptions{})
}

// IsClusterRoleRuleExist checks whether the namespace is already exist in the rule of cluster role
func IsClusterRoleRuleExist(oldRules []rbacv1.PolicyRule, newRule rbacv1.PolicyRule) bool {
	for _, oldRule := range oldRules {
		if reflect.DeepEqual(oldRule, newRule) {
			return true
		}
	}
	return false
}

// GetRole get a role
func GetRole(clientset *kubernetes.Clientset, namespace string, name string) (*rbacv1.Role, error) {
	return clientset.RbacV1().Roles(namespace).Get(name, metav1.GetOptions{})
}

// ListRoles list a role
func ListRoles(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*rbacv1.RoleList, error) {
	return clientset.RbacV1().Roles(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateRole updates the role
func UpdateRole(clientset *kubernetes.Clientset, namespace string, role *rbacv1.Role) (*rbacv1.Role, error) {
	return clientset.RbacV1().Roles(namespace).Update(role)
}

// DeleteRole delete a role
func DeleteRole(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.RbacV1().Roles(namespace).Delete(name, &metav1.DeleteOptions{})
}

// GetRoleBinding get a role binding
func GetRoleBinding(clientset *kubernetes.Clientset, namespace string, name string) (*rbacv1.RoleBinding, error) {
	return clientset.RbacV1().RoleBindings(namespace).Get(name, metav1.GetOptions{})
}

// ListRoleBindings list a role binding
func ListRoleBindings(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*rbacv1.RoleBindingList, error) {
	return clientset.RbacV1().RoleBindings(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateRoleBinding updates the role binding
func UpdateRoleBinding(clientset *kubernetes.Clientset, namespace string, role *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	return clientset.RbacV1().RoleBindings(namespace).Update(role)
}

// DeleteRoleBinding delete a role binding
func DeleteRoleBinding(clientset *kubernetes.Clientset, namespace string, name string) error {
	return clientset.RbacV1().RoleBindings(namespace).Delete(name, &metav1.DeleteOptions{})
}

// GetRouteClient attempts to get a Route Client. It returns nil if it
// fails due to an error or due to being on kubernetes (doesn't support routes)
func GetRouteClient(restConfig *rest.Config) *routeclient.RouteV1Client {
	routeClient, err := routeclient.NewForConfig(restConfig)
	if err != nil {
		return nil
	}
	return routeClient
}

// GetRoute gets an OpenShift routes
func GetRoute(routeClient *routeclient.RouteV1Client, namespace string, name string) (*routev1.Route, error) {
	return routeClient.Routes(namespace).Get(name, metav1.GetOptions{})
}

// ListRoutes list an OpenShift routes
func ListRoutes(routeClient *routeclient.RouteV1Client, namespace string, labelSelector string) (*routev1.RouteList, error) {
	return routeClient.Routes(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateRoute updates an OpenShift routes
func UpdateRoute(routeClient *routeclient.RouteV1Client, namespace string, route *routev1.Route) (*routev1.Route, error) {
	return routeClient.Routes(namespace).Update(route)
}

// CreateRoute creates an OpenShift routes
func CreateRoute(routeClient *routeclient.RouteV1Client, namespace string, route *routev1.Route) (*routev1.Route, error) {
	return routeClient.Routes(namespace).Create(route)
}

// DeleteRoute deletes an OpenShift routes
func DeleteRoute(routeClient *routeclient.RouteV1Client, namespace string, name string) error {
	return routeClient.Routes(namespace).Delete(name, &metav1.DeleteOptions{})
}

// GetOpenShiftSecurityConstraint gets an OpenShift security constraints
func GetOpenShiftSecurityConstraint(osSecurityClient *securityclient.SecurityV1Client, name string) (*securityv1.SecurityContextConstraints, error) {
	return osSecurityClient.SecurityContextConstraints().Get(name, metav1.GetOptions{})
}

// UpdateOpenShiftSecurityConstraint updates an OpenShift security constraints
func UpdateOpenShiftSecurityConstraint(osSecurityClient *securityclient.SecurityV1Client, serviceAccounts []string, name string) error {
	scc, err := GetOpenShiftSecurityConstraint(osSecurityClient, name)
	if err != nil {
		return fmt.Errorf("failed to get scc %s: %v", name, err)
	}

	newUsers := []string{}
	// Only add the service account if it isn't already in the list of users for the privileged scc
	for _, sa := range serviceAccounts {
		exist := false
		for _, user := range scc.Users {
			if strings.Compare(user, sa) == 0 {
				exist = true
				break
			}
		}

		if !exist {
			newUsers = append(newUsers, sa)
		}
	}

	if len(newUsers) > 0 {
		scc.Users = append(scc.Users, newUsers...)

		_, err = osSecurityClient.SecurityContextConstraints().Update(scc)
		if err != nil {
			return fmt.Errorf("failed to update scc %s: %v", name, err)
		}
	}
	return err
}

// PatchReplicationControllerForReplicas patch a replication controller for replica update
func PatchReplicationControllerForReplicas(clientset *kubernetes.Clientset, old *corev1.ReplicationController, replicas *int32) (*corev1.ReplicationController, error) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	new := old.DeepCopy()
	new.Spec.Replicas = replicas
	newData, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.ReplicationController{})
	if err != nil {
		return nil, err
	}
	rc, err := clientset.CoreV1().ReplicationControllers(new.Namespace).Patch(new.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

// PatchReplicationController patch a replication controller
func PatchReplicationController(clientset *kubernetes.Clientset, old corev1.ReplicationController, new corev1.ReplicationController) error {
	oldData, err := json.Marshal(old)
	if err != nil {
		return err
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.ReplicationController{})
	if err != nil {
		return err
	}
	newRc, err := clientset.CoreV1().ReplicationControllers(new.Namespace).Patch(new.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	newRc, err = PatchReplicationControllerForReplicas(clientset, newRc, IntToInt32(0))
	if err != nil {
		return err
	}

	newRc, err = PatchReplicationControllerForReplicas(clientset, newRc, new.Spec.Replicas)
	if err != nil {
		return err
	}
	return nil
}

// PatchDeploymentForReplicas patch a deployment for replica update
func PatchDeploymentForReplicas(clientset *kubernetes.Clientset, old *appsv1.Deployment, replicas *int32) (*appsv1.Deployment, error) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	new := old.DeepCopy()
	new.Spec.Replicas = replicas
	newData, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, appsv1.Deployment{})
	if err != nil {
		return nil, err
	}
	newDeployment, err := clientset.AppsV1().Deployments(new.Namespace).Patch(new.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return nil, err
	}
	return newDeployment, nil
}

// PatchDeployment patch a deployment
func PatchDeployment(clientset *kubernetes.Clientset, old appsv1.Deployment, new appsv1.Deployment) error {
	oldData, err := json.Marshal(old)
	if err != nil {
		return err
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, appsv1.Deployment{})
	if err != nil {
		return err
	}
	newDeployment, err := clientset.AppsV1().Deployments(new.Namespace).Patch(new.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	newDeployment, err = PatchDeploymentForReplicas(clientset, newDeployment, IntToInt32(0))
	if err != nil {
		return err
	}

	newDeployment, err = PatchDeploymentForReplicas(clientset, newDeployment, new.Spec.Replicas)
	if err != nil {
		return err
	}
	return nil
}

// UniqueValues returns a unique subset of the string slice provided.
func UniqueValues(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

// GetCustomResourceDefinition get the custom resource defintion
func GetCustomResourceDefinition(apiExtensionClient *apiextensionsclient.Clientset, name string) (*apiextensions.CustomResourceDefinition, error) {
	return apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
}

// ListCustomResourceDefinitions list the custom resource defintions
func ListCustomResourceDefinitions(apiExtensionClient *apiextensionsclient.Clientset, labelSelector string) (*apiextensions.CustomResourceDefinitionList, error) {
	return apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{LabelSelector: labelSelector})
}

// UpdateCustomResourceDefinition updates the custom resource defintion
func UpdateCustomResourceDefinition(apiExtensionClient *apiextensionsclient.Clientset, crd *apiextensions.CustomResourceDefinition) (*apiextensions.CustomResourceDefinition, error) {
	return apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(crd)
}

// DeleteCustomResourceDefinition deletes the custom resource defintion
func DeleteCustomResourceDefinition(apiExtensionClient *apiextensionsclient.Clientset, name string) error {
	return apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, &metav1.DeleteOptions{})
}

// WaitUntilPodsAreReady will wait for the pods to be ready
func WaitUntilPodsAreReady(clientset *kubernetes.Clientset, namespace string, labelSelector string, timeoutInSeconds int64) error {
	// timer starts the timer for timeoutInSeconds. If the task doesn't completed, return error
	timeout := time.NewTimer(time.Duration(timeoutInSeconds) * time.Second)
	// ticker starts and execute the task for every n intervals
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			// check right before the timeout; this will handle both when timeout is less than ticker; and also when timeout is not a multiple of 10 seconds
			podsAreReady, err := ArePodsReady(clientset, namespace, labelSelector)
			if err != nil {
				return err
			}
			if podsAreReady == false {
				return nil
			}
			return fmt.Errorf("[NS: %s | Label: %s] the pods weren't ready - timing out after %d seconds", namespace, labelSelector, timeoutInSeconds)
		case <-ticker.C:
			// log.Debugf("Ticker ticked at: %v", time.Now())
			podsAreReady, err := ArePodsReady(clientset, namespace, labelSelector)
			if err != nil {
				return err
			}
			if podsAreReady == true {
				return nil
			}
		}
	}
}

// ArePodsReady returns whether the pods are ready or not. Returns an error if pods will never become ready.
func ArePodsReady(clientset *kubernetes.Clientset, namespace string, labelSelector string) (bool, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return false, err
	}

	// Check if all the pods are ready
	arePodsReady := true
	for _, p := range pods.Items {
		// Skip if a pod is in a failed or unknown state
		if p.Status.Phase == corev1.PodFailed || p.Status.Phase == corev1.PodUnknown {
			continue
		}
		// verify the pod is ready, otherwise set arePodsReady to false
		for _, condition := range p.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				arePodsReady = false
			}
		}
	}
	return arePodsReady, nil
}

// GetClusterScopeByName returns whether the CRD is cluster scope
func GetClusterScopeByName(apiExtensionClient *apiextensionsclient.Clientset, name string) bool {
	cr, err := GetCustomResourceDefinition(apiExtensionClient, name)
	if err == nil && strings.EqualFold("CLUSTER", string(cr.Spec.Scope)) {
		return true
	}
	return false
}

// GetClusterScope returns whether any of the CRD is cluster scope
func GetClusterScope(apiExtensionClient *apiextensionsclient.Clientset) bool {
	crds := []string{AlertCRDName, BlackDuckCRDName, OpsSightCRDName}
	for _, crd := range crds {
		cr, err := GetCustomResourceDefinition(apiExtensionClient, crd)
		if err == nil && strings.EqualFold("CLUSTER", string(cr.Spec.Scope)) {
			return true
		}
	}
	return false
}

// GetOperatorNamespace uses labels to return the namespace of the synopsys operator based on
// the provided namespace. In cluster scoped mode it will return the namespaces of all operators if provided
// namespace=NamespaceAll. In namespace scoped mode it will return nil if there is no operator in
// the namespace
func GetOperatorNamespace(clientset *kubernetes.Clientset, namespace string) ([]string, error) {
	namespaces := make(map[string]string, 0)
	// check if operator is already installed
	rcs, err := ListReplicationControllers(clientset, namespace, "app=synopsys-operator,component=operator")
	if err == nil && len(rcs.Items) > 0 {
		for _, rc := range rcs.Items {
			if metav1.NamespaceAll != namespace {
				return []string{rc.Namespace}, nil
			}
			if _, ok := namespaces[rc.Namespace]; !ok {
				namespaces[rc.Namespace] = rc.Namespace
			}
		}
	}
	deployments, err := ListDeployments(clientset, namespace, "app=synopsys-operator,component=operator")
	if err == nil && len(deployments.Items) > 0 {
		for _, deployment := range deployments.Items {
			if metav1.NamespaceAll != namespace {
				return []string{deployment.Namespace}, nil
			}
			if _, ok := namespaces[deployment.Namespace]; !ok {
				namespaces[deployment.Namespace] = deployment.Namespace
			}
		}
	}
	pods, err := ListPodsWithLabels(clientset, namespace, "app=synopsys-operator,component=operator")
	if err == nil && len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if metav1.NamespaceAll != namespace {
				return []string{pod.Namespace}, nil
			}
			if _, ok := namespaces[pod.Namespace]; !ok {
				namespaces[pod.Namespace] = pod.Namespace
			}
		}
	}
	if len(namespaces) > 0 {
		return MapKeyToStringArray(namespaces), nil
	}
	return nil, fmt.Errorf("unable to find the synopsys operator namespace")
}

// MapKeyToStringArray will return map keys
func MapKeyToStringArray(maps map[string]string) []string {
	keys := make([]string, 0)
	for key := range maps {
		keys = append(keys, key)
	}
	return keys
}

// GetOperatorRoles returns the roles or the cluster role of the synopsys operator based on the labels
func GetOperatorRoles(clientset *kubernetes.Clientset, namespace string) ([]string, []string, error) {
	clusterRoles := []string{}
	roles := []string{}

	// synopsysctl case
	// list cluster roles with app=synopsys-operator
	crs, err := ListClusterRoles(clientset, "app=synopsys-operator")
	if err != nil {
		return clusterRoles, roles, fmt.Errorf("unable to list the cluster roles due to %+v", err)
	}
	for _, cr := range crs.Items {
		clusterRoles = append(clusterRoles, cr.Name)
	}

	// list roles with app=synopsys-operator
	rs, err := ListRoles(clientset, namespace, "app=synopsys-operator")
	if err != nil {
		return clusterRoles, roles, fmt.Errorf("unable to list the roles due to %+v", err)
	}
	for _, r := range rs.Items {
		roles = append(roles, r.Name)
	}

	// OLM case
	if len(roles) == 0 {
		// list cluster roles with app=synopsys-operator
		crs, err := ListClusterRoles(clientset, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil {
			return clusterRoles, roles, fmt.Errorf("unable to list the cluster roles due to %+v", err)
		}
		for _, cr := range crs.Items {
			clusterRoles = append(clusterRoles, cr.Name)
		}

		// list roles with app=synopsys-operator
		rs, err := ListRoles(clientset, namespace, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil {
			return clusterRoles, roles, fmt.Errorf("unable to list the roles due to %+v", err)
		}
		for _, r := range rs.Items {
			roles = append(roles, r.Name)
		}
	}
	return UniqueStringSlice(clusterRoles), UniqueStringSlice(roles), nil
}

// GetOperatorRoleBindings returns the cluster role bindings of the synopsys operator based on the labels
func GetOperatorRoleBindings(clientset *kubernetes.Clientset, namespace string) ([]string, []string, error) {
	clusterRolebindings := []string{}
	rolebindings := []string{}

	// synopsysctl case
	// list cluster role binding
	crbs, err := ListClusterRoleBindings(clientset, "app=synopsys-operator,component=operator")
	if err != nil {
		return clusterRolebindings, rolebindings, fmt.Errorf("unable to list the cluster role bindings due to %+v", err)
	}
	for _, crb := range crbs.Items {
		clusterRolebindings = append(clusterRolebindings, crb.Name)
	}

	// list role binding
	rbs, err := ListRoleBindings(clientset, namespace, "app=synopsys-operator,component=operator")
	if err != nil {
		return clusterRolebindings, rolebindings, fmt.Errorf("unable to list the role bindings due to %+v", err)
	}
	for _, rb := range rbs.Items {
		rolebindings = append(rolebindings, rb.Name)
	}

	// OLM case
	if len(rolebindings) == 0 {
		// list cluster role binding
		crbs, err := ListClusterRoleBindings(clientset, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil {
			return clusterRolebindings, rolebindings, fmt.Errorf("unable to list the cluster role bindings due to %+v", err)
		}
		for _, crb := range crbs.Items {
			clusterRolebindings = append(clusterRolebindings, crb.Name)
		}

		// list role binding
		rbs, err := ListRoleBindings(clientset, namespace, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil {
			return clusterRolebindings, rolebindings, fmt.Errorf("unable to list the role bindings due to %+v", err)
		}
		for _, rb := range rbs.Items {
			rolebindings = append(rolebindings, rb.Name)
		}
	}
	return UniqueStringSlice(clusterRolebindings), UniqueStringSlice(rolebindings), nil
}

// GetKubernetesVersion will return the kubernetes version
func GetKubernetesVersion(clientset *kubernetes.Clientset) (string, error) {
	k, err := clientset.Discovery().ServerVersion()
	if k != nil {
		return k.GitVersion, nil
	}
	return "", err
}

// IsOpenshift will whether it is an openshift cluster
func IsOpenshift(clientset *kubernetes.Clientset) bool {
	body, err := clientset.Discovery().RESTClient().Get().AbsPath("/").Do().Raw()
	if err != nil {
		return false
	}

	return strings.Contains(string(body), "openshift")
}

// IsOperatorExist returns whether the operator exist or not
func IsOperatorExist(clientset *kubernetes.Clientset, namespace string) bool {
	rcs, err := ListReplicationControllers(clientset, namespace, "app=synopsys-operator")
	if err == nil && len(rcs.Items) > 0 {
		return true
	}
	deployments, err := ListDeployments(clientset, namespace, "app=synopsys-operator")
	if err == nil && len(deployments.Items) > 0 {
		return true
	}
	pods, err := ListPodsWithLabels(clientset, namespace, "app=synopsys-operator")
	if err == nil && len(pods.Items) > 0 {
		return true
	}
	return false
}

// isSynopsysResourceExist returns whether any Synopsys resources like Alert, Black Duck or OpsSight etc. exist in the namespace
func isSynopsysResourceExist(namespace, label string) error {
	crdNames := []string{AlertName, BlackDuckName, OpsSightName}
	for _, crdName := range crdNames {
		if strings.HasPrefix(label, fmt.Sprintf("synopsys.com/%s", crdName)) {
			name := label[len(fmt.Sprintf("synopsys.com/%s", crdName))+1:]
			return fmt.Errorf("%s %s instance is already running in %s namespace", name, crdName, namespace)
		}
	}
	return nil
}

// IsOwnerLabelExistInNamespace check for owner label exist in the namespace
func IsOwnerLabelExistInNamespace(kubeClient *kubernetes.Clientset, namespace string) bool {
	// verify whether the namespace exist
	ns, err := GetNamespace(kubeClient, namespace)
	if err != nil {
		return false
	}

	// check for owner label
	if owner, ok := ns.Labels["owner"]; !ok || (ok && owner != OperatorName) {
		return false
	}
	return true
}

// CheckResourceNamespace checks whether namespace is having any resource types
func CheckResourceNamespace(kubeClient *kubernetes.Clientset, namespace string, label string, isOperator bool) (bool, error) {
	// verify whether the namespace exist
	ns, err := GetNamespace(kubeClient, namespace)
	if err != nil {
		return false, fmt.Errorf("error getting namespace due to %+v", err)
	}

	var isExist bool
	if !isOperator {
		// check whether the operator already exist in the namespace
		isExist = IsOperatorExist(kubeClient, namespace)
		if isExist {
			return true, fmt.Errorf("synopsys operator is already running in %s namespace... namespace cannot be deleted", namespace)
		}
	}

	// iterate over each label and verify whether any Synopsys resources like Alert, Black Duck or OpsSight etc. exist in the namespace
	for synopsysLabel := range ns.Labels {
		if synopsysLabel != label {
			err := isSynopsysResourceExist(namespace, synopsysLabel)
			if err != nil {
				return true, err
			}
		}
	}

	return true, nil
}

// DeleteResourceNamespace deletes the namespace if none of the other resource types are running
func DeleteResourceNamespace(kubeClient *kubernetes.Clientset, resourceName string, namespace string, name string, isOperator bool) error {
	isExist, err := CheckResourceNamespace(kubeClient, namespace, fmt.Sprintf("synopsys.com/%s.%s", resourceName, name), isOperator)
	if !isExist {
		return err
	} else if isExist && err == nil {
		log.Infof("deleting %s namespace", namespace)
		err = DeleteNamespace(kubeClient, namespace)
		if err != nil {
			return fmt.Errorf("unable to delete the %s namespace because %+v", namespace, err)
		}
	} else {
		_, checkErr := CheckAndUpdateNamespace(kubeClient, resourceName, namespace, name, "", true)
		if checkErr != nil {
			return fmt.Errorf("%+v and hence deleting the Synopsys label from the namespace. %+v", err, checkErr)
		}
	}
	return err
}

// CheckAndUpdateNamespace will check whether the namespace is exist and if exist, update the version label in namespace of the updated/deleted resource
func CheckAndUpdateNamespace(kubeClient *kubernetes.Clientset, resourceName string, namespace string, name string, version string, isDelete bool) (bool, error) {
	ns, err := GetNamespace(kubeClient, namespace)
	if err == nil {
		err = updateLabelsInNamespace(kubeClient, ns, resourceName, namespace, name, version, isDelete)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, fmt.Errorf("unable to get %s namespace due to %+v", namespace, err)
}

// updateLabelsInNamespace will update the labels in the namespace
func updateLabelsInNamespace(kubeClient *kubernetes.Clientset, ns *corev1.Namespace, resourceName string, namespace string, name string, version string, isDelete bool) error {
	isLabelUpdated := false
	ns.Labels = InitLabels(ns.Labels)
	if isDelete {
		// delete from labels
		delete(ns.Labels, fmt.Sprintf("synopsys.com/%s.%s", resourceName, name))
		isLabelUpdated = true
	} else {
		// add or update the label
		if existingVersion, ok := ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", resourceName, name)]; !ok || existingVersion != version {
			ns.Labels = InitLabels(ns.Labels)
			ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", resourceName, name)] = version
			isLabelUpdated = true
		}
	}
	if isLabelUpdated {
		// update the namespace with labels updated
		_, err := UpdateNamespace(kubeClient, ns)
		if err != nil {
			return fmt.Errorf("unable to update the %s namespace labels due to %+v", namespace, err)
		}
	}
	return nil
}

// GetOperatorNamespaceByCRDScope get the operator namespace by CRD scope
func GetOperatorNamespaceByCRDScope(kubeClient *kubernetes.Clientset, crdName string, scope apiextensions.ResourceScope, namespace string) (string, error) {
	operatorNamespace := namespace
	if scope == apiextensions.ClusterScoped {
		operatorNamespace = metav1.NamespaceAll
	}

	// Get all Synopsys Operator running namespaces
	namespaces, err := GetOperatorNamespace(kubeClient, operatorNamespace)
	if err != nil {
		return "", err
	}

	// For each namespace, get the CRD names and see if the CRD name is belong to input CRD name
	for _, namespace := range namespaces {
		crdNames, err := GetCRDNamesFromConfigMap(kubeClient, namespace)
		if err != nil {
			continue
		}
		crds := StringToStringSlice(crdNames, ",")
		for _, crd := range crds {
			if crd == crdName {
				return namespace, nil
			}
		}
	}

	return "", fmt.Errorf("%s is not enabled in any of the Synopsys Operator", crdName)
}

// GetCRDNamesFromConfigMap get CRD names from the Synopsys Operator config map
func GetCRDNamesFromConfigMap(kubeClient *kubernetes.Clientset, namespace string) (string, error) {
	cm, err := GetConfigMap(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		return "", fmt.Errorf("error getting the Synopsys Operator config map due to %+v", err)
	}
	data := cm.Data["config.json"]
	var cmData map[string]interface{}
	err = json.Unmarshal([]byte(data), &cmData)
	if err != nil {
		log.Errorf("unable to unmarshal config map data due to %+v", err)
	}
	if crdNames, ok := cmData["CrdNames"]; ok {
		return crdNames.(string), nil
	}
	return "", fmt.Errorf("unable to find CRD names in the Synopsys Operator config map")
}

// StringToStringSlice slices s into all substrings separated by sep
func StringToStringSlice(s string, sep string) []string {
	if len(s) > 0 {
		return strings.Split(s, sep)
	}
	return make([]string, 0)
}

// InitLabels initialize the label
func InitLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return make(map[string]string, 0)
	}
	return labels
}

// InitAnnotations initialize the annotation
func InitAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return make(map[string]string, 0)
	}
	return annotations
}

func WaitForCRD(name string, interval time.Duration, timeout time.Duration, apiextensionsclient *apiextensionsclient.Clientset) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		crd, err := apiextensionsclient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, v := range crd.Status.Conditions {
			if v.Type == apiextensions.Established {
				if v.Status == apiextensions.ConditionTrue {
					return true, nil
				}
				break
			}
		}
		return false, nil
	})
}

// MergeEnvMaps will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvMaps(source, destination map[string]string) map[string]string {
	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for key, value := range source {
		if len(value) == 0 {
			delete(destination, key)
		} else {
			destination[key] = value
		}
	}
	return destination
}

// MergeEnvSlices will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvSlices(source, destination []string) []string {
	// create a destination map
	destinationMap := make(map[string]string)
	for _, value := range destination {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for _, value := range source {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapValue) == 0 {
				delete(destinationMap, mapKey)
			} else {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// convert destination map to string array
	mergedValues := []string{}
	for key, value := range destinationMap {
		mergedValues = append(mergedValues, fmt.Sprintf("%s:%s", key, value))
	}
	return mergedValues
}

// UniqueStringSlice returns a unique subset of the string slice provided.
func UniqueStringSlice(input []string) []string {
	output := []string{}
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			output = append(output, val)
		}
	}

	return output
}

// RemoveFromStringSlice will remove the string from the slice and it will maintain the order
func RemoveFromStringSlice(slice []string, str string) []string {
	for index, value := range slice {
		if value == str {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

// IsExposeServiceValid validates the expose service type
func IsExposeServiceValid(serviceType string) bool {
	switch strings.ToUpper(serviceType) {
	case NONE, NODEPORT, LOADBALANCER, OPENSHIFT:
		return true
	}
	return false
}
