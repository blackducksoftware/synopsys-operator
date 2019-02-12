// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
)

// synopsysctl Defaults
var namespace = ""

// Start Command Defaults
var start_synopsysOperatorImage = "docker.io/blackducksoftware/synopsys-operator:2019.2.0-RC"
var start_prometheusImage = "docker.io/prom/prometheus:v2.1.0"
var start_blackduckRegistrationKey = ""
var start_dockerConfigPath = ""

var start_secretName = "blackduck-secret"
var start_secretType = "Opaque"
var start_secretAdminPassword = "YmxhY2tkdWNr"
var start_secretPostgresPassword = "YmxhY2tkdWNr"
var start_secretUserPassword = "YmxhY2tkdWNr"
var start_secretBlackduckPassword = "YmxhY2tkdWNr"

// Create Blackduck Defaults
var create_blackduck_size = "small"
var create_blackduck_dbPrototype = ""
var create_blackduck_externalPostgres = &blackduckv1.PostgresExternalDBConfig{}

var create_blackduck_externalPostgres_postgresHost = ""
var create_blackduck_externalPostgres_postgresPort = 0
var create_blackduck_externalPostgres_postgresAdmin = ""
var create_blackduck_externalPostgres_postgresUser = ""
var create_blackduck_externalPostgres_postgresSsl = false
var create_blackduck_externalPostgres_postgresAdminPassword = ""
var create_blackduck_externalPostgres_postgresUserPassword = ""

var create_blackduck_pvcStorageClass = "standard"
var create_blackduck_livenessProbes = false
var create_blackduck_scanType = ""
var create_blackduck_persistentStorage = true
var create_blackduck_PVC = []blackduckv1.PVC{
	blackduckv1.PVC{
		Name: "blackduck-postgres",
		Size: "200Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-authentication",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-cfssl",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-registration",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-solr",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-webapp",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-logstash",
		Size: "20Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-zookeeper-data",
		Size: "2Gi",
	},
	blackduckv1.PVC{
		Name: "blackduck-zookeeper-datalog",
		Size: "2Gi",
	},
}
var create_blackduck_PVC_json_slice = []string{
	"{\"name\": \"blackduck-postgres\",\"size\": \"200Gi\"}",
	"{\"name\": \"blackduck-authentication\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-cfssl\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-registration\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-solr\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-webapp\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-logstash\",\"size\": \"20Gi\"}",
	"{\"name\": \"blackduck-zookeeper-data\",\"size\": \"2Gi\"}",
	"{\"name\": \"blackduck-zookeeper-datalog\",\"size\": \"2Gi\"}",
}

var create_blackduck_PVC_json = "[{\"name\": \"blackduck-postgres\",\"size\": \"200Gi\"},{\"name\": \"blackduck-authentication\",\"size\": \"2Gi\"},{\"name\": \"blackduck-cfssl\",\"size\": \"2Gi\"},{\"name\": \"blackduck-registration\",\"size\": \"2Gi\"},{\"name\": \"blackduck-solr\",\"size\": \"2Gi\"},{\"name\": \"blackduck-webapp\",\"size\": \"2Gi\"},{\"name\": \"blackduck-logstash\",\"size\": \"20Gi\"},{\"name\": \"blackduck-zookeeper-data\",\"size\": \"2Gi\"},{\"name\": \"blackduck-zookeeper-datalog\",\"size\": \"2Gi\"]"

var create_blackduck_certificateName = "default"
var create_blackduck_certificate = ""
var create_blackduck_certificateKey = ""
var create_blackduck_proxyCertificate = ""
var create_blackduck_type = "worker"
var create_blackduck_desiredState = ""
var create_blackduck_environs = []string{
	"HTTPS_VERIFY_CERTS:yes",
	"HUB_POSTGRES_ENABLE_SSL:false",
	"HUB_VERSION:2018.12.2",
	"IPV4_ONLY:0",
	"RABBITMQ_DEFAULT_VHOST:protecodesc",
	"USE_ALERT:0",
	"CFSSL:cfssl:8888",
	"PUBLIC_HUB_WEBSERVER_HOST:localhost",
	"RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT:false",
	"HUB_POSTGRES_ADMIN:blackduck",
	"HUB_PROXY_NON_PROXY_HOSTS:solr",
	"PUBLIC_HUB_WEBSERVER_PORT:443",
	"DISABLE_HUB_DASHBOARD:#hub-webserver.env",
	"HUB_LOGSTASH_HOST:logstash",
	"RABBIT_MQ_PORT:5671",
	"USE_BINARY_UPLOADS:1",
	"BROKER_USE_SSL:yes",
	"RABBIT_MQ_HOST:rabbitmq",
	"CLIENT_CERT_CN:binaryscanner",
	"HUB_POSTGRES_USER:blackduck_user",
	"BLACKDUCK_REPORT_IGNORED_COMPONENTS:false",
	"BROKER_URL:amqps://rabbitmq/protecodesc",
	"SCANNER_CONCURRENCY:1",
	"HUB_WEBSERVER_PORT:8443",
}
var create_blackduck_imageRegistries = []string{
	"docker.io/blackducksoftware/blackduck-authentication:2018.12.2",
	"docker.io/blackducksoftware/blackduck-documentation:2018.12.2",
	"docker.io/blackducksoftware/blackduck-jobrunner:2018.12.2",
	"docker.io/blackducksoftware/blackduck-registration:2018.12.2",
	"docker.io/blackducksoftware/blackduck-scan:2018.12.2",
	"docker.io/blackducksoftware/blackduck-webapp:2018.12.2",
	"docker.io/blackducksoftware/blackduck-cfssl:1.0.0",
	"docker.io/blackducksoftware/blackduck-logstash:1.0.2",
	"docker.io/blackducksoftware/blackduck-nginx:1.0.0",
	"docker.io/blackducksoftware/blackduck-solr:1.0.0",
	"docker.io/blackducksoftware/blackduck-zookeeper:1.0.0",
	"docker.io/blackducksoftware/appcheck-worker:1.0.1",
	"docker.io/blackducksoftware/rabbitmq:1.0.0",
	"docker.io/blackducksoftware/blackduck-upload-cache:1.0.3",
}
var create_blackduck_imageUIDMap = map[string]int64{}
var create_blackduck_licenseKey = ""

// Create OpsSight Defaults
var create_opssight_perceptor = &opssightv1.Perceptor{}
var create_opssight_scannerPod = &opssightv1.ScannerPod{}
var create_opssight_perceiver = &opssightv1.Perceiver{}
var create_opssight_prometheus = &opssightv1.Prometheus{}
var create_opssight_enableSkyfire = false
var create_opssight_skyfire = &opssightv1.Skyfire{}
var create_opssight_blackduck = &blackduckv1.Blackduck{}
var create_opssight_enableMetrics = false
var create_opssight_defaultCPU = ""
var create_opssight_defaultMem = ""
var create_opssight_logLevel = ""
var create_opssight_configMapName = ""
var create_opssight_secretName = ""

// Create Alert Defaults
var create_alert_registry = "docker.io"
var create_alert_imagePath = "blackducksoftware"
var create_alert_alertImageName = "blackduck-alert"
var create_alert_alertImageVersion = "2.1.0"
var create_alert_cfsslImageName = "hub-cfssl"
var create_alert_cfsslImageVersion = "4.8.1"
var create_alert_blackduckHost = "sysadmin"
var create_alert_blackduckUser = ""
var create_alert_blackduckPort = 443
var create_alert_port = 8443
var create_alert_standAlone = true
var create_alert_alertMemory = "512M"
var create_alert_cfsslMemory = "640M"
var create_alert_state = ""
