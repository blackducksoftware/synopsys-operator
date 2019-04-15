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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource Ctl for edit
var updateBlackduckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl
var updateAlertCtl ResourceCtl

// Update Defaults
var updateSynopsysOperatorImage = ""
var updatePrometheusImage = ""
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
		return fmt.Errorf("Must specify a sub-command")
	},
}

// updateOperatorCmd lets the user update the Synopsys-Operator
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

		log.Debugf("Updating the Synopsys-Operator in namespace %s", namespace)
		// Create new Synopsys-Operator SpecConfig
		sOperatorSpecConfig, err := soperator.GetSpecConfigForCurrentComponents(restconfig, kubeClient, namespace)
		if err != nil {
			log.Errorf("Error Updating Operator: %s", err)
			return nil
		}
		// Update Spec with changed values
		if cmd.Flag("synopsys-operator-image").Changed {
			log.Debugf("Updating SynopsysOperatorImage to %s", updateSynopsysOperatorImage)
			sOperatorSpecConfig.Image = updateSynopsysOperatorImage
		}
		if cmd.Flag("admin-password").Changed {
			log.Debugf("Updating SecretAdminPassword")
			sOperatorSpecConfig.AdminPassword = updateSecretAdminPassword
		}
		if cmd.Flag("postgres-password").Changed {
			log.Debugf("Updating SecretPostgresPassword")
			sOperatorSpecConfig.PostgresPassword = updateSecretPostgresPassword
		}
		if cmd.Flag("user-password").Changed {
			log.Debugf("Updating SecretUserPassword")
			sOperatorSpecConfig.UserPassword = updateSecretUserPassword
		}
		if cmd.Flag("blackduck-password").Changed {
			log.Debugf("Updating SecretBlackduckPassword")
			sOperatorSpecConfig.BlackduckPassword = updateSecretBlackduckPassword
		}
		err = sOperatorSpecConfig.UpdateSynopsysOperator(restconfig, kubeClient, namespace, blackduckClient, opssightClient, alertClient)
		if err != nil {
			log.Errorf("Failed to update the Synopsys-Operator: %s", err)
		}

		log.Debugf("Updating Prometheus in namespace %s", namespace)
		// Create new Prometheus SpecConfig
		prometheusSpecConfig, err := soperator.GetSpecConfigForCurrentPrometheusComponents(restconfig, kubeClient, namespace)
		if err != nil {
			log.Errorf("Error Updating the Operator: %s", err)
		}
		if cmd.Flag("prometheus-image").Changed {
			log.Debugf("Updating PrometheusImage to %s", updatePrometheusImage)
			prometheusSpecConfig.Image = updatePrometheusImage
		}
		err = prometheusSpecConfig.UpdatePrometheus()
		if err != nil {
			log.Errorf("Failed to update Prometheus: %s", err)
		}
		return nil
	},
}

// updateBlackduckCmd lets the user update a BlackDuck instance
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
		blackduckNamespace := args[0]

		// Get the Blackuck
		currBlackduck, err := operatorutil.GetHub(blackduckClient, blackduckNamespace, blackduckNamespace)
		if err != nil {
			log.Errorf("Error getting Blackduck: %s", err)
			return nil
		}
		// Check if it can be updated
		updateBlackduckCtl.SetSpec(currBlackduck.Spec)
		canUpdate, err := updateBlackduckCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateBlackduckCtl.SetChangedFlags(flagset)
			newSpec := updateBlackduckCtl.GetSpec().(blackduckapi.BlackduckSpec)
			// Create new Blackduck CRD
			newBlackduck := *currBlackduck //make copy
			newBlackduck.Spec = newSpec
			// Update Blackduck
			_, err = operatorutil.UpdateBlackduck(blackduckClient, newBlackduck.Spec.Namespace, &newBlackduck)
			if err != nil {
				log.Errorf("Error updating the BlackDuck: %s", err)
				return nil
			}
		}
		return nil
	},
}

// updateBlackduckRootKeyCmd create new Black Duck root key for source code upload in the cluster
var updateBlackduckRootKeyCmd = &cobra.Command{
	Use:   "rootKey BLACK_DUCK_NAME NEW_SEAL_KEY MASTER_KEY_FILE_PATH",
	Short: "Update the root key of Black Duck for source code upload",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("Black Duck name, new seal key or file path to retrieve the master key is missing")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Updating Blackduck Root Key\n")
		namespace := args[0]
		newSealKey := args[1]
		filePath := args[2]

		_, err := operatorutil.GetHub(blackduckClient, metav1.NamespaceDefault, namespace)
		if err != nil {
			log.Errorf("unable to find Black Duck %s instance due to %+v", namespace, err)
			return nil
		}
		operatorNamespace, err := soperator.GetOperatorNamespace(restconfig)
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator instance due to %+v", err)
			return nil
		}

		fileName := filepath.Join(filePath, fmt.Sprintf("%s.key", namespace))
		masterKey, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("error reading the master key from %s because %+v", fileName, err)
			return nil
		}

		// Filter the upload cache pod to get the root key using the seal key
		uploadCachePod, err := operatorutil.FilterPodByNamePrefixInNamespace(kubeClient, namespace, "uploadcache")
		if err != nil {
			log.Errorf("unable to filter the upload cache pod of %s due to %+v", namespace, err)
			return nil
		}

		// Create the exec into kubernetes pod request
		req := operatorutil.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
		_, err = operatorutil.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://uploadcache:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
		if err != nil {
			log.Errorf("unable to exec into upload cache pod in %s because %+v", namespace, err)
			return nil
		}

		secret, err := operatorutil.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
			return nil
		}
		secret.Data["SEAL_KEY"] = []byte(newSealKey)

		err = operatorutil.UpdateSecret(kubeClient, operatorNamespace, secret)
		if err != nil {
			log.Errorf("unable to update the Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
			return nil
		}

		return nil
	},
}

// updateOpsSightCmd lets the user update an OpsSight instance
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
		opsSightNamespace := args[0]

		// Get the current OpsSight
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("Error getting OpsSight: %s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateOpsSightCtl.SetChangedFlags(flagset)
			newSpec := updateOpsSightCtl.GetSpec().(opssightapi.OpsSightSpec)
			// Create new OpsSight CRD
			newOpsSight := *currOpsSight //make copy
			newOpsSight.Spec = newSpec
			// Update OpsSight
			_, err = operatorutil.UpdateOpsSight(opssightClient, newOpsSight.Spec.Namespace, &newOpsSight)
			if err != nil {
				log.Errorf("Error updating the OpsSight: %s", err)
				return nil
			}
		}
		return nil
	},
}

// updateOpsSightImageCmd lets the user update an image in an OpsSight instance
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
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update OpsSight: %s\n", err)
			return nil
		}
		if canUpdate {
			// Update the Spec with new Image
			switch componentName {
			case "Perceptor":
				currOpsSight.Spec.Perceptor.Image = componentImage
			case "Scanner":
				currOpsSight.Spec.ScannerPod.Scanner.Image = componentImage
			case "ImageFacade":
				currOpsSight.Spec.ScannerPod.ImageFacade.Image = componentImage
			case "ImagePerceiver":
				currOpsSight.Spec.Perceiver.ImagePerceiver.Image = componentImage
			case "PodPerceiver":
				currOpsSight.Spec.Perceiver.PodPerceiver.Image = componentImage
			case "Skyfire":
				currOpsSight.Spec.Skyfire.Image = componentImage
			case "Prometheus":
				currOpsSight.Spec.Prometheus.Image = componentImage
			default:
				log.Errorf("%s is not a valid COMPONENT\n", componentName)
				log.Errorf("Valid Components: Perceptor, Scanner, ImageFacade, ImagePerceiver, PodPerceiver, Skyfire, Prometheus\n")
				return fmt.Errorf("Invalid Component Name")
			}
			// Update OpsSight with New Image
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("Error updating the OpsSight: %s", err)
				return nil
			}
		}
		return nil
	},
}

// updateOpsSightExternalHostCmd lets the user update an OpsSight with an External Host
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
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("Error getting the OpsSight: %s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update OpsSight: %s\n", err)
			return nil
		}
		if canUpdate {
			// Add External Host to Spec
			newHost := opssightapi.Host{
				Scheme:              hostScheme,
				Domain:              hostDomain,
				Port:                int(hostPort),
				User:                hostUser,
				Password:            hostPassword,
				ConcurrentScanLimit: int(hostScanLimit),
			}
			currOpsSight.Spec.Blackduck.ExternalHosts = append(currOpsSight.Spec.Blackduck.ExternalHosts, &newHost)
			// Update OpsSight with External Host
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("Error updating the OpsSight: %s", err)
				return nil
			}
		}
		return nil
	},
}

// updateOpsSightAddRegistryCmd lets the user update and OpsSight by
// adding a registry for the ImageFacade
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
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("Error adding Internal Registry while getting OpsSight: %s\n", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update OpsSight: %s\n", err)
			return nil
		}
		if canUpdate {
			// Add Internal Registry to Spec
			newReg := opssightapi.RegistryAuth{
				URL:      regURL,
				User:     regUser,
				Password: regPass,
			}
			currOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries = append(currOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
			// Update OpsSight with Internal Registry
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("Error adding Internal Registry with updating OpsSight: %s\n", err)
				return nil
			}
		}
		return nil
	},
}

// updateAlertCmd lets the user update an Alert Instance
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
		alertNamespace := args[0]

		// Get the Alert
		currAlert, err := operatorutil.GetAlert(alertClient, alertNamespace, alertNamespace)
		if err != nil {
			log.Errorf("Error Updaing Alert while getting the Alert: %s", err)
			return nil
		}
		updateAlertCtl.SetSpec(currAlert.Spec)

		// Check if it can be updated
		canUpdate, err := updateAlertCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update Alert: %s\n", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateAlertCtl.SetChangedFlags(flagset)
			newSpec := updateAlertCtl.GetSpec().(alertapi.AlertSpec)
			// Create new Alert CRD
			newAlert := *currAlert //make copy
			newAlert.Spec = newSpec
			// Update Alert
			_, err = operatorutil.UpdateAlert(alertClient, newAlert.Spec.Namespace, &newAlert)
			if err != nil {
				log.Errorf("Error Updating the Alert: %s\n", err)
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
	updateOperatorCmd.Flags().StringVar(&updateSecretAdminPassword, "admin-password", updateSecretAdminPassword, "postgres admin password")
	updateOperatorCmd.Flags().StringVar(&updateSecretPostgresPassword, "postgres-password", updateSecretPostgresPassword, "postgres password")
	updateOperatorCmd.Flags().StringVar(&updateSecretUserPassword, "user-password", updateSecretUserPassword, "postgres user password")
	updateOperatorCmd.Flags().StringVar(&updateSecretBlackduckPassword, "blackduck-password", updateSecretBlackduckPassword, "blackduck password for 'sysadmin' account")
	updateCmd.AddCommand(updateOperatorCmd)

	// Add Bladuck Commands
	updateBlackduckCtl.AddSpecFlags(updateBlackduckCmd, false)
	updateCmd.AddCommand(updateBlackduckCmd)
	updateBlackduckCmd.AddCommand(updateBlackduckRootKeyCmd)

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
