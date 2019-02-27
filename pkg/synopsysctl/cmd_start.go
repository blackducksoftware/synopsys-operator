// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Start Command Defaults
var exposeUI = false
var startNamespace = "synopsys-operator"
var startSynopsysOperatorImage = "docker.io/blackducksoftware/synopsys-operator:2019.2.0-RC"
var startPrometheusImage = "docker.io/prom/prometheus:v2.1.0"
var startBlackduckRegistrationKey = ""
var startDockerConfigPath = ""
var startSecretName = "blackduck-secret"
var startSecretType = "Opaque"
var startSecretAdminPassword = "YmxhY2tkdWNr"
var startSecretPostgresPassword = "YmxhY2tkdWNr"
var startSecretUserPassword = "YmxhY2tkdWNr"
var startSecretBlackduckPassword = "YmxhY2tkdWNr"

// Start Global Variables
var secretType horizonapi.SecretType

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [NAME]",
	Short: "Deploys the synopsys operator onto your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument - NAME")
		}
		// Check the Secret Type
		switch startSecretType {
		case "Opaque":
			secretType = horizonapi.SecretTypeOpaque
		case "ServiceAccountToken":
			secretType = horizonapi.SecretTypeServiceAccountToken
		case "Dockercfg":
			secretType = horizonapi.SecretTypeDockercfg
		case "DockerConfigJSON":
			secretType = horizonapi.SecretTypeDockerConfigJSON
		case "BasicAuth":
			secretType = horizonapi.SecretTypeBasicAuth
		case "SSHAuth":
			secretType = horizonapi.SecretTypeSSHAuth
		case "TypeTLS":
			secretType = horizonapi.SecretTypeTLS
		default:
			return fmt.Errorf("Invalid Secret Type: %s", startSecretType)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Starting the Synopsys-Operator: %s\n", startNamespace)
		// Read Commandline Parameters
		if len(args) == 1 {
			startNamespace = args[0]
		}
		// check if operator is already installed
		out, err := RunKubeCmd("get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
		if err == nil {
			log.Errorf("Synopsys-Operator is already installed in namespace %s.", out)
			return nil
		}

		// Create a Horizon Deployer to set up the environment for the Synopsys Operator
		environmentDeployer, err := deployer.NewDeployer(restconfig)

		// create a new namespace
		ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
			// APIVersion:  "string",
			// ClusterName: "string",
			Name:      startNamespace,
			Namespace: startNamespace,
		})
		environmentDeployer.AddNamespace(ns)

		// create a secret
		secret := horizoncomponents.NewSecret(horizonapi.SecretConfig{
			APIVersion: "v1",
			// ClusterName : "cluster",
			Name:      startSecretName,
			Namespace: startNamespace,
			Type:      secretType,
		})
		secret.AddData(map[string][]byte{
			"ADMIN_PASSWORD":    []byte(startSecretAdminPassword),
			"POSTGRES_PASSWORD": []byte(startSecretPostgresPassword),
			"USER_PASSWORD":     []byte(startSecretUserPassword),
			"HUB_PASSWORD":      []byte(startSecretBlackduckPassword),
		})
		environmentDeployer.AddSecret(secret)

		// Deploy Resources for the Synopsys Operator
		err = environmentDeployer.Run()
		if err != nil {
			log.Errorf("Error deploying Environment with Horizon : %s", err)
			return nil
		}

		// Deploy synopsys-operator
		soperatorSpec := SOperatorSpecConfig{
			Namespace:                startNamespace,
			SynopsysOperatorImage:    startSynopsysOperatorImage,
			BlackduckRegistrationKey: startBlackduckRegistrationKey,
		}
		synopsysOperatorDeployer, err := deployer.NewDeployer(restconfig)
		if err != nil {
			log.Errorf("Error creating Horizon Deployer for Synopsys Operator: %s", err)
			return nil
		}
		synopsysOperatorDeployer.AddReplicationController(soperatorSpec.GetOperatorReplicationController())
		synopsysOperatorDeployer.AddService(soperatorSpec.GetOperatorService())
		synopsysOperatorDeployer.AddConfigMap(soperatorSpec.GetOperatorConfigMap())
		synopsysOperatorDeployer.AddServiceAccount(soperatorSpec.GetOperatorServiceAccount())
		synopsysOperatorDeployer.AddClusterRoleBinding(soperatorSpec.GetOperatorClusterRoleBinding())
		err = synopsysOperatorDeployer.Run()
		if err != nil {
			return fmt.Errorf("Error deploying Synopsys Operator with Horizon : %s", err)
		}

		// Deploy prometheus
		promtheusSpec := PrometheusSpecConfig{
			Namespace:       startNamespace,
			PrometheusImage: startPrometheusImage,
		}
		prometheusDeployer, err := deployer.NewDeployer(restconfig)
		if err != nil {
			log.Errorf("Error creating Horizon Deployer for Prometheus: %s", err)
			return nil
		}
		prometheusDeployer.AddService(promtheusSpec.GetPrometheusService())
		prometheusDeployer.AddDeployment(promtheusSpec.GetPrometheusDeployment())
		prometheusDeployer.AddConfigMap(promtheusSpec.GetPrometheusConfigMap())
		err = prometheusDeployer.Run()
		if err != nil {
			log.Errorf("Error deploying Prometheus with Horizon : %s", err)
			return nil
		}

		// secret link stuff
		RunKubeCmd("create", "secret", "generic", "custom-registry-pull-secret", fmt.Sprintf("--from-file=.dockerconfigjson=%s", startDockerConfigPath), "--type=kubernetes.io/dockerconfigjson")
		RunKubeCmd("secrets", "link", "default", "custom-registry-pull-secret", "--for=pull")
		RunKubeCmd("secrets", "link", "synopsys-operator", "custom-registry-pull-secret", "--for=pull")
		RunKubeCmd("scale", "replicationcontroller", "synopsys-operator", "--replicas=0")
		RunKubeCmd("scale", "replicationcontroller", "synopsys-operator", "--replicas=1")

		// expose the routes
		if exposeUI {
			out, err = RunKubeCmd("expose", "replicationcontroller", "synopsys-operator", "--port=80", "--target-port=3000", "--name=synopsys-operator-tcp", "--type=LoadBalancer", fmt.Sprintf("--namespace=%s", startNamespace))
			if err != nil {
				log.Warnf("Error exposing the Synopsys-Operator's Replication Controller: %s", out)
			}
			out, err = RunKubeCmd("create", "route", "edge", "--service=synopsys-operator-tcp", "-n", startNamespace)
			if err != nil {
				log.Warnf("Could not create route (Possible Reason: Kubernetes doesn't support Routes): %s", out)
			}
		} else {
			log.Warnf("Synopsys-Operator UI is not exposed ( --expose-ui=true to expose )")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVar(&exposeUI, "expose-ui", exposeUI, "Expose the Synopsys-Operator's User Interface")
	startCmd.Flags().StringVarP(&startSynopsysOperatorImage, "synopsys-operator-image", "i", startSynopsysOperatorImage, "synopsys operator image URL")
	startCmd.Flags().StringVarP(&startPrometheusImage, "prometheus-image", "p", startPrometheusImage, "prometheus image URL")
	startCmd.Flags().StringVarP(&startBlackduckRegistrationKey, "blackduck-registration-key", "k", startBlackduckRegistrationKey, "key to register with KnowledgeBase")
	startCmd.Flags().StringVarP(&startDockerConfigPath, "docker-config", "d", startDockerConfigPath, "path to docker config (image pull secrets etc)")
	startCmd.Flags().StringVar(&startSecretName, "secret-name", startSecretName, "name of kubernetes secret for postgres and blackduck")
	startCmd.Flags().StringVar(&startSecretType, "secret-type", startSecretType, "type of kubernetes secret for postgres and blackduck")
	startCmd.Flags().StringVar(&startSecretAdminPassword, "admin-password", startSecretAdminPassword, "postgres admin password")
	startCmd.Flags().StringVar(&startSecretPostgresPassword, "postgres-password", startSecretPostgresPassword, "postgres password")
	startCmd.Flags().StringVar(&startSecretUserPassword, "user-password", startSecretUserPassword, "postgres user password")
	startCmd.Flags().StringVar(&startSecretBlackduckPassword, "blackduck-password", startSecretBlackduckPassword, "blackduck password for 'sysadmin' account")

	// Set Log Level
	log.SetLevel(log.DebugLevel)
}
