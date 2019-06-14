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
	"strconv"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetPostgres will return the postgres object
func (c *Creater) GetPostgres() *postgres.Postgres {
	postgresImage := c.GetFullContainerNameFromImageRegistryConf("postgres")
	if len(postgresImage) == 0 {
		postgresImage = "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1"
	}

	name := util.GetResourceName(c.name, util.BlackDuckName, "postgres", c.config.IsClusterScoped)

	var pvcName string
	if c.hubSpec.PersistentStorage {
		pvcName = util.GetResourceName(c.name, util.BlackDuckName, "postgres", c.config.IsClusterScoped)
	}

	return &postgres.Postgres{
		Name:                   name,
		Namespace:              c.hubSpec.Namespace,
		PVCName:                pvcName,
		Port:                   PostgresPort,
		Image:                  postgresImage,
		MinCPU:                 c.hubContainerFlavor.PostgresCPULimit,
		MaxCPU:                 "",
		MinMemory:              c.hubContainerFlavor.PostgresMemoryLimit,
		MaxMemory:              "",
		Database:               "blackduck",
		User:                   "blackduck",
		PasswordSecretName:     util.GetResourceName(c.name, util.BlackDuckName, "db-creds", c.config.IsClusterScoped),
		UserPasswordSecretKey:  "HUB_POSTGRES_ADMIN_PASSWORD_FILE",
		AdminPasswordSecretKey: "HUB_POSTGRES_POSTGRES_PASSWORD_FILE",
		MaxConnections:         300,
		SharedBufferInMB:       1024,
		EnvConfigMapRefs:       []string{util.GetResourceName(c.name, util.BlackDuckName, "db-config", c.config.IsClusterScoped)},
		Labels:                 c.GetLabel("postgres"),
	}
}

// GetPostgresConfigmap will return the postgres configMaps
func (c *Creater) GetPostgresConfigmap() *components.ConfigMap {
	// DB
	hubDbConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, util.BlackDuckName, "db-config", c.config.IsClusterScoped)})
	if c.hubSpec.ExternalPostgres != nil {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": c.hubSpec.ExternalPostgres.PostgresAdmin,
			"HUB_POSTGRES_USER":  c.hubSpec.ExternalPostgres.PostgresUser,
			"HUB_POSTGRES_PORT":  strconv.Itoa(c.hubSpec.ExternalPostgres.PostgresPort),
			"HUB_POSTGRES_HOST":  c.hubSpec.ExternalPostgres.PostgresHost,
		})
	} else {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": "blackduck",
			"HUB_POSTGRES_USER":  "blackduck_user",
			"HUB_POSTGRES_PORT":  "5432",
			"HUB_POSTGRES_HOST":  util.GetResourceName(c.name, util.BlackDuckName, "postgres", c.config.IsClusterScoped),
		})
	}

	if c.hubSpec.ExternalPostgres != nil {
		hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": strconv.FormatBool(c.hubSpec.ExternalPostgres.PostgresSsl)})
		if c.hubSpec.ExternalPostgres.PostgresSsl {
			hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL_CERT_AUTH": "false"})
		}
	} else {
		hubDbConfig.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})
	}
	hubDbConfig.AddLabels(c.GetVersionLabel("postgres"))

	return hubDbConfig
}

// GetPostgresSecret will return the postgres secret
func (c *Creater) GetPostgresSecret(adminPassword string, userPassword string, postgresPassword string) *components.Secret {
	hubSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, util.BlackDuckName, "db-creds", c.config.IsClusterScoped), Type: horizonapi.SecretTypeOpaque})

	if c.hubSpec.ExternalPostgres != nil {
		hubSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(c.hubSpec.ExternalPostgres.PostgresAdminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(c.hubSpec.ExternalPostgres.PostgresUserPassword)})
	} else {
		hubSecret.AddData(map[string][]byte{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": []byte(adminPassword), "HUB_POSTGRES_USER_PASSWORD_FILE": []byte(userPassword), "HUB_POSTGRES_POSTGRES_PASSWORD_FILE": []byte(postgresPassword)})
	}
	hubSecret.AddLabels(c.GetVersionLabel("postgres"))

	return hubSecret
}
