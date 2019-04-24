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

// describeCmd prints the CRD for a resource
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show details of a Synopsys Resource from you cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Not a Valid Command")
	},
}

// describeBlackduckCmd prints the CRD for a Blackduck
var describeBlackduckCmd = &cobra.Command{
	Use:     "blackduck [NAME]",
	Aliases: []string{"blackducks"},
	Short:   "Show details of one or many Black Ducks",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("this command takes up to 1 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing a Blackduck")
		kCmd := []string{"describe", "blackducks"}
		if len(args) > 0 {
			kCmd = append(kCmd, args[0])
		}
		out, err := RunKubeCmd(restconfig, kube, openshift, kCmd...)
		if err != nil {
			log.Errorf("error describing the Black Duck: %s - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeOpsSightCmd prints the CRD for an OpsSight
var describeOpsSightCmd = &cobra.Command{
	Use:     "opssight [NAME]",
	Aliases: []string{"opssights"},
	Short:   "Show details of one or many OpsSights",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("this command takes up to 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing an OpsSight")
		kCmd := []string{"describe", "opssights"}
		if len(args) > 0 {
			kCmd = append(kCmd, args[0])
		}
		out, err := RunKubeCmd(restconfig, kube, openshift, kCmd...)
		if err != nil {
			log.Errorf("error describing the OpsSight: %s - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// describeAlertCmd prints the CRD for an Alert
var describeAlertCmd = &cobra.Command{
	Use:     "alert [NAME]",
	Aliases: []string{"alerts"},
	Short:   "Show details of one or many Alerts",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("this command takes up to 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Describing an Alert")
		kCmd := []string{"describe", "alerts"}
		if len(args) > 0 {
			kCmd = append(kCmd, args[0])
		}
		out, err := RunKubeCmd(restconfig, kube, openshift, kCmd...)
		if err != nil {
			log.Errorf("error describing the Alert: %s - %s", out, err)
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
