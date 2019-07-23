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
	"strings"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	alertapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/alert"
	blackduckapp "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Command CRSpecBuilderFromCobraFlagsInterface
var createAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createBlackDuckCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface

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
		alertName, alertNamespace, scope, err := getInstanceInfo(mockMode, util.AlertCRDName, "", namespace, args[0])
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

		// Get the Synopsys Operator namespace by CRD scope
		operatorNamespace, err := util.GetOperatorNamespaceByCRDScope(kubeClient, util.AlertCRDName, scope, alertNamespace)
		if err != nil {
			return err
		}

		// Deploy the Alert instance
		_, err = alertClient.SynopsysV1().Alerts(operatorNamespace).Create(alert)
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
		alertName, alertNamespace, _, err := getInstanceInfo(true, util.AlertCRDName, "", namespace, args[0])
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
		return nil, err
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

	// Get PVCs from app
	app, err := getDefaultApp(nativeClusterType)
	if err != nil {
		return nil, err
	}
	cList, err := app.Blackduck().GetComponents(blackDuck, blackduckapp.PVCResources)
	m := map[string]string{}
	for _, pvc := range cList.PersistentVolumeClaims {
		name := pvc.GetName()
		sizeQuanitity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		sizeCanonical := sizeQuanitity.String()
		m[name] = sizeCanonical
	}
	// Add User's PVCs to the map
	currSpecPVCs := blackDuck.Spec.PVC
	for _, specPvc := range currSpecPVCs {
		m[specPvc.Name] = specPvc.Size

	}
	// Add the map's PVCs to the spec
	blackDuck.Spec.PVC = []blackduckv1.PVC{}
	for pvcName, pvcSize := range m {
		blackDuck.Spec.PVC = append(blackDuck.Spec.PVC,
			blackduckv1.PVC{
				Name: pvcName,
				Size: pvcSize,
			})
	}

	return blackDuck, nil
}

// createBlackDuckCmd creates a Black Duck instance
var createBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsysctl create blackduck <name>\nsynopsysctl create blackduck <name> -n <namespace>\nsynopsysctl create blackduck <name> --mock json",
	Short:         "Create a Black Duck instance",
	Aliases:       []string{"bds", "bd"},
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
		blackDuckName, blackDuckNamespace, scope, err := getInstanceInfo(mockMode, util.BlackDuckCRDName, "", namespace, args[0])
		if err != nil {
			return err
		}
		blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			return PrintResource(*blackDuck, mockFormat, false)
		}

		log.Infof("creating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

		// Verifying only one Black Duck instance per namespace
		blackducks, err := util.ListHubs(blackDuckClient, blackDuckNamespace)
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances in namespace '%s' due to %+v", blackDuckNamespace, err)
		}
		for _, v := range blackducks.Items {
			if strings.EqualFold(v.Spec.Namespace, blackDuckNamespace) {
				return fmt.Errorf("a Black Duck instance already exists in namespace '%s', only one instance per namespace is allowed", blackDuckNamespace)
			}
		}

		// Get the Synopsys Operator namespace by CRD scope
		operatorNamespace, err := util.GetOperatorNamespaceByCRDScope(kubeClient, util.BlackDuckCRDName, scope, blackDuckNamespace)
		if err != nil {
			return err
		}

		// Deploy the Black Duck instance
		log.Debugf("deploying Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		_, err = blackDuckClient.SynopsysV1().Blackducks(operatorNamespace).Create(blackDuck)
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
		return nil
	},
	PreRunE: createBlackDuckPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(true, util.BlackDuckCRDName, "", namespace, args[0])
		if err != nil {
			return err
		}
		blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
		if err != nil {
			return err
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
	Example:       "synopsysctl create opssight <name>\nsynopsysctl create opssight <name> --mock json",
	Short:         "Create an OpsSight instance",
	Aliases:       []string{"ops"},
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
		opsSightName, opsSightNamespace, scope, err := getInstanceInfo(mockMode, util.OpsSightCRDName, "", namespace, args[0])
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

		// Get the Synopsys Operator namespace by CRD scope
		operatorNamespace, err := util.GetOperatorNamespaceByCRDScope(kubeClient, util.OpsSightCRDName, scope, opsSightNamespace)
		if err != nil {
			return err
		}

		// Deploy the OpsSight instance
		_, err = opsSightClient.SynopsysV1().OpsSights(operatorNamespace).Create(opsSight)
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
	Example:       "synopsysctl create opssight native <name>\nsynopsysctl create opssight native <name> -o yaml",
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
		opsSightName, opsSightNamespace, _, err := getInstanceInfo(true, util.OpsSightCRDName, "", namespace, args[0])
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

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()
	createBlackDuckCobraHelper = blackduck.NewCRSpecBuilderFromCobraFlags()
	createOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()

	rootCmd.AddCommand(createCmd)

	// Add Alert Command
	createAlertCmd.PersistentFlags().StringVar(&baseAlertSpec, "template", baseAlertSpec, "Base resource configuration to modify with flags (empty|default)")
	createAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createAlertCobraHelper.AddCRSpecFlagsToCommand(createAlertCmd, true)
	addMockFlag(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)

	createAlertCobraHelper.AddCRSpecFlagsToCommand(createAlertNativeCmd, true)
	addNativeFormatFlag(createAlertNativeCmd)
	createAlertNativeCmd.Flags().BoolVar(&alertNativePVC, "output-pvc", alertNativePVC, "If true, output resources for only Alert's persistent volume claims")
	createAlertCmd.AddCommand(createAlertNativeCmd)

	// Add Black Duck Command
	createBlackDuckCmd.PersistentFlags().StringVar(&baseBlackDuckSpec, "template", baseBlackDuckSpec, "Base resource configuration to modify with flags (empty|persistentStorageLatest|persistentStorageV1|externalPersistentStorageLatest|externalPersistentStorageV1|bdba|ephemeral|ephemeralCustomAuthCA|externalDB|IPV6Disabled)")
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
	createOpsSightCmd.PersistentFlags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags (empty|upstream|default|disabledBlackDuck)")
	createOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightCmd, true)
	addMockFlag(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightNativeCmd, true)
	addNativeFormatFlag(createOpsSightNativeCmd)
	createOpsSightCmd.AddCommand(createOpsSightNativeCmd)

}
