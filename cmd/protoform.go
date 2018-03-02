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
	viper.AddConfigPath(configPath)
	pc := &model.ProtoformConfig{}
	log.Print(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Print(" ^^ Didnt see a config file ! Making a reasonable default.")
		return nil
	}
	viper.Unmarshal(pc)
	PrettyPrint(pc)
	log.Print("*************** [protoform] done reading in viper ****************")
	return pc
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

type PerceptorRC struct {
	configMapMounts map[string]string
	emptyDirMounts  map[string]string
	name            string
	image           string
	port            int32
	cmd             []string

	// key:value = name:mountPath
	emptyDirVolumeMounts map[string]string

	// Only needed for openshift.
	serviceAccount     string
	serviceAccountName string
}

// This function creates an RC and services that forward to it.
func NewRcSvc(podName string, descriptions []PerceptorRC) (*v1.ReplicationController, []*v1.Service) {
	defaultMem, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}

	TheVolumes := []v1.Volume{}
	TheContainers := []v1.Container{}
	mounts := []v1.VolumeMount{}
	addedMounts := map[string]string{}

	for _, desc := range descriptions {

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
				mounts = append(mounts, v1.VolumeMount{
					Name:      cfgMapName,
					MountPath: cfgMapMount,
				})
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", cfgMapName))
			}
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
				mounts = append(mounts, v1.VolumeMount{
					Name:      emptyDirName,
					MountPath: emptyDirMount,
				})
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", emptyDirName))
			}
		}

		for name, _ := range desc.emptyDirVolumeMounts {
			TheVolumes = append(TheVolumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			})
		}

		// Each RC has only one pod, but can have many containers.
		TheContainers = append(TheContainers, v1.Container{
			Name:            desc.name,
			Image:           desc.image,
			ImagePullPolicy: "Always",
			Command:         desc.cmd,
			Ports: []v1.ContainerPort{
				v1.ContainerPort{
					ContainerPort: desc.port,
					Protocol:      "TCP",
				},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    defaultCPU,
					v1.ResourceMemory: defaultMem,
				},
			},
			VolumeMounts: mounts,
		})

	}
	rc := &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{
			Name: podName,
		},
		Spec: v1.ReplicationControllerSpec{
			Selector: map[string]string{"name": podName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{
					Labels: map[string]string{"name": podName},
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
					"name": desc.name,
				},
			},
		})
	}
	return rc, services
}

// perceptor, pod-perceiver, image-perceiver, pod-perceiver

func CreatePerceptorResources(namespace string, clientset *kubernetes.Clientset, svcAcct map[string]string, dryRun bool) {

	// perceptor = only one container, very simple.
	rcPCP, svcPCP := NewRcSvc("perceptor", []PerceptorRC{
		PerceptorRC{
			configMapMounts: map[string]string{"perceptor-config": "/etc/perceptor"},
			name:            "perceptor",
			image:           "gcr.io/gke-verification/blackducksoftware/perceptor:latest",
			port:            3001,
			cmd:             []string{"./perceptor"},
		},
	})

	// perceivers
	rcPCVR, svcPCVR := NewRcSvc("pod-perceiver", []PerceptorRC{
		PerceptorRC{
			configMapMounts:    map[string]string{"kube-generic-perceiver-config": "/etc/perceiver"},
			name:               "pod-perceiver",
			image:              "gcr.io/gke-verification/blackducksoftware/pod-perceiver:latest",
			port:               4000,
			cmd:                []string{},
			serviceAccountName: svcAcct["pod-perceiver"],
			serviceAccount:     svcAcct["pod-perceiver"],
		},
	})

	rcPCVRo, svcPCVRo := NewRcSvc("image-perceiver", []PerceptorRC{
		PerceptorRC{
			configMapMounts:    map[string]string{"openshift-perceiver-config": "/etc/perceiver"},
			name:               "image-perceiver",
			image:              "gcr.io/gke-verification/blackducksoftware/image-perceiver:latest",
			port:               4000,
			cmd:                []string{},
			serviceAccount:     svcAcct["image-perceiver"],
			serviceAccountName: svcAcct["image-perceiver"],
		},
	})

	perceptorScanner := model.NewPerceptorScannerPod(svcAcct["perceptor-image-facade"])

	replicationControllers := []*v1.ReplicationController{rcPCP, rcPCVR, rcPCVRo, perceptorScanner.ReplicationController()}
	services := [][]*v1.Service{svcPCP, svcPCVR, svcPCVRo, []*v1.Service{perceptorScanner.ScannerService(), perceptorScanner.ImageFacadeService()}}

	for i, rc := range replicationControllers {
		// Now, create all the resources.  Note that we'll panic after creating ANY
		// resource that fails.  Thats intentional.
		PrettyPrint(rc)
		if !dryRun {
			_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
		for _, svcI := range services[i] {
			if dryRun {
				// service dont really need much debug...
				//PrettyPrint(svc)
			} else {
				_, err := clientset.Core().Services(namespace).Create(svcI)
				if err != nil {
					panic(err)
				}
			}
		}
	}
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

func CreateConfigMapsFromInput(namespace string, clientset *kubernetes.Clientset, configMaps []*v1.ConfigMap, dryRun bool) {
	for _, configMap := range configMaps {
		log.Println("*********************************************")
		log.Println("Creating config maps:", configMap)
		if !dryRun {
			log.Println("creating config map.")
			clientset.Core().ConfigMaps(namespace).Create(configMap)
		} else {
			PrettyPrint(configMap)
		}
	}
}

// protoform is an experimental installer which bootstraps perceptor and the other
// autobots.

// main installs prime
func main() {
	//configPath := os.Args[1]
	runProtoform("/etc/protoform/")
}

func runProtoform(configPath string) {
	namespace := "bds-perceptor"
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
			"pod-perceiver":          "openshift-perceiver",
			"image-perceiver":        "openshift-perceiver",
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

	CreateConfigMapsFromInput(namespace, clientset, pc.ToConfigMap(), pc.DryRun)
	CreatePerceptorResources(namespace, clientset, pc.ServiceAccounts, pc.DryRun)
}
