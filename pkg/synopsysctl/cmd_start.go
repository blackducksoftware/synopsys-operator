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
	Use:   "start",
	Short: "Start a Synopsys resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a resource")
	},
}

// startAlertCmd starts an Alert in the cluster
var startAlertCmd = &cobra.Command{
	Use:     "alert NAMES",
	Example: "synopsysctl start alert <name>\nsynopsysctl start alert <name> -n <namespace>",
	Short:   "Start an Alert",
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
		log.Infof("starting an Alert '%s' instance in '%s' namespace...", alertName, alertNamespace)

		// Get the Alert
		currAlert, err := util.GetAlert(alertClient, alertNamespace, alertName)
		if err != nil {
			log.Errorf("error getting %s Alert instance in %s namespace due to %+v", alertName, alertNamespace, err)
			return nil
		}

		// Make changes to Spec
		currAlert.Spec.DesiredState = ""
		// Update Alert
		_, err = util.UpdateAlert(alertClient, currAlert.Spec.Namespace, currAlert)
		if err != nil {
			log.Errorf("error updating the %s Alert instance in %s namespace due to %+v", alertName, alertNamespace, err)
			return nil
		}

		log.Infof("successfully started the '%s' Alert instance in '%s' namespace", alertName, alertNamespace)
		return nil
	},
}

// startBlackDuckCmd starts a Black Duck in the cluster
var startBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAMESPACE",
	Example: "synopsysctl start blackduck <name>\nsynopsysctl start blackduck <name> -n <namespace>",
	Short:   "Start a Blackduck",
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
		log.Infof("starting Black Duck '%s' instance in '%s' namespace...", blackDuckName, blackDuckNamespace)

		// Get the Black Duck
		currBlackDuck, err := util.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		// Make changes to Spec
		currBlackDuck.Spec.DesiredState = ""
		// Update Blackduck
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck.Spec.Namespace, currBlackDuck)
		if err != nil {
			log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}

		log.Infof("successfully started the '%s' Black Duck instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// startOpsSightCmd starts an OpsSight in the cluster
var startOpsSightCmd = &cobra.Command{
	Use:     "opssight NAMESPACE",
	Example: "synopsysctl start opssight <name>",
	Short:   "Start an OpsSight",
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
		log.Infof("starting OpsSight '%s' instance in '%s' namespace...", opsSightName, opsSightNamespace)

		// Get the OpsSight
		currOpsSight, err := util.GetOpsSight(opsSightClient, opsSightNamespace, opsSightName)

		if err != nil {
			log.Errorf("error getting %s OpsSight instance in %s namespace due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		// Make changes to Spec
		currOpsSight.Spec.DesiredState = ""
		// Update OpsSight
		_, err = util.UpdateOpsSight(opsSightClient, currOpsSight.Spec.Namespace, currOpsSight)
		if err != nil {
			log.Errorf("error updating the %s OpsSight instance in %s namespace due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}

		log.Infof("successfully started the '%s' OpsSight instance in '%s' namespace", opsSightName, opsSightNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to start the resource(s)")
	startCmd.AddCommand(startAlertCmd)

	startBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to start the resource(s)")
	startCmd.AddCommand(startBlackDuckCmd)

	startCmd.AddCommand(startOpsSightCmd)
}
