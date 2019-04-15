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

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater stores the configuration and clients to create specific versions of Alerts
type Creater struct {
	Config      *protoform.Config
	KubeConfig  *rest.Config
	KubeClient  *kubernetes.Clientset
	AlertClient *alertclientset.Clientset
	RouteClient *routeclient.RouteV1Client
}

// NewCreater returns this Alert Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, alertClient *alertclientset.Clientset, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, AlertClient: alertClient, RouteClient: routeClient}
}

// Versions is an Interface function that returns the versions supported by this Creater
func (ac *Creater) Versions() []string {
	return GetVersions()
}

// Ensure is an Interface function that will make sure the instance is correctly deployed or deploy it if needed
func (ac *Creater) Ensure(alert *alertapi.Alert) error {
	// Get Kubernetes Components for the Alert
	specConfig := NewSpecConfig(&alert.Spec)
	cpList, err := specConfig.GetComponents()
	if err != nil {
		return err
	}
	// Update components in cluster
	commonConfig := crdupdater.NewCRUDComponents(ac.KubeConfig, ac.KubeClient, ac.Config.DryRun, alert.Spec.Namespace, cpList, "app=alert")
	errors := commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update Alert components due to %+v", errors)
	}

	// Create Route if on Openshift
	if ac.RouteClient != nil && alert.Spec.ExposeService == "OPENSHIFT" {
		log.Debugf("Creating an Openshift Route for Alert")
		_, err := util.CreateOpenShiftRoutes(ac.RouteClient, alert.Spec.Namespace, alert.Spec.Namespace, "Service", "alert", routev1.TLSTerminationPassthrough)
		if err != nil {
			log.Errorf("unable to create the openshift route due to %+v", err)
		}
	}
	return nil
}
