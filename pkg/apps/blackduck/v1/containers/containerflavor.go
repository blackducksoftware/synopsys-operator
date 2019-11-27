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

package containers

import (
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// Flavor will determine the size of the Black Duck Blackduck
type Flavor string

const (
	// SMALL Black Duck Blackduck
	SMALL Flavor = "SMALL"
	// MEDIUM Black Duck Blackduck
	MEDIUM Flavor = "MEDIUM"
	// LARGE Black Duck Blackduck
	LARGE Flavor = "LARGE"
	// XLARGE Black Duck Blackduck
	XLARGE Flavor = "X-LARGE"
)

// ContainerFlavor configuration will have the settings for flavored Black Duck Blackduck
type ContainerFlavor struct {
	WebserverMemoryLimit       string
	SolrMemoryLimit            string
	WebappCPULimit             string
	WebappMemoryLimit          string
	WebappHubMaxMemory         string
	ScanReplicas               *int32
	ScanMemoryLimit            string
	ScanHubMaxMemory           string
	JobRunnerReplicas          *int32
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
	PostgresCPULimit           string
	BinaryScannerMemoryLimit   string
	RabbitmqMemoryLimit        string
	UploadCacheMemoryLimit     string
}

// GetContainersFlavor will return the default settings for the flavored Black Duck Blackduck TODO Make this typesafe, make flavor const into an enum.
func GetContainersFlavor(flavor string) *ContainerFlavor {
	switch Flavor(strings.ToUpper(flavor)) {
	case SMALL:
		return &ContainerFlavor{
			WebserverMemoryLimit:       smallWebServerMemoryRequestsAndLimits,
			SolrMemoryLimit:            smallSolrMemoryRequestsAndLimits,
			WebappCPULimit:             smallWebappLogstashCpuRequests,
			WebappMemoryLimit:          smallWebappLogstashMemoryRequestsAndLimits,
			WebappHubMaxMemory:         smallWebappLogstashHubMaxMemoryEnvVar,
			ScanReplicas:               util.IntToInt32(smallScanReplicas),
			ScanMemoryLimit:            smallScanMemoryRequestsAndLimits,
			ScanHubMaxMemory:           smallScanHubMaxMemoryEnvVar,
			JobRunnerReplicas:          util.IntToInt32(smallJobRunnerReplicas),
			JobRunnerMemoryLimit:       smallJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      smallJobRunnerHubMaxMemoryEnvVar,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           smallPostgresCpuRequests,
			PostgresMemoryLimit:        smallPostgresMemoryRequestsAndLimits,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case MEDIUM:
		return &ContainerFlavor{
			WebserverMemoryLimit:       mediumWebServerMemoryRequestsAndLimits,
			SolrMemoryLimit:            mediumSolrMemoryRequestsAndLimits,
			WebappCPULimit:             mediumWebappLogstashCpuRequests,
			WebappMemoryLimit:          mediumWebappLogstashMemoryRequestsAndLimits,
			WebappHubMaxMemory:         mediumWebappLogstashHubMaxMemoryEnvVar,
			ScanReplicas:               util.IntToInt32(mediumScanReplicas),
			ScanMemoryLimit:            mediumScanMemoryRequestsAndLimits,
			ScanHubMaxMemory:           mediumScanHubMaxMemoryEnvVar,
			JobRunnerReplicas:          util.IntToInt32(mediumJobRunnerReplicas),
			JobRunnerMemoryLimit:       mediumJobRunnerMemoryRequestsAndLimits,
			JobRunnerHubMaxMemory:      mediumJobRunnerHubMaxMemoryEnvVar,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           mediumPostgresCpuRequests,
			PostgresMemoryLimit:        mediumPostgresMemoryRequestsAndLimits,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case LARGE:
		return &ContainerFlavor{
			WebserverMemoryLimit:       largeWebServerMemoryRequestsAndLimits,
			SolrMemoryLimit:            largeSolrMemoryRequestsAndLimits,
			WebappCPULimit:             largeWebappLogstashCpuRequests,
			WebappMemoryLimit:          largeWebappLogstashMemoryRequestsAndLimits,
			WebappHubMaxMemory:         largeWebappLogstashHubMaxMemoryEnvVar,
			ScanReplicas:               util.IntToInt32(largeScanReplicas),
			ScanMemoryLimit:            largeScanMemoryRequestsAndLimits,
			ScanHubMaxMemory:           largeScanHubMaxMemoryEnvVar,
			JobRunnerReplicas:          util.IntToInt32(largeJobRunnerReplicas),
			JobRunnerMemoryLimit:       largeJobRunnerMemoryRequestsAndLimits,
			JobRunnerHubMaxMemory:      largeJobRunnerHubMaxMemoryEnvVar,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           largePostgresCpuRequests,
			PostgresMemoryLimit:        largePostgresMemoryRequestsAndLimits,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case XLARGE:
		return &ContainerFlavor{
			WebserverMemoryLimit:       xLargeWebServerMemoryRequestsAndLimits,
			SolrMemoryLimit:            xLargeSolrMemoryRequestsAndLimits,
			WebappCPULimit:             xLargeWebappLogstashCpuRequests,
			WebappMemoryLimit:          xLargeWebappLogstashMemoryRequestsAndLimits,
			WebappHubMaxMemory:         xLargeWebappLogstashHubMaxMemoryEnvVar,
			ScanReplicas:               util.IntToInt32(xLargeScanReplicas),
			ScanMemoryLimit:            xLargeScanMemoryRequestsAndLimits,
			ScanHubMaxMemory:           xLargeScanHubMaxMemoryEnvVar,
			JobRunnerReplicas:          util.IntToInt32(xLargeJobRunnerReplicas),
			JobRunnerMemoryLimit:       xLargeJobRunnerMemoryRequestsAndLimits,
			JobRunnerHubMaxMemory:      xLargeJobRunnerHubMaxMemoryEnvVar,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           xLargePostgresCpuRequests,
			PostgresMemoryLimit:        xLargePostgresMemoryRequestsAndLimits,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	default:
		return nil
	}
}
