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
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
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
		currPod, err := operatorutil.GetPod(kubeClient, namespace, "synopsys-operator")
		if err != nil {
			log.Errorf("Failed to get Synopsys-Operator Pod: %s", err)
			return nil
		}
		var currImage string
		for _, container := range currPod.Spec.Containers {
			if container.Name == "synopsys-operator" {
				continue
			}
			currImage = container.Image
		}
		imageChanged := false
		if currImage != deploySynopsysOperatorImage {
			imageChanged = true
		}
		currOperatorVersion := "2019.0.0"
		newOperatorVersion := "2019.0.0"
		if imageChanged {
			currCrdNames := soperator.OperatorVersionLookup[currOperatorVersion]
			newCrdNames := soperator.OperatorVersionLookup[newOperatorVersion]
			// Get local copies of CRD specs of all instances
			currBlackduckCRDs, err := operatorutil.GetBlackducks(blackduckClient)
			currOpsSightCRDs, err := operatorutil.GetOpsSights(opssightClient)
			currAlertCRDs, err := operatorutil.GetAlerts(alertClient)
			// Change CRD specs to have new versions
			newBlackduckCRDs, err := setBlackduckCrdVersions(currBlackduckCRDs.Items, newCrdNames.Blackduck.Version)
			newOpsSightCRDs, err := setOpsSightCrdVersions(currOpsSightCRDs.Items, newCrdNames.OpsSight.Version)
			newAlertCRDs, err := setAlertCrdVersions(currAlertCRDs.Items, newCrdNames.Alert.Version)
			// Delete the CRD definitions from the cluster
			for _, crd := range soperator.GetCrdDataList(currOperatorVersion) {
				RunKubeCmd("delete", "crd", crd.Name)
			}
			// Update the Synopsys-Operator's Kubernetes Components (TODO this will deploy new crds)
			updateSynopsysOperator(namespace)
			updatePrometheus(namespace)
			// Update the resources in the cluster with the new versions
			for _, crd := range newBlackduckCRDs {
				if crd.TypeMeta.APIVersion != currCrdNames.Blackduck.Version {
					_, err = operatorutil.UpdateBlackduck(blackduckClient, crd.Spec.Namespace, &crd)
				}
			}
			for _, crd := range newOpsSightCRDs {
				if crd.TypeMeta.APIVersion != currCrdNames.OpsSight.Version {
					_, err = operatorutil.UpdateOpsSight(opssightClient, crd.Spec.Namespace, &crd)
				}
			}
			for _, crd := range newAlertCRDs {
				if crd.TypeMeta.APIVersion != currCrdNames.Alert.Version {
					_, err = operatorutil.UpdateAlert(alertClient, crd.Spec.Namespace, &crd)
				}
			}
			if err != nil {
				log.Errorf("An Error Occurred")
				return nil
			}
		} else {
			updateSynopsysOperator(namespace)
			updatePrometheus(namespace)
		}

		return nil
	},
}

func setBlackduckCrdVersions(blackduckList []blackduckv1.Blackduck, version string) ([]blackduckv1.Blackduck, error) {
	for _, crd := range blackduckList {
		crd.TypeMeta.APIVersion = version
	}
	return blackduckList, nil
}

func setOpsSightCrdVersions(opsSightList []opssightv1.OpsSight, version string) ([]opssightv1.OpsSight, error) {
	for _, crd := range opsSightList {
		crd.TypeMeta.APIVersion = version
	}
	return opsSightList, nil
}

func setAlertCrdVersions(alertList []alertv1.Alert, version string) ([]alertv1.Alert, error) {
	for _, crd := range alertList {
		crd.TypeMeta.APIVersion = version
	}
	return alertList, nil
}

func updateSynopsysOperator(namespace string) error {
	// Get Components of Current Synopsys-Operator
	currPod, err := operatorutil.GetPod(kubeClient, namespace, "synopsys-operator")
	var currImage string
	var currRegKey string
	for _, container := range currPod.Spec.Containers {
		if container.Name == "synopsys-operator" {
			continue
		}
		currImage = container.Image
		for _, env := range container.Env {
			if env.Name != "REGISTRATION_KEY" {
				continue
			}
			currRegKey = container.Env[0].Value
		}
	}
	currSecret, err := operatorutil.GetSecret(kubeClient, namespace, "blackduck-secret")
	currSecretType, err := kubeSecretTypeToHorizon(currSecret.Type)
	currSOperatorSpec := soperator.SOperatorSpecConfig{
		Namespace:                namespace,
		SynopsysOperatorImage:    currImage,
		BlackduckRegistrationKey: currRegKey,
		SecretType:               currSecretType,
		SecretAdminPassword:      deploySecretAdminPassword,
		SecretPostgresPassword:   deploySecretPostgresPassword,
		SecretUserPassword:       deploySecretUserPassword,
		SecretBlackduckPassword:  deploySecretBlackduckPassword,
	}
	currSOperatorComponents, err := currSOperatorSpec.GetComponents()
	fmt.Printf("%+v\n", currSOperatorComponents)

	// Get Components of New Synopsys-Operator
	newSOperatorSpec := soperator.SOperatorSpecConfig{
		Namespace:                deployNamespace,
		SynopsysOperatorImage:    deploySynopsysOperatorImage,
		BlackduckRegistrationKey: deployBlackduckRegistrationKey,
		SecretType:               secretType,
		SecretAdminPassword:      deploySecretAdminPassword,
		SecretPostgresPassword:   deploySecretPostgresPassword,
		SecretUserPassword:       deploySecretUserPassword,
		SecretBlackduckPassword:  deploySecretBlackduckPassword,
	}
	newSOperatorComponents, err := newSOperatorSpec.GetComponents()
	fmt.Printf("%+v\n", newSOperatorComponents)

	// Update S-O ConfigMap if necessary
	isConfigMapUpdated, err := crdupdater.UpdateConfigMap(kubeClient, deployNamespace, "synopsys-operator", newSOperatorComponents.ConfigMaps[0])

	// Update S-O Secret if necessary
	isSecretUpdated, err := crdupdater.UpdateSecret(kubeClient, deployNamespace, "blackduck-secret", newSOperatorComponents.Secrets[0])

	operatorUpdater := crdupdater.NewUpdater()

	// Update S-O ReplicationController if necessary
	replicationControllerUpdater, err := crdupdater.NewReplicationController(restconfig, kubeClient, newSOperatorComponents.ReplicationControllers, namespace, "app=opssight", isConfigMapUpdated || isSecretUpdated)
	operatorUpdater.AddUpdater(replicationControllerUpdater)

	// Update S-O Service if necessary
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newSOperatorComponents.Services, namespace, "app=opssight")
	operatorUpdater.AddUpdater(serviceUpdater)

	// Update S-O ServiceAccount if necessary

	// Update S-O ClusterRoleBinding if necessary
	clusterRoleBindingUpdater, err := crdupdater.NewClusterRoleBinding(restconfig, kubeClient, newSOperatorComponents.ClusterRoleBindings, namespace, "app=opssight")
	operatorUpdater.AddUpdater(clusterRoleBindingUpdater)

	err = operatorUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func updatePrometheus(namespace string) error {
	// Get Components of Current Prometheus
	currPod, err := operatorutil.GetPod(kubeClient, namespace, "prometheus")
	currPrometheusImage := currPod.Spec.Containers[0].Image
	currPrometheusSpecConfig := soperator.PrometheusSpecConfig{
		Namespace:       deployNamespace,
		PrometheusImage: currPrometheusImage,
	}
	currPrometheusComponents, err := currPrometheusSpecConfig.GetComponents()
	fmt.Printf("%+v\n", currPrometheusComponents)

	// Get Components of New Prometheus
	newPrometheusSpecConfig := soperator.PrometheusSpecConfig{
		Namespace:       deployNamespace,
		PrometheusImage: deployPrometheusImage,
	}
	newPrometheusComponents, err := newPrometheusSpecConfig.GetComponents()
	fmt.Printf("%+v\n", newPrometheusComponents)

	prometheusUpdater := crdupdater.NewUpdater()

	// Update Prometheus ConfigMap
	_, err = crdupdater.UpdateConfigMap(kubeClient, deployNamespace, "prometheus", newPrometheusComponents.ConfigMaps[0])

	// Update Prometheus Deployment
	deploymentUpdater, err := crdupdater.NewDeployment(restconfig, kubeClient, newPrometheusComponents.Deployments, namespace, "app=prometheus", false)
	prometheusUpdater.AddUpdater(deploymentUpdater)

	// Update Prometheus Service
	serviceUpdater, err := crdupdater.NewService(restconfig, kubeClient, newPrometheusComponents.Services, namespace, "app=prometheus")
	prometheusUpdater.AddUpdater(serviceUpdater)

	err = prometheusUpdater.Update()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
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
	updateOperatorCmd.Flags().StringVarP(&deploySynopsysOperatorImage, "synopsys-operator-image", "i", deploySynopsysOperatorImage, "synopsys operator image URL")
	updateOperatorCmd.Flags().StringVarP(&deployPrometheusImage, "prometheus-image", "p", deployPrometheusImage, "prometheus image URL")
	updateOperatorCmd.Flags().StringVarP(&deployBlackduckRegistrationKey, "blackduck-registration-key", "k", deployBlackduckRegistrationKey, "key to register with KnowledgeBase")
	updateOperatorCmd.Flags().StringVarP(&deployDockerConfigPath, "docker-config", "d", deployDockerConfigPath, "path to docker config (image pull secrets etc)")
	updateOperatorCmd.Flags().StringVar(&deploySecretType, "secret-type", deploySecretType, "type of kubernetes secret for postgres and blackduck")
	updateOperatorCmd.Flags().StringVar(&deploySecretAdminPassword, "admin-password", deploySecretAdminPassword, "postgres admin password")
	updateOperatorCmd.Flags().StringVar(&deploySecretPostgresPassword, "postgres-password", deploySecretPostgresPassword, "postgres password")
	updateOperatorCmd.Flags().StringVar(&deploySecretUserPassword, "user-password", deploySecretUserPassword, "postgres user password")
	updateOperatorCmd.Flags().StringVar(&deploySecretBlackduckPassword, "blackduck-password", deploySecretBlackduckPassword, "blackduck password for 'sysadmin' account")
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
