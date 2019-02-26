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
	"encoding/json"
	"fmt"

	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// OpsSightCtl type provides functionality for an OpsSight
// for the Synopsysctl tool
type OpsSightCtl struct {
	Spec                                             *opssightv1.OpsSightSpec
	PerceptorName                                    string
	PerceptorImage                                   string
	PerceptorPort                                    int
	PerceptorCheckForStalledScansPauseHours          int
	PerceptorStalledScanClientTimeoutHours           int
	PerceptorModelMetricsPauseSeconds                int
	PerceptorUnknownImagePauseMilliseconds           int
	PerceptorClientTimeoutMilliseconds               int
	ScannerPodName                                   string
	ScannerPodScannerName                            string
	ScannerPodScannerImage                           string
	ScannerPodScannerPort                            int
	ScannerPodScannerClientTimeoutSeconds            int
	ScannerPodImageFacadeName                        string
	ScannerPodImageFacadeImage                       string
	ScannerPodImageFacadePort                        int
	ScannerPodImageFacadeInternalRegistriesJSONSlice []string
	ScannerPodImageFacadeImagePullerType             string
	ScannerPodImageFacadeServiceAccount              string
	ScannerPodReplicaCount                           int
	ScannerPodImageDirectory                         string
	PerceiverEnableImagePerceiver                    bool
	PerceiverEnablePodPerceiver                      bool
	PerceiverImagePerceiverName                      string
	PerceiverImagePerceiverImage                     string
	PerceiverPodPerceiverName                        string
	PerceiverPodPerceiverImage                       string
	PerceiverPodPerceiverNamespaceFilter             string
	PerceiverAnnotationIntervalSeconds               int
	PerceiverDumpIntervalMinutes                     int
	PerceiverServiceAccount                          string
	PerceiverPort                                    int
	PrometheusName                                   string
	PrometheusImage                                  string
	PrometheusPort                                   int
	EnableSkyfire                                    bool
	SkyfireName                                      string
	SkyfireImage                                     string
	SkyfirePort                                      int
	SkyfirePrometheusPort                            int
	SkyfireServiceAccount                            string
	SkyfireHubClientTimeoutSeconds                   int
	SkyfireHubDumpPauseSeconds                       int
	SkyfireKubeDumpIntervalSeconds                   int
	SkyfirePerceptorDumpIntervalSeconds              int
	BlackduckHosts                                   []string
	BlackduckUser                                    string
	BlackduckPort                                    int
	BlackduckConcurrentScanLimit                     int
	BlackduckTotalScanLimit                          int
	BlackduckPasswordEnvVar                          string
	BlackduckInitialCount                            int
	BlackduckMaxCount                                int
	BlackduckDeleteHubThresholdPercentage            int
	EnableMetrics                                    bool
	DefaultCPU                                       string
	DefaultMem                                       string
	LogLevel                                         string
	ConfigMapName                                    string
	SecretName                                       string
}

// NewOpsSightCtl creates a new OpsSightCtl struct
func NewOpsSightCtl() *OpsSightCtl {
	return &OpsSightCtl{
		Spec:                                             &opssightv1.OpsSightSpec{},
		PerceptorName:                                    "",
		PerceptorImage:                                   "",
		PerceptorPort:                                    0,
		PerceptorCheckForStalledScansPauseHours:          0,
		PerceptorStalledScanClientTimeoutHours:           0,
		PerceptorModelMetricsPauseSeconds:                0,
		PerceptorUnknownImagePauseMilliseconds:           0,
		PerceptorClientTimeoutMilliseconds:               0,
		ScannerPodName:                                   "",
		ScannerPodScannerName:                            "",
		ScannerPodScannerImage:                           "",
		ScannerPodScannerPort:                            0,
		ScannerPodScannerClientTimeoutSeconds:            0,
		ScannerPodImageFacadeName:                        "",
		ScannerPodImageFacadeImage:                       "",
		ScannerPodImageFacadePort:                        0,
		ScannerPodImageFacadeInternalRegistriesJSONSlice: []string{},
		ScannerPodImageFacadeImagePullerType:             "",
		ScannerPodImageFacadeServiceAccount:              "",
		ScannerPodReplicaCount:                           0,
		ScannerPodImageDirectory:                         "",
		PerceiverEnableImagePerceiver:                    false,
		PerceiverEnablePodPerceiver:                      false,
		PerceiverImagePerceiverName:                      "",
		PerceiverImagePerceiverImage:                     "",
		PerceiverPodPerceiverName:                        "",
		PerceiverPodPerceiverImage:                       "",
		PerceiverPodPerceiverNamespaceFilter:             "",
		PerceiverAnnotationIntervalSeconds:               0,
		PerceiverDumpIntervalMinutes:                     0,
		PerceiverServiceAccount:                          "",
		PerceiverPort:                                    0,
		PrometheusName:                                   "",
		PrometheusImage:                                  "",
		PrometheusPort:                                   0,
		EnableSkyfire:                                    false,
		SkyfireName:                                      "",
		SkyfireImage:                                     "",
		SkyfirePort:                                      0,
		SkyfirePrometheusPort:                            0,
		SkyfireServiceAccount:                            "",
		SkyfireHubClientTimeoutSeconds:                   0,
		SkyfireHubDumpPauseSeconds:                       0,
		SkyfireKubeDumpIntervalSeconds:                   0,
		SkyfirePerceptorDumpIntervalSeconds:              0,
		BlackduckHosts:                                   []string{},
		BlackduckUser:                                    "",
		BlackduckPort:                                    0,
		BlackduckConcurrentScanLimit:                     0,
		BlackduckTotalScanLimit:                          0,
		BlackduckPasswordEnvVar:                          "",
		BlackduckInitialCount:                            0,
		BlackduckMaxCount:                                0,
		BlackduckDeleteHubThresholdPercentage:            0,
		EnableMetrics:                                    false,
		DefaultCPU:                                       "",
		DefaultMem:                                       "",
		LogLevel:                                         "",
		ConfigMapName:                                    "",
		SecretName:                                       "",
	}
}

// GetSpec returns the Spec for the resource
func (ctl *OpsSightCtl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *OpsSightCtl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(opssightv1.OpsSightSpec)
	if !ok {
		return fmt.Errorf("Error setting OpsSight Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *OpsSightCtl) CheckSpecFlags() error {
	for _, registryJSON := range ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice {
		registry := &opssightv1.RegistryAuth{}
		err := json.Unmarshal([]byte(registryJSON), registry)
		if err != nil {
			return fmt.Errorf("Invalid Registry Format")
		}
	}
	return nil
}

// SwitchSpec switches the OpsSight's Spec to a different predefined spec
func (ctl *OpsSightCtl) SwitchSpec(createOpsSightSpecType string) error {
	switch createOpsSightSpecType {
	case "empty":
		ctl.Spec = &opssightv1.OpsSightSpec{}
	case "disabledBlackduck":
		ctl.Spec = crddefaults.GetOpsSightDefaultValueWithDisabledHub()
	case "default":
		ctl.Spec = crddefaults.GetOpsSightDefaultValue()
	default:
		return fmt.Errorf("OpsSight Spec Type %s does not match: empty, disabledBlackduck, default", createOpsSightSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for the OpsSight's Spec to the command
func (ctl *OpsSightCtl) AddSpecFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctl.PerceptorName, "spec-perceptor-name", ctl.PerceptorName, "TODO")
	cmd.Flags().StringVar(&ctl.PerceptorImage, "spec-perceptor-image", ctl.PerceptorImage, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorPort, "spec-perceptor-port", ctl.PerceptorPort, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "spec-perceptor-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "spec-perceptor-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "spec-perceptor-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "spec-perceptor-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "spec-perceptor-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodName, "spec-scannerpod-name", ctl.ScannerPodName, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerName, "spec-scannerpod-scanner-name", ctl.ScannerPodScannerName, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "spec-scannerpod-scanner-image", ctl.ScannerPodScannerImage, "TODO")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerPort, "spec-scannerpod-scanner-port", ctl.ScannerPodScannerPort, "TODO")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "spec-scannerpod-scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeName, "spec-scannerpod-imagefacade-name", ctl.ScannerPodImageFacadeName, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "spec-scannerpod-imagefacade-image", ctl.ScannerPodImageFacadeImage, "TODO")
	cmd.Flags().IntVar(&ctl.ScannerPodImageFacadePort, "spec-scannerpod-imagefacade-port", ctl.ScannerPodImageFacadePort, "TODO")
	cmd.Flags().StringSliceVar(&ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "spec-scannerpod-imagefacade-internal-registries", ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "spec-scannerpod-imagefacade-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "spec-scannerpod-imagefacade-service-account", ctl.ScannerPodImageFacadeServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "spec-scannerpod-replica-count", ctl.ScannerPodReplicaCount, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "spec-scannerpod-image-directory", ctl.ScannerPodImageDirectory, "TODO")
	cmd.Flags().BoolVar(&ctl.PerceiverEnableImagePerceiver, "spec-enable-image-perceiver", ctl.PerceiverEnableImagePerceiver, "TODO")
	cmd.Flags().BoolVar(&ctl.PerceiverEnablePodPerceiver, "spec-enable-pod-perceiver", ctl.PerceiverEnablePodPerceiver, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverName, "spec-imageperceiver-name", ctl.PerceiverImagePerceiverName, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverImage, "spec-imageperceiver-image", ctl.PerceiverImagePerceiverImage, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverName, "spec-podperceiver-name", ctl.PerceiverPodPerceiverName, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverImage, "spec-podperceiver-image", ctl.PerceiverPodPerceiverImage, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverNamespaceFilter, "spec-podperceiver-namespace-filter", ctl.PerceiverPodPerceiverNamespaceFilter, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverAnnotationIntervalSeconds, "spec-perceiver-annotation-interval-seconds", ctl.PerceiverAnnotationIntervalSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverDumpIntervalMinutes, "spec-perceiver-dump-interval-minutes", ctl.PerceiverDumpIntervalMinutes, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverServiceAccount, "spec-perceiver-service-account", ctl.PerceiverServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverPort, "spec-perceiver-port", ctl.PerceiverPort, "TODO")
	cmd.Flags().StringVar(&ctl.PrometheusName, "spec-prometheus-name", ctl.PrometheusName, "TODO")
	cmd.Flags().StringVar(&ctl.PrometheusName, "spec-prometheus-image", ctl.PrometheusName, "TODO")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "spec-prometheus-port", ctl.PrometheusPort, "TODO")
	cmd.Flags().BoolVar(&ctl.EnableSkyfire, "spec-enable-skyfire", ctl.EnableSkyfire, "TODO")
	cmd.Flags().StringVar(&ctl.SkyfireName, "spec-skyfire-name", ctl.SkyfireName, "TODO")
	cmd.Flags().StringVar(&ctl.SkyfireImage, "spec-skyfire-image", ctl.SkyfireImage, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfirePort, "spec-skyfire-port", ctl.SkyfirePort, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfirePrometheusPort, "spec-skyfire-prometheus-port", ctl.SkyfirePrometheusPort, "TODO")
	cmd.Flags().StringVar(&ctl.SkyfireServiceAccount, "spec-skyfire-service-account", ctl.SkyfireServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfireHubClientTimeoutSeconds, "spec-skyfire-hub-client-timeout-seconds", ctl.SkyfireHubClientTimeoutSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfireHubDumpPauseSeconds, "spec-skyfire-hub-dump-pause-seconds", ctl.SkyfireHubDumpPauseSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfireKubeDumpIntervalSeconds, "spec-skyfire-kube-dump-interval-seconds", ctl.SkyfireKubeDumpIntervalSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfirePerceptorDumpIntervalSeconds, "spec-skyfire-perceptor-dump-interval-seconds", ctl.SkyfirePerceptorDumpIntervalSeconds, "TODO")
	cmd.Flags().StringSliceVar(&ctl.BlackduckHosts, "spec-blackduck-hosts", ctl.BlackduckHosts, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckUser, "spec-blackduck-user", ctl.BlackduckUser, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckPort, "spec-blackduck-port", ctl.BlackduckPort, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckConcurrentScanLimit, "spec-blackduck-concurrent-scan-limit", ctl.BlackduckConcurrentScanLimit, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckTotalScanLimit, "spec-blackduck-total-scan-limit", ctl.BlackduckTotalScanLimit, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckPasswordEnvVar, "spec-blackduck-password-environment-variable", ctl.BlackduckPasswordEnvVar, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "spec-blackduck-initial-count", ctl.BlackduckInitialCount, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "spec-blackduck-max-count", ctl.BlackduckMaxCount, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckDeleteHubThresholdPercentage, "spec-blackduck-delete-blackduck-threshold-percentage", ctl.BlackduckDeleteHubThresholdPercentage, "TODO")
	cmd.Flags().BoolVar(&ctl.EnableMetrics, "spec-enable-metrics", ctl.EnableMetrics, "TODO")
	cmd.Flags().StringVar(&ctl.DefaultCPU, "spec-default-cpu", ctl.DefaultCPU, "TODO")
	cmd.Flags().StringVar(&ctl.DefaultMem, "spec-default-mem", ctl.DefaultMem, "TODO")
	cmd.Flags().StringVar(&ctl.LogLevel, "spec-log-level", ctl.LogLevel, "TODO")
	cmd.Flags().StringVar(&ctl.ConfigMapName, "spec-config-map-name", ctl.ConfigMapName, "TODO")
	cmd.Flags().StringVar(&ctl.SecretName, "spec-secret-name", ctl.SecretName, "TODO")
}

// SetFlags sets the OpsSight's Spec if a flag was changed
func (ctl *OpsSightCtl) SetFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s:   CHANGED\n", f.Name)
		switch f.Name {
		case "spec-perceptor-name":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Name = ctl.PerceptorName
		case "spec-perceptor-image":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Image = ctl.PerceptorImage
		case "spec-perceptor-port":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Port = ctl.PerceptorPort
		case "spec-perceptor-check-scan-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.CheckForStalledScansPauseHours = ctl.PerceptorCheckForStalledScansPauseHours
		case "spec-perceptor-scan-client-timeout-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.StalledScanClientTimeoutHours = ctl.PerceptorStalledScanClientTimeoutHours
		case "spec-perceptor-metrics-pause-seconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ModelMetricsPauseSeconds = ctl.PerceptorModelMetricsPauseSeconds
		case "spec-perceptor-unknown-image-pause-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.UnknownImagePauseMilliseconds = ctl.PerceptorUnknownImagePauseMilliseconds
		case "spec-perceptor-client-timeout-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ClientTimeoutMilliseconds = ctl.PerceptorClientTimeoutMilliseconds
		case "spec-scannerpod-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.Name = ctl.ScannerPodName
		case "spec-scannerpod-scanner-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Name = ctl.ScannerPodScannerName
		case "spec-scannerpod-scanner-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Image = ctl.ScannerPodScannerImage
		case "spec-scannerpod-scanner-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Port = ctl.ScannerPodScannerPort
		case "spec-scannerpod-scanner-client-timeout-seconds":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.ClientTimeoutSeconds = ctl.ScannerPodScannerClientTimeoutSeconds
		case "spec-scannerpod-imagefacade-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Name = ctl.ScannerPodImageFacadeName
		case "spec-scannerpod-imagefacade-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Image = ctl.ScannerPodImageFacadeImage
		case "spec-scannerpod-imagefacade-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Port = ctl.ScannerPodImageFacadePort
		case "spec-scannerpod-imagefacade-internal-registries":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice {
				registry := &opssightv1.RegistryAuth{}
				json.Unmarshal([]byte(registryJSON), registry)
				ctl.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ctl.Spec.ScannerPod.ImageFacade.InternalRegistries, *registry)
			}
		case "spec-scannerpod-imagefacade-image-puller-type":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ImagePullerType = ctl.ScannerPodImageFacadeImagePullerType
		case "spec-scannerpod-imagefacade-service-account":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ServiceAccount = ctl.ScannerPodImageFacadeServiceAccount
		case "spec-scannerpod-replica-count":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ReplicaCount = ctl.ScannerPodReplicaCount
		case "spec-scannerpod-image-directory":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ImageDirectory = ctl.ScannerPodImageDirectory
		case "spec-enable-image-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnableImagePerceiver = ctl.PerceiverEnableImagePerceiver
		case "spec-enable-pod-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnablePodPerceiver = ctl.PerceiverEnablePodPerceiver
		case "spec-imageperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Name = ctl.PerceiverImagePerceiverName
		case "spec-imageperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Image = ctl.PerceiverImagePerceiverImage
		case "spec-podperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Name = ctl.PerceiverPodPerceiverName
		case "spec-podperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Image = ctl.PerceiverPodPerceiverImage
		case "spec-podperceiver-namespace-filter":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.NamespaceFilter = ctl.PerceiverPodPerceiverNamespaceFilter
		case "spec-perceiver-annotation-interval-seconds":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.AnnotationIntervalSeconds = ctl.PerceiverAnnotationIntervalSeconds
		case "spec-perceiver-dump-interval-minutes":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.DumpIntervalMinutes = ctl.PerceiverDumpIntervalMinutes
		case "spec-perceiver-service-account":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.ServiceAccount = ctl.PerceiverServiceAccount
		case "spec-perceiver-port":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.Port = ctl.PerceiverPort
		case "spec-prometheus-name":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Name = ctl.PrometheusName
		case "spec-prometheus-image":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Image = ctl.PrometheusImage
		case "spec-prometheus-port":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Port = ctl.PrometheusPort
		case "spec-enable-skyfire":
			ctl.Spec.EnableSkyfire = ctl.EnableSkyfire
		case "spec-skyfire-name":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Name = ctl.SkyfireName
		case "spec-skyfire-image":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Image = ctl.SkyfireImage
		case "spec-skyfire-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Port = ctl.SkyfirePort
		case "spec-skyfire-prometheus-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PrometheusPort = ctl.SkyfirePrometheusPort
		case "spec-skyfire-service-account":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.ServiceAccount = ctl.SkyfireServiceAccount
		case "spec-skyfire-hub-client-timeout-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubClientTimeoutSeconds = ctl.SkyfireHubClientTimeoutSeconds
		case "spec-skyfire-hub-dump-pause-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubDumpPauseSeconds = ctl.SkyfireHubDumpPauseSeconds
		case "spec-skyfire-kube-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.KubeDumpIntervalSeconds = ctl.SkyfireKubeDumpIntervalSeconds
		case "spec-skyfire-perceptor-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PerceptorDumpIntervalSeconds = ctl.SkyfirePerceptorDumpIntervalSeconds
		case "spec-blackduck-hosts":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.Hosts = ctl.BlackduckHosts
		case "spec-blackduck-user":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.User = ctl.BlackduckUser
		case "spec-blackduck-port":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.Port = ctl.BlackduckPort
		case "spec-blackduck-concurrent-scan-limit":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.ConcurrentScanLimit = ctl.BlackduckConcurrentScanLimit
		case "spec-blackduck-total-scan-limit":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.TotalScanLimit = ctl.BlackduckTotalScanLimit
		case "spec-blackduck-password-environment-variable":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.PasswordEnvVar = ctl.BlackduckPasswordEnvVar
		case "spec-blackduck-initial-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.InitialCount = ctl.BlackduckInitialCount
		case "spec-blackduck-max-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.MaxCount = ctl.BlackduckMaxCount
		case "spec-blackduck-delete-blackduck-threshold-percentage":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.DeleteHubThresholdPercentage = ctl.BlackduckDeleteHubThresholdPercentage
		case "spec-enable-metrics":
			ctl.Spec.EnableMetrics = ctl.EnableMetrics
		case "spec-default-cpu":
			ctl.Spec.DefaultCPU = ctl.DefaultCPU
		case "spec-default-mem":
			ctl.Spec.DefaultMem = ctl.DefaultMem
		case "spec-log-level":
			ctl.Spec.LogLevel = ctl.LogLevel
		case "spec-config-map-name":
			ctl.Spec.ConfigMapName = ctl.ConfigMapName
		case "spec-secret-name":
			ctl.Spec.SecretName = ctl.SecretName
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}
