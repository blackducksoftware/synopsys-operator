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

// Creater will store the configuration to create the Blackduck
type Creater struct {
	kubeConfig   *rest.Config
	kubeClient   *kubernetes.Clientset
	sampleClient *sampleclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, sampleClient *sampleclientset.Clientset) *Creater {
	return &Creater{kubeConfig: kubeConfig, kubeClient: kubeClient, sampleClient: sampleClient}
}

// DeleteSample will delete the Sample
func (sampleCreater *Creater) DeleteSample(namespace string) {
	log.Debugf("Delete Sample details for %s", namespace)
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(sampleCreater.kubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(sampleCreater.kubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(sampleCreater.kubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}
}

// CreateSample will create the Sample
func (sampleCreater *Creater) CreateSample(createSample *v1.SampleSpec) error {
	log.Debugf("Create Sample details for %s: %+v", createSample.Namespace, createSample)
	sample := NewSample(createSample)
	components, err := sample.GetComponents()
	if err != nil {
		log.Errorf("unable to get sample components for %s due to %+v", createSample.Namespace, err)
		return err
	}
	deployer, err := util.NewDeployer(sampleCreater.kubeConfig)
	if err != nil {
		log.Errorf("unable to get deployer object for %s due to %+v", createSample.Namespace, err)
		return err
	}
	deployer.PreDeploy(components, createSample.Namespace)
	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to deploy sample app due to %+v", err)
	}
	deployer.StartControllers()
	return nil
}
