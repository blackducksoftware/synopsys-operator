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

package synopsysctl

import (
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd starts a resource in the cluster if it's stopped
var startCmd = &cobra.Command{
	Use:   "start [resource]",
	Short: "Start a Synopsys Resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Must specify a resource")
	},
}

// startBlackduckCmd starts a Blackduck in the cluster
var startBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Start a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackduckNamespace := args[0]
		log.Infof("Starting BlackDuck %s...", blackduckNamespace)

		// Get the Black Duck
		currBlackduck, err := util.GetHub(blackduckClient, blackduckNamespace, blackduckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackduckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackduck.Spec.DesiredState = ""
		// Update Blackduck
		_, err = util.UpdateBlackduck(blackduckClient,
			currBlackduck.Spec.Namespace, currBlackduck)
		if err != nil {
			log.Errorf("error starting the %s Black Duck instance due to %+v", blackduckNamespace, err)
			return nil
		}

		log.Infof("successfully started the '%s' Black Duck instance", blackduckNamespace)
		return nil
	},
}

func init() {
	startCmd.AddCommand(startBlackduckCmd)
	rootCmd.AddCommand(startCmd)
}
