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
)

// editCmd edits non-synopsys resources
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Not a Valid Command")
	},
}

// editBlackduckCmd edits a Blackduck by updating the spec
// or using the kube/oc editor
var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Edit an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackduckName := args[0]
		log.Debugf("editing Black Duck %s instance instance...", blackduckName)
		err := RunKubeEditorCmd(restconfig, kube, openshift, "edit", "blackduck", blackduckName, "-n", blackduckName)
		if err != nil {
			log.Errorf("error editing the Black Duck: %s", err)
			return nil
		}
		log.Infof("successfully edited Black Duck: '%s'", blackduckName)
		return nil
	},
}

// editOpsSightCmd edits an OpsSight by updating the spec
// or using the kube/oc editor
var editOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Edit an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
		log.Debugf("Editing OpsSight %s...", opsSightName)
		err := RunKubeEditorCmd(restconfig, kube, openshift, "edit", "opssight", opsSightName, "-n", opsSightName)
		if err != nil {
			log.Errorf("error Editing the OpsSight: %s", err)
			return nil
		}
		log.Infof("successfully edited OpsSight: '%s'", opsSightName)
		return nil
	},
}

// editAlertCmd edits an Alert by updating the spec
// or using the kube/oc editor
var editAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Edit an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName := args[0]
		log.Infof("editing Alert %s instance...", alertName)
		err := RunKubeEditorCmd(restconfig, kube, openshift, "edit", "alert", alertName, "-n", alertName)
		if err != nil {
			log.Errorf("error editing the Alert: %s", err)
			return nil
		}
		log.Infof("successfully edited Alert: '%s'", alertName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editBlackduckCmd)
	editCmd.AddCommand(editOpsSightCmd)
	editCmd.AddCommand(editAlertCmd)
}
