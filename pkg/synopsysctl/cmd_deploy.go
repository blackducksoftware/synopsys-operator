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
var blackduckPassword = "blackduck"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [NAMESPACE]",
	Short: "Deploys the synopsys operator onto your cluster",
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
			log.Errorf("synopsys operator is already installed in '%s' namespace", ns)
			return nil
		}

		log.Infof("deploying the synopsys operator in '%s' namespace......", deployNamespace)

		// if image image tag
		imageHasTag := len(strings.Split(synopsysOperatorImage, ":")) == 2
		if !imageHasTag {
			log.Errorf("synopsys operator image doesn't have a tag: %s", synopsysOperatorImage)
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

		// Deploy synopsys-operator
		log.Debugf("creating Synopsys-Operator components")
		soperatorSpec := soperator.NewSOperator(deployNamespace, synopsysOperatorImage, exposeUI, adminPassword, postgresPassword,
			userPassword, blackduckPassword, secretType, operatorTimeBombInSeconds, dryRun, logLevel, threadiness, postgresRestartInMins,
			podWaitTimeoutSeconds, resyncIntervalInSeconds, terminationGracePeriodSeconds, sealKey, restconfig, kubeClient, cert, key)

		err = soperatorSpec.UpdateSOperatorComponents()
		if err != nil {
			log.Errorf("error in deploying the synopsys operator due to %+v", err)
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

		log.Infof("successfully deployed the synopsys operator")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "expose the synopsys operator's user interface. possible values are [NODEPORT/LOADBALANCER/OPENSHIFT]")
	deployCmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", DefaultOperatorImage, "synopsys operator image URL")
	deployCmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "m", exposeMetrics, "expose the Synopsys-Operator's metrics application. possible values are [NODEPORT/LOADBALANCER/OPENSHIFT]")
	deployCmd.Flags().StringVarP(&metricsImage, "metrics-image", "k", DefaultMetricsImage, "image URL for the Synopsys-Operator's metrics pod")
	deployCmd.Flags().StringVar(&deploySecretType, "secret-type", deploySecretType, "type of kubernetes secret to store the postgres and blackduck credentials")
	deployCmd.Flags().StringVarP(&adminPassword, "admin-password", "a", adminPassword, "postgres admin password")
	deployCmd.Flags().StringVarP(&postgresPassword, "postgres-password", "p", postgresPassword, "postgres password")
	deployCmd.Flags().StringVarP(&userPassword, "user-password", "u", userPassword, "postgres user password")
	deployCmd.Flags().StringVarP(&blackduckPassword, "blackduck-password", "b", blackduckPassword, "blackduck password for 'sysadmin' account")
	deployCmd.Flags().Int64VarP(&operatorTimeBombInSeconds, "operator-time-bomb-in-seconds", "o", operatorTimeBombInSeconds, "termination grace period in seconds for shutting down crds")
	deployCmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "n", postgresRestartInMins, "check for postgres restart in minutes")
	deployCmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "wait for pod to be running in seconds")
	deployCmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "custom resources resync time period in seconds")
	deployCmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "termination grace period in seconds for shutting down postgres")
	deployCmd.Flags().BoolVar(&dryRun, "dryRun", dryRun, "dry run to run the test cases")
	deployCmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "log level of synopsys operator")
	deployCmd.Flags().IntVarP(&threadiness, "no-of-threads", "c", threadiness, "number of threads to process the custom resources")
}
