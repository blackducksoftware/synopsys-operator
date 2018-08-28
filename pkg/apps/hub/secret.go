/*
Copyright (C) 2018 Synopsys, Inc.

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

package hub

import (
	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

func (hc *Creater) createHubSecrets(namespace string, adminPassword string, userPassword string) []*components.Secret {
	var secrets []*components.Secret
	hubSecret := components.NewSecret(kapi.SecretConfig{Namespace: namespace, Name: "db-creds", Type: kapi.SecretTypeOpaque})
	hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": adminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": userPassword})
	secrets = append(secrets, hubSecret)
	return secrets
}
