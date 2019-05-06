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

package opssight

import (
	"testing"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewOpsSightCtl(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	assert.Equal(&Ctl{
		Spec:                                            &opssightapi.OpsSightSpec{},
		PerceptorName:                                   "",
		PerceptorImage:                                  "",
		PerceptorPort:                                   0,
		PerceptorCheckForStalledScansPauseHours:         0,
		PerceptorStalledScanClientTimeoutHours:          0,
		PerceptorModelMetricsPauseSeconds:               0,
		PerceptorUnknownImagePauseMilliseconds:          0,
		PerceptorClientTimeoutMilliseconds:              0,
		ScannerPodName:                                  "",
		ScannerPodScannerName:                           "",
		ScannerPodScannerImage:                          "",
		ScannerPodScannerPort:                           0,
		ScannerPodScannerClientTimeoutSeconds:           0,
		ScannerPodImageFacadeName:                       "",
		ScannerPodImageFacadeImage:                      "",
		ScannerPodImageFacadePort:                       0,
		ScannerPodImageFacadeInternalRegistriesFilePath: "",
		ScannerPodImageFacadeImagePullerType:            "",
		ScannerPodImageFacadeServiceAccount:             "",
		ScannerPodReplicaCount:                          0,
		ScannerPodImageDirectory:                        "",
		PerceiverEnableImagePerceiver:                   false,
		PerceiverEnablePodPerceiver:                     false,
		PerceiverImagePerceiverName:                     "",
		PerceiverImagePerceiverImage:                    "",
		PerceiverPodPerceiverName:                       "",
		PerceiverPodPerceiverImage:                      "",
		PerceiverPodPerceiverNamespaceFilter:            "",
		PerceiverAnnotationIntervalSeconds:              0,
		PerceiverDumpIntervalMinutes:                    0,
		PerceiverServiceAccount:                         "",
		PerceiverPort:                                   0,
		ConfigMapName:                                   "",
		SecretName:                                      "",
		DefaultCPU:                                      "",
		DefaultMem:                                      "",
		ScannerCPU:                                      "",
		ScannerMem:                                      "",
		LogLevel:                                        "",
		EnableMetrics:                                   false,
		PrometheusName:                                  "",
		PrometheusImage:                                 "",
		PrometheusPort:                                  0,
		EnableSkyfire:                                   false,
		SkyfireName:                                     "",
		SkyfireImage:                                    "",
		SkyfirePort:                                     0,
		SkyfirePrometheusPort:                           0,
		SkyfireServiceAccount:                           "",
		SkyfireHubClientTimeoutSeconds:                  0,
		SkyfireHubDumpPauseSeconds:                      0,
		SkyfireKubeDumpIntervalSeconds:                  0,
		SkyfirePerceptorDumpIntervalSeconds:             0,
		BlackduckExternalHostsFilePath:                  "",
		BlackduckConnectionsEnvironmentVaraiableName:    "",
		BlackduckTLSVerification:                        false,
		BlackduckInitialCount:                           0,
		BlackduckMaxCount:                               0,
	}, opsSightCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	assert.Equal(opssightapi.OpsSightSpec{}, opsSightCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	specToSet := opssightapi.OpsSightSpec{Namespace: "test"}
	opsSightCtl.SetSpec(specToSet)
	assert.Equal(specToSet, opsSightCtl.GetSpec())

	// check for error
	assert.Error(opsSightCtl.SetSpec(""))
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	specFlags := opsSightCtl.CheckSpecFlags()
	assert.Nil(specFlags)
}

func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()

	var tests = []struct {
		input    string
		expected *opssightapi.OpsSightSpec
	}{
		{input: EmptySpec, expected: &opssightapi.OpsSightSpec{}},
		{input: UpstreamSpec, expected: crddefaults.GetOpsSightUpstream()},
		{input: DefaultSpec, expected: crddefaults.GetOpsSightDefault()},
		{input: DisabledBlackDuckSpec, expected: crddefaults.GetOpsSightDefaultWithIPV6DisabledBlackDuck()},
	}

	// test cases: "empty", "default", "disabledBlackduck"
	for _, test := range tests {
		assert.Nil(opsSightCtl.SwitchSpec(test.input))
		assert.Equal(*test.expected, opsSightCtl.GetSpec())
	}

	// test cases: default
	createOpsSightSpecType := ""
	assert.Error(opsSightCtl.SwitchSpec(createOpsSightSpecType))

}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	ctl := NewOpsSightCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd, true)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.PerceptorImage, "opssight-core", ctl.PerceptorImage, "Image of the OpsSight Core")
	cmd.Flags().StringVar(&ctl.PerceptorExpose, "opssight-core-expose", ctl.PerceptorExpose, "Expose the OpsSight Core model. Possible values are NODEPORT/LOADBALANCER/OPENSHIFT")
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "opssight-core-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "Hours the Percpetor waits between checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "opssight-core-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "Hours until the OpsSight Core stops checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "opssight-core-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "Perceptor metrics pause in seconds")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "opssight-core-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "OpsSight Core's unknown image pause in milliseconds")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "opssight-core-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "OpsSight Core's timeout for Black Duck Scan Client in seconds")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "scanner-image", ctl.ScannerPodScannerImage, "Scanner Container's image")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "Scanner timeout for Black Duck Scan Client in seconds")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "image-getter-image", ctl.ScannerPodImageFacadeImage, "Image Getter Container's image")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "image-getter-secure-registries-file-path", ctl.ScannerPodImageFacadeInternalRegistriesFilePath, "Absolute path to a file for secure docker registries credentials to pull the images for scan")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "image-getter-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "Type of Image Getter's Image Puller - docker, skopeo")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "image-getter-service-account", ctl.ScannerPodImageFacadeServiceAccount, "Service Account for the Image Getter")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.ScannerPodReplicaCount, "Number of Scan Containers")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.ScannerPodImageDirectory, "Directory in the Scanner Pod where images are stored for scanning")
	cmd.Flags().BoolVar(&ctl.PerceiverEnableImagePerceiver, "enable-image-processor", ctl.PerceiverEnableImagePerceiver, "Enables the Image Processor to discover images for scanning")
	cmd.Flags().BoolVar(&ctl.PerceiverEnablePodPerceiver, "enable-pod-processor", ctl.PerceiverEnablePodPerceiver, "Enables the Pod Processor to discover Pods for scanning")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverImage, "image-processor-image", ctl.PerceiverImagePerceiverImage, "Image of the Image Processor Container")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverImage, "pod-processor-image", ctl.PerceiverPodPerceiverImage, "Image of the Pod Processor Container")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverNamespaceFilter, "pod-processor-namespace-filter", ctl.PerceiverPodPerceiverNamespaceFilter, "Pod Processor's filter to scan pods by their namespace")
	cmd.Flags().IntVar(&ctl.PerceiverAnnotationIntervalSeconds, "processor-annotation-interval-seconds", ctl.PerceiverAnnotationIntervalSeconds, "Refresh interval to get latest scan results and apply to Pods and Images")
	cmd.Flags().IntVar(&ctl.PerceiverDumpIntervalMinutes, "processor-dump-interval-minutes", ctl.PerceiverDumpIntervalMinutes, "Minutes the Image Processor and Pod Processor wait between creating dumps of data/metrics")
	cmd.Flags().StringVar(&ctl.DefaultCPU, "default-cpu", ctl.DefaultCPU, "CPU size for the OpsSight")
	cmd.Flags().StringVar(&ctl.DefaultMem, "default-memory", ctl.DefaultMem, "Memory size for the OpsSight")
	cmd.Flags().StringVar(&ctl.ScannerCPU, "scanner-cpu", ctl.ScannerCPU, "CPU size for the OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.ScannerMem, "scanner-memory", ctl.ScannerMem, "Memory size for the OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.LogLevel, "log-level", ctl.LogLevel, "Log-level for OpsSight's logs")
	cmd.Flags().BoolVar(&ctl.EnableMetrics, "enable-metrics", ctl.EnableMetrics, "Enable recording of OpsSight's Prometheus Metrics")
	cmd.Flags().StringVar(&ctl.PrometheusImage, "metrics-image", ctl.PrometheusImage, "Image for OpsSight's Prometheus Metrics")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "metrics-port", ctl.PrometheusPort, "Port for OpsSight's Prometheus Metrics")
	cmd.Flags().StringVar(&ctl.PrometheusExpose, "metrics-expose", ctl.PrometheusExpose, "Expose the OpsSight's Prometheus Metrics. Possible values are NODEPORT/LOADBALANCER/OPENSHIFT")
	cmd.Flags().StringVar(&ctl.BlackduckExternalHostsFilePath, "blackduck-external-hosts-file-path", ctl.BlackduckExternalHostsFilePath, "Absolute path to a file containing a list of Black Duck External Hosts")
	cmd.Flags().BoolVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "Perform TLS Verification for Black Duck")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Black Ducks to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Black Ducks that can be created")
	cmd.Flags().StringVar(&ctl.BlackduckType, "blackduck-type", ctl.BlackduckType, "Type of Black Duck")

	assert.Equal(cmd.Flags(), actualCmd.Flags())

}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewOpsSightCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, true)
	actualCtl.SetChangedFlags(cmd.Flags())

	expCtl := NewOpsSightCtl()

	assert.Equal(expCtl.Spec, actualCtl.Spec)

}

func TestSetFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *Ctl
		changedCtl  *Ctl
		changedSpec *opssightapi.OpsSightSpec
	}{
		// case
		{
			flagName:   "opssight-core-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightapi.OpsSightSpec{},
				PerceptorImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{Image: "changed"}},
		},
		// case
		{
			flagName:   "opssight-core-expose",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:            &opssightapi.OpsSightSpec{},
				PerceptorExpose: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{Expose: "changed"}},
		},
		// case
		{
			flagName:   "opssight-core-check-scan-hours",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                    &opssightapi.OpsSightSpec{},
				PerceptorCheckForStalledScansPauseHours: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{CheckForStalledScansPauseHours: 10}},
		},
		// case
		{
			flagName:   "opssight-core-scan-client-timeout-hours",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                   &opssightapi.OpsSightSpec{},
				PerceptorStalledScanClientTimeoutHours: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{StalledScanClientTimeoutHours: 10}},
		},
		// case
		{
			flagName:   "opssight-core-metrics-pause-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                              &opssightapi.OpsSightSpec{},
				PerceptorModelMetricsPauseSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{ModelMetricsPauseSeconds: 10}},
		},
		// case
		{
			flagName:   "opssight-core-unknown-image-pause-milliseconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                   &opssightapi.OpsSightSpec{},
				PerceptorUnknownImagePauseMilliseconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{UnknownImagePauseMilliseconds: 10}},
		},
		// case
		{
			flagName:   "opssight-core-client-timeout-milliseconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                               &opssightapi.OpsSightSpec{},
				PerceptorClientTimeoutMilliseconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{ClientTimeoutMilliseconds: 10}},
		},
		// case
		{
			flagName:   "scanner-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                   &opssightapi.OpsSightSpec{},
				ScannerPodScannerImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{Scanner: &opssightapi.Scanner{Image: "changed"}}},
		},
		// case
		{
			flagName:   "scanner-client-timeout-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                  &opssightapi.OpsSightSpec{},
				ScannerPodScannerClientTimeoutSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{Scanner: &opssightapi.Scanner{ClientTimeoutSeconds: 10}}},
		},
		// case
		{
			flagName:   "image-getter-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{Image: "changed"}}},
		},
		// case
		{
			flagName:   "image-getter-image-puller-type",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                 &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeImagePullerType: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{ImagePullerType: "changed"}}},
		},
		// case
		{
			flagName:   "image-getter-service-account",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeServiceAccount: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{ServiceAccount: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-replica-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                   &opssightapi.OpsSightSpec{},
				ScannerPodReplicaCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ReplicaCount: 10}},
		},
		// case
		{
			flagName:   "scannerpod-image-directory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                     &opssightapi.OpsSightSpec{},
				ScannerPodImageDirectory: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageDirectory: "changed"}},
		},
		// case
		{
			flagName:   "enable-pod-processor",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                        &opssightapi.OpsSightSpec{},
				PerceiverEnablePodPerceiver: true,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{EnablePodPerceiver: true}},
		},
		// case
		{
			flagName:   "image-processor-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                         &opssightapi.OpsSightSpec{},
				PerceiverImagePerceiverImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{ImagePerceiver: &opssightapi.ImagePerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "pod-processor-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightapi.OpsSightSpec{},
				PerceiverPodPerceiverImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{PodPerceiver: &opssightapi.PodPerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "pod-processor-namespace-filter",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                 &opssightapi.OpsSightSpec{},
				PerceiverPodPerceiverNamespaceFilter: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{PodPerceiver: &opssightapi.PodPerceiver{NamespaceFilter: "changed"}}},
		},
		// case
		{
			flagName:   "processor-annotation-interval-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                               &opssightapi.OpsSightSpec{},
				PerceiverAnnotationIntervalSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{AnnotationIntervalSeconds: 10}},
		},
		// case
		{
			flagName:   "processor-dump-interval-minutes",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                         &opssightapi.OpsSightSpec{},
				PerceiverDumpIntervalMinutes: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{DumpIntervalMinutes: 10}},
		},
		// case
		{
			flagName:   "default-cpu",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightapi.OpsSightSpec{},
				DefaultCPU: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{DefaultCPU: "changed"},
		},
		// case
		{
			flagName:   "default-memory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightapi.OpsSightSpec{},
				DefaultMem: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{DefaultMem: "changed"},
		},
		// case
		{
			flagName:   "scanner-cpu",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightapi.OpsSightSpec{},
				ScannerCPU: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerCPU: "changed"},
		},
		// case
		{
			flagName:   "scanner-memory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightapi.OpsSightSpec{},
				ScannerMem: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerMem: "changed"},
		},
		// case
		{
			flagName:   "log-level",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:     &opssightapi.OpsSightSpec{},
				LogLevel: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{LogLevel: "changed"},
		},
		// case
		{
			flagName:   "enable-metrics",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightapi.OpsSightSpec{},
				EnableMetrics: true,
			},
			changedSpec: &opssightapi.OpsSightSpec{EnableMetrics: true},
		},
		// case
		{
			flagName:   "metrics-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:            &opssightapi.OpsSightSpec{},
				PrometheusImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Image: "changed"}},
		},
		// case
		{
			flagName:   "metrics-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightapi.OpsSightSpec{},
				PrometheusPort: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Port: 10}},
		},
		// case
		{
			flagName:   "metrics-expose",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:             &opssightapi.OpsSightSpec{},
				PrometheusExpose: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Expose: "changed"}},
		},
		// case
		{
			flagName:   "blackduck-TLS-verification",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                     &opssightapi.OpsSightSpec{},
				BlackduckTLSVerification: true,
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{TLSVerification: true}},
		},
		// case
		{
			flagName:   "blackduck-initial-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightapi.OpsSightSpec{},
				BlackduckInitialCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{InitialCount: 10}},
		},
		// case
		{
			flagName:   "blackduck-max-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:              &opssightapi.OpsSightSpec{},
				BlackduckMaxCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{MaxCount: 10}},
		},
		// case
		{
			flagName:   "blackduck-type",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightapi.OpsSightSpec{},
				BlackduckType: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{BlackduckSpec: &blackduckapi.BlackduckSpec{Type: "changed"}}},
		},
		// case
		{
			flagName:   "",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec: &opssightapi.OpsSightSpec{},
			},
			changedSpec: &opssightapi.OpsSightSpec{},
		},
	}

	for _, test := range tests {
		actualCtl := NewOpsSightCtl()
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetFlag(f)
		assert.Equal(test.changedSpec, actualCtl.Spec)
	}
}
