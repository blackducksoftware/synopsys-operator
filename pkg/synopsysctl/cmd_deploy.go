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
	"crypto/x509/pkix"
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//  Deploy Command Defaults
var operatorNamespace = DefaultOperatorNamespace
var exposeUI = util.NONE
var synopsysOperatorImage string
var metricsImage = DefaultMetricsImage
var exposeMetrics = util.NONE
var terminationGracePeriodSeconds int64 = 180
var dryRun = "false"
var logLevel = "debug"
var threadiness = 5
var postgresRestartInMins int64 = 10
var podWaitTimeoutSeconds int64 = 600
var resyncIntervalInSeconds int64 = 120
var isEnabledAlert bool
var isEnabledBlackDuck bool
var isEnabledOpsSight bool
var isEnabledPrm bool
var isClusterScoped bool
var admissionWebhookListener = false

// getEnabledCrds returns a list of CRDs that are enabled based on synopsysctl's
// global variables
func getEnabledCrds() ([]string, error) {
	crds := []string{}
	if isEnabledAlert {
		crds = append(crds, util.AlertCRDName)
	}
	if isEnabledBlackDuck {
		crds = append(crds, util.BlackDuckCRDName)
	}
	if isEnabledOpsSight && isClusterScoped {
		crds = append(crds, util.OpsSightCRDName)
	} else if isEnabledOpsSight && !isClusterScoped {
		return []string{}, fmt.Errorf("unable to enable OpsSight because Synopsys Operator has namespace scope and OpsSight requires Synopsys Operator with cluster scope (deploy with flag --cluster-scoped)")
	}
	return crds, nil
}

// deployCrds deploys the Custom Resource Definitions for the Synopsys Operator into
// the cluster based on the scope Synopsys Operator
func deployCrds(namespace string, isClusterScoped bool, crdConfigs []*horizoncomponents.CustomResourceDefinition) error {
	// Create a Deployer
	deployer, err := util.NewDeployer(restconfig)
	if err != nil {
		return errors.Annotate(err, "unable to create the deployer object to deploy custom resource definitions")
	}

	// Add CRDs to the Deployer
	for _, crdConfig := range crdConfigs {
		var scope apiextensions.ResourceScope
		if isClusterScoped {
			scope = apiextensions.ClusterScoped
		} else {
			scope = apiextensions.NamespaceScoped
		}

		isExist, err := checkUpdateCustomResource(crdConfig.GetName(), namespace, scope, isClusterScoped)
		if isExist && err != nil {
			return err
		} else if isExist {
			continue
		}

		deployer.Deployer.AddComponent(horizonapi.CRDComponent, crdConfig)
	}
	// Deploy CRDs into the cluster
	err = deployer.Deployer.Run()
	if err != nil {
		return errors.Annotate(err, "unable to deploy custom resource definition")
	}

	return nil
}

func getCrdConfigs(namespace string, isClusterScoped bool, crds []string) ([]*horizoncomponents.CustomResourceDefinition, error) {
	crdConfigs := []*horizoncomponents.CustomResourceDefinition{}

	for _, crd := range crds {
		var crdConfig *horizoncomponents.CustomResourceDefinition
		var crdScope horizonapi.CRDScopeType
		if isClusterScoped {
			crdScope = horizonapi.CRDClusterScoped
		} else {
			crdScope = horizonapi.CRDNamespaceScoped
		}

		switch crd {
		case util.BlackDuckCRDName:
			crdConfig = horizoncomponents.NewCustomResourceDefintion(horizonapi.CRDConfig{
				APIVersion: "apiextensions.k8s.io/v1beta1",
				Name:       util.BlackDuckCRDName,
				Namespace:  namespace,
				Group:      "synopsys.com",
				CRDVersion: "v1",
				Kind:       "Blackduck",
				Plural:     "blackducks",
				Singular:   "blackduck",
				ShortNames: []string{"bds", "bd"},
				Scope:      crdScope,
			})
		case util.AlertCRDName:
			crdConfig = horizoncomponents.NewCustomResourceDefintion(horizonapi.CRDConfig{
				APIVersion: "apiextensions.k8s.io/v1beta1",
				Name:       util.AlertCRDName,
				Namespace:  namespace,
				Group:      "synopsys.com",
				CRDVersion: "v1",
				Kind:       "Alert",
				Plural:     "alerts",
				Singular:   "alert",
				Scope:      crdScope,
			})
		case util.OpsSightCRDName:
			crdConfig = horizoncomponents.NewCustomResourceDefintion(horizonapi.CRDConfig{
				APIVersion: "apiextensions.k8s.io/v1beta1",
				Name:       util.OpsSightCRDName,
				Namespace:  namespace,
				Group:      "synopsys.com",
				CRDVersion: "v1",
				Kind:       "OpsSight",
				Plural:     "opssights",
				Singular:   "opssight",
				ShortNames: []string{"ops"},
				Scope:      crdScope,
			})
		}
		if crdConfig != nil {
			crdConfig.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator", fmt.Sprintf("synopsys.com/operator.%s", namespace): namespace})
			crdConfigs = append(crdConfigs, crdConfig)
		}
	}
	return crdConfigs, nil
}

// checkUpdateCustomResource returns true if the Custom Resource definition can be updated (doesn't deploy)
func checkUpdateCustomResource(crdType string, namespace string, scope apiextensions.ResourceScope, isClusterScoped bool) (bool, error) {
	crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdType)
	if err != nil {
		return false, fmt.Errorf("unable to get Custom Resource Definition '%s' due to %+v", crdType, err)
	}

	// CRD exist with different scope and hence error out
	if scope != crd.Spec.Scope {
		return true, fmt.Errorf("the Custom Resource Definition '%s' already exists with scope '%v'. updating Custom Resource Definition scope '%s' is not supported", crd.Name, crd.Spec.Scope, crd.Name)
	}

	if isClusterScoped {
		for key, value := range crd.Labels {
			if strings.HasPrefix(key, "synopsys.com/operator.") {
				if value != namespace {
					return true, fmt.Errorf("there is already a Synopsys Operator managing '%s' in namespace '%s'. Only one Synopsys Operator may manage '%s' per namespace", crdType, value, crdType)
				}
			}
		}
	}

	// CRD exist with same scope and hence continue
	log.Infof("the Custom Resource Definition '%s' with scope '%s' already exists", crd.Name, crd.Spec.Scope)

	if _, ok := crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)]; !ok {
		crd.Labels = util.InitLabels(crd.Labels)
		crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)] = namespace
		_, err = util.UpdateCustomResourceDefinition(apiExtensionClient, crd)
		if err != nil {
			return true, fmt.Errorf("unable to update the labels for %s custom resource definition due to %+v", crd.Name, err)
		}
	}
	return true, nil
}

// getSpecToDeploySOperator returns the config from creating Synopsys Operator
func getSpecToDeploySOperator(crds []string) (*soperator.SpecConfig, error) {
	// verify Synopsys Operator image has a tag
	imageHasTag := len(strings.Split(synopsysOperatorImage, ":")) == 2
	if !imageHasTag {
		return nil, fmt.Errorf("Synopsys Operator image doesn't have a tag: %s", synopsysOperatorImage)
	}
	// generate random string as SEAL key
	log.Debugf("getting Seal Key")
	sealKey, err := util.GetRandomString(32)
	if err != nil {
		log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
	}
	// generate self signed nginx certs
	cert, key, err := util.GeneratePemSelfSignedCertificateAndKey(pkix.Name{
		CommonName: fmt.Sprintf("synopsys-operator.%s.svc", operatorNamespace),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't generate certificate and key due to %+v", err)
	}
	// Deploy Synopsys Operator
	log.Debugf("creating Synopsys Operator's components")
	soperatorSpec := soperator.NewSOperator(operatorNamespace, synopsysOperatorImage, exposeUI, soperator.GetClusterType(restconfig, kubeClient, operatorNamespace),
		strings.ToUpper(dryRun) == "TRUE", logLevel, threadiness, postgresRestartInMins, podWaitTimeoutSeconds, resyncIntervalInSeconds,
		terminationGracePeriodSeconds, sealKey, restconfig, kubeClient, cert, key, isClusterScoped, crds, admissionWebhookListener)
	return soperatorSpec, nil
}

// deployCmd deploys Synopsys Operator into your cluster
var deployCmd = &cobra.Command{
	Use:           "deploy",
	Example:       "synopsysctl deploy --enable-blackduck\nsynopsysctl deploy -n <namespace> --enable-blackduck\nsynopsysctl deploy --enable-blackduck --mock json",
	Short:         "Deploy Synopsys Operator into your cluster",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 0 {
			cmd.Help()
			return fmt.Errorf("this command doesn't take any arguments")
		}
		isValid := util.IsExposeServiceValid(exposeUI)
		if !isValid {
			cmd.Help()
			return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}

		isValid = util.IsExposeServiceValid(exposeMetrics)
		if !isValid {
			cmd.Help()
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		// Read Commandline Parameters
		if len(namespace) > 0 {
			operatorNamespace = namespace
		}

		// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
		crds, err := getEnabledCrds()
		if err != nil {
			return err
		}
		// Get CRD configs
		crdConfigs, err := getCrdConfigs(operatorNamespace, isClusterScoped, crds)
		if err != nil {
			return err
		}
		if len(crdConfigs) == 0 {
			return fmt.Errorf("no resources are enabled [include flag(s): --enable-alert --enable-blackduck --enable-opssight ]")
		}
		// Create Synopsys Operator Spec
		soperatorSpec, err := getSpecToDeploySOperator(crds)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating Spec for Synopsys Operator in namespace '%s'...", operatorNamespace)
			return PrintResource(*soperatorSpec, mockFormat, false)
		}

		// check if namespace exist in namespace scope, if not throw an error
		if !isClusterScoped {
			_, err = util.GetNamespace(kubeClient, operatorNamespace)
			if err != nil {
				return fmt.Errorf("please create the namespace '%s' to deploy the Synopsys Operator in namespace scoped", operatorNamespace)
			}
		}

		// check if operator is already installed
		_, err = util.GetOperatorNamespace(kubeClient, operatorNamespace)
		if err == nil {
			return fmt.Errorf("the Synopsys Operator instance is already deployed in namespace '%s'", namespace)
		}

		log.Infof("deploying Synopsys Operator in namespace '%s'...", operatorNamespace)

		log.Debugf("creating custom resource definitions")
		err = deployCrds(operatorNamespace, isClusterScoped, crdConfigs)
		if err != nil {
			return err
		}

		log.Debugf("creating Synopsys Operator components")
		sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
		err = sOperatorCreater.UpdateSOperatorComponents(soperatorSpec)
		if err != nil {
			return fmt.Errorf("error deploying Synopsys Operator due to %+v", err)
		}

		log.Debugf("creating Metrics components")
		promtheusSpec := soperator.NewPrometheus(operatorNamespace, metricsImage, exposeMetrics, restconfig, kubeClient)
		err = sOperatorCreater.UpdatePrometheus(promtheusSpec)
		if err != nil {
			return fmt.Errorf("error deploying metrics due to %s", err)
		}

		log.Infof("successfully submitted Synopsys Operator into namespace '%s'", operatorNamespace)
		return nil
	},
}

// deployNativeCmd prints the Kubernetes resources for a Synopsys Operator instance
var deployNativeCmd = &cobra.Command{
	Use:           "native",
	Example:       "synopsysctl deploy native --enable-blackduck\nsynopsysctl deploy native -n <namespace> --enable-blackduck\nsynopsysctl deploy native -n <namespace> --enable-blackduck -o yaml",
	Short:         "Print the Kubernetes resources for a Synopsys Operator instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 0 {
			cmd.Help()
			return fmt.Errorf("this command doesn't take any arguments")
		}
		isValid := util.IsExposeServiceValid(exposeUI)
		if !isValid {
			cmd.Help()
			return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}

		isValid = util.IsExposeServiceValid(exposeMetrics)
		if !isValid {
			cmd.Help()
			return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read Commandline Parameters
		if len(namespace) > 0 {
			operatorNamespace = namespace
		}
		// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
		crds, err := getEnabledCrds()
		if err != nil {
			return nil
		}
		// Get CRD configs
		crdConfigs, err := getCrdConfigs(operatorNamespace, isClusterScoped, crds)
		if err != nil {
			return err
		}
		if len(crdConfigs) == 0 {
			return fmt.Errorf("no resources are enabled [include flag(s): --enable-alert --enable-blackduck --enable-opssight ]")
		}
		// Create Synopsys Operator Spec
		sOperatorSpec, err := getSpecToDeploySOperator(crds)
		if err != nil {
			return err
		}

		log.Debugf("generating Kubernetes resources for Synopsys Operator in namespace '%s'...", operatorNamespace)
		if err := PrintResource(*sOperatorSpec, nativeFormat, true); err != nil {
			return err
		}
		for _, crdConfig := range crdConfigs {
			if _, err := PrintComponent(*crdConfig.CustomResourceDefinition, nativeFormat); err != nil {
				return err
			}
		}
		return nil
	},
}

func addOperatorDeployFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&isClusterScoped, "cluster-scoped", "c", isClusterScoped, "Enable/Disable Synopsys Operator with cluster scope")
	cmd.Flags().BoolVarP(&isEnabledAlert, "enable-alert", "a", isEnabledAlert, "Enable/Disable Alert Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().BoolVarP(&isEnabledBlackDuck, "enable-blackduck", "b", isEnabledBlackDuck, "Enable/Disable Black Duck Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().BoolVarP(&isEnabledOpsSight, "enable-opssight", "s", isEnabledOpsSight, "Enable/Disable OpsSight Custom Resource Definition (CRD) in your cluster")
	// cmd.Flags().BoolVarP(&isEnabledPrm, "enable-prm", "p", isEnabledPrm, "Enable/Disable Polaris Reporting Module Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	cmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "Image URL of Synopsys Operator")
	cmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "x", exposeMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	cmd.Flags().StringVarP(&metricsImage, "metrics-image", "m", metricsImage, "Image URL of Synopsys Operator's metrics pod")
	cmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "q", postgresRestartInMins, "Minutes to check for restarting postgres")
	cmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	cmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	cmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	cmd.Flags().StringVarP(&dryRun, "dry-run", "d", dryRun, "If true, Synopsys Operator runs without being connected to a cluster [true|false]")
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	cmd.Flags().IntVarP(&threadiness, "no-of-threads", "t", threadiness, "Number of threads to process the custom resources")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the Synopsys Operator instance")
}

func initFlags() {
	addOperatorDeployFlags(deployCmd)
	addMockFlag(deployCmd)
	rootCmd.AddCommand(deployCmd)

	addOperatorDeployFlags(deployNativeCmd)
	addNativeFormatFlag(deployNativeCmd)
	deployCmd.AddCommand(deployNativeCmd)
}
