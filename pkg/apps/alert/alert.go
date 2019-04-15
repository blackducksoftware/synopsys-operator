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
	"time"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	latestalert "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert/latest"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Alert is used to handle Alerts in the cluster
type Alert struct {
	config      *protoform.Config
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	alertClient *alertclientset.Clientset
	routeClient *routeclient.RouteV1Client
	creaters    []Creater
}

// NewAlert will return an Alert type
func NewAlert(config *protoform.Config, kubeConfig *rest.Config) *Alert {
	// Initialiase the clienset
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize the Alert client
	alertClient, err := alertclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize the Route Client for Openshift routes
	routeClient, err := routeclient.NewForConfig(kubeConfig)
	if err != nil {
		routeClient = nil
	} else {
		_, err := util.GetOpenShiftRoutes(routeClient, "default", "docker-registry")
		if err != nil {
			routeClient = nil
		}
	}
	// Initialize creaters for different versions of Alert (each Creater can support differernt versions)
	creaters := []Creater{
		latestalert.NewCreater(config, kubeConfig, kubeclient, alertClient, routeClient),
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

// getCreater loops through each Creater and returns the one
// that supports the specified version
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

// Versions returns the versions that the operator supports for Alert
func (a Alert) Versions() []string {
	var versions []string
	// Get versions that each Creater supports
	for _, c := range a.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will get the necessary Creater and make sure the instance
// is correctly deployed or deploy it if needed
func (a Alert) Ensure(alt *alertapi.Alert) error {
	creater, err := a.getCreater(alt.Spec.Version) // get Creater for the Alert Version
	if err != nil {
		return err
	}

	return creater.Ensure(alt) // Ensure the Alert
}

// Delete will delete the Alert from the cluster (all Alerts are deleted the same way)
func (a *Alert) Delete(namespace string) error {
	log.Debugf("Delete Alert details for %s", namespace)
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(a.kubeClient, namespace)
	if err != nil {
		return fmt.Errorf("unable to find the namespace %+v due to %+v", namespace, err)
	}
	// Delete the namespace
	err = util.DeleteNamespace(a.kubeClient, namespace)
	if err != nil {
		return fmt.Errorf("unable to delete the namespace %+v due to %+v", namespace, err)
	}
	// Verify whether the namespace deleted
	var attempts = 30
	var retryWait time.Duration = 10
	for i := 0; i <= attempts; i++ {
		_, err := util.GetNamespace(a.kubeClient, namespace)
		if err != nil {
			log.Infof("Deleted the namespace %+v", namespace)
			break
		}
		if i >= 10 {
			return fmt.Errorf("unable to delete the namespace %+v after %f minutes", namespace, float64(attempts)*retryWait.Seconds()/60)
		}
		time.Sleep(retryWait * time.Second)
	}
	return nil
}
