/*
Copyright (C) 2018 Synopsys, Inc.

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

package opssight

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelmValuesFromCobraFlags is a type for converting synopsysctl flags
// to Helm Chart fields and values
// args: map of helm chart field to value
type HelmValuesFromCobraFlags struct {
	args     map[string]interface{}
	flagTree FlagTree
}

// FlagTree is a set of fields needed to configure the Opssight Helm Chart
type FlagTree struct {
	// IsUpstream                                      string
	Registry          string
	RegistryNamespace string
	PullSecrets       []string
	// ImageRegistries                                 []string
	LogLevel string
	// BlackduckExternalHostsFilePath                  string
	// BlackduckConnectionsEnvironmentVaraiableName    string
	BlackduckTLSVerification string
	// BlackduckPassword                               string
	// BlackduckInitialCount                           int
	// BlackduckMaxCount                               int
	// BlackduckType                                   string
	PrometheusExpose                        string
	EnableMetrics                           string
	PerceptorExpose                         string
	PerceptorCheckForStalledScansPauseHours int
	PerceptorStalledScanClientTimeoutHours  int
	PerceptorModelMetricsPauseSeconds       int
	PerceptorUnknownImagePauseMilliseconds  int
	PerceptorClientTimeoutMilliseconds      int
	PerceiverTLSCertificatePath             string
	PerceiverTLSKeyPath                     string
	PerceiverAnnotationIntervalSeconds      int
	PerceiverDumpIntervalMinutes            int
	PerceiverEnablePodPerceiver             string
	PerceiverPodPerceiverNamespaceFilter    string
	ScannerPodScannerClientTimeoutSeconds   int
	ScannerPodReplicaCount                  int
	ScannerPodImageDirectory                string
	// ScannerPodImageFacadeInternalRegistriesFilePath string
	ScannerPodImageFacadeImagePullerType      string
	PerceiverEnableImagePerceiver             string
	PerceiverEnableQuayPerceiver              string
	PerceiverQuayExpose                       string
	PerceiverEnableArtifactoryPerceiver       string
	PerceiverEnableArtifactoryPerceiverDumper string
	PerceiverArtifactoryExpose                string
	DefaultCPU                                string
	DefaultMem                                string
	ScannerCPU                                string
	ScannerMem                                string
}

// NewHelmValuesFromCobraFlags returns an initialized HelmValuesFromCobraFlags
func NewHelmValuesFromCobraFlags() *HelmValuesFromCobraFlags {
	return &HelmValuesFromCobraFlags{
		args:     make(map[string]interface{}, 0),
		flagTree: FlagTree{},
	}
}

// GetArgs returns the map of helm chart fields to values
func (ctl *HelmValuesFromCobraFlags) GetArgs() map[string]interface{} {
	return ctl.args
}

// SetArgs set the map to values
func (ctl *HelmValuesFromCobraFlags) SetArgs(args map[string]interface{}) {
	for key, value := range args {
		ctl.args[key] = value
	}
}

// AddCobraFlagsToCommand adds flags for the Opssight helm chart to the cmd
// master=true is used to add all flags for creating an instance
// master=false is used to add a subset of flags for updating an instance
func (ctl *HelmValuesFromCobraFlags) AddCobraFlagsToCommand(cmd *cobra.Command, master bool) {
	// Registry Configuration
	// cmd.Flags().StringVar(&ctl.flagTree.IsUpstream, "is-upstream", ctl.flagTree.IsUpstream, "If true, Upstream images and names will be used [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.Registry, "registry", ctl.flagTree.Registry, "Name of the registry to use for images e.g. docker.io/blackducksoftware")
	cmd.Flags().StringSliceVar(&ctl.flagTree.PullSecrets, "pull-secret-name", ctl.flagTree.PullSecrets, "Only if the registry requires authentication")
	// cmd.Flags().StringSliceVar(&ctl.flagTree.ImageRegistries, "image-registries", ctl.flagTree.ImageRegistries, "List of image registries")

	cmd.Flags().StringVar(&ctl.flagTree.LogLevel, "log-level", ctl.flagTree.LogLevel, "Log level of Opssight")

	// Black Duck Configuration
	// TODO: What to do with black duck flags?
	// cmd.Flags().StringVar(&ctl.flagTree.BlackduckExternalHostsFilePath, "blackduck-external-hosts-file-path", ctl.flagTree.BlackduckExternalHostsFilePath, "Absolute path to a file containing a list of Black Duck External Hosts")
	cmd.Flags().StringVar(&ctl.flagTree.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.flagTree.BlackduckTLSVerification, "If true, Opssight performs TLS Verification for Black Duck [true|false]")
	// cmd.Flags().IntVar(&ctl.flagTree.BlackduckInitialCount, "blackduck-initial-count", ctl.flagTree.BlackduckInitialCount, "Initial number of Black Duck instances to create")
	// cmd.Flags().IntVar(&ctl.flagTree.BlackduckMaxCount, "blackduck-max-count", ctl.flagTree.BlackduckMaxCount, "Maximum number of Black Duck instances that can be created")
	// cmd.Flags().StringVar(&ctl.flagTree.BlackduckType, "blackduck-type", ctl.flagTree.BlackduckType, "Type of Black Duck")
	// cmd.Flags().StringVar(&ctl.flagTree.BlackduckPassword, "blackduck-password", ctl.flagTree.BlackduckPassword, "Password to use for all internal Blackduck 'sysadmin' account")

	// Metrics
	cmd.Flags().StringVar(&ctl.flagTree.EnableMetrics, "enable-metrics", ctl.flagTree.EnableMetrics, "If true, Opssight records Prometheus Metrics [true|false]")
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.PrometheusExpose, "expose-metrics", util.NONE, "Type of service of Opssight's Prometheus Metrics [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	} else {
		cmd.Flags().StringVar(&ctl.flagTree.PrometheusExpose, "expose-metrics", ctl.flagTree.PrometheusExpose, "Type of service of Opssight's Prometheus Metrics [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	}

	// Core
	cmd.Flags().StringVar(&ctl.flagTree.PerceptorExpose, "opssight-core-expose", ctl.flagTree.PerceptorExpose, "Type of service for Opssight's core model [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	cmd.Flags().IntVar(&ctl.flagTree.PerceptorCheckForStalledScansPauseHours, "opssight-core-check-scan-hours", ctl.flagTree.PerceptorCheckForStalledScansPauseHours, "Hours Opssight's Core waits between checking for scans")
	cmd.Flags().IntVar(&ctl.flagTree.PerceptorStalledScanClientTimeoutHours, "opssight-core-scan-client-timeout-hours", ctl.flagTree.PerceptorStalledScanClientTimeoutHours, "Hours until Opssight's Core stops checking for scans")
	cmd.Flags().IntVar(&ctl.flagTree.PerceptorModelMetricsPauseSeconds, "opssight-core-metrics-pause-seconds", ctl.flagTree.PerceptorModelMetricsPauseSeconds, "Core metrics pause in seconds")
	cmd.Flags().IntVar(&ctl.flagTree.PerceptorUnknownImagePauseMilliseconds, "opssight-core-unknown-image-pause-milliseconds", ctl.flagTree.PerceptorUnknownImagePauseMilliseconds, "Opssight Core's unknown image pause in milliseconds")
	cmd.Flags().IntVar(&ctl.flagTree.PerceptorClientTimeoutMilliseconds, "opssight-core-client-timeout-milliseconds", ctl.flagTree.PerceptorClientTimeoutMilliseconds, "Seconds for Opssight Core's timeout for Black Duck Scan Client")

	// Processor
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverTLSCertificatePath, "processor-TLS-certificate-path", ctl.flagTree.PerceiverTLSCertificatePath, "Accepts certificate file to start webhook receiver with TLS enabled, works in conjunction with Quay and Artifactory processors")
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverTLSKeyPath, "processor-TLS-key-path", ctl.flagTree.PerceiverTLSKeyPath, "Accepts key file to sign the TLS certificate, works in conjunction with Quay and Artifactory processors")
	cmd.Flags().IntVar(&ctl.flagTree.PerceiverAnnotationIntervalSeconds, "processor-annotation-interval-seconds", ctl.flagTree.PerceiverAnnotationIntervalSeconds, "Refresh interval to get latest scan results and apply to Pods and Images")
	cmd.Flags().IntVar(&ctl.flagTree.PerceiverDumpIntervalMinutes, "processor-dump-interval-minutes", ctl.flagTree.PerceiverDumpIntervalMinutes, "Minutes Image Processor and Pod Processor wait between creating dumps of data/metrics")

	// Pod Processor
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverEnablePodPerceiver, "enable-pod-processor", ctl.flagTree.PerceiverEnablePodPerceiver, "If true, Pod Processor discovers pods for scanning [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverPodPerceiverNamespaceFilter, "pod-processor-namespace-filter", ctl.flagTree.PerceiverPodPerceiverNamespaceFilter, "Pod Processor's filter to scan pods by their namespace")

	// Scanner
	cmd.Flags().IntVar(&ctl.flagTree.ScannerPodScannerClientTimeoutSeconds, "scanner-client-timeout-seconds", ctl.flagTree.ScannerPodScannerClientTimeoutSeconds, "Seconds before Scanner times out for Black Duck's Scan Client")
	cmd.Flags().IntVar(&ctl.flagTree.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.flagTree.ScannerPodReplicaCount, "Number of Containers for scanning")
	cmd.Flags().StringVar(&ctl.flagTree.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.flagTree.ScannerPodImageDirectory, "Directory in Scanner's pod where images are stored for scanning")

	// Image Getter
	// cmd.Flags().StringVar(&ctl.flagTree.ScannerPodImageFacadeInternalRegistriesFilePath, "image-getter-secure-registries-file-path", ctl.flagTree.ScannerPodImageFacadeInternalRegistriesFilePath, "Absolute path to a file for secure docker registries credentials to pull the images for scan")
	cmd.Flags().StringVar(&ctl.flagTree.ScannerPodImageFacadeImagePullerType, "image-getter-image-puller-type", ctl.flagTree.ScannerPodImageFacadeImagePullerType, "Type of Image Getter's Image Puller [docker|skopeo]")

	// Image Processor
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverEnableImagePerceiver, "enable-image-processor", ctl.flagTree.PerceiverEnableImagePerceiver, "If true, Image Processor discovers images for scanning [true|false]")

	// Quay Processor
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverEnableQuayPerceiver, "enable-quay-processor", ctl.flagTree.PerceiverEnableQuayPerceiver, "If true, Quay Processor discovers quay images for scanning [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverQuayExpose, "expose-quay-processor", ctl.flagTree.PerceiverQuayExpose, "Type of service for Quay processor [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")

	// Artifactory Processor
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverEnableArtifactoryPerceiver, "enable-artifactory-processor", ctl.flagTree.PerceiverEnableArtifactoryPerceiver, "If true, Artifactory Processor discovers artifactory images for scanning [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverEnableArtifactoryPerceiverDumper, "enable-artifactory-processor-dumper", ctl.flagTree.PerceiverEnableArtifactoryPerceiverDumper, "If true, Artifactory Processor dumps all docker images in an artifactory instance for scanning [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.PerceiverArtifactoryExpose, "expose-artifactory-processor", ctl.flagTree.PerceiverArtifactoryExpose, "Type of service for Artifactory processor [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")

	// Memory
	cmd.Flags().StringVar(&ctl.flagTree.DefaultCPU, "default-cpu", ctl.flagTree.DefaultCPU, "CPU size of Opssight")
	cmd.Flags().StringVar(&ctl.flagTree.DefaultMem, "default-memory", ctl.flagTree.DefaultMem, "Memory size of Opssight")
	cmd.Flags().StringVar(&ctl.flagTree.ScannerCPU, "scanner-cpu", ctl.flagTree.ScannerCPU, "CPU size of Opssight's Scanner")
	cmd.Flags().StringVar(&ctl.flagTree.ScannerMem, "scanner-memory", ctl.flagTree.ScannerMem, "Memory size of Opssight's Scanner")
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the opssightSpec
func (ctl *HelmValuesFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	if FlagWasSet(flagset, "opssight-core-expose") {
		isValid := util.IsExposeServiceValid(ctl.flagTree.PerceptorExpose)
		if !isValid {
			return fmt.Errorf("opssight core expose must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "expose-metrics") {
		isValid := util.IsExposeServiceValid(ctl.flagTree.PrometheusExpose)
		if !isValid {
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "expose-artifactory-processor") {
		isValid := util.IsExposeServiceValid(ctl.flagTree.PerceiverArtifactoryExpose)
		if !isValid {
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "expose-quay-processor") {
		isValid := util.IsExposeServiceValid(ctl.flagTree.PerceiverQuayExpose)
		if !isValid {
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	// TODO - add check for log level format
	return nil
}

// FlagWasSet returns true if a flag was changed and it exists, otherwise it returns false
func FlagWasSet(flagset *pflag.FlagSet, flagName string) bool {
	if flagset.Lookup(flagName) != nil && flagset.Lookup(flagName).Changed {
		return true
	}
	return false
}

// GenerateHelmFlagsFromCobraFlags checks each flag in synopsysctl and updates the map to
// contain the corresponding helm chart field and value
func (ctl *HelmValuesFromCobraFlags) GenerateHelmFlagsFromCobraFlags(flagset *pflag.FlagSet) (map[string]interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	var isErrorExist bool
	flagset.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			log.Debugf("flag '%s': CHANGED", f.Name)
			switch f.Name {
			// case "is-upstream":
			// 	isUpstream := strings.ToUpper(ctl.flagTree.IsUpstream) == "TRUE"
			// 	util.SetHelmValueInMap(ctl.args, []string{"isUpstream"}, isUpstream)
			case "registry":
				util.SetHelmValueInMap(ctl.args, []string{"registry"}, ctl.flagTree.Registry)
			case "pull-secret-name":
				util.SetHelmValueInMap(ctl.args, []string{"imagePullSecrets"}, ctl.flagTree.PullSecrets)
			// case "image-registries":
			// 	util.SetHelmValueInMap(ctl.args, []string{"imageRegistries"}, ctl.flagTree.ImageRegistries)
			case "log-level":
				util.SetHelmValueInMap(ctl.args, []string{"logLevel"}, ctl.flagTree.LogLevel)
			// TODO: What to do with black duck cases?
			// case "blackduck-external-hosts-file-path":
			// 	data, err := util.ReadFileData(ctl.flagTree.BlackduckExternalHostsFilePath)
			// 	if err != nil {
			// 		log.Fatalf("failed to read external hosts file: %+v", err)
			// 	}
			// 	hostStructs := []*opssightapi.Host{}
			// 	err = json.Unmarshal([]byte(data), &hostStructs)
			// 	if err != nil {
			// 		log.Fatalf("failed to unmarshal internal registry structs: %+v", err)
			// 	}
			// 	util.SetHelmValueInMap(ctl.args, []string{"blackduck", "externalHosts"}, hostStructs)
			case "blackduck-TLS-verification":
				enableTLSVerification := strings.ToUpper(ctl.flagTree.BlackduckTLSVerification) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"blackduck", "tlsVerification"}, enableTLSVerification)
			// case "blackduck-initial-count":
			// 	util.SetHelmValueInMap(ctl.args, []string{"blackduck", "initialCount"}, ctl.flagTree.BlackduckInitialCount)
			// case "blackduck-max-count":
			// 	util.SetHelmValueInMap(ctl.args, []string{"blackduck", "maxCount"}, ctl.flagTree.BlackduckMaxCount)
			// case "blackduck-type":
			// 	util.SetHelmValueInMap(ctl.args, []string{"blackduck", "blackduckSpec", "type"}, ctl.flagTree.BlackduckType)
			// case "blackduck-password":
			// 	util.SetHelmValueInMap(ctl.args, []string{"blackduck", "blackduckPassword"}, crddefaults.Base64Encode([]byte(ctl.flagTree.BlackduckPassword)))
			case "enable-metrics":
				enableMetrics := strings.ToUpper(ctl.flagTree.EnableMetrics) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"prometheus", "enabled"}, enableMetrics)
			case "expose-metrics":
				switch strings.ToUpper(ctl.flagTree.PrometheusExpose) {
				case util.NODEPORT:
					util.SetHelmValueInMap(ctl.args, []string{"prometheus", "expose"}, "NodePort")
				case util.LOADBALANCER:
					util.SetHelmValueInMap(ctl.args, []string{"prometheus", "expose"}, "LoadBalancer")
				case util.NONE:
					util.SetHelmValueInMap(ctl.args, []string{"prometheus", "expose"}, "ClusterIP")
				default:
					util.SetHelmValueInMap(ctl.args, []string{"prometheus", "expose"}, "None")
				}
			case "opssight-core-expose":
				switch strings.ToUpper(ctl.flagTree.PerceptorExpose) {
				case util.NODEPORT:
					util.SetHelmValueInMap(ctl.args, []string{"core", "expose"}, "NodePort")
				case util.LOADBALANCER:
					util.SetHelmValueInMap(ctl.args, []string{"core", "expose"}, "LoadBalancer")
				case util.NONE:
					util.SetHelmValueInMap(ctl.args, []string{"core", "expose"}, "ClusterIP")
				default:
					util.SetHelmValueInMap(ctl.args, []string{"core", "expose"}, "None")
				}
			case "opssight-core-check-scan-hours":
				util.SetHelmValueInMap(ctl.args, []string{"core", "checkForStalledScansPauseHours"}, ctl.flagTree.PerceptorCheckForStalledScansPauseHours)
			case "opssight-core-scan-client-timeout-hours":
				util.SetHelmValueInMap(ctl.args, []string{"core", "stalledScanClientTimeoutHours"}, ctl.flagTree.PerceptorStalledScanClientTimeoutHours)
			case "opssight-core-metrics-pause-seconds":
				util.SetHelmValueInMap(ctl.args, []string{"core", "modelMetricsPauseSeconds"}, ctl.flagTree.PerceptorModelMetricsPauseSeconds)
			case "opssight-core-unknown-image-pause-milliseconds":
				util.SetHelmValueInMap(ctl.args, []string{"core", "unknownImagePauseMilliseconds"}, ctl.flagTree.PerceptorUnknownImagePauseMilliseconds)
			case "opssight-core-client-timeout-milliseconds":
				util.SetHelmValueInMap(ctl.args, []string{"core", "clientTimeoutMilliseconds"}, ctl.flagTree.PerceptorClientTimeoutMilliseconds)
			case "processor-TLS-certificate-path":
				data, err := util.ReadFileData(ctl.flagTree.PerceiverTLSCertificatePath)
				if err != nil {
					log.Errorf("failed to read certificate file: %+v", err)
				}
				util.SetHelmValueInMap(ctl.args, []string{"processor", "certificate"}, data)
			case "processor-TLS-key-path":
				data, err := util.ReadFileData(ctl.flagTree.PerceiverTLSKeyPath)
				if err != nil {
					log.Errorf("failed to read certificate file: %+v", err)
				}
				util.SetHelmValueInMap(ctl.args, []string{"processor", "certificateKey"}, data)
			case "processor-annotation-interval-seconds":
				util.SetHelmValueInMap(ctl.args, []string{"processor", "annotationIntervalSeconds"}, ctl.flagTree.PerceiverAnnotationIntervalSeconds)
			case "processor-dump-interval-minutes":
				util.SetHelmValueInMap(ctl.args, []string{"processor", "dumpIntervalMinutes"}, ctl.flagTree.PerceiverDumpIntervalMinutes)
			case "enable-pod-processor":
				enablePodPerceiver := strings.ToUpper(ctl.flagTree.PerceiverEnablePodPerceiver) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"podProcessor", "enabled"}, enablePodPerceiver)
			case "pod-processor-namespace-filter":
				util.SetHelmValueInMap(ctl.args, []string{"podProcessor", "nameSpaceFilter"}, ctl.flagTree.PerceiverPodPerceiverNamespaceFilter)
			case "scanner-client-timeout-seconds":
				util.SetHelmValueInMap(ctl.args, []string{"scanner", "blackDuckClientTimeoutSeconds"}, ctl.flagTree.ScannerPodScannerClientTimeoutSeconds)
			case "scannerpod-replica-count":
				util.SetHelmValueInMap(ctl.args, []string{"scanner", "replicas"}, ctl.flagTree.ScannerPodReplicaCount)
			case "scannerpod-image-directory":
				util.SetHelmValueInMap(ctl.args, []string{"scanner", "imageDirectory"}, ctl.flagTree.ScannerPodImageDirectory)
			// case "image-getter-secure-registries-file-path":
			// 	data, err := util.ReadFileData(ctl.flagTree.ScannerPodImageFacadeInternalRegistriesFilePath)
			// 	if err != nil {
			// 		log.Fatalf("failed to read internal registries file: %+v", err)
			// 	}
			// 	registryStructs := []*opssightapi.RegistryAuth{}
			// 	err = json.Unmarshal([]byte(data), &registryStructs)
			// 	if err != nil {
			// 		log.Fatalf("failed to unmarshal internal registries: %+v", err)
			// 	}
			// 	util.SetHelmValueInMap(ctl.args, []string{"imageGetter"}, registryStructs)
			case "image-getter-image-puller-type":
				util.SetHelmValueInMap(ctl.args, []string{"imageGetter", "imagePullerType"}, ctl.flagTree.ScannerPodImageFacadeImagePullerType)
			case "enable-image-processor":
				enableImagePerceiver := strings.ToUpper(ctl.flagTree.PerceiverEnableImagePerceiver) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"imageProcessor", "enabled"}, enableImagePerceiver)
			case "enable-quay-processor":
				enableQuayPerceiver := strings.ToUpper(ctl.flagTree.PerceiverEnableQuayPerceiver) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "enabled"}, enableQuayPerceiver)
			case "expose-quay-processor":
				switch strings.ToUpper(ctl.flagTree.PerceiverQuayExpose) {
				case util.NODEPORT:
					util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "expose"}, "NodePort")
				case util.LOADBALANCER:
					util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "expose"}, "LoadBalancer")
				case util.NONE:
					util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "expose"}, "ClusterIP")
				default:
					util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "expose"}, "None")
				}
			case "enable-artifactory-processor":
				enableArtifactoryPerceiver := strings.ToUpper(ctl.flagTree.PerceiverEnableArtifactoryPerceiver) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "enabled"}, enableArtifactoryPerceiver)
			case "enable-artifactory-processor-dumper":
				enableArtifactoryPerceiverDumper := strings.ToUpper(ctl.flagTree.PerceiverEnableArtifactoryPerceiverDumper) == "TRUE"
				util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "dumper"}, enableArtifactoryPerceiverDumper)
			case "expose-artifactory-processor":
				switch strings.ToUpper(ctl.flagTree.PerceiverArtifactoryExpose) {
				case util.NODEPORT:
					util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "expose"}, "NodePort")
				case util.LOADBALANCER:
					util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "expose"}, "LoadBalancer")
				case util.NONE:
					util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "expose"}, "ClusterIP")
				default:
					util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "expose"}, "None")
				}
			case "default-cpu":
				util.SetHelmValueInMap(ctl.args, []string{"prometheus", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"core", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"podProcessor", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"imageGetter", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"imageProcessor", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				if !FlagWasSet(flagset, "scanner-cpu") {
					util.SetHelmValueInMap(ctl.args, []string{"scanner", "resources", "requests", "cpu"}, ctl.flagTree.DefaultCPU)
				}
			case "default-memory":
				util.SetHelmValueInMap(ctl.args, []string{"prometheus", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"core", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"podProcessor", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"imageGetter", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"imageProcessor", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"quayProcessor", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				util.SetHelmValueInMap(ctl.args, []string{"artifactoryProcessor", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				if !FlagWasSet(flagset, "scanner-memory") {
					util.SetHelmValueInMap(ctl.args, []string{"scanner", "resources", "requests", "memory"}, ctl.flagTree.DefaultMem)
				}
			case "scanner-cpu":
				util.SetHelmValueInMap(ctl.args, []string{"scanner", "resources", "requests", "cpu"}, ctl.flagTree.ScannerCPU)
			case "scanner-memory":
				util.SetHelmValueInMap(ctl.args, []string{"scanner", "resources", "requests", "memory"}, ctl.flagTree.ScannerMem)
			default:
				log.Debugf("flag '%s': NOT FOUND", f.Name)
			}
		} else {
			log.Debugf("flag '%s': UNCHANGED", f.Name)
		}
	})

	if isErrorExist {
		log.Fatalf("please fix all the above errors to continue")
	}

	return ctl.args, nil
}
