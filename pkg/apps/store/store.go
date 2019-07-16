package store

import (
	"fmt"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
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
		fmt.Println("loaded RC: " + name)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ServiceInterface:
		if ComponentStore.Service == nil {
			ComponentStore.Service = make(map[types.ComponentName]types.ServiceCreater)
		}
		ComponentStore.Service[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ServiceInterface)
		fmt.Println("loaded service: " + name)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ConfigMapInterface:
		if ComponentStore.Configmap == nil {
			ComponentStore.Configmap = make(map[types.ComponentName]types.ConfigmapCreater)
		}
		ComponentStore.Configmap[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.ConfigMapInterface)
		fmt.Println("loaded configmap: " + name)
	case types.PvcCreater:
		ComponentStore.PVC[name] = function.(types.PvcCreater)
	case func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.SecretInterface:
		if ComponentStore.Secret == nil {
			ComponentStore.Secret = make(map[types.ComponentName]types.SecretCreater)
		}
		ComponentStore.Secret[name] = function.(func(*protoform.Config, *kubernetes.Clientset, *v1.Blackduck) types.SecretInterface)
		fmt.Println("loaded secret: " + name)
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
				Image:  c,
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
		rcs = append(rcs, comp)
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
		secrets = append(secrets, c.GetSecrets()...)
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
		services = append(services, c.GetService())
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
		cms = append(cms, c.GetCM()...)
	}
	return cms, nil
}
