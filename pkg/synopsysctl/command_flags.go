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

package synopsysctl

import (
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
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

// Create Blackduck Command Defaults
var defaultBlackduckSpec = &blackduckv1.BlackduckSpec{}
var defaultBlackduckName = "blackduck"

var create_blackduck_size = ""
var create_blackduck_dbPrototype = ""
var create_blackduck_externalPostgres_postgresHost = ""
var create_blackduck_externalPostgres_postgresPort = 0
var create_blackduck_externalPostgres_postgresAdmin = ""
var create_blackduck_externalPostgres_postgresUser = ""
var create_blackduck_externalPostgres_postgresSsl = false
var create_blackduck_externalPostgres_postgresAdminPassword = ""
var create_blackduck_externalPostgres_postgresUserPassword = ""
var create_blackduck_pvcStorageClass = ""
var create_blackduck_livenessProbes = false
var create_blackduck_scanType = ""
var create_blackduck_persistentStorage = false
var create_blackduck_PVC_json_slice = []string{}
var create_blackduck_certificateName = ""
var create_blackduck_certificate = ""
var create_blackduck_certificateKey = ""
var create_blackduck_proxyCertificate = ""
var create_blackduck_type = ""
var create_blackduck_desiredState = ""
var create_blackduck_environs = []string{}
var create_blackduck_imageRegistries = []string{}
var create_blackduck_imageUIDMap_json_slice = []string{}
var create_blackduck_licenseKey = ""

// Create OpsSight Command Defaults
var defaultOpsSightSpec = &opssightv1.OpsSightSpec{}
var defaultOpsSightName = "opssight"

var create_opssight_perceptor_name = ""
var create_opssight_perceptor_image = ""
var create_opssight_perceptor_port = 0
var create_opssight_perceptor_checkForStalledScansPauseHours = 0
var create_opssight_perceptor_stalledScanClientTimeoutHours = 0
var create_opssight_perceptor_modelMetricsPauseSeconds = 0
var create_opssight_perceptor_unknownImagePauseMilliseconds = 0
var create_opssight_perceptor_clientTimeoutMilliseconds = 0
var create_opssight_scannerPod_name = ""
var create_opssight_scannerPod_scanner_name = ""
var create_opssight_scannerPod_scanner_image = ""
var create_opssight_scannerPod_scanner_port = 0
var create_opssight_scannerPod_scanner_clientTimeoutSeconds = 0
var create_opssight_scannerPod_imageFacade_name = ""
var create_opssight_scannerPod_imageFacade_image = ""
var create_opssight_scannerPod_imageFacade_port = 0
var create_opssight_scannerPod_imageFacade_internalRegistries_json_slice = []string{}
var create_opssight_scannerPod_imageFacade_imagePullerType = ""
var create_opssight_scannerPod_imageFacade_serviceAccount = ""
var create_opssight_scannerPod_replicaCount = 0
var create_opssight_scannerPod_imageDirectory = ""
var create_opssight_perceiver_enableImagePerceiver = false
var create_opssight_perceiver_enablePodPerceiver = false
var create_opssight_perceiver_imagePerceiver_name = ""
var create_opssight_perceiver_imagePerceiver_image = ""
var create_opssight_perceiver_podPerceiver_name = ""
var create_opssight_perceiver_podPerceiver_image = ""
var create_opssight_perceiver_podPerceiver_namespaceFilter = ""
var create_opssight_perceiver_annotationIntervalSeconds = 0
var create_opssight_perceiver_dumpIntervalMinutes = 0
var create_opssight_perceiver_serviceAccount = ""
var create_opssight_perceiver_port = 0
var create_opssight_prometheus_name = ""
var create_opssight_prometheus_image = ""
var create_opssight_prometheus_port = 0
var create_opssight_enableSkyfire = false
var create_opssight_skyfire_name = ""
var create_opssight_skyfire_image = ""
var create_opssight_skyfire_port = 0
var create_opssight_skyfire_prometheusPort = 0
var create_opssight_skyfire_serviceAccount = ""
var create_opssight_skyfire_hubClientTimeoutSeconds = 0
var create_opssight_skyfire_hubDumpPauseSeconds = 0
var create_opssight_skyfire_kubeDumpIntervalSeconds = 0
var create_opssight_skyfire_perceptorDumpIntervalSeconds = 0
var create_opssight_blackduck_hosts = []string{}
var create_opssight_blackduck_user = ""
var create_opssight_blackduck_port = 0
var create_opssight_blackduck_concurrentScanLimit = 0
var create_opssight_blackduck_totalScanLimit = 0
var create_opssight_blackduck_passwordEnvVar = ""
var create_opssight_blackduck_initialCount = 0
var create_opssight_blackduck_maxCount = 0
var create_opssight_blackduck_deleteHubThresholdPercentage = 0
var create_opssight_enableMetrics = false
var create_opssight_defaultCPU = ""
var create_opssight_defaultMem = ""
var create_opssight_logLevel = ""
var create_opssight_configMapName = ""
var create_opssight_secretName = ""

// Create Alert Command Defaults
var defaultAlertSpec = &alertv1.AlertSpec{}
var defaultAlertName = "alert"

var create_alert_registry = ""
var create_alert_imagePath = ""
var create_alert_alertImageName = ""
var create_alert_alertImageVersion = ""
var create_alert_cfsslImageName = ""
var create_alert_cfsslImageVersion = ""
var create_alert_blackduckHost = ""
var create_alert_blackduckUser = ""
var create_alert_blackduckPort = 0
var create_alert_port = 0
var create_alert_standAlone = false
var create_alert_alertMemory = ""
var create_alert_cfsslMemory = ""
var create_alert_state = ""
