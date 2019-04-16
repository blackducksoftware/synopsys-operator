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

	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewOpsSightCtl(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	assert.Equal(&Ctl{
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
		ConfigMapName:                                    "",
		SecretName:                                       "",
		DefaultCPU:                                       "",
		DefaultMem:                                       "",
		ScannerCPU:                                       "",
		ScannerMem:                                       "",
		LogLevel:                                         "",
		EnableMetrics:                                    false,
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
		BlackduckExternalHostsJSON:                       []string{},
		BlackduckConnectionsEnvironmentVaraiableName:     "",
		BlackduckTLSVerification:                         false,
		BlackduckInitialCount:                            0,
		BlackduckMaxCount:                                0,
	}, opsSightCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	assert.Equal(opssightv1.OpsSightSpec{}, opsSightCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()
	specToSet := opssightv1.OpsSightSpec{Namespace: "test"}
	opsSightCtl.SetSpec(specToSet)
	assert.Equal(specToSet, opsSightCtl.GetSpec())

	// check for error
	assert.Error(opsSightCtl.SetSpec(""))
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)

	opsSightCtl := NewOpsSightCtl()
	assert.Nil(opsSightCtl.CheckSpecFlags())

	var tests = []struct {
		input *Ctl
	}{ // case
		{input: &Ctl{
			Spec: &opssightv1.OpsSightSpec{},
			ScannerPodImageFacadeInternalRegistriesJSONSlice: []string{"notValid"},
		}},
	}

	for _, test := range tests {
		assert.Error(test.input.CheckSpecFlags())
	}

}

func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	opsSightCtl := NewOpsSightCtl()

	var tests = []struct {
		input    string
		expected *opssightv1.OpsSightSpec
	}{
		{input: EmptySpec, expected: &opssightv1.OpsSightSpec{}},
		{input: TemplateSpec, expected: crddefaults.GetOpsSightTemplate()},
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
	cmd.Flags().StringVar(&ctl.PerceptorName, "perceptor-name", ctl.PerceptorName, "Name of the Perceptor")
	cmd.Flags().StringVar(&ctl.ScannerPodName, "scannerpod-name", ctl.ScannerPodName, "Name of the ScannerPod")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerName, "scannerpod-scanner-name", ctl.ScannerPodScannerName, "Name of the ScannerPod's Scanner Container")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeName, "scannerpod-imagefacade-name", ctl.ScannerPodImageFacadeName, "Name of the ScannerPod's ImageFacade Container")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverName, "imageperceiver-name", ctl.PerceiverImagePerceiverName, "Name of the ImagePerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverName, "podperceiver-name", ctl.PerceiverPodPerceiverName, "Name of the PodPerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverServiceAccount, "perceiver-service-account", ctl.PerceiverServiceAccount, "Name of the Service Account Resource for the Perceiver")
	cmd.Flags().StringVar(&ctl.PrometheusName, "prometheus-name", ctl.PrometheusName, "Name of Prometheus")
	cmd.Flags().StringVar(&ctl.SkyfireName, "skyfire-name", ctl.SkyfireName, "Name of Skyfire")
	cmd.Flags().StringVar(&ctl.SkyfireServiceAccount, "skyfire-service-account", ctl.SkyfireServiceAccount, "Service Account for Skyfire")
	cmd.Flags().StringVar(&ctl.BlackduckConnectionsEnvironmentVaraiableName, "blackduck-connections-environment-variable-name", ctl.BlackduckConnectionsEnvironmentVaraiableName, "Environment Variable name to store the Black Duck connections")
	cmd.Flags().StringVar(&ctl.ConfigMapName, "config-map-name", ctl.ConfigMapName, "Name of the config map for OpsSight")
	cmd.Flags().StringVar(&ctl.SecretName, "secret-name", ctl.SecretName, "Name of the Secret for OpsSight")
	cmd.Flags().StringVar(&ctl.PerceptorImage, "perceptor-image", ctl.PerceptorImage, "Image of the Perceptor")
	cmd.Flags().IntVar(&ctl.PerceptorPort, "perceptor-port", ctl.PerceptorPort, "Port for the Perceptor")
	cmd.Flags().StringVar(&ctl.PerceptorExpose, "perceptor-expose", ctl.PerceptorExpose, "Expose the Perceptor model. Possible values are NODEPORT/LOADBALANCER/OPENSHIFT")
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "Hours the Percpetor waits between checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "Hours until Perceptor stops checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "Perceptor metrics pause in seconds")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "Perceptor unknown image pause in milliseconds")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "Perceptor timeout for Black Duck Scan Client in seconds")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "scannerpod-scanner-image", ctl.ScannerPodScannerImage, "Scanner Container's image")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerPort, "scannerpod-scanner-port", ctl.ScannerPodScannerPort, "Scanner Container's port")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "Scanner timeout for Black Duck Scan Client in seconds")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "scannerpod-imagefacade-image", ctl.ScannerPodImageFacadeImage, "ImageFacade Container's image")
	cmd.Flags().IntVar(&ctl.ScannerPodImageFacadePort, "scannerpod-imagefacade-port", ctl.ScannerPodImageFacadePort, "ImageFacade Container's port")
	cmd.Flags().StringSliceVar(&ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "Secure Docker registries credentials to pull the images for scan")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "Type of ImageFacade's Image Puller - docker, skopeo")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", ctl.ScannerPodImageFacadeServiceAccount, "Service Account for the ImageFacade")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.ScannerPodReplicaCount, "Number of Scan Containers")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.ScannerPodImageDirectory, "Directory in Pod where images are stored for scanning")
	cmd.Flags().BoolVar(&ctl.PerceiverEnableImagePerceiver, "enable-image-perceiver", ctl.PerceiverEnableImagePerceiver, "Enables the perceiver to discover images for scanning")
	cmd.Flags().BoolVar(&ctl.PerceiverEnablePodPerceiver, "enable-pod-perceiver", ctl.PerceiverEnablePodPerceiver, "Enables the perceriver to discover Pods for scanning")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverImage, "imageperceiver-image", ctl.PerceiverImagePerceiverImage, "Image of the ImagePerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverImage, "podperceiver-image", ctl.PerceiverPodPerceiverImage, "Image of the PodPerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", ctl.PerceiverPodPerceiverNamespaceFilter, "Filter pods to scan by namespace")
	cmd.Flags().IntVar(&ctl.PerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", ctl.PerceiverAnnotationIntervalSeconds, "Refresh interval to get latest scan results and apply to Pods and Images")
	cmd.Flags().IntVar(&ctl.PerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", ctl.PerceiverDumpIntervalMinutes, "Minutes the perceiver waits between creating dumps of data/metrics")
	cmd.Flags().IntVar(&ctl.PerceiverPort, "perceiver-port", ctl.PerceiverPort, "Port for the Perceiver")
	cmd.Flags().StringVar(&ctl.DefaultCPU, "default-cpu", ctl.DefaultCPU, "CPU size for the OpsSight")
	cmd.Flags().StringVar(&ctl.DefaultMem, "default-memory", ctl.DefaultMem, "Memory size for the OpsSight")
	cmd.Flags().StringVar(&ctl.ScannerCPU, "scanner-cpu", ctl.ScannerCPU, "CPU size for the OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.ScannerMem, "scanner-memory", ctl.ScannerMem, "Memory size for the OpsSight's Scanner")
	cmd.Flags().StringVar(&ctl.LogLevel, "log-level", ctl.LogLevel, "Log-level for OpsSight's logs")
	cmd.Flags().BoolVar(&ctl.EnableMetrics, "enable-metrics", ctl.EnableMetrics, "Enable recording of Prometheus Metrics")
	cmd.Flags().StringVar(&ctl.PrometheusImage, "prometheus-image", ctl.PrometheusImage, "Image for Prometheus")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "prometheus-port", ctl.PrometheusPort, "Port for Prometheus")
	cmd.Flags().StringVar(&ctl.PrometheusExpose, "prometheus-expose", ctl.PrometheusExpose, "Expose the Prometheus metrics. Possible values are NODEPORT/LOADBALANCER/OPENSHIFT")
	cmd.Flags().BoolVar(&ctl.EnableSkyfire, "enable-skyfire", ctl.EnableSkyfire, "Enables Skyfire Pod if true")
	cmd.Flags().StringVar(&ctl.SkyfireImage, "skyfire-image", ctl.SkyfireImage, "Image of Skyfire")
	cmd.Flags().IntVar(&ctl.SkyfirePort, "skyfire-port", ctl.SkyfirePort, "Port of Skyfire")
	cmd.Flags().IntVar(&ctl.SkyfirePrometheusPort, "skyfire-prometheus-port", ctl.SkyfirePrometheusPort, "Skyfire's Prometheus port")
	cmd.Flags().IntVar(&ctl.SkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", ctl.SkyfireHubClientTimeoutSeconds, "Seconds Skyfire waits to receive response from the Black Duck client")
	cmd.Flags().IntVar(&ctl.SkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", ctl.SkyfireHubDumpPauseSeconds, "Seconds Skyfire waits between querying Black Ducks")
	cmd.Flags().IntVar(&ctl.SkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", ctl.SkyfireKubeDumpIntervalSeconds, "Seconds Skyfire waits between querying the KubeAPI")
	cmd.Flags().IntVar(&ctl.SkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", ctl.SkyfirePerceptorDumpIntervalSeconds, "Seconds Skyfire waits between querying the Perceptor Model")
	cmd.Flags().StringSliceVar(&ctl.BlackduckExternalHostsJSON, "blackduck-external-hosts", ctl.BlackduckExternalHostsJSON, "List of Black Duck External Hosts")
	cmd.Flags().BoolVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "Perform TLS Verification for Black Duck")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Black Ducks to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Black Ducks that can be created")

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
		changedSpec *opssightv1.OpsSightSpec
	}{
		// case
		{
			flagName:   "perceptor-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				PerceptorName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{Name: "changed"}},
		},
		// case
		{
			flagName:   "perceptor-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightv1.OpsSightSpec{},
				PerceptorImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{Image: "changed"}},
		},
		// case
		{
			flagName:   "perceptor-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				PerceptorPort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{Port: 10}},
		},
		// case
		{
			flagName:   "perceptor-expose",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:            &opssightv1.OpsSightSpec{},
				PerceptorExpose: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{Expose: "changed"}},
		},
		// case
		{
			flagName:   "perceptor-check-scan-hours",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                    &opssightv1.OpsSightSpec{},
				PerceptorCheckForStalledScansPauseHours: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{CheckForStalledScansPauseHours: 10}},
		},
		// case
		{
			flagName:   "perceptor-scan-client-timeout-hours",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                   &opssightv1.OpsSightSpec{},
				PerceptorStalledScanClientTimeoutHours: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{StalledScanClientTimeoutHours: 10}},
		},
		// case
		{
			flagName:   "perceptor-metrics-pause-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                              &opssightv1.OpsSightSpec{},
				PerceptorModelMetricsPauseSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{ModelMetricsPauseSeconds: 10}},
		},
		// case
		{
			flagName:   "perceptor-unknown-image-pause-milliseconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                   &opssightv1.OpsSightSpec{},
				PerceptorUnknownImagePauseMilliseconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{UnknownImagePauseMilliseconds: 10}},
		},
		// case
		{
			flagName:   "perceptor-client-timeout-milliseconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                               &opssightv1.OpsSightSpec{},
				PerceptorClientTimeoutMilliseconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceptor: &opssightv1.Perceptor{ClientTimeoutMilliseconds: 10}},
		},
		// case
		{
			flagName:   "scannerpod-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightv1.OpsSightSpec{},
				ScannerPodName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{Name: "changed"}},
		},
		// case
		{
			flagName:   "scannerpod-scanner-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightv1.OpsSightSpec{},
				ScannerPodScannerName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{Scanner: &opssightv1.Scanner{Name: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-scanner-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                   &opssightv1.OpsSightSpec{},
				ScannerPodScannerImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{Scanner: &opssightv1.Scanner{Image: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-scanner-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightv1.OpsSightSpec{},
				ScannerPodScannerPort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{Scanner: &opssightv1.Scanner{Port: 10}}},
		},
		// case
		{
			flagName:   "scannerpod-scanner-client-timeout-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                  &opssightv1.OpsSightSpec{},
				ScannerPodScannerClientTimeoutSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{Scanner: &opssightv1.Scanner{ClientTimeoutSeconds: 10}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                      &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadeName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{Name: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadeImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{Image: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                      &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadePort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{Port: 10}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-internal-registries",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec: &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadeInternalRegistriesJSONSlice: []string{"{\"URL\": \"changed\"}"},
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{InternalRegistries: []*opssightv1.RegistryAuth{{URL: "changed"}}}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-image-puller-type",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                 &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadeImagePullerType: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{ImagePullerType: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-imagefacade-service-account",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                &opssightv1.OpsSightSpec{},
				ScannerPodImageFacadeServiceAccount: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageFacade: &opssightv1.ImageFacade{ServiceAccount: "changed"}}},
		},
		// case
		{
			flagName:   "scannerpod-replica-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                   &opssightv1.OpsSightSpec{},
				ScannerPodReplicaCount: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ReplicaCount: 10}},
		},
		// case
		{
			flagName:   "scannerpod-image-directory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                     &opssightv1.OpsSightSpec{},
				ScannerPodImageDirectory: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerPod: &opssightv1.ScannerPod{ImageDirectory: "changed"}},
		},
		// case
		{
			flagName:   "enable-pod-perceiver",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                        &opssightv1.OpsSightSpec{},
				PerceiverEnablePodPerceiver: true,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{EnablePodPerceiver: true}},
		},
		// case
		{
			flagName:   "imageperceiver-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                        &opssightv1.OpsSightSpec{},
				PerceiverImagePerceiverName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{ImagePerceiver: &opssightv1.ImagePerceiver{Name: "changed"}}},
		},
		// case
		{
			flagName:   "imageperceiver-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                         &opssightv1.OpsSightSpec{},
				PerceiverImagePerceiverImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{ImagePerceiver: &opssightv1.ImagePerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "podperceiver-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                      &opssightv1.OpsSightSpec{},
				PerceiverPodPerceiverName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{PodPerceiver: &opssightv1.PodPerceiver{Name: "changed"}}},
		},
		// case
		{
			flagName:   "podperceiver-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightv1.OpsSightSpec{},
				PerceiverPodPerceiverImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{PodPerceiver: &opssightv1.PodPerceiver{Image: "changed"}}},
		},
		// case
		{
			flagName:   "podperceiver-namespace-filter",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                 &opssightv1.OpsSightSpec{},
				PerceiverPodPerceiverNamespaceFilter: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{PodPerceiver: &opssightv1.PodPerceiver{NamespaceFilter: "changed"}}},
		},
		// case
		{
			flagName:   "perceiver-annotation-interval-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                               &opssightv1.OpsSightSpec{},
				PerceiverAnnotationIntervalSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{AnnotationIntervalSeconds: 10}},
		},
		// case
		{
			flagName:   "perceiver-dump-interval-minutes",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                         &opssightv1.OpsSightSpec{},
				PerceiverDumpIntervalMinutes: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{DumpIntervalMinutes: 10}},
		},
		// case
		{
			flagName:   "perceiver-service-account",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                    &opssightv1.OpsSightSpec{},
				PerceiverServiceAccount: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{ServiceAccount: "changed"}},
		},
		// case
		{
			flagName:   "perceiver-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				PerceiverPort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Perceiver: &opssightv1.Perceiver{Port: 10}},
		},
		// case
		{
			flagName:   "config-map-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				ConfigMapName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ConfigMapName: "changed"},
		},
		// case
		{
			flagName:   "secret-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightv1.OpsSightSpec{},
				SecretName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{SecretName: "changed"},
		},
		// case
		{
			flagName:   "default-cpu",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightv1.OpsSightSpec{},
				DefaultCPU: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{DefaultCPU: "changed"},
		},
		// case
		{
			flagName:   "default-memory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightv1.OpsSightSpec{},
				DefaultMem: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{DefaultMem: "changed"},
		},
		// case
		{
			flagName:   "scanner-cpu",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightv1.OpsSightSpec{},
				ScannerCPU: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerCPU: "changed"},
		},
		// case
		{
			flagName:   "scanner-memory",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:       &opssightv1.OpsSightSpec{},
				ScannerMem: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{ScannerMem: "changed"},
		},
		// case
		{
			flagName:   "log-level",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:     &opssightv1.OpsSightSpec{},
				LogLevel: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{LogLevel: "changed"},
		},
		// case
		{
			flagName:   "enable-metrics",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				EnableMetrics: true,
			},
			changedSpec: &opssightv1.OpsSightSpec{EnableMetrics: true},
		},
		// case
		{
			flagName:   "prometheus-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightv1.OpsSightSpec{},
				PrometheusName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Prometheus: &opssightv1.Prometheus{Name: "changed"}},
		},
		// case
		{
			flagName:   "prometheus-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:            &opssightv1.OpsSightSpec{},
				PrometheusImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Prometheus: &opssightv1.Prometheus{Image: "changed"}},
		},
		// case
		{
			flagName:   "prometheus-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:           &opssightv1.OpsSightSpec{},
				PrometheusPort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Prometheus: &opssightv1.Prometheus{Port: 10}},
		},
		// case
		{
			flagName:   "prometheus-expose",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:             &opssightv1.OpsSightSpec{},
				PrometheusExpose: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Prometheus: &opssightv1.Prometheus{Expose: "changed"}},
		},
		// case
		{
			flagName:   "enable-skyfire",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:          &opssightv1.OpsSightSpec{},
				EnableSkyfire: true,
			},
			changedSpec: &opssightv1.OpsSightSpec{EnableSkyfire: true},
		},
		// case
		{
			flagName:   "skyfire-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:        &opssightv1.OpsSightSpec{},
				SkyfireName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{Name: "changed"}},
		},
		// case
		{
			flagName:   "skyfire-image",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:         &opssightv1.OpsSightSpec{},
				SkyfireImage: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{Image: "changed"}},
		},
		// case
		{
			flagName:   "skyfire-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:        &opssightv1.OpsSightSpec{},
				SkyfirePort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{Port: 10}},
		},
		// case
		{
			flagName:   "skyfire-prometheus-port",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightv1.OpsSightSpec{},
				SkyfirePrometheusPort: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{PrometheusPort: 10}},
		},
		// case
		{
			flagName:   "skyfire-service-account",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightv1.OpsSightSpec{},
				SkyfireServiceAccount: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{ServiceAccount: "changed"}},
		},
		// case
		{
			flagName:   "skyfire-hub-client-timeout-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                           &opssightv1.OpsSightSpec{},
				SkyfireHubClientTimeoutSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{HubClientTimeoutSeconds: 10}},
		},
		// case
		{
			flagName:   "skyfire-hub-dump-pause-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightv1.OpsSightSpec{},
				SkyfireHubDumpPauseSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{HubDumpPauseSeconds: 10}},
		},
		// case
		{
			flagName:   "skyfire-kube-dump-interval-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                           &opssightv1.OpsSightSpec{},
				SkyfireKubeDumpIntervalSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{KubeDumpIntervalSeconds: 10}},
		},
		// case
		{
			flagName:   "skyfire-perceptor-dump-interval-seconds",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                                &opssightv1.OpsSightSpec{},
				SkyfirePerceptorDumpIntervalSeconds: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Skyfire: &opssightv1.Skyfire{PerceptorDumpIntervalSeconds: 10}},
		},
		// case
		{
			flagName:   "blackduck-external-hosts",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                       &opssightv1.OpsSightSpec{},
				BlackduckExternalHostsJSON: []string{"{\"Scheme\": \"changed\"}"},
			},
			changedSpec: &opssightv1.OpsSightSpec{Blackduck: &opssightv1.Blackduck{ExternalHosts: []*opssightv1.Host{{Scheme: "changed"}}}},
		},
		// case
		{
			flagName:   "blackduck-connections-environment-variable-name",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec: &opssightv1.OpsSightSpec{},
				BlackduckConnectionsEnvironmentVaraiableName: "changed",
			},
			changedSpec: &opssightv1.OpsSightSpec{Blackduck: &opssightv1.Blackduck{ConnectionsEnvironmentVariableName: "changed"}},
		},
		// case
		{
			flagName:   "blackduck-TLS-verification",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                     &opssightv1.OpsSightSpec{},
				BlackduckTLSVerification: true,
			},
			changedSpec: &opssightv1.OpsSightSpec{Blackduck: &opssightv1.Blackduck{TLSVerification: true}},
		},
		// case
		{
			flagName:   "blackduck-initial-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:                  &opssightv1.OpsSightSpec{},
				BlackduckInitialCount: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Blackduck: &opssightv1.Blackduck{InitialCount: 10}},
		},
		// case
		{
			flagName:   "blackduck-max-count",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec:              &opssightv1.OpsSightSpec{},
				BlackduckMaxCount: 10,
			},
			changedSpec: &opssightv1.OpsSightSpec{Blackduck: &opssightv1.Blackduck{MaxCount: 10}},
		},
		// case
		{
			flagName:   "",
			initialCtl: NewOpsSightCtl(),
			changedCtl: &Ctl{
				Spec: &opssightv1.OpsSightSpec{},
			},
			changedSpec: &opssightv1.OpsSightSpec{},
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
