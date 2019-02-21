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
	Args: func(cmd *cobra.Command, args []string) error {
		numArgs := 1
		if len(args) < numArgs {
			return fmt.Errorf("Not enough arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Getting a Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"get"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			fmt.Printf("Error Getting the Resource: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var getBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Get a list of Blackducks in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Getting a Blackduck\n")
		out, err := RunKubeCmd("get", "blackducks")
		if err != nil {
			fmt.Printf("Error getting Blackducks: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var getOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Get a list of OpsSights in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Getting an OpsSight\n")
		out, err := RunKubeCmd("get", "opssights")
		if err != nil {
			fmt.Printf("Error getting OpsSights: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var getAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Get a list of Alerts in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Getting an OpsSight\n")
		out, err := RunKubeCmd("get", "alerts")
		if err != nil {
			fmt.Printf("Error getting Alerts with KubeCmd: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

func init() {
	getCmd.DisableFlagParsing = true
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getBlackduckCmd)
	getCmd.AddCommand(getOpsSightCmd)
	getCmd.AddCommand(getAlertCmd)
}
