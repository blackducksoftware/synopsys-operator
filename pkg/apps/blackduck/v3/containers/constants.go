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

	// SOURCE: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// cpu: 1000m means 1 CPU
	// mem: Mi means (2**20)
	// hubMaxMemory: this is the environment variable HUB_MAX_MEMORY passed to the container and eventually passed to JVM

	// [FIX ME]: NONE OF THIS IS TESTED/INSTRUMENTED/PERFORMANCE VERIFIED, BASICALLY SOME DUDES JUST BALL PARKED SOME NUMBERS

	// SHARED VALUES
	cfsslMemoryLimit           = "640Mi"
	logstashMemoryLimit        = "1Gi"
	registrationMemoryLimit    = "1Gi"
	zookeeperMemoryLimit       = "640Mi"
	authenticationMemoryLimit  = "1024Mi"
	authenticationHubMaxMemory = "512m"
	documentationMemoryLimit   = "512Mi"
	binaryScannerMemoryLimit   = "2048Mi"
	rabbitmqMemoryLimit        = "1024Mi"
	uploadCacheMemoryLimit     = "512Mi"

	registrationMinCPUUsage  = "1000m"
	zookeeperMinCPUUsage     = "1000m"
	jobRunnerMinCPUUsage     = "1000m"
	jobRunnerMaxCPUUsage     = "1000m"
	binaryScannerMinCPUUsage = "1000m"
	binaryScannerMaxCPUUsage = "1000m"

	// Ports
	cfsslPort          = int32(8888)
	webserverPort      = int32(8443)
	documentationPort  = int32(8443)
	solrPort           = int32(8983)
	registrationPort   = int32(8443)
	zookeeperPort      = int32(2181)
	jobRunnerPort      = int32(3001)
	scannerPort        = int32(8443)
	authenticationPort = int32(8443)
	webappPort         = int32(8443)
	logstashPort       = int32(5044)
	// PostgresPort will hold the port number of Postgres
	PostgresPort      = int32(5432)
	binaryScannerPort = int32(3001)
	rabbitmqPort      = int32(5671)
	uploadCachePort1  = int32(9443)
	uploadCachePort2  = int32(9444)

	// Small Flavor
	smallWebServerMemoryRequestsAndLimits = "512Mi"

	smallSolrMemoryRequestsAndLimits = "640Mi"

	smallWebappLogstashCpuRequests             = "1000m"  // 1000m means 1 CPU
	smallWebappLogstashMemoryRequestsAndLimits = "2560Mi" // currently, we are using the same for both requests and limits; Mi means (2**20)
	smallWebappLogstashHubMaxMemoryEnvVar      = "2048m"  // this is the environment variable HUB_MAX_MEMORY passed to the container and eventually passed to JVM

	smallScanReplicas                = 1
	smallScanMemoryRequestsAndLimits = "2560Mi"
	smallScanHubMaxMemoryEnvVar      = "2048m"

	smallJobRunnerReplicas           = 1
	smallJobRunnerMemoryLimit        = "4608Mi" // currently, we are using the same for both requests and limits; Mi means (2**20)
	smallJobRunnerHubMaxMemoryEnvVar = "4096m"  // this is the environment variable HUB_MAX_MEMORY passed to the container and eventually passed to JVM

	smallPostgresCpuRequests             = "1000m"  // 1000m means 1 CPU
	smallPostgresMemoryRequestsAndLimits = "3072Mi" // currently, we are using the same for both requests and limits; Mi means (2**20)

	// Medium Flavor
	mediumWebServerMemoryRequestsAndLimits = "2048Mi"

	mediumSolrMemoryRequestsAndLimits = "1024Mi"

	mediumWebappLogstashCpuRequests             = "2000m"
	mediumWebappLogstashMemoryRequestsAndLimits = "5120Mi"
	mediumWebappLogstashHubMaxMemoryEnvVar      = "4096m"

	mediumScanReplicas                = 2
	mediumScanMemoryRequestsAndLimits = "5120Mi"
	mediumScanHubMaxMemoryEnvVar      = "4096m"

	mediumJobRunnerReplicas                = 4
	mediumJobRunnerMemoryRequestsAndLimits = "7168Mi"
	mediumJobRunnerHubMaxMemoryEnvVar      = "6144m"

	mediumPostgresCpuRequests             = "2000m"
	mediumPostgresMemoryRequestsAndLimits = "8192Mi"

	// Large Flavor
	largeWebServerMemoryRequestsAndLimits = "2048Mi"

	largeSolrMemoryRequestsAndLimits = "1024Mi"

	largeWebappLogstashCpuRequests             = "2000m"
	largeWebappLogstashMemoryRequestsAndLimits = "9728Mi"
	largeWebappLogstashHubMaxMemoryEnvVar      = "8192m"

	largeScanReplicas                = 3
	largeScanMemoryRequestsAndLimits = "9728Mi"
	largeScanHubMaxMemoryEnvVar      = "8192m"

	largeJobRunnerReplicas                = 6
	largeJobRunnerMemoryRequestsAndLimits = "13824Mi"
	largeJobRunnerHubMaxMemoryEnvVar      = "12288m"

	largePostgresCpuRequests             = "2000m"
	largePostgresMemoryRequestsAndLimits = "12288Mi"

	// XLarge Flavor
	xLargeWebServerMemoryRequestsAndLimits = "2048Mi"

	xLargeSolrMemoryRequestsAndLimits = "1024Mi"

	xLargeWebappLogstashCpuRequests             = "3000m"
	xLargeWebappLogstashMemoryRequestsAndLimits = "19728Mi"
	xLargeWebappLogstashHubMaxMemoryEnvVar      = "8192m"

	xLargeScanReplicas                = 5
	xLargeScanMemoryRequestsAndLimits = "9728Mi"
	xLargeScanHubMaxMemoryEnvVar      = "8192m"

	xLargeJobRunnerReplicas                = 10
	xLargeJobRunnerMemoryRequestsAndLimits = "13824Mi"
	xLargeJobRunnerHubMaxMemoryEnvVar      = "12288m"

	xLargePostgresCpuRequests             = "3000m"
	xLargePostgresMemoryRequestsAndLimits = "12288Mi"
)
