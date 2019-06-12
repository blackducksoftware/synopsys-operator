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

	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

// deleteAlertCmd deletes an Alert from the cluster
var deleteAlertCmd = &cobra.Command{
	Use:     "alert NAME...",
	Example: "synopsysctl delete alert <name>\nsynopsysctl delete alert <name1> <name2> <name3>\nsynopsysctl delete alert <name> -n <namespace>\nsynopsysctl delete alert <name1> <name2> <name3> -n <namespace>",
	Short:   "Delete one or many Alerts",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.AlertCRDName)
		if err != nil {
			return fmt.Errorf("unable to get the %s custom resource definition in your cluster due to %+v", util.AlertCRDName, err)
		}

		// Check Number of Arguments
		if crd.Spec.Scope != apiextensions.ClusterScoped && len(namespace) == 0 {
			return fmt.Errorf("namespace to delete an Alert instance need to be provided")
		}

		for _, alertName := range args {
			var operatorNamespace string
			if crd.Spec.Scope == apiextensions.ClusterScoped {
				if len(namespace) == 0 {
					operatorNamespace = alertName
				} else {
					operatorNamespace = namespace
				}
			} else {
				operatorNamespace = namespace
			}
			log.Infof("deleting an Alert '%s' instance in '%s' namespace...", alertName, operatorNamespace)
			err := alertClient.SynopsysV1().Alerts(operatorNamespace).Delete(alertName, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting an Alert %s instance in %s namespace due to %+v", alertName, operatorNamespace, err)
			}
			log.Infof("successfully deleted an Alert '%s' instance in '%s' namespace", alertName, operatorNamespace)
		}
		return nil
	},
}

// deleteBlackDuckCmd deletes a Black Duck from the cluster
var deleteBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME...",
	Example: "synopsysctl delete blackduck <name>\nsynopsysctl delete blackduck <name1> <name2> <name3>\nsynopsysctl delete blackduck <name> -n <namespace>\nsynopsysctl delete blackduck <name1> <name2> <name3> -n <namespace>",
	Short:   "Delete one or many Black Ducks",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.BlackDuckCRDName)
		if err != nil {
			return fmt.Errorf("unable to get the %s custom resource definition in your cluster due to %+v", util.BlackDuckCRDName, err)
		}

		// Check Number of Arguments
		if crd.Spec.Scope != apiextensions.ClusterScoped && len(namespace) == 0 {
			return fmt.Errorf("namespace to delete the Black Duck instance need to be provided")
		}

		for _, blackDuckName := range args {
			var operatorNamespace string
			if crd.Spec.Scope == apiextensions.ClusterScoped {
				if len(namespace) == 0 {
					operatorNamespace = blackDuckName
				} else {
					operatorNamespace = namespace
				}
			} else {
				operatorNamespace = namespace
			}
			log.Infof("deleting Black Duck '%s' instance in '%s' namespace...", blackDuckName, operatorNamespace)
			err := blackDuckClient.SynopsysV1().Blackducks(operatorNamespace).Delete(blackDuckName, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting Black Duck %s instance in %s namespace due to '%s'", blackDuckName, operatorNamespace, err)
			}
			log.Infof("successfully deleted Black Duck '%s' in %s namespace", blackDuckName, operatorNamespace)
		}
		return nil
	},
}

// deleteOpsSightCmd deletes an OpsSight from the cluster
var deleteOpsSightCmd = &cobra.Command{
	Use:     "opssight NAME...",
	Example: "synopsysctl delete opssight <name>\nsynopsysctl delete opssight <name1> <name2> <name3>",
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

func init() {
	//(PassCmd) deleteCmd.DisableFlagParsing = true // lets deleteCmd pass flags to kube/oc
	rootCmd.AddCommand(deleteCmd)

	// Add Delete Alert Command
	deleteAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to delete the resource(s)")
	deleteCmd.AddCommand(deleteAlertCmd)

	// Add Delete Black Duck Command
	deleteBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to delete the resource(s)")
	deleteCmd.AddCommand(deleteBlackDuckCmd)

	// Add Delete OpsSight Command
	deleteCmd.AddCommand(deleteOpsSightCmd)
}
