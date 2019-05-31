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
	Short: "Stops a Synopsys resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a resource")
	},
}

// stopAlertCmd stops an Alert in the cluster
var stopAlertCmd = &cobra.Command{
	Use:     "alert NAMESPACE",
	Example: "synopsysctl stop alert altnamespace",
	Short:   "Stops an Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertNamespace := args[0]
		log.Infof("stopping Alert %s...", alertNamespace)

		// Get the Alert
		currAlert, err := util.GetAlert(alertClient, alertNamespace, alertNamespace)
		if err != nil {
			log.Errorf("error getting %s Alert instance due to %+v", alertNamespace, err)
			return nil
		}

		// Make changes to Spec
		currAlert.Spec.DesiredState = "STOP"
		// Update Alert
		_, err = util.UpdateAlert(alertClient,
			currAlert.Spec.Namespace, currAlert)
		if err != nil {
			log.Errorf("error stopping the %s Alert instance due to %+v", alertNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' Alert instance", alertNamespace)
		return nil
	},
}

// stopBlackDuckCmd stops a Black Duck in the cluster
var stopBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAMESPACE",
	Example: "synopsysctl stop blackduck bdnamespace",
	Short:   "Stops a Black Duck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		log.Infof("stopping Black Duck %s...", blackDuckNamespace)

		// Get the Black Duck
		currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackDuck.Spec.DesiredState = "STOP"
		// Update Black Duck
		_, err = util.UpdateBlackduck(blackDuckClient,
			currBlackDuck.Spec.Namespace, currBlackDuck)
		if err != nil {
			log.Errorf("error stopping the %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' Black Duck instance", blackDuckNamespace)
		return nil
	},
}

// stopOpsSightCmd stops an OpsSight in the cluster
var stopOpsSightCmd = &cobra.Command{
	Use:     "opssight NAMESPACE",
	Example: "synopsysctl stop opssight opsnamespace",
	Short:   "Stops an OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		log.Infof("stopping OpsSight %s...", opsSightNamespace)

		// Get the OpsSight
		currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("error getting %s OpsSight instance due to %+v", opsSightNamespace, err)
			return nil
		}

		// Make changes to Spec
		currOpsSight.Spec.DesiredState = "STOP"
		// Update OpsSight
		_, err = util.UpdateOpsSight(opsSightClient,
			currOpsSight.Spec.Namespace, currOpsSight)
		if err != nil {
			log.Errorf("error stopping the %s OpsSight instance due to %+v", opsSightNamespace, err)
			return nil
		}

		log.Infof("successfully stopped the '%s' OpsSight instance", opsSightNamespace)
		return nil
	},
}

func init() {
	stopCmd.AddCommand(stopAlertCmd)
	stopCmd.AddCommand(stopBlackDuckCmd)
	stopCmd.AddCommand(stopOpsSightCmd)
	rootCmd.AddCommand(stopCmd)
}
