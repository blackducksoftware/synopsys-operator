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

package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var secretType horizonapi.SecretType

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Deploys the synopsys operator onto your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Secret Type
		switch init_secretType {
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
			fmt.Printf("Invalid Secret Type: %s\n", init_secretType)
			return errors.New("Bad Secret Type")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("at this point we would call kube/install.sh -i %s -p %s -k %s -d %s\n", init_synopsysOperatorImage, init_promethiusImage, init_blackduckRegistrationKey, init_dockerConfigPath)

		// check if operator is already installed
		out, err := RunKubeCmd("get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
		if err == nil {
			fmt.Printf("You have already installed the operator in namespace %s.\n", out)
			fmt.Printf("To delete the operator run: synopsysctl stop --namespace %s\n", out)
			fmt.Printf("Nothing to do...\n")
			return
		}

		// Start Horizon
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// Use the current context in kubeconfig
		rc, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

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
			Name:      init_secretName,
			Namespace: namespace,
			Type:      secretType,
		})
		secret.AddData(map[string][]byte{
			"ADMIN_PASSWORD":    []byte(init_secretAdminPassword),
			"POSTGRES_PASSWORD": []byte(init_secretPostgresPassword),
			"USER_PASSWORD":     []byte(init_secretUserPassword),
			"HUB_PASSWORD":      []byte(init_secretBlackduckPassword),
		})
		environmentDeployer.AddSecret(secret)

		// Deploy Resources for the Synopsys Operator
		err = environmentDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying Environment with Horizon : %s\n", err)
			return
		}

		// Create a Horizon Deployer for the Synopsys Operator
		synopsysOperatorDeployer, err := deployer.NewDeployer(rc)

		// Add the Replication Controller to the Deployer
		var synopsysOperatorRCReplicas int32 = 1
		synopsysOperatorRC := horizoncomponents.NewReplicationController(horizonapi.ReplicationControllerConfig{
			APIVersion: "v1",
			//ClusterName:  "string",
			Name:      "synopsys-operator",
			Namespace: namespace,
			Replicas:  &synopsysOperatorRCReplicas,
			//ReadySeconds: "int32",
		})

		synopsysOperatorRC.AddLabelSelectors(map[string]string{"name": "synopsys-operator"})

		synopsysOperatorPod := horizoncomponents.NewPod(horizonapi.PodConfig{
			APIVersion: "v1",
			//ClusterName:            "string",
			Name:           "synopsys-operator",
			Namespace:      namespace,
			ServiceAccount: "synopsys-operator",
			//RestartPolicy:          "RestartPolicyType",
			//TerminationGracePeriod: "*int64",
			//ActiveDeadline:         "*int64",
			//Node:                   "string",
			//FSGID:                  "*int64",
			//Hostname:               "string",
			//SchedulerName:          "string",
			//DNSPolicy:              "DNSPolicType",
			//PriorityValue:          "*int32",
			//PriorityClass:          "string",
			//SELinux:                "*SELinuxType",
			//RunAsUser:              "*int64",
			//RunAsGroup:             "*int64",
			//ForceNonRoot:           "*bool",
		})

		synopsysOperatorPodLabels := map[string]string{"name": "synopsys-operator"}

		synopsysOperatorContainer := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
			Name:       "synopsys-operator",
			Args:       []string{"/etc/synopsys-operator/config.json"},
			Command:    []string{"./operator"},
			Image:      init_synopsysOperatorImage,
			PullPolicy: horizonapi.PullAlways,
			//MinCPU:                   "string",
			//MaxCPU:                   "string",
			//MinMem:                   "string",
			//MaxMem:                   "string",
			//Privileged:               "*bool",
			//AllowPrivilegeEscalation: "*bool",
			//ReadOnlyFS:               "*bool",
			//ForceNonRoot:             "*bool",
			//SELinux:                  "*SELinuxType",
			//UID:                      "*int64",
			//AllocateStdin:            "bool",
			//StdinOnce:                "bool",
			//AllocateTTY:              "bool",
			//WorkingDirectory:         "string",
			//TerminationMsgPath:       "string",
			//TerminationMsgPolicy:     "TerminationMessagePolicyType",
		})
		synopsysOperatorContainer.AddPort(horizonapi.PortConfig{
			//Name:          "string",
			//Protocol:      "ProtocolType",
			//IP:            "string",
			//HostPort:      "string",
			ContainerPort: "8080",
		})
		synopsysOperatorContainer.AddEnv(horizonapi.EnvConfig{
			NameOrPrefix: "REGISTRATION_KEY",
			Type:         horizonapi.EnvVal,
			KeyOrVal:     init_blackduckRegistrationKey,
			//FromName:     "string",
		})
		synopsysOperatorContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
			MountPath: "/etc/synopsys-operator",
			//Propagation: "*MountPropagationType",
			Name: "synopsys-operator",
			//SubPath:     "string",
			//ReadOnly:    "*bool",
		})

		synopsysOperatorContainerUI := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
			Name: "synopsys-operator-ui",
			//Args:                     "[]string",
			Command:    []string{"./app"},
			Image:      init_synopsysOperatorImage,
			PullPolicy: horizonapi.PullAlways,
			//MinCPU:                   "string",
			//MaxCPU:                   "string",
			//MinMem:                   "string",
			//MaxMem:                   "string",
			//Privileged:               "*bool",
			//AllowPrivilegeEscalation: "*bool",
			//ReadOnlyFS:               "*bool",
			//ForceNonRoot:             "*bool",
			//SELinux:                  "*SELinuxType",
			//UID:                      "*int64",
			//AllocateStdin:            "bool",
			//StdinOnce:                "bool",
			//AllocateTTY:              "bool",
			//WorkingDirectory:         "string",
			//TerminationMsgPath:       "string",
			//TerminationMsgPolicy:     "TerminationMessagePolicyType",
		})
		synopsysOperatorContainerUI.AddPort(horizonapi.PortConfig{
			//Name:          "string",
			//Protocol:      "ProtocolType",
			//IP:            "string",
			//HostPort:      "string",
			ContainerPort: "3000",
		})
		synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
			NameOrPrefix: "ADDR",
			Type:         horizonapi.EnvVal,
			KeyOrVal:     "0.0.0.0",
			//FromName:     "string",
		})
		synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
			NameOrPrefix: "PORT",
			Type:         horizonapi.EnvVal,
			KeyOrVal:     "3000",
			//FromName:     "string",
		})
		synopsysOperatorContainerUI.AddEnv(horizonapi.EnvConfig{
			NameOrPrefix: "GO_ENV",
			Type:         horizonapi.EnvVal,
			KeyOrVal:     "development",
			//FromName:     "string",
		})

		// Create config map volume
		var synopsysOperatorVolumeDefaultMode int32 = 420
		synopsysOperatorVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
			VolumeName:      "synopsys-operator",
			MapOrSecretName: "synopsys-operator",
			//Items:           "map[string]KeyAndMode",
			DefaultMode: &synopsysOperatorVolumeDefaultMode,
			//Required:        "*bool",
		})

		synopsysOperatorPod.AddLabels(synopsysOperatorPodLabels)
		synopsysOperatorPod.AddContainer(synopsysOperatorContainer)
		synopsysOperatorPod.AddContainer(synopsysOperatorContainerUI)
		synopsysOperatorPod.AddVolume(synopsysOperatorVolume)
		synopsysOperatorRC.AddPod(synopsysOperatorPod)

		synopsysOperatorDeployer.AddReplicationController(synopsysOperatorRC)

		// Add the Service to the Deployer
		synopsysOperatorService := horizoncomponents.NewService(horizonapi.ServiceConfig{
			APIVersion: "v1",
			//ClusterName:              "string",
			Name:      "synopsys-operator",
			Namespace: namespace,
			//ExternalName:             "string",
			//IPServiceType:            "ClusterIPServiceType",
			//ClusterIP:                "string",
			//PublishNotReadyAddresses: "bool",
			//TrafficPolicy:            "TrafficPolicyType",
			//Affinity:                 "string",
		})

		synopsysOperatorService.AddSelectors(map[string]string{"name": "synopsys-operator"})
		synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
			Name:       "synopsys-operator-ui",
			Port:       3000,
			TargetPort: "3000",
			//NodePort:   "int32",
			Protocol: horizonapi.ProtocolTCP,
		})
		synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
			Name:       "synopsys-operator-ui-standard-port",
			Port:       80,
			TargetPort: "3000",
			//NodePort:   "int32",
			Protocol: horizonapi.ProtocolTCP,
		})
		synopsysOperatorService.AddPort(horizonapi.ServicePortConfig{
			Name:       "synopsys-operator",
			Port:       8080,
			TargetPort: "8080",
			//NodePort:   "int32",
			Protocol: horizonapi.ProtocolTCP,
		})

		synopsysOperatorDeployer.AddService(synopsysOperatorService)

		// Config Map
		synopsysOperatorConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
			APIVersion: "v1",
			//ClusterName: "string",
			Name:      "synopsys-operator",
			Namespace: namespace,
		})

		synopsysOperatorConfigMap.AddData(map[string]string{"config.json": fmt.Sprintf("{\"OperatorTimeBombInSeconds\":\"315576000\", \"DryRun\": false, \"LogLevel\": \"debug\", \"Namespace\": \"%s\", \"Threadiness\": 5, \"PostgresRestartInMins\": 10, \"NFSPath\" : \"/kubenfs\"}", namespace)})

		synopsysOperatorDeployer.AddConfigMap(synopsysOperatorConfigMap)

		// Service Account
		synopsysOperatorServiceAccount := horizoncomponents.NewServiceAccount(horizonapi.ServiceAccountConfig{
			APIVersion: "v1",
			//ClusterName:    "string",
			Name:      "synopsys-operator",
			Namespace: namespace,
			//AutomountToken: "*bool",
		})

		synopsysOperatorDeployer.AddServiceAccount(synopsysOperatorServiceAccount)

		// Cluster Role Binding
		synopsysOperatorClusterRoleBinding := horizoncomponents.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			//ClusterName: "string",
			Name:      "synopsys-operator-admin",
			Namespace: namespace,
		})
		synopsysOperatorClusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
			Kind: "ServiceAccount",
			//APIGroup:  "string",
			Name:      "synopsys-operator",
			Namespace: namespace,
		})
		synopsysOperatorClusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
			APIGroup: "",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		})

		synopsysOperatorDeployer.AddClusterRoleBinding(synopsysOperatorClusterRoleBinding)

		// Deploy Resources for the Synopsys Operator
		err = synopsysOperatorDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying Synopsys Operator with Horizon : %s\n", err)
			return
		}

		// Create a Horizon Deployer for Prometheus
		prometheusDeployer, err := deployer.NewDeployer(rc)

		// Add Service for Prometheus
		prometheusService := horizoncomponents.NewService(horizonapi.ServiceConfig{
			APIVersion: "v1",
			//ClusterName:              "string",
			Name:      "prometheus",
			Namespace: namespace,
			//ExternalName:             "string",
			IPServiceType: horizonapi.ClusterIPServiceTypeNodePort,
			//ClusterIP:                "string",
			//PublishNotReadyAddresses: "bool",
			//TrafficPolicy:            "TrafficPolicyType",
			//Affinity:                 "string",
		})
		prometheusService.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
		prometheusService.AddLabels(map[string]string{"name": "prometheus"})
		prometheusService.AddSelectors(map[string]string{"app": "prometheus"})
		prometheusService.AddPort(horizonapi.ServicePortConfig{
			Name:       "prometheus",
			Port:       9090,
			TargetPort: "9090",
			//NodePort:   "int32",
			Protocol: horizonapi.ProtocolTCP,
		})

		prometheusDeployer.AddService(prometheusService)

		// Deployment
		var prometheusDeploymentReplicas int32 = 1
		prometheusDeployment := horizoncomponents.NewDeployment(horizonapi.DeploymentConfig{
			APIVersion: "extensions/v1beta1",
			//ClusterName:             "string",
			Name:      "prometheus",
			Namespace: namespace,
			Replicas:  &prometheusDeploymentReplicas,
			//Recreate:                "bool",
			//MaxUnavailable:          "string",
			//MaxExtra:                "string",
			//MinReadySeconds:         "int32",
			//RevisionHistoryLimit:    "*int32",
			//Paused:                  "bool",
			//ProgressDeadlineSeconds: "*int32",
		})
		prometheusDeployment.AddMatchLabelsSelectors(map[string]string{"app": "prometheus"})

		prometheusPod := horizoncomponents.NewPod(horizonapi.PodConfig{
			APIVersion: "v1",
			//ClusterName          :  "string",
			Name:      "prometheus",
			Namespace: namespace,
			//ServiceAccount       :  "string",
			//RestartPolicy        :  "RestartPolicyType",
			//TerminationGracePeriod : "*int64",
			//ActiveDeadline       :  "*int64",
			//Node                 :  "string",
			//FSGID                :  "*int64",
			//Hostname             :  "string",
			//SchedulerName        :  "string",
			//DNSPolicy           :   "DNSPolicyType",
			//PriorityValue       :   "*int32",
			//PriorityClass        :  "string",
			//SELinux              :  "*SELinuxType",
			//RunAsUser            :  "*int64",
			//RunAsGroup           :  "*int64",
			//ForceNonRoot         :  "*bool",
		})

		prometheusContainer := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
			Name: "prometheus",
			Args: []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/"},
			//Command:                  "[]string",
			Image: init_promethiusImage,
			//PullPolicy:               "PullPolicyType",
			//MinCPU:                   "string",
			//MaxCPU:                   "string",
			//MinMem:                   "string",
			//MaxMem:                   "string",
			//Privileged:               "*bool",
			//AllowPrivilegeEscalation: "*bool",
			//ReadOnlyFS:               "*bool",
			//ForceNonRoot:             "*bool",
			//SELinux:                  "*SELinuxType",
			//UID:                      "*int64",
			//AllocateStdin:            "bool",
			//StdinOnce:                "bool",
			//AllocateTTY:              "bool",
			//WorkingDirectory:         "string",
			//TerminationMsgPath:       "string",
			//TerminationMsgPolicy:     "TerminationMessagePolicyType",
		})

		prometheusContainer.AddPort(horizonapi.PortConfig{
			Name: "web",
			//Protocol:      "ProtocolType",
			//IP:            "string",
			//HostPort:      "string",
			ContainerPort: "9090",
		})

		prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
			MountPath: "/data",
			//Propagation: "*MountPropagationType",
			Name: "data",
			//SubPath:     "string",
			//ReadOnly:    "*bool",
		})
		prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
			MountPath: "/etc/prometheus",
			//Propagation: "*MountPropagationType",
			Name: "config-volume",
			//SubPath:     "string",
			//ReadOnly:    "*bool",
		})

		prometheusEmptyDirVolume, err := horizoncomponents.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
			VolumeName: "data",
			//Medium:     "StorageMediumType",
			//SizeLimit:  "string",
		})
		prometheusConfigMapVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
			VolumeName:      "config-volume",
			MapOrSecretName: "prometheus",
			//Items:           "map[string]KeyAndMode",
			//DefaultMode:     "*int32",
			//Required:        "*bool",
		})

		prometheusPod.AddContainer(prometheusContainer)
		prometheusPod.AddVolume(prometheusEmptyDirVolume)
		prometheusPod.AddVolume(prometheusConfigMapVolume)
		prometheusDeployment.AddPod(prometheusPod)

		prometheusDeployer.AddDeployment(prometheusDeployment)

		// Config Map
		prometheusConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
			APIVersion: "v1",
			//ClusterName: "string",
			Name:      "prometheus",
			Namespace: namespace,
		})
		prometheusConfigMap.AddData(map[string]string{"prometheus.yml": "{'global':{'scrape_interval':'5s'},'scrape_configs':[{'job_name':'synopsys-operator-scrape','scrape_interval':'5s','static_configs':[{'targets':['synopsys-operator:8080', 'synopsys-operator-ui:3000']}]}]}"})
		prometheusDeployer.AddConfigMap(prometheusConfigMap)

		// Deploy Resources for Prometheus
		err = prometheusDeployer.Run()
		if err != nil {
			fmt.Printf("Error deploying Prometheus with Horizon : %s\n", err)
			return
		}

		// secret link stuff

		// expose the routes

	},
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&init_synopsysOperatorImage, "synopsys-operator-image", "i", init_synopsysOperatorImage, "synopsys operator image URL")
	initCmd.Flags().StringVarP(&init_promethiusImage, "promethius-image", "p", init_promethiusImage, "promethius image URL")
	initCmd.Flags().StringVarP(&init_blackduckRegistrationKey, "blackduck-registration-key", "k", init_blackduckRegistrationKey, "key to register with KnowledgeBase")
	initCmd.Flags().StringVarP(&init_dockerConfigPath, "docker-config", "d", init_dockerConfigPath, "path to docker config (image pull secrets etc)")

	initCmd.Flags().StringVar(&init_secretName, "secret-name", init_secretName, "name of kubernetes secret for postgres and blackduck")
	initCmd.Flags().StringVar(&init_secretType, "secret-type", init_secretType, "type of kubernetes secret for postgres and blackduck")
	initCmd.Flags().StringVar(&init_secretAdminPassword, "admin-password", init_secretAdminPassword, "postgres admin password")
	initCmd.Flags().StringVar(&init_secretPostgresPassword, "postgres-password", init_secretPostgresPassword, "postgres password")
	initCmd.Flags().StringVar(&init_secretUserPassword, "user-password", init_secretUserPassword, "postgres user password")
	initCmd.Flags().StringVar(&init_secretBlackduckPassword, "blackduck-password", init_secretBlackduckPassword, "blackduck password for 'sysadmin' account")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
