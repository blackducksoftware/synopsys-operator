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

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"edit"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			log.Errorf("Error Editing the Resource with KubeCmd: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Edit an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		err := RunKubeEditorCmd("edit", "blackduck", blackduckNamespace, "-n", blackduckNamespace)
		if err != nil {
			fmt.Printf("Error Editing the Blackduck: %s\n", err)
		}
	},
}

var editBlackduckAddPVCCmd = &cobra.Command{
	Use:   "addPVC",
	Short: "Add a PVC to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding PVC to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddEnvironCmd = &cobra.Command{
	Use:   "addEnviron",
	Short: "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Environ to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry",
	Short: "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image Registry to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddUIDCmd = &cobra.Command{
	Use:   "addUID",
	Short: "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image UID to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Edit an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an OpsSight\n")
		// Read Commandline Parameters
		opSightNamespace := args[0]

		err := RunKubeEditorCmd("edit", "opssight", opSightNamespace, "-n", opSightNamespace)
		if err != nil {
			fmt.Printf("Error Editing the OpsSight: %s\n", err)
		}
	},
}

var editOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry",
	Short: "Add an Internal Registry to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Internal Registryto OpsSight\n")
		// Read Commandline Parameters
		//opSightNamespace := args[0]
		return
	},
}

var editOpsSightAddHostCmd = &cobra.Command{
	Use:   "addHost",
	Short: "Add a Blackduck Host to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command only accepts 2 arguments - NAME  BLACKDUCK_HOST")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Blackduck Host to OpsSight\n")
		// Read Commandline Parameters
		//opSightNamespace := args[0]
		return
	},
}

var editAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Edit an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an Alert\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		err := RunKubeEditorCmd("edit", "alert", alertNamespace, "-n", alertNamespace)
		if err != nil {
			fmt.Printf("Error Editing the Alert: %s\n", err)
		}
	},
}

func init() {
	editCmd.DisableFlagParsing = true
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editBlackduckCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddPVCCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddEnvironCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddRegistryCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddUIDCmd)
	editCmd.AddCommand(editOpsSightCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddRegistryCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddHostCmd)
	editCmd.AddCommand(editAlertCmd)
}
