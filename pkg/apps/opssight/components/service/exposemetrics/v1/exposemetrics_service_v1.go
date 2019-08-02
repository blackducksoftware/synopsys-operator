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

package v1

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// OpsSightService holds the OpsSight service configuration
type OpsSightService struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.OpsSightExposeMetricsServiceV1, NewOpsSightService)
}

// NewOpsSightService returns the OpsSight service account configuration
func NewOpsSightService(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ServiceInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightService{config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetService returns the service
func (o *OpsSightService) GetService() (*components.Service, error) {
	if !o.opsSight.Spec.EnableMetrics {
		return nil, nil
	}

	var svc *components.Service
	var err error
	switch strings.ToUpper(o.opsSight.Spec.Prometheus.Expose) {
	case util.NODEPORT:
		svc, err = o.opsSightMetricsNodePortService()
		break
	case util.LOADBALANCER:
		svc, err = o.opsSightMetricsLoadBalancerService()
		break
	default:
	}
	return svc, err
}

// opsSightMetricsNodePortService creates a nodeport service for OpsSight metrics
func (o *OpsSightService) opsSightMetricsNodePortService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus-exposed"),
		Namespace: o.opsSight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"component": "prometheus-exposed", "app": "opssight", "name": o.opsSight.Name})
	service.AddSelectors(map[string]string{"component": "prometheus-exposed", "app": "opssight", "name": o.opsSight.Name})

	return service, err
}

// opsSightMetricsLoadBalancerService creates a loadbalancer service for OpsSight metrics
func (o *OpsSightService) opsSightMetricsLoadBalancerService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "prometheus-exposed"),
		Namespace: o.opsSight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": o.opsSight.Name})
	service.AddSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": o.opsSight.Name})

	return service, err
}
