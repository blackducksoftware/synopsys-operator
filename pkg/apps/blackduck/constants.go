package blackduck

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
)

// TODO a type for each component type
const (
	// Size
	BlackDuckSizeV1 types.ComponentName = "blackDuckSizeV1"

	// RC
	BlackDuckAuthenticationRCV1 types.ComponentName = "blackDuckAuthenticationRCV1"
	BlackDuckBinaryScannerRCV1  types.ComponentName = "blackDuckBinaryscannerRCV1"
	BlackDuckCfsslRCV1          types.ComponentName = "blackDuckCfsslRCV1"
	BlackDuckDocumentationRCV1  types.ComponentName = "blackDuckDocumentationRCV1"
	BlackDuckJobRunnerRCV1      types.ComponentName = "blackDuckJobRunnerRCV1"
	BlackDuckPostgresRCV1       types.ComponentName = "blackDuckPostgresRCV1"
	BlackDuckRabbitMQRCV1       types.ComponentName = "blackDuckRabbitMQRCV1"
	BlackDuckRegistrationRCV1   types.ComponentName = "blackDuckRegistrationRCV1"
	BlackDuckScanRCV1           types.ComponentName = "blackDuckScanRCV1"
	BlackDuckSolrRCV1           types.ComponentName = "blackDuckSolrRCV1"
	BlackDuckUploadCacheRCV1    types.ComponentName = "blackDuckUploadCacheRCV1"
	BlackDuckWebappLogstashRCV1 types.ComponentName = "blackDuckWebappLogstashRCV1"
	BlackDuckWebserverRCV1      types.ComponentName = "blackDuckWebserverRCV1"
	BlackDuckZookeeperRCV1      types.ComponentName = "blackDuckZookeeperRCV1"

	// Service
	BlackDuckAuthentivationServiceV1 types.ComponentName = "blackDuckAuthenticationServiceV1"
	BlackDuckCfsslServiceV1          types.ComponentName = "blackDuckCfsslServiceV1"
	BlackDuckDocumentationServiceV1  types.ComponentName = "blackDuckDocumentationServiceV1"
	BlackDuckPostgresServiceV1       types.ComponentName = "blackDuckPostgresServiceV1"
	BlackDuckRabbitMQServiceV1       types.ComponentName = "blackDuckRabbitMQServiceV1"
	BlackDuckRegistrationServiceV1   types.ComponentName = "blackDuckRegistrationServiceV1"
	BlackDuckScanServiceV1           types.ComponentName = "blackDuckScanServiceV1"
	BlackDuckSolrServiceV1           types.ComponentName = "blackDuckSolrServiceV1"
	BlackDuckUploadCacheServiceV1    types.ComponentName = "blackDuckUploadCacheServiceV1"
	BlackDuckWebappServiceV1         types.ComponentName = "blackDuckWebappServiceV1"
	BlackDuckLogstashServiceV1       types.ComponentName = "blackDuckLogstashServiceV1"
	BlackDuckWebserverServiceV1      types.ComponentName = "blackDuckWebserverServiceV1"
	BlackDuckZookeeperServiceV1      types.ComponentName = "blackDuckZookeeperServiceV1"
	BlackDuckExposeServiceV1         types.ComponentName = "blackDuckExposeServiceV1"

	// ConfigMap
	BlackDuckGlobalConfigmapV1   types.ComponentName = "blackDuckGlobalConfigmapV1"
	BlackDuckDatabaseConfigmapV1 types.ComponentName = "blackDuckDatabaseConfigmapV1"

	// Secret
	BlackDuckWebCertificateSecretV1   types.ComponentName = "blackDuckWebCertificateSecretV1"
	BlackDuckAuthCertificateSecretV1  types.ComponentName = "blackDuckAuthCertificateSecretV1"
	BlackDuckProxyCertificateSecretV1 types.ComponentName = "blackDuckProxyCertificateSecretV1"
	BlackDuckUploadCacheSecretV1      types.ComponentName = "blackDuckUploadCacheSecretV1"
	BlackDuckPostgresSecretV1         types.ComponentName = "blackDuckPostgresSecretV1"

	// PVC
	BlackDuckPVCV1 types.ComponentName = "blackDuckPVCV1"
	BlackDuckPVCV2 types.ComponentName = "blackDuckPVCV2"
)

const (
	// AuthenticationContainerName ...
	AuthenticationContainerName types.ContainerName = "authentication"
	// BinaryScannerContainerName ...
	BinaryScannerContainerName types.ContainerName = "binaryscanner"
	// CfsslContainerName ...
	CfsslContainerName types.ContainerName = "cfssl"
	// DocumentationContainerName ...
	DocumentationContainerName types.ContainerName = "documentation"
	// JobrunnerContainerName ...
	JobrunnerContainerName types.ContainerName = "jobrunner"
	// RabbitMQContainerName ...
	RabbitMQContainerName types.ContainerName = "rabbitmq"
	// RegistrationContainerName ...
	RegistrationContainerName types.ContainerName = "registration"
	// ScanContainerName ...
	ScanContainerName types.ContainerName = "scan"
	// SolrContainerName ...
	SolrContainerName types.ContainerName = "solr"
	// UploadCacheContainerName ...
	UploadCacheContainerName types.ContainerName = "uploadcache"
	// WebappContainerName ...
	WebappContainerName types.ContainerName = "webapp"
	// LogstashContainerName ...
	LogstashContainerName types.ContainerName = "logstash"
	// WebserverContainerName ...
	WebserverContainerName types.ContainerName = "webserver"
	// ZookeeperContainerName ...
	ZookeeperContainerName types.ContainerName = "zookeeper"
	// PostgresContainerName ...
	PostgresContainerName types.ContainerName = "postgres"
)
