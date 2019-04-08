/*
Copyright (C) 2019 Synopsys, Inc.

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
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/sirupsen/logrus"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	latestalert "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert/latest"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Alert is used for the Alert deployment
type Alert struct {
	config      *protoform.Config
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	alertClient *alertclientset.Clientset
	routeClient *routeclient.RouteV1Client
	creaters    []Creater
}

// NewAlert will return a Alert
func NewAlert(config *protoform.Config, kubeConfig *rest.Config) *Alert {
	// Initialiase the clienset using kubeConfig
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	alertClient, err := alertclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	routeClient, err := routeclient.NewForConfig(kubeConfig)
	if err != nil {
		routeClient = nil
	} else {
		_, err := util.GetOpenShiftRoutes(routeClient, "default", "docker-registry")
		if err != nil {
			routeClient = nil
		}
	}

	creaters := []Creater{
		latestalert.NewCreater(kubeConfig, kubeclient, alertClient, routeClient),
	}

	return &Alert{
		config:      config,
		kubeConfig:  kubeConfig,
		kubeClient:  kubeclient,
		alertClient: alertClient,
		routeClient: routeClient,
		creaters:    creaters,
	}
}

func (a Alert) getCreater(version string) (Creater, error) {
	for _, c := range a.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

// Delete will be used to delete an Alert instance
func (a Alert) Delete(name string) {
	logrus.Infof("deleting %s", name)
	err := crdupdater.NewCRUDComponents(a.kubeConfig, a.kubeClient, false, name, &api.ComponentList{}, "app=alert").CRUDComponents()
	if err != nil {
		logrus.Error(err)
	}
}

// Versions returns the versions that the operator supports for Alert
func (a Alert) Versions() []string {
	var versions []string
	for _, c := range a.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (a Alert) Ensure(alt *v1.Alert) error {
	creater, err := a.getCreater(alt.Spec.AlertImageVersion)
	if err != nil {
		return err
	}
	return creater.Ensure(alt)
}
