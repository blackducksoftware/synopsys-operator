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

// describeCmd prints the CRD for a resource
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Print a detailed description of the selected resource",
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	// Display synopsysctl's Help instead of sending to oc/kubectl
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Describing a Non-Synopsys Resource\n")
		//(PassCmd) kubeCmdArgs := append([]string{"describe"}, args...)
		//(PassCmd) out, err := util.RunKubeCmd(restconfig, kube, openshift, kubeCmdArgs...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	log.Errorf("Error Describing the Resource: %s", out)
		//(PassCmd) 	return nil
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
	},
}

// describeBlackduckCmd prints the CRD for a Blackduck
var describeBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Describe an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing a Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		out, err := util.RunKubeCmd(restconfig, kube, openshift, "describe", "blackduck", blackduckNamespace, "-n", blackduckNamespace)
		if err != nil {
			log.Errorf("Error Describing the Blackduck: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeOpsSightCmd prints the CRD for an OpsSight
var describeOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Describe an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing an OpsSight\n")
		// Read Commandline Parameters
		opsSightNamespace := args[0]

		out, err := util.RunKubeCmd(restconfig, kube, openshift, "describe", "opssight", opsSightNamespace, "-n", opsSightNamespace)
		if err != nil {
			log.Errorf("Error Describing the OpsSight: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeAlertCmd prints the CRD for an Alert
var describeAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Describe an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing an Alert\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		out, err := util.RunKubeCmd(restconfig, kube, openshift, "describe", "alert", alertNamespace, "-n", alertNamespace)
		if err != nil {
			log.Errorf("Error Describing the Alert: %s\n", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

func init() {
	//(PassCmd) describeCmd.DisableFlagParsing = true // lets describeCmd pass flags to kube/oc
	rootCmd.AddCommand(describeCmd)

	// Add Commands
	describeCmd.AddCommand(describeBlackduckCmd)
	describeCmd.AddCommand(describeOpsSightCmd)
	describeCmd.AddCommand(describeAlertCmd)
}
