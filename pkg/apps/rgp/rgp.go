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

package rgp

import (
	"fmt"
	"strings"
	"time"

	rgpapi "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	rgplatest "github.com/blackducksoftware/synopsys-operator/pkg/apps/rgp/latest"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	rgpclientset "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Rgp is used to handle Rgps in the cluster
type Rgp struct {
	config      *protoform.Config
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	rgpClient   *rgpclientset.Clientset
	routeClient *routeclient.RouteV1Client
	creaters    []Creater
}

// NewRgp will return an Rgp type
func NewRgp(config *protoform.Config, kubeConfig *rest.Config) *Rgp {
	// Initialiase the clienset
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize the Rgp client
	rgpClient, err := rgpclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize the Route Client for Openshift routes
	routeClient, err := routeclient.NewForConfig(kubeConfig)
	if err != nil {
		routeClient = nil
	}
	// Initialize creaters for different versions of Rgp (each Creater can support differernt versions)
	creaters := []Creater{
		rgplatest.NewCreater(config, kubeConfig, kubeclient, rgpClient, routeClient),
	}

	return &Rgp{
		config:      config,
		kubeConfig:  kubeConfig,
		kubeClient:  kubeclient,
		rgpClient:   rgpClient,
		routeClient: routeClient,
		creaters:    creaters,
	}
}

// getCreater loops through each Creater and returns the one
// that supports the specified version
func (a Rgp) getCreater(version string) (Creater, error) {
	for _, c := range a.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

// Versions returns the versions that the operator supports for Rgp
func (a Rgp) Versions() []string {
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
func (a Rgp) Ensure(rgp *rgpapi.Rgp) error {
	creater, err := a.getCreater(rgp.Spec.Version) // get Creater for the Rgp Version
	if err != nil {
		return err
	}

	return creater.Ensure(rgp) // Ensure the Rgp
}

// Delete will delete the Rgp from the cluster (all Rgps are deleted the same way)
func (a *Rgp) Delete(namespace string) error {
	log.Debugf("Delete Rgp details for %s", namespace)
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
