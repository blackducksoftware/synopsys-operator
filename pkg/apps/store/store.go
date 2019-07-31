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
	sizev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/size/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/size"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	sizeclientset "github.com/blackducksoftware/synopsys-operator/pkg/size/client/clientset/versioned"
)

// Components contains the list of components to be added/updated
type Components struct {
	Rc        map[types.ComponentName]types.ReplicationControllerCreater
	Service   map[types.ComponentName]types.ServiceCreater
	Configmap map[types.ComponentName]types.ConfigmapCreater
	PVC       map[types.ComponentName]types.PvcCreater
	Secret    map[types.ComponentName]types.SecretCreater
}

// ComponentStore stores the components
var ComponentStore Components

// Register registers the components to the store
func Register(name types.ComponentName, function interface{}) {
	switch function.(type) {
	case func(*types.ReplicationController, *protoform.Config, *kubernetes.Clientset, interface{}) (types.ReplicationControllerInterface, error):
		if ComponentStore.Rc == nil {
			ComponentStore.Rc = make(map[types.ComponentName]types.ReplicationControllerCreater)
		}
		ComponentStore.Rc[name] = function.(func(*types.ReplicationController, *protoform.Config, *kubernetes.Clientset, interface{}) (types.ReplicationControllerInterface, error))
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
func GetComponents(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, sizeClient *sizeclientset.Clientset, cr interface{}) (api.ComponentList, error) {
	var cp api.ComponentList

	// get the custom resource info
	customResource, err := getCR(cr)
	if err != nil {
		return api.ComponentList{}, fmt.Errorf("unable to get the components because %+v", err)
	}

	// Rc
	rcs, err := generateRc(v, config, kubeclient, sizeClient, customResource, cr)
	if err != nil {
		return cp, err
	}
	cp.ReplicationControllers = rcs

	// Services
	services, err := generateService(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.Services = services

	// Configmap
	cm, err := generateConfigmap(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.ConfigMaps = cm

	// Secret
	secrets, err := generateSecret(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.Secrets = secrets

	// PVC
	pvcs, err := generatePVCs(v, config, kubeclient, cr)
	if err != nil {
		return cp, err
	}
	cp.PersistentVolumeClaims = pvcs

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
	for k, v := range v.RCs {
		rcSize, rcFound := componentSize.Spec.Rc[k]

		rc := &types.ReplicationController{
			Namespace:  customResource.namespace,
			Replicas:   1,
			Containers: map[types.ContainerName]types.Container{},
		}

		if rcFound {
			rc.Replicas = rcSize.Replica
		}

		for containerName, defaultImage := range v.Container {
			tmpContainer := types.Container{
				Image: generateImageTag(defaultImage, customResource.imageRegistries, customResource.registryConfiguration),
			}
			if rcFound {
				if containerSize, containerSizeFound := rcSize.ContainerLimit[string(containerName)]; containerSizeFound {
					tmpContainer.MinMem = containerSize.MinMem
					tmpContainer.MaxMem = containerSize.MaxMem
					tmpContainer.MinCPU = containerSize.MinCPU
					tmpContainer.MaxCPU = containerSize.MaxCPU
				}
			}
			rc.Containers[containerName] = tmpContainer
		}

		component, ok := ComponentStore.Rc[v.Identifier]
		if !ok {
			return nil, fmt.Errorf("rc %s couldn't be found", v.Identifier)
		}

		rcCreater, err := component(rc, config, kubeclient, cr)
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
		res := secretCreater.GetSecrets()
		if len(res) > 0 {
			secrets = append(secrets, res...)
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
		res := serviceCreater.GetService()
		if res != nil {
			services = append(services, res)
		}
	}
	return services, nil
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
		res := cmCreater.GetCM()
		if len(res) > 0 {
			cms = append(cms, res...)
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
