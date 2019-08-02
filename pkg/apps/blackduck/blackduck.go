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

package blackduck

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	blackduckutil "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sizeclientset "github.com/blackducksoftware/synopsys-operator/pkg/size/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Constants for each unit of a deployment of Black Duck
const (
	CRDResources      = "BLACKDUCK"
	DatabaseResources = "DATABASE"
	PVCResources      = "PVC"
)

// Blackduck is used for the Blackduck deployment
type Blackduck struct {
	config          *protoform.Config
	kubeConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	blackduckClient *blackduckclientset.Clientset
	sizeClient      *sizeclientset.Clientset
	routeClient     *routeclient.RouteV1Client
}

var publicVersions = map[string]types.PublicVersion{
	"2018.12.0": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2018.12.0",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2018.12.0",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2018.12.0",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2018.12.0",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2018.12.0",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2018.12.0",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.0",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2018.12.1": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2018.12.1",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2018.12.1",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2018.12.1",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2018.12.1",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2018.12.1",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2018.12.1",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.0",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2018.12.2": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2018.12.2",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2018.12.2",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2018.12.2",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2018.12.2",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2018.12.2",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2018.12.2",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.0",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2018.12.3": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2018.12.3",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2018.12.3",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2018.12.3",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2018.12.3",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2018.12.3",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2018.12.3",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.0",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2018.12.4": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2018.12.4",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2018.12.4",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2018.12.4",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2018.12.4",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2018.12.4",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2018.12.4",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.0",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2019.2.0": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.2.0",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.2.0",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.2.0",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.2.0",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.2.0",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.2.0",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.2",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2019.2.1": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.2.1",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.2.1",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.2.1",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.2.1",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.2.1",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.2.1",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.2",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2019.2.2": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.2.2",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.2.2",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.2.2",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.2.2",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.2.2",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.3",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.2.2",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.2",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.2",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
			"solr": {
				Identifier: types.BlackDuckSolrRCV1,
				Container: map[types.ContainerName]string{
					types.SolrContainerName: "blackducksoftware/blackduck-solr:1.0.0",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckSolrServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV1},
	},
	"2019.4.0": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.4.0",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.4.0",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.4.0",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.4.0",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.4.0",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.4.0",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
	"2019.4.1": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.4.1",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.4.1",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.4.1",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.4.1",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.4.1",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.4.1",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
	"2019.4.2": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.4.2",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.4.2",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.4.2",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.4.2",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.4.2",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.4.2",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
	"2019.4.3": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.4.3",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.01",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.4.3",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.4.3",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.4.3",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.4.3",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.4.3",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
	"2019.6.0": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.6.0",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.03",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.6.0",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.6.0",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.6.0",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.6.0",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.6.0",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
	"2019.6.1": {
		Size: types.BlackDuckSizeV1,
		RCs: map[string]types.PublicPodResource{
			"authentication": {
				Identifier: types.BlackDuckAuthenticationRCV1,
				Container: map[types.ContainerName]string{
					types.AuthenticationContainerName: "blackducksoftware/blackduck-authentication:2019.6.1",
				},
			},
			"binaryscanner": {
				Identifier: types.BlackDuckBinaryScannerRCV1,
				Container: map[types.ContainerName]string{
					types.BinaryScannerContainerName: "blackducksoftware/appcheck-worker:2019.03",
				},
			},
			"cfssl": {
				Identifier: types.BlackDuckCfsslRCV1,
				Container: map[types.ContainerName]string{
					types.CfsslContainerName: "blackducksoftware/blackduck-cfssl:1.0.0",
				},
			},
			"documentation": {
				Identifier: types.BlackDuckDocumentationRCV1,
				Container: map[types.ContainerName]string{
					types.DocumentationContainerName: "blackducksoftware/blackduck-documentation:2019.6.1",
				},
			},
			"jobrunner": {
				Identifier: types.BlackDuckJobRunnerRCV1,
				Container: map[types.ContainerName]string{
					types.JobrunnerContainerName: "blackducksoftware/blackduck-jobrunner:2019.6.1",
				},
			},
			"rabbitmq": {
				Identifier: types.BlackDuckRabbitMQRCV1,
				Container: map[types.ContainerName]string{
					types.RabbitMQContainerName: "blackducksoftware/rabbitmq:1.0.0",
				},
			},
			"registration": {
				Identifier: types.BlackDuckRegistrationRCV1,
				Container: map[types.ContainerName]string{
					types.RegistrationContainerName: "blackducksoftware/blackduck-registration:2019.6.1",
				},
			},
			"scan": {
				Identifier: types.BlackDuckScanRCV1,
				Container: map[types.ContainerName]string{
					types.ScanContainerName: "blackducksoftware/blackduck-scan:2019.6.1",
				},
			},
			"uploadcache": {
				Identifier: types.BlackDuckUploadCacheRCV1,
				Container: map[types.ContainerName]string{
					types.UploadCacheContainerName: "blackducksoftware/blackduck-upload-cache:1.0.8",
				},
			},
			"webapp-logstash": {
				Identifier: types.BlackDuckWebappLogstashRCV1,
				Container: map[types.ContainerName]string{
					types.WebappContainerName:   "blackducksoftware/blackduck-webapp:2019.6.1",
					types.LogstashContainerName: "blackducksoftware/blackduck-logstash:1.0.4",
				},
			},
			"webserver": {
				Identifier: types.BlackDuckWebserverRCV1,
				Container: map[types.ContainerName]string{
					types.WebserverContainerName: "blackducksoftware/blackduck-nginx:1.0.7",
				},
			},
			"zookeeper": {
				Identifier: types.BlackDuckZookeeperRCV1,
				Container: map[types.ContainerName]string{
					types.ZookeeperContainerName: "blackducksoftware/blackduck-zookeeper:1.0.0",
				},
			},
			"postgres": {
				Identifier: types.BlackDuckPostgresRCV1,
				Container: map[types.ContainerName]string{
					types.PostgresContainerName: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
				},
			},
		},
		Secrets: []types.ComponentName{
			types.BlackDuckUploadCacheSecretV1,
			types.BlackDuckProxyCertificateSecretV1,
			types.BlackDuckAuthCertificateSecretV1,
			types.BlackDuckWebCertificateSecretV1,
			types.BlackDuckPostgresSecretV1,
		},
		ConfigMaps: []types.ComponentName{
			types.BlackDuckGlobalConfigmapV1,
			types.BlackDuckDatabaseConfigmapV1,
		},
		Services: []types.ComponentName{
			types.BlackDuckAuthentivationServiceV1,
			types.BlackDuckCfsslServiceV1,
			types.BlackDuckDocumentationServiceV1,
			types.BlackDuckRabbitMQServiceV1,
			types.BlackDuckRegistrationServiceV1,
			types.BlackDuckScanServiceV1,
			types.BlackDuckUploadCacheServiceV1,
			types.BlackDuckWebappServiceV1,
			types.BlackDuckLogstashServiceV1,
			types.BlackDuckWebserverServiceV1,
			types.BlackDuckZookeeperServiceV1,
			types.BlackDuckPostgresServiceV1,
			types.BlackDuckExposeServiceV1,
		},
		PVC: []types.ComponentName{types.BlackDuckPVCV2},
	},
}

// NewBlackduck will return a Blackduck
func NewBlackduck(protoformDeployer *protoform.Deployer) *Blackduck {
	kubeConfig := protoformDeployer.KubeConfig
	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	sizeClient, err := sizeclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	return &Blackduck{
		config:          protoformDeployer.Config,
		kubeConfig:      kubeConfig,
		kubeClient:      protoformDeployer.KubeClient,
		blackduckClient: blackduckClient,
		sizeClient:      sizeClient,
		routeClient:     protoformDeployer.RouteClient,
	}
}

func (b *Blackduck) ensureVersion(bd *blackduckapi.Blackduck) error {
	versions := b.Versions()
	// If the version is not provided, then we set it to be the latest
	if len(bd.Spec.Version) == 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		bd.Spec.Version = versions[0]
	} else {
		// If the verion is provided, check that it's supported
		for _, v := range versions {
			if strings.Compare(v, bd.Spec.Version) == 0 {
				return nil
			}
		}
		return fmt.Errorf("version '%s' is not supported.  Supported versions: %s", bd.Spec.Version, strings.Join(versions, ", "))
	}
	return nil
}

// Delete will be used to delete a blackduck instance
func (b *Blackduck) Delete(name string) error {
	log.Infof("deleting a %s Black Duck instance", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("invalid name to delete the Black Duck instance")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
		ns, err := util.ListNamespaces(b.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.BlackDuckName, name))
		if err != nil {
			log.Errorf("unable to list %s Black Duck instance namespaces %s due to %+v", name, namespace, err)
		}
		if len(ns.Items) > 0 {
			namespace = ns.Items[0].Name
		} else {
			return fmt.Errorf("unable to find %s Black Duck instance namespace", name)
		}
	} else {
		name = values[1]
		namespace = values[0]
	}

	// delete the Black Duck instance
	commonConfig := crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, false, namespace, "",
		&api.ComponentList{}, fmt.Sprintf("app=%s,name=%s", util.BlackDuckName, name), false)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		return fmt.Errorf("unable to delete the %s Black Duck instance in %s namespace due to %+v", name, namespace, crudErrors)
	}

	var err error
	// if cluster scope, if no other instance running in Synopsys Operator namespace, delete the namespace or delete the Synopsys labels in the namespace
	if b.config.IsClusterScoped {
		err = util.DeleteResourceNamespace(b.kubeClient, util.BlackDuckName, namespace, name, false)
	} else {
		// if namespace scope, delete the label from the namespace
		_, err = util.CheckAndUpdateNamespace(b.kubeClient, util.BlackDuckName, namespace, name, "", true)
	}
	if err != nil {
		return err
	}

	return nil
}

// Versions returns the versions that the operator supports
func (b *Blackduck) Versions() []string {
	var versions []string
	for v := range publicVersions {
		versions = append(versions, v)
	}
	return versions
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (b *Blackduck) Ensure(blackDuck *blackduckapi.Blackduck) error {
	// If the version is not specified then we set it to be the latest.
	if err := b.ensureVersion(blackDuck); err != nil {
		return err
	}

	newBlackDuck := blackDuck.DeepCopy()

	version, ok := publicVersions[blackDuck.Spec.Version]
	if !ok {
		return fmt.Errorf("version %s is not supported", blackDuck.Spec.Version)
	}

	cp, err := store.GetComponents(version, b.config, b.kubeClient, b.sizeClient, blackDuck)
	if err != nil {
		return err
	}

	if strings.EqualFold(blackDuck.Spec.DesiredState, "STOP") {
		// Save/Update the PVCs for the Black Duck
		commonConfig := crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, false, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			&api.ComponentList{PersistentVolumeClaims: cp.PersistentVolumeClaims}, fmt.Sprintf("app=%s,name=%s", util.BlackDuckName, blackDuck.Name), false)
		_, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("stop Black Duck: %+v", errors)
		}
	} else if strings.EqualFold(blackDuck.Spec.DesiredState, "DbMigrate") {
		// Save/Update the PVCs for the Black Duck
		commonConfig := crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, false, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			&api.ComponentList{PersistentVolumeClaims: cp.PersistentVolumeClaims}, fmt.Sprintf("app=%s,name=%s", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("migrate blackduck: %+v", errors)
		}

		commonConfig = crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, isPatched, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			cp, fmt.Sprintf("app=%s,name=%s,component=postgres", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update postgres component: %+v", errors)
		}
	} else {
		// Save/Update the PVCs for the Black Duck
		commonConfig := crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, false, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			&api.ComponentList{PersistentVolumeClaims: cp.PersistentVolumeClaims}, fmt.Sprintf("app=%s,name=%s,component=pvc", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update pvc: %+v", errors)
		}

		// install postgres
		commonConfig = crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, isPatched, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			cp, fmt.Sprintf("app=%s,name=%s,component=postgres", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update postgres component: %+v", errors)
		}

		//Check postgres and initialize if needed.
		if blackDuck.Spec.ExternalPostgres == nil {
			// TODO return whether we re-initialized or not
			err = b.initPostgres(blackDuck.Name, &blackDuck.Spec)
			if err != nil {
				return err
			}
		}

		// install cfssl
		commonConfig = crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, isPatched, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			cp, fmt.Sprintf("app=%s,name=%s,component in (configmap,serviceAccount,cfssl)", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update cfssl component: %+v", errors)
		}

		err = util.WaitUntilPodsAreReady(b.kubeClient, blackDuck.Spec.Namespace, fmt.Sprintf("app=%s,name=%s,component=cfssl", util.BlackDuckName, blackDuck.Name), b.config.PodWaitTimeoutSeconds)
		if err != nil {
			return fmt.Errorf("the cfssl pod is not ready: %v", err)
		}

		// deploy non postgres and uploadcache component
		commonConfig = crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, isPatched, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			cp, fmt.Sprintf("app=%s,name=%s,component notin (postgres,cfssl,configmap,serviceAccount,uploadcache,route)", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update non postgres, cfssl and uploadcache components: %+v", errors)
		}

		// add routes to component list
		if util.OPENSHIFT == strings.ToUpper(blackDuck.Spec.ExposeService) {
			cp.Routes = []*api.Route{{
				Name:               apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, ""),
				Namespace:          blackDuck.Spec.Namespace,
				Kind:               "Service",
				ServiceName:        apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, "webserver"),
				PortName:           fmt.Sprintf("port-%d", 443),
				Labels:             apputils.GetLabel("route", blackDuck.Name),
				TLSTerminationType: routev1.TLSTerminationPassthrough,
			}}
		}

		// deploy upload cache and route component
		commonConfig = crdupdater.NewCRUDComponents(b.kubeConfig, b.kubeClient, b.config.DryRun, isPatched, blackDuck.Spec.Namespace, blackDuck.Spec.Version,
			cp, fmt.Sprintf("app=%s,name=%s,component in (uploadcache,route)", util.BlackDuckName, blackDuck.Name), false)
		isPatched, errors = commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("update upload cache component: %+v", errors)
		}

		if strings.ToUpper(blackDuck.Spec.ExposeService) == util.NODEPORT {
			newBlackDuck.Status.IP, err = blackduckutil.GetNodePortIPAddress(b.kubeClient, blackDuck.Spec.Namespace, apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, "webserver-exposed"))
		} else if strings.ToUpper(blackDuck.Spec.ExposeService) == util.LOADBALANCER {
			newBlackDuck.Status.IP, err = blackduckutil.GetLoadBalancerIPAddress(b.kubeClient, blackDuck.Spec.Namespace, apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, "webserver-exposed"))
		}

		// get Route on Openshift
		if strings.ToUpper(blackDuck.Spec.ExposeService) == util.OPENSHIFT && b.routeClient != nil {
			route, err := util.GetRoute(b.routeClient, blackDuck.Spec.Namespace, apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, ""))
			if err != nil {
				log.Errorf("unable to get route %s in %s namespace due to %+v", apputils.GetResourceName(blackDuck.Name, util.BlackDuckName, ""), blackDuck.Spec.Namespace, err)
			}
			if route != nil {
				newBlackDuck.Status.IP = route.Spec.Host
			}
		}

		err = util.WaitUntilPodsAreReady(b.kubeClient, blackDuck.Spec.Namespace, fmt.Sprintf("app=%s,name=%s,component notin (postgres,cfssl)", util.BlackDuckName, blackDuck.Name), b.config.PodWaitTimeoutSeconds)
		if err != nil {
			return fmt.Errorf("the remaining pods are not ready: %v", err)
		}

		// TODO wait for webserver to be up before we register
		if len(blackDuck.Spec.LicenseKey) > 0 {
			if err := b.registerIfNeeded(blackDuck); err != nil {
				log.Infof("couldn't register blackduck %s: %v", blackDuck.Name, err)
			}
		}
	}

	if !reflect.DeepEqual(blackDuck.Status, newBlackDuck.Status) {
		bd, err := util.GetBlackDuck(b.blackduckClient, blackDuck.Spec.Namespace, blackDuck.Name)
		if err != nil {
			return err
		}
		bd.Status = newBlackDuck.Status
		if _, err := b.blackduckClient.SynopsysV1().Blackducks(blackDuck.Spec.Namespace).Update(bd); err != nil {
			return err
		}
	}

	return nil
}

// GetComponents gets the Black Duck's creater and returns the components
func (b Blackduck) GetComponents(bd *blackduckapi.Blackduck, compType string) (*api.ComponentList, error) {
	//If the version is not specified then we set it to be the latest.
	if err := b.ensureVersion(bd); err != nil {
		return nil, err
	}

	version, ok := publicVersions[bd.Spec.Version]
	if !ok {
		return nil, fmt.Errorf("version %s is not supported", bd.Spec.Version)
	}

	cp, err := store.GetComponents(version, b.config, b.kubeClient, b.sizeClient, bd)
	if err != nil {
		return nil, err
	}

	switch strings.ToUpper(compType) {
	case CRDResources:
		return cp.Filter("component notin (postgres, pvc)")
	case DatabaseResources:
		return cp.Filter("component in (postgres)")
	case PVCResources:
		return cp.Filter("component in (pvc)")
	}
	return nil, fmt.Errorf("invalid components type '%s'", compType)
}

func (b Blackduck) initPostgres(name string, bdspec *blackduckapi.BlackduckSpec) error {
	adminPassword, err := util.Base64Decode(bdspec.AdminPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode adminPassword due to: %+v", bdspec.Namespace, err)
	}
	userPassword, err := util.Base64Decode(bdspec.UserPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode userPassword due to: %+v", bdspec.Namespace, err)
	}
	postgresPassword, err := util.Base64Decode(bdspec.PostgresPassword)
	if err != nil {
		return fmt.Errorf("%v: unable to decode postgresPassword due to: %+v", bdspec.Namespace, err)
	}

	err = util.WaitUntilPodsAreReady(b.kubeClient, bdspec.Namespace, fmt.Sprintf("app=%s,name=%s,component=postgres", util.BlackDuckName, name), b.config.PodWaitTimeoutSeconds)
	if err != nil {
		return fmt.Errorf("the postgres pod is not yet ready: %v", err)
	}

	// Check if initialization is required.
	db, err := database.NewDatabase(fmt.Sprintf("%s.%s.svc.cluster.local", apputils.GetResourceName(name, util.BlackDuckName, "postgres"), bdspec.Namespace), "postgres", "postgres", postgresPassword, "postgres")
	if err != nil {
		return err
	}
	defer db.Connection.Close()

	// Wait for the DB to be up
	if !db.WaitForDatabase(10) {
		return fmt.Errorf("database %s is not accessible", bdspec.Namespace)
	}

	result, err := db.Connection.Exec("SELECT datname FROM pg_catalog.pg_database WHERE datname='bds_hub';")
	if err != nil {
		return err
	}
	nbRow, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// We initialize the DB if the bds_hub database doesn't exist
	if nbRow == 0 {
		log.Infof("postres instance %s requires to be re-initialized", bdspec.Namespace)
		if len(bdspec.DbPrototype) == 0 {
			err := InitDatabase(name, bdspec, b.config.IsClusterScoped, adminPassword, userPassword, postgresPassword)
			if err != nil {
				log.Errorf("%v: error: %+v", bdspec.Namespace, err)
				return fmt.Errorf("%v: error: %+v", bdspec.Namespace, err)
			}
		} else {
			fromNamespaces, err := util.ListNamespaces(b.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.BlackDuckName, bdspec.DbPrototype))
			if len(fromNamespaces.Items) == 0 {
				return fmt.Errorf("unable to find the %s Black Duck instance", bdspec.DbPrototype)
			}
			fromNamespace := fromNamespaces.Items[0].Name
			_, fromPw, err := blackduckutil.GetBlackDuckDBPassword(b.kubeClient, fromNamespace, bdspec.DbPrototype)
			if err != nil {
				return err
			}
			err = blackduckutil.CloneJob(b.kubeClient, fromNamespace, bdspec.DbPrototype, bdspec.Namespace, name, fromPw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b Blackduck) getPVCVolumeName(namespace string, name string) (string, error) {
	pvc, err := util.GetPVC(b.kubeClient, namespace, name)
	if err != nil {
		return "", fmt.Errorf("unable to get pvc in %s namespace because %s", namespace, err.Error())
	}

	return pvc.Spec.VolumeName, nil
}

func (b Blackduck) registerIfNeeded(bd *blackduckapi.Blackduck) error {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s.%s.svc:443/api/v1/registrations?summary=true", apputils.GetResourceName(bd.Name, util.BlackDuckName, "webserver"), bd.Spec.Namespace))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var objmap map[string]*json.RawMessage

	err = dec.Decode(&objmap)
	if err != nil {
		return err
	}

	// Check whether the registration is valid
	if val, ok := objmap["valid"]; ok {
		var r bool
		err := json.Unmarshal(*val, &r)
		if err != nil {
			return err
		}

		// We register if the registration is invalid
		if !r {
			if err := b.autoRegisterBlackDuck(bd.Name, &bd.Spec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b Blackduck) autoRegisterBlackDuck(name string, bdspec *blackduckapi.BlackduckSpec) error {
	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod, err := util.FilterPodByNamePrefixInNamespace(b.kubeClient, bdspec.Namespace, apputils.GetResourceName(name, util.BlackDuckName, "registration"))
	if err != nil {
		return err
	}

	registrationKey := bdspec.LicenseKey

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for i := 0; i < 20; i++ {
			registrationPod, err := util.GetPod(b.kubeClient, bdspec.Namespace, registrationPod.Name)
			if err != nil {
				return err
			}

			// Create the exec into Kubernetes pod request
			req := util.CreateExecContainerRequest(b.kubeClient, registrationPod, "/bin/sh")
			// Exec into the Kubernetes pod and execute the commands
			_, err = util.ExecContainer(b.kubeConfig, req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate" -k --cert /opt/blackduck/hub/hub-registration/security/blackduck_system.crt --key /opt/blackduck/hub/hub-registration/security/blackduck_system.key`, registrationKey)})

			if err == nil {
				log.Infof("blackduck %s has been registered", bdspec.Namespace)
				return nil
			}
			time.Sleep(10 * time.Second)
		}
	}
	return fmt.Errorf("unable to register the blackduck %s", bdspec.Namespace)
}

func (b Blackduck) isBinaryAnalysisEnabled(bdspec *blackduckapi.BlackduckSpec) bool {
	for _, value := range bdspec.Environs {
		if strings.Contains(value, "USE_BINARY_UPLOADS") {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				mapValue := strings.TrimSpace(values[1])
				if strings.EqualFold(mapValue, "1") {
					return true
				}
			}
			return false
		}
	}
	return false
}

// GenPVC returns the list of Black Duck PVCs
func GenPVC(blackDuck blackduckapi.Blackduck, defaultPVC map[string]string) ([]*components.PersistentVolumeClaim, error) {
	var pvcs []*components.PersistentVolumeClaim
	if blackDuck.Spec.PersistentStorage {
		pvcMap := make(map[string]blackduckapi.PVC)
		for _, claim := range blackDuck.Spec.PVC {
			pvcMap[claim.Name] = claim
		}

		for name, size := range defaultPVC {
			var claim blackduckapi.PVC

			if _, ok := pvcMap[name]; ok {
				claim = pvcMap[name]
			} else {
				claim = blackduckapi.PVC{
					Name:         name,
					Size:         size,
					StorageClass: blackDuck.Spec.PVCStorageClass,
				}
			}

			// Set the claim name to be app specific if the PVC was not created by an operator version prior to
			// 2019.6.0
			if blackDuck.Annotations["synopsys.com/created.by"] != "pre-2019.6.0" {
				claim.Name = apputils.GetResourceName(blackDuck.Name, "", name)
			}

			pvc, err := createPVC(claim, horizonapi.ReadWriteOnce, apputils.GetLabel("pvc", blackDuck.Name), blackDuck.Spec.Namespace)
			if err != nil {
				return nil, err
			}
			pvcs = append(pvcs, pvc)
		}
	}
	return pvcs, nil
}

func createPVC(claim blackduckapi.PVC, accessMode horizonapi.PVCAccessModeType, label map[string]string, namespace string) (*components.PersistentVolumeClaim, error) {
	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string
	if len(claim.StorageClass) > 0 {
		class = &claim.StorageClass
	} else {
		class = nil
	}

	var size string
	_, err := resource.ParseQuantity(claim.Size)
	if err != nil {
		return nil, err
	}
	size = claim.Size

	config := horizonapi.PVCConfig{
		Name:      claim.Name,
		Namespace: namespace,
		Size:      size,
		Class:     class,
	}

	if len(claim.VolumeName) > 0 {
		// Needed so that it doesn't use the default storage class
		var tmp = ""
		config.Class = &tmp
		config.VolumeName = claim.VolumeName
	}

	pvc, err := components.NewPersistentVolumeClaim(config)
	if err != nil {
		return nil, err
	}

	pvc.AddAccessMode(accessMode)
	pvc.AddLabels(label)

	return pvc, nil
}
