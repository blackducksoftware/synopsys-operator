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
	"sort"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	polarisreporting "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	alertapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert"
	blackduckapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Command CRSpecBuilderFromCobraFlagsInterface
var createAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createBlackDuckCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createPolarisCobraHelper polaris.HelmValuesFromCobraFlags
var createPolarisReportingCobraHelper polarisreporting.HelmValuesFromCobraFlags

// Default Base Specs for Create
var baseAlertSpec string
var baseBlackDuckSpec string
var baseOpsSightSpec string

var namespace string

var alertNativePVC bool
var blackDuckNativeDatabase bool
var blackDuckNativePVC bool

// createCmd creates a Synopsys resource in the cluster
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

/*
Create Alert Commands
*/

var createAlertPreRun = func(cmd *cobra.Command, args []string) error {
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseAlertSpec = defaultBaseAlertSpec
	}
	log.Debugf("setting Alert's base spec to '%s'", baseAlertSpec)
	err := createAlertCobraHelper.SetPredefinedCRSpec(baseAlertSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateAlertSpecWithFlags(cmd *cobra.Command, alertName string, alertNamespace string) (*alertv1.Alert, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	alertInterface, err := createAlertCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	alertSpec, _ := alertInterface.(alertv1.AlertSpec)
	alertSpec.Namespace = alertNamespace

	// Create Alert CRD
	alert := &alertv1.Alert{
		ObjectMeta: metav1.ObjectMeta{
			Name:      alertName,
			Namespace: alertNamespace,
		},
		Spec: alertSpec,
	}
	alert.Kind = "Alert"
	alert.APIVersion = "synopsys.com/v1"
	return alert, nil
}

// createCmd creates an Alert instance
var createAlertCmd = &cobra.Command{
	Use:           "alert NAME",
	Example:       "synopsysctl create alert <name>\nsynopsysctl create alert <name> -n <namespace>\nsynopsysctl create alert <name> --mock json",
	Short:         "Create an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	PreRunE: createAlertPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		alertName := args[0]
		alertNamespace, crdNamespace, _, err := getInstanceInfo(mockMode, util.AlertCRDName, "", namespace, alertName)
		if err != nil {
			return err
		}
		alert, err := updateAlertSpecWithFlags(cmd, alertName, alertNamespace)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
			return PrintResource(*alert, mockFormat, false)
		}

		log.Infof("creating Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		if len(alert.Spec.Version) == 0 {
			versions := apps.NewApp(&protoform.Config{}, restconfig).Alert().Versions()
			sort.Sort(sort.Reverse(sort.StringSlice(versions)))
			alert.Spec.Version = versions[0]
		}

		// Deploy the Alert instance
		_, err = util.CreateAlert(alertClient, crdNamespace, alert)
		if err != nil {
			return fmt.Errorf("error creating Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
		}
		log.Infof("successfully submitted Alert '%s' into namespace '%s'", alertName, alertNamespace)
		return nil
	},
}

// createAlertNativeCmd prints the Kubernetes resources for creating an Alert instance
var createAlertNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create alert native <name>\nsynopsysctl create alert native <name> -n <namespace>\nsynopsysctl create alert native <name> -o yaml",
	Short:         "Print the Kubernetes resources for creating an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	PreRunE: createAlertPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName := args[0]
		alertNamespace, _, _, err := getInstanceInfo(true, util.AlertCRDName, "", namespace, alertName)
		if err != nil {
			return err
		}
		alert, err := updateAlertSpecWithFlags(cmd, alertName, alertNamespace)
		if err != nil {
			return err
		}

		log.Debugf("generating Kubernetes resources for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		app, err := getDefaultApp(nativeClusterType)
		if err != nil {
			return err
		}
		var cList *api.ComponentList
		switch {
		case alertNativePVC:
			cList, err = app.Alert().GetComponents(alert, alertapp.PVCResources)
		case !alertNativePVC:
			return PrintResource(*alert, nativeFormat, true)
		}
		if err != nil {
			return fmt.Errorf("failed to generate Black Duck components due to %+v", err)
		}
		if cList == nil {
			return fmt.Errorf("unable to genreate Black Duck components")
		}
		return PrintComponentListKube(cList, nativeFormat)
	},
}

/*
Create Black Duck Commands
*/

func checkPasswords(flagset *pflag.FlagSet) {
	if flagset.Lookup("external-postgres-host").Changed ||
		flagset.Lookup("external-postgres-port").Changed ||
		flagset.Lookup("external-postgres-admin").Changed ||
		flagset.Lookup("external-postgres-user").Changed ||
		flagset.Lookup("external-postgres-ssl").Changed ||
		flagset.Lookup("external-postgres-admin-password").Changed ||
		flagset.Lookup("external-postgres-user-password").Changed {
		// require all external-postgres parameters
		cobra.MarkFlagRequired(flagset, "external-postgres-host")
		cobra.MarkFlagRequired(flagset, "external-postgres-port")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin")
		cobra.MarkFlagRequired(flagset, "external-postgres-user")
		cobra.MarkFlagRequired(flagset, "external-postgres-ssl")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin-password")
		cobra.MarkFlagRequired(flagset, "external-postgres-user-password")
	} else {
		// user is explicitly required to set the postgres passwords for: 'admin', 'postgres', and 'user'
		cobra.MarkFlagRequired(flagset, "admin-password")
		cobra.MarkFlagRequired(flagset, "postgres-password")
		cobra.MarkFlagRequired(flagset, "user-password")
	}
}

func checkSealKey(flagset *pflag.FlagSet) {
	cobra.MarkFlagRequired(flagset, "seal-key")
}

var createBlackDuckPreRun = func(cmd *cobra.Command, args []string) error {
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseBlackDuckSpec = defaultBaseBlackDuckSpec
	}
	log.Debugf("setting Black Duck's base spec to '%s'", baseBlackDuckSpec)
	err := createBlackDuckCobraHelper.SetPredefinedCRSpec(baseBlackDuckSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateBlackDuckSpecWithFlags(cmd *cobra.Command, blackDuckName string, blackDuckNamespace string) (*blackduckv1.Blackduck, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	blackDuckInterface, err := createBlackDuckCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, fmt.Errorf("failed to generate the spec from the flags: %s", err)
	}

	// Set Namespace in Spec
	blackDuckSpec, _ := blackDuckInterface.(blackduckv1.BlackduckSpec)
	blackDuckSpec.Namespace = blackDuckNamespace

	// Create and Deploy Black Duck CRD
	blackDuck := &blackduckv1.Blackduck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      blackDuckName,
			Namespace: blackDuckNamespace,
		},
		Spec: blackDuckSpec,
	}
	blackDuck.Kind = "Blackduck"
	blackDuck.APIVersion = "synopsys.com/v1"

	return blackDuck, nil
}

// addPVCValuesToBlackDuckSpec returns the baseBlackDuckSpec with it's PVC values
func addPVCValuesToBlackDuckSpec(cmd *cobra.Command, blackDuckName string, blackDuckNamespace string, baseBlackDuckSpec blackduckv1.BlackduckSpec) (*blackduckv1.BlackduckSpec, error) {
	// Create a Black Duck configuration based on the flags
	blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to update Spec with flags: %+v", err)
	}
	// Get the PVCs based on the Black Duck configuration
	defaultPvcComponentsList, err := getBlackDuckPVCValues(blackDuck)
	if err != nil {
		return nil, fmt.Errorf("failed to get Black Duck PVC values: %+v", err)
	}
	// Add the PVCs to the base Black Duck spec
	baseBlackDuckSpec.PVC = convertHorizonPVCComponentToBlackDuckPVC(defaultPvcComponentsList)
	return &baseBlackDuckSpec, nil
}

func getBlackDuckPVCValues(bd *blackduckv1.Blackduck) ([]*components.PersistentVolumeClaim, error) {
	app, err := getDefaultApp(nativeClusterType)
	if err != nil {
		return nil, fmt.Errorf("failed to get Default App: %+v", err)
	}
	defaultPvcComponentsList, err := app.Blackduck().GetComponents(&blackduckv1.Blackduck{Spec: bd.Spec}, blackduckapp.PVCResources)
	if err != nil {
		return nil, fmt.Errorf("failed to get PVC Components List: %+v", err)
	}
	return defaultPvcComponentsList.PersistentVolumeClaims, nil
}

func convertHorizonPVCComponentToBlackDuckPVC(horizonPVCList []*components.PersistentVolumeClaim) []blackduckv1.PVC {
	blackDuckPVC := []blackduckv1.PVC{}
	for _, defaultPvcComponent := range horizonPVCList {
		defaultPvcComponentResourceQuantitySize := defaultPvcComponent.Spec.Resources.Requests[v1.ResourceStorage]
		pvc := blackduckv1.PVC{
			Name: defaultPvcComponent.Name[1:],
			Size: defaultPvcComponentResourceQuantitySize.String(),
		}
		if defaultPvcComponent.Spec.StorageClassName != nil {
			pvc.StorageClass = *defaultPvcComponent.Spec.StorageClassName
		}
		if defaultPvcComponent.Spec.VolumeName != "" {
			pvc.VolumeName = defaultPvcComponent.Spec.VolumeName
		}
		blackDuckPVC = append(blackDuckPVC, pvc)
	}
	return blackDuckPVC
}

// createBlackDuckCmd creates a Black Duck instance
var createBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsysctl create blackduck <name>\nsynopsysctl create blackduck <name> -n <namespace>\nsynopsysctl create blackduck <name> --mock json",
	Short:         "Create a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		checkPasswords(cmd.Flags())
		return nil
	},
	PreRunE: createBlackDuckPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		blackDuckName := args[0]
		blackDuckNamespace, crdNamespace, _, err := getInstanceInfo(mockMode, util.BlackDuckCRDName, "", namespace, blackDuckName)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

			// Update the BlackDuck spec in 'createBlackDuckCobraHelper' with the correct PVC values
			blackDuckSpecInterface := createBlackDuckCobraHelper.GetCRSpec()
			baseBlackDuckSpec, _ := blackDuckSpecInterface.(blackduckv1.BlackduckSpec)
			baseBlackDuckSpecWithPVCs, err := addPVCValuesToBlackDuckSpec(cmd, blackDuckName, blackDuckNamespace, baseBlackDuckSpec)
			if err != nil {
				return fmt.Errorf("failed to add PVCs to Black Duck spec: %+v", err)
			}

			err = createBlackDuckCobraHelper.SetCRSpec(*baseBlackDuckSpecWithPVCs)
			if err != nil {
				return fmt.Errorf("error setting Spec with PVC values: %s", err)
			}

			// Update the CR in createBlackDuckCobraHelper with user's flags
			cmd.Flag("pvc-file-path").Changed = false // we already did the special logic above to set the PVCs
			blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
			if err != nil {
				return err
			}

			// add versions
			if len(blackDuck.Spec.Version) == 0 {
				// versions := apps.NewApp(&protoform.Config{}, restconfig).Blackduck().Versions()
				// sort.Sort(sort.Reverse(sort.StringSlice(versions)))
				// TODO: fix the sort logic for Black Duck version
				blackDuck.Spec.Version = "2020.2.1"
			}
			versionSupportsSecurityContexts, err := util.IsVersionGreaterThanOrEqualTo(blackDuck.Spec.Version, 2019, time.December, 0)
			if err != nil {
				return fmt.Errorf("failed to check Black Duck version: %s", err)
			}
			if !versionSupportsSecurityContexts && cmd.Flags().Changed("security-context-file-path") {
				log.Warnf("security contexts from --security-context-file-path are ignored for versions before 2019.12.0, you're using version %s", blackDuck.Spec.Version)
			}

			return PrintResource(*blackDuck, mockFormat, false)
		}

		blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
		if err != nil {
			return err
		}

		log.Infof("creating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		if len(blackDuck.Spec.Version) == 0 {
			// versions := apps.NewApp(&protoform.Config{}, restconfig).Blackduck().Versions()
			// sort.Sort(sort.Reverse(sort.StringSlice(versions)))
			// TODO: fix the sort logic for Black Duck version
			blackDuck.Spec.Version = "2020.2.1"
		}
		versionSupportsSecurityContexts, err := util.IsVersionGreaterThanOrEqualTo(blackDuck.Spec.Version, 2019, time.December, 0)
		if err != nil {
			return fmt.Errorf("failed to check Black Duck version: %s", err)
		}
		if !versionSupportsSecurityContexts && cmd.Flags().Changed("security-context-file-path") {
			return fmt.Errorf("security contexts from --security-context-file-path cannot be set for versions before 2019.12.0, you're using version %s", blackDuck.Spec.Version)
		}
		if util.IsOpenshift(kubeClient) && cmd.Flags().Changed("security-context-file-path") {
			return fmt.Errorf("cannot set security contexts with --security-context-file-path in an Openshift environment")
		}

		if isBlackDuckVersionSupportMultipleInstance, _ := util.IsBlackDuckVersionSupportMultipleInstance(blackDuck.Spec.Version); !isBlackDuckVersionSupportMultipleInstance {
			// Verifying only one Black Duck instance per namespace
			blackducks, err := util.ListBlackduck(blackDuckClient, crdNamespace, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("unable to list Black Duck instances in namespace '%s' due to %+v", blackDuckNamespace, err)
			}

			for _, v := range blackducks.Items {
				if strings.EqualFold(v.Spec.Namespace, blackDuckNamespace) {
					return fmt.Errorf("a Black Duck instance already exists in namespace '%s', only one instance per namespace is allowed", blackDuckNamespace)
				}
			}
		}

		// Deploy the Black Duck instance
		log.Debugf("deploying Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		_, err = util.CreateBlackduck(blackDuckClient, crdNamespace, blackDuck)
		if err != nil {
			return fmt.Errorf("error creating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		log.Infof("successfully submitted Black Duck '%s' into namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// createBlackDuckNativeCmd prints the Kubernetes resources for creating a Black Duck instance
var createBlackDuckNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create blackduck native <name>\nsynopsysctl create blackduck native <name> -n <namespace>\nsynopsysctl create blackduck native <name> -o yaml",
	Short:         "Print the Kubernetes resources for creating a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 arguments")
		}
		if blackDuckNativeDatabase && blackDuckNativePVC {
			return fmt.Errorf("cannot enable --output-database and --output-pvc, please only specify one")
		}
		if blackDuckNativeDatabase {
			checkPasswords(cmd.Flags())
		}
		checkSealKey(cmd.Flags())
		return nil
	},
	PreRunE: createBlackDuckPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName := args[0]
		blackDuckNamespace, _, _, err := getInstanceInfo(true, util.BlackDuckCRDName, "", namespace, blackDuckName)
		if err != nil {
			return err
		}
		blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
		if err != nil {
			return err
		}

		// Security Contexts check
		newVersionIsGreaterThanOrEqualv2019x12x0, err := util.IsVersionGreaterThanOrEqualTo(blackDuck.Spec.Version, 2019, time.December, 0)
		if err != nil {
			return err
		}
		if !newVersionIsGreaterThanOrEqualv2019x12x0 && cmd.Flags().Changed("security-context-file-path") {
			return fmt.Errorf("security contexts from --security-context-file-path cannot be set for versions before 2019.12.0, you're using version %s", blackDuck.Spec.Version)
		}

		log.Debugf("generating Kubernetes resources for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		app, err := getDefaultApp(nativeClusterType)
		if err != nil {
			return err
		}
		var cList *api.ComponentList
		blackDuck.Spec.LivenessProbes = true // enable LivenessProbes when generating Kubernetes resources for customers
		switch {
		case !blackDuckNativeDatabase && !blackDuckNativePVC:
			return PrintResource(*blackDuck, nativeFormat, true)
		case blackDuckNativeDatabase:
			cList, err = app.Blackduck().GetComponents(blackDuck, blackduckapp.DatabaseResources)
		case blackDuckNativePVC:
			cList, err = app.Blackduck().GetComponents(blackDuck, blackduckapp.PVCResources)
		}
		if err != nil {
			return fmt.Errorf("failed to generate Black Duck components due to %+v", err)
		}
		if cList == nil {
			return fmt.Errorf("unable to genreate Black Duck components")
		}
		return PrintComponentListKube(cList, nativeFormat)
	},
}

/*
Create OpsSight Commands
*/

var createOpsSightPreRun = func(cmd *cobra.Command, args []string) error {
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseOpsSightSpec = defaultBaseOpsSightSpec
	}
	log.Debugf("setting OpsSight's base spec to '%s'", baseOpsSightSpec)
	err := createOpsSightCobraHelper.SetPredefinedCRSpec(baseOpsSightSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateOpsSightSpecWithFlags(cmd *cobra.Command, opsSightName string, opsSightNamespace string) (*opssightv1.OpsSight, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	opsSightInterface, err := createOpsSightCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	opsSightSpec, _ := opsSightInterface.(opssightv1.OpsSightSpec)
	opsSightSpec.Namespace = opsSightNamespace

	// Create and Deploy OpsSight CRD
	opsSight := &opssightv1.OpsSight{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opsSightName,
			Namespace: opsSightNamespace,
		},
		Spec: opsSightSpec,
	}
	opsSight.Kind = "OpsSight"
	opsSight.APIVersion = "synopsys.com/v1"
	return opsSight, nil
}

// createOpsSightCmd creates an OpsSight instance
var createOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME",
	Example:       "synopsysctl create opssight <name>\nsynopsysctl create opssight <name> -n <namespace>\nsynopsysctl create opssight <name> --mock json",
	Short:         "Create an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 arguments")
		}
		return nil
	},
	PreRunE: createOpsSightPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		opsSightName := args[0]
		opsSightNamespace, crdNamespace, _, err := getInstanceInfo(mockMode, util.OpsSightCRDName, "", namespace, opsSightName)
		if err != nil {
			return err
		}
		opsSight, err := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*opsSight, mockFormat, false)
		}

		log.Infof("creating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

		// Deploy the OpsSight instance
		_, err = util.CreateOpsSight(opsSightClient, crdNamespace, opsSight)
		if err != nil {
			return fmt.Errorf("error creating the OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		log.Infof("successfully submitted OpsSight '%s' into namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// createOpsSightNativeCmd prints the Kubernetes resources for creating an OpsSight instance
var createOpsSightNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create opssight native <name>\nsynopsysctl create opssight native <name> -n <namespace>\nsynopsysctl create opssight native <name> -o yaml",
	Short:         "Print the Kubernetes resources for creating an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	PreRunE: createOpsSightPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
		opsSightNamespace, _, _, err := getInstanceInfo(true, util.OpsSightCRDName, "", namespace, opsSightName)
		if err != nil {
			return err
		}
		opsSight, err := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)
		if err != nil {
			return err
		}

		log.Debugf("generating Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*opsSight, nativeFormat, true)
	},
}

// createPolarisCmd creates a Polaris instance
var createPolarisCmd = &cobra.Command{
	Use:           "polaris",
	Short:         "Create a Polaris instance. (Please make sure you have read and understand prerequisites before installing Polaris: [https://synopsys.atlassian.net/wiki/spaces/POP/overview])",
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: "\nRequried flags for setup with external database:\n\n 	synopsysctl create polaris --namespace 'onprem' --version '2020.04' --gcp-service-account-path '<PATH>/gcp-service-account-file.json' --coverity-license-path '<PATH>/coverity-license-file.xml' --fqdn 'example.polaris.com' --smtp-host 'example.smtp.com' --smtp-port 25 --smtp-username 'example' --smtp-password 'example' --smtp-sender-email 'example.email.com' --postgres-host 'example.postgres.com' --postgres-port 5432 --postgres-username 'example' --postgres-password 'example' ",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		// TODO: allow user to specify --version and --chart-location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf("failed to create Polaris resources: %+v", err)
		}

		// Deploy Polaris-Reporting Resources
		err = util.CreateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf("failed to create Polaris resources: %+v", err)
		}

		log.Infof("Polaris has been successfully Created!")
		return nil
	},
}

// createPolarisNativeCmd prints the Kubernetes resources for creating a Polaris instance
var createPolarisNativeCmd = &cobra.Command{
	Use:           "native",
	Short:         "Print Kubernetes resources for creating a Polaris instance (Please make sure you have read and understand prerequisites before installing Polaris: [https://synopsys.atlassian.net/wiki/spaces/POP/overview])",
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: "\nRequried flags for setup with external database:\n\n 	synopsysctl create polaris native --namespace 'onprem' --version '2020.04' --gcp-service-account-path '<PATH>/gcp-service-account-file.json' --coverity-license-path '<PATH>/coverity-license-file.xml' --fqdn 'example.polaris.com' --smtp-host 'example.smtp.com' --smtp-port 25 --smtp-username 'example' --smtp-password 'example' --smtp-sender-email 'example.email.com' --postgres-host 'example.postgres.com' --postgres-port 5432 --postgres-username 'example' --postgres-password 'example' ",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisReportingChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreporting.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to create GCP Service Account Secrets: %+v", err)
		}

		// Print the Secret
		for _, obj := range gcpServiceAccountSecrets {
			PrintComponent(obj, "YAML") // helm only supports yaml
		}

		// Print Polaris-Reporting Resources
		err = util.TemplateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to generate Polaris-Reporting resources: %+v", err)
		}

		return nil
	},
}

func polarisPostgresCheck(flagset *pflag.FlagSet) error {
	usingPostgresContainer, _ := flagset.GetBool("enable-postgres-container")
	if usingPostgresContainer {
		if flagset.Lookup("postgres-host").Changed || flagset.Lookup("postgres-port").Changed || flagset.Lookup("postgres-username").Changed {
			return fmt.Errorf("cannot change the host, port and username when using the postgres container")
		}
		if flagset.Lookup("postgres-ssl-mode").Changed {
			return fmt.Errorf("cannot enable SSL when using postgres container")
		}
	} else {
		if flagset.Lookup("postgres-size").Changed {
			return fmt.Errorf("cannot configure the postgresql size when using an external database")
		}
		// External DB. Host, port and username are mandatory
		cobra.MarkFlagRequired(flagset, "postgres-host")
		cobra.MarkFlagRequired(flagset, "postgres-port")
		cobra.MarkFlagRequired(flagset, "postgres-username")
	}

	return nil
}

// createPolarisReportingCmd creates a Polaris-Reporting instance
var createPolarisReportingCmd = &cobra.Command{
	Use:           "polaris-reporting",
	Example:       "",
	Short:         "Create a Polaris-Reporting instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisReportingCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		// TODO: allow user to specify --version and --chart-location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisReportingChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf("failed to create Polaris-Reporting resources: %+v", err)
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreporting.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to create GCP Service Account Secrets: %+v", err)
		}

		// Deploy the Secret
		err = KubectlApplyRuntimeObjects(gcpServiceAccountSecrets)
		if err != nil {
			return fmt.Errorf("failed to deploy the gcpServiceAccount Secrets: %s", err)
		}

		// Deploy Polaris-Reporting Resources
		err = util.CreateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf("failed to create Polaris-Reporting resources: %+v", err)
		}

		log.Infof("Polaris-Reporting has been successfully Created!")
		return nil
	},
}

// createPolarisReportingNativeCmd prints Polaris-Reporting resources
var createPolarisReportingNativeCmd = &cobra.Command{
	Use:           "native",
	Example:       "",
	Short:         "Print Kubernetes resources for creating a Polaris-Reporting instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisReportingCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisReportingChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreporting.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to create GCP Service Account Secrets: %+v", err)
		}

		// Print the Secret
		for _, obj := range gcpServiceAccountSecrets {
			PrintComponent(obj, "YAML") // helm only supports yaml
		}

		// Print Polaris-Reporting Resources
		err = util.TemplateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to generate Polaris-Reporting resources: %+v", err)
		}

		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()
	createBlackDuckCobraHelper = blackduck.NewCRSpecBuilderFromCobraFlags()
	createOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	createPolarisCobraHelper = *polaris.NewHelmValuesFromCobraFlags()
	createPolarisReportingCobraHelper = *polarisreporting.NewHelmValuesFromCobraFlags()

	rootCmd.AddCommand(createCmd)

	// Add Alert Command
	createAlertCmd.PersistentFlags().StringVar(&baseAlertSpec, "template", baseAlertSpec, "Base resource configuration to modify with flags [empty|default]")
	createAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createAlertCobraHelper.AddCRSpecFlagsToCommand(createAlertCmd, true)
	addMockFlag(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)

	createAlertCobraHelper.AddCRSpecFlagsToCommand(createAlertNativeCmd, true)
	addNativeFormatFlag(createAlertNativeCmd)
	createAlertNativeCmd.Flags().BoolVar(&alertNativePVC, "output-pvc", alertNativePVC, "If true, output resources for only Alert's persistent volume claims")
	createAlertCmd.AddCommand(createAlertNativeCmd)

	// Add Black Duck Command
	createBlackDuckCmd.PersistentFlags().StringVar(&baseBlackDuckSpec, "template", baseBlackDuckSpec, "Base resource configuration to modify with flags [empty|persistentStorageLatest|persistentStorageV1|externalPersistentStorageLatest|externalPersistentStorageV1|bdba|ephemeral|ephemeralCustomAuthCA|externalDB|IPV6Disabled]")
	createBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createBlackDuckCobraHelper.AddCRSpecFlagsToCommand(createBlackDuckCmd, true)
	addMockFlag(createBlackDuckCmd)
	createCmd.AddCommand(createBlackDuckCmd)

	createBlackDuckCobraHelper.AddCRSpecFlagsToCommand(createBlackDuckNativeCmd, true)
	addNativeFormatFlag(createBlackDuckNativeCmd)
	createBlackDuckNativeCmd.Flags().BoolVar(&blackDuckNativeDatabase, "output-database", blackDuckNativeDatabase, "If true, output resources for only Black Duck's database")
	createBlackDuckNativeCmd.Flags().BoolVar(&blackDuckNativePVC, "output-pvc", blackDuckNativePVC, "If true, output resources for only Black Duck's persistent volume claims")
	createBlackDuckCmd.AddCommand(createBlackDuckNativeCmd)

	// Add OpsSight Command
	createOpsSightCmd.PersistentFlags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags [empty|upstream|default|disabledBlackDuck]")
	createOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightCmd, true)
	addMockFlag(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightNativeCmd, true)
	addNativeFormatFlag(createOpsSightNativeCmd)
	createOpsSightCmd.AddCommand(createOpsSightNativeCmd)

	// Add Polaris commands
	createPolarisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createPolarisCmd.PersistentFlags(), "namespace")
	createPolarisCobraHelper.AddCobraFlagsToCommand(createPolarisCmd, true)
	addChartLocationPathFlag(createPolarisCmd)
	createCmd.AddCommand(createPolarisCmd)

	createPolarisCobraHelper.AddCobraFlagsToCommand(createPolarisNativeCmd, true)
	addChartLocationPathFlag(createPolarisNativeCmd)
	createPolarisCmd.AddCommand(createPolarisNativeCmd)

	// Add Polaris-Reporting commands
	createPolarisReportingCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createPolarisReportingCmd.PersistentFlags(), "namespace")
	createPolarisReportingCobraHelper.AddCobraFlagsToCommand(createPolarisReportingCmd, true)
	addChartLocationPathFlag(createPolarisReportingCmd)
	createCmd.AddCommand(createPolarisReportingCmd)

	createPolarisReportingCobraHelper.AddCobraFlagsToCommand(createPolarisReportingNativeCmd, true)
	addChartLocationPathFlag(createPolarisReportingNativeCmd)
	createPolarisReportingCmd.AddCommand(createPolarisReportingNativeCmd)

}
