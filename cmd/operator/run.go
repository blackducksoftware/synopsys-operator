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

package main

import (
	"fmt"
	"os"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/webhook"
	log "github.com/sirupsen/logrus"

	//"github.com/blackducksoftware/synopsys-operator/pkg/sample"

	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components"
	_ "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/components"
)

var version string

func main() {
	log.Infof("version: %s", version)
	if len(os.Args) > 1 {
		configPath := os.Args[1]
		runProtoform(configPath, version)
		fmt.Printf("Config path: %s", configPath)
		return
	}
	log.Warn("no config file sent. running operator with environment variable and default settings")
	runProtoform("", version)
}

// runProtoform will add CRD controllers to the Protoform Deployer which
// will call each of their Deploy functions
func runProtoform(configPath string, version string) {
	// Add controllers to the Protoform Deployer
	deployer, err := protoform.NewController(configPath, version)
	if err != nil {
		panic(err)
	}

	// check for the existence of operator configmap, if not create it
	_, err = util.GetConfigMap(deployer.KubeClient, deployer.Config.Namespace, "synopsys-operator")
	if err != nil {
		d, err := horizon.NewDeployer(deployer.KubeConfig)
		if err != nil {
			panic("unable to create deployer object")
		}

		operatorConfig := soperator.SpecConfig{
			Namespace:                     deployer.Config.Namespace,
			Image:                         "",
			Expose:                        "",
			DryRun:                        deployer.Config.DryRun,
			LogLevel:                      deployer.Config.LogLevel,
			Threadiness:                   deployer.Config.Threadiness,
			PostgresRestartInMins:         deployer.Config.PostgresRestartInMins,
			PodWaitTimeoutSeconds:         deployer.Config.PodWaitTimeoutSeconds,
			ResyncIntervalInSeconds:       deployer.Config.ResyncIntervalInSeconds,
			TerminationGracePeriodSeconds: deployer.Config.TerminationGracePeriodSeconds,
			Crds:                          strings.Split(deployer.Config.CrdNames, ","),
			IsClusterScoped:               deployer.Config.IsClusterScoped,
		}
		operatorCm, err := operatorConfig.GetOperatorConfigMap()
		if err != nil {
			panic("unable to create operator configmap")
		}
		d.AddComponent(horizonapi.ConfigMapComponent, operatorCm)
		d.Run()
	}

	// check for the existence of prometheus configmap, if not create it
	_, err = util.GetConfigMap(deployer.KubeClient, deployer.Config.Namespace, "prometheus")
	if err != nil {
		d, err := horizon.NewDeployer(deployer.KubeConfig)
		if err != nil {
			panic("unable to create deployer object")
		}
		prometheusConfig := soperator.PrometheusSpecConfig{
			Namespace: deployer.Config.Namespace,
			Image:     "",
			Expose:    "",
		}
		prometheusCm, err := prometheusConfig.GetPrometheusConfigMap()
		if err != nil {
			panic("unable to create prometheus configmap")
		}
		d.AddComponent(horizonapi.ConfigMapComponent, prometheusCm)
		d.Run()
	}

	// Log Kubernetes version
	kversion, err := util.GetKubernetesVersion(deployer.KubeClient)
	if err == nil {
		log.Infof("Kubernetes: %s", kversion)
	}

	stopCh := make(chan struct{})

	// get the list of crds from an environment variable
	if len(deployer.Config.CrdNames) > 0 {
		crds := strings.Split(deployer.Config.CrdNames, ",")
		for _, crd := range crds {
			// start the CRD controller
			startController(deployer, strings.TrimSpace(crd), stopCh)
		}
		if err := deployer.Deploy(); err != nil {
			log.Errorf("unable to deploy the CRD controllers due to  %+v", err)
			os.Exit(1)
		}
	} else {
		log.Errorf("unable to start any CRD controllers. Please set the CrdNames environment variable to start any CRD controllers...")
		os.Exit(1)
	}

	if deployer.Config.AdmissionWebhookListener {
		go func() {
			webhook.NewOperatorWebhook(deployer.KubeConfig).Start()
		}()
	}

	// Start the prometheus endpoint
	protoform.SetupHTTPServer()
	<-stopCh
}

// startController will start the CRD controller
func startController(deployer *protoform.Deployer, name string, stopCh chan struct{}) {
	// Add controllers to the Operator
	switch strings.ToLower(name) {
	case util.BlackDuckCRDName:
		blackduckController := blackduck.NewCRDInstaller(deployer, util.GetBlackDuckTemplate(), stopCh)
		deployer.AddController(blackduckController)
	case util.AlertCRDName:
		alertController := alert.NewCRDInstaller(deployer, util.GetAlertTemplate(), stopCh)
		deployer.AddController(alertController)
	case util.OpsSightCRDName:
		opsSightController := opssight.NewCRDInstaller(deployer, util.GetOpsSightDefault(), stopCh)
		deployer.AddController(opsSightController)
	default:
		log.Warnf("unable to start the %s custom resource definition controller due to invalid custom resource definition name", name)
	}
	log.Infof("added %s crd controller", name)
}

func kill(stopCh chan struct{}) {
	// TODO: no idea why this doesnt actually cause the program to exit.  must be another
	// channel open somewhere....
	go func() {
		stopCh <- struct{}{}
	}()
	// hard exit b/c of the above comment ^
	os.Exit(0)
}
