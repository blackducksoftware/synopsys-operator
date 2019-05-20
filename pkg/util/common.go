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

package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/version"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
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
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// OPENSHIFT denotes to create an OpenShift routes
	OPENSHIFT = "OPENSHIFT"
)

// CreateContainer will create the container
func CreateContainer(config *horizonapi.ContainerConfig, envs []*horizonapi.EnvConfig, volumeMounts []*horizonapi.VolumeMountConfig, ports []*horizonapi.PortConfig,
	actionConfig *horizonapi.ActionConfig, preStopConfig *horizonapi.ActionConfig, livenessProbeConfigs []*horizonapi.ProbeConfig, readinessProbeConfigs []*horizonapi.ProbeConfig) (*components.Container, error) {

	container, err := components.NewContainer(*config)
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		container.AddEnv(*env)
	}

	for _, volumeMount := range volumeMounts {
		container.AddVolumeMount(*volumeMount)
	}

	for _, port := range ports {
		container.AddPort(*port)
	}

	if actionConfig != nil {
		container.AddPostStartAction(*actionConfig)
	}

	// Adds a PreStop if given, originally added to enable graceful pg shutdown
	if preStopConfig != nil {
		container.AddPreStopAction(*preStopConfig)
	}

	for _, livenessProbe := range livenessProbeConfigs {
		container.AddLivenessProbe(*livenessProbe)
	}

	for _, readinessProbe := range readinessProbeConfigs {
		container.AddReadinessProbe(*readinessProbe)
	}

	return container, nil
}

// CreateGCEPersistentDiskVolume will create a GCE Persistent disk volume for a pod
func CreateGCEPersistentDiskVolume(volumeName string, diskName string, fsType string) *components.Volume {
	gcePersistentDiskVol := components.NewGCEPersistentDiskVolume(horizonapi.GCEPersistentDiskVolumeConfig{
		VolumeName: volumeName,
		DiskName:   diskName,
		FSType:     fsType,
	})

	return gcePersistentDiskVol
}

// CreateEmptyDirVolumeWithoutSizeLimit will create a empty directory for a pod
func CreateEmptyDirVolumeWithoutSizeLimit(volumeName string) (*components.Volume, error) {
	emptyDirVol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: volumeName,
	})

	return emptyDirVol, err
}

// CreatePersistentVolumeClaimVolume will create a PVC claim for a pod
func CreatePersistentVolumeClaimVolume(volumeName string, pvcName string) (*components.Volume, error) {
	pvcVol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		PVCName:    pvcName,
		VolumeName: volumeName,
	})

	return pvcVol, nil
}

// CreateEmptyDirVolume will create a empty directory for a pod
func CreateEmptyDirVolume(volumeName string, sizeLimit string) (*components.Volume, error) {
	emptyDirVol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: volumeName,
		SizeLimit:  sizeLimit,
	})

	return emptyDirVol, err
}

// CreateConfigMapVolume will mount the config map for a pod
func CreateConfigMapVolume(volumeName string, mapName string, defaultMode int) (*components.Volume, error) {
	configMapVol := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		DefaultMode:     IntToInt32(defaultMode),
		MapOrSecretName: mapName,
	})

	return configMapVol, nil
}

// CreateSecretVolume will mount the secret for a pod
func CreateSecretVolume(volumeName string, secretName string, defaultMode int) (*components.Volume, error) {
	secretVol := components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		DefaultMode:     IntToInt32(defaultMode),
		MapOrSecretName: secretName,
	})

	return secretVol, nil
}

// CreatePod will create the pod
func CreatePod(podConfig *PodConfig) (*components.Pod, error) {
	name := podConfig.Name
	// create pod config
	pod := components.NewPod(horizonapi.PodConfig{
		Name: name,
	})

	// set service account
	if len(podConfig.ServiceAccount) > 0 {
		pod.Spec.ServiceAccountName = podConfig.ServiceAccount
	}

	// add volumes
	for _, volume := range podConfig.Volumes {
		pod.AddVolume(volume)
	}

	// add labels
	pod.AddLabels(podConfig.Labels)

	// add pod affinities
	for affinityType, podAffinityConfigs := range podConfig.PodAffinityConfigs {
		for _, podAffinityConfig := range podAffinityConfigs {
			pod.AddPodAffinity(affinityType, *podAffinityConfig)
		}
	}

	// add pod anti affinities
	for affinityType, podAntiAffinityConfigs := range podConfig.PodAntiAffinityConfigs {
		for _, podAntiAffinityConfig := range podAntiAffinityConfigs {
			pod.AddPodAntiAffinity(affinityType, *podAntiAffinityConfig)
		}
	}

	// add node affinities
	for affinityType, nodeAffinityConfigs := range podConfig.NodeAffinityConfigs {
		for _, nodeAffinityConfig := range nodeAffinityConfigs {
			pod.AddNodeAffinity(affinityType, *nodeAffinityConfig)
		}
	}

	// add containers
	for _, containerConfig := range podConfig.Containers {
		container, err := CreateContainer(containerConfig.ContainerConfig, containerConfig.EnvConfigs, containerConfig.VolumeMounts, containerConfig.PortConfig,
			containerConfig.ActionConfig, containerConfig.PreStopConfig, containerConfig.LivenessProbeConfigs, containerConfig.ReadinessProbeConfigs)
		if len(containerConfig.Capabilities) > 0 {
			container.AddAddCapabilities(containerConfig.Capabilities)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create the container for pod %s because %+v", name, err)
		}
		pod.AddContainer(container)
	}

	// add init containers
	for _, initContainerConfig := range podConfig.InitContainers {
		initContainer, err := CreateContainer(initContainerConfig.ContainerConfig, initContainerConfig.EnvConfigs, initContainerConfig.VolumeMounts,
			initContainerConfig.PortConfig, initContainerConfig.ActionConfig, initContainerConfig.PreStopConfig, initContainerConfig.LivenessProbeConfigs, initContainerConfig.ReadinessProbeConfigs)
		if err != nil {
			return nil, fmt.Errorf("failed to create the init container for pod %s because %+v", name, err)
		}
		err = pod.AddInitContainer(initContainer)
		if err != nil {
			return nil, fmt.Errorf("failed to create the init container for pod %s because %+v", name, err)
		}
	}

	if len(podConfig.ImagePullSecrets) > 0 {
		pod.AddImagePullSecrets(podConfig.ImagePullSecrets)
	}

	return pod, nil
}

// CreateDeployment will create a deployment
func CreateDeployment(deploymentConfig *horizonapi.DeploymentConfig, pod *components.Pod, labelSelector map[string]string) *components.Deployment {
	deployment := components.NewDeployment(*deploymentConfig)
	deployment.AddMatchLabelsSelectors(labelSelector)
	deployment.AddPod(pod)
	return deployment
}

// CreateDeploymentFromContainer will create a deployment with multiple containers inside a pod
func CreateDeploymentFromContainer(deploymentConfig *horizonapi.DeploymentConfig, podConfig *PodConfig, labelSelector map[string]string) (*components.Deployment, error) {
	podConfig.Name = deploymentConfig.Name
	pod, err := CreatePod(podConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create pod for the deployment %s due to %+v", deploymentConfig.Name, err)
	}
	deployment := CreateDeployment(deploymentConfig, pod, labelSelector)
	return deployment, nil
}

// CreateReplicationController will create a replication controller
func CreateReplicationController(replicationControllerConfig *horizonapi.ReplicationControllerConfig, pod *components.Pod, labels map[string]string,
	labelSelector map[string]string) *components.ReplicationController {
	rc := components.NewReplicationController(*replicationControllerConfig)
	rc.AddSelectors(labelSelector)
	rc.AddLabels(labels)
	rc.AddPod(pod)
	return rc
}

// CreateReplicationControllerFromContainer will create a replication controller with multiple containers inside a pod
func CreateReplicationControllerFromContainer(replicationControllerConfig *horizonapi.ReplicationControllerConfig, podConfig *PodConfig, labelSelector map[string]string) (*components.ReplicationController, error) {
	podConfig.Name = replicationControllerConfig.Name
	pod, err := CreatePod(podConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create pod for the replication controller %s due to %+v", replicationControllerConfig.Name, err)
	}
	// TODO:pod.AddImagePullSecrets()
	rc := CreateReplicationController(replicationControllerConfig, pod, podConfig.Labels, labelSelector)
	return rc, nil
}

// CreateStateFulSet will create a statefulset
func CreateStateFulSet(stateFulSetConfig *horizonapi.StatefulSetConfig, pod *components.Pod) *components.StatefulSet {
	stateFulSet := components.NewStatefulSet(*stateFulSetConfig)
	stateFulSet.AddMatchLabelsSelectors(pod.GetLabels())
	stateFulSet.AddPod(pod)
	return stateFulSet
}

// CreateStateFulSetFromContainer will create a statefulset from a container
func CreateStateFulSetFromContainer(stateFulSetConfig *horizonapi.StatefulSetConfig, serviceAccount string, containers []*Container, volumes []*components.Volume, initContainers []*Container, affinityConfigs map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig, labelSelector map[string]string) *components.StatefulSet {
	pod, _ := CreatePod(&PodConfig{
		Name:                stateFulSetConfig.Name,
		ServiceAccount:      serviceAccount,
		Volumes:             volumes,
		Containers:          containers,
		InitContainers:      initContainers,
		NodeAffinityConfigs: affinityConfigs,
		Labels:              labelSelector,
	})
	stateFulSet := CreateStateFulSet(stateFulSetConfig, pod)
	return stateFulSet
}

// CreateService will create the service
func CreateService(name string, selectLabel map[string]string, namespace string, port int32, target int32, serviceType horizonapi.ServiceType, label map[string]string) *components.Service {
	svcConfig := horizonapi.ServiceConfig{
		Name:      name,
		Namespace: namespace,
		Type:      serviceType,
	}

	mySvc := components.NewService(svcConfig)
	myPort := &horizonapi.ServicePortConfig{
		Name:       fmt.Sprintf("port-%d", port),
		Port:       port,
		TargetPort: fmt.Sprint(target),
		Protocol:   horizonapi.ProtocolTCP,
	}

	mySvc.AddPort(*myPort)
	mySvc.AddSelectors(selectLabel)
	mySvc.AddLabels(label)

	return mySvc
}

// CreateServiceWithMultiplePort will create the service with multiple port
func CreateServiceWithMultiplePort(name string, selectLabel map[string]string, namespace string, ports []int32, serviceType horizonapi.ServiceType, label map[string]string) *components.Service {
	svcConfig := horizonapi.ServiceConfig{
		Name:      name,
		Namespace: namespace,
		Type:      serviceType,
	}

	mySvc := components.NewService(svcConfig)

	for _, port := range ports {
		myPort := &horizonapi.ServicePortConfig{
			Name:       fmt.Sprintf("port-%d", port),
			Port:       port,
			TargetPort: fmt.Sprint(port),
			Protocol:   horizonapi.ProtocolTCP,
		}
		mySvc.AddPort(*myPort)
	}

	mySvc.AddSelectors(selectLabel)
	mySvc.AddLabels(label)

	return mySvc
}

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
func UpdateSecret(clientset *kubernetes.Clientset, namespace string, secret *corev1.Secret) error {
	_, err := clientset.CoreV1().Secrets(namespace).Update(secret)
	return err
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
func UpdateConfigMap(clientset *kubernetes.Clientset, namespace string, configMap *corev1.ConfigMap) error {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Update(configMap)
	return err
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

// GetReplicationController will get the replication controller corresponding to a namespace and name
func GetReplicationController(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.ReplicationController, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).Get(name, metav1.GetOptions{})
}

// ListReplicationControllers will get the replication controllers corresponding to a namespace
func ListReplicationControllers(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.ReplicationControllerList, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
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

// CreatePersistentVolumeClaim will create the persistent volume claim
func CreatePersistentVolumeClaim(name string, namespace string, pvcClaimSize string, storageClass string, accessMode horizonapi.PVCAccessModeType) (*components.PersistentVolumeClaim, error) {

	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string
	if len(storageClass) == 0 {
		class = nil
	} else {
		class = &storageClass
	}

	postgresPVC, err := components.NewPersistentVolumeClaim(horizonapi.PVCConfig{
		Name:      name,
		Namespace: namespace,
		// VolumeName: createHub.Name,
		Size:  pvcClaimSize,
		Class: class,
	})
	if err != nil {
		return nil, err
	}
	postgresPVC.AddAccessMode(accessMode)

	return postgresPVC, nil
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

// ValidatePodsAreRunningInNamespace will validate whether the pods are running in a given namespace
func ValidatePodsAreRunningInNamespace(clientset *kubernetes.Clientset, namespace string, timeoutInSeconds int64) error {
	// timer starts the timer for timeoutInSeconds. If the task doesn't completed, return error
	timeout := time.NewTimer(time.Duration(timeoutInSeconds) * time.Second)
	// ticker starts and execute the task for every n intervals
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timeout.C:
			ticker.Stop()
			return fmt.Errorf("the pods weren't able to start - timing out after %d seconds", timeoutInSeconds)
		case <-ticker.C:
			pods, err := ListPods(clientset, namespace)
			if err != nil {
				timeout.Stop()
				ticker.Stop()
				return fmt.Errorf("unable to list the pods in namespace %s due to %+v", namespace, err)
			}
			if ValidatePodsAreRunning(clientset, pods) {
				timeout.Stop()
				ticker.Stop()
				return nil
			}
		}
	}
}

// ValidatePodsAreRunning will validate whether the pods are running
func ValidatePodsAreRunning(clientset *kubernetes.Clientset, pods *corev1.PodList) bool {
	// Check whether all pods are running
	for _, podList := range pods.Items {
		pod, _ := clientset.CoreV1().Pods(podList.Namespace).Get(podList.Name, metav1.GetOptions{})
		if !strings.EqualFold(string(pod.Status.Phase), "Running") {
			log.Infof("pod %s is in %s status...", pod.Name, string(pod.Status.Phase))
			return false
		}
	}
	return true
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

// CreateHub will create hub in the cluster
func CreateHub(hubClientset *hubclientset.Clientset, namespace string, createHub *blackduckapi.Blackduck) (*blackduckapi.Blackduck, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Create(createHub)
}

// ListHubs will list all hubs in the cluster
func ListHubs(hubClientset *hubclientset.Clientset, namespace string) (*blackduckapi.BlackduckList, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).List(metav1.ListOptions{})
}

// GetHub will get hubs in the cluster
func GetHub(hubClientset *hubclientset.Clientset, namespace string, name string) (*blackduckapi.Blackduck, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Get(name, metav1.GetOptions{})
}

// GetBlackducks gets all blackducks
func GetBlackducks(clientSet *hubclientset.Clientset) (*blackduckapi.BlackduckList, error) {
	return clientSet.SynopsysV1().Blackducks(metav1.NamespaceAll).List(metav1.ListOptions{})
}

// UpdateBlackduck will update Blackduck in the cluster
func UpdateBlackduck(blackduckClientset *hubclientset.Clientset, namespace string, blackduck *blackduckapi.Blackduck) (*blackduckapi.Blackduck, error) {
	return blackduckClientset.SynopsysV1().Blackducks(namespace).Update(blackduck)
}

// UpdateBlackducks will update a set of Blackducks in the cluster
func UpdateBlackducks(clientSet *hubclientset.Clientset, blackduckCRDs []blackduckapi.Blackduck) error {
	for _, crd := range blackduckCRDs {
		_, err := UpdateBlackduck(clientSet, crd.Spec.Namespace, &crd)
		if err != nil {
			return err
		}
	}
	return nil
}

// WatchHubs will watch for hub events in the cluster
func WatchHubs(hubClientset *hubclientset.Clientset, namespace string) (watch.Interface, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Watch(metav1.ListOptions{})
}

// CreateOpsSight will create opsSight in the cluster
func CreateOpsSight(opssightClientset *opssightclientset.Clientset, namespace string, opssight *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).Create(opssight)
}

// ListOpsSights will list all opssights in the cluster
func ListOpsSights(opssightClientset *opssightclientset.Clientset, namespace string) (*opssightapi.OpsSightList, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).List(metav1.ListOptions{})
}

// GetOpsSight will get OpsSight in the cluster
func GetOpsSight(opssightClientset *opssightclientset.Clientset, namespace string, name string) (*opssightapi.OpsSight, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).Get(name, metav1.GetOptions{})
}

// GetOpsSights gets all opssights
func GetOpsSights(clientSet *opssightclientset.Clientset) (*opssightapi.OpsSightList, error) {
	return clientSet.SynopsysV1().OpsSights(metav1.NamespaceAll).List(metav1.ListOptions{})
}

// UpdateOpsSight will update OpsSight in the cluster
func UpdateOpsSight(opssightClientset *opssightclientset.Clientset, namespace string, opssight *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).Update(opssight)
}

// UpdateOpsSights will update a set of OpsSights in the cluster
func UpdateOpsSights(clientSet *opssightclientset.Clientset, opsSightCRDs []opssightapi.OpsSight) error {
	for _, crd := range opsSightCRDs {
		_, err := UpdateOpsSight(clientSet, crd.Spec.Namespace, &crd)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateAlert will create alert in the cluster
func CreateAlert(alertClientset *alertclientset.Clientset, namespace string, createAlert *alertapi.Alert) (*alertapi.Alert, error) {
	return alertClientset.SynopsysV1().Alerts(namespace).Create(createAlert)
}

// ListAlerts will list all alerts in the cluster
func ListAlerts(clientSet *alertclientset.Clientset, namespace string) (*alertapi.AlertList, error) {
	return clientSet.SynopsysV1().Alerts(namespace).List(metav1.ListOptions{})
}

// GetAlert will get Alert in the cluster
func GetAlert(clientSet *alertclientset.Clientset, namespace string, name string) (*alertapi.Alert, error) {
	return clientSet.SynopsysV1().Alerts(namespace).Get(name, metav1.GetOptions{})
}

// GetAlerts gets all alerts
func GetAlerts(clientSet *alertclientset.Clientset) (*alertapi.AlertList, error) {
	return clientSet.SynopsysV1().Alerts(metav1.NamespaceAll).List(metav1.ListOptions{})
}

// UpdateAlert will update an Alert in the cluster
func UpdateAlert(clientSet *alertclientset.Clientset, namespace string, alert *alertapi.Alert) (*alertapi.Alert, error) {
	return clientSet.SynopsysV1().Alerts(namespace).Update(alert)
}

// UpdateAlerts will update a set of Alerts in the cluster
func UpdateAlerts(clientSet *alertclientset.Clientset, alertCRDs []alertapi.Alert) error {
	for _, crd := range alertCRDs {
		_, err := UpdateAlert(clientSet, crd.Spec.Namespace, &crd)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListHubPV will list all the persistent volumes attached to each hub in the cluster
func ListHubPV(hubClientset *hubclientset.Clientset, namespace string) (map[string]string, error) {
	var pvList map[string]string
	pvList = make(map[string]string)
	hubs, err := ListHubs(hubClientset, namespace)
	if err != nil {
		log.Errorf("unable to list the hubs due to %+v", err)
		return pvList, err
	}
	for _, hub := range hubs.Items {
		if hub.Spec.PersistentStorage {
			pvList[hub.Name] = fmt.Sprintf("%s (%s)", hub.Name, hub.Status.PVCVolumeName["blackduck-postgres"])
		}
	}
	return pvList, nil
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

// CreateServiceAccount creates a service account
func CreateServiceAccount(namespace string, name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      name,
		Namespace: namespace,
	})

	return serviceAccount
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

// CreateClusterRoleBinding creates a cluster role binding
func CreateClusterRoleBinding(namespace string, name string, serviceAccountName string, clusterRoleAPIGroup string, clusterRoleKind string, clusterRoleName string) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      serviceAccountName,
		Namespace: namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: clusterRoleAPIGroup,
		Kind:     clusterRoleKind,
		Name:     clusterRoleName,
	})

	return clusterRoleBinding
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

// IsClusterRoleBindingSubjectExist checks whether the namespace is already exist in the subject of cluster role binding
func IsClusterRoleBindingSubjectExist(subjects []rbacv1.Subject, namespace string, name string) bool {
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

// DeleteClusterRole delete a cluster role binding
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

// GetRouteClient attempts to get a Route Client. It returns nil if it
// fails due to an error or due to being on kubernetes (doesn't support routes)
func GetRouteClient(restConfig *rest.Config) *routeclient.RouteV1Client {
	routeClient, err := routeclient.NewForConfig(restConfig)
	if routeClient == nil || err != nil {
		log.Debugf("unable to get route client")
		return nil
	}
	_, err = ListRoutes(routeClient, "default", "")
	if err != nil {
		log.Debugf("ignoring routes for kubernetes cluster")
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

// GetRouteComponent returns the route component
func GetRouteComponent(routeClient *routeclient.RouteV1Client, route *api.Route, labels map[string]string) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      route.Name,
			Namespace: route.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			TLS: &routev1.TLSConfig{Termination: route.TLSTerminationType},
			To: routev1.RouteTargetReference{
				Kind: route.Kind,
				Name: route.ServiceName,
			},
			Port: &routev1.RoutePort{TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: route.PortName}},
		},
	}
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
func WaitUntilPodsAreReady(clientset *kubernetes.Clientset, namespace string, labelSelector string, timeoutInSeconds int64) (bool, error) {
	// timer starts the timer for timeoutInSeconds. If the task doesn't completed, return error
	timeout := time.NewTimer(time.Duration(timeoutInSeconds) * time.Second)
	// ticker starts and execute the task for every n intervals
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timeout.C:
			ticker.Stop()
			return false, fmt.Errorf("[NS: %s | Label: %s] the pods weren't ready - timing out after %d seconds", namespace, labelSelector, timeoutInSeconds)
		case <-ticker.C:
			ready, err := IsPodReady(clientset, namespace, labelSelector)
			if err != nil || ready {
				timeout.Stop()
				ticker.Stop()
				return ready, err
			}
		}
	}
}

// IsPodReady returns whether the pods are ready or not
func IsPodReady(clientset *kubernetes.Clientset, namespace string, labelSelector string) (bool, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return false, err
	}

	for _, p := range pods.Items {
		for _, condition := range p.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				return false, nil
			}
		}
	}
	return true, nil
}

// GetOperatorNamespace returns the namespace of Synopsys Operator based on the labels
func GetOperatorNamespace(clientset *kubernetes.Clientset) (string, error) {
	// check if operator is already installed
	rcs, err := ListReplicationControllers(clientset, metav1.NamespaceAll, "app=synopsys-operator,component=operator")
	if err == nil && len(rcs.Items) > 0 {
		return rcs.Items[0].Namespace, nil
	}
	deployments, err := ListDeployments(clientset, metav1.NamespaceAll, "app=synopsys-operator,component=operator")
	if err == nil && len(deployments.Items) > 0 {
		return deployments.Items[0].Namespace, nil
	}
	pods, err := ListPodsWithLabels(clientset, metav1.NamespaceAll, "app=synopsys-operator,component=operator")
	if err == nil && len(pods.Items) > 0 {
		return pods.Items[0].Namespace, nil
	}
	return "", fmt.Errorf("Synopsys Operator namespace not found")
}

// GetOperatorClusterRole returns the cluster role of Synopsys Operator based on the labels
func GetOperatorClusterRole(clientset *kubernetes.Clientset) (string, error) {
	crs, err := ListClusterRoles(clientset, "app=synopsys-operator,component=operator")

	if err != nil || len(crs.Items) == 0 {
		namespace, err := GetOperatorNamespace(clientset)
		if err != nil {
			return "", fmt.Errorf("Synopsys Operator namespace not found")
		}

		crs, err = ListClusterRoles(clientset, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil || len(crs.Items) == 0 {
			return "", fmt.Errorf("Synopsys Operator cluster role not found")
		}
	}
	return crs.Items[0].Name, nil
}

// GetOperatorClusterRoleBinding returns the cluster role bindings of Synopsys Operator based on the labels
func GetOperatorClusterRoleBinding(clientset *kubernetes.Clientset) (string, error) {
	crbs, err := ListClusterRoleBindings(clientset, "app=synopsys-operator,component=operator")

	if err != nil || len(crbs.Items) == 0 {
		namespace, err := GetOperatorNamespace(clientset)
		if err != nil {
			return "", fmt.Errorf("Synopsys Operator namespace not found")
		}

		crbs, err = ListClusterRoleBindings(clientset, fmt.Sprintf("olm.owner.namespace=%s,olm.owner.kind=ClusterServiceVersion", namespace))
		if err != nil || len(crbs.Items) == 0 {
			return "", fmt.Errorf("Synopsys Operator cluster role binding not found")
		}
	}
	return crbs.Items[0].Name, nil
}

// GetKubernetesVersion will return the kubernetes version
func GetKubernetesVersion(clientset *kubernetes.Clientset) (string, error) {
	k, err := clientset.Discovery().ServerVersion()
	if k != nil {
		return k.GitVersion, nil
	}
	return "", err
}

// GetOcVersion will return the version of openshift
func GetOcVersion(clientset *kubernetes.Clientset) (string, error) {
	body, err := clientset.Discovery().RESTClient().Get().AbsPath("/version/openshift").Do().Raw()
	if err != nil {
		return "", err
	}

	var info version.Info
	err = json.Unmarshal(body, &info)
	if err != nil {
		return "", fmt.Errorf("unable to parse the server version: %v", err)
	}

	return info.GitVersion, err
}
