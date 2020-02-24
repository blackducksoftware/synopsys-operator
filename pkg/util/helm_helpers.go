/*
Copyright (C) 2020 Synopsys, Inc.

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

package util

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var settings = cli.New()

// CreateWithHelm3 uses the helm NewInstall action to create a resource in the cluster
// Modified from https://github.com/openshift/console/blob/cdf6b189b71e488033ecaba7d90258d9f9453478/pkg/helm/actions/install_chart.go
// Helm Actions: https://github.com/helm/helm/tree/9bc7934f350233fa72a11d2d29065aa78ab62792/pkg/action
func CreateWithHelm3(releaseName, namespace, chartURL string, vals map[string]interface{}, kubeConfig string, dryRun bool) error {
	actionConfig, err := CreateHelmActionConfiguration(kubeConfig, "", namespace)
	if err != nil {
		return err
	}

	chart, err := LoadChart(chartURL, actionConfig)
	if err != nil {
		return err
	}
	validInstallableChart, err := isChartInstallable(chart)
	if !validInstallableChart {
		return fmt.Errorf("release at '%s' is not installable: %+v", chartURL, err)
	}
	if chart.Metadata.Deprecated {
		log.Warnf("the release at '%s' is deprecated", chartURL)
	}
	// TODO: determine if we will check dependencies...
	// if req := chart.Metadata.Dependencies; req != nil {
	// 	// If CheckDependencies returns an error, we have unfulfilled dependencies.
	// 	// As of Helm 2.4.0, this is treated as a stopping condition:
	// 	// https://github.com/helm/helm/issues/2209
	// 	if err := action.CheckDependencies(chart, req); err != nil {
	// 		if client.DependencyUpdate {
	// 			man := &downloader.Manager{
	// 				Out:              out,
	// 				ChartPath:        cp,
	// 				Keyring:          client.ChartPathOptions.Keyring,
	// 				SkipUpdate:       false,
	// 				Getters:          p,
	// 				RepositoryConfig: settings.RepositoryConfig,
	// 				RepositoryCache:  settings.RepositoryCache,
	// 			}
	// 			if err := man.Update(); err != nil {
	// 				return nil, err
	// 			}
	// 		} else {
	// 			return nil, err
	// 		}
	// 	}
	// }

	client := action.NewInstall(actionConfig)
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	client.ReleaseName = releaseName
	client.DryRun = dryRun
	_, err = client.Run(chart, vals) // deploy the chart into the namespace from the actionConfig
	if err != nil {
		return fmt.Errorf("failed to run install: %+v", err)
	}
	return nil
}

// UpdateWithHelm3 uses the helm NewUpgrade action to update a resource in the cluster
func UpdateWithHelm3(releaseName, namespace, chartURL string, vals map[string]interface{}, kubeConfig string) error {
	actionConfig, err := CreateHelmActionConfiguration(kubeConfig, "", namespace)
	if err != nil {
		return err
	}
	if releaseExists := ReleaseExists(releaseName, namespace, kubeConfig); !releaseExists {
		return fmt.Errorf("release '%s' does not exist", releaseName)
	}

	chart, err := LoadChart(chartURL, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to load release at '%s' for updating: %s", chartURL, err)
	}
	validInstallableChart, err := isChartInstallable(chart)
	if !validInstallableChart {
		return fmt.Errorf("release at '%s' is not installable: %+v", chartURL, err)
	}

	client := action.NewUpgrade(actionConfig)
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	client.ReuseValues = true                     // rememeber the values that have been set previously
	_, err = client.Run(releaseName, chart, vals) // updates the release in the namespace from the actionConfig
	if err != nil {
		return fmt.Errorf("failed to run upgrade: %+v", err)
	}
	return nil
}

// TemplateWithHelm3 prints the kube manifest files for a resource
func TemplateWithHelm3(releaseName, namespace, chartURL string, vals map[string]interface{}) error {
	actionConfig, err := CreateHelmActionConfiguration("", "", namespace)
	if err != nil {
		return err
	}
	chart, err := LoadChart(chartURL, actionConfig)
	validInstallableChart, err := isChartInstallable(chart)
	if !validInstallableChart {
		return err
	}
	templateOutput, err := RenderManifests(releaseName, namespace, chart, vals, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to render kube manifest files: %s", err)
	}
	fmt.Printf("%+v\n", templateOutput)
	return nil
}

// RenderManifests converts a helm chart to a string of the kube manifest files
// Modified from https://github.com/openshift/console/blob/cdf6b189b71e488033ecaba7d90258d9f9453478/pkg/helm/actions/template_test.go
func RenderManifests(releaseName, namespace string, chart *chart.Chart, vals map[string]interface{}, actionConfig *action.Configuration) (string, error) {
	var showFiles []string
	response := make(map[string]string)
	validate := false
	includeCrds := true
	emptyResponse := ""

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = releaseName
	client.Namespace = namespace
	client.Replace = true // Skip the releaseName check
	client.ClientOnly = !validate

	rel, err := client.Run(chart, vals)
	if err != nil {
		return emptyResponse, err
	}

	var manifests bytes.Buffer
	var output bytes.Buffer

	if includeCrds {
		for _, f := range rel.Chart.CRDs() {
			fmt.Fprintf(&manifests, "---\n# Source: %s\n%s\n", f.Name, f.Data)
		}
	}

	fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))

	if !client.DisableHooks {
		for _, m := range rel.Hooks {
			fmt.Fprintf(&manifests, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
		}
	}

	// if we have a list of files to render, then check that each of the
	// provided files exists in the chart.
	if len(showFiles) > 0 {
		splitManifests := releaseutil.SplitManifests(manifests.String())
		manifestNameRegex := regexp.MustCompile("# Source: [^/]+/(.+)")
		var manifestsToRender []string
		for _, f := range showFiles {
			missing := true
			for _, manifest := range splitManifests {
				submatch := manifestNameRegex.FindStringSubmatch(manifest)
				if len(submatch) == 0 {
					continue
				}
				manifestName := submatch[1]
				// manifest.Name is rendered using linux-style filepath separators on Windows as
				// well as macOS/linux.
				manifestPathSplit := strings.Split(manifestName, "/")
				manifestPath := filepath.Join(manifestPathSplit...)

				// if the filepath provided matches a manifest path in the
				// chart, render that manifest
				if f == manifestPath {
					manifestsToRender = append(manifestsToRender, manifest)
					missing = false
				}
			}
			if missing {
				return "", fmt.Errorf("could not find template %s in chart", f)
			}
			for _, m := range manifestsToRender {
				response[f] = m
				fmt.Fprintf(&output, "---\n%s\n", m)
			}
		}
	} else {
		fmt.Fprintf(&output, "%s", manifests.String())
	}
	return output.String(), nil
}

// DeleteWithHelm3 uses the helm NewUninstall action to delete a resource from the cluster
func DeleteWithHelm3(releaseName, namespace, kubeConfig string) error {
	actionConfig, err := CreateHelmActionConfiguration(kubeConfig, "", namespace)
	if err != nil {
		return err
	}
	if releaseExists := ReleaseExists(releaseName, namespace, kubeConfig); !releaseExists {
		return fmt.Errorf("release '%s' does not exist", releaseName)
	}
	client := action.NewUninstall(actionConfig)
	_, err = client.Run(releaseName) // deletes the releaseName from the namespace in the actionConfig
	if err != nil {
		return fmt.Errorf("failed to run uninstall: %+v", err)
	}
	return nil
}

// GetWithHelm3 uses the helm NewGet action to return a Release with information about
// a resource from the cluster
func GetWithHelm3(releaseName, namespace, kubeConfig string) (*release.Release, error) {
	actionConfig, err := CreateHelmActionConfiguration(kubeConfig, "", namespace)
	if err != nil {
		return nil, err
	}
	client := action.NewGet(actionConfig)
	release, err := client.Run(releaseName) // lists the releases in the namespace from the actionConfig
	if err != nil {
		return nil, fmt.Errorf("failed to run get: %+v", err)
	}
	return release, nil
}

// CreateHelmActionConfiguration creates an action.Configuration that points to the specified cluster and namespace
func CreateHelmActionConfiguration(kubeConfig, kubeContext, namespace string) (*action.Configuration, error) {
	// TODO: look into using GetActionConfigurations()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(kube.GetConfig(kubeConfig, kubeContext, namespace), namespace, "secret", func(format string, v ...interface{}) {}); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

type configFlagsWithTransport struct {
	*genericclioptions.ConfigFlags
	Transport *http.RoundTripper
}

// GetActionConfigurations creates an action.Configuration that points to the specified cluster and namespace
// TODO - this function specifies more values than CreateHelmActionConfiguration(), consider using this
func GetActionConfigurations(host, namespace, token string, transport *http.RoundTripper) *action.Configuration {

	confFlags := &configFlagsWithTransport{
		ConfigFlags: &genericclioptions.ConfigFlags{
			APIServer:   &host,
			BearerToken: &token,
			Namespace:   &namespace,
		},
		Transport: transport,
	}
	inClusterCfg, err := rest.InClusterConfig()

	if err != nil {
		fmt.Print("Running outside cluster, CAFile is unset")
	} else {
		confFlags.CAFile = &inClusterCfg.CAFile
	}
	conf := new(action.Configuration)
	conf.Init(confFlags, namespace, "secrets", klog.Infof)

	return conf
}

// LoadChart returns a chart from the specified chartURL
// Modified from https://github.com/openshift/console/blob/master/pkg/helm/actions/template_test.go
func LoadChart(chartURL string, actionConfig *action.Configuration) (*chart.Chart, error) {
	client := action.NewInstall(actionConfig)

	// Get full path - checks local machine and chart repository
	chartFullPath, err := client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart with '%s': %s", chartURL, err)
	}

	chart, err := loader.Load(chartFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart from %s", chartFullPath)
	}
	return chart, nil
}

// isChartInstallable validates if a chart can be installed
//
// Only the "application" chart type is installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// ReleaseExists verifies that a resources is deployed in the cluster
func ReleaseExists(releaseName, namespace, kubeConfig string) bool {
	actionConfig, err := CreateHelmActionConfiguration(kubeConfig, "", namespace)
	if err != nil {
		return false
	}
	client := action.NewGet(actionConfig)
	release, err := client.Run(releaseName) // lists the releases in the namespace from the actionConfig
	if err != nil || release == nil {
		return false
	}
	return true
}
