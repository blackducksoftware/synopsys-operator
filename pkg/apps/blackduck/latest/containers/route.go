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

package containers

import (
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
)

// GetOpenShiftRoute creates the OpenShift route component for the webserver
func (c *Creater) GetOpenShiftRoute() *api.Route {
	if strings.ToUpper(c.hubSpec.ExposeService) == util.OPENSHIFT {
		return &api.Route{
			Name:               c.hubSpec.Namespace,
			Namespace:          c.hubSpec.Namespace,
			Kind:               "Service",
			ServiceName:        "webserver",
			PortName:           "port-webserver",
			Labels:             map[string]string{"app": "blackduck", "component": "webserver"},
			TLSTerminationType: routev1.TLSTerminationPassthrough,
		}
	}
	return nil
}
