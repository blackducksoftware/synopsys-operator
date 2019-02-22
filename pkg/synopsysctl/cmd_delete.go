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

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a Synopsys Resource from your cluster",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Deleting a Non-Synopsys Resource")
		kubeCmdArgs := append([]string{"delete"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			fmt.Printf("Error Deleting the Resource: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var deleteBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Delete a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Deleting a Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		blackduckClient, err := blackduckclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the Blackduck Clientset: %s\n", err)
			return
		}
		err = blackduckClient.SynopsysV1().Blackducks(blackduckNamespace).Delete(blackduckNamespace, &metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting the Blackduck: %s\n", err)
			return
		}
	},
}

var deleteOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Delete an OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Deleting an OpsSight\n")
		// Read Commandline Parameters
		opsSightNamespace := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		opssightClient, err := opssightclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the OpsSight Clientset: %s\n", err)
			return
		}
		err = opssightClient.SynopsysV1().OpsSights(opsSightNamespace).Delete(opsSightNamespace, &metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting the OpsSight: %s\n", err)
			return
		}
	},
}

var deleteAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Delete an Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Deleting an Alert\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		alertClient, err := alertclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the Alert Clientset: %s\n", err)
			return
		}
		err = alertClient.SynopsysV1().Alerts(alertNamespace).Delete(alertNamespace, &metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting the Alert: %s\n", err)
			return
		}
	},
}

func init() {
	deleteCmd.DisableFlagParsing = true
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.AddCommand(deleteBlackduckCmd)
	deleteCmd.AddCommand(deleteOpsSightCmd)
	deleteCmd.AddCommand(deleteAlertCmd)
}
