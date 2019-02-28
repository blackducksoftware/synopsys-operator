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
	"encoding/json"
	"fmt"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource Ctls for Create Command
var createBlackduckCtl ResourceCtl
var createOpsSightCtl ResourceCtl
var createAlertCtl ResourceCtl

// Flags for the Base Spec (template)
var baseBlackduckSpec = "persistentStorage"
var baseOpsSightSpec = "disabledBlackduck"
var baseAlertSpec = "spec1"

// Flags for using mock mode - don't deploy
var mockBlackduck bool
var mockOpsSight bool
var mockAlert bool

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
		// Display synopsysctl's Help instead of sending to oc/kubectl
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
	Use:   "blackduck NAMESPACE",
	Short: "Create an instance of a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		// Check the Arguments
		err := createBlackduckCtl.CheckSpecFlags()
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Set the Spec Type
		err = createBlackduckCtl.SwitchSpec(baseBlackduckSpec)
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

		// Update Spec with user's flags
		createBlackduckCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		blackduckSpec, _ := createBlackduckCtl.GetSpec().(blackduckv1.BlackduckSpec)
		blackduckSpec.Namespace = blackduckName

		// Create and Deploy Blackduck CRD
		blackduck := &blackduckv1.Blackduck{
			ObjectMeta: metav1.ObjectMeta{
				Name: blackduckName,
			},
			Spec: blackduckSpec,
		}
		if mockBlackduck {
			prettyPrint, _ := json.MarshalIndent(blackduck, "", "    ")
			fmt.Printf("%s\n", prettyPrint)
		} else {
			// Create namespace for the Blackduck
			err := DeployCRDNamespace(restconfig, blackduckName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create Blackduck with Client
			_, err = blackduckClient.SynopsysV1().Blackducks(blackduckName).Create(blackduck)
			if err != nil {
				log.Errorf("Error creating the Blackduck : %s", err)
				return nil
			}
		}
		return nil
	},
}

// createCmd represents the create command for OpsSight
var createOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Create an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		// Check the Arguments
		err := createOpsSightCtl.CheckSpecFlags()
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Set the Spec Type
		err = createOpsSightCtl.SwitchSpec(baseOpsSightSpec)
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

		// Update Spec with user's flags
		createOpsSightCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		opssightSpec, _ := createOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
		opssightSpec.Namespace = opsSightName

		// Create and Deploy OpsSight CRD
		opssight := &opssightv1.OpsSight{
			ObjectMeta: metav1.ObjectMeta{
				Name: opsSightName,
			},
			Spec: opssightSpec,
		}
		if mockOpsSight {
			prettyPrint, _ := json.MarshalIndent(opssight, "", "    ")
			fmt.Printf("%s\n", prettyPrint)
		} else {
			// Create namespace for the OpsSight
			err := DeployCRDNamespace(restconfig, opsSightName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create OpsSight with Client
			_, err = opssightClient.SynopsysV1().OpsSights(opsSightName).Create(opssight)
			if err != nil {
				log.Errorf("Error creating the OpsSight : %s", err)
				return nil
			}
		}
		return nil
	},
}

// createCmd represents the create command for Alert
var createAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Create an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument")
		}
		err := createAlertCtl.CheckSpecFlags()
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Check/Set the Spec Type
		err = createAlertCtl.SwitchSpec(baseAlertSpec)
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

		// Update Spec with user's flags
		createAlertCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		alertSpec, _ := createAlertCtl.GetSpec().(alertv1.AlertSpec)
		alertSpec.Namespace = alertName

		// Create and Deploy Alert CRD
		alert := &alertv1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name: alertName,
			},
			Spec: alertSpec,
		}
		if mockAlert {
			prettyPrint, _ := json.MarshalIndent(alert, "", "    ")
			fmt.Printf("%s\n", prettyPrint)
		} else {
			// Create namespace for the Alert
			err := DeployCRDNamespace(restconfig, alertName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create the Alert with Client
			_, err = alertClient.SynopsysV1().Alerts(alertName).Create(alert)
			if err != nil {
				log.Errorf("Error creating the Alert : %s", err)
				return nil
			}
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createBlackduckCtl = blackduck.NewBlackduckCtl()
	createOpsSightCtl = opssight.NewOpsSightCtl()
	createAlertCtl = alert.NewAlertCtl()

	createCmd.DisableFlagParsing = true // lets createCmd pass flags to kube/oc
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Command
	createBlackduckCmd.Flags().StringVar(&baseBlackduckSpec, "template", baseBlackduckSpec, "Base resource configuration to modify with flags")
	createBlackduckCmd.Flags().BoolVar(&mockBlackduck, "mock", false, "Prints resource spec instead of creating")
	createBlackduckCtl.AddSpecFlags(createBlackduckCmd)
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Command
	createOpsSightCmd.Flags().StringVar(&baseOpsSightSpec, "template", baseBlackduckSpec, "Base resource configuration to modify with flags")
	createOpsSightCmd.Flags().BoolVar(&mockOpsSight, "mock", false, "Prints resource spec instead of creating")
	createOpsSightCtl.AddSpecFlags(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Command
	createAlertCmd.Flags().StringVar(&baseAlertSpec, "template", baseBlackduckSpec, "Base resource configuration to modify with flags")
	createAlertCmd.Flags().BoolVar(&mockAlert, "mock", false, "Prints resource spec instead of creating")
	createAlertCtl.AddSpecFlags(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)
}
