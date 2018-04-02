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
	"k8s.io/api/core/v1"

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
func NewRcSvc(descriptions []*PerceptorRC) (*v1.ReplicationController, []*v1.Service) {

	TheVolumes := []v1.Volume{}
	TheContainers := []v1.Container{}
	addedMounts := map[string]string{}

	for _, desc := range descriptions {
		mounts := []v1.VolumeMount{}

		for cfgMapName, cfgMapMount := range desc.configMapMounts {
			log.Print("Adding config mounts now.")
			if addedMounts[cfgMapName] == "" {
				addedMounts[cfgMapName] = cfgMapName
				TheVolumes = append(TheVolumes,
					v1.Volume{
						Name: cfgMapName,
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: cfgMapName,
								},
							},
						},
					})
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", cfgMapName))
			}
			mounts = append(mounts, v1.VolumeMount{
				Name:      cfgMapName,
				MountPath: cfgMapMount,
			})

		}

		// keep track of emptyDirs, only once, since it can be referenced in
		// multiple pods
		for emptyDirName, emptyDirMount := range desc.emptyDirMounts {
			log.Print("Adding empty mounts now.")
			if addedMounts[emptyDirName] == "" {
				addedMounts[emptyDirName] = emptyDirName
				TheVolumes = append(TheVolumes,
					v1.Volume{
						Name: emptyDirName,
						VolumeSource: v1.VolumeSource{
							EmptyDir: &v1.EmptyDirVolumeSource{},
						},
					})
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", emptyDirName))
			}
			mounts = append(mounts, v1.VolumeMount{
				Name:      emptyDirName,
				MountPath: emptyDirMount,
			})

		}

		if desc.dockerSocket {
			dockerSock := v1.VolumeMount{
				Name:      "dir-docker-socket",
				MountPath: "/var/run/docker.sock",
			}
			mounts = append(mounts, dockerSock)
			TheVolumes = append(TheVolumes, v1.Volume{
				Name: dockerSock.Name,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: dockerSock.MountPath,
					},
				},
			})
		}

		for name, _ := range desc.emptyDirVolumeMounts {
			TheVolumes = append(TheVolumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			})
		}

		envVar := []v1.EnvVar{}
		for _, env := range desc.env {
			envVar = append(envVar, v1.EnvVar{
				Name: env.EnvName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: env.SecretName,
						},
						Key: env.KeyFromSecret,
					},
				},
			})
		}

		container := v1.Container{
			Name:            desc.name,
			Image:           desc.image,
			ImagePullPolicy: "Always",
			Command:         desc.cmd,
			Env:             envVar,
			Ports: []v1.ContainerPort{
				v1.ContainerPort{
					ContainerPort: desc.port,
					Protocol:      "TCP",
				},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    desc.cpu,
					v1.ResourceMemory: desc.memory,
				},
			},
			VolumeMounts: mounts,
			SecurityContext: &v1.SecurityContext{
				Privileged: &desc.dockerSocket,
			},
		}
		// Each RC has only one pod, but can have many containers.
		TheContainers = append(TheContainers, container)

		log.Print(fmt.Sprintf("privileged = %v %v %v", desc.name, desc.dockerSocket, *container.SecurityContext.Privileged))
	}

	rc := &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{
			Name: descriptions[0].name,
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &descriptions[0].replicas,
			Selector: map[string]string{"name": descriptions[0].name},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{
					Labels: map[string]string{"name": descriptions[0].name},
				},
				Spec: v1.PodSpec{
					Volumes:            TheVolumes,
					Containers:         TheContainers,
					ServiceAccountName: descriptions[0].serviceAccountName,
				},
			},
		},
	}

	services := []*v1.Service{}
	for _, desc := range descriptions {
		services = append(services, &v1.Service{
			ObjectMeta: v1meta.ObjectMeta{
				Name: desc.name,
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					v1.ServicePort{
						Name: desc.name,
						Port: desc.port,
					},
				},
				Selector: map[string]string{
					// The POD name of the first container will be used as selector name for all containers
					"name": descriptions[0].name,
				},
			},
		})
	}
	return rc, services
}

// perceptor, pod-perceiver, image-perceiver, pod-perceiver

func CreatePerceptorResources(clientset *kubernetes.Clientset, paths map[string]string, pc *model.ProtoformConfig) []*v1.ReplicationController {

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
			emptyDirMounts: map[string]string{
				"logs": "/tmp",
			},
			name:   "perceptor",
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
			name:               "pod-perceiver",
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
				"var-images": "/var/images", "logs": "/tmp",
			},
			name:               "perceptor-scanner",
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
				"var-images": "/var/images", "logs": "/tmp",
			},
			name:               "perceptor-image-facade",
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

	// rcs := []*v1.ReplicationController{rcSCAN}
	// svc := [][]*v1.Service{svcSCAN}
	rcs := []*v1.ReplicationController{rcPCP, rcPCVR, rcSCAN} //rcPCVRo
	svc := [][]*v1.Service{svcPCP, svcPCVR, svcSCAN}          //svcPCVRo

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
				name:               "image-perceiver",
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

	// TODO MAKE SURE WE VERIFY THAT SERVICE ACCOUNTS ARE EQUAL

	for i, rc := range rcs {
		// Now, create all the resources.  Note that we'll panic after creating ANY
		// resource that fails.  Thats intentional.
		PrettyPrint(rc)
		if !pc.DryRun {
			_, err := clientset.Core().ReplicationControllers(pc.Namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
		for _, svcI := range svc[i] {
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
	for cn, _ := range svcAccounts {
		if !isValid(cn) {
			log.Print("[protoform] failed at verifiying that the container name for a svc account was valid!")
			log.Fatalln(cn)
			return false
		}
	}
	return true
}

func CreateConfigMapsFromInput(clientset *kubernetes.Clientset, pc *model.ProtoformConfig) {
	for _, configMap := range pc.ToConfigMap() {
		log.Println("*********************************************")
		log.Println("Creating config maps:", configMap)
		if !pc.DryRun {
			log.Println("creating config map.")
			clientset.Core().ConfigMaps(pc.Namespace).Create(configMap)
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
	//configPath := os.Args[1]
	runProtoform("/etc/protoform/")
}

func runProtoform(configPath string) []*v1.ReplicationController {
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
