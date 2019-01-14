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
	"strconv"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/hub/v2"
	"github.com/blackducksoftware/synopsys-operator/pkg/hub/containers"
)

// CreateHubConfig will create the hub configMaps
func (hc *Creater) createHubConfig(createHub *v2.HubSpec, hubContainerFlavor *containers.ContainerFlavor) map[string]*components.ConfigMap {
	configMaps := make(map[string]*components.ConfigMap)

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-config"})
	hubData := map[string]string{
		"PUBLIC_HUB_WEBSERVER_HOST": "localhost",
		"PUBLIC_HUB_WEBSERVER_PORT": "443",
		"HUB_WEBSERVER_PORT":        "8443",
		"IPV4_ONLY":                 "0",
		"RUN_SECRETS_DIR":           "/tmp/secrets",
		"HUB_PROXY_NON_PROXY_HOSTS": "solr",
	}

	for _, value := range createHub.Environs {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.Trim(values[0], " ")
			mapValue := strings.Trim(values[1], " ")
			if len(mapKey) > 0 && len(mapValue) > 0 {
				hubData[mapKey] = mapValue
			}
		}
	}

	hubConfig.AddData(hubData)

	configMaps["hub-config"] = hubConfig

	hubDbConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-db-config"})

	if createHub.ExternalPostgres != (v2.PostgresExternalDBConfig{}) {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": createHub.ExternalPostgres.PostgresAdmin,
			"HUB_POSTGRES_USER":  createHub.ExternalPostgres.PostgresUser,
			"HUB_POSTGRES_PORT":  strconv.Itoa(createHub.ExternalPostgres.PostgresPort),
			"HUB_POSTGRES_HOST":  createHub.ExternalPostgres.PostgresHost,
		})
	} else {
		hubDbConfig.AddData(map[string]string{
			"HUB_POSTGRES_ADMIN": "blackduck",
			"HUB_POSTGRES_USER":  "blackduck_user",
			"HUB_POSTGRES_PORT":  "5432",
			"HUB_POSTGRES_HOST":  "postgres",
		})
	}

	configMaps["hub-db-config"] = hubDbConfig

	hubConfigResources := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-config-resources"})
	hubConfigResources.AddData(map[string]string{
		"webapp-mem":    hubContainerFlavor.WebappHubMaxMemory,
		"jobrunner-mem": hubContainerFlavor.JobRunnerHubMaxMemory,
		"scan-mem":      hubContainerFlavor.ScanHubMaxMemory,
	})

	configMaps["hub-config-resources"] = hubConfigResources

	hubDbConfigGranular := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-db-config-granular"})
	if createHub.ExternalPostgres != (v2.PostgresExternalDBConfig{}) {
		hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": strconv.FormatBool(createHub.ExternalPostgres.PostgresSsl)})
		if createHub.ExternalPostgres.PostgresSsl {
			hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL_CERT_AUTH": "false"})
		}
	} else {
		hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})
	}

	configMaps["hub-db-config-granular"] = hubDbConfigGranular

	return configMaps
}
