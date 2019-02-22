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

// Blackduck Spec Flags
var blackduckSize = ""
var blackduckDbPrototype = ""
var blackduckExternalPostgresPostgresHost = ""
var blackduckExternalPostgresPostgresPort = 0
var blackduckExternalPostgresPostgresAdmin = ""
var blackduckExternalPostgresPostgresUser = ""
var blackduckExternalPostgresPostgresSsl = false
var blackduckExternalPostgresPostgresAdminPassword = ""
var blackduckExternalPostgresPostgresUserPassword = ""
var blackduckPvcStorageClass = ""
var blackduckLivenessProbes = false
var blackduckScanType = ""
var blackduckPersistentStorage = false
var blackduckPVCJSONSlice = []string{}
var blackduckCertificateName = ""
var blackduckCertificate = ""
var blackduckCertificateKey = ""
var blackduckProxyCertificate = ""
var blackduckType = ""
var blackduckDesiredState = ""
var blackduckEnvirons = []string{}
var blackduckImageRegistries = []string{}
var blackduckImageUIDMapJSONSlice = []string{}
var blackduckLicenseKey = ""

// OpsSight Spec Flags
var opssightPerceptorName = ""
var opssightPerceptorImage = ""
var opssightPerceptorPort = 0
var opssightPerceptorCheckForStalledScansPauseHours = 0
var opssightPerceptorStalledScanClientTimeoutHours = 0
var opssightPerceptorModelMetricsPauseSeconds = 0
var opssightPerceptorUnknownImagePauseMilliseconds = 0
var opssightPerceptorClientTimeoutMilliseconds = 0
var opssightScannerPodName = ""
var opssightScannerPodScannerName = ""
var opssightScannerPodScannerImage = ""
var opssightScannerPodScannerPort = 0
var opssightScannerPodScannerClientTimeoutSeconds = 0
var opssightScannerPodImageFacadeName = ""
var opssightScannerPodImageFacadeImage = ""
var opssightScannerPodImageFacadePort = 0
var opssightScannerPodImageFacadeInternalRegistriesJSONSlice = []string{}
var opssightScannerPodImageFacadeImagePullerType = ""
var opssightScannerPodImageFacadeServiceAccount = ""
var opssightScannerPodReplicaCount = 0
var opssightScannerPodImageDirectory = ""
var opssightPerceiverEnableImagePerceiver = false
var opssightPerceiverEnablePodPerceiver = false
var opssightPerceiverImagePerceiverName = ""
var opssightPerceiverImagePerceiverImage = ""
var opssightPerceiverPodPerceiverName = ""
var opssightPerceiverPodPerceiverImage = ""
var opssightPerceiverPodPerceiverNamespaceFilter = ""
var opssightPerceiverAnnotationIntervalSeconds = 0
var opssightPerceiverDumpIntervalMinutes = 0
var opssightPerceiverServiceAccount = ""
var opssightPerceiverPort = 0
var opssightPrometheusName = ""
var opssightPrometheusImage = ""
var opssightPrometheusPort = 0
var opssightEnableSkyfire = false
var opssightSkyfireName = ""
var opssightSkyfireImage = ""
var opssightSkyfirePort = 0
var opssightSkyfirePrometheusPort = 0
var opssightSkyfireServiceAccount = ""
var opssightSkyfireHubClientTimeoutSeconds = 0
var opssightSkyfireHubDumpPauseSeconds = 0
var opssightSkyfireKubeDumpIntervalSeconds = 0
var opssightSkyfirePerceptorDumpIntervalSeconds = 0
var opssightBlackduckHosts = []string{}
var opssightBlackduckUser = ""
var opssightBlackduckPort = 0
var opssightBlackduckConcurrentScanLimit = 0
var opssightBlackduckTotalScanLimit = 0
var opssightBlackduckPasswordEnvVar = ""
var opssightBlackduckInitialCount = 0
var opssightBlackduckMaxCount = 0
var opssightBlackduckDeleteHubThresholdPercentage = 0
var opssightEnableMetrics = false
var opssightDefaultCPU = ""
var opssightDefaultMem = ""
var opssightLogLevel = ""
var opssightConfigMapName = ""
var opssightSecretName = ""

// Create Alert Spec Flags
var alertRegistry = ""
var alertImagePath = ""
var alertAlertImageName = ""
var alertAlertImageVersion = ""
var alertCfsslImageName = ""
var alertCfsslImageVersion = ""
var alertBlackduckHost = ""
var alertBlackduckUser = ""
var alertBlackduckPort = 0
var alertPort = 0
var alertStandAlone = false
var alertAlertMemory = ""
var alertCfsslMemory = ""
var alertState = ""
