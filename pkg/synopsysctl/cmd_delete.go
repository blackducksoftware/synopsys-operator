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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteCmd deletes a resource from the cluster
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a Synopsys Resource from your cluster",
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	// Display synopsysctl's Help instead of sending to oc/kubectl
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Deleting a Non-Synopsys Resource")
		//(PassCmd) kubeCmdArgs := append([]string{"delete"}, args...)
		//(PassCmd) out, err := util.RunKubeCmd(restconfig, kube, openshift, kubeCmdArgs...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	log.Errorf("Error Deleting the Resource: %s", out)
		//(PassCmd) 	return nil
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
	},
}

// deleteBlackduckCmd deletes a Blackduck from the cluster
var deleteBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Delete a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Deleting a Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		// Delete Blackduck with Client
		err := blackduckClient.SynopsysV1().Blackducks(blackduckNamespace).Delete(blackduckNamespace, &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Error deleting the Blackduck: %s", err)
			return nil
		}
		return nil
	},
}

// deleteOpsSightCmd deletes an OpsSight from the cluster
var deleteOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Delete an OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Deleting an OpsSight\n")
		// Read Commandline Parameters
		opsSightNamespace := args[0]

		// Delete OpsSight with Client
		err := opssightClient.SynopsysV1().OpsSights(opsSightNamespace).Delete(opsSightNamespace, &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Error deleting the OpsSight: %s", err)
			return nil
		}
		return nil
	},
}

// deleteAlertCmd deletes an Alert from the cluster
var deleteAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Delete an Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Deleting an Alert\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		// Delete Alert with Client
		err := alertClient.SynopsysV1().Alerts(alertNamespace).Delete(alertNamespace, &metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Error deleting the Alert: %s", err)
			return nil
		}
		return nil
	},
}

func init() {
	//(PassCmd) deleteCmd.DisableFlagParsing = true // lets deleteCmd pass flags to kube/oc
	rootCmd.AddCommand(deleteCmd)

	// Add Delete Commands
	deleteCmd.AddCommand(deleteBlackduckCmd)
	deleteCmd.AddCommand(deleteOpsSightCmd)
	deleteCmd.AddCommand(deleteAlertCmd)
}
