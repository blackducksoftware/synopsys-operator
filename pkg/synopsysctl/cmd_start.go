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

// startCmd starts a resource in the cluster
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a resource")
	},
}

// startAlertCmd starts an Alert instance
var startAlertCmd = &cobra.Command{
	Use:     "alert NAMESPACE",
	Example: "synopsysctl start alert altnamespace",
	Short:   "Start an Alert instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertNamespace := args[0]
		log.Infof("starting Alert %s...", alertNamespace)

		// Get the Alert
		currAlert, err := util.GetAlert(alertClient, alertNamespace, alertNamespace)

		if err != nil {
			log.Errorf("error starting Alert '%s' due to %+v", alertNamespace, err)
			return nil
		}

		// Make changes to Spec
		currAlert.Spec.DesiredState = ""
		// Update Alert
		_, err = util.UpdateAlert(alertClient,
			currAlert.Spec.Namespace, currAlert)
		if err != nil {
			log.Errorf("error starting Alert '%s' due to %+v", alertNamespace, err)
			return nil
		}

		log.Infof("successfully started Alert '%s'", alertNamespace)
		return nil
	},
}

// startBlackDuckCmd starts a Black Duck instance
var startBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAMESPACE",
	Example: "synopsysctl start blackduck bdnamespace",
	Short:   "Start a Black Duck instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		log.Infof("starting Black Duck '%s'...", blackDuckNamespace)

		// Get the Black Duck
		currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error starting Black Duck '%s' due to %+v", blackDuckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackDuck.Spec.DesiredState = ""
		// Update Blackduck
		_, err = util.UpdateBlackduck(blackDuckClient,
			currBlackDuck.Spec.Namespace, currBlackDuck)
		if err != nil {
			log.Errorf("error starting Black Duck '%s' due to %+v", blackDuckNamespace, err)
			return nil
		}

		log.Infof("successfully started Black Duck '%s'", blackDuckNamespace)
		return nil
	},
}

// startOpsSightCmd starts an OpsSight instance
var startOpsSightCmd = &cobra.Command{
	Use:     "opssight NAMESPACE",
	Example: "synopsysctl start opssight opsnamespace",
	Short:   "Start an OpsSight instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		log.Infof("starting OpsSight '%s'...", opsSightNamespace)

		// Get the OpsSight
		currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)

		if err != nil {
			log.Errorf("error starting OpsSight '%s' due to %+v", opsSightNamespace, err)
			return nil
		}

		// Make changes to Spec
		currOpsSight.Spec.DesiredState = ""
		// Update OpsSight
		_, err = util.UpdateOpsSight(opsSightClient, currOpsSight.Spec.Namespace, currOpsSight)
		if err != nil {
			log.Errorf("error starting OpsSight '%s' due to %+v", opsSightNamespace, err)
			return nil
		}

		log.Infof("successfully started OpsSight '%s'", opsSightNamespace)
		return nil
	},
}

func init() {
	startCmd.AddCommand(startAlertCmd)
	startCmd.AddCommand(startBlackDuckCmd)
	startCmd.AddCommand(startOpsSightCmd)
	rootCmd.AddCommand(startCmd)
}
