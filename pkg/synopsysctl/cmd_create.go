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

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

// Flags for using mock mode - doesn't deploy
var createMockFormat string
var createMockKubeFormat string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

var namespace string

// createCmd represents the create command for Alert
var createAlertCmd = &cobra.Command{
	Use:     "alert NAME",
	Example: "synopsysctl create alert name\nsynopsysctl create alert name -n <namespace>\nsynopsysctl create alert <name> --mock json\nsynopsysctl create alert <name> -n <namespace> --mock json",
	Short:   "Create an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		// Check the Arguments
		err := createAlertCtl.CheckSpecFlags(cmd.Flags())
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Check/Set the Spec Type
		if !cmd.Flags().Lookup("template").Changed {
			baseAlertSpec = defaultBaseAlertSpec
		}
		log.Debugf("setting Alert template spec %s", baseAlertSpec)
		err = createAlertCtl.SwitchSpec(baseAlertSpec)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName, alertNamespace, crdScope, err := getInstanceInfo(cmd, args, util.AlertCRDName, namespace)
		if err != nil {
			return err
		}

		log.Infof("creating Alert '%s' instance in '%s' namespace...", alertName, alertNamespace)

		// Update Spec with user's flags
		log.Debugf("updating Spec with User's Flags")
		createAlertCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		alertSpec, _ := createAlertCtl.GetSpec().(alertv1.AlertSpec)
		alertSpec.Namespace = alertNamespace

		// Create and Deploy Alert CRD
		alert := &alertv1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name:      alertName,
				Namespace: alertNamespace,
			},
			Spec: alertSpec,
		}
		alert.Kind = "Alert"
		alert.APIVersion = "synopsys.com/v1"
		if cmd.LocalFlags().Lookup("mock").Changed {
			log.Debugf("running mock mode")
			err := PrintResource(*alert, createMockFormat, false)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
			log.Debugf("running kube mock mode")
			err := PrintResource(*alert, createMockKubeFormat, true)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else {
			// Create namespace for Alert
			var err error
			if crdScope == apiextensions.ClusterScoped {
				err = DeployCRDNamespace(restconfig, alertName)
			} else {
				err = DeployCRDNamespace(restconfig, alertNamespace)
			}
			if err != nil {
				log.Warn(err)
			}
			// Create Alert with Client
			log.Debugf("deploying Alert '%s' instance in '%s' namespace", alertName, alertNamespace)
			_, err = alertClient.SynopsysV1().Alerts(alertNamespace).Create(alert)
			if err != nil {
				log.Errorf("error creating the Alert %s instance in %s namespace due to %+v", alertName, alertNamespace, err)
				return nil
			}
			log.Infof("successfully created Alert '%s' instance in '%s namespace", alertName, alertNamespace)
		}
		return nil
	},
}

// createBlackDuckCmd represents the create command for Black Duck
var createBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME",
	Example: "synopsysctl create blackduck <name>\nsynopsysctl create blackduck <name> -n <namespace>\nsynopsysctl create blackduck name --mock json\nsynopsysctl create blackduck name -n <namespace> --mock json",
	Short:   "Create an instance of a Black Duck",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		// Check the Arguments
		err := createBlackDuckCtl.CheckSpecFlags(cmd.Flags())
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Set the Spec Type
		if !cmd.Flags().Lookup("template").Changed {
			baseBlackDuckSpec = defaultBaseBlackDuckSpec
		}
		log.Debugf("setting template spec %s", baseBlackDuckSpec)
		err = createBlackDuckCtl.SwitchSpec(baseBlackDuckSpec)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, crdScope, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			return err
		}
		log.Infof("creating Black Duck '%s' instance in '%s' namespace...", blackDuckName, blackDuckNamespace)

		blackducks, err := util.ListHubs(blackDuckClient, blackDuckNamespace)
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances in %s namespace due to %+v", blackDuckNamespace, err)
		}

		if len(blackducks.Items) > 0 {
			return fmt.Errorf("due to known restriction, only one Black Duck instance per namespace is allowed. stay tuned for the updates")
		}

		// Update Spec with user's flags
		log.Debugf("updating Spec with User's Flags")
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
		if cmd.LocalFlags().Lookup("mock").Changed {
			log.Debugf("running mock mode")
			err := PrintResource(*blackDuck, createMockFormat, false)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
			log.Debugf("running kube mock mode")
			err := PrintResource(*blackDuck, createMockKubeFormat, true)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else {
			// Create namespace for Black Duck
			var err error
			if crdScope == apiextensions.ClusterScoped {
				err = DeployCRDNamespace(restconfig, blackDuckName)
			} else {
				err = DeployCRDNamespace(restconfig, blackDuckNamespace)
			}
			if err != nil {
				log.Warn(err)
			}

			// Create Black Duck with Client
			log.Debugf("deploying Black Duck '%s' instance in '%s' namespace", blackDuckName, blackDuckNamespace)
			_, err = blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Create(blackDuck)
			if err != nil {
				log.Errorf("error creating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
			log.Infof("successfully created Black Duck '%s' instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		}
		return nil
	},
}

// createCmd represents the create command for OpsSight
var createOpsSightCmd = &cobra.Command{
	Use:     "opssight NAME",
	Short:   "Create an instance of OpsSight",
	Example: "synopsysctl create opssight <name>\nsynopsysctl create opssight <name> --mock json",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		// Check the Arguments
		err := createOpsSightCtl.CheckSpecFlags(cmd.Flags())
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		// Set the Spec Type
		if !cmd.Flags().Lookup("template").Changed {
			baseOpsSightSpec = defaultBaseOpsSightSpec
		}
		log.Debugf("setting OpsSight template spec %s", baseOpsSightSpec)
		err = createOpsSightCtl.SwitchSpec(baseOpsSightSpec)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		log.Infof("creating OpsSight %s instance...", opsSightNamespace)

		// Update Spec with user's flags
		log.Debugf("updating Spec with User's Flags")
		createOpsSightCtl.SetChangedFlags(cmd.Flags())

		// Set Namespace in Spec
		opssightSpec, _ := createOpsSightCtl.GetSpec().(opssightv1.OpsSightSpec)
		opssightSpec.Namespace = opsSightNamespace

		// Create and Deploy OpsSight CRD
		opsSight := &opssightv1.OpsSight{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opsSightNamespace,
				Namespace: opsSightNamespace,
			},
			Spec: opssightSpec,
		}
		opsSight.Kind = "OpsSight"
		opsSight.APIVersion = "synopsys.com/v1"
		if cmd.LocalFlags().Lookup("mock").Changed {
			log.Debugf("running mock mode")
			err := PrintResource(*opsSight, createMockFormat, false)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
			log.Debugf("running kube mock mode")
			err := PrintResource(*opsSight, createMockKubeFormat, true)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else {
			// Create namespace for OpsSight
			err := DeployCRDNamespace(restconfig, opsSightNamespace)
			if err != nil {
				log.Errorf("%s", err)
			}
			// Create OpsSight with Client
			log.Debugf("deploying OpsSight in namespace %s", opsSightNamespace)
			_, err = opsSightClient.SynopsysV1().OpsSights(opsSightNamespace).Create(opsSight)
			if err != nil {
				log.Errorf("error creating the %s OpsSight instance due to %+v", opsSightNamespace, err)
				return nil
			}
			log.Infof("successfully created OpsSight '%s' instance", opsSightNamespace)
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createAlertCtl = alert.NewAlertCtl()
	createBlackDuckCtl = blackduck.NewBlackDuckCtl()
	createOpsSightCtl = opssight.NewOpsSightCtl()

	//(PassCmd) createCmd.DisableFlagParsing = true // lets createCmd pass flags to kube/oc
	rootCmd.AddCommand(createCmd)

	// Add Alert Command
	createAlertCmd.Flags().StringVar(&baseAlertSpec, "template", baseAlertSpec, "Base resource configuration to modify with flags [empty|default]")
	createAlertCmd.Flags().StringVarP(&createMockFormat, "mock", "o", createMockFormat, "Prints the resource spec in the specified format instead of creating it [json|yaml]")
	createAlertCmd.Flags().StringVarP(&createMockKubeFormat, "mock-kube", "k", createMockKubeFormat, "Prints the Kubernetes resource specs in the specified format instead of creating it [json|yaml]")
	createAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to create the resource(s)")
	createAlertCtl.AddSpecFlags(createAlertCmd, true)
	createCmd.AddCommand(createAlertCmd)

	// Add Black Duck Command
	createBlackDuckCmd.Flags().StringVar(&baseBlackDuckSpec, "template", baseBlackDuckSpec, "Base resource configuration to modify with flags [empty|persistentStorageLatest|persistentStorageV1|externalPersistentStorageLatest|externalPersistentStorageV1|bdba|ephemeral|ephemeralCustomAuthCA|externalDB|IPV6Disabled]")
	createBlackDuckCmd.Flags().StringVarP(&createMockFormat, "mock", "o", createMockFormat, "Prints the CRD resource spec in the specified format instead of creating it [json|yaml]")
	createBlackDuckCmd.Flags().StringVarP(&createMockKubeFormat, "mock-kube", "k", createMockKubeFormat, "Prints the Kubernetes resource specs in the specified format instead of creating it [json|yaml]")
	createBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to create the resource(s)")
	createBlackDuckCtl.AddSpecFlags(createBlackDuckCmd, true)
	createCmd.AddCommand(createBlackDuckCmd)

	// Add OpsSight Command
	createOpsSightCmd.Flags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags [empty|upstream|default|disabledBlackDuck]")
	createOpsSightCmd.Flags().StringVarP(&createMockFormat, "mock", "o", createMockFormat, "Prints the resource spec in the specified format instead of creating it [json|yaml]")
	createOpsSightCmd.Flags().StringVarP(&createMockKubeFormat, "mock-kube", "k", createMockKubeFormat, "Prints the Kubernetes resource specs in the specified format instead of creating it [json|yaml]")
	createOpsSightCtl.AddSpecFlags(createOpsSightCmd, true)
	createCmd.AddCommand(createOpsSightCmd)
}
