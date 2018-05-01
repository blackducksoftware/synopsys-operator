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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/spf13/viper"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/api/resource"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func readConfig(configPath string) *model.ProtoformConfig {
	log.Print("*************** [protoform] initializing viper ****************")
	viper.SetConfigName("protoform")

	// these need to be set before we read in the config!
	viper.SetEnvPrefix("PCP")
	viper.BindEnv("HubUserPassword")
	if viper.GetString("hubuserpassword") == "" {
		viper.Debug()
		panic("No hub database password secret supplied.  Please inject PCP_HUBUSERPASSWORD as a secret and restart!")
	}

	viper.AddConfigPath(configPath)

	pc := &model.ProtoformConfig{}
	pc.HubUserPasswordEnvVar = "PCP_HUBUSERPASSWORD"
	pc.ViperSecret = "viper-secret"
	log.Print(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Print(" ^^ Didnt see a config file ! Making a reasonable default.")
		return nil
	}

	internalRegistry := viper.GetStringSlice("InternalDockerRegistries")
	viper.Set("InternalDockerRegistries", internalRegistry)

	viper.Unmarshal(pc)
	PrettyPrint(pc)
	log.Print("*************** [protoform] done reading in viper ****************")
	return pc
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

type EnvSecret struct {
	EnvName       string
	SecretName    string
	KeyFromSecret string
}

type PerceptorRC struct {
	configMapMounts map[string]string
	emptyDirMounts  map[string]string
	name            string
	image           string
	port            int32
	cmd             []string
	replicas        int32
	env             []EnvSecret

	// key:value = name:mountPath
	emptyDirVolumeMounts map[string]string

	// if true, then container is privileged /var/run/docker.sock.
	dockerSocket bool

	// Only needed for openshift.
	serviceAccount     string
	serviceAccountName string

	memory resource.Quantity
	cpu    resource.Quantity
}

// This function creates an RC and services that forward to it.
func NewRcSvc(descriptions []*PerceptorRC) (*types.ReplicationController, []*types.Service) {

	TheVolumes := map[string]types.Volume{}
	TheContainers := []types.Container{}
	addedMounts := map[string]string{}

	for _, desc := range descriptions {
		mounts := []types.VolumeMount{}

		for cfgMapName, cfgMapMount := range desc.configMapMounts {
			log.Print("Adding config mounts now.")
			if addedMounts[cfgMapName] == "" {
				addedMounts[cfgMapName] = cfgMapName
				TheVolumes[cfgMapName] = types.Volume{
					ConfigMap: &types.ConfigMapVolume{
						Name: cfgMapName,
					},
				}
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", cfgMapName))
			}
			mounts = append(mounts, types.VolumeMount{
				Store:     cfgMapName,
				MountPath: cfgMapMount,
			})

		}

		// keep track of emptyDirs, only once, since it can be referenced in
		// multiple pods
		for emptyDirName, emptyDirMount := range desc.emptyDirMounts {
			log.Print("Adding empty mounts now.")
			if addedMounts[emptyDirName] == "" {
				addedMounts[emptyDirName] = emptyDirName
				TheVolumes[emptyDirName] = types.Volume{
					EmptyDir: &types.EmptyDirVolume{
						Medium: types.StorageMediumDefault,
					},
				}
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", emptyDirName))
			}
			mounts = append(mounts, types.VolumeMount{
				Store:     emptyDirName,
				MountPath: emptyDirMount,
			})

		}

		if desc.dockerSocket {
			dockerSock := types.VolumeMount{
				Store:     "dir-docker-socket",
				MountPath: "/var/run/docker.sock",
			}
			mounts = append(mounts, dockerSock)
			TheVolumes[dockerSock.Store] = types.Volume{
				HostPath: &types.HostPathVolume{
					Path: dockerSock.MountPath,
				},
			}
		}

		for name := range desc.emptyDirVolumeMounts {
			TheVolumes[name] = types.Volume{
				EmptyDir: &types.EmptyDirVolume{
					Medium: types.StorageMediumDefault,
				},
			}
		}

		envVar := []types.Env{}
		for _, env := range desc.env {
			new, err := types.NewEnvFromSecret(env.EnvName, env.SecretName, env.KeyFromSecret)
			if err != nil {
				panic(err)
			}
			envVar = append(envVar, new)
		}

		container := types.Container{
			Name:    desc.name,
			Image:   desc.image,
			Pull:    types.PullAlways,
			Command: desc.cmd,
			Env:     envVar,
			Expose: []types.Port{
				{
					ContainerPort: fmt.Sprintf("%d", desc.port),
					Protocol:      types.ProtocolTCP,
				},
			},
			CPU: &types.CPU{
				Min: desc.cpu.String(),
			},
			Mem: &types.Mem{
				Min: desc.memory.String(),
			},
			VolumeMounts: mounts,
			Privileged:   &desc.dockerSocket,
		}
		// Each RC has only one pod, but can have many containers.
		TheContainers = append(TheContainers, container)

		log.Print(fmt.Sprintf("privileged = %v %v %v", desc.name, desc.dockerSocket, *container.Privileged))
	}

	rc := &types.ReplicationController{
		Name:     descriptions[0].name,
		Replicas: &descriptions[0].replicas,
		Selector: map[string]string{"name": descriptions[0].name},
		TemplateMetadata: &types.PodTemplateMeta{
			Labels: map[string]string{"name": descriptions[0].name},
		},
		PodTemplate: types.PodTemplate{
			Volumes:    TheVolumes,
			Containers: TheContainers,
			Account:    descriptions[0].serviceAccountName,
		},
	}

	services := []*types.Service{}
	for _, desc := range descriptions {
		services = append(services, &types.Service{
			Name: desc.name,
			Ports: []types.NamedServicePort{
				{
					Name: desc.name,
					Port: types.ServicePort{
						Expose: desc.port,
					},
				},
			},
			Selector: map[string]string{
				// The POD name of the first container will be used as selector name for all containers
				"name": descriptions[0].name,
			},
		})
	}
	return rc, services
}

// perceptor, pod-perceiver, image-perceiver, pod-perceiver

func CreatePerceptorResources(clientset *kubernetes.Clientset, paths map[string]string, pc *model.ProtoformConfig) []*types.ReplicationController {

	// WARNING: THE SERVICE ACCOUNT IN THE FIRST CONTAINER IS USED FOR THE GLOBAL SVC ACCOUNT FOR ALL PODS !!!!!!!!!!!!!
	// MAKE SURE IF YOU NEED A SVC ACCOUNT THAT ITS IN THE FIRST CONTAINER...
	defaultMem, err := resource.ParseQuantity(pc.DefaultMem)
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity(pc.DefaultCPU)
	if err != nil {
		panic(err)
	}

	rcPCP, svcPCP := NewRcSvc([]*PerceptorRC{
		&PerceptorRC{
			replicas:        1,
			configMapMounts: map[string]string{"perceptor-config": "/etc/perceptor"},
			env: []EnvSecret{
				{
					EnvName:       pc.HubUserPasswordEnvVar,
					SecretName:    pc.ViperSecret,
					KeyFromSecret: "HubUserPassword",
				},
			},
			name:   pc.PerceptorImageName,
			image:  paths["perceptor"],
			port:   int32(pc.PerceptorPort),
			cmd:    []string{"./perceptor"},
			cpu:    defaultCPU,
			memory: defaultMem,
		},
	})

	// perceivers
	rcPCVR, svcPCVR := NewRcSvc([]*PerceptorRC{
		&PerceptorRC{
			replicas:        1,
			configMapMounts: map[string]string{"perceiver": "/etc/perceiver"},
			emptyDirMounts: map[string]string{
				"logs": "/tmp",
			},
			name:               pc.PodPerceiverImageName,
			image:              paths["pod-perceiver"],
			port:               int32(pc.PerceiverPort),
			cmd:                []string{},
			serviceAccountName: pc.ServiceAccounts["pod-perceiver"],
			serviceAccount:     pc.ServiceAccounts["pod-perceiver"],
			cpu:                defaultCPU,
			memory:             defaultMem,
		},
	})

	rcSCAN, svcSCAN := NewRcSvc([]*PerceptorRC{
		&PerceptorRC{
			replicas:        int32(math.Ceil(float64(pc.ConcurrentScanLimit) / 2.0)),
			configMapMounts: map[string]string{"perceptor-scanner-config": "/etc/perceptor_scanner"},
			env: []EnvSecret{
				{
					EnvName:       pc.HubUserPasswordEnvVar,
					SecretName:    pc.ViperSecret,
					KeyFromSecret: "HubUserPassword",
				},
			},
			emptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			name:               pc.ScannerImageName,
			image:              paths["perceptor-scanner"],
			dockerSocket:       false,
			port:               int32(pc.ScannerPort),
			cmd:                []string{},
			serviceAccount:     pc.ServiceAccounts["perceptor-image-facade"],
			serviceAccountName: pc.ServiceAccounts["perceptor-image-facade"],
			cpu:                defaultCPU,
			memory:             defaultMem,
		},
		&PerceptorRC{
			configMapMounts: map[string]string{"perceptor-imagefacade-config": "/etc/perceptor_imagefacade"},
			emptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			name:               pc.ImageFacadeImageName,
			image:              paths["perceptor-imagefacade"],
			dockerSocket:       true,
			port:               int32(pc.ImageFacadePort),
			cmd:                []string{},
			serviceAccount:     pc.ServiceAccounts["perceptor-image-facade"],
			serviceAccountName: pc.ServiceAccounts["perceptor-image-facade"],
			cpu:                defaultCPU,
			memory:             defaultMem,
		},
	})

	rcs := []*types.ReplicationController{rcPCP, rcPCVR, rcSCAN} //rcPCVRo
	svc := [][]*types.Service{svcPCP, svcPCVR, svcSCAN}          //svcPCVRo

	// We dont create openshift perceivers if running kube... This needs to be avoided b/c the svc accounts
	// won't exist.
	if pc.Openshift {
		rcOpenshift, svcOpenshift := NewRcSvc([]*PerceptorRC{
			&PerceptorRC{
				replicas:        1,
				configMapMounts: map[string]string{"perceiver": "/etc/perceiver"},
				emptyDirMounts: map[string]string{
					"logs": "/tmp",
				},
				name:               pc.ImagePerceiverImageName,
				image:              paths["image-perceiver"],
				port:               int32(pc.PerceiverPort),
				cmd:                []string{},
				serviceAccount:     pc.ServiceAccounts["image-perceiver"],
				serviceAccountName: pc.ServiceAccounts["image-perceiver"],
			},
		})
		rcs = append(rcs, rcOpenshift)
		svc = append(svc, svcOpenshift)
	}

	if pc.PerceptorSkyfire {
		rcSkyfire, svcSkyfire := NewRcSvc([]*PerceptorRC{
			&PerceptorRC{
				replicas:        1,
				configMapMounts: map[string]string{"skyfire": "/etc/skyfire"},
				emptyDirMounts: map[string]string{
					"logs": "/tmp",
				},
				env: []EnvSecret{
					{
						EnvName:       pc.HubUserPasswordEnvVar,
						SecretName:    pc.ViperSecret,
						KeyFromSecret: "HubUserPassword",
					},
				},
				name:               "skyfire",
				image:              "gcr.io/blackducksoftware/skyfire-daemon:master",
				port:               3005,
				cmd:                []string{},
				serviceAccount:     pc.ServiceAccounts["image-perceiver"],
				serviceAccountName: pc.ServiceAccounts["image-perceiver"],
			},
		})
		rcs = append(rcs, rcSkyfire)
		svc = append(svc, svcSkyfire)
	}

	// TODO MAKE SURE WE VERIFY THAT SERVICE ACCOUNTS ARE EQUAL

	for i, krc := range rcs {
		// Now, create all the resources.  Note that we'll panic after creating ANY
		// resource that fails.  Thats intentional.
		wrapper := &types.ReplicationControllerWrapper{ReplicationController: *krc}
		rc, err := converters.Convert_Koki_ReplicationController_to_Kube_v1_ReplicationController(wrapper)
		if err != nil {
			panic(err)
		}
		PrettyPrint(rc)
		if !pc.DryRun {
			_, err := clientset.Core().ReplicationControllers(pc.Namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
		for _, ksvcI := range svc[i] {
			sWrapper := &types.ServiceWrapper{Service: *ksvcI}
			svcI, err := converters.Convert_Koki_Service_To_Kube_v1_Service(sWrapper)
			if err != nil {
				panic(err)
			}
			if pc.DryRun {
				// service dont really need much debug...
				// PrettyPrint(svcI)
			} else {
				_, err := clientset.Core().Services(pc.Namespace).Create(svcI)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return rcs
}

func sanityCheckServices(svcAccounts map[string]string) bool {
	isValid := func(cn string) bool {
		for _, valid := range []string{"perceptor", "pod-perceiver", "image-perceiver", "perceptor-scanner", "perceptor-image-facade"} {
			if cn == valid {
				return true
			}
		}
		return false
	}
	for cn := range svcAccounts {
		if !isValid(cn) {
			log.Print("[protoform] failed at verifiying that the container name for a svc account was valid!")
			log.Fatalln(cn)
			return false
		}
	}
	return true
}

func CreateConfigMapsFromInput(clientset *kubernetes.Clientset, pc *model.ProtoformConfig) {
	for _, kconfigMap := range pc.ToConfigMap() {
		wrapper := &types.ConfigMapWrapper{ConfigMap: *kconfigMap}
		configMap, err := converters.Convert_Koki_ConfigMap_to_Kube_v1_ConfigMap(wrapper)
		if err != nil {
			panic(err)
		}
		log.Println("*********************************************")
		log.Println("Creating config maps:", configMap)
		if !pc.DryRun {
			log.Println("creating config map.")
			_, err := clientset.Core().ConfigMaps(pc.Namespace).Create(configMap)
			if err != nil {
				panic(err)
			}
		} else {
			PrettyPrint(configMap)
		}
	}
}

// GenerateContainerPaths creates paths with reasonable defaults.
func GenerateContainerPaths(config *model.ProtoformConfig) map[string]string {
	return map[string]string{
		"perceptor":             fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.PerceptorImageName, config.PerceptorContainerVersion),
		"perceptor-scanner":     fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ScannerImageName, config.ScannerContainerVersion),
		"pod-perceiver":         fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.PodPerceiverImageName, config.PerceiverContainerVersion),
		"image-perceiver":       fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ImagePerceiverImageName, config.PerceiverContainerVersion),
		"perceptor-imagefacade": fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ImageFacadeImageName, config.ImageFacadeContainerVersion),
	}
}

// protoform is an experimental installer which bootstraps perceptor and the other
// autobots.

// main installs prime
func main() {
	runProtoform("/etc/protoform/")
}

func runProtoform(configPath string) []*types.ReplicationController {
	var clientset *kubernetes.Clientset
	pc := readConfig(configPath)
	if pc == nil {

	}
	if !pc.DryRun {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		} else {
			// creates the clientset
			clientset, err = kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	// TODO Viperize these env vars.
	if pc.ServiceAccounts == nil {
		log.Println("[viper] NO SERVICE ACCOUNTS FOUND.  USING DEFAULTS: MAKE SURE THESE EXIST!")

		svcAccounts := map[string]string{
			// WARNINNG: These service accounts need to exist !
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner-sa",
		}
		// TODO programatically validate rather then sanity check.
		PrettyPrint(svcAccounts)
		pc.ServiceAccounts = svcAccounts
	}

	isValid := sanityCheckServices(pc.ServiceAccounts)
	if isValid == false {
		panic("Please set the service accounts correctly!")
	}

	log.Println("Creating config maps : Dry Run ")

	CreateConfigMapsFromInput(clientset, pc)
	imagePaths := GenerateContainerPaths(pc)
	rcsCreated := CreatePerceptorResources(clientset, imagePaths, pc)

	log.Println("Entering pod listing loop!")

	// continually print out pod statuses .  can exit any time.  maybe use this as a stub for self testing.
	if !pc.DryRun {
		for i := 0; i < 10; i++ {
			pods, _ := clientset.Core().Pods(pc.Namespace).List(v1meta.ListOptions{})
			for _, pod := range pods.Items {
				log.Printf("Pod = %v -> %v", pod.Name, pod.Status.Phase)
			}
			log.Printf("***************")
			time.Sleep(10 * time.Second)
		}
	}

	return rcsCreated
}
