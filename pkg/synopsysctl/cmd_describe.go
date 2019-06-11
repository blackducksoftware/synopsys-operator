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

// Describe Command flag for -selector functionality
var describeSelector string

// describeCmd Show details of a Synopsys resource from your cluster
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show details of a Synopsys resource from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

// describeAlertCmd details of one or many Alerts
var describeAlertCmd = &cobra.Command{
	Use:     "alert [namespace]...",
	Example: "synopsysctl describe alerts\nsynopsysctl describe alert altnamespace\nsynopsysctl describe alerts altnamespace1 altnamespace2",
	Aliases: []string{"alerts"},
	Short:   "Show details of one or many Alerts",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing an Alert")
		kubectlCmd := []string{"describe", "alerts"}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, describeSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error describing Alert: %s - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeBlackDuckCmd Show details of one or many Black Ducks
var describeBlackDuckCmd = &cobra.Command{
	Use:     "blackduck [namespace]...",
	Example: "synopsysctl describe blackducks\nsynopsysctl describe blackduck bdnamespace\nnsynopsysctl describe blackducks bdnamespace1 bdnamespace2",
	Aliases: []string{"blackducks"},
	Short:   "Show details of one or many Black Ducks",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing a Blackduck")
		kubectlCmd := []string{"describe", "blackducks"}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, describeSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error describing Black Duck: %s - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeOpsSightCmd Show details of one or many OpsSights
var describeOpsSightCmd = &cobra.Command{
	Use:     "opssight [namespace]...",
	Example: "synopsysctl describe opssights\nsynopsysctl describe opssight opsnamespace\nsynopsysctl describe opssights opsnamespace1 opsnamespace2",
	Aliases: []string{"opssights"},
	Short:   "Show details of one or many OpsSights",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing an OpsSight")
		kubectlCmd := []string{"describe", "opssights"}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, describeSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error describing OpsSight: %s - %s", out, err)
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
	describeBlackDuckCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeBlackDuckCmd)

	describeOpsSightCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeOpsSightCmd)

	describeAlertCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeAlertCmd)
}
