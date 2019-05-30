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
	Short: "Remove a Synopsys resource from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

// deleteBlackDuckCmd deletes a Black Duck from the cluster
var deleteBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAMESPACE...",
	Example: "synopsysctl delete blackduck bdnamespace\nsynopsysctl delete blackduck bdnamespace1 bdnamespace2 bdnamespace3",
	Short:   "Delete one or many Black Ducks",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, blackDuckNamespace := range args {
			log.Infof("deleting Black Duck %s...", blackDuckNamespace)
			err := blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Delete(blackDuckNamespace, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting Black Duck %s: '%s'", blackDuckNamespace, err)
			}
			log.Infof("successfully deleted Black Duck: %s", blackDuckNamespace)
		}
		return nil
	},
}

// deleteOpsSightCmd deletes an OpsSight from the cluster
var deleteOpsSightCmd = &cobra.Command{
	Use:     "opssight NAMESPACE...",
	Example: "synopsysctl delete opssight opsnamespace\nsynopsysctl delete opssight opsnamespace1 opsnamespace2 opsnamespace3",
	Short:   "Delete one or many OpsSights",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, opsSightNamespace := range args {
			log.Infof("deleting OpsSight %s...", opsSightNamespace)
			err := opsSightClient.SynopsysV1().OpsSights(opsSightNamespace).Delete(opsSightNamespace, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting OpsSight %s: '%s'", opsSightNamespace, err)
			}
			log.Infof("successfully deleted OpsSight: %s", opsSightNamespace)
		}
		return nil
	},
}

// deleteAlertCmd deletes an Alert from the cluster
var deleteAlertCmd = &cobra.Command{
	Use:     "alert NAMESPACE...",
	Example: "synopsysctl delete alert altnamespace\nsynopsysctl delete alert altnamespace1 altnamespace2 altnamespace3",
	Short:   "Delete one or many Alerts",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, alertNamespace := range args {
			log.Infof("deleting Alert %s...", alertNamespace)
			err := alertClient.SynopsysV1().Alerts(alertNamespace).Delete(alertNamespace, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting Alert %s: %s", alertNamespace, err)
			}
			log.Infof("successfully deleted Alert: %s", alertNamespace)
		}
		return nil
	},
}

func init() {
	//(PassCmd) deleteCmd.DisableFlagParsing = true // lets deleteCmd pass flags to kube/oc
	rootCmd.AddCommand(deleteCmd)

	// Add Delete Commands
	deleteCmd.AddCommand(deleteBlackDuckCmd)
	deleteCmd.AddCommand(deleteOpsSightCmd)
	deleteCmd.AddCommand(deleteAlertCmd)
}
