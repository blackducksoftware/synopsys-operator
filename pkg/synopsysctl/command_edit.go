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

	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	Args: func(cmd *cobra.Command, args []string) error {
		numArgs := 1
		if len(args) < numArgs {
			return fmt.Errorf("Not enough arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Describing Non-Synopsys Resource")
		kubeCmdArgs := append([]string{"describe"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			fmt.Printf("Error Describing the Resource with KubeCmd: %s\n", err)
		}
		fmt.Printf("%+v\n", out)
	},
}

var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Edit an instance of Blackduck",
	Run: func(cmd *cobra.Command, args []string) {
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		fmt.Println("Editing Blackduck")
		err := RunKubeEditorCmd("edit", "blackduck", blackduckNamespace, "-n", blackduckNamespace)
		if err != nil {
			fmt.Printf("Error Editing the Blackduck with KubeCmd: %s\n", err)
		}
	},
}

var editOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Edit an instance of OpsSight",
	Run: func(cmd *cobra.Command, args []string) {
		// Read Commandline Parameters
		opSightNamespace := args[0]

		fmt.Println("Editing OpsSight")
		err := RunKubeEditorCmd("edit", "opssight", opSightNamespace, "-n", opSightNamespace)
		if err != nil {
			fmt.Printf("Error Editing the OpsSight with KubeCmd: %s\n", err)
		}
	},
}

var editAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Edit an instance of Alert",
	Run: func(cmd *cobra.Command, args []string) {
		// Read Commandline Parameters
		alertNamespace := args[0]

		fmt.Println("Editing Alert")
		err := RunKubeEditorCmd("edit", "alert", alertNamespace, "-n", alertNamespace)
		if err != nil {
			fmt.Printf("Error Editing the Alert with KubeCmd: %s\n", err)
		}
	},
}

func init() {
	editCmd.DisableFlagParsing = true
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editBlackduckCmd)
	editCmd.AddCommand(editOpsSightCmd)
	editCmd.AddCommand(editAlertCmd)
}
