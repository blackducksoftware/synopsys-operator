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
	Spec                                                     *opssightv1.OpsSightSpec
	OpssightPerceptorName                                    string
	OpssightPerceptorImage                                   string
	OpssightPerceptorPort                                    int
	OpssightPerceptorCheckForStalledScansPauseHours          int
	OpssightPerceptorStalledScanClientTimeoutHours           int
	OpssightPerceptorModelMetricsPauseSeconds                int
	OpssightPerceptorUnknownImagePauseMilliseconds           int
	OpssightPerceptorClientTimeoutMilliseconds               int
	OpssightScannerPodName                                   string
	OpssightScannerPodScannerName                            string
	OpssightScannerPodScannerImage                           string
	OpssightScannerPodScannerPort                            int
	OpssightScannerPodScannerClientTimeoutSeconds            int
	OpssightScannerPodImageFacadeName                        string
	OpssightScannerPodImageFacadeImage                       string
	OpssightScannerPodImageFacadePort                        int
	OpssightScannerPodImageFacadeInternalRegistriesJSONSlice []string
	OpssightScannerPodImageFacadeImagePullerType             string
	OpssightScannerPodImageFacadeServiceAccount              string
	OpssightScannerPodReplicaCount                           int
	OpssightScannerPodImageDirectory                         string
	OpssightPerceiverEnableImagePerceiver                    bool
	OpssightPerceiverEnablePodPerceiver                      bool
	OpssightPerceiverImagePerceiverName                      string
	OpssightPerceiverImagePerceiverImage                     string
	OpssightPerceiverPodPerceiverName                        string
	OpssightPerceiverPodPerceiverImage                       string
	OpssightPerceiverPodPerceiverNamespaceFilter             string
	OpssightPerceiverAnnotationIntervalSeconds               int
	OpssightPerceiverDumpIntervalMinutes                     int
	OpssightPerceiverServiceAccount                          string
	OpssightPerceiverPort                                    int
	OpssightPrometheusName                                   string
	OpssightPrometheusImage                                  string
	OpssightPrometheusPort                                   int
	OpssightEnableSkyfire                                    bool
	OpssightSkyfireName                                      string
	OpssightSkyfireImage                                     string
	OpssightSkyfirePort                                      int
	OpssightSkyfirePrometheusPort                            int
	OpssightSkyfireServiceAccount                            string
	OpssightSkyfireHubClientTimeoutSeconds                   int
	OpssightSkyfireHubDumpPauseSeconds                       int
	OpssightSkyfireKubeDumpIntervalSeconds                   int
	OpssightSkyfirePerceptorDumpIntervalSeconds              int
	OpssightBlackduckHosts                                   []string
	OpssightBlackduckUser                                    string
	OpssightBlackduckPort                                    int
	OpssightBlackduckConcurrentScanLimit                     int
	OpssightBlackduckTotalScanLimit                          int
	OpssightBlackduckPasswordEnvVar                          string
	OpssightBlackduckInitialCount                            int
	OpssightBlackduckMaxCount                                int
	OpssightBlackduckDeleteHubThresholdPercentage            int
	OpssightEnableMetrics                                    bool
	OpssightDefaultCPU                                       string
	OpssightDefaultMem                                       string
	OpssightLogLevel                                         string
	OpssightConfigMapName                                    string
	OpssightSecretName                                       string
}

// NewOpsSightCtl creates a new OpsSightCtl struct
func NewOpsSightCtl() *OpsSightCtl {
	return &OpsSightCtl{
		Spec:                   &opssightv1.OpsSightSpec{},
		OpssightPerceptorName:  "",
		OpssightPerceptorImage: "",
		OpssightPerceptorPort:  0,
		OpssightPerceptorCheckForStalledScansPauseHours:          0,
		OpssightPerceptorStalledScanClientTimeoutHours:           0,
		OpssightPerceptorModelMetricsPauseSeconds:                0,
		OpssightPerceptorUnknownImagePauseMilliseconds:           0,
		OpssightPerceptorClientTimeoutMilliseconds:               0,
		OpssightScannerPodName:                                   "",
		OpssightScannerPodScannerName:                            "",
		OpssightScannerPodScannerImage:                           "",
		OpssightScannerPodScannerPort:                            0,
		OpssightScannerPodScannerClientTimeoutSeconds:            0,
		OpssightScannerPodImageFacadeName:                        "",
		OpssightScannerPodImageFacadeImage:                       "",
		OpssightScannerPodImageFacadePort:                        0,
		OpssightScannerPodImageFacadeInternalRegistriesJSONSlice: []string{},
		OpssightScannerPodImageFacadeImagePullerType:             "",
		OpssightScannerPodImageFacadeServiceAccount:              "",
		OpssightScannerPodReplicaCount:                           0,
		OpssightScannerPodImageDirectory:                         "",
		OpssightPerceiverEnableImagePerceiver:                    false,
		OpssightPerceiverEnablePodPerceiver:                      false,
		OpssightPerceiverImagePerceiverName:                      "",
		OpssightPerceiverImagePerceiverImage:                     "",
		OpssightPerceiverPodPerceiverName:                        "",
		OpssightPerceiverPodPerceiverImage:                       "",
		OpssightPerceiverPodPerceiverNamespaceFilter:             "",
		OpssightPerceiverAnnotationIntervalSeconds:               0,
		OpssightPerceiverDumpIntervalMinutes:                     0,
		OpssightPerceiverServiceAccount:                          "",
		OpssightPerceiverPort:                                    0,
		OpssightPrometheusName:                                   "",
		OpssightPrometheusImage:                                  "",
		OpssightPrometheusPort:                                   0,
		OpssightEnableSkyfire:                                    false,
		OpssightSkyfireName:                                      "",
		OpssightSkyfireImage:                                     "",
		OpssightSkyfirePort:                                      0,
		OpssightSkyfirePrometheusPort:                            0,
		OpssightSkyfireServiceAccount:                            "",
		OpssightSkyfireHubClientTimeoutSeconds:                   0,
		OpssightSkyfireHubDumpPauseSeconds:                       0,
		OpssightSkyfireKubeDumpIntervalSeconds:                   0,
		OpssightSkyfirePerceptorDumpIntervalSeconds:              0,
		OpssightBlackduckHosts:                                   []string{},
		OpssightBlackduckUser:                                    "",
		OpssightBlackduckPort:                                    0,
		OpssightBlackduckConcurrentScanLimit:                     0,
		OpssightBlackduckTotalScanLimit:                          0,
		OpssightBlackduckPasswordEnvVar:                          "",
		OpssightBlackduckInitialCount:                            0,
		OpssightBlackduckMaxCount:                                0,
		OpssightBlackduckDeleteHubThresholdPercentage:            0,
		OpssightEnableMetrics:                                    false,
		OpssightDefaultCPU:                                       "",
		OpssightDefaultMem:                                       "",
		OpssightLogLevel:                                         "",
		OpssightConfigMapName:                                    "",
		OpssightSecretName:                                       "",
	}
}

// GetSpec returns the Spec for the resource
func (ctl *OpsSightCtl) GetSpec() opssightv1.OpsSightSpec {
	return *ctl.Spec
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
	cmd.Flags().StringVar(&ctl.OpssightPerceptorName, "perceptor-name", ctl.OpssightPerceptorName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceptorImage, "perceptor-image", ctl.OpssightPerceptorImage, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorPort, "perceptor-port", ctl.OpssightPerceptorPort, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", ctl.OpssightPerceptorCheckForStalledScansPauseHours, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", ctl.OpssightPerceptorStalledScanClientTimeoutHours, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", ctl.OpssightPerceptorModelMetricsPauseSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", ctl.OpssightPerceptorUnknownImagePauseMilliseconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", ctl.OpssightPerceptorClientTimeoutMilliseconds, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodName, "scannerpod-name", ctl.OpssightScannerPodName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodScannerName, "scannerpod-scanner-name", ctl.OpssightScannerPodScannerName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodScannerImage, "scannerpod-scanner-image", ctl.OpssightScannerPodScannerImage, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightScannerPodScannerPort, "scannerpod-scanner-port", ctl.OpssightScannerPodScannerPort, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", ctl.OpssightScannerPodScannerClientTimeoutSeconds, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodImageFacadeName, "scannerpod-imagefacade-name", ctl.OpssightScannerPodImageFacadeName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodImageFacadeImage, "scannerpod-imagefacade-image", ctl.OpssightScannerPodImageFacadeImage, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightScannerPodImageFacadePort, "scannerpod-imagefacade-port", ctl.OpssightScannerPodImageFacadePort, "TODO")
	cmd.Flags().StringSliceVar(&ctl.OpssightScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", ctl.OpssightScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", ctl.OpssightScannerPodImageFacadeImagePullerType, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", ctl.OpssightScannerPodImageFacadeServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightScannerPodReplicaCount, "scannerpod-replica-count", ctl.OpssightScannerPodReplicaCount, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightScannerPodImageDirectory, "scannerpod-image-directory", ctl.OpssightScannerPodImageDirectory, "TODO")
	cmd.Flags().BoolVar(&ctl.OpssightPerceiverEnableImagePerceiver, "enable-image-perceiver", ctl.OpssightPerceiverEnableImagePerceiver, "TODO")
	cmd.Flags().BoolVar(&ctl.OpssightPerceiverEnablePodPerceiver, "enable-pod-perceiver", ctl.OpssightPerceiverEnablePodPerceiver, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverImagePerceiverName, "imageperceiver-name", ctl.OpssightPerceiverImagePerceiverName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverImagePerceiverImage, "imageperceiver-image", ctl.OpssightPerceiverImagePerceiverImage, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverPodPerceiverName, "podperceiver-name", ctl.OpssightPerceiverPodPerceiverName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverPodPerceiverImage, "podperceiver-image", ctl.OpssightPerceiverPodPerceiverImage, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", ctl.OpssightPerceiverPodPerceiverNamespaceFilter, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", ctl.OpssightPerceiverAnnotationIntervalSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", ctl.OpssightPerceiverDumpIntervalMinutes, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPerceiverServiceAccount, "perceiver-service-account", ctl.OpssightPerceiverServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPerceiverPort, "perceiver-port", ctl.OpssightPerceiverPort, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPrometheusName, "prometheus-name", ctl.OpssightPrometheusName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightPrometheusName, "prometheus-image", ctl.OpssightPrometheusName, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightPrometheusPort, "prometheus-port", ctl.OpssightPrometheusPort, "TODO")
	cmd.Flags().BoolVar(&ctl.OpssightEnableSkyfire, "enable-skyfire", ctl.OpssightEnableSkyfire, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightSkyfireName, "skyfire-name", ctl.OpssightSkyfireName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightSkyfireImage, "skyfire-image", ctl.OpssightSkyfireImage, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfirePort, "skyfire-port", ctl.OpssightSkyfirePort, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfirePrometheusPort, "skyfire-prometheus-port", ctl.OpssightSkyfirePrometheusPort, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightSkyfireServiceAccount, "skyfire-service-account", ctl.OpssightSkyfireServiceAccount, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", ctl.OpssightSkyfireHubClientTimeoutSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", ctl.OpssightSkyfireHubDumpPauseSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", ctl.OpssightSkyfireKubeDumpIntervalSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightSkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", ctl.OpssightSkyfirePerceptorDumpIntervalSeconds, "TODO")
	cmd.Flags().StringSliceVar(&ctl.OpssightBlackduckHosts, "blackduck-hosts", ctl.OpssightBlackduckHosts, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightBlackduckUser, "blackduck-user", ctl.OpssightBlackduckUser, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckPort, "blackduck-port", ctl.OpssightBlackduckPort, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckConcurrentScanLimit, "blackduck-concurrent-scan-limit", ctl.OpssightBlackduckConcurrentScanLimit, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckTotalScanLimit, "blackduck-total-scan-limit", ctl.OpssightBlackduckTotalScanLimit, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightBlackduckPasswordEnvVar, "blackduck-password-environment-variable", ctl.OpssightBlackduckPasswordEnvVar, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckInitialCount, "blackduck-initial-count", ctl.OpssightBlackduckInitialCount, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckMaxCount, "blackduck-max-count", ctl.OpssightBlackduckMaxCount, "TODO")
	cmd.Flags().IntVar(&ctl.OpssightBlackduckDeleteHubThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", ctl.OpssightBlackduckDeleteHubThresholdPercentage, "TODO")
	cmd.Flags().BoolVar(&ctl.OpssightEnableMetrics, "enable-metrics", ctl.OpssightEnableMetrics, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightDefaultCPU, "default-cpu", ctl.OpssightDefaultCPU, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightDefaultMem, "default-mem", ctl.OpssightDefaultMem, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightLogLevel, "log-level", ctl.OpssightLogLevel, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightConfigMapName, "config-map-name", ctl.OpssightConfigMapName, "TODO")
	cmd.Flags().StringVar(&ctl.OpssightSecretName, "secret-name", ctl.OpssightSecretName, "TODO")
}

// SetFlags sets the OpsSight's Spec if a flag was changed
func (ctl *OpsSightCtl) SetFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "perceptor-name":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Name = ctl.OpssightPerceptorName
		case "perceptor-image":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Image = ctl.OpssightPerceptorImage
		case "perceptor-port":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Port = ctl.OpssightPerceptorPort
		case "perceptor-check-scan-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.CheckForStalledScansPauseHours = ctl.OpssightPerceptorCheckForStalledScansPauseHours
		case "perceptor-scan-client-timeout-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.StalledScanClientTimeoutHours = ctl.OpssightPerceptorStalledScanClientTimeoutHours
		case "perceptor-metrics-pause-seconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ModelMetricsPauseSeconds = ctl.OpssightPerceptorModelMetricsPauseSeconds
		case "perceptor-unknown-image-pause-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.UnknownImagePauseMilliseconds = ctl.OpssightPerceptorUnknownImagePauseMilliseconds
		case "perceptor-client-timeout-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ClientTimeoutMilliseconds = ctl.OpssightPerceptorClientTimeoutMilliseconds
		case "scannerpod-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.Name = ctl.OpssightScannerPodName
		case "scannerpod-scanner-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Name = ctl.OpssightScannerPodScannerName
		case "scannerpod-scanner-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Image = ctl.OpssightScannerPodScannerImage
		case "scannerpod-scanner-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Port = ctl.OpssightScannerPodScannerPort
		case "scannerpod-scanner-client-timeout-seconds":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.ClientTimeoutSeconds = ctl.OpssightScannerPodScannerClientTimeoutSeconds
		case "scannerpod-imagefacade-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Name = ctl.OpssightScannerPodImageFacadeName
		case "scannerpod-imagefacade-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Image = ctl.OpssightScannerPodImageFacadeImage
		case "scannerpod-imagefacade-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Port = ctl.OpssightScannerPodImageFacadePort
		case "scannerpod-imagefacade-internal-registries":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range ctl.OpssightScannerPodImageFacadeInternalRegistriesJSONSlice {
				registry := &opssightv1.RegistryAuth{}
				json.Unmarshal([]byte(registryJSON), registry)
				ctl.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ctl.Spec.ScannerPod.ImageFacade.InternalRegistries, *registry)
			}
		case "scannerpod-imagefacade-image-puller-type":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ImagePullerType = ctl.OpssightScannerPodImageFacadeImagePullerType
		case "scannerpod-imagefacade-service-account":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ServiceAccount = ctl.OpssightScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ReplicaCount = ctl.OpssightScannerPodReplicaCount
		case "scannerpod-image-directory":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ImageDirectory = ctl.OpssightScannerPodImageDirectory
		case "enable-image-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnableImagePerceiver = ctl.OpssightPerceiverEnableImagePerceiver
		case "enable-pod-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnablePodPerceiver = ctl.OpssightPerceiverEnablePodPerceiver
		case "imageperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Name = ctl.OpssightPerceiverImagePerceiverName
		case "imageperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Image = ctl.OpssightPerceiverImagePerceiverImage
		case "podperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Name = ctl.OpssightPerceiverPodPerceiverName
		case "podperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Image = ctl.OpssightPerceiverPodPerceiverImage
		case "podperceiver-namespace-filter":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.NamespaceFilter = ctl.OpssightPerceiverPodPerceiverNamespaceFilter
		case "perceiver-annotation-interval-seconds":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.AnnotationIntervalSeconds = ctl.OpssightPerceiverAnnotationIntervalSeconds
		case "perceiver-dump-interval-minutes":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.DumpIntervalMinutes = ctl.OpssightPerceiverDumpIntervalMinutes
		case "perceiver-service-account":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.ServiceAccount = ctl.OpssightPerceiverServiceAccount
		case "perceiver-port":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.Port = ctl.OpssightPerceiverPort
		case "prometheus-name":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Name = ctl.OpssightPrometheusName
		case "prometheus-image":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Image = ctl.OpssightPrometheusImage
		case "prometheus-port":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Port = ctl.OpssightPrometheusPort
		case "enable-skyfire":
			ctl.Spec.EnableSkyfire = ctl.OpssightEnableSkyfire
		case "skyfire-name":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Name = ctl.OpssightSkyfireName
		case "skyfire-image":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Image = ctl.OpssightSkyfireImage
		case "skyfire-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Port = ctl.OpssightSkyfirePort
		case "skyfire-prometheus-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PrometheusPort = ctl.OpssightSkyfirePrometheusPort
		case "skyfire-service-account":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.ServiceAccount = ctl.OpssightSkyfireServiceAccount
		case "skyfire-hub-client-timeout-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubClientTimeoutSeconds = ctl.OpssightSkyfireHubClientTimeoutSeconds
		case "skyfire-hub-dump-pause-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubDumpPauseSeconds = ctl.OpssightSkyfireHubDumpPauseSeconds
		case "skyfire-kube-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.KubeDumpIntervalSeconds = ctl.OpssightSkyfireKubeDumpIntervalSeconds
		case "skyfire-perceptor-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PerceptorDumpIntervalSeconds = ctl.OpssightSkyfirePerceptorDumpIntervalSeconds
		case "blackduck-hosts":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.Hosts = ctl.OpssightBlackduckHosts
		case "blackduck-user":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.User = ctl.OpssightBlackduckUser
		case "blackduck-port":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.Port = ctl.OpssightBlackduckPort
		case "blackduck-concurrent-scan-limit":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.ConcurrentScanLimit = ctl.OpssightBlackduckConcurrentScanLimit
		case "blackduck-total-scan-limit":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.TotalScanLimit = ctl.OpssightBlackduckTotalScanLimit
		case "blackduck-password-environment-variable":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.PasswordEnvVar = ctl.OpssightBlackduckPasswordEnvVar
		case "blackduck-initial-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.InitialCount = ctl.OpssightBlackduckInitialCount
		case "blackduck-max-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.MaxCount = ctl.OpssightBlackduckMaxCount
		case "blackduck-delete-blackduck-threshold-percentage":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.DeleteHubThresholdPercentage = ctl.OpssightBlackduckDeleteHubThresholdPercentage
		case "enable-metrics":
			ctl.Spec.EnableMetrics = ctl.OpssightEnableMetrics
		case "default-cpu":
			ctl.Spec.DefaultCPU = ctl.OpssightDefaultCPU
		case "default-mem":
			ctl.Spec.DefaultMem = ctl.OpssightDefaultMem
		case "log-level":
			ctl.Spec.LogLevel = ctl.OpssightLogLevel
		case "config-map-name":
			ctl.Spec.ConfigMapName = ctl.OpssightConfigMapName
		case "secret-name":
			ctl.Spec.SecretName = ctl.OpssightSecretName
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)

}
