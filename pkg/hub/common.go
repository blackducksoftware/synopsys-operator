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

package hub

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	kapi "github.com/blackducksoftware/horizon/pkg/api"
	types "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// CreateContainer will create the container
func CreateContainer(config *kapi.ContainerConfig, envs []*kapi.EnvConfig, volumeMounts []*kapi.VolumeMountConfig, port *kapi.PortConfig) *types.Container {

	container := types.NewContainer(kapi.ContainerConfig{
		Name:       config.Name,
		Image:      config.Image,
		PullPolicy: config.PullPolicy,
		MinCPU:     config.MinCPU,
		MaxCPU:     config.MaxCPU,
		MinMem:     config.MinMem,
		MaxMem:     config.MaxMem,
		Privileged: config.Privileged,
		Command:    config.Command,
		Args:       config.Args,
	})

	for _, env := range envs {
		container.AddEnv(*env)
	}

	for _, volumeMount := range volumeMounts {
		container.AddVolumeMount(*volumeMount)
	}

	container.AddPort(*port)

	return container
}

// CreateGCEPersistentDiskVolume will create a GCE Persistent disk volume for a pod
func CreateGCEPersistentDiskVolume(volumeName string, diskName string, fsType string) *types.Volume {
	gcePersistentDiskVol := types.NewGCEPersistentDiskVolume(kapi.GCEPersistentDiskVolumeConfig{
		VolumeName: volumeName,
		DiskName:   diskName,
		FSType:     fsType,
	})

	return gcePersistentDiskVol
}

// CreateEmptyDirVolume will create a empty directory for a pod
func CreateEmptyDirVolume(volumeName string, sizeLimit string) (*types.Volume, error) {
	emptyDirVol, err := types.NewEmptyDirVolume(kapi.EmptyDirVolumeConfig{
		VolumeName: volumeName,
		SizeLimit:  sizeLimit,
	})

	return emptyDirVol, err
}

// CreatePod will create the pod
func CreatePod(name string, volumes []*types.Volume, containers []*api.Container, initContainers []*api.Container, affinityConfigs []kapi.AffinityConfig) *types.Pod {
	pod := types.NewPod(kapi.PodConfig{
		Name: name,
	})

	for _, volume := range volumes {
		pod.AddVolume(volume)
	}

	pod.AddLabels(map[string]string{
		"app":  name,
		"tier": name,
	})

	for _, affinityConfig := range affinityConfigs {
		pod.AddAffinity(affinityConfig)
	}

	for _, containerConfig := range containers {
		container := CreateContainer(containerConfig.ContainerConfig, containerConfig.EnvConfigs, containerConfig.VolumeMounts, containerConfig.PortConfig)
		pod.AddContainer(container)
	}

	for _, initContainerConfig := range initContainers {
		initContainer := CreateContainer(initContainerConfig.ContainerConfig, initContainerConfig.EnvConfigs, initContainerConfig.VolumeMounts, initContainerConfig.PortConfig)
		err := pod.AddInitContainer(initContainer)
		if err != nil {
			log.Printf("failed to create the init container because %+v", err)
		}
	}

	return pod
}

// CreateDeployment will create a deployment
func CreateDeployment(deploymentConfig *kapi.DeploymentConfig, pod *types.Pod) *types.Deployment {
	deployment := types.NewDeployment(*deploymentConfig)

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  deploymentConfig.Name,
		"tier": deploymentConfig.Name,
	})
	deployment.AddPod(pod)
	return deployment
}

// CreateDeploymentFromContainer will create a deployment with multiple containers inside a pod
func CreateDeploymentFromContainer(deploymentConfig *kapi.DeploymentConfig, containers []*api.Container, volumes []*types.Volume, initContainers []*api.Container, affinityConfigs []kapi.AffinityConfig) *types.Deployment {
	pod := CreatePod(deploymentConfig.Name, volumes, containers, initContainers, affinityConfigs)
	deployment := CreateDeployment(deploymentConfig, pod)
	return deployment
}

// CreateService will create the service
func CreateService(name string, label string, namespace string, port string, target string, exp bool) *types.Service {
	svcConfig := kapi.ServiceConfig{
		Name:      name,
		Namespace: namespace,
	}
	// services w/ -exp are exposed
	if exp {
		svcConfig.IPServiceType = kapi.ClusterIPServiceTypeLoadBalancer
	}
	mySvc := types.NewService(svcConfig)
	portVal, _ := strconv.Atoi(port)
	myPort := &kapi.ServicePortConfig{
		Name:       fmt.Sprintf("port-" + name),
		Port:       int32(portVal),
		TargetPort: target,
		Protocol:   kapi.ProtocolTCP,
	}

	mySvc.AddPort(*myPort)
	mySvc.AddSelectors(map[string]string{"app": label})

	return mySvc
}

// CreateSecretFromFile will create the secret from file
func CreateSecretFromFile(clientset *kubernetes.Clientset, jsonFile string, namespace string, name string, dataKey string) (*v1.Secret, error) {
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Panicf("Unable to read the secret file %s due to error: %v\n", jsonFile, err)
	}

	return clientset.CoreV1().Secrets(namespace).Create(&v1.Secret{
		Type:       v1.SecretTypeOpaque,
		StringData: map[string]string{dataKey: string(file)},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// CreateSecret will create the secret
func CreateSecret(clientset *kubernetes.Clientset, namespace string, name string, stringData map[string]string) (*v1.Secret, error) {

	return clientset.CoreV1().Secrets(namespace).Create(&v1.Secret{
		Type:       v1.SecretTypeOpaque,
		StringData: stringData,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// ReadFromFile will read the file
func ReadFromFile(filePath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filePath)

	return file, err
}

// // CreateSecret will create the secret
// func CreateNewSecret(secretConfig *kapi.SecretConfig) *types.Secret {
//
// 	secret := types.NewSecret(kapi.SecretConfig{Namespace: secretConfig.Namespace, Name: secretConfig.Name, Type: secretConfig.Type})
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
// func CreateConfigMap(configMapConfig *kapi.ConfigMapConfig) *types.ConfigMap {
//
// 	configMap := types.NewConfigMap(kapi.ConfigMapConfig{Namespace: configMapConfig.Namespace, Name: configMapConfig.Name})
//
// 	configMap.AddData(configMapConfig.Data)
// 	configMap.AddLabels(configMapConfig.Labels)
// 	configMap.AddAnnotations(configMapConfig.Annotations)
//
// 	return configMap
// }

// CreateNamespace will create the namespace
func CreateNamespace(clientset *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
}

// GetNamespace will get the namespace
func GetNamespace(clientset *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
}

// DeleteNamespace will delete the namespace
func DeleteNamespace(clientset *kubernetes.Clientset, namespace string) error {
	return clientset.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
}

// GetAllPodsForNamespace will get all the pods corresponding to a namespace
func GetAllPodsForNamespace(clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
	return clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

// ValidatePodsAreRunning will validate whether the pods are running
func ValidatePodsAreRunning(clientset *kubernetes.Clientset, pods *corev1.PodList) {
	// Check whether all pods are running
	for _, podList := range pods.Items {
		for {
			pod, _ := clientset.CoreV1().Pods(podList.Namespace).Get(podList.Name, metav1.GetOptions{})
			if strings.EqualFold(string(pod.Status.Phase), "Running") {
				break
			}
			log.Infof("pod %s is in %s status!!!", pod.Name, string(pod.Status.Phase))
			time.Sleep(10 * time.Second)
		}
	}
}

// FilterPodByNamePrefix will filter the pod based on pod name prefix from a list a pods
func FilterPodByNamePrefix(pods *corev1.PodList) *corev1.Pod {
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "registration") {
			return &pod
		}
	}
	return nil
}

// CreateExecContainerRequest will create the request to exec into kubernetes pod
func CreateExecContainerRequest(clientset *kubernetes.Clientset, pod *corev1.Pod) *rest.Request {
	return clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", pod.Spec.Containers[0].Name).
		VersionedParams(&corev1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
}

// NewStringReader will convert string array to string reader object
func NewStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}

// NewKubeClientFromOutsideCluster will get the kube Configuration from outside the cluster
func newKubeClientFromOutsideCluster() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Errorf("error creating default client config: %s", err)
		return nil, err
	}
	return config, err
}

// GetKubeConfig will get the kube configuration
func GetKubeConfig() (*rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Infof("unable to get in cluster config due to %v", err)
		log.Infof("trying to use local config")
		config, err = newKubeClientFromOutsideCluster()
		if err != nil {
			log.Errorf("unable to retrive the local config due to %v", err)
			log.Panicf("failed to find a valid cluster config")
		}
	}

	return config, err
}

// GetService will get the service information for the input service name inside the input namespace
func GetService(clientset *kubernetes.Clientset, namespace string, serviceName string) (*v1.Service, error) {
	return clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
}

// IntToInt32 will convert from int to int32
func IntToInt32(i int) *int32 {
	j := int32(i)
	return &j
}

func getBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// RandomString will generate the random string
func RandomString(n int) (string, error) {
	b, err := getBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
