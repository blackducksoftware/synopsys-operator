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
	store.Register(types.OpsSightImageProcessorServiceV1, NewOpsSightService)
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
	if !o.opsSight.Spec.Perceiver.EnableImagePerceiver {
		return nil, nil
	}

	name := o.opsSight.Spec.Perceiver.ImagePerceiver.Name
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, name),
		Namespace: o.opsSight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(o.opsSight.Spec.Perceiver.Port),
		TargetPort: fmt.Sprintf("%d", o.opsSight.Spec.Perceiver.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})

	service.AddLabels(map[string]string{"component": name, "app": "opssight", "name": o.opsSight.Name})
	service.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": o.opsSight.Name})

	return service, nil
}
