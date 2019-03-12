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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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
		// 1. Get local copies of specs of all instances of crds (ex: opssight crds)
		// 2. Delete the CRD definition
		// 3. Create the new CRD definition
		// 4. Update the local specs of all instances with the new versions
		// 5. Update the resources in the cluster with the new specs (that contain the new version)

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

		// Load Spec into ctl tool
		updateBlackduckCtl.SetSpec(blackduck.Spec)

		// Check if it can be updated
		canUpdate, err := updateBlackduckCtl.CanUpdate()
		if err != nil {
			log.Errorf("Cannot Update: %s\n", err)
			return nil
		}
		if canUpdate {
			log.Debugf("Updating Blackduck...\n")
			// Update spec in the ctl tool with flags
			flagset := cmd.Flags()
			updateBlackduckCtl.SetChangedFlags(flagset)
			// Check differences between updated spec

			// Set Spec in the cluster
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
			newSpec := updateOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
			// Build New Horizon-Components from the updated OpsSight Spec
			opsSightSpecConfig := opssight.NewSpecConfig(&newSpec)
			horizonComponents, err := opsSightSpecConfig.GetComponents()
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's Config Map
			newConfigMapHorizon := horizonComponents.ConfigMaps[0]
			newConfigMapKube := newConfigMapHorizon.ToKube()
			err = updateOpsSightConfigMap(newConfigMapKube)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's Secret
			newSecretHorizon := horizonComponents.Secrets[0]
			newSecretKube := newSecretHorizon.ToKube()
			err = updateOpsSightSecret(newSecretKube)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's Services
			err = updateOpsSightServices(opsSightNamespace, horizonComponents.Services)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's ClusterRoles
			err = updateOpsSightClusterRoles(opsSightNamespace, horizonComponents.ClusterRoles)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's ClusterRoleBindings
			err = updateOpsSightClusterRoleBindings(opsSightNamespace, horizonComponents.ClusterRoleBindings)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			// Update OpsSight's Replication Controllers
			err = updateOpsSightReplicationControllers(opsSightNamespace, horizonComponents.ReplicationControllers, true, true)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}

			// Update in cluster
			opsSight.Spec = newSpec
			updateOpsSightInCluster(opsSightNamespace, opsSight)
		}
		return nil
	},
}

func updateOpsSightConfigMap(newConfigMap *corev1.ConfigMap) error {
	// Get Current Config Map
	oldConfigMap, err := util.GetConfigMap(kubeClient, newConfigMap.Namespace, newConfigMap.Namespace)
	if err != nil {
		return err
	}
	// Compare Data
	newConfigMapData := newConfigMap.Data
	oldConfigMapData := oldConfigMap.Data
	if !reflect.DeepEqual(newConfigMapData, oldConfigMapData) {
		oldConfigMap.Data = newConfigMapData
		err = util.UpdateConfigMap(kubeClient, oldConfigMap.Namespace, oldConfigMap)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func updateOpsSightSecret(newSecret *corev1.Secret) error {
	// Get Current Config Map
	oldSecret, err := util.GetSecret(kubeClient, newSecret.Namespace, newSecret.Namespace)
	if err != nil {
		return err
	}
	// TODO addSecret
	// Compare Data
	newSecretData := newSecret.Data
	oldSecretData := oldSecret.Data
	if !reflect.DeepEqual(newSecretData, oldSecretData) {
		oldSecret.Data = newSecretData
		err = util.UpdateSecret(kubeClient, oldSecret.Namespace, oldSecret)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func updateOpsSightServices(namespace string, services []*components.Service) error {
	deployer, err := util.NewDeployer(restconfig)
	if err != nil {
		return fmt.Errorf("unable to get deployer object for %s: %s", namespace, err)
	}
	isRun := false
	for _, service := range services {
		_, err := util.GetService(kubeClient, namespace, service.GetName())
		if err != nil {
			deployer.Deployer.AddService(service)
			isRun = true
		}
	}
	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy service object due to %+v", err)
		}
	}
	return nil
}

func updateOpsSightClusterRoles(namespace string, clusterRoles []*components.ClusterRole) error {
	deployer, err := util.NewDeployer(restconfig)
	if err != nil {
		return fmt.Errorf("unable to get deployer object for %s: %s", namespace, err)
	}
	isRun := false
	for _, clusterRole := range clusterRoles {
		_, err := util.GetClusterRole(kubeClient, clusterRole.GetName())
		if err != nil {
			deployer.Deployer.AddClusterRole(clusterRole)
			isRun = true
		}
	}
	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy cluster role object due to %+v", err)
		}
	}
	return nil
}

func updateOpsSightClusterRoleBindings(namespace string, clusterRoleBindings []*components.ClusterRoleBinding) error {
	deployer, err := util.NewDeployer(restconfig)
	if err != nil {
		return fmt.Errorf("unable to get deployer object for %s: %s", namespace, err)
	}
	isRun := false
	for _, clusterRoleBinding := range clusterRoleBindings {
		_, err := util.GetClusterRoleBinding(kubeClient, clusterRoleBinding.GetName())
		if err != nil {
			deployer.Deployer.AddClusterRoleBinding(clusterRoleBinding)
			isRun = true
		}
	}

	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy cluster role binding object due to %+v", err)
		}
	}
	return nil
}

func updateOpsSightReplicationControllers(opssight *opssightv1.OpsSightSpec, replicationControllers []*components.ReplicationController, isConfigMapUpdated bool, isSecretUpdated bool) error {
	// get old replication controller
	rcl, err := util.GetReplicationControllerList(kubeClient, opssight.Namespace, "app=opssight")
	if err != nil {
		return fmt.Errorf("unable to get opssight replication controllers for %s: %s", opssight.Namespace, err)
	}

	oldRCs := make(map[string]corev1.ReplicationController)
	for _, rc := range rcl.Items {
		oldRCs[rc.GetName()] = rc
	}

	// iterate through the replication controller list for any changes
	for _, component := range replicationControllers {
		newRCKube, err := component.ToKube()
		if err != nil {
			return fmt.Errorf("unable to convert rc %s to kube in opssight namespace %s: %s", component.GetName(), opssight.Namespace, err)
		}

		newRC := newRCKube.(*corev1.ReplicationController)
		oldRC := oldRCs[newRC.GetName()]

		// if the replication controller is not found in the cluster, create it
		if _, ok := oldRCs[newRC.GetName()]; !ok {
			deployer, err := util.NewDeployer(restconfig)
			if err != nil {
				return fmt.Errorf("unable to get deployer object for %s: %s", opssight.Namespace, err)
			}
			deployer.Deployer.AddReplicationController(component)
			deployer.Deployer.Run()
		}

		// if config map or secret is updated, patch the replication controller
		if isConfigMapUpdated || isSecretUpdated {
			err = util.PatchReplicationController(kubeClient, oldRC, *newRC)
			if err != nil {
				return fmt.Errorf("unable to patch rc %s to kube in opssight namespace %s: %s", component.GetName(), opssight.Namespace, err)
			}
			continue
		}

		// check whether the replication controller or its container got changed
		isChanged := false
		for _, oldContainer := range oldRC.Spec.Template.Spec.Containers {
			for _, newContainer := range newRC.Spec.Template.Spec.Containers {
				if strings.EqualFold(oldContainer.Name, newContainer.Name) &&
					!reflect.DeepEqual(
						opssight.ReplicationControllerComparator{
							Image:    oldContainer.Image,
							Replicas: oldRC.Spec.Replicas,
							MinCPU:   oldContainer.Resources.Requests.Cpu(),
							MaxCPU:   oldContainer.Resources.Limits.Cpu(),
							MinMem:   oldContainer.Resources.Requests.Memory(),
							MaxMem:   oldContainer.Resources.Limits.Memory(),
						},
						opssight.ReplicationControllerComparator{
							Image:    newContainer.Image,
							Replicas: newRC.Spec.Replicas,
							MinCPU:   newContainer.Resources.Requests.Cpu(),
							MaxCPU:   newContainer.Resources.Limits.Cpu(),
							MinMem:   newContainer.Resources.Requests.Memory(),
							MaxMem:   newContainer.Resources.Limits.Memory(),
						}) {
					isChanged = true
				}
			}
		}

		// if changed from the above step, patch the replication controller
		if isChanged {
			err = util.PatchReplicationController(kubeClient, oldRC, *newRC)
			if err != nil {
				return fmt.Errorf("unable to patch rc %s to kube in opssight namespace %s: %s", component.GetName(), opssight.Namespace, err)
			}
		}
	}
	return nil
}

func addSecretData(opsSight *opssightv1.OpsSightSpec, secret *components.Secret) error {
	blackduckPasswords := make(map[string]interface{})
	// adding External Black Duck passwords
	for _, host := range opsSight.Blackduck.ExternalHosts {
		blackduckPasswords[host.Domain] = &host
	}
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return err
	}
	secret.AddData(map[string][]byte{opsSight.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credential
	securedRegistries := make(map[string]interface{})
	for _, internalRegistry := range opsSight.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = &internalRegistry
	}
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return err
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})
	return nil
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
