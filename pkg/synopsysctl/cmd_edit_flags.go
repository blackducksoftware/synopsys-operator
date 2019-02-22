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

// Edit Blackduck Command Flags

// Edit Blackduck Spec Flags
var editBlackduckSize = ""
var editBlackduckDbPrototype = ""
var editBlackduckExternalPostgresPostgresHost = ""
var editBlackduckExternalPostgresPostgresPort = 0
var editBlackduckExternalPostgresPostgresAdmin = ""
var editBlackduckExternalPostgresPostgresUser = ""
var editBlackduckExternalPostgresPostgresSsl = false
var editBlackduckExternalPostgresPostgresAdminPassword = ""
var editBlackduckExternalPostgresPostgresUserPassword = ""
var editBlackduckPvcStorageClass = ""
var editBlackduckLivenessProbes = false
var editBlackduckScanType = ""
var editBlackduckPersistentStorage = false
var editBlackduckPVCJSONSlice = []string{}
var editBlackduckCertificateName = ""
var editBlackduckCertificate = ""
var editBlackduckCertificateKey = ""
var editBlackduckProxyCertificate = ""
var editBlackduckType = ""
var editBlackduckDesiredState = ""
var editBlackduckEnvirons = []string{}
var editBlackduckImageRegistries = []string{}
var editBlackduckImageUIDMapJSONSlice = []string{}
var editBlackduckLicenseKey = ""

// Edit OpsSight Command Flags

// Edit OpsSight Spec Flags
var editOpssightPerceptorName = ""
var editOpssightPerceptorImage = ""
var editOpssightPerceptorPort = 0
var editOpssightPerceptorCheckForStalledScansPauseHours = 0
var editOpssightPerceptorStalledScanClientTimeoutHours = 0
var editOpssightPerceptorModelMetricsPauseSeconds = 0
var editOpssightPerceptorUnknownImagePauseMilliseconds = 0
var editOpssightPerceptorClientTimeoutMilliseconds = 0
var editOpssightScannerPodName = ""
var editOpssightScannerPodScannerName = ""
var editOpssightScannerPodScannerImage = ""
var editOpssightScannerPodScannerPort = 0
var editOpssightScannerPodScannerClientTimeoutSeconds = 0
var editOpssightScannerPodImageFacadeName = ""
var editOpssightScannerPodImageFacadeImage = ""
var editOpssightScannerPodImageFacadePort = 0
var editOpssightScannerPodImageFacadeInternalRegistriesJSONSlice = []string{}
var editOpssightScannerPodImageFacadeImagePullerType = ""
var editOpssightScannerPodImageFacadeServiceAccount = ""
var editOpssightScannerPodReplicaCount = 0
var editOpssightScannerPodImageDirectory = ""
var editOpssightPerceiverEnableImagePerceiver = false
var editOpssightPerceiverEnablePodPerceiver = false
var editOpssightPerceiverImagePerceiverName = ""
var editOpssightPerceiverImagePerceiverImage = ""
var editOpssightPerceiverPodPerceiverName = ""
var editOpssightPerceiverPodPerceiverImage = ""
var editOpssightPerceiverPodPerceiverNamespaceFilter = ""
var editOpssightPerceiverAnnotationIntervalSeconds = 0
var editOpssightPerceiverDumpIntervalMinutes = 0
var editOpssightPerceiverServiceAccount = ""
var editOpssightPerceiverPort = 0
var editOpssightPrometheusName = ""
var editOpssightPrometheusImage = ""
var editOpssightPrometheusPort = 0
var editOpssightEnableSkyfire = false
var editOpssightSkyfireName = ""
var editOpssightSkyfireImage = ""
var editOpssightSkyfirePort = 0
var editOpssightSkyfirePrometheusPort = 0
var editOpssightSkyfireServiceAccount = ""
var editOpssightSkyfireHubClientTimeoutSeconds = 0
var editOpssightSkyfireHubDumpPauseSeconds = 0
var editOpssightSkyfireKubeDumpIntervalSeconds = 0
var editOpssightSkyfirePerceptorDumpIntervalSeconds = 0
var editOpssightBlackduckHosts = []string{}
var editOpssightBlackduckUser = ""
var editOpssightBlackduckPort = 0
var editOpssightBlackduckConcurrentScanLimit = 0
var editOpssightBlackduckTotalScanLimit = 0
var editOpssightBlackduckPasswordEnvVar = ""
var editOpssightBlackduckInitialCount = 0
var editOpssightBlackduckMaxCount = 0
var editOpssightBlackduckDeleteHubThresholdPercentage = 0
var editOpssightEnableMetrics = false
var editOpssightDefaultCPU = ""
var editOpssightDefaultMem = ""
var editOpssightLogLevel = ""
var editOpssightConfigMapName = ""
var editOpssightSecretName = ""

// Edit Alert Command Flags

// Edit Alert Spec Flags
var editAlertRegistry = ""
var editAlertImagePath = ""
var editAlertAlertImageName = ""
var editAlertAlertImageVersion = ""
var editAlertCfsslImageName = ""
var editAlertCfsslImageVersion = ""
var editAlertBlackduckHost = ""
var editAlertBlackduckUser = ""
var editAlertBlackduckPort = 0
var editAlertPort = 0
var editAlertStandAlone = false
var editAlertAlertMemory = ""
var editAlertCfsslMemory = ""
var editAlertState = ""
