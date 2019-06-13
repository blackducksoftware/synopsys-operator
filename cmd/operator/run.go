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
	"strconv"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/webhook"
	log "github.com/sirupsen/logrus"
	//"github.com/blackducksoftware/synopsys-operator/pkg/sample"
)

func main() {
	if len(os.Args) > 1 {
		configPath := os.Args[1]
		runProtoform(configPath)
		fmt.Printf("Config path: %s", configPath)
		return
	}
	fmt.Println("WARNING: Running protoform with defaults, no config file sent.")
	runProtoform("")
}

// runProtoform will add CRD controllers to the Protoform Deployer which
// will call each of their Deploy functions
func runProtoform(configPath string) {
	// Add controllers to the Protoform Deployer
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	// Log Kubernetes version
	kversion, err := util.GetKubernetesVersion(deployer.KubeClientSet)
	if err == nil {
		log.Infof("Kubernetes: %s", kversion)
	}

	// Log Openshift version
	oversion, err := util.GetOcVersion(deployer.KubeClientSet)
	if err == nil {
		log.Infof("Openshift: %s", oversion)
	}

	stopCh := make(chan struct{})

	// get the operator scope from an environment variable
	var isClusterScoped bool
	clusterScopeEnv, ok := os.LookupEnv("CLUSTER_SCOPE")
	if ok && len(clusterScopeEnv) > 0 {
		clusterScope, err := strconv.ParseBool(strings.TrimSpace(clusterScopeEnv))
		if err != nil {
			panic("cluster scope environment variable is not set properly. possible values are true/false")
		}
		isClusterScoped = clusterScope
	}

	// get the list of crds from an environment variable
	crdEnv, ok := os.LookupEnv("CRD_NAMES")
	if ok && len(crdEnv) > 0 {
		crds := strings.Split(crdEnv, ",")
		for _, crd := range crds {
			// start the CRD controller
			startController(configPath, strings.TrimSpace(crd), isClusterScoped, stopCh)
		}
	} else {
		log.Errorf("unable to start any CRD controllers. Please set the CRD_NAMES environment variable to start any CRD controllers...")
		os.Exit(1)
	}

	if deployer.Config.AdmissionWebhookListener {
		go func() {
			webhook.NewOperatorWebhook(deployer.KubeConfig).Start()
		}()
	}

	// Start the prometheus endpoint
	protoform.SetupHTTPServer()

	if deployer.Config.OperatorTimeBombInSeconds > 0 {
		go func() {
			timeout := time.Duration(deployer.Config.OperatorTimeBombInSeconds) * time.Second
			log.Warnf("timeout to stop all controllers is enabled to %v seconds", timeout)
			time.Sleep(timeout)

			// trip the stop channel after done sleeping.  wait 20 seconds for debuggability.
			log.Warn("timeout tripped.  exiting in 20 seconds !")
			time.Sleep(time.Duration(20) * time.Second)
			kill(stopCh)
		}()
	}
	<-stopCh
}

// startController will start the CRD controller
func startController(configPath string, name string, isClusterScoped bool, stopCh chan struct{}) {
	// Add controllers to the Operator
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}
	deployer.Config.IsClusterScoped = isClusterScoped

	switch strings.ToLower(name) {
	case util.BlackDuckCRDName:
		blackduckController := blackduck.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, util.GetBlackDuckTemplate(), stopCh)
		deployer.AddController(blackduckController)
	case util.AlertCRDName:
		alertController := alert.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, util.GetAlertTemplate(), stopCh)
		deployer.AddController(alertController)
	case util.OpsSightCRDName:
		opsSightController := opssight.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, util.GetOpsSightDefault(), stopCh)
		deployer.AddController(opsSightController)
	default:
		log.Warnf("unable to start the %s custom resource definition controller due to invalid custom resource definition name", name)
	}
	if err = deployer.Deploy(); err != nil {
		log.Errorf("unable to deploy the CRD controllers due to  %+v", err)
		os.Exit(1)
	}
	log.Infof("started %s crd controller", name)
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
