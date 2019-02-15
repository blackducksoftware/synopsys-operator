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

package ctl

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")
	},
}

var getBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Get a list of Blackducks in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Getting Blackducks")
		out, err := RunKubeCmd("get", "blackducks")
		if err != nil {
			fmt.Printf("Error getting Blackducks with KubeCmd: %s\n", err)
		}
		fmt.Printf("%+v\n", out)
	},
}

var getOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Get a list of OpsSights in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Getting OpsSights")
		out, err := RunKubeCmd("get", "opssights")
		if err != nil {
			fmt.Printf("Error getting OpsSights with KubeCmd: %s\n", err)
		}
		fmt.Printf("%+v\n", out)
	},
}

var getAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Get a list of Alerts in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Getting Alerts")
		out, err := RunKubeCmd("get", "alerts")
		if err != nil {
			fmt.Printf("Error getting Alerts with KubeCmd: %s\n", err)
		}
		fmt.Printf("%+v\n", out)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getBlackduckCmd)
	getCmd.AddCommand(getOpsSightCmd)
	getCmd.AddCommand(getAlertCmd)
}
