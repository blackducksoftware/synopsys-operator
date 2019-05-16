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

// stopCmd stops a resource in the cluster
var stopCmd = &cobra.Command{
	Use:   "stop [resource]",
	Short: "Stops a Synopsys Resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Must specify a resource")
	},
}

// stopBlackduckCmd stops a Blackduck in the cluster
var stopBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Stops a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackduckNamespace := args[0]
		log.Infof("Stopping BlackDuck %s...", blackduckNamespace)

		// Get the Black Duck
		currBlackduck, err := util.GetHub(blackduckClient, blackduckNamespace, blackduckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackduckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackduck.Spec.DesiredState = "STOP"
		// Update Blackduck
		_, err = util.UpdateBlackduck(blackduckClient,
			currBlackduck.Spec.Namespace, currBlackduck)
		if err != nil {
			log.Errorf("error stopping the %s Black Duck instance due to %+v", blackduckNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' Black Duck instance", blackduckNamespace)
		return nil
	},
}

func init() {
	stopCmd.AddCommand(stopBlackduckCmd)
	rootCmd.AddCommand(stopCmd)
}
