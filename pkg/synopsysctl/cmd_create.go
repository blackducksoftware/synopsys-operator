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
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Command Resource Ctls
var createAlertCtl ResourceCtl
var createBlackDuckCtl ResourceCtl
var createOpsSightCtl ResourceCtl

// Default Base Specs for Create
var baseAlertSpec string
var baseBlackDuckSpec string
var baseOpsSightSpec string

var namespace string
var blackDuckNativeDatabase bool
var blackDuckNativePVC bool

const (
	blackDuckStageDatabase = "DATABASE"
	blackDuckStagePVC      = "PVC"
)

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
	// Check the user's flags
	err := createAlertCtl.CheckSpecFlags(cmd.Flags())
	if err != nil {
		cmd.Help()
		return err
	}
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseAlertSpec = defaultBaseAlertSpec
	}
	log.Debugf("setting Alert's base spec to '%s'", baseAlertSpec)
	err = createAlertCtl.SwitchSpec(baseAlertSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateAlertSpecWithFlags(cmd *cobra.Command, alertName string, alertNamespace string) *alertv1.Alert {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	createAlertCtl.SetChangedFlags(cmd.Flags())

	// Set Namespace in Spec
	alertSpec, _ := createAlertCtl.GetSpec().(alertv1.AlertSpec)
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
	return alert
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
		alertName, alertNamespace, scope, err := getInstanceInfo(mockMode, args[0], util.AlertCRDName, "", namespace)
		if err != nil {
			return err
		}
		alert := updateAlertSpecWithFlags(cmd, alertName, alertNamespace)

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
			return PrintResource(*alert, mockFormat, false)
		}

		log.Infof("creating Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		// Check if Synopsys Operator is running
		if err := checkOperatorIsRunning(scope, alertNamespace); err != nil {
			return err
		}
		// Create namespace for an Alert instance
		err = util.DeployCRDNamespace(restconfig, kubeClient, util.AlertName, alertNamespace, alertName, alert.Spec.Version)
		if err != nil {
			log.Warn(err)
		}
		// Deploy the Alert instance
		_, err = alertClient.SynopsysV1().Alerts(alertNamespace).Create(alert)
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
		alertName, alertNamespace, _, err := getInstanceInfo(true, args[0], util.AlertCRDName, "", namespace)
		if err != nil {
			return err
		}
		alert := updateAlertSpecWithFlags(cmd, alertName, alertNamespace)

		log.Debugf("generating Kubernetes resources for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		return PrintResource(*alert, nativeFormat, true)
	},
}

/*
Create Black Duck Commands
*/

var createBlackDuckPreRun = func(cmd *cobra.Command, args []string) error {
	// Check the user's flags
	err := createBlackDuckCtl.CheckSpecFlags(cmd.Flags())
	if err != nil {
		cmd.Help()
		return err
	}
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseBlackDuckSpec = defaultBaseBlackDuckSpec
	}
	log.Debugf("setting Black Duck's base spec to '%s'", baseBlackDuckSpec)
	err = createBlackDuckCtl.SwitchSpec(baseBlackDuckSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateBlackDuckSpecWithFlags(cmd *cobra.Command, blackDuckName string, blackDuckNamespace string) *blackduckv1.Blackduck {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	createBlackDuckCtl.SetChangedFlags(cmd.Flags())

	// Set Namespace in Spec
	blackDuckSpec, _ := createBlackDuckCtl.GetSpec().(blackduckv1.BlackduckSpec)
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
	return blackDuck
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
		return nil
	},
	PreRunE: createBlackDuckPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		blackDuckName, blackDuckNamespace, scope, err := getInstanceInfo(mockMode, args[0], util.BlackDuckCRDName, "", namespace)
		if err != nil {
			return err
		}
		blackDuck := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			return PrintResource(*blackDuck, mockFormat, false)
		}

		log.Infof("creating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		// Check if Synopsys Operator is running
		if err := checkOperatorIsRunning(scope, blackDuckNamespace); err != nil {
			return err
		}
		// Verifying only one Black Duck instance per namespace
		blackducks, err := util.ListHubs(blackDuckClient, blackDuckNamespace)
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances in namespace '%s' due to %+v", blackDuckNamespace, err)
		}
		for _, v := range blackducks.Items {
			if strings.EqualFold(v.Spec.Namespace, blackDuckNamespace) {
				return fmt.Errorf("due to issues with this version of Black Duck, only one instance per namespace is allowed")
			}
		}
		// Create namespace for the Black Duck instance
		err = util.DeployCRDNamespace(restconfig, kubeClient, util.BlackDuckName, blackDuckNamespace, blackDuckName, blackDuck.Spec.Version)
		if err != nil {
			log.Warn(err)
		}
		// Deploy the Black Duck instance
		log.Debugf("deploying Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		_, err = blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Create(blackDuck)
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
		return nil
	},
	PreRunE: createBlackDuckPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(true, args[0], util.BlackDuckCRDName, "", namespace)
		if err != nil {
			return err
		}
		blackDuck := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)

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
			cList, err = app.Blackduck().GetComponents(blackDuck, blackDuckStageDatabase)
		case blackDuckNativePVC:
			cList, err = app.Blackduck().GetComponents(blackDuck, blackDuckStagePVC)
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
	// Check the user's flags
	err := createOpsSightCtl.CheckSpecFlags(cmd.Flags())
	if err != nil {
		cmd.Help()
		return err
	}
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseOpsSightSpec = defaultBaseOpsSightSpec
	}
	log.Debugf("setting OpsSight's base spec to '%s'", baseOpsSightSpec)
	err = createOpsSightCtl.SwitchSpec(baseOpsSightSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateOpsSightSpecWithFlags(cmd *cobra.Command, opsSightName string, opsSightNamespace string) *opssightv1.OpsSight {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	createOpsSightCtl.SetChangedFlags(cmd.Flags())

	// Set Namespace in Spec
	opsSightSpec, _ := createOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
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
	return opsSight
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
		opsSightName, opsSightNamespace, scope, err := getInstanceInfo(mockMode, args[0], util.OpsSightCRDName, "", namespace)
		if err != nil {
			return err
		}
		opsSight := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*opsSight, mockFormat, false)
		}

		log.Infof("creating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		// Check if Synopsys Operator is running
		if err := checkOperatorIsRunning(scope, opsSightNamespace); err != nil {
			return err
		}
		// Create namespace for OpsSight
		// TODO: when opssight versioning PR is merged, the hard coded 2.2.3 version to be replaced with opsSight
		err = util.DeployCRDNamespace(restconfig, kubeClient, util.OpsSightName, opsSightNamespace, opsSightName, "2.2.3")
		if err != nil {
			log.Warnf("%s", err)
		}
		// Deploy the OpsSight instance
		_, err = opsSightClient.SynopsysV1().OpsSights(opsSightNamespace).Create(opsSight)
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
	Aliases:       []string{"ops"},
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
		opsSightName, opsSightNamespace, _, err := getInstanceInfo(true, args[0], util.OpsSightCRDName, "", namespace)
		if err != nil {
			return err
		}
		opsSight := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)

		log.Debugf("generating Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*opsSight, nativeFormat, true)
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCtl = alert.NewAlertCtl()
	createBlackDuckCtl = blackduck.NewBlackDuckCtl()
	createOpsSightCtl = opssight.NewOpsSightCtl()

	rootCmd.AddCommand(createCmd)

	// Add Alert Command
	createAlertCmd.PersistentFlags().StringVar(&baseAlertSpec, "template", baseAlertSpec, "Base resource configuration to modify with flags [empty|default]")
	createAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createAlertCtl.AddSpecFlags(createAlertCmd, true)
	addMockFlag(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)

	createAlertCtl.AddSpecFlags(createAlertNativeCmd, true)
	addNativeFormatFlag(createAlertNativeCmd)
	createAlertCmd.AddCommand(createAlertNativeCmd)

	// Add Black Duck Command
	createBlackDuckCmd.PersistentFlags().StringVar(&baseBlackDuckSpec, "template", baseBlackDuckSpec, "Base resource configuration to modify with flags [empty|persistentStorageLatest|persistentStorageV1|externalPersistentStorageLatest|externalPersistentStorageV1|bdba|ephemeral|ephemeralCustomAuthCA|externalDB|IPV6Disabled]")
	createBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createBlackDuckCtl.AddSpecFlags(createBlackDuckCmd, true)
	addMockFlag(createBlackDuckCmd)
	createCmd.AddCommand(createBlackDuckCmd)

	createBlackDuckCtl.AddSpecFlags(createBlackDuckNativeCmd, true)
	addNativeFormatFlag(createBlackDuckNativeCmd)
	createBlackDuckNativeCmd.Flags().BoolVar(&blackDuckNativeDatabase, "output-database", blackDuckNativeDatabase, "If true, output resources for only Black Duck's database")
	createBlackDuckNativeCmd.Flags().BoolVar(&blackDuckNativePVC, "output-pvc", blackDuckNativePVC, "If true, output resources for only Black Duck's persistent volume claims")
	createBlackDuckCmd.AddCommand(createBlackDuckNativeCmd)

	// Add OpsSight Command
	createOpsSightCmd.PersistentFlags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags [empty|upstream|default|disabledBlackDuck]")
	createOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createOpsSightCtl.AddSpecFlags(createOpsSightCmd, true)
	addMockFlag(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	createOpsSightCtl.AddSpecFlags(createOpsSightNativeCmd, true)
	addNativeFormatFlag(createOpsSightNativeCmd)
	createOpsSightCmd.AddCommand(createOpsSightNativeCmd)

}
