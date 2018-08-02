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

package controller

import (
	"fmt"

	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Controller will store the controller configuration
type Controller struct {
	logger       *log.Entry
	clientset    kubernetes.Interface
	informer     cache.SharedIndexInformer
	hubclientset *hubclientset.Clientset
}

// Run will be executed to create the informers or controllers
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	c.logger.Info("Initiating controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		runtime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	c.logger.Info("Controller cache sync complete")
	<-stopCh
}

// HasSynced will check for informer sync
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// HubNamespaces will list the hub namespaces
func (c *Controller) HubNamespaces() ([]string, error) {
	// 1. get Hub CDR list from default ns
	hubList, err := c.hubclientset.SynopsysV1().Hubs(corev1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 2. extract the namespaces
	hubNamespaces := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 {
			hubNamespaces = append(hubNamespaces, hub.Spec.Namespace)
		}
	}
	return hubNamespaces, nil
}
