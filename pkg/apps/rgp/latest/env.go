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

package rgp

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
)

func (g *SpecConfig) getCommonEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "CONNECTION_POOL_SIZE", KeyOrVal: "10"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "LOG_LEVEL", KeyOrVal: "INFO"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SPRING_PROFILE", KeyOrVal: "production"})
	return envs
}

func (g *SpecConfig) getSwipEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_ROOT_DOMAIN", KeyOrVal: g.config.IngressHost})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "SWIP_ENVIRONMENT_NAME", KeyOrVal: g.config.Namespace})
	return envs
}

func (g *SpecConfig) getPostgresEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_HOST", KeyOrVal: "postgres"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_PORT", KeyOrVal: "5432"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRES_USERNAME", KeyOrVal: "postgres"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRES_PASSWORD", KeyOrVal: "POSTGRES_PASSWORD", FromName: "db-creds"})
	return envs
}

func (g *SpecConfig) getMongoEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_HOST", KeyOrVal: "mongodb"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_PORT", KeyOrVal: "27017"})
	return envs
}

func (g *SpecConfig) getEventStoreLegacyEnvConfigs() []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_ADDR", KeyOrVal: "eventstore"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_USERNAME", KeyOrVal: "admin"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "EVENT_STORE_PASSWORD", KeyOrVal: "password", FromName: "swip-eventstore-creds"})
	return envs
}

func (g *SpecConfig) getEventStoreEnvConfigs(role string) []*horizonapi.EnvConfig {
	var envs []*horizonapi.EnvConfig
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "EVENT_STORE_ADDR", KeyOrVal: "eventstore"})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: fmt.Sprintf("EVENT_STORE_%s_USERNAME", strings.ToUpper(role)), KeyOrVal: role})
	envs = append(envs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: fmt.Sprintf("EVENT_STORE_%s_PASSWORD", strings.ToUpper(role)), KeyOrVal: "password", FromName: "swip-eventstore-creds"})
	return envs
}
