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
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"

	"fmt"

	"sort"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	v1 "k8s.io/api/core/v1"

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

	return blackDuck, nil
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
		if len(blackDuck.Spec.Version) == 0 {
			versions := apps.NewApp(&protoform.Config{}, restconfig).Blackduck().Versions()
			sort.Sort(sort.Reverse(sort.StringSlice(versions)))
			blackDuck.Spec.Version = versions[0]
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

// createCmd creates a Polaris instance
var createPolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsysctl create polaris -n <namespace>",
	Short:         "Create a Polaris instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument")
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
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-username")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-password")

		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-sender-email")

		cobra.MarkFlagRequired(cmd.Flags(), "organization-description")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-name")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-name")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-username")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-email")
		cobra.MarkFlagRequired(cmd.Flags(), "coverity-license-path")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		polarisObj, err := updatePolarisSpecWithFlags(cmd, namespace)
		if err != nil {
			return err
		}

		type deploy struct {
			name string
			obj  map[string]runtime.Object
		}

		var deployments []deploy

		dbComponents, err := polaris.GetPolarisDBComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}
		deployments = append(deployments, deploy{name: "Polaris DB", obj: dbComponents})

		polarisComponents, err := polaris.GetPolarisComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}
		deployments = append(deployments, deploy{name: "Polaris Core", obj: polarisComponents})

		if polarisObj.EnableReporting {
			reportingComponents, err := polaris.GetPolarisReportingComponents(baseURL, *polarisObj)
			if err != nil {
				return err
			}
			deployments = append(deployments, deploy{name: "Polaris Reporting", obj: reportingComponents})
		}

		provisionComponents, err := polaris.GetPolarisProvisionComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}

		log.Infof("provisionComponents: %+v", provisionComponents)
		deployments = append(deployments, deploy{name: "Polaris Organization Provision", obj: provisionComponents})

		// Marshal Polaris
		polarisByte, err := json.Marshal(polarisObj)
		if err != nil {
			return err
		}

		polarisSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "polaris",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"polaris": polarisByte,
			},
		}
		// Create secret
		polarisSecret, err = kubeClient.CoreV1().Secrets(namespace).Create(polarisSecret)
		if err != nil {
			return err
		}

		log.Info("Deploying Polaris")

		for _, v := range deployments {
			log.Infof("Deploying %s", v.name)
			var content []byte
			for _, v := range v.obj {
				polarisComponentsByte, err := json.Marshal(v)
				if err != nil {
					return err
				}
				content = append(content, polarisComponentsByte...)
			}
			out, err := RunKubeCmdWithStdin(restconfig, kubeClient, string(content), "apply", "--validate=false", "-f", "-")
			if err != nil {
				kubeClient.CoreV1().Secrets(namespace).Delete("polaris", &metav1.DeleteOptions{})

				return fmt.Errorf("couldn't deploy polaris |  %+v - %s", out, err)
			}
		}

		log.Info("Polaris has been successfully deployed!")
		return nil
	},
}

//createPolarisNativeCmd prints the Kubernetes resources for creating a Polaris instance
var createPolarisNativeCmd = &cobra.Command{
	Use:           "native",
	Example:       "synopsysctl create polaris native",
	Short:         "Print the Kubernetes resources for creating a Polaris instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments")
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
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-username")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-password")

		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-sender-email")

		cobra.MarkFlagRequired(cmd.Flags(), "organization-description")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-name")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-name")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-username")
		cobra.MarkFlagRequired(cmd.Flags(), "organization-admin-email")
		cobra.MarkFlagRequired(cmd.Flags(), "coverity-license-path")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		polarisObj, err := updatePolarisSpecWithFlags(cmd, namespace)
		if err != nil {
			return err
		}

		components, err := polaris.GetComponents(baseURL, *polarisObj)
		if err != nil {
			return err
		}

		var objectArr []interface{}
		for _, v := range components {
			objectArr = append(objectArr, v)
		}

		PrintComponents(objectArr, nativeFormat)

		return nil
	},
}

func updatePolarisSpecWithFlags(cmd *cobra.Command, namespace string) (*polaris.Polaris, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	polarisInterface, err := createPolarisCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	polarisSpec, ok := polarisInterface.(polaris.Polaris)
	if !ok {
		panic("Couldn't cast polarisInterface to polarisSpec")
	}
	polarisSpec.Namespace = namespace

	return &polarisSpec, nil
}

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()
	createBlackDuckCobraHelper = blackduck.NewCRSpecBuilderFromCobraFlags()
	createOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	createPolarisCobraHelper = polaris.NewCRSpecBuilderFromCobraFlags()

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
	createPolarisCobraHelper.AddCRSpecFlagsToCommand(createPolarisCmd, true)
	addbaseURLFlag(createPolarisCmd)
	createCmd.AddCommand(createPolarisCmd)

	createPolarisCobraHelper.AddCRSpecFlagsToCommand(createPolarisNativeCmd, true)
	addNativeFormatFlag(createPolarisNativeCmd)
	addbaseURLFlag(createPolarisNativeCmd)
	createPolarisCmd.AddCommand(createPolarisNativeCmd)

}
