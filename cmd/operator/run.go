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

	crdEnv, ok := os.LookupEnv("CRD_NAMES")
	if ok && len(crdEnv) > 0 {
		crds := strings.Split(crdEnv, ",")
		for _, crd := range crds {
			startController(configPath, crd, stopCh)
		}
	} else {
		log.Errorf("unable to start any CRD controllers. Please set the CRD_NAMES environment variable to start any CRD controllers...")
	}

	go func() {
		webhook.NewOperatorWebhook(deployer.KubeConfig).Start()
	}()

	// Start the prometheus endpoint
	protoform.SetupHTTPServer()

	if deployer.Config.OperatorTimeBombInSeconds > 0 {
		go func() {
			timeout := time.Duration(deployer.Config.OperatorTimeBombInSeconds) * time.Second
			log.Warnf("self timeout is enabled to %v seconds", timeout)
			time.Sleep(timeout)

			// trip the stop channel after done sleeping.  wait 20 seconds for debuggability.
			log.Warn("timeout tripped.  exiting in 20 seconds !")
			time.Sleep(time.Duration(20) * time.Second)
			kill(stopCh)
		}()
	}
	<-stopCh
}

// addController will start the CRD controller
func startController(configPath string, name string, stopCh chan struct{}) {
	crd := strings.SplitN(name, ":", 2)
	if len(crd) != 2 {
		panic(fmt.Errorf("CRD_NAMES environment variable are not set properly"))
	}
	name = crd[0]
	// Add controllers to the Operator
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	switch strings.ToLower(name) {
	case util.BlackDuckCRDName:
		hubController := blackduck.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, getClusterScope(crd[1]), util.GetBlackDuckTemplate(), stopCh)
		deployer.AddController(hubController)
	case util.AlertCRDName:
		alertController := alert.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, getClusterScope(crd[1]), util.GetAlertTemplate(), stopCh)
		deployer.AddController(alertController)
	case util.OpsSightCRDName:
		opssSightController, err := opssight.NewCRDInstaller(&opssight.Config{
			Config:                  deployer.Config,
			KubeConfig:              deployer.KubeConfig,
			KubeClientSet:           deployer.KubeClientSet,
			Defaults:                util.GetOpsSightDefault(),
			Threadiness:             deployer.Config.Threadiness,
			StopCh:                  stopCh,
			IsBlackDuckClusterScope: getClusterScopeByName(util.BlackDuckCRDName),
		})
		if err != nil {
			panic(err)
		}
		deployer.AddController(opssSightController)
	case util.PrmCRDName:
		log.Info("Polaris Reporting Module will be coming soon!!!")
	default:
		log.Warnf("unable to start the %s custom resource definition controller due to invalid custom resource definition name", name)
	}
	if err = deployer.Deploy(); err != nil {
		log.Errorf("ran into errors during deployment, but continuing anyway: %s", err.Error())
	}
	log.Infof("started %s crd controller", name)
}

// getClusterScope returns whether the CRD scope is cluster scope
func getClusterScope(crdScope string) bool {
	switch strings.ToLower(crdScope) {
	case "cluster":
		return true
	}
	return false
}

func getClusterScopeByName(name string) bool {
	crdEnv, ok := os.LookupEnv("CRD_NAMES")
	if ok {
		crdList := strings.Split(crdEnv, ",")
		for _, crds := range crdList {
			crd := strings.SplitN(crds, ":", 2)
			if len(crd) != 2 {
				panic(fmt.Errorf("CRD_NAMES environment variable are not set properly"))
			}
			if name == crd[0] {
				return getClusterScope(crd[1])
			}
		}
	}
	return false
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
