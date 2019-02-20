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
	"errors"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/spf13/cobra"
)

var secretType horizonapi.SecretType

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Deploys the synopsys operator onto your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
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
			fmt.Printf("Invalid Secret Type: %s\n", startSecretType)
			return errors.New("Bad Secret Type")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// check if operator is already installed
		out, err := RunKubeCmd("get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
		if err == nil {
			fmt.Printf("You have already installed the operator in namespace %s.\n", out)
			fmt.Printf("To delete the operator run: synopsysctl stop --namespace %s\n", out)
			fmt.Printf("Nothing to do...\n")
			return
		}

		// Start Horizon
		rc := getKubeRestConfig()

		// Create a Horizon Deployer to set up the environment for the Synopsys Operator
		environmentDeployer, err := deployer.NewDeployer(rc)

		// create a new namespace
		ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
			// APIVersion:  "string",
			// ClusterName: "string",
			Name:      namespace,
			Namespace: namespace,
		})
		environmentDeployer.AddNamespace(ns)

		// create a secret
		secret := horizoncomponents.NewSecret(horizonapi.SecretConfig{
			APIVersion: "v1",
			// ClusterName : "cluster",
			Name:      startSecretName,
			Namespace: namespace,
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
			fmt.Printf("Error deploying Environment with Horizon : %s\n", err)
			return
		}

		// Deploy synopsys-operator
		soperatorSpec := SOperatorSpecConfig{
			Namespace:                namespace,
			SynopsysOperatorImage:    startSynopsysOperatorImage,
			BlackduckRegistrationKey: startBlackduckRegistrationKey,
		}
		synopsysOperatorDeployer, err := deployer.NewDeployer(rc)
		if err != nil {
			fmt.Printf("Error creating Horizon Deployer for Synopsys Operator: %s\n", err)
			return
		}
		synopsysOperatorDeployer.AddReplicationController(soperatorSpec.GetOperatorReplicationController())
		synopsysOperatorDeployer.AddService(soperatorSpec.GetOperatorService())
		synopsysOperatorDeployer.AddConfigMap(soperatorSpec.GetOperatorConfigMap())
		synopsysOperatorDeployer.AddServiceAccount(soperatorSpec.GetOperatorServiceAccount())
		synopsysOperatorDeployer.AddClusterRoleBinding(soperatorSpec.GetOperatorClusterRoleBinding())
		err = synopsysOperatorDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying Synopsys Operator with Horizon : %s\n", err)
			return
		}

		// Deploy prometheus
		promtheusSpec := PrometheusSpecConfig{
			Namespace:       namespace,
			PrometheusImage: startPrometheusImage,
		}
		prometheusDeployer, err := deployer.NewDeployer(rc)
		if err != nil {
			fmt.Printf("Error creating Horizon Deployer for Prometheus: %s\n", err)
			return
		}
		prometheusDeployer.AddService(promtheusSpec.GetPrometheusService())
		prometheusDeployer.AddDeployment(promtheusSpec.GetPrometheusDeployment())
		prometheusDeployer.AddConfigMap(promtheusSpec.GetPrometheusConfigMap())
		err = prometheusDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying Prometheus with Horizon : %s\n", err)
			return
		}

		// secret link stuff
		RunKubeCmd("create", "secret", "generic", "custom-registry-pull-secret", fmt.Sprintf("--from-file=.dockerconfigjson=%s", startDockerConfigPath), "--type=kubernetes.io/dockerconfigjson")
		RunKubeCmd("secrets", "link", "default", "custom-registry-pull-secret", "--for=pull")
		RunKubeCmd("secrets", "link", "synopsys-operator", "custom-registry-pull-secret", "--for=pull")
		RunKubeCmd("scale", "rc", "synopsys-operator", "--replicas=0")
		RunKubeCmd("scale", "rc", "synopsys-operator", "--replicas=1")

		// expose the routes
		out, err = RunKubeCmd("expose", "rc", "synopsys-operator", "--port=80", "--target-port=3000", "--name=synopsys-operator-tcp", "--type=LoadBalancer", fmt.Sprintf("--namespace=%s", namespace))
		if err != nil {
			fmt.Printf("Error exposing the Synopsys-Operator's Replication Controller: %s", out)
		}
		out, err = RunKubeCmd("create", "route", "edge", "--service=synopsys-operator-tcp", "-n", namespace)
		if err != nil {
			fmt.Printf("Could not create route (Possible Reason: Kubernetes doesn't support Routes): %s", out)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

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
}
