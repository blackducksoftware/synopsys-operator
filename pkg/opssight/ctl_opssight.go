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

// Ctl type provides functionality for an OpsSight
// for the Synopsysctl tool
type Ctl struct {
	Spec                                            *opssightapi.OpsSightSpec
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

// NewOpsSightCtl creates a new Ctl struct
func NewOpsSightCtl() *Ctl {
	return &Ctl{
		Spec: &opssightapi.OpsSightSpec{},
	}
}

// GetSpec returns the Spec for the resource
func (ctl *Ctl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *Ctl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(opssightapi.OpsSightSpec)
	if !ok {
		return fmt.Errorf("error setting OpsSight spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags(flagset *pflag.FlagSet) error {
	return nil
}

// Constants
const (
	EmptySpec             string = "empty"
	UpstreamSpec          string = "upstream"
	DefaultSpec           string = "default"
	DisabledBlackDuckSpec string = "disabledBlackDuck"
)

// SwitchSpec switches OpsSight's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(createOpsSightSpecType string) error {
	switch createOpsSightSpecType {
	case EmptySpec:
		ctl.Spec = &opssightapi.OpsSightSpec{}
	case UpstreamSpec:
		ctl.Spec = crddefaults.GetOpsSightUpstream()
		ctl.Spec.Perceiver.EnablePodPerceiver = true
		ctl.Spec.EnableMetrics = true
	case DefaultSpec:
		ctl.Spec = crddefaults.GetOpsSightDefault()
		ctl.Spec.Perceiver.EnablePodPerceiver = true
		ctl.Spec.EnableMetrics = true
	case DisabledBlackDuckSpec:
		ctl.Spec = crddefaults.GetOpsSightDefaultWithIPV6DisabledBlackDuck()
		ctl.Spec.Perceiver.EnablePodPerceiver = true
		ctl.Spec.EnableMetrics = true
	default:
		return fmt.Errorf("OpsSight spec type '%s' is not valid", createOpsSightSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for OpsSight's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.PerceptorImage, "opssight-core-image", ctl.PerceptorImage, "Image of OpsSight's Core")
	cmd.Flags().StringVar(&ctl.PerceptorExpose, "opssight-core-expose", ctl.PerceptorExpose, "Type of service for OpsSight's core model [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "opssight-core-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "Hours OpsSight's Core waits between checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "opssight-core-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "Hours until OpsSight's Core stops checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "opssight-core-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "Core metrics pause in seconds")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "opssight-core-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "OpsSight Core's unknown image pause in milliseconds")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "opssight-core-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "Seconds for OpsSight Core's timeout for Black Duck Scan Client")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "scanner-image", ctl.ScannerPodScannerImage, "Image URL of Scanner")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "Seconds before Scanner times out for Black Duck's Scan Client")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "image-getter-image", ctl.ScannerPodImageFacadeImage, "Image Getter Container's image")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "image-getter-secure-registries-file-path", ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "Absolute path to a file for secure docker registries credentials to pull the images for scan")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "image-getter-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "Type of Image Getter's Image Puller [docker|skopeo]")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "image-getter-service-account", ctl.ScannerPodImageFacadeServiceAccount, "Service Account of Image Getter")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.ScannerPodReplicaCount, "Number of Containers for scanning")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.ScannerPodImageDirectory, "Directory in Scanner's pod where images are stored for scanning")
	cmd.Flags().StringVar(&ctl.PerceiverEnableImagePerceiver, "enable-image-processor", ctl.PerceiverEnableImagePerceiver, "If true, Image Processor discovers images for scanning [true|false]")
	cmd.Flags().StringVar(&ctl.PerceiverEnablePodPerceiver, "enable-pod-processor", ctl.PerceiverEnablePodPerceiver, "If true, Pod Processor discovers pods for scanning [true|false]")
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
	cmd.Flags().StringVar(&ctl.EnableMetrics, "enable-metrics", ctl.EnableMetrics, "If true, OpsSight records Prometheus Metrics [true|false]")
	cmd.Flags().StringVar(&ctl.PrometheusImage, "metrics-image", ctl.PrometheusImage, "Image of OpsSight's Prometheus Metrics")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "metrics-port", ctl.PrometheusPort, "Port of OpsSight's Prometheus Metrics")
	cmd.Flags().StringVar(&ctl.PrometheusExpose, "expose-metrics", ctl.PrometheusExpose, "Type of service of OpsSight's Prometheus Metrics [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().StringVar(&ctl.BlackduckExternalHostsFilePath, "blackduck-external-hosts-file-path", ctl.BlackduckExternalHostsFilePath, "Absolute path to a file containing a list of Black Duck External Hosts")
	cmd.Flags().StringVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "If true, OpsSight performs TLS Verification for Black Duck [true|false]")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Black Duck instances to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Black Duck instances that can be created")
	cmd.Flags().StringVar(&ctl.BlackduckType, "blackduck-type", ctl.BlackduckType, "Type of Black Duck")
	cmd.Flags().StringVar(&ctl.BlackduckPassword, "blackduck-password", ctl.BlackduckPassword, "Password to use for all internal Blackduck 'sysadmin' account")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	flagset.VisitAll(ctl.SetFlag)
}

// InternalRegistryStructs - file format for reading data
type InternalRegistryStructs struct {
	Data []opssightapi.RegistryAuth
}

// ExternalHostStructs - file format for reading data
type ExternalHostStructs struct {
	Data []opssightapi.Host
}

// SetFlag sets an OpsSights's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "opssight-core-image":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.Image = ctl.PerceptorImage
		case "opssight-core-expose":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.Expose = ctl.PerceptorExpose
		case "opssight-core-check-scan-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.CheckForStalledScansPauseHours = ctl.PerceptorCheckForStalledScansPauseHours
		case "opssight-core-scan-client-timeout-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.StalledScanClientTimeoutHours = ctl.PerceptorStalledScanClientTimeoutHours
		case "opssight-core-metrics-pause-seconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.ModelMetricsPauseSeconds = ctl.PerceptorModelMetricsPauseSeconds
		case "opssight-core-unknown-image-pause-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.UnknownImagePauseMilliseconds = ctl.PerceptorUnknownImagePauseMilliseconds
		case "opssight-core-client-timeout-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightapi.Perceptor{}
			}
			ctl.Spec.Perceptor.ClientTimeoutMilliseconds = ctl.PerceptorClientTimeoutMilliseconds
		case "scanner-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightapi.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Image = ctl.ScannerPodScannerImage
		case "scanner-client-timeout-seconds":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightapi.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.ClientTimeoutSeconds = ctl.ScannerPodScannerClientTimeoutSeconds
		case "image-getter-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Image = ctl.ScannerPodImageFacadeImage
		case "image-getter-secure-registries-file-path":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
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
			ctl.Spec.ScannerPod.ImageFacade.InternalRegistries = registryStructs
		case "image-getter-image-puller-type":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ImagePullerType = ctl.ScannerPodImageFacadeImagePullerType
		case "image-getter-service-account":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightapi.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ServiceAccount = ctl.ScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ReplicaCount = ctl.ScannerPodReplicaCount
		case "scannerpod-image-directory":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightapi.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ImageDirectory = ctl.ScannerPodImageDirectory
		case "enable-image-processor":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.Spec.Perceiver.EnableImagePerceiver = strings.ToUpper(ctl.PerceiverEnableImagePerceiver) == "TRUE"
		case "enable-pod-processor":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.Spec.Perceiver.EnablePodPerceiver = strings.ToUpper(ctl.PerceiverEnablePodPerceiver) == "TRUE"
		case "image-processor-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightapi.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Image = ctl.PerceiverImagePerceiverImage
		case "pod-processor-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightapi.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Image = ctl.PerceiverPodPerceiverImage
		case "pod-processor-namespace-filter":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightapi.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.NamespaceFilter = ctl.PerceiverPodPerceiverNamespaceFilter
		case "processor-annotation-interval-seconds":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.Spec.Perceiver.AnnotationIntervalSeconds = ctl.PerceiverAnnotationIntervalSeconds
		case "processor-dump-interval-minutes":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightapi.Perceiver{}
			}
			ctl.Spec.Perceiver.DumpIntervalMinutes = ctl.PerceiverDumpIntervalMinutes
		case "default-cpu":
			ctl.Spec.DefaultCPU = ctl.DefaultCPU
		case "default-memory":
			ctl.Spec.DefaultMem = ctl.DefaultMem
		case "scanner-cpu":
			ctl.Spec.ScannerCPU = ctl.ScannerCPU
		case "scanner-memory":
			ctl.Spec.ScannerMem = ctl.ScannerMem
		case "log-level":
			ctl.Spec.LogLevel = ctl.LogLevel
		case "enable-metrics":
			ctl.Spec.EnableMetrics = strings.ToUpper(ctl.EnableMetrics) == "TRUE"
		case "metrics-image":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.Spec.Prometheus.Image = ctl.PrometheusImage
		case "metrics-port":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.Spec.Prometheus.Port = ctl.PrometheusPort
		case "expose-metrics":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightapi.Prometheus{}
			}
			ctl.Spec.Prometheus.Expose = ctl.PrometheusExpose
		case "blackduck-external-hosts-file-path":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
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
			ctl.Spec.Blackduck.ExternalHosts = hostStructs
		case "blackduck-TLS-verification":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.Spec.Blackduck.TLSVerification = strings.ToUpper(ctl.BlackduckTLSVerification) == "TRUE"
		case "blackduck-initial-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.Spec.Blackduck.InitialCount = ctl.BlackduckInitialCount
		case "blackduck-max-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
			}
			ctl.Spec.Blackduck.MaxCount = ctl.BlackduckMaxCount
		case "blackduck-type":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
			}
			if ctl.Spec.Blackduck.BlackduckSpec == nil {
				ctl.Spec.Blackduck.BlackduckSpec = &blackduckapi.BlackduckSpec{}
			}
			ctl.Spec.Blackduck.BlackduckSpec.Type = ctl.BlackduckType
		case "blackduck-password":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightapi.Blackduck{}
			}
			if ctl.Spec.Blackduck.BlackduckSpec == nil {
				ctl.Spec.Blackduck.BlackduckSpec = &blackduckapi.BlackduckSpec{}
			}
			ctl.Spec.Blackduck.BlackduckPassword = crddefaults.Base64Encode([]byte(ctl.BlackduckPassword))
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}

// SpecIsValid verifies the spec has necessary fields to deploy
func (ctl *Ctl) SpecIsValid() (bool, error) {
	return true, nil
}

// CanUpdate checks if a user has permission to modify based on the spec
func (ctl *Ctl) CanUpdate() (bool, error) {
	return true, nil
}
