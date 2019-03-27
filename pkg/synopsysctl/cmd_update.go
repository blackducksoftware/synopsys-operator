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
	"strconv"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Resource Ctl for edit
var updateBlackduckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl
var updateAlertCtl ResourceCtl

// Update Defaults
var updateSynopsysOperatorImage = ""
var updatePrometheusImage = ""
var updateSecretType = ""
var updateSecretAdminPassword = ""
var updateSecretPostgresPassword = ""
var updateSecretUserPassword = ""
var updateSecretBlackduckPassword = ""

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
		namespace, err := soperator.GetOperatorNamespace(restconfig)
		if err != nil {
			log.Errorf("Error finding Synopsys-Operator: %s", err)
			return nil
		}
		log.Debugf("Updating the Synopsys-Operator in namespace %s\n", namespace)
		// Create new Synopsys-Operator Spec
		sOperatorSpecConfig, err := soperator.GetCurrentComponentsSpecConfig(kubeClient, namespace)
		if err != nil {
			log.Errorf("Error Updating Operator: %s", err)
			return nil
		}
		// Update Spec with changed values
		if cmd.Flag("synopsys-operator-image").Changed {
			log.Debugf("Updating SynopsysOperatorImage to %s", updateSynopsysOperatorImage)
			sOperatorSpecConfig.SynopsysOperatorImage = updateSynopsysOperatorImage
		}
		if cmd.Flag("secret-type").Changed {
			log.Debugf("Updating SecretType to %s", updateSecretType)
			updateSecretTypeConverted, err := operatorutil.SecretTypeNameToHorizon(updateSecretType)
			if err != nil {
				log.Errorf("Failed to convert SecretType: %s", err)
				return nil
			}
			sOperatorSpecConfig.SecretType = updateSecretTypeConverted
		}
		if cmd.Flag("admin-password").Changed {
			log.Debugf("Updating SecretAdminPassword")
			sOperatorSpecConfig.SecretAdminPassword = updateSecretAdminPassword
		}
		if cmd.Flag("postgres-password").Changed {
			log.Debugf("Updating SecretPostgresPassword")
			sOperatorSpecConfig.SecretPostgresPassword = updateSecretPostgresPassword
		}
		if cmd.Flag("user-password").Changed {
			log.Debugf("Updating SecretUserPassword")
			sOperatorSpecConfig.SecretUserPassword = updateSecretUserPassword
		}
		if cmd.Flag("blackduck-password").Changed {
			log.Debugf("Updating SecretBlackduckPassword")
			sOperatorSpecConfig.SecretBlackduckPassword = updateSecretBlackduckPassword
		}
		err = soperator.UpdateSynopsysOperator(restconfig, kubeClient, namespace, sOperatorSpecConfig, blackduckClient, opssightClient, alertClient)
		if err != nil {
			log.Errorf("Failed to Updated Synopsys-Operator: %s", err)
		}

		log.Debugf("Updating Prometheus in namespace %s\n", namespace)
		prometheusSpecConfig, err := soperator.GetCurrentComponentsSpecConfigPrometheus(kubeClient, namespace)
		if err != nil {
			log.Errorf("Error Updating the Operator: %s", err)
		}
		if cmd.Flag("prometheus-image").Changed {
			log.Debugf("Updating PrometheusImage to %s", updatePrometheusImage)
			prometheusSpecConfig.PrometheusImage = updatePrometheusImage
		}
		err = soperator.UpdatePrometheus(restconfig, kubeClient, namespace, prometheusSpecConfig)
		if err != nil {
			log.Errorf("Failed to Updated Prometheus: %s", err)
		}
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
		currBlackduck, err := operatorutil.GetHub(blackduckClient, blackduckNamespace, blackduckNamespace)
		if err != nil {
			log.Errorf("Error getting Blackduck: %s", err)
			return nil
		}
		updateBlackduckCtl.SetSpec(currBlackduck.Spec)

		// Check if it can be updated
		canUpdate, err := updateBlackduckCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateBlackduckCtl.SetChangedFlags(flagset)
			newSpec := updateBlackduckCtl.GetSpec().(blackduckv1.BlackduckSpec)
			// Create new Blackduck CRD
			newBlackduck := *currBlackduck //make copy
			newBlackduck.Spec = newSpec
			// Update Blackduck
			_, err = operatorutil.UpdateBlackduck(blackduckClient, newBlackduck.Spec.Namespace, &newBlackduck)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
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

		// Get the current OpsSight
		currOpsSight, err := getOpsSightFromCluster(opsSightNamespace)
		if err != nil {
			log.Errorf("Error getting OpsSight: %s", err)
			return nil
		}
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)

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
			newSpec := updateOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
			// Create new OpsSight CRD
			newOpsSight := *currOpsSight //make copy
			newOpsSight.Spec = newSpec
			// Update OpsSight
			_, err = operatorutil.UpdateOpsSight(opssightClient, newOpsSight.Spec.Namespace, &newOpsSight)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
		}
		return nil
	},
}

var updateOpsSightImageCmd = &cobra.Command{
	Use:   "image NAMESPACE COMPONENT IMAGE",
	Short: "Update an image for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("This command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating an Image of OpsSight\n")
		opsSightName := args[0]
		componentName := args[1]
		componentImage := args[2]
		// Get OpsSight Spec
		opsSightCRD, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Update the Spec with new Image
		switch componentName {
		case "Perceptor":
			opsSightCRD.Spec.Perceptor.Image = componentImage
		case "Scanner":
			opsSightCRD.Spec.ScannerPod.Scanner.Image = componentImage
		case "ImageFacade":
			opsSightCRD.Spec.ScannerPod.ImageFacade.Image = componentImage
		case "ImagePerceiver":
			opsSightCRD.Spec.Perceiver.ImagePerceiver.Image = componentImage
		case "PodPerceiver":
			opsSightCRD.Spec.Perceiver.PodPerceiver.Image = componentImage
		case "Skyfire":
			opsSightCRD.Spec.Skyfire.Image = componentImage
		case "Prometheus":
			opsSightCRD.Spec.Prometheus.Image = componentImage
		default:
			log.Errorf("Invalid COMPONENT")
			return fmt.Errorf("Invalid Component Name")
		}
		// Update OpsSight with New Image
		_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, opsSightCRD)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// updateOpsSightAddRegistryCmd
var updateOpsSightExternalHostCmd = &cobra.Command{
	Use:   "externalHost NAMESPACE SCHEME DOMAIN PORT USER PASSWORD SCANLIMIT",
	Short: "Update an external host for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 7 {
			return fmt.Errorf("This command takes 7 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding External Host to OpsSight\n")
		opsSightName := args[0]
		hostScheme := args[1]
		hostDomain := args[2]
		hostPort, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			log.Errorf("Invalid Port Number: %s", err)
		}
		hostUser := args[4]
		hostPassword := args[5]
		hostScanLimit, err := strconv.ParseInt(args[6], 0, 64)
		if err != nil {
			log.Errorf("Invalid Concurrent Scan Limit: %s", err)
		}
		// Get OpsSight Spec
		opsSightCRD, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add External Host to Spec
		newHost := opssightv1.Host{
			Scheme:              hostScheme,
			Domain:              hostDomain,
			Port:                int(hostPort),
			User:                hostUser,
			Password:            hostPassword,
			ConcurrentScanLimit: int(hostScanLimit),
		}
		opsSightCRD.Spec.Blackduck.ExternalHosts = append(opsSightCRD.Spec.Blackduck.ExternalHosts, &newHost)
		// Update OpsSight with External Host
		_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, opsSightCRD)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// updateOpsSightAddRegistryCmd adds a registry to an OpsSight
var updateOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "registry NAMESPACE URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight's ImageFacade",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
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
		opsSightCRD, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
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
		opsSightCRD.Spec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightCRD.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
		// Update OpsSight with Internal Registry
		_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, opsSightCRD)
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
		currAlert, err := getAlertFromCluster(alertNamespace)
		if err != nil {
			log.Errorf("Error getting Alert: %s", err)
			return nil
		}
		updateAlertCtl.SetSpec(currAlert.Spec)

		// Check if it can be updated
		canUpdate, err := updateAlertCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateAlertCtl.SetChangedFlags(flagset)
			newSpec := updateAlertCtl.GetSpec().(alertv1.AlertSpec)
			// Create new Alert CRD
			newAlert := *currAlert //make copy
			newAlert.Spec = newSpec
			// Update Alert
			_, err = operatorutil.UpdateAlert(alertClient, newAlert.Spec.Namespace, &newAlert)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
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
	updateOperatorCmd.Flags().StringVarP(&updateSynopsysOperatorImage, "synopsys-operator-image", "i", updateSynopsysOperatorImage, "synopsys operator image URL")
	updateOperatorCmd.Flags().StringVarP(&updatePrometheusImage, "prometheus-image", "p", updatePrometheusImage, "prometheus image URL")
	updateOperatorCmd.Flags().StringVar(&updateSecretType, "secret-type", updateSecretType, "type of kubernetes secret for postgres and blackduck")
	updateOperatorCmd.Flags().StringVar(&updateSecretAdminPassword, "admin-password", updateSecretAdminPassword, "postgres admin password")
	updateOperatorCmd.Flags().StringVar(&updateSecretPostgresPassword, "postgres-password", updateSecretPostgresPassword, "postgres password")
	updateOperatorCmd.Flags().StringVar(&updateSecretUserPassword, "user-password", updateSecretUserPassword, "postgres user password")
	updateOperatorCmd.Flags().StringVar(&updateSecretBlackduckPassword, "blackduck-password", updateSecretBlackduckPassword, "blackduck password for 'sysadmin' account")
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
