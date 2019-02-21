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

package sample

import (
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
	sampleclientset "github.com/blackducksoftware/synopsys-operator/pkg/sample/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// The Creater type is used to Create and Delete a Sample from your cluster
type Creater struct {
	kubeConfig   *rest.Config
	kubeClient   *kubernetes.Clientset
	sampleClient *sampleclientset.Clientset
}

// NewSampleCreater will instantiate the Sample's Creater
func NewSampleCreater(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, sampleClient *sampleclientset.Clientset) *Creater {
	return &Creater{kubeConfig: kubeConfig, kubeClient: kubeClient, sampleClient: sampleClient}
}

// CreateSample will deploy a Sample into the Cluster
func (sampleCreater *Creater) CreateSample(sampleObj *v1.SampleSpec) error {
	log.Debugf("Creating a Sample in namespace %s: %+v", sampleObj.Namespace, sampleObj)
	// Get a ComponentList for the Sample
	sample := NewSample(sampleObj)
	componentList, err := sample.GetComponents(sampleObj)
	if err != nil {
		log.Errorf("Unable to get components for Sample %s due to %+v", sampleObj.Namespace, err)
		return err
	}
	// Use a DeployerHelper to add components to a Horizon Deployer and deploy them
	deployerHelper, err := util.NewDeployer(sampleCreater.kubeConfig)
	if err != nil {
		log.Errorf("Unable to create a DeployerHelper for the Sample %s: %+v", sampleObj.Namespace, err)
		return err
	}
	deployerHelper.PreDeploy(componentList, sampleObj.Namespace)
	err = deployerHelper.Run()
	if err != nil {
		log.Errorf("Unable to deploy the Sample in the namespace '%s': %+v", sampleObj.Namespace, err)
	}
	deployerHelper.StartControllers()
	return nil
}

// DeleteSample will delete a Sample from the Cluster
func (sampleCreater *Creater) DeleteSample(namespace string) {
	log.Debugf("Deleting a Sample from the namespace %s", namespace)
	// Verify whether the namespace exist
	_, err := util.GetNamespace(sampleCreater.kubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v: %+v", namespace, err)
	}
	// Delete the namespace with its contents
	err = util.DeleteNamespace(sampleCreater.kubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to delete the namespace %+v: %+v", namespace, err)
	}
	// Wait until the namespace is deleted
	for {
		ns, err := util.GetNamespace(sampleCreater.kubeClient, namespace)
		log.Infof("Namespace %v Status: %v", namespace, ns.Status)
		time.Sleep(10 * time.Second)
		if err != nil {
			log.Infof("Deleted the namespace %+v", namespace)
			break
		}
	}

}
