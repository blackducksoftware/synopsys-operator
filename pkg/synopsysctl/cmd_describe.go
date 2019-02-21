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

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Print a detailed description of the selected resource",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Describing a Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"describe"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			fmt.Printf("Error Describing the Resource: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var describeBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Describe an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Describing a Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		fmt.Println("Describing Blackduck")
		out, err := RunKubeCmd("describe", "blackduck", blackduckNamespace, "-n", blackduckNamespace)
		if err != nil {
			fmt.Printf("Error Describing the Blackduck: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var describeOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Describe an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Describing an OpsSight\n")
		// Read Commandline Parameters
		opsSightNamespace := args[0]

		fmt.Println("Describing OpsSight")
		out, err := RunKubeCmd("describe", "opssight", opsSightNamespace, "-n", opsSightNamespace)
		if err != nil {
			fmt.Printf("Error Describing the OpsSight: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var describeAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Describe an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Describing an OpsSight\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		fmt.Println("Describing Alert")
		out, err := RunKubeCmd("describe", "alert", alertNamespace, "-n", alertNamespace)
		if err != nil {
			fmt.Printf("Error Describing the Alert: %s\n", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

func init() {
	describeCmd.DisableFlagParsing = true
	rootCmd.AddCommand(describeCmd)
	describeCmd.AddCommand(describeBlackduckCmd)
	describeCmd.AddCommand(describeOpsSightCmd)
	describeCmd.AddCommand(describeAlertCmd)
}
