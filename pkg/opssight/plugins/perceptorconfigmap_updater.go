/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package plugins

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	opssiteclient "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/extensions"

	//extensions "github.com/kubernetes/kubernetes/pkg/apis/extensions"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type hubConfig struct {
	Hosts                     []string
	User                      string
	PasswordEnvVar            string
	ClientTimeoutMilliseconds int
	Port                      int
	ConcurrentScanLimit       int
	TotalScanLimit            int
}

type timings struct {
	CheckForStalledScansPauseHours int
	StalledScanClientTimeoutHours  int
	ModelMetricsPauseSeconds       int
	UnknownImagePauseMilliseconds  int
}

type perceptorConfig struct {
	Hub         *hubConfig
	Timings     *timings
	UseMockMode bool
	Port        int
	LogLevel    string
}

// PerceptorConfigMap ...
type PerceptorConfigMap struct{}

// sendHubs is one possible way to configure the perceptor hub family.
// TODO replace w/ configmap mutation if we want to.
func sendHubs(kubeClient *kubernetes.Clientset, namespace string, hubs []string) error {
	configmapList, err := kubeClient.Core().ConfigMaps(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var configMap *v1.ConfigMap
	for _, cm := range configmapList.Items {
		if cm.Name == "perceptor" {
			configMap = &cm
			break
		}
	}

	if configMap == nil {
		return fmt.Errorf("unable to find configmap perceptor-config")
	}

	var value perceptorConfig
	err = json.Unmarshal([]byte(configMap.Data["perceptor.yaml"]), &value)
	if err != nil {
		return err
	}

	value.Hub.Hosts = hubs

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	configMap.Data["perceptor_conf.yaml"] = string(jsonBytes)
	kubeClient.Core().ConfigMaps(namespace).Update(configMap)
	return nil
}

// Run is a BLOCKING function which should be run by the framework .
func (p *PerceptorConfigMap) Run(resources api.ControllerResources, ch chan struct{}) error {
	syncFunc := func() {
		updateAllHubs(resources.KubeClient)
	}
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return hubclient.New(resources.KubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return hubclient.New(resources.KubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&extensions.Deployment{},
		2*time.Second,
		cache.ResourceEventHandlerFuncs{
			// TODO kinda dumb, we just do a complete re-list of all hubs,
			// every time an event happens... But thats all we need to do, so its good enough.
			DeleteFunc: func(obj interface{}) {
				logrus.Infof("Hub deleted ! %v ", obj)
				syncFunc()
			},

			AddFunc: func(obj interface{}) {
				logrus.Infof("Hub added ! %v ", obj)
				syncFunc()
			},
		},
	)
	logrus.Infof("Starting controller for hub<->perceptor updates... this blocks, so running in a go func.")

	// make sure this is called from a go func.
	// This blocks!
	ctrl.Run(ch)
	return nil
}

// updateAllHubs will list all hubs in the cluster, and send them to opssight as scan targets.
// TODO there may be hubs which we dont want opssight to use.  Not sure how to deal with that yet.
func updateAllHubs(kubeClient *kubernetes.Clientset) error {
	allHubNamespaces := func() []string {
		allHubNamespaces := []string{}

		hubsList, _ := hubclient.New(kubeClient.RESTClient()).SynopsysV1().Hubs(v1.NamespaceAll).List(metav1.ListOptions{})
		hubs := hubsList.Items
		for _, hub := range hubs {
			ns := hub.Namespace
			allHubNamespaces = append(allHubNamespaces, ns)
			logrus.Infof("Hub config map controller, namespace is %v", ns)
		}
		return allHubNamespaces
	}()

	// for opssight 3.0, only support one opssight
	opssiteList, err := opssiteclient.New(kubeClient.RESTClient()).SynopsysV1().OpsSights(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// TODO, replace w/ configmap mutat ?
	// curl perceptor w/ the latest hub list
	for _, opssight := range opssiteList.Items {
		sendHubs(kubeClient, opssight.Namespace, allHubNamespaces)
	}
	return nil
}
