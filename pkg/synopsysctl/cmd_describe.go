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

// describeCmd shows details of Synopsys resources from your cluster
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show details of Synopsys resources from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// describeAlertCmd shows details of one or many Alert instances
var describeAlertCmd = &cobra.Command{
	Use:     "alert [NAME...]",
	Example: "synopsysctl describe alerts\nsynopsysctl describe alert <name>\nsynopsysctl describe alerts <name1> <name2>\nsynopsysctl describe alerts -n <namespace>\nsynopsysctl describe alert <name> -n <namespace>\nsynopsysctl describe alerts <name1> <name2> -n <namespace>",
	Aliases: []string{"alerts"},
	Short:   "Show details of one or many Alert instances",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing Alert instances")
		kubectlCmd := []string{"describe", "alerts"}
		if len(namespace) > 0 {
			kubectlCmd = append(kubectlCmd, "-n", namespace)
		}
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

// describeBlackDuckCmd shows details of one or many Black Duck instances
var describeBlackDuckCmd = &cobra.Command{
	Use:     "blackduck [NAME...]",
	Example: "synopsysctl describe blackducks\nsynopsysctl describe blackduck <name>\nsynopsysctl describe blackducks <name1> <name2>\nsynopsysctl describe blackducks -n <namespace>\nsynopsysctl describe blackduck <name> -n <namespace>\nsynopsysctl describe blackducks <name1> <name2> -n <namespace>",
	Aliases: []string{"blackducks", "bds", "bd"},
	Short:   "Show details of one or many Black Duck instances",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing Black Duck instances")
		kubectlCmd := []string{"describe", "blackducks"}
		if len(namespace) > 0 {
			kubectlCmd = append(kubectlCmd, "-n", namespace)
		}
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

// describeOpsSightCmd shows details of one or many OpsSight instances
var describeOpsSightCmd = &cobra.Command{
	Use:     "opssight [NAME...]",
	Example: "synopsysctl describe opssights\nsynopsysctl describe opssight <name>\nsynopsysctl describe opssights <name1> <name2>\nsynopsysctl describe opssights -n <namespace>\nsynopsysctl describe opssight <name> -n <namespace>\nsynopsysctl describe opssights <name1> <name2> -n <namespace>",
	Aliases: []string{"opssights", "ops"},
	Short:   "Show details of one or many OpsSight instances",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("describing OpsSight instances")
		kubectlCmd := []string{"describe", "opssights"}
		if len(namespace) > 0 {
			kubectlCmd = append(kubectlCmd, "-n", namespace)
		}
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
	describeBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	describeBlackDuckCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeBlackDuckCmd)

	describeOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	describeOpsSightCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeOpsSightCmd)

	describeAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	describeAlertCmd.Flags().StringVarP(&describeSelector, "selector", "l", describeSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	describeCmd.AddCommand(describeAlertCmd)
}
