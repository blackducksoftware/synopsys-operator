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

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"edit"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			log.Errorf("Error Editing the Resource with KubeCmd: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAME",
	Short: "Edit an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			bd, err := getBlackduckSpec(blackduckName)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			globalBlackduckSpec = &bd.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(setBlackduckFlags)
			// Update Blackduck with Updates
			err = updateBlackduckSpec(bd)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "blackduck", blackduckName, "-n", blackduckName)
			if err != nil {
				fmt.Printf("Error Editing the Blackduck: %s\n", err)
			}
		}
	},
}

var blackduckPVCSize = "2Gi"
var blackduckPVCStorageClass = ""
var editBlackduckAddPVCCmd = &cobra.Command{
	Use:   "addPVC BLACKDUCK_NAME PVC_NAME",
	Short: "Add a PVC to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command takes 2 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding PVC to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		pvcName := args[1]
		// Get Blackduck Spec
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
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
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editBlackduckAddEnvironCmd = &cobra.Command{
	Use:   "addEnviron BLACKDUCK_NAME ENVIRON_NAME:ENVIRON_VALUE",
	Short: "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command accepts 2 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Environ to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		environ := args[1]
		// Get Blackduck Spec
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		// Add Environ to Spec
		bd.Spec.Environs = append(bd.Spec.Environs, environ)
		// Update Blackduck with Environ
		err = updateBlackduckSpec(bd)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editBlackduckAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry BLACKDUCK_NAME REGISTRY",
	Short: "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command accepts 2 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image Registry to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		registry := args[1]
		// Get Blackduck Spec
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		// Add Registry to Spec
		bd.Spec.ImageRegistries = append(bd.Spec.ImageRegistries, registry)
		// Update Blackduck with Environ
		err = updateBlackduckSpec(bd)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editBlackduckAddUIDCmd = &cobra.Command{
	Use:   "addUID BLACKDUCK_NAME UID_KEY UID_VALUE",
	Short: "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("This command accepts 3 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image UID to Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]
		uidKey := args[1]
		uidVal := args[2]
		// Get Blackduck Spec
		bd, err := getBlackduckSpec(blackduckName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
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
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editOpsSightCmd = &cobra.Command{
	Use:   "opssight OPSSIGHT_NAME",
	Short: "Edit an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			ops, err := getOpsSightSpec(opsSightName)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			globalOpsSightSpec = &ops.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(setOpsSightFlags)
			// Update OpsSight with Updates
			err = updateOpsSightSpec(ops)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "opssight", opsSightName, "-n", opsSightName)
			if err != nil {
				fmt.Printf("Error Editing the OpsSight: %s\n", err)
			}
		}
	},
}

var editOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry OPSSIGHT_NAME URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command takes 4 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Internal Registry to OpsSight\n")
		opssightName := args[0]
		regURL := args[1]
		regUser := args[2]
		regPass := args[3]
		// Get OpsSight Spec
		ops, err := getOpsSightSpec(opssightName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
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
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editOpsSightAddHostCmd = &cobra.Command{
	Use:   "addHost OPSSIGHT_NAME BLACKDUCK_HOST",
	Short: "Add a Blackduck Host to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command takes 2 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Blackduck Host to OpsSight\n")
		opssightName := args[0]
		host := args[1]
		// Get OpsSight Spec
		ops, err := getOpsSightSpec(opssightName)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		// Add Host to Spec
		ops.Spec.Blackduck.Hosts = append(ops.Spec.Blackduck.Hosts, host)
		// Update OpsSight with Host
		err = updateOpsSightSpec(ops)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	},
}

var editAlertCmd = &cobra.Command{
	Use:   "alert NAME",
	Short: "Edit an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an Alert\n")
		// Read Commandline Parameters
		alertName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			alt, err := getAlertSpec(alertName)
			if err != nil {
				fmt.Printf("Get Spec: %s\n", err)
				return
			}
			globalAlertSpec = &alt.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(setAlertFlags)
			// Update Alert with Updates
			err = updateAlertSpec(alt)
			if err != nil {
				fmt.Printf("Update Spec: %s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "alert", alertName, "-n", alertName)
			if err != nil {
				fmt.Printf("Error Editing the Alert: %s\n", err)
			}
		}
	},
}

func init() {
	editCmd.DisableFlagParsing = true
	rootCmd.AddCommand(editCmd)

	// Add Blackduck Spec Flags
	addBlackduckSpecFlags(editBlackduckCmd)
	editCmd.AddCommand(editBlackduckCmd)

	// Add Blackduck PVC Command
	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCSize, "size", blackduckPVCSize, "TODO")
	editBlackduckAddPVCCmd.Flags().StringVar(&blackduckPVCStorageClass, "storage-class", blackduckPVCStorageClass, "TODO")
	editBlackduckCmd.AddCommand(editBlackduckAddPVCCmd)
	// Add Blackduck Environ Command
	editBlackduckCmd.AddCommand(editBlackduckAddEnvironCmd)
	// Add Blackduck Registry Command
	editBlackduckCmd.AddCommand(editBlackduckAddRegistryCmd)
	// Add Blackduck Registry Command
	editBlackduckCmd.AddCommand(editBlackduckAddUIDCmd)

	// Add OpsSight Spec Flags
	addOpsSightSpecFlags(editOpsSightCmd)
	editCmd.AddCommand(editOpsSightCmd)

	// Add OpsSight Commands
	editOpsSightCmd.AddCommand(editOpsSightAddRegistryCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddHostCmd)

	// Add Alert Spec Flags
	addAlertSpecFlags(editAlertCmd)
	editCmd.AddCommand(editAlertCmd)
}
