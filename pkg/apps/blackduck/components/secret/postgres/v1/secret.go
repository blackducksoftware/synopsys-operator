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
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdRSecret holds the Black Duck secret configuration
type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// GetSecrets returns the secret
func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret

	var adminPassword, userPassword, postgresPassword string
	var err error
	if b.blackDuck.Spec.ExternalPostgres != nil {

		adminPassword, err = util.Base64Decode(b.blackDuck.Spec.ExternalPostgres.PostgresAdminPassword)
		if err != nil {
			return nil
		}

		userPassword, err = util.Base64Decode(b.blackDuck.Spec.ExternalPostgres.PostgresUserPassword)
		if err != nil {
			return nil
		}

	} else {
		adminPassword, err = util.Base64Decode(b.blackDuck.Spec.AdminPassword)
		if err != nil {
			return nil
		}

		userPassword, err = util.Base64Decode(b.blackDuck.Spec.UserPassword)
		if err != nil {
			return nil
		}

		postgresPassword, err = util.Base64Decode(b.blackDuck.Spec.PostgresPassword)
		if err != nil {
			return nil
		}

	}

	postgresSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "db-creds"), Type: horizonapi.SecretTypeOpaque})

	if b.blackDuck.Spec.ExternalPostgres != nil {
		postgresSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(adminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(userPassword)})
	} else {
		postgresSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(adminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(userPassword), "HUB_POSTGRES_POSTGRES_PASSWORD_FILE": []byte(postgresPassword)})
	}
	postgresSecret.AddLabels(apputils.GetVersionLabel("postgres", b.blackDuck.Name, b.blackDuck.Spec.Version))

	secrets = append(secrets, postgresSecret)
	return secrets
}

// NewBdRSecret returns the Black Duck secret configuration
func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.SecretInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdRSecret{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

func init() {
	store.Register(types.BlackDuckPostgresSecretV1, NewBdRSecret)
}
