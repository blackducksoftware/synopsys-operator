package store

import (
	"fmt"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
	"strings"

	//_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"log"
)

type Components struct {
	Rc        map[types.ComponentName]types.ReplicationControllerCreater
	Service   map[types.ComponentName]types.ServiceCreater
	Configmap map[types.ComponentName]types.ConfigmapCreater
	PVC       map[types.ComponentName]types.PvcCreater
	Secret    map[types.ComponentName]types.SecretCreater
	Size      map[types.ComponentName]types.SizeInterface
}

var ComponentStore Components

func Register(name types.ComponentName, function interface{}) {
	switch function.(type) {
	case func(*types.ReplicationController, *protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ReplicationControllerInterface:
		if ComponentStore.Rc == nil {
			ComponentStore.Rc = make(map[types.ComponentName]types.ReplicationControllerCreater)
		}
		ComponentStore.Rc[name] = function.(func(*types.ReplicationController, *protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ReplicationControllerInterface)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ServiceInterface:
		if ComponentStore.Service == nil {
			ComponentStore.Service = make(map[types.ComponentName]types.ServiceCreater)
		}
		ComponentStore.Service[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ServiceInterface)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ConfigMapInterface:
		if ComponentStore.Configmap == nil {
			ComponentStore.Configmap = make(map[types.ComponentName]types.ConfigmapCreater)
		}
		ComponentStore.Configmap[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ConfigMapInterface)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.PVCInterface:
		if ComponentStore.PVC == nil {
			ComponentStore.PVC = make(map[types.ComponentName]types.PvcCreater)
		}
		ComponentStore.PVC[name] = function.(func(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *v1.Blackduck) types.PVCInterface)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.SecretInterface:
		if ComponentStore.Secret == nil {
			ComponentStore.Secret = make(map[types.ComponentName]types.SecretCreater)
		}
		ComponentStore.Secret[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.SecretInterface)
	case types.SizeInterface:
		if ComponentStore.Size == nil {
			ComponentStore.Size = make(map[types.ComponentName]types.SizeInterface)
		}
		ComponentStore.Size[name] = function.(types.SizeInterface)
	default:
		log.Fatal("Couldn't load " + name)
	}
}

func GetComponents(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) (api.ComponentList, error) {
	var cp api.ComponentList

	// Rc
	rcs, err := generateRc(v, config, kubeclient, blackduck)
	if err != nil {
		return cp, err
	}
	cp.ReplicationControllers = rcs

	// Services
	services, err := generateService(v, config, kubeclient, blackduck)
	if err != nil {
		return cp, err
	}
	cp.Services = services

	// Configmap
	cm, err := generateConfigmap(v, config, kubeclient, blackduck)
	if err != nil {
		return cp, err
	}
	cp.ConfigMaps = cm

	// Secret
	secrets, err := generateSecret(v, config, kubeclient, blackduck)
	if err != nil {
		return cp, err
	}
	cp.Secrets = secrets

	// PVC
	pvcs, err := generatePVCs(v, config, kubeclient, blackduck)
	if err != nil {
		return cp, err
	}
	cp.PersistentVolumeClaims = pvcs

	return cp, nil
}

func generateRc(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) ([]*components.ReplicationController, error) {
	var rcs []*components.ReplicationController
	// Size
	sizegen, okk := ComponentStore.Size[v.Size]
	if !okk {
		return nil, fmt.Errorf("size %s couldn't be found", v.Size)
	}

	size := sizegen.GetSize(blackduck.Spec.Size)
	if size == nil {
		return nil, fmt.Errorf("size %s couldn't be found in %s", blackduck.Spec.Size, v.Size)
	}

	// RC
	for k, v := range v.RCs {
		rcSize, ok := size[k]
		if !ok {
			return nil, fmt.Errorf("replication controller %s couldn't be found in size [%s]", k, blackduck.Spec.Size)
		}
		rc := &types.ReplicationController{
			Namespace:  blackduck.Spec.Namespace,
			Replicas:   rcSize.Replica,
			Containers: map[types.ContainerName]types.Container{},
		}

		for cName, c := range v.Container {
			rc.Containers[cName] = types.Container{
				Image:  generateImageTag(c, blackduck),
				MinMem: size[k].Containers[cName].MinMem,
				MaxMem: size[k].Containers[cName].MaxMem,
				MinCPU: size[k].Containers[cName].MinCPU,
				MaxCPU: size[k].Containers[cName].MaxCPU,
			}
		}

		component, ok := ComponentStore.Rc[v.Identifier]
		if !ok {
			return nil, fmt.Errorf("rc %s couldn't be found", v.Identifier)
		}

		test := component(rc, config, kubeclient, blackduck)
		comp, err := test.GetRc()
		if err != nil {
			return nil, err
		}
		if comp != nil {
			rcs = append(rcs, comp)
		}
	}
	return rcs, nil
}

func generateSecret(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) ([]*components.Secret, error) {
	var secrets []*components.Secret
	for _, v := range v.Secrets {
		component, ok := ComponentStore.Secret[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find secret %s", v)
		}
		c := component(config, kubeclient, blackduck)
		res := c.GetSecrets()
		if len(res) > 0 {
			secrets = append(secrets, res...)
		}
	}
	return secrets, nil
}

func generateService(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) ([]*components.Service, error) {
	var services []*components.Service
	for _, v := range v.Services {
		component, ok := ComponentStore.Service[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find secret %s", v)
		}
		c := component(config, kubeclient, blackduck)
		res := c.GetService()
		if res != nil {
			services = append(services, res)
		}
	}
	return services, nil
}

func generateConfigmap(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) ([]*components.ConfigMap, error) {
	var cms []*components.ConfigMap
	for _, v := range v.ConfigMaps {
		component, ok := ComponentStore.Configmap[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find secret %s", v)
		}
		c := component(config, kubeclient, blackduck)
		res := c.GetCM()
		if len(res) > 0 {
			cms = append(cms, res...)
		}
	}
	return cms, nil
}

func generatePVCs(v types.PublicVersion, config *protoform.Config, kubeclient *kubernetes.Clientset, blackduck *v1.Blackduck) ([]*components.PersistentVolumeClaim, error) {
	var pvcs []*components.PersistentVolumeClaim
	for _, v := range v.PVC {
		component, ok := ComponentStore.PVC[v]
		if !ok {
			return nil, fmt.Errorf("couldn't find secret %s", v)
		}
		c := component(config, kubeclient, blackduck)
		res, err := c.GetPVCs()
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			pvcs = append(pvcs, res...)
		}
	}
	return pvcs, nil
}

func generateImageTag(currentImage string, blackduck *v1.Blackduck) string {
	if len(blackduck.Spec.ImageRegistries) > 0 {
		imageName, err := util.GetImageName(currentImage)
		if err != nil {
			return currentImage
		}
		return getFullContainerNameFromImageRegistryConf(imageName, blackduck.Spec.ImageRegistries)
	}

	if len(blackduck.Spec.RegistryConfiguration.Registry) > 0 && len(blackduck.Spec.RegistryConfiguration.Namespace) > 0 {
		imageName, err := util.GetImageName(currentImage)
		if err != nil {
			return currentImage
		}
		imageTag, err := util.GetImageTag(currentImage)
		if err != nil {
			return currentImage
		}
		return fmt.Sprintf("%s/%s/%s:%s", blackduck.Spec.RegistryConfiguration.Registry, blackduck.Spec.RegistryConfiguration.Namespace, imageName, imageTag)
	}

	return currentImage
}

func getFullContainerNameFromImageRegistryConf(baseContainer string, images []string) string {
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
	return baseContainer
}
