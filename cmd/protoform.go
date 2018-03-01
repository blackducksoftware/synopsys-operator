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
	"log"

	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var pc *model.ProtoformConfig

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func init() {
	log.Print("*************** [protoform] initializing viper ****************")
	viper.SetConfigName("protoform")
	viper.AddConfigPath("/etc/protoform/")
	pc := &model.ProtoformConfig{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(pc)
	PrettyPrint(pc)
	log.Print("*************** [protoform] done reading in viper ****************")
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

	// if true, then container is privileged /var/run/docker.sock.
	dockerSocket bool

	// Only needed for openshift.
	serviceAccount     string
	serviceAccountName string
}

// This function creates an RC and services that forward to it.
func NewRcSvc(descriptions []PerceptorRC) (*v1.ReplicationController, []*v1.Service) {
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

	for _, desc := range descriptions {
		for cfgMapName, cfgMapMount := range desc.configMapMounts {
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
		}
		for emptyDirName, emptyDirMount := range desc.emptyDirMounts {
			TheVolumes = append(TheVolumes,
				v1.Volume{
					Name: emptyDirName,
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: emptyDirName,
							},
						},
					},
				})
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
			SecurityContext: &v1.SecurityContext{
				Privileged: &desc.dockerSocket,
			},
		})

	}
	rc := &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{
			Name: descriptions[0].name,
		},
		Spec: v1.ReplicationControllerSpec{
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
	rcPCP, svcPCP := NewRcSvc([]PerceptorRC{
		PerceptorRC{
			configMapMounts: map[string]string{"perceptor-config": "/etc/perceptor"},
			name:            "perceptor",
			image:           "gcr.io/gke-verification/blackducksoftware/perceptor:latest",
			port:            3001,
			cmd:             []string{"./perceptor"},
		},
	})

	// perceivers
	rcPCVR, svcPCVR := NewRcSvc([]PerceptorRC{
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

	rcPCVRo, svcPCVRo := NewRcSvc([]PerceptorRC{
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

	rcSCAN, svcSCAN := NewRcSvc([]PerceptorRC{
		PerceptorRC{
			configMapMounts: map[string]string{"perceptor-scanner-config": "/etc/perceptor_scanner"},
			emptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			name:         "image-perceiver",
			image:        "gcr.io/gke-verification/blackducksoftware/perceptor-imagefacade:latest",
			dockerSocket: false,
			port:         4000,
			cmd:          []string{},
		},
		PerceptorRC{
			configMapMounts: map[string]string{"perceptor-scanner-config": "/etc/perceptor_scanner"},
			emptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			name:               "perceptor-scanner",
			image:              "gcr.io/gke-verification/blackducksoftware/perceptor-scanner:latest",
			dockerSocket:       true,
			port:               3003,
			cmd:                []string{},
			serviceAccount:     svcAcct["perceptor-scanner"],
			serviceAccountName: svcAcct["perceptor-scanner"],
		},
	})

	rcs := []*v1.ReplicationController{rcPCP, rcPCVR, rcPCVRo, rcSCAN}
	svc := [][]*v1.Service{svcPCP, svcPCVR, svcPCVRo, svcSCAN}

	for i, rc := range rcs {
		// Now, create all the resources.  Note that we'll panic after creating ANY
		// resource that fails.  Thats intentional.
		PrettyPrint(rc)
		if !dryRun {
			_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
		for _, svcI := range svc[i] {
			if dryRun {
				PrettyPrint(svc)
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
		for _, valid := range []string{"perceptor", "pod-perceiver", "image-perceiver", "perceptor-scanner"} {
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

func CreateConfigMapsFromInput(namespace string, clientset *kubernetes.Clientset, dryRun bool) {
	configMaps := pc.ToConfigMap()

	for _, configMap := range configMaps {
		clientset.Core().ConfigMaps(namespace).Create(configMap)
	}

}

// protoform is an experimental installer which bootstraps perceptor and the other
// autobots.

// main installs prime
func main() {

	namespace := "bds-perceptor"

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// TODO Viperize these env vars.
	if pc.ServiceAccounts == nil {
		log.Println("[viper] NO SERVICE ACCOUNTS FOUND.  USING DEFAULTS: MAKE SURE THESE EXIST!")

		svcAccounts := map[string]string{
			// WARNINNG: These service accounts need to exist !
			"pod-perceiver":     "openshift-perceiver",
			"image-perceiver":   "openshift-perceiver",
			"perceptor-scanner": "perceptor-scanner-sa",
		}
		// TODO programmatically validate rather then sanity check.
		PrettyPrint(svcAccounts)
		pc.ServiceAccounts = svcAccounts
	}

	isValid := sanityCheckServices(pc.ServiceAccounts)
	if isValid == false {
		panic("Please set the service accounts correctly!")
	}

	dryRun := pc.DryRun
	CreateConfigMapsFromInput(namespace, clientset, dryRun)
	CreatePerceptorResources(namespace, clientset, pc.ServiceAccounts, dryRun)
}
