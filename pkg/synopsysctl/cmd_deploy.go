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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//  Deploy Command Defaults
var deployNamespace string
var exposeUI = ""
var synopsysOperatorImage string
var metricsImage string
var exposeMetrics = ""
var terminationGracePeriodSeconds int64 = 180
var operatorTimeBombInSeconds int64 = 315576000
var dryRun = false
var logLevel = "debug"
var threadiness = 5
var postgresRestartInMins int64 = 10
var podWaitTimeoutSeconds int64 = 600
var resyncIntervalInSeconds int64 = 120
var deploySecretType = "Opaque"
var adminPassword = "blackduck"
var postgresPassword = "blackduck"
var userPassword = "blackduck"
var blackDuckPassword = "blackduck"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:     "deploy [namespace]",
	Example: "synopsysctl deploy\nsynopsysctl deploy sonamespace\nsynopsysctl deploy --expose-ui LOADBALANCER",
	Short:   "Deploys Synopsys Operator into your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 1 {
			return fmt.Errorf("this command takes up to 1 argument")
		}
		// Check the Secret Type
		var err error
		secretType, err = operatorutil.SecretTypeNameToHorizon(deploySecretType)
		if err != nil {
			log.Errorf("failed to convert secret type: %s", err)
			return nil
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read Commandline Parameters
		deployNamespace = DefaultDeployNamespace
		if len(args) == 1 {
			deployNamespace = args[0]
		}

		if !cmd.Flags().Lookup("synopsys-operator-image").Changed {
			synopsysOperatorImage = DefaultOperatorImage
		}

		if !cmd.Flags().Lookup("metrics-image").Changed {
			metricsImage = DefaultMetricsImage
		}

		// check if operator is already installed
		ns, err := operatorutil.GetOperatorNamespace(kubeClient)
		if err == nil {
			log.Errorf("Synopsys Operator is already installed in '%s' namespace", ns)
			return nil
		}

		log.Infof("deploying Synopsys Operator in '%s' namespace......", deployNamespace)

		// if image image tag
		imageHasTag := len(strings.Split(synopsysOperatorImage, ":")) == 2
		if !imageHasTag {
			log.Errorf("Synopsys Operator image doesn't have a tag: %s", synopsysOperatorImage)
			return nil
		}

		log.Debugf("getting Seal Key")
		sealKey, err := operatorutil.GetRandomString(32)
		if err != nil {
			log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
		}

		cert, key, err := operatorutil.GeneratePemSelfSignedCertificateAndKey(pkix.Name{
			CommonName: fmt.Sprintf("synopsys-operator.%s.svc", deployNamespace),
		})
		if err != nil {
			log.Errorf("couldn't generate certificate and key due to %+v", err)
			return nil
		}

		// Deploy Synopsys Operator
		log.Debugf("creating Synopsys Operator components")
		soperatorSpec := soperator.NewSOperator(deployNamespace, synopsysOperatorImage, exposeUI, adminPassword, postgresPassword,
			userPassword, blackDuckPassword, secretType, operatorTimeBombInSeconds, dryRun, logLevel, threadiness, postgresRestartInMins,
			podWaitTimeoutSeconds, resyncIntervalInSeconds, terminationGracePeriodSeconds, sealKey, restconfig, kubeClient, cert, key)

		err = soperatorSpec.UpdateSOperatorComponents()
		if err != nil {
			log.Errorf("error in deploying Synopsys Operator due to %+v", err)
			return nil
		}

		// Deploy Metrics Components for Prometheus
		log.Debugf("creating Metrics components")
		promtheusSpec := soperator.NewPrometheus(deployNamespace, metricsImage, exposeMetrics, restconfig, kubeClient)
		err = promtheusSpec.UpdatePrometheus()
		if err != nil {
			log.Errorf("error deploying metrics: %s", err)
			return nil
		}

		log.Infof("successfully deployed Synopsys Operator")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	deployCmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", DefaultOperatorImage, "Image URL of Synopsys Operator")
	deployCmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "m", exposeMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT]")
	deployCmd.Flags().StringVarP(&metricsImage, "metrics-image", "k", DefaultMetricsImage, "Image URL of Synopsys Operator's metrics pod")
	deployCmd.Flags().StringVar(&deploySecretType, "secret-type", deploySecretType, "Type of kubernetes secret to store the postgres and Black Duck credentials")
	deployCmd.Flags().StringVarP(&adminPassword, "admin-password", "a", adminPassword, "Postgres admin password")
	deployCmd.Flags().StringVarP(&postgresPassword, "postgres-password", "p", postgresPassword, "Postgres password")
	deployCmd.Flags().StringVarP(&userPassword, "user-password", "u", userPassword, "Postgres user password")
	deployCmd.Flags().StringVarP(&blackDuckPassword, "blackduck-password", "b", blackDuckPassword, "Black Duck password of 'sysadmin' account")
	deployCmd.Flags().Int64VarP(&operatorTimeBombInSeconds, "operator-time-bomb-in-seconds", "o", operatorTimeBombInSeconds, "Termination grace period in seconds for shutting down crds")
	deployCmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "n", postgresRestartInMins, "Minutes to check for restarting postgres")
	deployCmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	deployCmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	deployCmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	deployCmd.Flags().BoolVar(&dryRun, "dryRun", dryRun, "If true, Synopsys Operator runs without being connected to a cluster")
	deployCmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	deployCmd.Flags().IntVarP(&threadiness, "no-of-threads", "c", threadiness, "Number of threads to process the custom resources")
}
