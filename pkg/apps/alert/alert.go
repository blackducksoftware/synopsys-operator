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
	"sort"
	"strings"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	latestalert "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert/latest"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Constants for each unit of a deployment of Alert
const (
	CRDResources = "ALERT"
	PVCResources = "PVC"
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
func (a *Alert) getCreater(version string) (Creater, error) {
	for _, c := range a.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

func (a *Alert) ensureVersion(alt *alertapi.Alert) error {
	versions := a.Versions()
	// If the version is not provided, then we set it to be the latest
	if len(alt.Spec.Version) == 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		alt.Spec.Version = versions[0]
	} else {
		// If the verion is provided, check that it's supported
		for _, v := range versions {
			if strings.Compare(v, alt.Spec.Version) == 0 {
				return nil
			}
		}
		return fmt.Errorf("version '%s' is not supported", alt.Spec.Version)
	}
	return nil
}

// Versions returns the versions that the operator supports for Alert
func (a *Alert) Versions() []string {
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
func (a *Alert) Ensure(alt *alertapi.Alert) error {
	// If the version is not specified then we set it to be the latest.
	if err := a.ensureVersion(alt); err != nil {
		return err
	}
	creater, err := a.getCreater(alt.Spec.Version) // get Creater for the Alert Version
	if err != nil {
		return err
	}

	return creater.Ensure(alt) // Ensure the Alert
}

// Delete will delete the Alert from the cluster (all Alerts are deleted the same way)
func (a *Alert) Delete(name string) error {
	log.Debugf("deleting %s Alert instance", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("invalid name to delete the Alert instance")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
		ns, err := util.ListNamespaces(a.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.AlertName, name))
		if err != nil {
			log.Errorf("unable to list %s Alert instance namespaces %s due to %+v", name, namespace, err)
		}
		if len(ns.Items) > 0 {
			namespace = ns.Items[0].Name
		} else {
			log.Errorf("unable to find %s Alert instance namespace", name)
			return fmt.Errorf("unable to find %s Alert instance namespace", name)
		}
	} else {
		name = values[1]
		namespace = values[0]
	}

	// delete an Alert instance
	commonConfig := crdupdater.NewCRUDComponents(a.kubeConfig, a.kubeClient, a.config.DryRun, false, namespace,
		&api.ComponentList{}, fmt.Sprintf("app=%s,name=%s", util.AlertName, name), false)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		return fmt.Errorf("unable to delete the %s Alert instance in %s namespace due to %+v", name, namespace, crudErrors)
	}

	if a.config.IsClusterScoped {
		err := util.DeleteResourceNamespace(a.kubeConfig, a.kubeClient, a.config.CrdNames, namespace, false)
		if err != nil {
			return errors.Annotatef(err, "unable to delete namespace %s", namespace)
		}
	}

	// update the namespace label if the version of the app got deleted
	if isNamespaceExist, err := util.CheckAndUpdateNamespace(a.kubeClient, util.AlertName, namespace, name, "", true); isNamespaceExist {
		return err
	}

	return nil
}

// GetComponents gets the necessary creater and returns the Alert's components
func (a *Alert) GetComponents(alt *alertapi.Alert, compType string) (*api.ComponentList, error) {
	// If the version is not specified then we set it to be the latest.
	if err := a.ensureVersion(alt); err != nil {
		return nil, err
	}
	creater, err := a.getCreater(alt.Spec.Version) // get Creater for the Alert Version
	if err != nil {
		return nil, err
	}
	switch strings.ToUpper(compType) {
	case CRDResources:
		return creater.GetComponents(alt)
	case PVCResources:
		pvcs, err := creater.GetPVC(alt)
		return &api.ComponentList{PersistentVolumeClaims: pvcs}, err
	}
	return nil, fmt.Errorf("invalid components type '%s'", compType)
}
