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
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/webhook"

	"github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	bdutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
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

func kill(stopCh chan struct{}) {
	// TODO: no idea why this doesnt actually cause the program to exit.  must be another
	// channel open somewhere....
	go func() {
		stopCh <- struct{}{}
	}()
	// hard exit b/c of the above comment ^
	os.Exit(0)
}

// runProtoform will add CRD controllers to the Protoform Deployer which
// will call each of their Deploy functions
func runProtoform(configPath string) {
	// Add controllers to the Protoform Deployer
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})

	//sampleController := sample.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, bdutil.GetSampleDefaultValue(), stopCh)
	//deployer.AddController(sampleController)

	alertController := alert.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, bdutil.GetAlertTemplate(), stopCh)
	deployer.AddController(alertController)

	hubController := blackduck.NewCRDInstaller(deployer.Config, deployer.KubeConfig, deployer.KubeClientSet, bdutil.GetBlackDuckTemplate(), stopCh)
	deployer.AddController(hubController)

	opssSightController, err := opssight.NewCRDInstaller(&opssight.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      bdutil.GetOpsSightTemplate(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	if err != nil {
		panic(err)
	}
	deployer.AddController(opssSightController)

	logrus.Info("Starting deployer.  All controllers have been added to Protoform.")
	if err = deployer.Deploy(); err != nil {
		logrus.Errorf("ran into errors during deployment, but continuing anyway: %s", err.Error())
	}

	go func() {
		webhook.NewOperatorWebhook(deployer.KubeConfig).Start()
	}()

	// Start the prometheus endpoint
	protoform.SetupHTTPServer()

	if deployer.Config.OperatorTimeBombInSeconds > 0 {
		go func() {
			timeout := time.Duration(deployer.Config.OperatorTimeBombInSeconds) * time.Second
			logrus.Warnf("Self timeout is enabled to %v seconds", timeout)
			time.Sleep(timeout)

			// trip the stop channel after done sleeping.  wait 20 seconds for debuggability.
			logrus.Warn("Timeout tripped.  Exiting In 20 seconds !")
			time.Sleep(time.Duration(20) * time.Second)
			kill(stopCh)
		}()
	}
	<-stopCh
}
