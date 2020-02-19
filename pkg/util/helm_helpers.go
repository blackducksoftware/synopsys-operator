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
	"fmt"
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
)

var settings = cli.New()

// InstallChart https://github.com/openshift/console/blob/master/pkg/helm/actions/install_chart.go
func InstallChart(namespace, name, url string, vals map[string]interface{}, conf *action.Configuration) (*release.Release, error) {
	client := action.NewInstall(conf)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	name, chart, err := client.NameAndChart([]string{name, url})
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		log.Info("WARNING: This chart is deprecated")
	}

	// if req := chartRequested.Metadata.Dependencies; req != nil {
	// 	// If CheckDependencies returns an error, we have unfulfilled dependencies.
	// 	// As of Helm 2.4.0, this is treated as a stopping condition:
	// 	// https://github.com/helm/helm/issues/2209
	// 	if err := action.CheckDependencies(chartRequested, req); err != nil {
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

	client.Namespace = namespace
	return client.Run(chartRequested, vals)
}

// DeployWithHelm ...
func DeployWithHelm(kubeConfig, kubeContext, namespace, name, url string, vals map[string]interface{}) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(kube.GetConfig(kubeConfig, kubeContext, namespace), namespace, "memory", func(format string, v ...interface{}) {}); err != nil {
		log.Fatal(err)
	}
	_, err := InstallChart(namespace, name, url, vals, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to install chart: %+v", err)
	}
	// fmt.Printf("Release: %+v\n", rel)
	// fmt.Printf("Release Namespace: %+v\n", rel.Namespace)
	// fmt.Printf("Release Name: %+v\n", rel.Name)
	// fmt.Printf("Release Version: %+v\n", rel.Version)
	// fmt.Printf("Release Config: %+v\n", rel.Config)
	// fmt.Printf("Release Manifest: %+v\n", rel.Manifest)
	// fmt.Printf("Release Chart Path: %+v\n", rel.Chart.ChartPath())
	// fmt.Printf("Release Chart FullPath: %+v\n", rel.Chart.ChartFullPath())
	// fmt.Printf("Release Chart Values: %+v\n", rel.Chart.Values)
	// fmt.Printf("Release Chart Metadata: %+v\n", rel.Chart.Metadata)
	return nil
}

// RunHelm3 executes a helm command
// It takes in a helm command, arguments to the command, and values to set in the helm chart
func RunHelm3(commandName string, kubeConfig, kubeContext, namespace, name, url string, vals map[string]interface{}) (string, error) {

	err := DeployWithHelm(kubeConfig, kubeContext, namespace, name, url, vals)
	if err != nil {
		return "", err
	}
	return "", nil

	// var helmExists bool
	// var err error
	// if helmExists, err = HelmV3Exists(); err != nil {
	// 	return "", err
	// }
	// if !helmExists {
	// 	return "", fmt.Errorf("helm v3 is not installed in PATH")
	// }
	// cmdArgs := genHelm3Args(commandName, args, setValuesMap)
	// cmd := exec.Command("helm", cmdArgs...)
	// log.Debugf("%+v", cmd)
	// stdoutErr, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return string(stdoutErr), fmt.Errorf("failed to run Helm command of args %+v with error %s", cmdArgs, err)
	// }
	// return string(stdoutErr), nil
}

// isChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func genHelm3Args(command string, args []string, setValuesMap map[string]string) []string {
	helmArgs := append([]string{command}, args...)
	for name, value := range setValuesMap {
		helmArgs = append(helmArgs, "--set", fmt.Sprintf("%s=%s", name, value))
	}
	return helmArgs
}

// HelmV3Exists returns true if it can find the helm binary and it is v3
func HelmV3Exists() (bool, error) {
	helmExists, err := HelmIsInPath()
	if err != nil {
		return false, fmt.Errorf("failed to look for Helm in PATH: %s", err)
	}
	if !helmExists {
		return false, nil
	}
	isV3, err := HelmIsV3()
	if err != nil {
		return false, fmt.Errorf("failed to determine if Helm is V3: %+v", err)
	}
	if !isV3 {
		return false, nil
	}
	return true, nil
}

// HelmIsInPath returns true if it finds the helm binary in the
// user's PATH
func HelmIsInPath() (bool, error) {
	_, err := exec.LookPath("helm")
	if err != nil {
		return false, err
	}
	return true, nil
}

// HelmIsV3 returns true if the helm binary on the user's system is v3
func HelmIsV3() (bool, error) {
	cmd := exec.Command("helm", "version", "--short")
	stdoutErr, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s - %s", string(stdoutErr), err)
	}
	version, err := ValidateHelmVersion(string(stdoutErr))
	if err != nil {
		return false, fmt.Errorf("failed to validate Helm version: %s", err)
	}
	if version != "3" {
		return false, fmt.Errorf("Helm version is not v3")
	}
	return true, nil
}

// ValidateHelmVersion takes in output from "helm version --short" and verifies that it's
// formatted correctly. It returns the first value from the version
func ValidateHelmVersion(helmVersionOutput string) (string, error) {
	var rgx = regexp.MustCompile(`v([0-9])\.[0-9]\.[0-9]\+[0-9a-z]+`)

	versionMatches := rgx.FindStringSubmatch(helmVersionOutput)
	if len(versionMatches) != 2 {
		return "", fmt.Errorf("invalid 'helm version --short' output: %s", helmVersionOutput)
	}
	return versionMatches[1], nil
}
