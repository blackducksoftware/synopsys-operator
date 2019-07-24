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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewCRSpecBuilderFromCobraFlags(t *testing.T) {
	assert := assert.New(t)
	opsSightCobraHelper := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(&CRSpecBuilderFromCobraFlags{
		opsSightSpec: &opssightapi.OpsSightSpec{},
	}, opsSightCobraHelper)
}

func TestGetCRSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCobraHelper := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(opssightapi.OpsSightSpec{}, opsSightCobraHelper.GetCRSpec())
}

func TestSetCRSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCobraHelper := NewCRSpecBuilderFromCobraFlags()
	specToSet := opssightapi.OpsSightSpec{Namespace: "test"}
	opsSightCobraHelper.SetCRSpec(specToSet)
	assert.Equal(specToSet, opsSightCobraHelper.GetCRSpec())

	// check for error
	assert.Error(opsSightCobraHelper.SetCRSpec(""))
}

func TestCheckValuesFromFlags(t *testing.T) {
	assert := assert.New(t)
	opsSightCobraHelper := NewCRSpecBuilderFromCobraFlags()
	opsSightCobraHelper.PerceptorExpose = util.NONE
	opsSightCobraHelper.PrometheusExpose = util.NONE
	cmd := &cobra.Command{}
	specFlags := opsSightCobraHelper.CheckValuesFromFlags(cmd.Flags())
	assert.Nil(specFlags)

	var tests = []struct {
		input          *CRSpecBuilderFromCobraFlags
		flagNameToTest string
		flagValue      string
	}{
		// invalid opssight core expose case
		{input: &CRSpecBuilderFromCobraFlags{
			opsSightSpec:     &opssightapi.OpsSightSpec{},
			PerceptorExpose:  "",
			PrometheusExpose: util.NONE,
		},
			flagNameToTest: "opssight-core-expose",
			flagValue:      "",
		},
		// invalid prometheus metrics expose case
		{input: &CRSpecBuilderFromCobraFlags{
			opsSightSpec:     &opssightapi.OpsSightSpec{},
			PerceptorExpose:  util.NONE,
			PrometheusExpose: "",
		},
			flagNameToTest: "expose-metrics",
			flagValue:      "",
		},
	}

	for _, test := range tests {
		cmd := &cobra.Command{}
		opsSightCobraHelper.AddCRSpecFlagsToCommand(cmd, true)
		flagset := cmd.Flags()
		flagset.Set(test.flagNameToTest, test.flagValue)
		err := test.input.CheckValuesFromFlags(flagset)
		if err == nil {
			t.Errorf("Expected an error but got nil, test: %+v", test)
		}
	}
}

func TestSetPredefinedCRSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCobraHelper := NewCRSpecBuilderFromCobraFlags()
	defaultSpec := util.GetOpsSightDefault()
	defaultSpec.Perceiver.EnablePodPerceiver = true
	defaultSpec.EnableMetrics = true

	var tests = []struct {
		input    string
		expected *opssightapi.OpsSightSpec
	}{
		{input: EmptySpec, expected: &opssightapi.OpsSightSpec{}},
		{input: UpstreamSpec, expected: util.GetOpsSightUpstream()},
		{input: DefaultSpec, expected: defaultSpec},
		{input: DisabledBlackDuckSpec, expected: util.GetOpsSightDefaultWithIPV6DisabledBlackDuck()},
	}

	// test cases: "empty", "default", "disabledBlackduck"
	for _, test := range tests {
		assert.Nil(opsSightCobraHelper.SetPredefinedCRSpec(test.input))
		assert.Equal(*test.expected, opsSightCobraHelper.GetCRSpec())
	}

	// test cases: default
	createOpsSightSpecType := ""
	assert.Error(opsSightCobraHelper.SetPredefinedCRSpec(createOpsSightSpecType))

}

func TestAddCRSpecFlagsToCommand(t *testing.T) {
	assert := assert.New(t)

	ctl := NewCRSpecBuilderFromCobraFlags()
	actualCmd := &cobra.Command{}
	ctl.AddCRSpecFlagsToCommand(actualCmd, true)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.PerceptorImage, "opssight-core-image", ctl.PerceptorImage, "Image of OpsSight's Core")
	cmd.Flags().StringVar(&ctl.PerceptorExpose, "opssight-core-expose", ctl.PerceptorExpose, "Type of service for OpsSight's core model (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
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
	cmd.Flags().StringVar(&ctl.PrometheusExpose, "expose-metrics", ctl.PrometheusExpose, "Type of service of OpsSight's Prometheus Metrics (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	cmd.Flags().StringVar(&ctl.BlackduckExternalHostsFilePath, "blackduck-external-hosts-file-path", ctl.BlackduckExternalHostsFilePath, "Absolute path to a file containing a list of Black Duck External Hosts")
	cmd.Flags().StringVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "If true, OpsSight performs TLS Verification for Black Duck (true|false)")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Black Duck instances to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Black Duck instances that can be created")
	cmd.Flags().StringVar(&ctl.BlackduckType, "blackduck-type", ctl.BlackduckType, "Type of Black Duck")
	cmd.Flags().StringVar(&ctl.BlackduckPassword, "blackduck-password", ctl.BlackduckPassword, "Password to use for all internal Blackduck 'sysadmin' account")

	assert.Equal(cmd.Flags(), actualCmd.Flags())

}

func TestGenerateCRSpecFromFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewCRSpecBuilderFromCobraFlags()
	cmd := &cobra.Command{}
	actualCtl.AddCRSpecFlagsToCommand(cmd, true)
	actualCtl.GenerateCRSpecFromFlags(cmd.Flags())

	expCtl := NewCRSpecBuilderFromCobraFlags()

	assert.Equal(expCtl.opsSightSpec, actualCtl.opsSightSpec)

}

func TestSetCRSpecFieldByFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *CRSpecBuilderFromCobraFlags
		changedCtl  *CRSpecBuilderFromCobraFlags
		changedSpec *opssightapi.OpsSightSpec
	}{
		// case
		{
			flagName:   "opssight-core-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:   &opssightapi.OpsSightSpec{},
				PerceptorImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{Image: "changed"}},
		},
		// case
		{
			flagName:   "opssight-core-expose",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:    &opssightapi.OpsSightSpec{},
				PerceptorExpose: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{Expose: "changed"}},
		},
		// case
		{
			flagName:   "opssight-core-check-scan-hours",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                            &opssightapi.OpsSightSpec{},
				PerceptorCheckForStalledScansPauseHours: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{CheckForStalledScansPauseHours: 10}},
		},
		// case
		{
			flagName:   "opssight-core-scan-client-timeout-hours",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                           &opssightapi.OpsSightSpec{},
				PerceptorStalledScanClientTimeoutHours: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{StalledScanClientTimeoutHours: 10}},
		},
		// case
		{
			flagName:   "opssight-core-metrics-pause-seconds",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                      &opssightapi.OpsSightSpec{},
				PerceptorModelMetricsPauseSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{ModelMetricsPauseSeconds: 10}},
		},
		// case
		{
			flagName:   "opssight-core-unknown-image-pause-milliseconds",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                           &opssightapi.OpsSightSpec{},
				PerceptorUnknownImagePauseMilliseconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{UnknownImagePauseMilliseconds: 10}},
		},
		// case
		{
			flagName:   "opssight-core-client-timeout-milliseconds",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                       &opssightapi.OpsSightSpec{},
				PerceptorClientTimeoutMilliseconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceptor: &opssightapi.Perceptor{ClientTimeoutMilliseconds: 10}},
		},
		// case
		{
			flagName:   "scanner-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:           &opssightapi.OpsSightSpec{},
				ScannerPodScannerImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{Scanner: &opssightapi.Scanner{Image: "changed"}}},
		},
		// case
		{
			flagName:   "scanner-client-timeout-seconds",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                          &opssightapi.OpsSightSpec{},
				ScannerPodScannerClientTimeoutSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{Scanner: &opssightapi.Scanner{ClientTimeoutSeconds: 10}}},
		},
		// case
		{
			flagName:   "image-getter-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:               &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{Image: "changed"}}},
		},
		// case
		{
			flagName:   "image-getter-secure-registries-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeInternalRegistriesFilePath: "../../examples/synopsysctl/imageGetterSecureRegistries.json",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{InternalRegistries: []*opssightapi.RegistryAuth{
				{
					URL:      "url",
					User:     "user",
					Password: "password",
				},
			}}}},
		},
		// case
		{
			flagName:   "image-getter-image-puller-type",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                         &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeImagePullerType: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{ImagePullerType: "changed"}}},
		},
		// case
		{
			flagName:   "image-getter-service-account",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                        &opssightapi.OpsSightSpec{},
				ScannerPodImageFacadeServiceAccount: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageFacade: &opssightapi.ImageFacade{ServiceAccount: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-replica-count",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:           &opssightapi.OpsSightSpec{},
				ScannerPodReplicaCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ReplicaCount: 10}},
		},
		// case
		{
			flagName:   "scannerpod-image-directory",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:             &opssightapi.OpsSightSpec{},
				ScannerPodImageDirectory: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerPod: &opssightapi.ScannerPod{ImageDirectory: "changed"}},
		},
		// case
		{
			flagName:   "enable-pod-processor",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                &opssightapi.OpsSightSpec{},
				PerceiverEnablePodPerceiver: "false",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{EnablePodPerceiver: false}},
		},
		// case
		{
			flagName:   "enable-pod-processor",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                &opssightapi.OpsSightSpec{},
				PerceiverEnablePodPerceiver: "true",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{EnablePodPerceiver: true}},
		},
		// case
		{
			flagName:   "image-processor-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                 &opssightapi.OpsSightSpec{},
				PerceiverImagePerceiverImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{ImagePerceiver: &opssightapi.ImagePerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "pod-processor-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:               &opssightapi.OpsSightSpec{},
				PerceiverPodPerceiverImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{PodPerceiver: &opssightapi.PodPerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "pod-processor-namespace-filter",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                         &opssightapi.OpsSightSpec{},
				PerceiverPodPerceiverNamespaceFilter: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{PodPerceiver: &opssightapi.PodPerceiver{NamespaceFilter: "changed"}}},
		},
		// case
		{
			flagName:   "processor-annotation-interval-seconds",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                       &opssightapi.OpsSightSpec{},
				PerceiverAnnotationIntervalSeconds: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{AnnotationIntervalSeconds: 10}},
		},
		// case
		{
			flagName:   "processor-dump-interval-minutes",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                 &opssightapi.OpsSightSpec{},
				PerceiverDumpIntervalMinutes: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Perceiver: &opssightapi.Perceiver{DumpIntervalMinutes: 10}},
		},
		// case
		{
			flagName:   "default-cpu",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				DefaultCPU:   "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{DefaultCPU: "changed"},
		},
		// case
		{
			flagName:   "default-memory",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				DefaultMem:   "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{DefaultMem: "changed"},
		},
		// case
		{
			flagName:   "scanner-cpu",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				ScannerCPU:   "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerCPU: "changed"},
		},
		// case
		{
			flagName:   "scanner-memory",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				ScannerMem:   "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{ScannerMem: "changed"},
		},
		// case
		{
			flagName:   "log-level",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec: &opssightapi.OpsSightSpec{},
				LogLevel:     "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{LogLevel: "changed"},
		},
		// case
		{
			flagName:   "enable-metrics",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:  &opssightapi.OpsSightSpec{},
				EnableMetrics: "false",
			},
			changedSpec: &opssightapi.OpsSightSpec{EnableMetrics: false},
		},
		// case
		{
			flagName:   "enable-metrics",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:  &opssightapi.OpsSightSpec{},
				EnableMetrics: "true",
			},
			changedSpec: &opssightapi.OpsSightSpec{EnableMetrics: true},
		},
		// case
		{
			flagName:   "metrics-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:    &opssightapi.OpsSightSpec{},
				PrometheusImage: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Image: "changed"}},
		},
		// case
		{
			flagName:   "metrics-port",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:   &opssightapi.OpsSightSpec{},
				PrometheusPort: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Port: 10}},
		},
		// case
		{
			flagName:   "expose-metrics",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:     &opssightapi.OpsSightSpec{},
				PrometheusExpose: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Prometheus: &opssightapi.Prometheus{Expose: "changed"}},
		},
		// case
		{
			flagName:   "blackduck-external-hosts-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:                   &opssightapi.OpsSightSpec{},
				BlackduckExternalHostsFilePath: "../../examples/synopsysctl/blackduckExternalHosts.json",
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{ExternalHosts: []*opssightapi.Host{
				{
					Scheme:              "scheme",
					Domain:              "domain",
					Port:                99,
					User:                "user",
					Password:            "password",
					ConcurrentScanLimit: 88,
				},
			}}},
		},
		// case
		{
			flagName:   "blackduck-TLS-verification",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:             &opssightapi.OpsSightSpec{},
				BlackduckTLSVerification: "false",
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{TLSVerification: false}},
		},
		// case
		{
			flagName:   "blackduck-TLS-verification",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:             &opssightapi.OpsSightSpec{},
				BlackduckTLSVerification: "true",
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{TLSVerification: true}},
		},
		// case
		{
			flagName:   "blackduck-initial-count",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:          &opssightapi.OpsSightSpec{},
				BlackduckInitialCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{InitialCount: 10}},
		},
		// case
		{
			flagName:   "blackduck-max-count",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:      &opssightapi.OpsSightSpec{},
				BlackduckMaxCount: 10,
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{MaxCount: 10}},
		},
		// case
		{
			flagName:   "blackduck-type",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				opsSightSpec:  &opssightapi.OpsSightSpec{},
				BlackduckType: "changed",
			},
			changedSpec: &opssightapi.OpsSightSpec{Blackduck: &opssightapi.Blackduck{BlackduckSpec: &blackduckapi.BlackduckSpec{Type: "changed"}}},
		},
	}

	// get the CRSpecBuilderFromCobraFlags's flags
	cmd := &cobra.Command{}
	actualCtl := NewCRSpecBuilderFromCobraFlags()
	actualCtl.AddCRSpecFlagsToCommand(cmd, true)
	flagset := cmd.Flags()

	for _, test := range tests {
		actualCtl = NewCRSpecBuilderFromCobraFlags()
		// check the Flag exists
		foundFlag := flagset.Lookup(test.flagName)
		if foundFlag == nil {
			t.Errorf("flag %s is not in the spec", test.flagName)
		}
		// check the correct CRSpecBuilderFromCobraFlags is used
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		// test setting a flag
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetCRSpecFieldByFlag(f)
		assert.Equal(test.changedSpec, actualCtl.opsSightSpec)
	}

	// case: nothing set if flag doesn't exist
	actualCtl = NewCRSpecBuilderFromCobraFlags()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	actualCtl.SetCRSpecFieldByFlag(f)
	assert.Equal(&opssightapi.OpsSightSpec{}, actualCtl.opsSightSpec)

}
