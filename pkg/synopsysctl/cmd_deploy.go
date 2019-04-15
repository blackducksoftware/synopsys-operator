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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//  Deploy Command Defaults
var exposeUI = ""
var deployNamespace = "synopsys-operator"
var synopsysOperatorImage = "docker.io/blackducksoftware/synopsys-operator:2019.4.0-RC"
var prometheusImage = "docker.io/prom/prometheus:v2.1.0"
var terminationGracePeriodSeconds int64 = 180
var operatorTimeBombInSeconds int64 = 315576000
var dryRun = false
var logLevel = "debug"
var threadiness = 5
var postgresRestartInMins int64 = 10
var podWaitTimeoutSeconds int64 = 600
var resyncIntervalInSeconds int64 = 120
var dockerConfigPath = ""
var deploySecretType = "Opaque"
var adminPassword = "blackduck"
var postgresPassword = "blackduck"
var userPassword = "blackduck"
var blackduckPassword = "blackduck"

// Deploy Global Variables
var secretType horizonapi.SecretType

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [NAMESPACE]",
	Short: "Deploys the synopsys operator onto your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 1 {
			return fmt.Errorf("namespace to deploy the synopsys operator is missing")
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
		log.Debugf("deploying the synopsys operator in %s namespace", deployNamespace)
		// Read Commandline Parameters
		if len(args) == 1 {
			deployNamespace = args[0]
		}
		// check if operator is already installed
		crb, err := util.GetClusterRoleBinding(kubeClient, "synopsys-operator-admin")
		if err == nil {
			log.Errorf("synopsys operator is already installed in %s namespace", crb.Subjects[0].Namespace)
			return nil
		}

		sealKey, err := operatorutil.GetRandomString(32)
		if err != nil {
			log.Panicf("unable to generate the random string for SEAL_KEY due to %+v", err)
		}

		// Deploy synopsys-operator
		soperatorSpec := soperator.NewSOperator(deployNamespace, synopsysOperatorImage, exposeUI, adminPassword, postgresPassword,
			userPassword, blackduckPassword, secretType, operatorTimeBombInSeconds, dryRun, logLevel, threadiness, postgresRestartInMins,
			podWaitTimeoutSeconds, resyncIntervalInSeconds, terminationGracePeriodSeconds, sealKey, restconfig, kubeClient)

		err = soperatorSpec.UpdateSOperatorComponents()
		if err != nil {
			log.Errorf("error in deploying the synopsys operator due to %+v", err)
			return nil
		}

		// Deploy prometheus
		promtheusSpec := soperator.NewPrometheus(deployNamespace, prometheusImage, restconfig, kubeClient)
		err = promtheusSpec.UpdatePrometheus()
		if err != nil {
			log.Errorf("error deploying Prometheus: %s", err)
			return nil
		}

		// create secrets (TDDO I think this only works on OpenShift)
		if openshift && len(dockerConfigPath) > 0 {
			util.RunKubeCmd(restconfig, kube, openshift, "create", "secret", "generic", "custom-registry-pull-secret", fmt.Sprintf("--from-file=.dockerconfigjson=%s", dockerConfigPath), "--type=kubernetes.io/dockerconfigjson")
			util.RunKubeCmd(restconfig, kube, openshift, "secrets", "link", "default", "custom-registry-pull-secret", "--for=pull")
			util.RunKubeCmd(restconfig, kube, openshift, "secrets", "link", "synopsys-operator", "custom-registry-pull-secret", "--for=pull")
			util.RunKubeCmd(restconfig, kube, openshift, "scale", "replicationcontroller", "synopsys-operator", "--replicas=0")
			util.RunKubeCmd(restconfig, kube, openshift, "scale", "replicationcontroller", "synopsys-operator", "--replicas=1")
		}

		// expose the routes
		if strings.EqualFold(exposeUI, "OPENSHIFT") {
			routeClient, err := routeclient.NewForConfig(restconfig)
			if err != nil {
				log.Errorf("unable to create the route client due to %+v", err)
				return nil
			}
			_, err = util.CreateOpenShiftRoutes(routeClient, deployNamespace, "synopsys-operator", "Service", "synopsys-operator", routev1.TLSTerminationEdge)
			if err != nil {
				log.Warnf("could not create route (possible reason: kubernetes doesn't support routes) due to %+v", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "expose the synopsys operator's user interface. possible values are NODEPORT, LOADBALANCER, OPENSHIFT (to create routes)")
	deployCmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "synopsys operator image URL")
	deployCmd.Flags().StringVarP(&prometheusImage, "prometheus-image", "k", prometheusImage, "prometheus image URL")
	deployCmd.Flags().StringVarP(&dockerConfigPath, "docker-config", "d", dockerConfigPath, "path to docker config (image pull secrets etc)")
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

	// Set Log Level
	level, err := getLogLevel(logLevel)
	if err != nil {
		log.Errorf("invalid log level, error: %+v", err)
	}
	log.SetLevel(level)
}

// getLogLevel returns the log level
func getLogLevel(logLevel string) (log.Level, error) {
	return log.ParseLevel(logLevel)
}
