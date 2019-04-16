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
	"strconv"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Resource Ctl for edit
var editBlackduckCtl ResourceCtl
var editOpsSightCtl ResourceCtl
var editAlertCtl ResourceCtl

// editCmd edits non-synopsys resources
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	// Display synopsysctl's Help instead of sending to oc/kubectl
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Editing Non-Synopsys Resource\n")
		//(PassCmd) kubeCmdArgs := append([]string{"edit"}, args...)
		//(PassCmd) out, err := operatorutil.RunKubeCmd(restconfig, kube, openshift, kubeCmdArgs...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	log.Errorf("Error Editing the Resource with KubeCmd: %s", out)
		//(PassCmd) 	return nil
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
	},
}

// editBlackduckCmd edits a Blackduck by updating the spec
// or using the kube/oc editor
var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Edit an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Editing Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			bd, err := operatorutil.GetHub(blackduckClient, blackduckName, blackduckName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			editBlackduckCtl.SetSpec(bd.Spec)
			// Update Spec with User's Flags
			editBlackduckCtl.SetChangedFlags(flagset)
			// Update Blackduck with Updates
			blackduckSpec := editBlackduckCtl.GetSpec().(blackduckv1.BlackduckSpec)
			bd.Spec = blackduckSpec
			_, err = operatorutil.UpdateBlackduck(blackduckClient, blackduckName, bd)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
		} else {
			err := operatorutil.RunKubeEditorCmd(restconfig, kube, openshift, "edit", "blackduck", blackduckName, "-n", blackduckName)
			if err != nil {
				log.Errorf("Error Editing the Blackduck: %s", err)
				return nil
			}
		}
		return nil
	},
}

var blackduckPVCSize = "2Gi"
var blackduckPVCStorageClass = ""

// editBlackduckAddPVCCmd adds a PVC to a Blackduck
var editBlackduckAddPVCCmd = &cobra.Command{
	Use:   "addPVC NAMESPACE PVC_NAME",
	Short: "Add a PVC to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command takes 2 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding PVC to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		pvcName := args[1]
		// Get Blackduck Spec
		bd, err := operatorutil.GetHub(blackduckClient, blackduckName, blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add PVC to Spec
		newPVC := blackduckv1.PVC{
			Name:         pvcName,
			Size:         blackduckPVCSize,
			StorageClass: blackduckPVCStorageClass,
		}
		bd.Spec.PVC = append(bd.Spec.PVC, newPVC)
		// Update Blackduck with PVC
		_, err = operatorutil.UpdateBlackduck(blackduckClient, blackduckName, bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddEnvironCmd adds an environ to a Blackduck
var editBlackduckAddEnvironCmd = &cobra.Command{
	Use:   "addEnviron NAMESPACE ENVIRON_NAME:ENVIRON_VALUE",
	Short: "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command accepts 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding Environ to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		environ := args[1]
		// Get Blackduck Spec
		bd, err := operatorutil.GetHub(blackduckClient, blackduckName, blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Environ to Spec
		bd.Spec.Environs = append(bd.Spec.Environs, environ)
		// Update Blackduck with Environ
		_, err = operatorutil.UpdateBlackduck(blackduckClient, blackduckName, bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddRegistryCmd adds an Image Registry to a Blackduck
var editBlackduckAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry NAMESPACE REGISTRY",
	Short: "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command accepts 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding an Image Registry to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		registry := args[1]
		// Get Blackduck Spec
		bd, err := operatorutil.GetHub(blackduckClient, blackduckName, blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Registry to Spec
		bd.Spec.ImageRegistries = append(bd.Spec.ImageRegistries, registry)
		// Update Blackduck with Environ
		_, err = operatorutil.UpdateBlackduck(blackduckClient, blackduckName, bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddUIDCmd adds a UID mapping to a Blackduck
var editBlackduckAddUIDCmd = &cobra.Command{
	Use:   "addUID NAMESPACE UID_KEY UID_VALUE",
	Short: "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("This command accepts 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding an Image UID to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		uidKey := args[1]
		uidVal := args[2]
		// Get Blackduck Spec
		bd, err := operatorutil.GetHub(blackduckClient, blackduckName, blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add UID Mapping to Spec
		intUIDVal, err := strconv.ParseInt(uidVal, 0, 64)
		if err != nil {
			fmt.Printf("Couldn't convert UID_VAL to int: %s\n", err)
		}
		if bd.Spec.ImageUIDMap == nil {
			bd.Spec.ImageUIDMap = make(map[string]int64)
		}
		bd.Spec.ImageUIDMap[uidKey] = intUIDVal
		// Update Blackduck with UID mapping
		_, err = operatorutil.UpdateBlackduck(blackduckClient, blackduckName, bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editOpsSightCmd edits an OpsSight by updating the spec
// or using the kube/oc editor
var editOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Edit an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Editing an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			ops, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			editOpsSightCtl.SetSpec(ops.Spec)
			// Update Spec with User's Flags
			editOpsSightCtl.SetChangedFlags(flagset)
			// Update OpsSight with Updates
			opsSightSpec := editOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
			ops.Spec = opsSightSpec
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, ops)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
		} else {
			err := operatorutil.RunKubeEditorCmd(restconfig, kube, openshift, "edit", "opssight", opsSightName, "-n", opsSightName)
			if err != nil {
				log.Errorf("Error Editing the OpsSight: %s", err)
				return nil
			}
		}
		return nil
	},
}

// editOpsSightAddRegistryCmd adds a registry to an OpsSight
var editOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry NAMESPACE URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight",
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
		ops, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
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
		ops.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ops.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
		// Update OpsSight with Internal Registry
		_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, ops)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editOpsSightAddHostCmd adds a Blackduck Host to an OpsSight
var editOpsSightAddHostCmd = &cobra.Command{
	Use:   "addHost NAMESPACE DOMAIN PORT",
	Short: "Add a Blackduck Host to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("This command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding Blackduck Host to OpsSight\n")
		opsSightName := args[0]
		domain := args[1]
		port := args[2]
		// Get OpsSight Spec
		ops, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Host to Spec
		host := opssightv1.Host{}
		host.Domain = domain
		intPort, err := strconv.ParseInt(port, 0, 64)
		if err != nil {
			log.Errorf("Couldn't convert Port '%s' to int", port)
		}
		host.Port = int(intPort)
		ops.Spec.Blackduck.ExternalHosts = append(ops.Spec.Blackduck.ExternalHosts, &host)
		// Update OpsSight with Host
		_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, ops)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editAlertCmd edits an Alert by updating the spec
// or using the kube/oc editor
var editAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Edit an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Editing an Alert\n")
		// Read Commandline Parameters
		alertName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			alt, err := operatorutil.GetAlert(alertClient, alertName, alertName)
			if err != nil {
				log.Errorf("Get Spec: %s", err)
				return nil
			}
			editAlertCtl.SetSpec(alt.Spec)
			// Update Spec with User's Flags
			editAlertCtl.SetChangedFlags(flagset)
			// Update Alert with Updates
			alertSpec := editAlertCtl.GetSpec().(alertv1.AlertSpec)
			alt.Spec = alertSpec
			_, err = operatorutil.UpdateAlert(alertClient, alertName, alt)
			if err != nil {
				log.Errorf("Update Spec: %s", err)
				return nil
			}
		} else {
			err := operatorutil.RunKubeEditorCmd(restconfig, kube, openshift, "edit", "alert", alertName, "-n", alertName)
			if err != nil {
				log.Errorf("Error Editing the Alert: %s", err)
				return nil
			}
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	editBlackduckCtl = blackduck.NewBlackduckCtl()
	editOpsSightCtl = opssight.NewOpsSightCtl()
	editAlertCtl = alert.NewAlertCtl()

	//(PassCmd) editCmd.DisableFlagParsing = true // lets editCmd pass flags to kube/oc
	rootCmd.AddCommand(editCmd)

	// Add Blackduck Edit Commands
	editBlackduckCtl.AddSpecFlags(editBlackduckCmd, true)
	editCmd.AddCommand(editBlackduckCmd)

	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCSize, "size", blackduckPVCSize, "TODO")
	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCStorageClass, "storage-class", blackduckPVCStorageClass, "TODO")
	editBlackduckCmd.AddCommand(editBlackduckAddPVCCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddEnvironCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddRegistryCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddUIDCmd)

	// Add OpsSight Edit Commands
	editOpsSightCtl.AddSpecFlags(editOpsSightCmd, true)
	editCmd.AddCommand(editOpsSightCmd)

	editOpsSightCmd.AddCommand(editOpsSightAddRegistryCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddHostCmd)

	// Add Alert Edit Comamnds
	editAlertCtl.AddSpecFlags(editAlertCmd, true)
	editCmd.AddCommand(editAlertCmd)
}
