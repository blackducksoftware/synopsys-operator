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

package store

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	sizev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/size/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/size"
	sizeclientset "github.com/blackducksoftware/synopsys-operator/pkg/size/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Components contains the list of components to be added/updated
type Components struct {
	ClusterRole        map[types.ComponentName]types.ClusterRoleCreater
	ClusterRoleBinding map[types.ComponentName]types.ClusterRoleBindingCreater
	Configmap          map[types.ComponentName]types.ConfigmapCreater
	Deployment         map[types.ComponentName]types.DeploymentCreater
	PVC                map[types.ComponentName]types.PvcCreater
	Rc                 map[types.ComponentName]types.ReplicationControllerCreater
	Route              map[types.ComponentName]types.RouteCreater
	Secret             map[types.ComponentName]types.SecretCreater
	Service            map[types.ComponentName]types.ServiceCreater
	ServiceAccount     map[types.ComponentName]types.ServiceAccountCreater
}

// ComponentStore stores the components
var ComponentStore Components

// Register registers the components to the store
func Register(name types.ComponentName, function interface{}) {
	switch function.(type) {
	case func(*types.PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (types.DeploymentInterface, error):
		if ComponentStore.Deployment == nil {
			ComponentStore.Deployment = make(map[types.ComponentName]types.DeploymentCreater)
		}
		ComponentStore.Deployment[name] = function.(func(*types.PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (types.DeploymentInterface, error))
	case func(*types.PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (types.ReplicationControllerInterface, error):
		if ComponentStore.Rc == nil {
			ComponentStore.Rc = make(map[types.ComponentName]types.ReplicationControllerCreater)
		}
		ComponentStore.Rc[name] = function.(func(*types.PodResource, *protoform.Config, *kubernetes.Clientset, interface{}) (types.ReplicationControllerInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ServiceInterface, error):
		if ComponentStore.Service == nil {
			ComponentStore.Service = make(map[types.ComponentName]types.ServiceCreater)
		}
		ComponentStore.Service[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ServiceInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ConfigMapInterface, error):
		if ComponentStore.Configmap == nil {
			ComponentStore.Configmap = make(map[types.ComponentName]types.ConfigmapCreater)
		}
		ComponentStore.Configmap[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ConfigMapInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.PVCInterface, error):
		if ComponentStore.PVC == nil {
			ComponentStore.PVC = make(map[types.ComponentName]types.PvcCreater)
		}
		ComponentStore.PVC[name] = function.(func(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck interface{}) (types.PVCInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.SecretInterface, error):
		if ComponentStore.Secret == nil {
			ComponentStore.Secret = make(map[types.ComponentName]types.SecretCreater)
		}
		ComponentStore.Secret[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.SecretInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ClusterRoleInterface, error):
		if ComponentStore.ClusterRole == nil {
			ComponentStore.ClusterRole = make(map[types.ComponentName]types.ClusterRoleCreater)
		}
		ComponentStore.ClusterRole[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ClusterRoleInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ClusterRoleBindingInterface, error):
		if ComponentStore.ClusterRoleBinding == nil {
			ComponentStore.ClusterRoleBinding = make(map[types.ComponentName]types.ClusterRoleBindingCreater)
		}
		ComponentStore.ClusterRoleBinding[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ClusterRoleBindingInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ServiceAccountInterface, error):
		if ComponentStore.ServiceAccount == nil {
			ComponentStore.ServiceAccount = make(map[types.ComponentName]types.ServiceAccountCreater)
		}
		ComponentStore.ServiceAccount[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.ServiceAccountInterface, error))
	case func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.RouteInterface, error):
		if ComponentStore.Route == nil {
			ComponentStore.Route = make(map[types.ComponentName]types.RouteCreater)
		}
		ComponentStore.Route[name] = function.(func(*protoform.Config, *kubernetes.Clientset, interface{}) (types.RouteInterface, error))
	default:
		log.Fatalf("couldn't load %s because unable to find the type %+v", name, reflect.TypeOf(function))
	}
}

// customResource holds the custom resource configuration
type customResource struct {
	namespace             string
	size                  string
	imageRegistries       []string
	registryConfiguration *api.RegistryConfiguration
}

// GetComponents get the components and generate the corresponding horizon object
func GetComponents(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, sizeClient *sizeclientset.Clientset, cr interface{}) (*api.ComponentList, error) {
	cp := &api.ComponentList{}

	// get the custom resource info
	customResource, err := getCR(cr)
	if err != nil {
		return &api.ComponentList{}, fmt.Errorf("unable to get the components because %+v", err)
	}

	// Cluster Roles
	crs, err := generateClusterRole(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.ClusterRoles = crs

	// Cluster Role Bindings
	crbs, err := generateClusterRoleBinding(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.ClusterRoleBindings = crbs

	// Config Maps
	cm, err := generateConfigmap(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.ConfigMaps = cm

	// Deployments
	deployments, err := generateDeployment(v, config, kubeclient, sizeClient, customResource, cr)
	if err != nil {
		return cp, err
	}
	cp.Deployments = deployments

	// PVC
	pvcs, err := generatePVCs(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.PersistentVolumeClaims = pvcs

	// Rc
	rcs, err := generateRc(v, config, kubeclient, sizeClient, customResource, cr)
	if err != nil {
		return cp, err
	}
	cp.ReplicationControllers = rcs

	// Routes
	routes, err := generateRoute(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.Routes = routes

	// Secret
	secrets, err := generateSecret(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.Secrets = secrets

	// Services
	services, err := generateService(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.Services = services

	// Service Accounts
	serviceAccounts, err := generateServiceAccount(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.ServiceAccounts = serviceAccounts

	return cp, nil
}

func getCR(cr interface{}) (*customResource, error) {
	spec := reflect.ValueOf(cr).Elem().FieldByName("Spec")

	namespace, ok := spec.FieldByName("Namespace").Interface().(string)
	if !ok {
		return nil, fmt.Errorf("namespace can't be retrieved from the custom resource because of type mismatch. expected: string, actual: %+v", reflect.TypeOf(namespace))
	}

	size, ok := spec.FieldByName("Size").Interface().(string)
	if !ok {
		return nil, fmt.Errorf("size can't be retrieved from the custom resource because of type mismatch. expected: string, actual: %+v", reflect.TypeOf(size))
	}

	imageRegistries, ok := spec.FieldByName("ImageRegistries").Interface().([]string)
	if !ok {
		return nil, fmt.Errorf("image registries can't be retrieved from the custom resource because of type mismatch. expected: []string, actual: %+v", reflect.TypeOf(imageRegistries))
	}

	registryConfiguration, ok := spec.FieldByName("RegistryConfiguration").Interface().(api.RegistryConfiguration)
	if !ok {
		return nil, fmt.Errorf("registry configuration can't be retrieved from the custom resource because of type mismatch. expected: api.RegistryConfiguration, actual: %+v", reflect.TypeOf(registryConfiguration))
	}

	return &customResource{namespace: namespace, size: size, imageRegistries: imageRegistries, registryConfiguration: &registryConfiguration}, nil
}

func generateDeployment(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, sizeClient *sizeclientset.Clientset, customResource *customResource, cr interface{}) ([]*components.Deployment, error) {
	deployments := make([]*components.Deployment, 0)
	// Size

	if len(customResource.size) == 0 {
		return nil, errors.New("size couldn't be found")
	}

	var componentSize *sizev1.Size
	var err error

	if config.DryRun {
		componentSize = size.GetDefaultSize(customResource.size)
		if componentSize == nil {
			return nil, fmt.Errorf("couldn't find size %s", customResource.size)
		}
	} else {
		componentSize, err = sizeClient.SynopsysV1().Sizes(config.Namespace).Get(strings.ToLower(customResource.size), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	// RC
	for k, publicPod := range v.Deployments {
		pod := generatePodResource(publicPod, componentSize, customResource, k)

		component, ok := ComponentStore.Deployment[publicPod.Identifier]
		if !ok {
			return nil, fmt.Errorf("deployment %s couldn't be found", publicPod.Identifier)
		}

		deploymentCreater, err := component(pod, config, kubeclient, cr)
		comp, err := deploymentCreater.GetDeployment()
		if err != nil {
			return nil, err
		}
		if comp != nil {
			deployments = append(deployments, comp)
		}
	}
	return deployments, nil
}

func generateRc(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, sizeClient *sizeclientset.Clientset, customResource *customResource, cr interface{}) ([]*components.ReplicationController, error) {
	rcs := make([]*components.ReplicationController, 0)
	// Size

	if len(customResource.size) == 0 {
		return nil, errors.New("size couldn't be found")
	}

	var componentSize *sizev1.Size
	var err error

	if config.DryRun {
		componentSize = size.GetDefaultSize(customResource.size)
		if componentSize == nil {
			return nil, fmt.Errorf("couldn't find size %s", customResource.size)
		}
	} else {
		componentSize, err = sizeClient.SynopsysV1().Sizes(config.Namespace).Get(strings.ToLower(customResource.size), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	// RC
	for k, publicPod := range v.RCs {
		pod := generatePodResource(publicPod, componentSize, customResource, k)

		component, ok := ComponentStore.Rc[publicPod.Identifier]
		if !ok {
			return nil, fmt.Errorf("rc %s couldn't be found", publicPod.Identifier)
		}

		rcCreater, err := component(pod, config, kubeclient, cr)
		comp, err := rcCreater.GetRc()
		if err != nil {
			return nil, err
		}
		if comp != nil {
			rcs = append(rcs, comp)
		}
	}
	return rcs, nil
}

func generatePodResource(v types.PublicPodResource, componentSize *sizev1.Size, customResource *customResource, componentName string) *types.PodResource {
	defaultPodResource, found := componentSize.Spec.PodResources[componentName]
	podResource := &types.PodResource{
		Namespace:  customResource.namespace,
		Replicas:   1,
		Containers: map[types.ContainerName]types.Container{},
	}

	if found {
		podResource.Replicas = defaultPodResource.Replica
	}

	for containerName, defaultImage := range v.Container {
		tmpContainer := types.Container{
			Image: generateImageTag(defaultImage, customResource.imageRegistries, customResource.registryConfiguration),
		}
		if found {
			if containerSize, containerSizeFound := defaultPodResource.ContainerLimit[string(containerName)]; containerSizeFound {
				tmpContainer.MinMem = containerSize.MinMem
				tmpContainer.MaxMem = containerSize.MaxMem
				tmpContainer.MinCPU = containerSize.MinCPU
				tmpContainer.MaxCPU = containerSize.MaxCPU
			}
		}
		podResource.Containers[containerName] = tmpContainer
	}

	return podResource
}

func generateSecret(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.Secret, error) {
	secrets := make([]*components.Secret, 0)
	for _, v := range v.Secrets {
		component, ok := ComponentStore.Secret[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find secret %s", v)
		}
		secretCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := secretCreater.GetSecret()
		if err != nil {
			return nil, err
		}
		if res != nil {
			secrets = append(secrets, res)
		}
	}
	return secrets, nil
}

func generateService(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.Service, error) {
	services := make([]*components.Service, 0)
	for _, v := range v.Services {
		component, ok := ComponentStore.Service[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find service %s", v)
		}
		serviceCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := serviceCreater.GetService()
		if err != nil {
			return nil, err
		}
		if res != nil {
			services = append(services, res)
		}
	}
	return services, nil
}

func generateClusterRole(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.ClusterRole, error) {
	clusterRoles := make([]*components.ClusterRole, 0)
	for _, v := range v.ClusterRoles {
		component, ok := ComponentStore.ClusterRole[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find service account %s", v)
		}
		crCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := crCreater.GetClusterRole()
		if err != nil {
			return nil, err
		}
		if res != nil {
			clusterRoles = append(clusterRoles, res)
		}
	}
	return clusterRoles, nil
}

func generateClusterRoleBinding(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.ClusterRoleBinding, error) {
	clusterRoleBindings := make([]*components.ClusterRoleBinding, 0)
	for _, v := range v.ClusterRoles {
		component, ok := ComponentStore.ClusterRoleBinding[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find service account %s", v)
		}
		crbCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := crbCreater.GetClusterRoleBinding()
		if err != nil {
			return nil, err
		}
		if res != nil {
			clusterRoleBindings = append(clusterRoleBindings, res)
		}
	}
	return clusterRoleBindings, nil
}

func generateServiceAccount(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.ServiceAccount, error) {
	serviceAccounts := make([]*components.ServiceAccount, 0)
	for _, v := range v.ServiceAccounts {
		component, ok := ComponentStore.ServiceAccount[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find service account %s", v)
		}
		saCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := saCreater.GetServiceAccount()
		if err != nil {
			return nil, err
		}
		if res != nil {
			serviceAccounts = append(serviceAccounts, res)
		}
	}
	return serviceAccounts, nil
}

func generateRoute(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*api.Route, error) {
	routes := make([]*api.Route, 0)
	for _, v := range v.Routes {
		component, ok := ComponentStore.Route[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find service account %s", v)
		}
		routeCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := routeCreater.GetRoute()
		if err != nil {
			return nil, err
		}
		if res != nil {
			routes = append(routes, res)
		}
	}
	return routes, nil
}

func generateConfigmap(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.ConfigMap, error) {
	cms := make([]*components.ConfigMap, 0)
	for _, v := range v.ConfigMaps {
		component, ok := ComponentStore.Configmap[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find configmap %s", v)
		}
		cmCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := cmCreater.GetCM()
		if err != nil {
			return nil, err
		}
		if res != nil {
			cms = append(cms, res)
		}
	}
	return cms, nil
}

func generatePVCs(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, cr interface{}) ([]*components.PersistentVolumeClaim, error) {
	pvcs := make([]*components.PersistentVolumeClaim, 0)
	for _, v := range v.PVC {
		component, ok := ComponentStore.PVC[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find pvc %s", v)
		}
		pvcCreater, err := component(config, kubeclient, cr)
		if err != nil {
			return nil, err
		}
		res, err := pvcCreater.GetPVCs()
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			pvcs = append(pvcs, res...)
		}
	}
	return pvcs, nil
}

func generateImageTag(defaultImage string, imageRegistries []string, registryConfig *api.RegistryConfiguration) string {
	if len(imageRegistries) > 0 {
		imageName, err := util.GetImageName(defaultImage)
		if err != nil {
			return defaultImage
		}
		defaultImage = getFullContainerNameFromImageRegistryConf(imageName, imageRegistries, defaultImage)
	}

	if len(registryConfig.Registry) > 0 && len(registryConfig.Namespace) > 0 {
		return getRegistryConfiguration(defaultImage, registryConfig)
	}
	return defaultImage
}

func getRegistryConfiguration(image string, registryConfig *api.RegistryConfiguration) string {
	if len(registryConfig.Registry) > 0 && len(registryConfig.Namespace) > 0 {
		imageName, err := util.GetImageName(image)
		if err != nil {
			return image
		}
		imageTag, err := util.GetImageTag(image)
		if err != nil {
			return image
		}
		return fmt.Sprintf("%s/%s/%s:%s", registryConfig.Registry, registryConfig.Namespace, imageName, imageTag)
	}
	return image
}

func getFullContainerNameFromImageRegistryConf(baseContainer string, images []string, defaultImage string) string {
	for _, reg := range images {
		// normal case: we expect registries
		if strings.Contains(reg, baseContainer) {
			_, err := util.ValidateImageString(reg)
			if err != nil {
				break
			}
			return reg
		}
	}
	return defaultImage
}
