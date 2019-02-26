// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List Synopsys Resources in your cluster",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting a Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"get"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			log.Errorf("Error Getting the Resource: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

var getBlackduckCmd = &cobra.Command{
	Use:     "blackduck",
	Aliases: []string{"blackducks"},
	Short:   "Get a list of Blackducks in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting Blackducks\n")
		out, err := RunKubeCmd("get", "blackducks")
		if err != nil {
			log.Errorf("Error getting Blackducks: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

var getOpsSightCmd = &cobra.Command{
	Use:     "opssight",
	Aliases: []string{"opssights"},
	Short:   "Get a list of OpsSights in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting OpsSights\n")
		out, err := RunKubeCmd("get", "opssights")
		if err != nil {
			log.Errorf("Error getting OpsSights: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

var getAlertCmd = &cobra.Command{
	Use:     "alert",
	Aliases: []string{"alerts"},
	Short:   "Get a list of Alerts in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting Alerts\n")
		out, err := RunKubeCmd("get", "alerts")
		if err != nil {
			log.Errorf("Error getting Alerts with KubeCmd: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

func init() {
	getCmd.DisableFlagParsing = true // lets getCmd pass flags to kube/oc
	rootCmd.AddCommand(getCmd)

	// Add Commands
	getCmd.AddCommand(getBlackduckCmd)
	getCmd.AddCommand(getOpsSightCmd)
	getCmd.AddCommand(getAlertCmd)
}
