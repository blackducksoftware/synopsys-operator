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
		num_args := 1
		if len(args) != num_args {
			return fmt.Errorf("Must pass Namespace")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edit called")
	},
}

var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Edit an instance of Blackduck",
	Run: func(cmd *cobra.Command, args []string) {
		// Read Commandline Parameters
		namespace = args[0]

		fmt.Println("Editing Blackduck")
		err := RunKubeEditorCmd("edit", "blackduck", namespace, "-n", namespace)
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
		namespace = args[0]

		fmt.Println("Editing OpsSight")
		err := RunKubeEditorCmd("edit", "opssight", namespace, "-n", namespace)
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
		namespace = args[0]

		fmt.Println("Editing Alert")
		err := RunKubeEditorCmd("edit", "alert", namespace, "-n", namespace)
		if err != nil {
			fmt.Printf("Error Editing the Alert with KubeCmd: %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editBlackduckCmd)
	editCmd.AddCommand(editOpsSightCmd)
	editCmd.AddCommand(editAlertCmd)
}
