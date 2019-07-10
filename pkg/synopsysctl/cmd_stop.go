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

// stopCmd stops a Synopsys resource in the cluster
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// stopAlertCmd stops an Alert instance
var stopAlertCmd = &cobra.Command{
	Use:           "alert NAME",
	Example:       "synopsysctl stop alert <name>\nsynopsysctl stop alert <name1> <name2>\nsynopsysctl stop alert <name> -n <namespace>\nsynopsysctl stop alert <name1> <name2> -n <namespace>",
	Short:         "Stop an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes one or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		errors := []error{}
		for _, altArg := range args {
			alertName, alertNamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, altArg)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			log.Infof("stopping Alert '%s' in namespace '%s'...", alertName, alertNamespace)

			// Get the Alert
			currAlert, err := util.GetAlert(alertClient, alertNamespace, alertName)
			if err != nil {
				errors = append(errors, fmt.Errorf("error getting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err))
				continue
			}

			// Make changes to Spec
			currAlert.Spec.DesiredState = "STOP"
			// Update Alert
			_, err = util.UpdateAlert(alertClient,
				currAlert.Spec.Namespace, currAlert)
			if err != nil {
				errors = append(errors, fmt.Errorf("error stopping Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err))
				continue
			}

			log.Infof("successfully submitted stop Alert '%s' in namespace '%s'", alertName, alertNamespace)
		}
		if len(errors) > 0 {
			return fmt.Errorf("%v", errors)
		}
		return nil
	},
}

// stopBlackDuckCmd stops a Black Duck instance
var stopBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsysctl stop blackduck <name>\nsynopsysctl stop blackduck <name1> <name2>\nsynopsysctl stop blackduck <name> -n <namespace>\nsynopsysctl stop blackduck <name1> <name2> -n <namespace>",
	Short:         "Stop a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes one or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		errors := []error{}
		for _, bdArg := range args {
			blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, bdArg)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			log.Infof("stopping Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

			// Get the Black Duck
			currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
			if err != nil {
				errors = append(errors, fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err))
				continue
			}

			// Make changes to Spec
			currBlackDuck.Spec.DesiredState = "STOP"
			// Update Black Duck
			_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck.Spec.Namespace, currBlackDuck)
			if err != nil {
				errors = append(errors, fmt.Errorf("error stopping Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err))
				continue
			}

			log.Infof("successfully submitted stop Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		}
		if len(errors) > 0 {
			return fmt.Errorf("%v", errors)
		}
		return nil
	},
}

// stopOpsSightCmd stops an OpsSight instance
var stopOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME",
	Example:       "synopsysctl stop opssight <name>\nsynopsysctl stop opssight <name1> <name2>",
	Short:         "Stop an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes one or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		errors := []error{}
		for _, opsArg := range args {
			opsSightName, opsSightNamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsArg)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			log.Infof("stopping OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

			// Get the OpsSight
			currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightName)
			if err != nil {
				errors = append(errors, fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err))
				continue
			}

			// Make changes to Spec
			currOpsSight.Spec.DesiredState = "STOP"
			// Update OpsSight
			_, err = util.UpdateOpsSight(opsSightClient,
				currOpsSight.Spec.Namespace, currOpsSight)
			if err != nil {
				errors = append(errors, fmt.Errorf("error stopping OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err))
				continue
			}

			log.Infof("successfully submitted stop OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		}
		if len(errors) > 0 {
			return fmt.Errorf("%v", errors)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	stopCmd.AddCommand(stopAlertCmd)

	stopBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	stopCmd.AddCommand(stopBlackDuckCmd)

	stopCmd.AddCommand(stopOpsSightCmd)
}
