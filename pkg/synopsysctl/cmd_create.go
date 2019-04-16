/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package synopsysctl

import (
	"fmt"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource Ctls for Create Command
var createBlackduckCtl ResourceCtl
var createOpsSightCtl ResourceCtl
var createAlertCtl ResourceCtl

// Flags for the Base Spec (template)
var baseBlackduckSpec = "persistentStorageLatest"
var baseOpsSightSpec = "disabledBlackDuck"
var baseAlertSpec = "default"

// Flags for using mock mode - don't deploy
var mockBlackduck bool
var mockOpsSight bool
var mockAlert bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys Resource in your cluster",
	//(PassCmd)Args: func(cmd *cobra.Command, args []string) error {
	//(PassCmd)	numArgs := 1
	//(PassCmd)	if len(args) < numArgs {
	//(PassCmd)		return fmt.Errorf("Not enough arguments")
	//(PassCmd)	}
	//(PassCmd)	return nil
	//(PassCmd)},
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	// Display synopsysctl's Help instead of sending to oc/kubectl
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Creating a Non-Synopsys Resource\n")
		//(PassCmd) kubeCmdArgs := append([]string{"create"}, args...)
		//(PassCmd) out, err := util.RunKubeCmd(restconfig, kube, openshift, kubeCmdArgs...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	return fmt.Errorf("Error Creating the Resource with KubeCmd: %s", err)
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
	},
}

// createCmd represents the create command for Blackduck
var createBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Create an instance of a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
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
		blackduckNamespace := args[0]

		// Update Spec with user's flags
		createBlackduckCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		blackduckSpec, _ := createBlackduckCtl.GetSpec().(blackduckv1.BlackduckSpec)
		blackduckSpec.Namespace = blackduckNamespace

		// Create and Deploy Blackduck CRD
		blackduck := &blackduckv1.Blackduck{
			ObjectMeta: metav1.ObjectMeta{
				Name:      blackduckNamespace,
				Namespace: blackduckNamespace,
			},
			Spec: blackduckSpec,
		}
		blackduck.Kind = "Blackduck"
		blackduck.APIVersion = "synopsys.com/v1"
		if mockBlackduck {
			util.PrettyPrint(blackduck)
		} else {
			// Create namespace for the Blackduck
			err := DeployCRDNamespace(restconfig, blackduckNamespace)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create Blackduck with Client
			_, err = blackduckClient.SynopsysV1().Blackducks(blackduckNamespace).Create(blackduck)
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
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
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
		opsSightNamespace := args[0]

		// Update Spec with user's flags
		createOpsSightCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		opssightSpec, _ := createOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
		opssightSpec.Namespace = opsSightNamespace

		// Create and Deploy OpsSight CRD
		opssight := &opssightv1.OpsSight{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opsSightNamespace,
				Namespace: opsSightNamespace,
			},
			Spec: opssightSpec,
		}
		opssight.Kind = "OpsSight"
		opssight.APIVersion = "synopsys.com/v1"
		if mockOpsSight {
			util.PrettyPrint(opssight)
		} else {
			// Create namespace for the OpsSight
			err := DeployCRDNamespace(restconfig, opsSightNamespace)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create OpsSight with Client
			_, err = opssightClient.SynopsysV1().OpsSights(opsSightNamespace).Create(opssight)
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
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
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
		alertNamespace := args[0]

		// Update Spec with user's flags
		createAlertCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		alertSpec, _ := createAlertCtl.GetSpec().(alertv1.AlertSpec)
		alertSpec.Namespace = alertNamespace

		// Create and Deploy Alert CRD
		alert := &alertv1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name:      alertNamespace,
				Namespace: alertNamespace,
			},
			Spec: alertSpec,
		}
		alert.Kind = "Alert"
		alert.APIVersion = "synopsys.com/v1"
		if mockAlert {
			util.PrettyPrint(alert)
		} else {
			// Create namespace for the Alert
			err := DeployCRDNamespace(restconfig, alertNamespace)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Create the Alert with Client
			_, err = alertClient.SynopsysV1().Alerts(alertNamespace).Create(alert)
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

	//(PassCmd) createCmd.DisableFlagParsing = true // lets createCmd pass flags to kube/oc
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Command
	createBlackduckCmd.Flags().StringVar(&baseBlackduckSpec, "template", baseBlackduckSpec, "Base resource configuration to modify with flags [empty/template/persistentStorageLatest/persistentStorageV1/externalPersistentStorageLatest/externalPersistentStorageV1/bdba/ephemeral/ephemeralCustomAuthCA/externalDB/IPV6Disabled]")
	createBlackduckCmd.Flags().BoolVar(&mockBlackduck, "mock", false, "Prints resource spec instead of creating")
	createBlackduckCtl.AddSpecFlags(createBlackduckCmd, true)
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Command
	createOpsSightCmd.Flags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags [empty/template/default/disabledBlackDuck]")
	createOpsSightCmd.Flags().BoolVar(&mockOpsSight, "mock", false, "Prints resource spec instead of creating")
	createOpsSightCtl.AddSpecFlags(createOpsSightCmd, true)
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Command
	createAlertCmd.Flags().StringVar(&baseAlertSpec, "template", baseAlertSpec, "Base resource configuration to modify with flags [empty/template/default]")
	createAlertCmd.Flags().BoolVar(&mockAlert, "mock", false, "Prints resource spec instead of creating")
	createAlertCtl.AddSpecFlags(createAlertCmd, true)
	createCmd.AddCommand(createAlertCmd)
}
