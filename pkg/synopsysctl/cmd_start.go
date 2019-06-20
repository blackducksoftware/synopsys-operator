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

// startCmd starts a Synopsys resource in the cluster
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// startAlertCmd starts an Alert instance
var startAlertCmd = &cobra.Command{
	Use:     "alert NAME",
	Example: "synopsysctl start alert <name>\nsynopsysctl start alert <name> -n <namespace>",
	Short:   "Start an Alert instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName, alertNamespace, _, err := getInstanceInfo(cmd, args[0], util.AlertCRDName, util.AlertName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("starting Alert '%s' in namespace '%s'...", alertName, alertNamespace)

		// Get the Alert
		currAlert, err := util.GetAlert(alertClient, alertNamespace, alertName)
		if err != nil {
			log.Errorf("error getting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
			return nil
		}

		// Make changes to Spec
		currAlert.Spec.DesiredState = ""
		// Update Alert
		_, err = util.UpdateAlert(alertClient, currAlert.Spec.Namespace, currAlert)
		if err != nil {
			log.Errorf("error updating Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
			return nil
		}

		log.Infof("successfully submitted start Alert '%s' in namespace '%s'", alertName, alertNamespace)
		return nil
	},
}

// startBlackDuckCmd starts a Black Duck instance
var startBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME",
	Example: "synopsysctl start blackduck <name>\nsynopsysctl start blackduck <name> -n <namespace>",
	Short:   "Start a Black Duck instance",
	Aliases: []string{"bds", "bd"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args[0], util.BlackDuckCRDName, util.BlackDuckName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("starting Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

		// Get the Black Duck
		currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackDuck.Spec.DesiredState = ""
		// Update Blackduck
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck.Spec.Namespace, currBlackDuck)
		if err != nil {
			log.Errorf("error updating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		log.Infof("successfully submitted start Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// startOpsSightCmd starts an OpsSight instance
var startOpsSightCmd = &cobra.Command{
	Use:     "opssight NAME",
	Example: "synopsysctl start opssight <name>\nsynopsysctl start opssight <name> -n <namespace>",
	Short:   "Start an OpsSight instance",
	Aliases: []string{"ops"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName, opsSightNamespace, _, err := getInstanceInfo(cmd, args[0], util.OpsSightCRDName, util.OpsSightName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("starting OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

		// Get the OpsSight
		currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightName)

		if err != nil {
			log.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		// Make changes to Spec
		currOpsSight.Spec.DesiredState = ""
		// Update OpsSight
		_, err = util.UpdateOpsSight(opsSightClient, currOpsSight.Spec.Namespace, currOpsSight)
		if err != nil {
			log.Errorf("error updating OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		log.Infof("successfully submitted start OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	startCmd.AddCommand(startAlertCmd)

	startBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	startCmd.AddCommand(startBlackDuckCmd)

	startOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	startCmd.AddCommand(startOpsSightCmd)
}
