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

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource Ctl for create
var createBlackduckCtl = NewBlackduckCtl()
var createOpsSightCtl = NewOpsSightCtl()
var createAlertCtl = NewAlertCtl()

// Create Flags
var createBlackduckSpecType = "persistentStorage"
var createOpsSightSpecType = "disabledBlackduck"
var createAlertSpecType = "spec1"

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys Resource in your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		numArgs := 1
		if len(args) < numArgs {
			return fmt.Errorf("Not enough arguments")
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Creating a Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"create"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			return fmt.Errorf("Error Creating the Resource with KubeCmd: %s", err)
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// createCmd represents the create command for Blackduck
var createBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAME",
	Short: "Create an instance of a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		// Check the Spec Type
		err := createBlackduckCtl.SwitchSpec(createBlackduckSpecType)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Creating a Blackduck\n")
		// Read Commandline Parameters
		blackduckName := "blackduck"
		if len(args) == 1 {
			blackduckName = args[0]
		}

		// Create namespace for the Blackduck
		err := DeployCRDNamespace(restconfig, blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}

		// Read Flags Into Default Blackduck Spec
		flagset := cmd.Flags()
		flagset.VisitAll(createBlackduckCtl.SetFlags)

		// Set Namespace in Spec
		blackduckSpec := createBlackduckCtl.GetSpec()
		blackduckSpec.Namespace = blackduckName

		// Create and Deploy Blackduck CRD
		blackduck := &blackduckv1.Blackduck{
			ObjectMeta: metav1.ObjectMeta{
				Name: blackduckName,
			},
			Spec: blackduckSpec,
		}
		log.Debugf("%+v\n", blackduck)
		_, err = blackduckClient.SynopsysV1().Blackducks(blackduckName).Create(blackduck)
		if err != nil {
			log.Errorf("Error creating the Blackduck : %s", err)
			return nil
		}
		return nil
	},
}

// createCmd represents the create command for OpsSight
var createOpsSightCmd = &cobra.Command{
	Use:   "opssight NAME",
	Short: "Create an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		// Check the Spec Type
		err := createOpsSightCtl.SwitchSpec(createOpsSightSpecType)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Creating an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := "opssight"
		if len(args) == 1 {
			opsSightName = args[0]
		}

		// Create namespace for the OpsSight
		err := DeployCRDNamespace(restconfig, opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}

		// Read Flags Into Default OpsSight Spec
		flagset := cmd.Flags()
		flagset.VisitAll(createOpsSightCtl.SetFlags)

		// Set Namespace in Spec
		opssightSpec := createOpsSightCtl.GetSpec()
		opssightSpec.Namespace = opsSightName

		// Create and Deploy OpsSight CRD
		opssight := &opssightv1.OpsSight{
			ObjectMeta: metav1.ObjectMeta{
				Name: opsSightName,
			},
			Spec: opssightSpec,
		}
		log.Debugf("%+v\n", opssight)
		_, err = opssightClient.SynopsysV1().OpsSights(opsSightName).Create(opssight)
		if err != nil {
			log.Errorf("Error creating the OpsSight : %s", err)
			return nil
		}
		return nil
	},
}

// createCmd represents the create command for Alert
var createAlertCmd = &cobra.Command{
	Use:   "alert NAME",
	Short: "Create an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		// Check the Spec Type
		err := createAlertCtl.SwitchSpec(createAlertSpecType)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Creating an Alert\n")
		// Read Commandline Parameters
		alertName := "alert"
		if len(args) == 1 {
			alertName = args[0]
		}

		// Create namespace for the Alert
		err := DeployCRDNamespace(restconfig, alertName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}

		// Read Flags Into Default Alert Spec
		flagset := cmd.Flags()
		flagset.VisitAll(createAlertCtl.SetFlags)

		// Set Namespace in Spec
		alertSpec := createAlertCtl.GetSpec()
		alertSpec.Namespace = alertName

		// Create and Deploy Alert CRD
		alert := &alertv1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name: alertName,
			},
			Spec: alertSpec,
		}
		log.Debugf("%+v\n", alert)
		_, err = alertClient.SynopsysV1().Alerts(alertName).Create(alert)
		if err != nil {
			log.Errorf("Error creating the Alert : %s", err)
			return nil
		}
		return nil
	},
}

func init() {
	createCmd.DisableFlagParsing = true // lets createCmd pass flags to kube/oc
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Command Flags
	createBlackduckCmd.Flags().StringVar(&createBlackduckSpecType, "spec", createBlackduckSpecType, "TODO")
	createBlackduckCtl.AddSpecFlags(createBlackduckCmd)
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Command Flags
	createOpsSightCmd.Flags().StringVar(&createOpsSightSpecType, "spec", createOpsSightSpecType, "TODO")
	createOpsSightCtl.AddSpecFlags(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Command Flags
	createAlertCmd.Flags().StringVar(&createAlertSpecType, "spec", createAlertSpecType, "TODO")
	createAlertCtl.AddSpecFlags(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)
}
