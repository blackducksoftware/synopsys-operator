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
	"strings"
)

type Flavor string

const (
	SMALL    Flavor = "SMALL"
	MEDIUM   Flavor = "MEDIUM"
	LARGE    Flavor = "LARGE"
	OPSSIGHT Flavor = "OPSSIGHT"
)

type HubContainerFlavor struct {
	WebserverMemoryLimit       string
	SolrMemoryLimit            string
	WebappCpuLimit             string
	WebappMemoryLimit          string
	WebappHubMaxMemory         string
	ScanReplicas               int32
	ScanMemoryLimit            string
	ScanHubMaxMemory           string
	JobRunnerReplicas          int32
	JobRunnerMemoryLimit       string
	JobRunnerHubMaxMemory      string
	CfsslMemoryLimit           string
	LogstashMemoryLimit        string
	RegistrationMemoryLimit    string
	ZookeeperMemoryLimit       string
	AuthenticationMemoryLimit  string
	AuthenticationHubMaxMemory string
	DocumentationMemoryLimit   string
	PostgresMemoryLimit        string
	PostgresCpuLimit           string
}

func GetHubContainersFlavor(flavor string) *HubContainerFlavor {
	switch Flavor(strings.ToUpper(flavor)) {
	case SMALL:
		return &HubContainerFlavor{
			WebserverMemoryLimit:       SMALL_WEBSERVER_MEMORY_LIMIT,
			SolrMemoryLimit:            SMALL_SOLR_MEMORY_LIMIT,
			WebappCpuLimit:             SMALL_WEBAPP_CPU_LIMIT,
			WebappMemoryLimit:          SMALL_WEBAPP_MEMORY_LIMIT,
			WebappHubMaxMemory:         SMALL_WEBAPP_HUB_MAX_MEMORY,
			ScanReplicas:               SMALL_SCAN_REPLICAS,
			ScanMemoryLimit:            SMALL_SCAN_MEMORY_LIMIT,
			ScanHubMaxMemory:           SMALL_SCAN_HUB_MAX_MEMORY,
			JobRunnerReplicas:          SMALL_JOBRUNNER_REPLICAS,
			JobRunnerMemoryLimit:       SMALL_JOBRUNNER_MEMORY_LIMIT,
			JobRunnerHubMaxMemory:      SMALL_JOBRUNNER_HUB_MAX_MEMORY,
			CfsslMemoryLimit:           CFSSL_MEMORY_LIMIT,
			LogstashMemoryLimit:        LOGSTASH_MEMORY_LIMIT,
			RegistrationMemoryLimit:    REGISTRATION_MEMORY_LIMIT,
			ZookeeperMemoryLimit:       ZOOKEEPER_MEMORY_LIMIT,
			AuthenticationMemoryLimit:  AUTHENTICATION_MEMORY_LIMIT,
			AuthenticationHubMaxMemory: AUTHENTICATION_HUB_MAX_MEMORY,
			DocumentationMemoryLimit:   DOCUMENTATION_MEMORY_LIMIT,
			PostgresMemoryLimit:        SMALL_POSTGRES_MEMORY_LIMIT,
			PostgresCpuLimit:           SMALL_POSTGRES_CPU_LIMIT,
		}
	case MEDIUM:
		return &HubContainerFlavor{
			WebserverMemoryLimit:       MEDIUM_WEBSERVER_MEMORY_LIMIT,
			SolrMemoryLimit:            MEDIUM_SOLR_MEMORY_LIMIT,
			WebappCpuLimit:             MEDIUM_WEBAPP_CPU_LIMIT,
			WebappMemoryLimit:          MEDIUM_WEBAPP_MEMORY_LIMIT,
			WebappHubMaxMemory:         MEDIUM_WEBAPP_HUB_MAX_MEMORY,
			ScanReplicas:               MEDIUM_SCAN_REPLICAS,
			ScanMemoryLimit:            MEDIUM_SCAN_MEMORY_LIMIT,
			ScanHubMaxMemory:           MEDIUM_SCAN_HUB_MAX_MEMORY,
			JobRunnerReplicas:          MEDIUM_JOBRUNNER_REPLICAS,
			JobRunnerMemoryLimit:       MEDIUM_JOBRUNNER_MEMORY_LIMIT,
			JobRunnerHubMaxMemory:      MEDIUM_JOBRUNNER_HUB_MAX_MEMORY,
			CfsslMemoryLimit:           CFSSL_MEMORY_LIMIT,
			LogstashMemoryLimit:        LOGSTASH_MEMORY_LIMIT,
			RegistrationMemoryLimit:    REGISTRATION_MEMORY_LIMIT,
			ZookeeperMemoryLimit:       ZOOKEEPER_MEMORY_LIMIT,
			AuthenticationMemoryLimit:  AUTHENTICATION_MEMORY_LIMIT,
			AuthenticationHubMaxMemory: AUTHENTICATION_HUB_MAX_MEMORY,
			DocumentationMemoryLimit:   DOCUMENTATION_MEMORY_LIMIT,
			PostgresMemoryLimit:        MEDIUM_POSTGRES_MEMORY_LIMIT,
			PostgresCpuLimit:           MEDIUM_POSTGRES_CPU_LIMIT,
		}
	case LARGE:
		return &HubContainerFlavor{
			WebserverMemoryLimit:       LARGE_WEBSERVER_MEMORY_LIMIT,
			SolrMemoryLimit:            LARGE_SOLR_MEMORY_LIMIT,
			WebappCpuLimit:             LARGE_WEBAPP_CPU_LIMIT,
			WebappMemoryLimit:          LARGE_WEBAPP_MEMORY_LIMIT,
			WebappHubMaxMemory:         LARGE_WEBAPP_HUB_MAX_MEMORY,
			ScanReplicas:               LARGE_SCAN_REPLICAS,
			ScanMemoryLimit:            LARGE_SCAN_MEMORY_LIMIT,
			ScanHubMaxMemory:           LARGE_SCAN_HUB_MAX_MEMORY,
			JobRunnerReplicas:          LARGE_JOBRUNNER_REPLICAS,
			JobRunnerMemoryLimit:       LARGE_JOBRUNNER_MEMORY_LIMIT,
			JobRunnerHubMaxMemory:      LARGE_JOBRUNNER_HUB_MAX_MEMORY,
			CfsslMemoryLimit:           CFSSL_MEMORY_LIMIT,
			LogstashMemoryLimit:        LOGSTASH_MEMORY_LIMIT,
			RegistrationMemoryLimit:    REGISTRATION_MEMORY_LIMIT,
			ZookeeperMemoryLimit:       ZOOKEEPER_MEMORY_LIMIT,
			AuthenticationMemoryLimit:  AUTHENTICATION_MEMORY_LIMIT,
			AuthenticationHubMaxMemory: AUTHENTICATION_HUB_MAX_MEMORY,
			DocumentationMemoryLimit:   DOCUMENTATION_MEMORY_LIMIT,
			PostgresMemoryLimit:        LARGE_POSTGRES_MEMORY_LIMIT,
			PostgresCpuLimit:           LARGE_POSTGRES_CPU_LIMIT,
		}
	case OPSSIGHT:
		return &HubContainerFlavor{
			WebserverMemoryLimit:       OPSSIGHT_WEBSERVER_MEMORY_LIMIT,
			SolrMemoryLimit:            OPSSIGHT_SOLR_MEMORY_LIMIT,
			WebappCpuLimit:             OPSSIGHT_WEBAPP_CPU_LIMIT,
			WebappMemoryLimit:          OPSSIGHT_WEBAPP_MEMORY_LIMIT,
			WebappHubMaxMemory:         OPSSIGHT_WEBAPP_HUB_MAX_MEMORY,
			ScanReplicas:               OPSSIGHT_SCAN_REPLICAS,
			ScanMemoryLimit:            OPSSIGHT_SCAN_MEMORY_LIMIT,
			ScanHubMaxMemory:           OPSSIGHT_SCAN_HUB_MAX_MEMORY,
			JobRunnerReplicas:          OPSSIGHT_JOBRUNNER_REPLICAS,
			JobRunnerMemoryLimit:       OPSSIGHT_JOBRUNNER_MEMORY_LIMIT,
			JobRunnerHubMaxMemory:      OPSSIGHT_JOBRUNNER_HUB_MAX_MEMORY,
			CfsslMemoryLimit:           CFSSL_MEMORY_LIMIT,
			LogstashMemoryLimit:        LOGSTASH_MEMORY_LIMIT,
			RegistrationMemoryLimit:    REGISTRATION_MEMORY_LIMIT,
			ZookeeperMemoryLimit:       ZOOKEEPER_MEMORY_LIMIT,
			AuthenticationMemoryLimit:  AUTHENTICATION_MEMORY_LIMIT,
			AuthenticationHubMaxMemory: AUTHENTICATION_HUB_MAX_MEMORY,
			DocumentationMemoryLimit:   DOCUMENTATION_MEMORY_LIMIT,
			PostgresMemoryLimit:        OPSSIGHT_POSTGRES_MEMORY_LIMIT,
			PostgresCpuLimit:           OPSSIGHT_POSTGRES_CPU_LIMIT,
		}
	default:
		return nil
	}
}
