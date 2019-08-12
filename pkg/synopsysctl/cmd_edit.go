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
)

// editCmd edits Synopsys resources
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// editAlertCmd edits an Alert instance by using the kube/oc editor
var editAlertCmd = &cobra.Command{
	Use:           "alert NAME",
	Example:       "synopsysctl edit alert <name>\nsynopsysctl edit alert <name> -n <namespace>",
	Short:         "Edit an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName, alertNamespace, _, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, args[0])
		if err != nil {
			return err
		}
		log.Infof("editing Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		err = RunKubeEditorCmd(restconfig, kubeClient, "edit", "alert", alertName, "-n", alertNamespace)
		if err != nil {
			return fmt.Errorf("error editing Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
		}
		log.Infof("successfully edited Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		return nil
	},
}

// editBlackDuckCmd edits a Black Duck instance by using the kube/oc editor
var editBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsysctl edit blackduck <name>\nsynopsysctl edit blackduck <name> -n <namespace>",
	Short:         "Edit a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		log.Infof("editing Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		err = RunKubeEditorCmd(restconfig, kubeClient, "edit", "blackduck", blackDuckName, "-n", blackDuckNamespace)
		if err != nil {
			return fmt.Errorf("error editing Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		log.Infof("successfully edited Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// editOpsSightCmd edits an OpsSight instance by using the kube/oc editor
var editOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME",
	Example:       "synopsysctl edit opssight <name>\nsynopsysctl edit opssight <name> -n <namespace>",
	Short:         "Edit an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName, opsSightNamespace, _, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		log.Infof("editing OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		err = RunKubeEditorCmd(restconfig, kubeClient, "edit", "opssight", opsSightName, "-n", opsSightNamespace)
		if err != nil {
			return fmt.Errorf("error editing OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		log.Infof("successfully edited OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	editCmd.AddCommand(editAlertCmd)

	editBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	editCmd.AddCommand(editBlackDuckCmd)

	editOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	editCmd.AddCommand(editOpsSightCmd)
}
