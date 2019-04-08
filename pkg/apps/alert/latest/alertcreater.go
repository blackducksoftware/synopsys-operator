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

package alert

import (
	"fmt"
	"time"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the Blackduck
type Creater struct {
	Config      *protoform.Config
	KubeConfig  *rest.Config
	KubeClient  *kubernetes.Clientset
	AlertClient *alertclientset.Clientset
	RouteClient *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, alertClient *alertclientset.Clientset, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, AlertClient: alertClient, RouteClient: routeClient}
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (ac *Creater) Ensure(alert *alertapi.Alert) error {
	// Get Components
	specConfig := NewAlert(&alert.Spec)
	cpList, err := specConfig.GetComponents()
	if err != nil {
		return err
	}
	// Update components in cluster
	commonConfig := crdupdater.NewCRUDComponents(ac.KubeConfig, ac.KubeClient, ac.Config.DryRun, alert.Spec.Namespace,
		cpList, "app=alert")
	errors := commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update Alert components due to %+v", errors)
	}
	// Check Pods are running
	if err := util.ValidatePodsAreRunningInNamespace(ac.KubeClient, alert.Spec.Namespace, 600); err != nil {
		return err
	}
	return nil
}

// Versions will return the versions supported
func (ac *Creater) Versions() []string {
	return GetVersions()
}

// DeleteAlert will delete the Black Duck Alert
func (ac *Creater) DeleteAlert(namespace string) {
	log.Debugf("Delete Alert details for %s", namespace)
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(ac.KubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(ac.KubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(ac.KubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}
}

// CreateAlert will create the Black Duck Alert
func (ac *Creater) CreateAlert(createAlert *alertapi.AlertSpec) error {
	log.Debugf("Create Alert details for %s: %+v", createAlert.Namespace, createAlert)
	alert := NewAlert(createAlert)
	components, err := alert.GetComponents()
	if err != nil {
		log.Errorf("unable to get alert components for %s due to %+v", createAlert.Namespace, err)
		return err
	}
	deployer, err := util.NewDeployer(ac.KubeConfig)
	if err != nil {
		log.Errorf("unable to get deployer object for %s due to %+v", createAlert.Namespace, err)
		return err
	}
	deployer.PreDeploy(components, createAlert.Namespace)
	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to deploy alert app due to %+v", err)
	}
	deployer.StartControllers()

	// Create Route if on Openshift
	if ac.RouteClient != nil {
		log.Debugf("Creating an Openshift Route for Alert")
		_, err := util.CreateOpenShiftRoutes(ac.RouteClient, createAlert.Namespace, createAlert.Namespace, "Service", "alert")
		if err != nil {
			log.Errorf("unable to create the openshift route due to %+v", err)
		}
	}
	return nil
}
