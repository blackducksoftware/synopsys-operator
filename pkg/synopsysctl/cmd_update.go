/*
Copyright (C) 2018 Synopsys, Inc.

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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Resource Ctl for edit
var updateBlackduckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl
var updateAlertCtl ResourceCtl

type OperatorVersions struct {
	Blackduck string
	OpsSight  string
	Alert     string
}

// Lookup table for crd versions that are compatible with operator verions
var operatorVersionLookup = map[string]OperatorVersions{
	"2019.0.0": OperatorVersions{
		Blackduck: "v1",
		OpsSight:  "v1",
		Alert:     "v1",
	},
	"2019.1.1": OperatorVersions{
		Blackduck: "v1",
		OpsSight:  "v1",
		Alert:     "v1",
	},
}

// updateCmd provides functionality to update/upgrade features of
// Synopsys resources
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Synopsys Resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating a Synopsys-Operator Resource\n")
		return nil
	},
}

var updateOperatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Update the Synopsys-Operator",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, err := GetOperatorNamespace()
		if err != nil {
			log.Errorf("Error finding Synopsys-Operator: %s", err)
			return nil
		}
		log.Debugf("Updating the Synopsys-Operator: %s\n", namespace)
		// Get Spec of Synopsys-Operator

		// Check if Version has changed -> migration script

		// else just change spec fields

		return nil
	},
}

var updateBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Update an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating a Blackduck\n")
		// Read Commandline Parameters
		blackduckNamespace := args[0]

		// Get the Blackuck
		blackduck, err := getBlackduckFromCluster(blackduckNamespace)
		if err != nil {
			log.Errorf("Error getting Blackduck: %s", err)
			return nil
		}
		updateBlackduckCtl.SetSpec(blackduck.Spec)

		// Check if it can be updated
		canUpdate, err := updateBlackduckCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			log.Debugf("Updating...\n")
			// Make changes to Spec
			flagset := cmd.Flags()
			updateBlackduckCtl.SetChangedFlags(flagset)
			// Update in cluster
			newSpec := updateBlackduckCtl.GetSpec().(blackduckv1.BlackduckSpec)
			blackduck.Spec = newSpec
			updateBlackduckInCluster(blackduckNamespace, blackduck)
		}
		return nil
	},
}

var updateOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Update an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating an OpsSight\n")
		// Read Commandline Parameters
		opsSightNamespace := args[0]

		// Get the OpsSight
		opsSight, err := getOpsSightFromCluster(opsSightNamespace)
		if err != nil {
			log.Errorf("Error getting OpsSight: %s", err)
			return nil
		}
		updateOpsSightCtl.SetSpec(opsSight.Spec)

		// Check if it can be updated
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			log.Debugf("Updating...\n")
			// Make changes to Spec
			flagset := cmd.Flags()
			updateOpsSightCtl.SetChangedFlags(flagset)
			// Update in cluster
			newSpec := updateOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
			opsSight.Spec = newSpec
			updateOpsSightInCluster(opsSightNamespace, opsSight)
		}
		return nil
	},
}

var updateOpsSightImageCmd = &cobra.Command{
	Use:   "image NAMESPACE COMPONENT IMAGE",
	Short: "Update an image for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the Spec

		// Modify the Spec's Image

		// Update in the cluster

		// Restart the pod
		return nil
	},
}

// updateOpsSightAddRegistryCmd
var updateOpsSightExternalHostCmd = &cobra.Command{
	Use:   "externalHost NAMESPACE HOST",
	Short: "Update an external host for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// updateOpsSightAddRegistryCmd adds a registry to an OpsSight
var updateOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry NAMESPACE URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight's ImageFacade",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding Internal Registry to OpsSight\n")
		opsSightName := args[0]
		regURL := args[1]
		regUser := args[2]
		regPass := args[3]
		// Get OpsSight Spec
		ops, err := getOpsSightFromCluster(opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Internal Registry to Spec
		newReg := opssightv1.RegistryAuth{
			URL:      regURL,
			User:     regUser,
			Password: regPass,
		}
		ops.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ops.Spec.ScannerPod.ImageFacade.InternalRegistries, newReg)
		// Update OpsSight with Internal Registry
		err = updateOpsSightInCluster(opsSightName, ops)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

var updateAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Describe an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating an Alert\n")
		// Read Commandline Parameters
		alertNamespace := args[0]

		// Get the Alert
		alert, err := getAlertFromCluster(alertNamespace)
		if err != nil {
			log.Errorf("Error getting Alert: %s", err)
			return nil
		}
		updateAlertCtl.SetSpec(alert.Spec)

		// Check if it can be updated
		canUpdate, err := updateAlertCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			log.Debugf("Updating...\n")
			// Make changes to Spec
			flagset := cmd.Flags()
			updateAlertCtl.SetChangedFlags(flagset)
			// Update in cluster
			newSpec := updateAlertCtl.GetSpec().(alertv1.AlertSpec)
			alert.Spec = newSpec
			updateAlertInCluster(alertNamespace, alert)
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackduckCtl = blackduck.NewBlackduckCtl()
	updateOpsSightCtl = opssight.NewOpsSightCtl()
	updateAlertCtl = alert.NewAlertCtl()

	rootCmd.AddCommand(updateCmd)

	// Add Operator Commands
	updateCmd.AddCommand(updateOperatorCmd)

	// Add Bladuck Commands
	updateAlertCtl.AddSpecFlags(updateBlackduckCmd, false)
	updateCmd.AddCommand(updateBlackduckCmd)

	// Add OpsSight Commands
	updateOpsSightCtl.AddSpecFlags(updateOpsSightCmd, false)
	updateCmd.AddCommand(updateOpsSightCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightImageCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightExternalHostCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightAddRegistryCmd)

	// Add Alert Commands
	updateAlertCtl.AddSpecFlags(updateAlertCmd, false)
	updateCmd.AddCommand(updateAlertCmd)
}
