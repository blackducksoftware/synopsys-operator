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

	polarisreporting "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var promptAnswerYes bool

// deleteCmd deletes a resource from the cluster
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove Synopsys resources from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// deleteAlertCmd deletes Alert instances from the cluster
var deleteAlertCmd = &cobra.Command{
	Use:           "alert NAME...",
	Example:       "synopsysctl delete alert <name>\nsynopsysctl delete alert <name1> <name2> <name3>\nsynopsysctl delete alert <name> -n <namespace>\nsynopsysctl delete alert <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many Alert instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, alertName := range args {
			alertNamespace, crdNamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, alertName)
			if err != nil {
				return err
			}
			log.Infof("deleting Alert '%s' in namespace '%s'...", alertName, alertNamespace)
			err = util.DeleteAlert(alertClient, alertName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
			}
			log.Infof("successfully submitted delete Alert '%s' in namespace '%s'", alertName, alertNamespace)
		}
		return nil
	},
}

// deleteBlackDuckCmd deletes Black Duck instances from the cluster
var deleteBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME...",
	Example:       "synopsysctl delete blackduck <name>\nsynopsysctl delete blackduck <name1> <name2> <name3>\nsynopsysctl delete blackduck <name> -n <namespace>\nsynopsysctl delete blackduck <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many Black Duck instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, blackDuckName := range args {
			blackDuckNamespace, crdNamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
			if err != nil {
				return err
			}
			log.Infof("deleting Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			err = util.DeleteBlackduck(blackDuckClient, blackDuckName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting Black Duck '%s' in namespace '%s' due to '%s'", blackDuckName, blackDuckNamespace, err)
			}
			log.Infof("successfully submitted delete Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		}
		return nil
	},
}

// deleteOpsSightCmd deletes OpsSight instances from the cluster
var deleteOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME...",
	Example:       "synopsysctl delete opssight <name>\nsynopsysctl delete opssight <name1> <name2> <name3>\nsynopsysctl delete opssight <name> -n <namespace>\nsynopsysctl delete opssight <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many OpsSight instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, opsSightName := range args {
			opsSightNamespace, crdNamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
			if err != nil {
				return err
			}
			log.Infof("deleting OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			err = util.DeleteOpsSight(opsSightClient, opsSightName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting OpsSight '%s' in namespace '%s' due to '%s'", opsSightName, opsSightNamespace, err)
			}
			log.Infof("successfully submitted delete OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		}
		return nil
	},
}

// deletePolarisCmd deletes a Polaris instance
var deletePolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsysctl delete polaris -n <namespace>",
	Short:         "Delete a Polaris instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delete Polaris Resources
		err := util.DeleteWithHelm3(polarisName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to delete Polaris resources: %+v", err)
		}

		log.Infof("Polaris has been successfully Deleted!")
		return nil
	},
}

// deletePolarisReportingCmd deletes a Polaris-Reporting instance
var deletePolarisReportingCmd = &cobra.Command{
	Use:           "polaris-reporting",
	Example:       "synopsysctl delete polaris-reportinng -n <namespace>",
	Short:         "Delete a Polaris-Reporting instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get Secret For the GCP Key
		gcpServiceAccountSecrets, err := polarisreporting.GetPolarisReportingSecrets(namespace, "EMPTY_DATA")
		if err != nil {
			return fmt.Errorf("failed to generate GCP Service Account Secrets: %+v", err)
		}

		// Delete the Secret
		err = KubectlDeleteRuntimeObjects(gcpServiceAccountSecrets)
		if err != nil {
			return fmt.Errorf("failed to delete the gcpServiceAccount Secrets: %s", err)
		}

		// Delete Polaris-Reporting Resources
		err = util.DeleteWithHelm3(polarisReportingName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to delete Polaris-Reporting resources: %+v", err)
		}

		log.Infof("Polaris-Reporting has been successfully Deleted!")
		return nil
	},
}

// deleteBDBACmd deletes a BDBA instance
var deleteBDBACmd = &cobra.Command{
	Use:           "bdba",
	Example:       "synopsysctl delete bdba -n <namespace>",
	Short:         "Delete a BDBA instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delete the extra resources like secrets...
		// ...

		// Delete Resources
		err := util.DeleteWithHelm3(bdbaName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to delete BDBA resources: %+v", err)
		}

		log.Infof("BDBA has been successfully Deleted!")
		return nil
	},
}

func init() {
	// deletePolarisReportingCobraHelper = *polarisreporting.NewArgsFromCobraFlags()

	//(PassCmd) deleteCmd.DisableFlagParsing = true // lets deleteCmd pass flags to kube/oc
	rootCmd.AddCommand(deleteCmd)

	// Add Delete Alert Command
	deleteAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteAlertCmd)

	// Add Delete Black Duck Command
	deleteBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteBlackDuckCmd)

	// Add Delete OpsSight Command
	deleteOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteOpsSightCmd)

	// Add Delete Polaris Command
	deletePolarisCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deletePolarisCmd.Flags().BoolVarP(&promptAnswerYes, "yes", "y", promptAnswerYes, "Automatic yes to prompts")
	addbaseURLFlag(deletePolarisCmd)
	deleteCmd.AddCommand(deletePolarisCmd)

	// Add Delete Polaris-Reporting Command
	deletePolarisReportingCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(deletePolarisReportingCmd.Flags(), "namespace")
	deleteCmd.AddCommand(deletePolarisReportingCmd)

	// Add Delete BDBA Command
	deleteBDBACmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(deleteBDBACmd.Flags(), "namespace")
	deleteCmd.AddCommand(deleteBDBACmd)
}
