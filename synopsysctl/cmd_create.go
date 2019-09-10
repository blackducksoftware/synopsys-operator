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
	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	"github.com/blackducksoftware/synopsys-operator/controllers"
	"github.com/blackducksoftware/synopsys-operator/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Command CRSpecBuilderFromCobraFlagsInterface
var createAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createBlackDuckCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createPolarisCobraHelper CRSpecBuilderFromCobraFlagsInterface

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

func updateAlertSpecWithFlags(cmd *cobra.Command, alertName string, alertNamespace string) (*synopsysv1.Alert, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	alertInterface, err := createAlertCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	alertSpec, _ := alertInterface.(synopsysv1.AlertSpec)
	alertSpec.Namespace = alertNamespace

	// Create Alert CRD
	alert := &synopsysv1.Alert{
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

func updatePolarisSpecWithFlags(cmd *cobra.Command, name string, namespace string) (*polarisMultiCR, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	polaris, err := createPolarisCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	multiCRSpec, ok := polaris.(polarisMultiCRSpec)
	if !ok {
		panic("Error")
	}
	multiCRSpec.polarisSpec.Namespace = namespace
	multiCRSpec.polarisDBSpec.Namespace = namespace
	multiCRSpec.authSpec.Namespace = namespace

	polarisMultiCR := &polarisMultiCR{
		auth: &synopsysv1.AuthServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "synopsys.com/v1",
				Kind:       "AuthServer",
			},
			Spec: *multiCRSpec.authSpec,
		},
		polarisDB: &synopsysv1.PolarisDB{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "synopsys.com/v1",
				Kind:       "PolarisDB",
			},
			Spec: *multiCRSpec.polarisDBSpec,
		},
		polaris: &synopsysv1.Polaris{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "synopsys.com/v1",
				Kind:       "Polaris",
			},
			Spec: *multiCRSpec.polarisSpec,
		},
	}

	return polarisMultiCR, nil
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
		alertName, alertNamespace, _, err := getInstanceInfo(mockMode, utils.AlertCRDName, "", namespace, args[0])
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
			//return PrintResource(*alert, mockFormat, false)
		}

		log.Infof("creating Alert '%s' in namespace '%s'...", alertName, alertNamespace)

		// Deploy the Alert instance
		_, err = utils.CreateAlert(restClient, alert)
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
		alertName, alertNamespace, _, err := getInstanceInfo(true, utils.AlertCRDName, "", namespace, args[0])
		if err != nil {
			return err
		}
		alert, err := updateAlertSpecWithFlags(cmd, alertName, alertNamespace)
		if err != nil {
			return err
		}

		reconciler := controllers.AlertReconciler{
			IsOpenShift: nativeClusterType == clusterTypeOpenshift,
			IsDryRun:    true,
		}

		objectsMap, err := reconciler.GetRuntimeObjects(alert)
		if err != nil {
			return err
		}

		var objectArr []runtime.Object

		// TODO PVC / DB filtering
		for _, v := range objectsMap {
			objectArr = append(objectArr, v)
		}

		switch {
		case !blackDuckNativeDatabase && !blackDuckNativePVC:
			if objectArr, err = filterByLabel("component notin (postgres, pvc)", objectArr); err != nil {
				return err
			}
		case blackDuckNativeDatabase:
			if objectArr, err = filterByLabel("component in (postgres)", objectArr); err != nil {
				return err
			}
		case blackDuckNativePVC:
			if objectArr, err = filterByLabel("component in (pvc)", objectArr); err != nil {
				return err
			}
		}

		for _, obj := range objectArr {
			PrintComponent(obj, nativeFormat)
		}

		return nil
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

func updateBlackDuckSpecWithFlags(cmd *cobra.Command, blackDuckName string, blackDuckNamespace string) (*synopsysv1.Blackduck, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	blackDuckInterface, err := createBlackDuckCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	blackDuckSpec, _ := blackDuckInterface.(synopsysv1.BlackduckSpec)
	blackDuckSpec.Namespace = blackDuckNamespace

	// Create and Deploy Black Duck CRD
	blackDuck := &synopsysv1.Blackduck{
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
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(mockMode, utils.BlackDuckCRDName, "", namespace, args[0])
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
			//return PrintResource(*blackDuck, mockFormat, false)
		}

		log.Infof("creating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

		// Verifying only one Black Duck instance per namespace
		blackducks, err := utils.ListBlackduck(restClient, blackDuckName, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances in namespace '%s' due to %+v", blackDuckNamespace, err)
		}
		for _, v := range blackducks.Items {
			if strings.EqualFold(v.Spec.Namespace, blackDuckNamespace) {
				return fmt.Errorf("a Black Duck instance already exists in namespace '%s', only one instance per namespace is allowed", blackDuckNamespace)
			}
		}

		// Deploy the Black Duck instance
		log.Debugf("deploying Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		_, err = utils.CreateBlackduck(restClient, blackDuck)
		if err != nil {
			return fmt.Errorf("error creating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		log.Infof("successfully submitted Black Duck '%s' into namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

//createBlackDuckNativeCmd prints the Kubernetes resources for creating a Black Duck instance
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
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(true, utils.BlackDuckCRDName, "", namespace, args[0])
		if err != nil {
			return err
		}
		blackDuck, err := updateBlackDuckSpecWithFlags(cmd, blackDuckName, blackDuckNamespace)
		if err != nil {
			return err
		}

		blackDuck.Spec.LivenessProbes = true // enable LivenessProbes when generating Kubernetes resources for customers
		log.Debugf("generating Kubernetes resources for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

		reconciler := controllers.BlackduckReconciler{
			IsDryRun:    true,
			IsOpenShift: nativeClusterType == clusterTypeOpenshift,
		}

		objectsMap, err := reconciler.GetRuntimeObjects(blackDuck)
		if err != nil {
			return err
		}

		var objectArr []runtime.Object

		for _, v := range objectsMap {
			objectArr = append(objectArr, v)
		}

		switch {
		case !blackDuckNativeDatabase && !blackDuckNativePVC:
			if objectArr, err = filterByLabel("component notin (postgres, pvc)", objectArr); err != nil {
				return err
			}
		case blackDuckNativeDatabase:
			if objectArr, err = filterByLabel("component in (postgres)", objectArr); err != nil {
				return err
			}
		case blackDuckNativePVC:
			if objectArr, err = filterByLabel("component in (pvc)", objectArr); err != nil {
				return err
			}
		}

		PrintComponents(objectArr, nativeFormat)

		return nil
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

func updateOpsSightSpecWithFlags(cmd *cobra.Command, opsSightName string, opsSightNamespace string) (*synopsysv1.OpsSight, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	opsSightInterface, err := createOpsSightCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	opsSightSpec, _ := opsSightInterface.(synopsysv1.OpsSightSpec)
	opsSightSpec.Namespace = opsSightNamespace

	// Create and Deploy OpsSight CRD
	opsSight := &synopsysv1.OpsSight{
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
		opsSightName, opsSightNamespace, _, err := getInstanceInfo(mockMode, utils.OpsSightCRDName, "", namespace, args[0])
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
			//return PrintResource(*opsSight, mockFormat, false)
		}

		log.Infof("creating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

		// Deploy the OpsSight instance
		_, err = utils.CreateOpsSight(restClient, opsSight)
		if err != nil {
			return fmt.Errorf("error creating the OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		log.Infof("successfully submitted OpsSight '%s' into namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// createOpsSightNativeCmd prints the Kubernetes resources for creating an OpsSight instance
//var createOpsSightNativeCmd = &cobra.Command{
//	Use:           "native NAME",
//	Example:       "synopsysctl create opssight native <name>\nsynopsysctl create opssight native <name> -o yaml",
//	Short:         "Print the Kubernetes resources for creating an OpsSight instance",
//	SilenceUsage:  true,
//	SilenceErrors: true,
//	Args: func(cmd *cobra.Command, args []string) error {
//		// Check the Number of Arguments
//		if len(args) != 1 {
//			cmd.Help()
//			return fmt.Errorf("this command takes 1 argument")
//		}
//		checkRegistryConfiguration(cmd.Flags())
//		return nil
//	},
//	PreRunE: createOpsSightPreRun,
//	RunE: func(cmd *cobra.Command, args []string) error {
//		opsSightName, opsSightNamespace, _, err := getInstanceInfo(true, util.OpsSightCRDName, "", namespace, args[0])
//		if err != nil {
//			return err
//		}
//		opsSight, err := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)
//		if err != nil {
//			return err
//		}
//
//		log.Debugf("generating Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
//		return PrintResource(*opsSight, nativeFormat, true)
//	},
//}

// createCmd creates a Polaris instance
var createPolarisCmd = &cobra.Command{
	Use:           "polaris NAME",
	Example:       "synopsysctl create polaris <name>\nsynopsysctl create polaris <name> -n <namespace>\nsynopsysctl create polaris <name> --mock json",
	Short:         "Create a Polaris instance",
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := createPolarisCobraHelper.SetPredefinedCRSpec("")
		if err != nil {
			cmd.Help()
			return err
		}
		cobra.MarkFlagRequired(cmd.Flags(), "version")
		cobra.MarkFlagRequired(cmd.Flags(), "environment-dns")
		cobra.MarkFlagRequired(cmd.Flags(), "environment-name")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-username")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-password")

		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed

		fmt.Println(namespace)
		polarisName, polarisNamespace, _, err := getInstanceInfo(mockMode, utils.PolarisCRDName, "", namespace, args[0])
		if err != nil {
			return err
		}
		polaris, err := updatePolarisSpecWithFlags(cmd, polarisName, polarisNamespace)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for Polaris '%s' in namespace '%s'...", polarisName, polarisNamespace)
			//return PrintResource(*alert, mockFormat, false)
		}

		log.Infof("creating Polaris '%s' in namespace '%s'...", polarisName, polarisNamespace)

		if _, err := utils.CreateAuthServer(restClient, polaris.auth); err != nil {
			return err
		}
		if _, err := utils.CreatePolarisDB(restClient, polaris.polarisDB); err != nil {
			return err
		}
		if _, err := utils.CreatePolaris(restClient, polaris.polaris); err != nil {
			return err
		}

		log.Infof("successfully submitted Polaris '%s' into namespace '%s'", polarisName, polarisNamespace)
		return nil
	},
}

//createPolarisNativeCmd prints the Kubernetes resources for creating a Polaris instance
var createPolarisNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create polaris native <name>",
	Short:         "Print the Kubernetes resources for creating a Polaris instance",
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := createPolarisCobraHelper.SetPredefinedCRSpec("")
		if err != nil {
			cmd.Help()
			return err
		}
		cobra.MarkFlagRequired(cmd.Flags(), "version")
		cobra.MarkFlagRequired(cmd.Flags(), "environment-dns")
		cobra.MarkFlagRequired(cmd.Flags(), "environment-name")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-username")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-password")

		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		polarisName, polarisNamespace, _, err := getInstanceInfo(true, utils.PolarisName, "", namespace, args[0])
		if err != nil {
			return err
		}
		polaris, err := updatePolarisSpecWithFlags(cmd, polarisName, polarisNamespace)
		if err != nil {
			return err
		}

		// Get runtime objects
		polarisReconciler := controllers.PolarisReconciler{
			IsOpenShift: nativeClusterType == clusterTypeOpenshift,
			IsDryRun:    true,
		}
		polarisObjectsMap, err := polarisReconciler.GetRuntimeObjects(polaris.polaris)
		if err != nil {
			return err
		}

		polarisDBReconciler := controllers.PolarisDBReconciler{
			IsOpenShift: nativeClusterType == clusterTypeOpenshift,
			IsDryRun:    true,
		}
		polarisDBObjectsMap, err := polarisDBReconciler.GetRuntimeObjects(polaris.polarisDB)
		if err != nil {
			return err
		}

		authReconciler := controllers.AuthServerReconciler{
			IsOpenShift: nativeClusterType == clusterTypeOpenshift,
			IsDryRun:    true,
		}
		authObjectsMap, err := authReconciler.GetRuntimeObjects(polaris.auth)
		if err != nil {
			return err
		}

		var objectArr []runtime.Object
		for _, v := range polarisObjectsMap {
			objectArr = append(objectArr, v)
		}
		for _, v := range polarisDBObjectsMap {
			objectArr = append(objectArr, v)
		}
		for _, v := range authObjectsMap {
			objectArr = append(objectArr, v)
		}

		PrintComponents(objectArr, nativeFormat)

		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCobraHelper = NewAlertCRSpecBuilderFromCobraFlags()
	createBlackDuckCobraHelper = NewBlackduckCRSpecBuilderFromCobraFlags()
	//createOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	createPolarisCobraHelper = NewPolarisCRSpecBuilderFromCobraFlags()

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
	//
	//// Add OpsSight Command
	//createOpsSightCmd.PersistentFlags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags (empty|upstream|default|disabledBlackDuck)")
	//createOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	//createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightCmd, true)
	//addMockFlag(createOpsSightCmd)
	//createCmd.AddCommand(createOpsSightCmd)
	//
	//createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightNativeCmd, true)
	//addNativeFormatFlag(createOpsSightNativeCmd)
	//createOpsSightCmd.AddCommand(createOpsSightNativeCmd)

	createPolarisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createPolarisCobraHelper.AddCRSpecFlagsToCommand(createPolarisCmd, true)
	addMockFlag(createPolarisCmd)
	createCmd.AddCommand(createPolarisCmd)

	createPolarisCobraHelper.AddCRSpecFlagsToCommand(createPolarisNativeCmd, true)
	addNativeFormatFlag(createPolarisNativeCmd)
	createPolarisCmd.AddCommand(createPolarisNativeCmd)
}
