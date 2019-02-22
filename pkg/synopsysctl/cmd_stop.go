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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Stop Command Defaults
var stopNamespace = "synopsys-operator"

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Removes the Synopsys-Operator and CRDs from Cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Stopping the Synopsys-Operator: %s\n", stopNamespace)
		// Read Commandline Parameters
		if len(args) == 1 {
			stopNamespace = args[0]
		}
		out, err := RunKubeCmd("delete", "ns", stopNamespace)
		if err != nil {
			fmt.Printf("Synopsys-Operator does not exist in namespace %s", stopNamespace)
			return
		}
		cleanCommands := [...]string{
			"delete crd alerts.synopsys.com",
			"delete crd blackducks.synopsys.com",
			"delete crd hubs.synopsys.com",
			"delete crd opssights.synopsys.com",
			"delete clusterrolebinding synopsys-operator-admin",
			"delete clusterrole skyfire",
			"delete clusterrole pod-perceiver",
		}

		for cmd := range cleanCommands {
			fmt.Printf("Command: %s\n", cleanCommands[cmd])
			out, err = RunKubeCmd(strings.Split(cleanCommands[cmd], " ")...)
			if err != nil {
				fmt.Printf(" > %s", out)
			} else {
				fmt.Printf(" > %s", out)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
