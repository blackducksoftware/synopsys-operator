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
	"encoding/json"
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

// OpsSightSecret holds the OpsSight secret configuration
type OpsSightSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	opsSight   *opssightapi.OpsSight
}

func init() {
	store.Register(types.OpsSightSecretV1, NewOpsSightSecret)
}

// NewOpsSightSecret returns the OpsSight secret configuration
func NewOpsSightSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.SecretInterface, error) {
	opsSight, ok := cr.(*opssightapi.OpsSight)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to OpsSight object")
	}
	return &OpsSightSecret{config: config, kubeClient: kubeClient, opsSight: opsSight}, nil
}

// GetSecret returns the secret
func (o *OpsSightSecret) GetSecret() (*components.Secret, error) {
	secretConfig := horizonapi.SecretConfig{
		Name:      utils.GetResourceName(o.opsSight.Name, util.OpsSightName, o.opsSight.Spec.SecretName),
		Namespace: o.opsSight.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)

	// empty data fields that will be overwritten
	emptyHosts := make(map[string]*opssightapi.Host)
	bytes, err := json.Marshal(emptyHosts)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal OpsSight's Host struct: %+v", err)
	}
	secret.AddData(map[string][]byte{o.opsSight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	emptySecuredRegistries := make(map[string]*opssightapi.RegistryAuth)
	bytes, err = json.Marshal(emptySecuredRegistries)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal secured registries struct: %+v", err)
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	secret.AddLabels(map[string]string{"component": o.opsSight.Spec.SecretName, "app": "opssight", "name": o.opsSight.Name})

	return secret, nil
}
