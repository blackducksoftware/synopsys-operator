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

package types

// Black Duck component constansts
const (
	// Size
	BlackDuckSizeV1 ComponentName = "blackDuckSizeV1"

	// RC
	BlackDuckAuthenticationRCV1 ComponentName = "blackDuckAuthenticationRCV1"
	BlackDuckBinaryScannerRCV1  ComponentName = "blackDuckBinaryscannerRCV1"
	BlackDuckCfsslRCV1          ComponentName = "blackDuckCfsslRCV1"
	BlackDuckDocumentationRCV1  ComponentName = "blackDuckDocumentationRCV1"
	BlackDuckJobRunnerRCV1      ComponentName = "blackDuckJobRunnerRCV1"
	BlackDuckPostgresRCV1       ComponentName = "blackDuckPostgresRCV1"
	BlackDuckRabbitMQRCV1       ComponentName = "blackDuckRabbitMQRCV1"
	BlackDuckRegistrationRCV1   ComponentName = "blackDuckRegistrationRCV1"
	BlackDuckScanRCV1           ComponentName = "blackDuckScanRCV1"
	BlackDuckSolrRCV1           ComponentName = "blackDuckSolrRCV1"
	BlackDuckUploadCacheRCV1    ComponentName = "blackDuckUploadCacheRCV1"
	BlackDuckWebappLogstashRCV1 ComponentName = "blackDuckWebappLogstashRCV1"
	BlackDuckWebserverRCV1      ComponentName = "blackDuckWebserverRCV1"
	BlackDuckZookeeperRCV1      ComponentName = "blackDuckZookeeperRCV1"

	// Service
	BlackDuckAuthentivationServiceV1 ComponentName = "blackDuckAuthenticationServiceV1"
	BlackDuckCfsslServiceV1          ComponentName = "blackDuckCfsslServiceV1"
	BlackDuckDocumentationServiceV1  ComponentName = "blackDuckDocumentationServiceV1"
	BlackDuckPostgresServiceV1       ComponentName = "blackDuckPostgresServiceV1"
	BlackDuckRabbitMQServiceV1       ComponentName = "blackDuckRabbitMQServiceV1"
	BlackDuckRegistrationServiceV1   ComponentName = "blackDuckRegistrationServiceV1"
	BlackDuckScanServiceV1           ComponentName = "blackDuckScanServiceV1"
	BlackDuckSolrServiceV1           ComponentName = "blackDuckSolrServiceV1"
	BlackDuckUploadCacheServiceV1    ComponentName = "blackDuckUploadCacheServiceV1"
	BlackDuckWebappServiceV1         ComponentName = "blackDuckWebappServiceV1"
	BlackDuckLogstashServiceV1       ComponentName = "blackDuckLogstashServiceV1"
	BlackDuckWebserverServiceV1      ComponentName = "blackDuckWebserverServiceV1"
	BlackDuckZookeeperServiceV1      ComponentName = "blackDuckZookeeperServiceV1"
	BlackDuckExposeServiceV1         ComponentName = "blackDuckExposeServiceV1"

	// ConfigMap
	BlackDuckGlobalConfigmapV1   ComponentName = "blackDuckGlobalConfigmapV1"
	BlackDuckDatabaseConfigmapV1 ComponentName = "blackDuckDatabaseConfigmapV1"

	// Secret
	BlackDuckWebCertificateSecretV1   ComponentName = "blackDuckWebCertificateSecretV1"
	BlackDuckAuthCertificateSecretV1  ComponentName = "blackDuckAuthCertificateSecretV1"
	BlackDuckProxyCertificateSecretV1 ComponentName = "blackDuckProxyCertificateSecretV1"
	BlackDuckUploadCacheSecretV1      ComponentName = "blackDuckUploadCacheSecretV1"
	BlackDuckPostgresSecretV1         ComponentName = "blackDuckPostgresSecretV1"

	// PVC
	BlackDuckPVCV1 ComponentName = "blackDuckPVCV1"
	BlackDuckPVCV2 ComponentName = "blackDuckPVCV2"
)

const (
	// AuthenticationContainerName ...
	AuthenticationContainerName ContainerName = "authentication"
	// BinaryScannerContainerName ...
	BinaryScannerContainerName ContainerName = "binaryscanner"
	// CfsslContainerName ...
	CfsslContainerName ContainerName = "cfssl"
	// DocumentationContainerName ...
	DocumentationContainerName ContainerName = "documentation"
	// JobrunnerContainerName ...
	JobrunnerContainerName ContainerName = "jobrunner"
	// RabbitMQContainerName ...
	RabbitMQContainerName ContainerName = "rabbitmq"
	// RegistrationContainerName ...
	RegistrationContainerName ContainerName = "registration"
	// ScanContainerName ...
	ScanContainerName ContainerName = "scan"
	// SolrContainerName ...
	SolrContainerName ContainerName = "solr"
	// UploadCacheContainerName ...
	UploadCacheContainerName ContainerName = "uploadcache"
	// WebappContainerName ...
	WebappContainerName ContainerName = "webapp"
	// LogstashContainerName ...
	LogstashContainerName ContainerName = "logstash"
	// WebserverContainerName ...
	WebserverContainerName ContainerName = "webserver"
	// ZookeeperContainerName ...
	ZookeeperContainerName ContainerName = "zookeeper"
	// PostgresContainerName ...
	PostgresContainerName ContainerName = "postgres"
)

// OpsSight component constants
const (
	// Size
	OpsSightSizeV1 ComponentName = "opsSightSizeV1"

	// Cluster Role
	OpsSightPodProcessorClusterRoleV1   ComponentName = "opsSightPodProcessorClusterRoleV1"
	OpsSightImageProcessorClusterRoleV1 ComponentName = "opsSightImageProcessorClusterRoleV1"
	OpsSightScannerClusterRoleV1        ComponentName = "OpsSightScannerClusterRoleV1"
	SkyfireClusterRoleV1                ComponentName = "skyfireClusterRoleV1"

	// Cluster Role Binding
	OpsSightPodProcessorClusterRoleBindingV1   ComponentName = "opsSightPodProcessorClusterRoleBindingV1"
	OpsSightImageProcessorClusterRoleBindingV1 ComponentName = "opsSightImageProcessorClusterRoleBindingV1"
	OpsSightScannerClusterRoleBindingV1        ComponentName = "opsSightScannerClusterRoleBindingV1"
	SkyfireClusterRoleBindingV1                ComponentName = "skyfireClusterRoleBindingV1"

	// Config Map
	OpsSightConfigMapV1        ComponentName = "opsSightConfigMapV1"
	OpsSightMetricsConfigMapV1 ComponentName = "opsSightMetricsConfigMapV1"

	// Deployment
	OpsSightMetricsDeploymentV1 ComponentName = "opsSightMetricsDeploymentV1"

	// RC
	OpsSightCoreRCV1           ComponentName = "opsSightCoreRCV1"
	OpsSightPodProcessorRCV1   ComponentName = "opsSightPodProcessorRCV1"
	OpsSightImageProcessorRCV1 ComponentName = "opsSightImageProcessorRCV1"
	OpsSightScannerRCV1        ComponentName = "opsSightScannerRCV1"
	SkyfireRCV1                ComponentName = "skyfireRCV1"

	// Route
	OpsSightCoreRouteV1    ComponentName = "opsSightCoreRouteV1"
	OpsSightMetricsRouteV1 ComponentName = "opsSightMetricsRouteV1"

	// Secret
	OpsSightSecretV1 ComponentName = "opsSightSecretV1"

	// Service
	OpsSightCoreServiceV1           ComponentName = "opsSightCoreServiceV1"
	OpsSightExposeCoreServiceV1     ComponentName = "opsSightExposeCoreServiceV1"
	OpsSightPodProcessorServiceV1   ComponentName = "opsSightPodProcessorServiceV1"
	OpsSightImageProcessorServiceV1 ComponentName = "opsSightImageProcessorServiceV1"
	OpsSightImageGetterServiceV1    ComponentName = "opsSightImageGetterServiceV1"
	OpsSightScannerServiceV1        ComponentName = "opsSightScannerServiceV1"
	OpsSightMetricsServiceV1        ComponentName = "opsSightMetricsServiceV1"
	OpsSightExposeMetricsServiceV1  ComponentName = "opsSightExposeMetricsServiceV1"
	SkyfireServiceV1                ComponentName = "skyfireServiceV1"

	// Service Account
	OpsSightPodProcessorServiceAccountV1   ComponentName = "opsSightPodProcessorServiceAccountV1"
	OpsSightImageProcessorServiceAccountV1 ComponentName = "opsSightImageProcessorServiceAccountV1"
	OpsSightScannerServiceAccountV1        ComponentName = "opsSightScannerServiceAccountV1"
	SkyfireServiceAccountV1                ComponentName = "skyfireServiceAccountV1"
)

// OpsSight container name
const (
	// Downstream
	OpsSightCoreContainerName           ContainerName = "opssight-core"
	OpsSightPodProcessorContainerName   ContainerName = "opssight-pod-processor"
	OpsSightImageProcessorContainerName ContainerName = "opssight-image-processor"
	OpsSightImageGetterContainerName    ContainerName = "opssight-image-getter"
	OpsSightScannerContainerName        ContainerName = "opssight-scanner"

	// Common between upstream and downstream
	OpsSightMetricsContainerName ContainerName = "prometheus"
	SkyfireContainerName         ContainerName = "skyfire"

	// Upstream
	PerceptorContainerName            ContainerName = "perceptor"
	PodPerceiverContainerName         ContainerName = "pod-perceiver"
	ImagePerceiverContainerName       ContainerName = "image-perceiver"
	PerceptorImageFacadeContainerName ContainerName = "perceptor-imagefacade"
	PerceptorScannerContainerName     ContainerName = "perceptor-scanner"
)
