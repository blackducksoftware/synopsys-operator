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
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdService holds the Black Duck service configuration
type BdService struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// GetService returns the service
func (b BdService) GetService() *components.Service {
	var svc *components.Service

	switch strings.ToUpper(b.blackDuck.Spec.ExposeService) {
	case util.NODEPORT:
		svc = util.CreateService(apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "webserver-exposed"), apputils.GetLabel("webserver", b.blackDuck.Name), b.blackDuck.Spec.Namespace, int32(443), int32(8443), horizonapi.ServiceTypeNodePort, apputils.GetLabel("webserver-exposed", b.blackDuck.Name))

		break
	case util.LOADBALANCER:
		svc = util.CreateService(apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "webserver-exposed"), apputils.GetLabel("webserver", b.blackDuck.Name), b.blackDuck.Spec.Namespace, int32(443), int32(8443), horizonapi.ServiceTypeLoadBalancer, apputils.GetLabel("webserver-exposed", b.blackDuck.Name))
		break
	default:
	}

	return svc
}

// NewBdService returns the Black Duck service configuration
func NewBdService(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ServiceInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdService{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

func init() {
	store.Register(types.BlackDuckExposeServiceV1, NewBdService)
}
