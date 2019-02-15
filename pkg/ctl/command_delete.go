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

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a Synopsys Resource from your cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")

	},
}

var deleteBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Delete a Blackduck",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		blackduckClient, err := blackduckclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the Blackduck Clientset: %s\n", err)
			return
		}
		blackduckClient.SynopsysV1().Blackducks(namespace).Delete(namespace, &metav1.DeleteOptions{})

	},
}

var deleteOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Delete an OpsSight",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		opssightClient, err := opssightclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the OpsSight Clientset: %s\n", err)
			return
		}
		opssightClient.SynopsysV1().OpsSights(namespace).Delete(namespace, &metav1.DeleteOptions{})
	},
}

var deleteAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Delete an Alert",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()
		alertClient, err := alertclientset.NewForConfig(restconfig)
		if err != nil {
			fmt.Printf("Error creating the Alert Clientset: %s\n", err)
			return
		}
		alertClient.SynopsysV1().Alerts(namespace).Delete(namespace, &metav1.DeleteOptions{})
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.AddCommand(deleteBlackduckCmd)
	deleteCmd.AddCommand(deleteOpsSightCmd)
	deleteCmd.AddCommand(deleteAlertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
