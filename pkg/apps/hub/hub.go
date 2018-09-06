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
	"fmt"
	"os"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	Config     *rest.Config
	KubeClient *kubernetes.Clientset
	HubClient  *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset) (*Creater, error) {
	return &Creater{Config: config, KubeClient: kubeClient, HubClient: hubClient}, nil
}

// DeleteHub will delete the Black Duck Hub
func (hc *Creater) DeleteHub(namespace string) {
	var err error
	// Verify whether the namespace exist
	_, err = GetNamespace(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = DeleteNamespace(hc.KubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := GetNamespace(hc.KubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
		err = DeletePersistentVolume(hc.KubeClient, namespace)
		if err != nil {
			log.Errorf("unable to delete the pv for %+v", namespace)
		}
	}
}

// CreateHub will create the Black Duck Hub
func (hc *Creater) CreateHub(createHub *v1.Hub) (string, string, bool, error) {
	log.Debugf("Create Hub details for %s: %+v", createHub.Spec.Namespace, createHub)
	// Create a horizon deployer for each hub
	deployer, err := horizon.NewDeployer(hc.Config)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to create the horizon deployer due to %+v", err)
	}

	// Get Containers Flavor
	hubContainerFlavor := GetContainersFlavor(createHub.Spec.Flavor)
	log.Debugf("Hub Container Flavor: %+v", hubContainerFlavor)

	if hubContainerFlavor == nil {
		return "", "", true, fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) OpsSight, Actual: %s", createHub.Spec.Flavor)
	}

	// All ConfigMap environment variables
	allConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config-granular"},
	}

	if createHub.Spec.IsRandomPassword {
		createHub.Spec.AdminPassword, _ = RandomString(12)
		createHub.Spec.UserPassword, _ = RandomString(12)
	} else {
		createHub.Spec.AdminPassword = createHub.Spec.PostgresPassword
		createHub.Spec.UserPassword = createHub.Spec.PostgresPassword
	}
	log.Debugf("Before init: %+v", createHub)
	// Create the config-maps, secrets and postgres container
	err = hc.init(deployer, createHub, hubContainerFlavor, allConfigEnv)
	if err != nil {
		return "", "", true, err
	}
	// Deploy config-maps, secrets and postgres container
	err = deployer.Run()
	if err != nil {
		log.Errorf("deployments failed because %+v", err)
	}
	// time.Sleep(20 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err := GetAllPodsForNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Spec.Namespace, err)
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.KubeClient, pods)
	// Initialize the hub database
	if strings.EqualFold(createHub.Spec.DbPrototype, "empty") {
		InitDatabase(createHub)
	}

	// Create all hub deployments
	deployer, _ = horizon.NewDeployer(hc.Config)
	hc.createDeployer(deployer, createHub, hubContainerFlavor, allConfigEnv)
	log.Debugf("%+v", deployer)
	// Deploy all hub containers
	err = deployer.Run()
	if err != nil {
		log.Errorf("deployments failed because %+v", err)
		return "", "", true, fmt.Errorf("unable to deploy the hub in %s due to %+v", createHub.Spec.Namespace, err)
	}
	time.Sleep(10 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err = GetAllPodsForNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Spec.Namespace, err)
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.KubeClient, pods)

	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod := FilterPodByNamePrefix(pods, "registration")
	log.Debugf("registration pod: %+v", registrationPod)
	registrationKey := os.Getenv("REGISTRATION_KEY")
	log.Debugf("registration key: %s", registrationKey)

	if registrationPod != nil {
		for {
			// Create the exec into kubernetes pod request
			req := CreateExecContainerRequest(hc.KubeClient, registrationPod)
			// Exec into the kubernetes pod and execute the commands
			err = hc.execContainer(req, []string{fmt.Sprintf("curl -k -X POST https://127.0.0.1:8443/registration/HubRegistration?action=activate\\&registrationid=%s", registrationKey)})
			if err != nil {
				log.Infof("error in Stream: %v", err)
			} else {
				// Hub created and auto registered. Exit!!!!
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	// Retrieve the PVC volume name
	pvcVolumeName := ""
	if strings.EqualFold(createHub.Spec.BackupSupport, "Yes") || !strings.EqualFold(createHub.Spec.PVCStorageClass, "") {
		pvcVolumeName, err = hc.getPVCVolumeName(createHub.Spec.Namespace)
		if err != nil {
			return "", "", false, err
		}
	}

	ipAddress, err := hc.getLoadBalancerIPAddress(createHub.Spec.Namespace, "webserver-lb")
	if err != nil {
		return "", pvcVolumeName, false, err
	}
	log.Infof("hub Ip address: %s", ipAddress)
	return ipAddress, pvcVolumeName, false, nil
}

func (hc *Creater) getPVCVolumeName(namespace string) (string, error) {
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		pvc, err := GetPVC(hc.KubeClient, namespace, namespace)
		if err != nil {
			return "", fmt.Errorf("unable to get pvc in %s namespace due to %s", namespace, err.Error())
		}

		log.Debugf("pvc: %v", pvc)

		if strings.EqualFold(pvc.Spec.VolumeName, "") {
			continue
		} else {
			return pvc.Spec.VolumeName, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get pvc %s in %s namespace", namespace, namespace)
}

func (hc *Creater) getLoadBalancerIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		service, err := GetService(hc.KubeClient, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace due to %s", serviceName, namespace, err.Error())
		}

		log.Debugf("Service: %v", service)

		if len(service.Status.LoadBalancer.Ingress) > 0 {
			ipAddress := service.Status.LoadBalancer.Ingress[0].IP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}
