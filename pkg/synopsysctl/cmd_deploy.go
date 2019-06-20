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

	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//  Deploy Command Defaults
var operatorNamespace = DefaultOperatorNamespace
var exposeUI = ""
var synopsysOperatorImage = DefaultOperatorImage
var metricsImage = DefaultMetricsImage
var exposeMetrics = ""
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

// Flags for using mock mode - doesn't deploy
var mockFormat string
var mockKubeFormat string

// deployCmd creates a Synopsys Operator instance in the cluster
var deployCmd = &cobra.Command{
	Use:           "deploy",
	Example:       "synopsysctl deploy --enable-blackduck\nsynopsysctl deploy -n <namespace> --enable-blackduck",
	Short:         "Deploy Synopsys Operator into your cluster",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 0 {
			cmd.Help()
			return fmt.Errorf("this command doesn't take any arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read Commandline Parameters
		if len(namespace) > 0 {
			operatorNamespace = namespace
		}

		var err error
		if !cmd.Flags().Lookup("mock").Changed && !cmd.Flags().Lookup("mock-kube").Changed {
			if isClusterScoped && len(operatorNamespace) == 0 {
				namespace = metav1.NamespaceAll
			}

			// check if operator is already installed
			namespace, err = operatorutil.GetOperatorNamespace(kubeClient, namespace)
			if err == nil {
				return fmt.Errorf("Synopsys Operator is already installed in namespace '%s'", namespace)
			}

			if metav1.NamespaceAll != namespace {
				operatorNamespace = namespace
			}
		}

		// verify Synopsys Operator image has a tag
		imageHasTag := len(strings.Split(synopsysOperatorImage, ":")) == 2
		if !imageHasTag {
			return fmt.Errorf("Synopsys Operator image doesn't have a tag: %s", synopsysOperatorImage)
		}

		// generate random string as SEAL key
		log.Debugf("getting Seal Key")
		sealKey, err := operatorutil.GetRandomString(32)
		if err != nil {
			log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
		}

		// generate self signed nginx certs
		cert, key, err := operatorutil.GeneratePemSelfSignedCertificateAndKey(pkix.Name{
			CommonName: fmt.Sprintf("synopsys-operator.%s.svc", operatorNamespace),
		})
		if err != nil {
			return fmt.Errorf("couldn't generate certificate and key due to %+v", err)
		}

		// validate each CRD enable parameter is enabled/disabled and cluster scope are from supported values
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
			log.Error("unable to create the OpsSight Custom Resource Definition (CRD) due to having a namespaced scope for Synopsys Operator. Please enable the cluster scope to install an OpsSight CRD")
		}

		// if isEnabledPrm {
		// 	crds = append(crds, util.PrmCRDName)
		// }

		// Deploy Synopsys Operator
		log.Debugf("creating Synopsys Operator components")
		soperatorSpec := soperator.NewSOperator(operatorNamespace, synopsysOperatorImage, exposeUI, soperator.GetClusterType(restconfig, operatorNamespace),
			strings.ToUpper(dryRun) == "TRUE", logLevel, threadiness, postgresRestartInMins, podWaitTimeoutSeconds, resyncIntervalInSeconds,
			terminationGracePeriodSeconds, sealKey, restconfig, kubeClient, cert, key, isClusterScoped, crds, admissionWebhookListener)

		if cmd.Flags().Lookup("mock").Changed {
			log.Debugf("running mock mode")
			err := PrintResource(*soperatorSpec, mockFormat, false)
			if err != nil {
				return err
			}
		} else if cmd.Flags().Lookup("mock-kube").Changed {
			log.Debugf("running kube mock mode")
			err := PrintResource(*soperatorSpec, mockKubeFormat, true)
			if err != nil {
				return err
			}
		} else {
			sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
			// Deploy Synopsys Operator
			log.Infof("deploying Synopsys Operator in namespace '%s'...", operatorNamespace)
			err = sOperatorCreater.UpdateSOperatorComponents(soperatorSpec)
			if err != nil {
				return fmt.Errorf("error deploying Synopsys Operator due to %+v", err)
			}

			// Deploy Prometheus Metrics Components for Synopsys Operator
			log.Debugf("creating Metrics components")
			promtheusSpec := soperator.NewPrometheus(operatorNamespace, metricsImage, exposeMetrics, restconfig, kubeClient)
			err = sOperatorCreater.UpdatePrometheus(promtheusSpec)
			if err != nil {
				return fmt.Errorf("error deploying metrics: %s", err)
			}
			log.Infof("successfully submitted Synopsys Operator into namespace '%s'", operatorNamespace)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&isClusterScoped, "cluster-scoped", "c", isClusterScoped, "Enable/Disable Synopsys Operator with cluster scope")
	deployCmd.Flags().BoolVarP(&isEnabledAlert, "enable-alert", "a", isEnabledAlert, "Enable/Disable Alert Custom Resource Definition (CRD) in your cluster")
	deployCmd.Flags().BoolVarP(&isEnabledBlackDuck, "enable-blackduck", "b", isEnabledBlackDuck, "Enable/Disable Black Duck Custom Resource Definition (CRD) in your cluster")
	deployCmd.Flags().BoolVarP(&isEnabledOpsSight, "enable-opssight", "s", isEnabledOpsSight, "Enable/Disable OpsSight Custom Resource Definition (CRD) in your cluster")
	// deployCmd.Flags().BoolVarP(&isEnabledPrm, "enable-prm", "p", isEnabledPrm, "Enable/Disable Polaris Reporting Module Custom Resource Definition (CRD) in your cluster")
	deployCmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	deployCmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "Image URL of Synopsys Operator")
	deployCmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "x", exposeMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT]")
	deployCmd.Flags().StringVarP(&metricsImage, "metrics-image", "m", metricsImage, "Image URL of Synopsys Operator's metrics pod")
	deployCmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "q", postgresRestartInMins, "Minutes to check for restarting postgres")
	deployCmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	deployCmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	deployCmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	deployCmd.Flags().StringVarP(&dryRun, "dry-run", "d", dryRun, "If true, Synopsys Operator runs without being connected to a cluster [true|false]")
	deployCmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	deployCmd.Flags().IntVarP(&threadiness, "no-of-threads", "t", threadiness, "Number of threads to process the custom resources")
	deployCmd.Flags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Prints Synopsys Operator's spec in the specified format instead of creating it [json|yaml]")
	deployCmd.Flags().StringVarP(&mockKubeFormat, "mock-kube", "k", mockKubeFormat, "Prints Synopsys Operator's Kubernetes resource specs in the specified format instead of creating it [json|yaml]")
	deployCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the Synopsys Operator instance")
}
