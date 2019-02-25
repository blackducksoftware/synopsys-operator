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
	"strconv"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Resource Ctl for edit
var editBlackduckCtl = NewBlackduckCtl()
var editOpsSightCtl = NewOpsSightCtl()
var editAlertCtl = NewAlertCtl()

// editCmd edits non-synopsys resources
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Editing Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"edit"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			log.Errorf("Error Editing the Resource with KubeCmd: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// editBlackduckCmd edits a Blackduck by updating the spec
// or using the kube/oc editor
var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAME",
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
			bd, err := getBlackduckSpec(blackduckName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			editBlackduckCtl.Spec = &bd.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editBlackduckCtl.SetFlags)
			// Update Blackduck with Updates
			err = updateBlackduckSpec(bd)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
		} else {
			err := RunKubeEditorCmd("edit", "blackduck", blackduckName, "-n", blackduckName)
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
	Use:   "addPVC NAME PVC_NAME",
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
		bd, err := getBlackduckSpec(blackduckName)
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
		err = updateBlackduckSpec(bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddEnvironCmd adds an environ to a Blackduck
var editBlackduckAddEnvironCmd = &cobra.Command{
	Use:   "addEnviron NAME ENVIRON_NAME:ENVIRON_VALUE",
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
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Environ to Spec
		bd.Spec.Environs = append(bd.Spec.Environs, environ)
		// Update Blackduck with Environ
		err = updateBlackduckSpec(bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddRegistryCmd adds an Image Registry to a Blackduck
var editBlackduckAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry NAME REGISTRY",
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
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Registry to Spec
		bd.Spec.ImageRegistries = append(bd.Spec.ImageRegistries, registry)
		// Update Blackduck with Environ
		err = updateBlackduckSpec(bd)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editBlackduckAddUIDCmd adds a UID mapping to a Blackduck
var editBlackduckAddUIDCmd = &cobra.Command{
	Use:   "addUID NAME UID_KEY UID_VALUE",
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
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add UID Mapping to Spec
		intUIDVal, err := strconv.ParseInt(uidVal, 0, 64)
		if err != nil {
			fmt.Printf("Couldn't convert UID_VAL to int: %s\n", err)
		}
		bd.Spec.ImageUIDMap[uidKey] = intUIDVal
		// Update Blackduck with UID mapping
		err = updateBlackduckSpec(bd)
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
	Use:   "opssight NAME",
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
			ops, err := getOpsSightSpec(opsSightName)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
			editOpsSightCtl.Spec = &ops.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editOpsSightCtl.SetFlags)
			// Update OpsSight with Updates
			err = updateOpsSightSpec(ops)
			if err != nil {
				log.Errorf("%s", err)
				return nil
			}
		} else {
			err := RunKubeEditorCmd("edit", "opssight", opsSightName, "-n", opsSightName)
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
	Use:   "addRegistry NAME URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding Internal Registry to OpsSight\n")
		opssightName := args[0]
		regURL := args[1]
		regUser := args[2]
		regPass := args[3]
		// Get OpsSight Spec
		ops, err := getOpsSightSpec(opssightName)
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
		err = updateOpsSightSpec(ops)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
}

// editOpsSightAddHostCmd adds a Blackduck Host to an OpsSight
var editOpsSightAddHostCmd = &cobra.Command{
	Use:   "addHost NAME BLACKDUCK_HOST",
	Short: "Add a Blackduck Host to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Adding Blackduck Host to OpsSight\n")
		opssightName := args[0]
		host := args[1]
		// Get OpsSight Spec
		ops, err := getOpsSightSpec(opssightName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Add Host to Spec
		ops.Spec.Blackduck.Hosts = append(ops.Spec.Blackduck.Hosts, host)
		// Update OpsSight with Host
		err = updateOpsSightSpec(ops)
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
	Use:   "alert NAME",
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
			alt, err := getAlertSpec(alertName)
			if err != nil {
				log.Errorf("Get Spec: %s", err)
				return nil
			}
			editAlertCtl.Spec = &alt.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editAlertCtl.SetFlags)
			// Update Alert with Updates
			err = updateAlertSpec(alt)
			if err != nil {
				log.Errorf("Update Spec: %s", err)
				return nil
			}
		} else {
			err := RunKubeEditorCmd("edit", "alert", alertName, "-n", alertName)
			if err != nil {
				log.Errorf("Error Editing the Alert: %s", err)
				return nil
			}
		}
		return nil
	},
}

func init() {
	editCmd.DisableFlagParsing = true // lets editCmd pass flags to kube/oc
	rootCmd.AddCommand(editCmd)

	// Add Blackduck Commands
	editBlackduckCtl.AddSpecFlags(editBlackduckCmd)
	editCmd.AddCommand(editBlackduckCmd)

	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCSize, "size", blackduckPVCSize, "TODO")
	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCStorageClass, "storage-class", blackduckPVCStorageClass, "TODO")
	editBlackduckCmd.AddCommand(editBlackduckAddPVCCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddEnvironCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddRegistryCmd)

	editBlackduckCmd.AddCommand(editBlackduckAddUIDCmd)

	// Add OpsSight Commands
	editOpsSightCtl.AddSpecFlags(editOpsSightCmd)
	editCmd.AddCommand(editOpsSightCmd)

	editOpsSightCmd.AddCommand(editOpsSightAddRegistryCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddHostCmd)

	// Add Alert Spec Flags
	editAlertCtl.AddSpecFlags(editAlertCmd)
	editCmd.AddCommand(editAlertCmd)
}
