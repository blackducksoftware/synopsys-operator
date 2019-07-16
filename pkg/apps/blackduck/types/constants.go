package types

type ComponentName string

// TODO a type for each component type
const (
	// Size
	SizeV1 ComponentName = "sizeV1"

	// RC
	RcAuthenticationV1 ComponentName = "rcauthenticationv1"
	RcBinaryScannerV1  ComponentName = "rcbinaryscannerv1"
	RcCfsslV1          ComponentName = "rccfsslv1"
	RcDocumentationV1  ComponentName = "rcdocumentationv1"
	RcJobrunnerV1      ComponentName = "rcjobrunnerv1"
	RcPostgresV1       ComponentName = "rcpostgresv1"
	RcRabbitmqV1       ComponentName = "rcrabbitmqv1"
	RcRegistrationV1   ComponentName = "rcregistrationv1"
	RcScanV1           ComponentName = "rcscanv1"
	RcSolrV1           ComponentName = "rcsolrv1"
	RcUploadCacheV1    ComponentName = "rcuploadcachev1"
	RcWebappLogstashV1 ComponentName = "rcwebapplogstashv1"
	RcWebserverV1      ComponentName = "rcwebserverv1"
	RcZookeeperV1      ComponentName = "rczookeeperv1"

	// Service
	ServiceAuthentivationV1 ComponentName = "serviceAuthenticationV1"
	ServiceCfsslV1          ComponentName = "serviceCfsslV1"
	ServiceDocumentationV1  ComponentName = "serviceDocumentationV1"
	ServicePostgresV1       ComponentName = "servicePostgresV1"
	ServiceRabbitMQV1       ComponentName = "serviceRabbitMQV1"
	ServiceRegistrationV1   ComponentName = "serviceRegistrationV1"
	ServiceScanV1           ComponentName = "serviceScanV1"
	ServiceSolrV1           ComponentName = "serviceSolrV1"
	ServiceUploadCacheV1    ComponentName = "serviceUploadCacheV1"
	ServiceWebappV1         ComponentName = "serviceWebappV1"
	ServiceLogstashV1       ComponentName = "serviceLogstashV1"
	ServiceWebserverV1      ComponentName = "serviceWebserverV1"
	ServiceZookeeperV1      ComponentName = "serviceZookeeperV1"
	ServiceExposeV1         ComponentName = "serviceExposeV1"

	// ConfigMap
	GlobalConfigmapV1   ComponentName = "globalConfigmapV1"
	DatabaseConfigmapV1 ComponentName = "databaseConfigmapV1"

	// Secret
	SecretWebCertificateV1   ComponentName = "secretWebCertificateV1"
	SecretAuthCertificateV1  ComponentName = "secretAuthCertificateV1"
	SecretProxyCertificateV1 ComponentName = "secretProxyCertificateV1"
	SecretUploadCacheV1      ComponentName = "secretUploadCacheV1"
	SecretPostgresV1         ComponentName = "secretPostgresV1"
)

type ContainerName string

const (
	AuthenticationContainerName ContainerName = "authentication"
	BinaryScannerContainerName  ContainerName = "binaryscanner"
	CfsslContainerName          ContainerName = "cfssl"
	DocumentationContainerName  ContainerName = "documentation"
	JobrunnerContainerName      ContainerName = "jobrunner"
	RabbitMQContainerName       ContainerName = "rabbitmq"
	RegistrationContainerName   ContainerName = "registration"
	ScanContainerName           ContainerName = "scan"
	SolrContainerName           ContainerName = "solr"
	UploadCacheContainerName    ContainerName = "uploadcache"
	WebappContainerName         ContainerName = "webapp"
	LogstashContainerName       ContainerName = "logstash"
	WebserverContainerName      ContainerName = "webserver"
	ZookeeperContainerName      ContainerName = "zookeeper"
	PostgresContainerName       ContainerName = "postgres"
)
