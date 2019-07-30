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

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/client-go/kubernetes"
)

// OpsSightRoute holds the OpsSight Route configuration
type OpsSightRoute struct {
	opsSight *opssightapi.OpsSight
}

func init() {
	store.Register(types.OpsSightCoreRouteV1, NewOpsSightRoute)
}

// NewOpsSightRoute returns the OpsSight Route configuration
func NewOpsSightRoute(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.RouteInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightRoute{opsSight: opsSight}, nil
}

// GetRoute returns the route
func (o *OpsSightRoute) GetRoute() (*api.Route, error) {
	namespace := o.opsSight.Spec.Namespace
	if strings.ToUpper(o.opsSight.Spec.Perceptor.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceptor.Name),
			Namespace:          namespace,
			Kind:               "Service",
			ServiceName:        utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.Perceptor.Name),
			PortName:           fmt.Sprintf("port-%s", o.opsSight.Spec.Perceptor.Name),
			Labels:             map[string]string{"app": "opssight", "name": o.opsSight.Name, "component": fmt.Sprintf("%s-ui", o.opsSight.Spec.Perceptor.Name)},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}, nil
	}
	return nil, nil
}
