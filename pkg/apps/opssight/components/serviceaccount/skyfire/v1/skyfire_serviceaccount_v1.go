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

// OpsSightServiceAccount holds the OpsSight service account configuration
type OpsSightServiceAccount struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.SkyfireServiceAccountV1, NewOpsSightServiceAccount)
}

// NewOpsSightServiceAccount returns the OpsSight service account configuration
func NewOpsSightServiceAccount(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ServiceAccountInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightServiceAccount{config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetServiceAccount returns the service account
func (o *OpsSightServiceAccount) GetServiceAccount() (*components.ServiceAccount, error) {
	if !o.opsSight.Spec.EnableSkyfire {
		return nil, nil
	}

	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, "skyfire"),
		Namespace: o.opsSight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"component": "skyfire", "app": "opssight", "name": o.opsSight.Name})
	return serviceAccount, nil
}
