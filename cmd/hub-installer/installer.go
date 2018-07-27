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

	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/webservice"
	log "github.com/sirupsen/logrus"
)

func main() {
	configPath := os.Args[1]
	runHubCreater(configPath)
}

func runHubCreater(configPath string) {
	log.Infof("Config path: %s", configPath)
	config, err := hub.GetConfig(configPath)
	if err != nil {
		log.Panicf("Unable to read the config file from %s", configPath)
	}
	log.Infof("Config : %v", config)

	level, err := config.GetLogLevel()
	if err != nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		webservice.SetupHTTPServer()
	}()
	for {
		fmt.Println("....hub installer heartbeat: still alive...")
		time.Sleep(120 * time.Second)
	}
}
