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
			WebserverMemoryLimit:       smallWebServerMemoryLimit,
			SolrMemoryLimit:            smallSolrMemoryLimit,
			WebappCPULimit:             smallWebappCPULimit,
			WebappMemoryLimit:          smallWebappMemoryLimit,
			WebappHubMaxMemory:         smallWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(smallScanReplicas),
			ScanMemoryLimit:            smallScanMemoryLimit,
			ScanHubMaxMemory:           smallScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(smallJobRunnerReplicas),
			JobRunnerMemoryLimit:       smallJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      smallJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           smallPostgresCPULimit,
			PostgresMemoryLimit:        smallPostgresMemoryLimit,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case MEDIUM:
		return &ContainerFlavor{
			WebserverMemoryLimit:       mediumWebServerMemoryLimit,
			SolrMemoryLimit:            mediumSolrMemoryLimit,
			WebappCPULimit:             mediumWebappCPULimit,
			WebappMemoryLimit:          mediumWebappMemoryLimit,
			WebappHubMaxMemory:         mediumWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(mediumScanReplicas),
			ScanMemoryLimit:            mediumScanMemoryLimit,
			ScanHubMaxMemory:           mediumScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(mediumJobRunnerReplicas),
			JobRunnerMemoryLimit:       mediumJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      mediumJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           mediumPostgresCPULimit,
			PostgresMemoryLimit:        mediumPostgresMemoryLimit,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case LARGE:
		return &ContainerFlavor{
			WebserverMemoryLimit:       largeWebServerMemoryLimit,
			SolrMemoryLimit:            largeSolrMemoryLimit,
			WebappCPULimit:             largeWebappCPULimit,
			WebappMemoryLimit:          largeWebappMemoryLimit,
			WebappHubMaxMemory:         largeWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(largeScanReplicas),
			ScanMemoryLimit:            largeScanMemoryLimit,
			ScanHubMaxMemory:           largeScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(largeJobRunnerReplicas),
			JobRunnerMemoryLimit:       largeJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      largeJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           largePostgresCPULimit,
			PostgresMemoryLimit:        largePostgresMemoryLimit,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	case XLARGE:
		return &ContainerFlavor{
			WebserverMemoryLimit:       xLargeWebServerMemoryLimit,
			SolrMemoryLimit:            xLargeSolrMemoryLimit,
			WebappCPULimit:             xLargeWebappCPULimit,
			WebappMemoryLimit:          xLargeWebappMemoryLimit,
			WebappHubMaxMemory:         xLargeWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(xLargeScanReplicas),
			ScanMemoryLimit:            xLargeScanMemoryLimit,
			ScanHubMaxMemory:           xLargeScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(xLargeJobRunnerReplicas),
			JobRunnerMemoryLimit:       xLargeJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      xLargeJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           xLargePostgresCPULimit,
			PostgresMemoryLimit:        xLargePostgresMemoryLimit,
			BinaryScannerMemoryLimit:   binaryScannerMemoryLimit,
			RabbitmqMemoryLimit:        rabbitmqMemoryLimit,
			UploadCacheMemoryLimit:     uploadCacheMemoryLimit,
		}
	default:
		return nil
	}
}
