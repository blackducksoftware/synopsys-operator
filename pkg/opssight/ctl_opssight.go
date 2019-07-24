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
	"encoding/json"
	"fmt"
	"strings"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create an OpsSight CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making an OpsSight Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	opsSightSpec                                    *opssightapi.OpsSightSpec
	PerceptorName                                   string
	PerceptorImage                                  string
	PerceptorPort                                   int
	PerceptorExpose                                 string
	PerceptorCheckForStalledScansPauseHours         int
	PerceptorStalledScanClientTimeoutHours          int
	PerceptorModelMetricsPauseSeconds               int
	PerceptorUnknownImagePauseMilliseconds          int
	PerceptorClientTimeoutMilliseconds              int
	ScannerPodName                                  string
	ScannerPodScannerName                           string
	ScannerPodScannerImage                          string
	ScannerPodScannerPort                           int
	ScannerPodScannerClientTimeoutSeconds           int
	ScannerPodImageFacadeName                       string
	ScannerPodImageFacadeImage                      string
	ScannerPodImageFacadePort                       int
	ScannerPodImageFacadeInternalRegistriesFilePath string
	ScannerPodImageFacadeImagePullerType            string
	ScannerPodImageFacadeServiceAccount             string
	ScannerPodReplicaCount                          int
	ScannerPodImageDirectory                        string
	PerceiverEnableImagePerceiver                   string
	PerceiverEnablePodPerceiver                     string
	PerceiverImagePerceiverName                     string
	PerceiverImagePerceiverImage                    string
	PerceiverPodPerceiverName                       string
	PerceiverPodPerceiverImage                      string
	PerceiverPodPerceiverNamespaceFilter            string
	PerceiverAnnotationIntervalSeconds              int
	PerceiverDumpIntervalMinutes                    int
	PerceiverServiceAccount                         string
	PerceiverPort                                   int
	ConfigMapName                                   string
	SecretName                                      string
	DefaultCPU                                      string
	DefaultMem                                      string
	ScannerCPU                                      string
	ScannerMem                                      string
	LogLevel                                        string
	EnableMetrics                                   string
	PrometheusName                                  string
	PrometheusImage                                 string
	PrometheusPort                                  int
	PrometheusExpose                                string
	SkyfireName                                     string
	SkyfireImage                                    string
	EnableSkyfire                                   string
	SkyfirePort                                     int
	SkyfirePrometheusPort                           int
	SkyfireServiceAccount                           string
	SkyfireHubClientTimeoutSeconds                  int
	SkyfireHubDumpPauseSeconds                      int
	SkyfireKubeDumpIntervalSeconds                  int
	SkyfirePerceptorDumpIntervalSeconds             int
	BlackduckExternalHostsFilePath                  string
	BlackduckConnectionsEnvironmentVaraiableName    string
	BlackduckTLSVerification                        string
	BlackduckPassword                               string
	BlackduckInitialCount                           int
	BlackduckMaxCount                               int
	BlackduckType                                   string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		opsSightSpec: &opssightapi.OpsSightSpec{},
	}
}

// GetCRSpec returns a pointer to the OpsSightSpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return *ctl.opsSightSpec
}

// SetCRSpec sets the opsSightSpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(opssightapi.OpsSightSpec)
	if !ok {
		return fmt.Errorf("error setting OpsSight spec")
	}
	ctl.opsSightSpec = &convertedSpec
	return nil
}

// Constants for predefined specs
const (
	EmptySpec             string = "empty"
	UpstreamSpec          string = "upstream"
	DefaultSpec           string = "default"
	DisabledBlackDuckSpec string = "disabledBlackDuck"
)

// SetPredefinedCRSpec sets the opsSightSpec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	switch specType {
	case EmptySpec:
		ctl.opsSightSpec = &opssightapi.OpsSightSpec{}
	case UpstreamSpec:
		ctl.opsSightSpec = crddefaults.GetOpsSightUpstream()
		ctl.opsSightSpec.Perceiver.EnablePodPerceiver = true
		ctl.opsSightSpec.EnableMetrics = true
	case DefaultSpec:
		ctl.opsSightSpec = crddefaults.GetOpsSightDefault()
		ctl.opsSightSpec.Perceiver.EnablePodPerceiver = true
		ctl.opsSightSpec.EnableMetrics = true
	case DisabledBlackDuckSpec:
		ctl.opsSightSpec = crddefaults.GetOpsSightDefaultWithIPV6DisabledBlackDuck()
		ctl.opsSightSpec.Perceiver.EnablePodPerceiver = true
		ctl.opsSightSpec.EnableMetrics = true
	default:
		return fmt.Errorf("OpsSight spec type '%s' is not valid", specType)
	}
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for OpsSight's Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.PerceptorImage, "opssight-core-image", ctl.PerceptorImage, "Image of OpsSight's Core")
	if master {
		cmd.Flags().StringVar(&ctl.PerceptorExpose, "opssight-core-expose", util.NONE, "Type of service for OpsSight's core model (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	} else {
		cmd.Flags().StringVar(&ctl.PerceptorExpose, "opssight-core-expose", ctl.PerceptorExpose, "Type of service for OpsSight's core model (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	}
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "opssight-core-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "Hours OpsSight's Core waits between checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "opssight-core-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "Hours until OpsSight's Core stops checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "opssight-core-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "Core metrics pause in seconds")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "opssight-core-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "OpsSight Core's unknown image pause in milliseconds")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "opssight-core-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "Seconds for OpsSight Core's timeout for Black Duck Scan Client")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "scanner-image", ctl.ScannerPodScannerImage, "Image URL of Scanner")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "Seconds before Scanner times out for Black Duck's Scan Client")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "image-getter-image", ctl.ScannerPodImageFacadeImage, "Image Getter Container's image")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "image-getter-secure-registries-file-path", ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "Absolute path to a file for secure docker registries credentials to pull the images for scan")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "image-getter-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "Type of Image Getter's Image Puller (docker|skopeo)")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "image-getter-service-account", ctl.ScannerPodImageFacadeServiceAccount, "Service Account of Image Getter")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.ScannerPodReplicaCount, "Number of Containers for scanning")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.ScannerPodImageDirectory, "Directory in Scanner's pod where images are stored for scanning")
	cmd.Flags().StringVar(&ctl.PerceiverEnableImagePerceiver, "enable-image-processor", ctl.PerceiverEnableImagePerceiver, "If true, Image Processor discovers images for scanning (true|false)")
	cmd.Flags().StringVar(&ctl.PerceiverEnablePodPerceiver, "enable-pod-processor", ctl.PerceiverEnablePodPerceiver, "If true, Pod Processor discovers pods for scanning (true|false)")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverImage, "image-processor-image", ctl.PerceiverImagePerceiverImage, "Image of Image Processor")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverImage, "pod-processor-image", ctl.PerceiverPodPerceiverImage, "Image of Pod Processor")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverNamespaceFilter, "pod-processor-namespace-filter", ctl.PerceiverPodPerceiverNamespaceFilter, "Pod Processor's filter to scan pods by their namespace")
	cmd.Flags().IntVar(&ctl.PerceiverAnnotationIntervalSeconds, "processor-annotation-interval-seconds", ctl.PerceiverAnnotationIntervalSeconds, "Refresh interval to get latest scan results and apply to Pods and Images")
	cmd.Flags().IntVar(&ctl.PerceiverDumpIntervalMinutes, "processor-dump-interval-minutes", ctl.PerceiverDumpIntervalMinutes, "Minutes Image Processor and Pod Processor wait between creating dumps of data/metrics")
	cmd.Flags().StringVar(&ctl.DefaultCPU, "default-cpu", ctl.DefaultCPU, "CPU size of OpsSight")
	cmd.Flags().StringVar(&ctl.DefaultMem, "default-memory", ctl.DefaultMem, "Memory size of OpsSight")
	cmd.Flags().StringVar(&ctl.ScannerCPU, "scanner-cpu", ctl.ScannerCPU, "CPU size of OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.ScannerMem, "scanner-memory", ctl.ScannerMem, "Memory size of OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.LogLevel, "log-level", ctl.LogLevel, "Log level of OpsSight")
	cmd.Flags().StringVar(&ctl.EnableMetrics, "enable-metrics", ctl.EnableMetrics, "If true, OpsSight records Prometheus Metrics (true|false)")
	cmd.Flags().StringVar(&ctl.PrometheusImage, "metrics-image", ctl.PrometheusImage, "Image of OpsSight's Prometheus Metrics")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "metrics-port", ctl.PrometheusPort, "Port of OpsSight's Prometheus Metrics")
	if master {
		cmd.Flags().StringVar(&ctl.PrometheusExpose, "expose-metrics", util.NONE, "Type of service of OpsSight's Prometheus Metrics (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	} else {
		cmd.Flags().StringVar(&ctl.PrometheusExpose, "expose-metrics", ctl.PrometheusExpose, "Type of service of OpsSight's Prometheus Metrics (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	}
	cmd.Flags().StringVar(&ctl.BlackduckExternalHostsFilePath, "blackduck-external-hosts-file-path", ctl.BlackduckExternalHostsFilePath, "Absolute path to a file containing a list of Black Duck External Hosts")
	cmd.Flags().StringVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "If true, OpsSight performs TLS Verification for Black Duck (true|false)")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Black Duck instances to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Black Duck instances that can be created")
	cmd.Flags().StringVar(&ctl.BlackduckType, "blackduck-type", ctl.BlackduckType, "Type of Black Duck")
	cmd.Flags().StringVar(&ctl.BlackduckPassword, "blackduck-password", ctl.BlackduckPassword, "Password to use for all internal Blackduck 'sysadmin' account")
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the opsSightSpec
func (ctl *CRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	if FlagWasSet(flagset, "opssight-core-expose") {
		isValid := util.IsExposeServiceValid(ctl.PerceptorExpose)
		if !isValid {
			return fmt.Errorf("opssight core expose must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "expose-metrics") {
		isValid := util.IsExposeServiceValid(ctl.PrometheusExpose)
		if !isValid {
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	return nil
}

// FlagWasSet returns true if a flag was changed and it exists, otherwise it returns false
func FlagWasSet(flagset *pflag.FlagSet, flagName string) bool {
	if flagset.Lookup(flagName) != nil && flagset.Lookup(flagName).Changed {
		return true
	}
	return false
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the opsSightSpec with the value that's stored
// in the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return *ctl.opsSightSpec, nil
}

// InternalRegistryStructs - file format for reading data
type InternalRegistryStructs struct {
	Data []opssightapi.RegistryAuth
}

// ExternalHostStructs - file format for reading data
type ExternalHostStructs struct {
	Data []opssightapi.Host
}

// SetCRSpecFieldByFlag updates a field in the opsSightSpec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "opssight-core-image":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.Image = ctl.PerceptorImage
		case "opssight-core-expose":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.Expose = ctl.PerceptorExpose
		case "opssight-core-check-scan-hours":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.CheckForStalledScansPauseHours = ctl.PerceptorCheckForStalledScansPauseHours
		case "opssight-core-scan-client-timeout-hours":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.StalledScanClientTimeoutHours = ctl.PerceptorStalledScanClientTimeoutHours
		case "opssight-core-metrics-pause-seconds":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.ModelMetricsPauseSeconds = ctl.PerceptorModelMetricsPauseSeconds
		case "opssight-core-unknown-image-pause-milliseconds":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.UnknownImagePauseMilliseconds = ctl.PerceptorUnknownImagePauseMilliseconds
		case "opssight-core-client-timeout-milliseconds":
			if ctl.opsSightSpec.Perceptor == nil {
				ctl.opsSightSpec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.opsSightSpec.Perceptor.ClientTimeoutMilliseconds = ctl.PerceptorClientTimeoutMilliseconds
		case "scanner-image":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.Scanner == nil {
				ctl.opsSightSpec.ScannerPod.Scanner = &opssightapi.Scanner{}
			}
			ctl.opsSightSpec.ScannerPod.Scanner.Image = ctl.ScannerPodScannerImage
		case "scanner-client-timeout-seconds":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.Scanner == nil {
				ctl.opsSightSpec.ScannerPod.Scanner = &opssightapi.Scanner{}
			}
			ctl.opsSightSpec.ScannerPod.Scanner.ClientTimeoutSeconds = ctl.ScannerPodScannerClientTimeoutSeconds
		case "image-getter-image":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.ImageFacade == nil {
				ctl.opsSightSpec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.opsSightSpec.ScannerPod.ImageFacade.Image = ctl.ScannerPodImageFacadeImage
		case "image-getter-secure-registries-file-path":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.ImageFacade == nil {
				ctl.opsSightSpec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			data, err := util.ReadFileData(ctl.ScannerPodImageFacadeInternalRegistriesFilePath)
			if err != nil {
				log.Errorf("failed to read internal registries file: %+v", err)
				return
			}
			registryStructs := []*opssightapi.RegistryAuth{}
			err = json.Unmarshal([]byte(data), &registryStructs)
			if err != nil {
				log.Errorf("failed to unmarshal internal registries: %+v", err)
				return
			}
			ctl.opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = registryStructs
		case "image-getter-image-puller-type":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.ImageFacade == nil {
				ctl.opsSightSpec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.opsSightSpec.ScannerPod.ImageFacade.ImagePullerType = ctl.ScannerPodImageFacadeImagePullerType
		case "image-getter-service-account":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.opsSightSpec.ScannerPod.ImageFacade == nil {
				ctl.opsSightSpec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.opsSightSpec.ScannerPod.ImageFacade.ServiceAccount = ctl.ScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			ctl.opsSightSpec.ScannerPod.ReplicaCount = ctl.ScannerPodReplicaCount
		case "scannerpod-image-directory":
			if ctl.opsSightSpec.ScannerPod == nil {
				ctl.opsSightSpec.ScannerPod = &opssightapi.ScannerPod{}
			}
			ctl.opsSightSpec.ScannerPod.ImageDirectory = ctl.ScannerPodImageDirectory
		case "enable-image-processor":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.opsSightSpec.Perceiver.EnableImagePerceiver = strings.ToUpper(ctl.PerceiverEnableImagePerceiver) == "TRUE"
		case "enable-pod-processor":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.opsSightSpec.Perceiver.EnablePodPerceiver = strings.ToUpper(ctl.PerceiverEnablePodPerceiver) == "TRUE"
		case "image-processor-image":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.opsSightSpec.Perceiver.ImagePerceiver == nil {
				ctl.opsSightSpec.Perceiver.ImagePerceiver = &opssightapi.ImagePerceiver{}
			}
			ctl.opsSightSpec.Perceiver.ImagePerceiver.Image = ctl.PerceiverImagePerceiverImage
		case "pod-processor-image":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.opsSightSpec.Perceiver.PodPerceiver == nil {
				ctl.opsSightSpec.Perceiver.PodPerceiver = &opssightapi.PodPerceiver{}
			}
			ctl.opsSightSpec.Perceiver.PodPerceiver.Image = ctl.PerceiverPodPerceiverImage
		case "pod-processor-namespace-filter":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.opsSightSpec.Perceiver.PodPerceiver == nil {
				ctl.opsSightSpec.Perceiver.PodPerceiver = &opssightapi.PodPerceiver{}
			}
			ctl.opsSightSpec.Perceiver.PodPerceiver.NamespaceFilter = ctl.PerceiverPodPerceiverNamespaceFilter
		case "processor-annotation-interval-seconds":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.opsSightSpec.Perceiver.AnnotationIntervalSeconds = ctl.PerceiverAnnotationIntervalSeconds
		case "processor-dump-interval-minutes":
			if ctl.opsSightSpec.Perceiver == nil {
				ctl.opsSightSpec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.opsSightSpec.Perceiver.DumpIntervalMinutes = ctl.PerceiverDumpIntervalMinutes
		case "default-cpu":
			ctl.opsSightSpec.DefaultCPU = ctl.DefaultCPU
		case "default-memory":
			ctl.opsSightSpec.DefaultMem = ctl.DefaultMem
		case "scanner-cpu":
			ctl.opsSightSpec.ScannerCPU = ctl.ScannerCPU
		case "scanner-memory":
			ctl.opsSightSpec.ScannerMem = ctl.ScannerMem
		case "log-level":
			ctl.opsSightSpec.LogLevel = ctl.LogLevel
		case "enable-metrics":
			ctl.opsSightSpec.EnableMetrics = strings.ToUpper(ctl.EnableMetrics) == "TRUE"
		case "metrics-image":
			if ctl.opsSightSpec.Prometheus == nil {
				ctl.opsSightSpec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.opsSightSpec.Prometheus.Image = ctl.PrometheusImage
		case "metrics-port":
			if ctl.opsSightSpec.Prometheus == nil {
				ctl.opsSightSpec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.opsSightSpec.Prometheus.Port = ctl.PrometheusPort
		case "expose-metrics":
			if ctl.opsSightSpec.Prometheus == nil {
				ctl.opsSightSpec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.opsSightSpec.Prometheus.Expose = ctl.PrometheusExpose
		case "blackduck-external-hosts-file-path":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			data, err := util.ReadFileData(ctl.BlackduckExternalHostsFilePath)
			if err != nil {
				log.Errorf("failed to read external hosts file: %+v", err)
				return
			}
			hostStructs := []*opssightapi.Host{}
			err = json.Unmarshal([]byte(data), &hostStructs)
			if err != nil {
				log.Errorf("failed to unmarshal internal registry structs: %+v", err)
				return
			}
			ctl.opsSightSpec.Blackduck.ExternalHosts = hostStructs
		case "blackduck-TLS-verification":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.opsSightSpec.Blackduck.TLSVerification = strings.ToUpper(ctl.BlackduckTLSVerification) == "TRUE"
		case "blackduck-initial-count":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.opsSightSpec.Blackduck.InitialCount = ctl.BlackduckInitialCount
		case "blackduck-max-count":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.opsSightSpec.Blackduck.MaxCount = ctl.BlackduckMaxCount
		case "blackduck-type":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			if ctl.opsSightSpec.Blackduck.BlackduckSpec == nil {
				ctl.opsSightSpec.Blackduck.BlackduckSpec = &blackduckapi.BlackduckSpec{}
			}
			ctl.opsSightSpec.Blackduck.BlackduckSpec.Type = ctl.BlackduckType
		case "blackduck-password":
			if ctl.opsSightSpec.Blackduck == nil {
				ctl.opsSightSpec.Blackduck = &opssightapi.Blackduck{}
			}
			if ctl.opsSightSpec.Blackduck.BlackduckSpec == nil {
				ctl.opsSightSpec.Blackduck.BlackduckSpec = &blackduckapi.BlackduckSpec{}
			}
			ctl.opsSightSpec.Blackduck.BlackduckPassword = crddefaults.Base64Encode([]byte(ctl.BlackduckPassword))
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
