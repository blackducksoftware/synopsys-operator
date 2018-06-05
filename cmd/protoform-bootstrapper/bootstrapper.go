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

	"github.com/blackducksoftware/perceptor-protoform/cmd/protoform-bootstrapper/app"
	"github.com/blackducksoftware/perceptor-protoform/cmd/protoform-bootstrapper/app/options"
)

func main() {
	opts := options.NewBootstrapperOptions()
	configFile := os.Args[1]
	if len(configFile) <= 0 {
		panic(fmt.Errorf("no config provided"))
	}
	err := opts.ReadConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("failed to read config: %v", err))
	}

	// Get values if interactive mode is enabled
	if opts.Interactive {
		fmt.Println("============================================")
		fmt.Println("Blackduck Hub Configuration Information")
		fmt.Println("============================================")
		fmt.Println("Interactive")
		fmt.Println("============================================")
		err := opts.InteractiveConfig()
		fmt.Println("============================================")
		if err != nil {
			panic(err)
		}
	}

	bootstrapper, err := app.NewBootstrapper(opts)
	if err != nil {
		panic(fmt.Errorf("failed to create bootstrapper: %v", err))
	}

	err = bootstrapper.Run()
	if err != nil {
		fmt.Printf("failed to run bootstrapper: %v\n", err)
	}
}
