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

package blackduck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	latestblackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest"
	v1blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Constants for each unit of a deployment of Black Duck
const (
	CRDResources      = "BLACKDUCK"
	DatabaseResources = "DATABASE"
	PVCResources      = "PVC"
)

// Blackduck is used for the Blackduck deployment
type Blackduck struct {
	config          *protoform.Config
	kubeConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	blackduckClient *blackduckclientset.Clientset
	routeClient     *routeclient.RouteV1Client
	creaters        []Creater
}

// NewBlackduck will return a Blackduck
func NewBlackduck(config *protoform.Config, kubeConfig *rest.Config) *Blackduck {
	// Initialiase the clienset using kubeConfig
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	routeClient := util.GetRouteClient(kubeConfig, kubeclient, config.Namespace)

	creaters := []Creater{
		v1blackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, routeClient),
		latestblackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, routeClient),
	}

	return &Blackduck{
		config:          config,
		kubeConfig:      kubeConfig,
		kubeClient:      kubeclient,
		blackduckClient: blackduckClient,
		routeClient:     routeClient,
		creaters:        creaters,
	}
}

func (b *Blackduck) getCreater(version string) (Creater, error) {
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version '%s' is not supported.  Supported versions: %s", version, strings.Join(b.Versions(), ", "))
}

func (b *Blackduck) ensureVersion(bd *v1.Blackduck) error {
	versions := b.Versions()
	// If the version is not provided, then we set it to be the latest
	if len(bd.Spec.Version) == 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		bd.Spec.Version = versions[0]
	} else {
		// If the verion is provided, check that it's supported
		for _, v := range versions {
			if strings.Compare(v, bd.Spec.Version) == 0 {
				return nil
			}
		}
		return fmt.Errorf("version '%s' is not supported.  Supported versions: %s", bd.Spec.Version, strings.Join(versions, ", "))
	}
	return nil
}

// Delete will be used to delete a blackduck instance
func (b *Blackduck) Delete(name string) error {
	log.Infof("deleting a %s Black Duck instance", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("invalid name to delete the Black Duck instance")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
		ns, err := util.ListNamespaces(b.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.BlackDuckName, name))
		if err != nil {
			log.Errorf("unable to list %s Black Duck instance namespaces %s due to %+v", name, namespace, err)
		}
		if len(ns.Items) > 0 {
			namespace = ns.Items[0].Name
		} else {
			return fmt.Errorf("unable to find %s Black Duck instance namespace", name)
		}
	} else {
		name = values[1]
		namespace = values[0]
	}

	// delete the Black Duck instance
	commonConfig := crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, false, namespace, "",
		&api.ComponentList{}, fmt.Sprintf("app=%s,name=%s", util.BlackDuckName, name), false)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		return fmt.Errorf("unable to delete the %s Black Duck instance in %s namespace due to %+v", name, namespace, crudErrors)
	}

	var err error
	// if cluster scope, if no other instance running in Synopsys Operator namespace, delete the namespace or delete the Synopsys labels in the namespace
	if b.config.IsClusterScoped {
		err = util.DeleteResourceNamespace(b.kubeClient, util.BlackDuckName, namespace, name, false)
	} else {
		// if namespace scope, delete the label from the namespace
		_, err = util.CheckAndUpdateNamespace(b.kubeClient, util.BlackDuckName, namespace, name, "", true)
	}
	if err != nil {
		return err
	}

	return nil
}

// Versions returns the versions that the operator supports
func (b *Blackduck) Versions() []string {
	var versions []string
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (b *Blackduck) Ensure(bd *v1.Blackduck) error {
	// If the version is not specified then we set it to be the latest.
	if err := b.ensureVersion(bd); err != nil {
		return err
	}
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}

	return creater.Ensure(bd)
}

// GetComponents gets the Black Duck's creater and returns the components
func (b Blackduck) GetComponents(bd *blackduckapi.Blackduck, compType string) (*api.ComponentList, error) {
	// If the version is not specified then we set it to be the latest.
	if err := b.ensureVersion(bd); err != nil {
		return nil, err
	}
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return nil, err
	}
	switch strings.ToUpper(compType) {
	case CRDResources:
		return creater.GetComponents(bd)
	case DatabaseResources:
		return creater.GetPostgresComponents(bd)
	case PVCResources:
		pvcs := creater.GetPVC(bd)
		return &api.ComponentList{PersistentVolumeClaims: pvcs}, nil
	}
	return nil, fmt.Errorf("invalid components type '%s'", compType)
}
