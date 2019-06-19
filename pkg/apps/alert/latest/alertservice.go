/*
Copyright (C) 2019 Synopsys, Inc.

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

package alert

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// getAlertService returns a new cluster Service for an Alert
func (a *SpecConfig) getAlertService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "alert"),
		Namespace: a.alert.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*a.alert.Spec.Port),
		TargetPort: fmt.Sprintf("%d", *a.alert.Spec.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("%d-tcp", *a.alert.Spec.Port),
	})

	service.AddSelectors(map[string]string{"app": util.AlertName, "component": "alert"})
	service.AddLabels(map[string]string{"app": util.AlertName, "component": "alert"})
	return service
}

// getAlertServiceNodePort returns a new Node Port Service for an Alert
func (a *SpecConfig) getAlertServiceNodePort() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "exposed"),
		Namespace: a.alert.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*a.alert.Spec.Port),
		TargetPort: fmt.Sprintf("%d", *a.alert.Spec.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("%d-tcp", *a.alert.Spec.Port),
	})

	service.AddSelectors(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	service.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	return service
}

// getAlertServiceLoadBalancer returns a new Load Balancer Service for an Alert
func (a *SpecConfig) getAlertServiceLoadBalancer() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "exposed"),
		Namespace: a.alert.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*a.alert.Spec.Port),
		TargetPort: fmt.Sprintf("%d", *a.alert.Spec.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("%d-tcp", *a.alert.Spec.Port),
	})

	service.AddSelectors(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	service.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	return service
}
