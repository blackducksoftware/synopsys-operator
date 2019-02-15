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
		switch start_secretType {
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
			fmt.Printf("Invalid Secret Type: %s\n", start_secretType)
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
			Name:      start_secretName,
			Namespace: namespace,
			Type:      secretType,
		})
		secret.AddData(map[string][]byte{
			"ADMIN_PASSWORD":    []byte(start_secretAdminPassword),
			"POSTGRES_PASSWORD": []byte(start_secretPostgresPassword),
			"USER_PASSWORD":     []byte(start_secretUserPassword),
			"HUB_PASSWORD":      []byte(start_secretBlackduckPassword),
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
			SynopsysOperatorImage:    start_synopsysOperatorImage,
			BlackduckRegistrationKey: start_blackduckRegistrationKey,
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
			PrometheusImage: start_prometheusImage,
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
		RunKubeCmd("create", "secret", "generic", "custom-registry-pull-secret", fmt.Sprintf("--from-file=.dockerconfigjson=%s", start_dockerConfigPath), "--type=kubernetes.io/dockerconfigjson")
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

	startCmd.Flags().StringVarP(&start_synopsysOperatorImage, "synopsys-operator-image", "i", start_synopsysOperatorImage, "synopsys operator image URL")
	startCmd.Flags().StringVarP(&start_prometheusImage, "prometheus-image", "p", start_prometheusImage, "prometheus image URL")
	startCmd.Flags().StringVarP(&start_blackduckRegistrationKey, "blackduck-registration-key", "k", start_blackduckRegistrationKey, "key to register with KnowledgeBase")
	startCmd.Flags().StringVarP(&start_dockerConfigPath, "docker-config", "d", start_dockerConfigPath, "path to docker config (image pull secrets etc)")

	startCmd.Flags().StringVar(&start_secretName, "secret-name", start_secretName, "name of kubernetes secret for postgres and blackduck")
	startCmd.Flags().StringVar(&start_secretType, "secret-type", start_secretType, "type of kubernetes secret for postgres and blackduck")
	startCmd.Flags().StringVar(&start_secretAdminPassword, "admin-password", start_secretAdminPassword, "postgres admin password")
	startCmd.Flags().StringVar(&start_secretPostgresPassword, "postgres-password", start_secretPostgresPassword, "postgres password")
	startCmd.Flags().StringVar(&start_secretUserPassword, "user-password", start_secretUserPassword, "postgres user password")
	startCmd.Flags().StringVar(&start_secretBlackduckPassword, "blackduck-password", start_secretBlackduckPassword, "blackduck password for 'sysadmin' account")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
