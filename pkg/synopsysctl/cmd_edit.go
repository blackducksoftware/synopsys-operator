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

// editCmd edits non-Synopsys resources
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

// editAlertCmd edits an Alert by updating the spec
// or using the kube/oc editor
var editAlertCmd = &cobra.Command{
	Use:     "alert NAME",
	Example: "synopsysctl edit alert <name>\nsynopsysctl edit alert <name> -n <namespace>",
	Short:   "Edit an instance of Alert",
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
		log.Infof("editing an Alert '%s' instance in '%s' namespace...", alertName, alertNamespace)
		err = RunKubeEditorCmd(restconfig, "edit", "alert", alertName, "-n", alertNamespace)
		if err != nil {
			log.Errorf("error editing an Alert '%s' instance in '%s' namespace due to %+v", alertName, alertNamespace, err)
			return nil
		}
		log.Infof("successfully edited '%s' Alert instance in '%s' namespace...", alertName, alertNamespace)
		return nil
	},
}

// editBlackDuckCmd edits a Black Duck by updating the spec
// or using the kube/oc editor
var editBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME",
	Example: "synopsysctl edit blackduck <name>\nsynopsysctl edit blackduck <name> -n <namespace>",
	Short:   "Edit a Black Duck instance",
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
		log.Infof("editing Black Duck '%s' instance in '%s' namespace...", blackDuckName, blackDuckNamespace)
		err = RunKubeEditorCmd(restconfig, "edit", "blackduck", blackDuckName, "-n", blackDuckNamespace)
		if err != nil {
			log.Errorf("error editing Black Duck '%s' instance in '%s' namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		log.Infof("successfully edited '%s' Black Duck instance in '%s' namespace...", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// editOpsSightCmd edits an OpsSight by updating the spec
// or using the kube/oc editor
var editOpsSightCmd = &cobra.Command{
	Use:     "opssight NAME",
	Example: "synopsysctl edit opssight <name>\nsynopsysctl edit opssight <name> -n <namespace>",
	Short:   "Edit an instance of OpsSight",
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
		log.Infof("editing an OpsSight '%s' instance in '%s' namespace...", opsSightName, opsSightNamespace)
		err = RunKubeEditorCmd(restconfig, "edit", "opssight", opsSightName, "-n", opsSightNamespace)
		if err != nil {
			log.Errorf("error editing OpsSight '%s' instance in '%s' namespace due to %+v", opsSightName, opsSightNamespace, err)
			return nil
		}
		log.Infof("successfully edited '%s' OpsSight instance in '%s' namespace...", opsSightName, opsSightNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to edit the resource(s)")
	editCmd.AddCommand(editAlertCmd)

	editBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to edit the resource(s)")
	editCmd.AddCommand(editBlackDuckCmd)

	editOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to edit the resource(s)")
	editCmd.AddCommand(editOpsSightCmd)
}
