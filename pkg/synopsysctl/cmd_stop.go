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
	Use:   "stop",
	Short: "Stop a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a resource")
	},
}

// stopAlertCmd stops an Alert instance
var stopAlertCmd = &cobra.Command{
	Use:     "alert NAME",
	Example: "synopsysctl stop alert <name>\nsynopsysctl stop alert <name> -n <namespace>",
	Short:   "Stops an Alert",
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
		log.Infof("stopping an Alert '%s' instance in '%s' namespace...", alertName, alertNamespace)

		// Get the Alert
		currAlert, err := util.GetAlert(alertClient, alertNamespace, alertNamespace)
		if err != nil {
			log.Errorf("error stopping Alert '%s' due to %+v", alertNamespace, err)
			return nil
		}

		// Make changes to Spec
		currAlert.Spec.DesiredState = "STOP"
		// Update Alert
		_, err = util.UpdateAlert(alertClient,
			currAlert.Spec.Namespace, currAlert)
		if err != nil {
			log.Errorf("error stopping the %s Alert instance in %s namespace due to %+v", alertName, alertNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' Alert instance in '%s' namespace", alertName, alertNamespace)
		return nil
	},
}

// stopBlackDuckCmd stops a Black Duck instance
var stopBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME",
	Example: "synopsysctl stop blackduck <name>\nsynopsysctl stop blackduck <name> -n <namespace>",
	Short:   "Stops a Black Duck",
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
		log.Infof("stopping Black Duck '%s' instance in '%s' namespace...", blackDuckName, blackDuckNamespace)

		// Get the Black Duck
		currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackDuck.Spec.DesiredState = "STOP"
		// Update Black Duck
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck.Spec.Namespace, currBlackDuck)
		if err != nil {
			log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		log.Infof("successfully started the '%s' Black Duck instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// stopOpsSightCmd stops an OpsSight instance
var stopOpsSightCmd = &cobra.Command{
	Use:     "opssight NAME",
	Example: "synopsysctl stop opssight <name>",
	Short:   "Stops an OpsSight",
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
		log.Infof("stopping OpsSight '%s' instance in '%s' namespace...", opsSightName, opsSightNamespace)

		// Get the OpsSight
		currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightName)
		if err != nil {
			log.Errorf("error getting %s OpsSight instance in %s namespace due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		// Make changes to Spec
		currOpsSight.Spec.DesiredState = "STOP"
		// Update OpsSight
		_, err = util.UpdateOpsSight(opsSightClient,
			currOpsSight.Spec.Namespace, currOpsSight)
		if err != nil {
			log.Errorf("error updating the %s OpsSight instance in %s namespace due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' OpsSight instance in '%s' namespace", opsSightName, opsSightNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to stop the resource(s)")
	stopCmd.AddCommand(stopAlertCmd)

	stopBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to stop the resource(s)")
	stopCmd.AddCommand(stopBlackDuckCmd)

	stopCmd.AddCommand(stopOpsSightCmd)
}
