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

const (
	// SHARED VALUES
	cfsslMemoryLimit           = "640M"
	logstashMemoryLimit        = "1G"
	registrationMemoryLimit    = "640M"
	zookeeperMemoryLimit       = "640M"
	authenticationMemoryLimit  = "1024M"
	authenticationHubMaxMemory = "512m"
	documentationMemoryLimit   = "512M"
	binaryScannerMemoryLimit   = "2048M"
	rabbitmqMemoryLimit        = "1024M"
	uploadCacheMemoryLimit     = "512M"

	registrationMinCPUUsage  = "1"
	zookeeperMinCPUUsage     = "1"
	jobRunnerMinCPUUsage     = "1"
	jobRunnerMaxCPUUsage     = "1"
	binaryScannerMinCPUUsage = "1"
	binaryScannerMaxCPUUsage = "1"

	// Ports
	cfsslPort          = "8888"
	webserverPort      = "8443"
	documentationPort  = "8443"
	solrPort           = "8983"
	registrationPort   = "8443"
	zookeeperPort      = "2181"
	jobRunnerPort      = "3001"
	scannerPort        = "8443"
	authenticationPort = "8443"
	webappPort         = "8443"
	logstashPort       = "5044"
	// PostgresPort will hold the port number of Postgres
	PostgresPort      = "5432"
	binaryScannerPort = "3001"
	rabbitmqPort      = "5671"
	uploadCachePort1  = "9443"
	uploadCachePort2  = "9444"

	// Small Flavor
	smallWebServerMemoryLimit = "512M"

	smallSolrMemoryLimit = "640M"

	smallWebappCPULimit     = "1"
	smallWebappMemoryLimit  = "2560M"
	smallWebappHubMaxMemory = "2048m"

	smallScanReplicas     = 1
	smallScanMemoryLimit  = "2560M"
	smallScanHubMaxMemory = "2048m"

	smallJobRunnerReplicas     = 1
	smallJobRunnerMemoryLimit  = "4608M"
	smallJobRunnerHubMaxMemory = "4096m"

	smallPostgresCPULimit    = "1"
	smallPostgresMemoryLimit = "3072M"

	// Medium Flavor
	mediumWebServerMemoryLimit = "2048M"

	mediumSolrMemoryLimit = "1024M"

	mediumWebappCPULimit     = "2"
	mediumWebappMemoryLimit  = "5120M"
	mediumWebappHubMaxMemory = "4096m"

	mediumScanReplicas     = 2
	mediumScanMemoryLimit  = "5120M"
	mediumScanHubMaxMemory = "4096m"

	mediumJobRunnerReplicas     = 4
	mediumJobRunnerMemoryLimit  = "7168M"
	mediumJobRunnerHubMaxMemory = "6144m"

	mediumPostgresCPULimit    = "2"
	mediumPostgresMemoryLimit = "8192M"

	// Large Flavor
	largeWebServerMemoryLimit = "2048M"

	largeSolrMemoryLimit = "1024M"

	largeWebappCPULimit     = "2"
	largeWebappMemoryLimit  = "9728M"
	largeWebappHubMaxMemory = "8192m"

	largeScanReplicas     = 3
	largeScanMemoryLimit  = "9728M"
	largeScanHubMaxMemory = "8192m"

	largeJobRunnerReplicas     = 6
	largeJobRunnerMemoryLimit  = "13824M"
	largeJobRunnerHubMaxMemory = "12288m"

	largePostgresCPULimit    = "2"
	largePostgresMemoryLimit = "12288M"

	// XLarge Flavor
	xLargeWebServerMemoryLimit = "2048M"

	xLargeSolrMemoryLimit = "1024M"

	xLargeWebappCPULimit     = "3"
	xLargeWebappMemoryLimit  = "19728M"
	xLargeWebappHubMaxMemory = "8192m"

	xLargeScanReplicas     = 5
	xLargeScanMemoryLimit  = "9728M"
	xLargeScanHubMaxMemory = "8192m"

	xLargeJobRunnerReplicas     = 10
	xLargeJobRunnerMemoryLimit  = "13824M"
	xLargeJobRunnerHubMaxMemory = "12288m"

	xLargePostgresCPULimit    = "3"
	xLargePostgresMemoryLimit = "12288M"
)
