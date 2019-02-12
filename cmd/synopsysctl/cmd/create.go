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
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys Resource in your cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")
	},
}

var createBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Create an instance of a Blackduck",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the Blackduck
		deployCRDNamespace(restconfig)

		// Create Spec for a Blackduck CRD
		blackduck := &blackduckv1.Blackduck{}
		populateBlackduckConfig(blackduck)
		fmt.Printf("%+v\n", blackduck)

		blackduckClient, err := blackduckclientset.NewForConfig(restconfig)

		_, err = blackduckClient.SynopsysV1().Blackducks(namespace).Create(blackduck)
		if err != nil {
			fmt.Printf("Error creating the Blackduck : %s\n", err)
			return
		}
	},
}

var createOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Create an instance of OpsSight",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the OpsSight
		deployCRDNamespace(restconfig)

		// Create OpsSight Spec
		opssight := &opssightv1.OpsSight{}
		populateOpssightConfig(opssight)
		opssightClient, err := opssightclientset.NewForConfig(restconfig)
		_, err = opssightClient.SynopsysV1().OpsSights(namespace).Create(opssight)
		if err != nil {
			fmt.Printf("Error creating the OpsSight : %s\n", err)
			return
		}
	},
}

var createAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Create an instance of Alert",
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the Alert
		deployCRDNamespace(restconfig)

		// Create Alert Spec
		alert := &alertv1.Alert{}
		populateAlertConfig(alert)
		alertClient, err := alertclientset.NewForConfig(restconfig)
		_, err = alertClient.SynopsysV1().Alerts(namespace).Create(alert)
		if err != nil {
			fmt.Printf("Error creating the Alert : %s\n", err)
			return
		}
	},
}

func deployCRDNamespace(restconfig *rest.Config) {

	// Create Horizon Deployer
	namespaceDeployer, err := deployer.NewDeployer(restconfig)
	ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
		// APIVersion:  "string",
		// ClusterName: "string",
		Name:      namespace,
		Namespace: namespace,
	})
	namespaceDeployer.AddNamespace(ns)
	err = namespaceDeployer.Run()
	if err != nil {
		fmt.Printf("Error deploying namespace with Horizon : %s\n", err)
		return
	}
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Flags
	createBlackduckCmd.Flags().StringVar(&create_blackduck_size, "size", create_blackduck_size, "Blackduck size - small, medium, large")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_dbPrototype, "db-prototype", create_blackduck_dbPrototype, "TODO")
	//TODO - var create_blackduck_externalPostgres = &blackduckv1.PostgresExternalDBConfig{}
	createBlackduckCmd.Flags().StringVar(&create_blackduck_externalPostgres_postgresHost, "external-postgres-host", create_blackduck_externalPostgres_postgresHost, "TODO")
	createBlackduckCmd.Flags().IntVar(&create_blackduck_externalPostgres_postgresPort, "external-postgres-port", create_blackduck_externalPostgres_postgresPort, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_externalPostgres_postgresAdmin, "external-postgres-admin", create_blackduck_externalPostgres_postgresAdmin, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_externalPostgres_postgresUser, "external-postgres-user", create_blackduck_externalPostgres_postgresUser, "TODO")
	createBlackduckCmd.Flags().BoolVar(&create_blackduck_externalPostgres_postgresSsl, "external-postgres-ssl", create_blackduck_externalPostgres_postgresSsl, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_externalPostgres_postgresAdminPassword, "external-postgres-admin-password", create_blackduck_externalPostgres_postgresAdminPassword, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_externalPostgres_postgresUserPassword, "external-postgres-user-password", create_blackduck_externalPostgres_postgresUserPassword, "TODO")

	createBlackduckCmd.Flags().StringVar(&create_blackduck_pvcStorageClass, "pvc-storage-class", create_blackduck_pvcStorageClass, "TODO")
	createBlackduckCmd.Flags().BoolVar(&create_blackduck_livenessProbes, "liveness-probes", create_blackduck_livenessProbes, "Enable liveness probes")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_scanType, "scan-type", create_blackduck_scanType, "TODO")
	createBlackduckCmd.Flags().BoolVar(&create_blackduck_persistentStorage, "persistent-storage", create_blackduck_persistentStorage, "Enable persistent storage")
	//TODO - var create_blackduck_PVC = []blackduckv1.PVC{}
	createBlackduckCmd.Flags().StringVar(&create_blackduck_certificateName, "db-certificate-name", create_blackduck_certificateName, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_certificate, "certificate", create_blackduck_certificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_certificateKey, "certificate-key", create_blackduck_certificateKey, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_proxyCertificate, "proxy-certificate", create_blackduck_proxyCertificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_type, "type", create_blackduck_type, "TODO")
	createBlackduckCmd.Flags().StringVar(&create_blackduck_desiredState, "desired-state", create_blackduck_desiredState, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&create_blackduck_environs, "environs", create_blackduck_environs, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&create_blackduck_imageRegistries, "image-registries", create_blackduck_imageRegistries, "List of image registries")
	//TODO - var create_blackduck_imageUIDMap = map[string]int64{}
	createBlackduckCmd.Flags().StringVar(&create_blackduck_licenseKey, "license-key", create_blackduck_licenseKey, "TODO")
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Flags
	//TODO - var create_opssight_perceptor = &opssightv1.Perceptor{}
	//TODO - var create_opssight_scannerPod = &opssightv1.ScannerPod{}
	//TODO - var create_opssight_perceiver = &opssightv1.Perceiver{}
	//TODO - var create_opssight_prometheus = &opssightv1.Prometheus{}
	createOpsSightCmd.Flags().BoolVar(&create_opssight_enableSkyfire, "enable-skyfire", create_opssight_enableSkyfire, "TODO")
	//TODO - var create_opssight_skyfire = &opssightv1.Skyfire{}
	//TODO - var create_opssight_blackduck = &blackduckv1.Blackduck{}
	createOpsSightCmd.Flags().BoolVar(&create_opssight_enableMetrics, "enable-metrics", create_opssight_enableMetrics, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_defaultCPU, "default-cpu", create_opssight_defaultCPU, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_defaultMem, "default-mem", create_opssight_defaultMem, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_logLevel, "log-level", create_opssight_logLevel, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_configMapName, "config-map-name", create_opssight_configMapName, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_secretName, "secret-name", create_opssight_secretName, "TODO")
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Flags
	createAlertCmd.Flags().StringVar(&create_alert_registry, "alert-registry", create_alert_registry, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_imagePath, "image-path", create_alert_imagePath, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_alertImageName, "alert-image-name", create_alert_alertImageName, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_alertImageVersion, "alert-image-version", create_alert_alertImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_cfsslImageName, "cfssl-image-name", create_alert_cfsslImageName, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_cfsslImageVersion, "cfssl-image-version", create_alert_cfsslImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_blackduckHost, "blackduck-host", create_alert_blackduckHost, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_blackduckUser, "blackduck-user", create_alert_blackduckUser, "TODO")
	createAlertCmd.Flags().IntVar(&create_alert_blackduckPort, "blackduck-port", create_alert_blackduckPort, "TODO")
	createAlertCmd.Flags().IntVar(&create_alert_port, "port", create_alert_port, "TODO")
	createAlertCmd.Flags().BoolVar(&create_alert_standAlone, "stand-alone", create_alert_standAlone, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_alertMemory, "alert-memory", create_alert_alertMemory, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_cfsslMemory, "cfssl-memory", create_alert_cfsslMemory, "TODO")
	createAlertCmd.Flags().StringVar(&create_alert_state, "alert-state", create_alert_state, "TODO")
	createCmd.AddCommand(createAlertCmd)
}

func populateBlackduckConfig(bd *blackduckv1.Blackduck) {
	// Add Meta Data
	bd.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}

	// Get Default Blackduck Spec
	bdDefaultSpec := crddefaults.GetHubDefaultPersistentStorage()

	// Update values with User input
	bdDefaultSpec.Namespace = namespace
	bdDefaultSpec.Size = create_blackduck_size
	bdDefaultSpec.DbPrototype = create_blackduck_dbPrototype
	bdDefaultSpec.ExternalPostgres.PostgresHost = create_blackduck_externalPostgres_postgresHost
	bdDefaultSpec.ExternalPostgres.PostgresPort = create_blackduck_externalPostgres_postgresPort
	bdDefaultSpec.ExternalPostgres.PostgresAdmin = create_blackduck_externalPostgres_postgresAdmin
	bdDefaultSpec.ExternalPostgres.PostgresUser = create_blackduck_externalPostgres_postgresUser
	bdDefaultSpec.ExternalPostgres.PostgresSsl = create_blackduck_externalPostgres_postgresSsl
	bdDefaultSpec.ExternalPostgres.PostgresAdminPassword = create_blackduck_externalPostgres_postgresAdminPassword
	bdDefaultSpec.ExternalPostgres.PostgresUserPassword = create_blackduck_externalPostgres_postgresUserPassword
	bdDefaultSpec.PVCStorageClass = create_blackduck_pvcStorageClass
	bdDefaultSpec.LivenessProbes = create_blackduck_livenessProbes
	bdDefaultSpec.ScanType = create_blackduck_scanType
	bdDefaultSpec.PersistentStorage = create_blackduck_persistentStorage
	bdDefaultSpec.PVC = create_blackduck_PVC
	bdDefaultSpec.CertificateName = create_blackduck_certificateName
	bdDefaultSpec.Certificate = create_blackduck_certificate
	bdDefaultSpec.CertificateKey = create_blackduck_certificateKey
	bdDefaultSpec.ProxyCertificate = create_blackduck_proxyCertificate
	bdDefaultSpec.Type = create_blackduck_type
	bdDefaultSpec.DesiredState = create_blackduck_desiredState
	bdDefaultSpec.Environs = create_blackduck_environs
	bdDefaultSpec.ImageRegistries = create_blackduck_imageRegistries
	bdDefaultSpec.ImageUIDMap = create_blackduck_imageUIDMap
	bdDefaultSpec.LicenseKey = create_blackduck_licenseKey

	// Add updated spec
	bd.Spec = *bdDefaultSpec
}

func populateOpssightConfig(opssight *opssightv1.OpsSight) {
	// Add Meta Data
	opssight.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}

	// Get Default OpsSight Spec
	opssightDefaultSpec := crddefaults.GetOpsSightDefaultValueWithDisabledHub()

	// Update values with User input
	opssightDefaultSpec.Namespace = namespace
	//TODO - opssightDefaultSpec.Perceptor{}
	//TODO - opssightDefaultSpec.ScannerPod{}
	//TODO - opssightDefaultSpec.Perceiver{}
	//TODO - opssightDefaultSpec.Prometheus{}
	opssightDefaultSpec.EnableSkyfire = create_opssight_enableSkyfire
	//TODO - opssightDefaultSpec.Skyfire{}
	//TODO - opssightDefaultSpec.Blackduck{}
	opssightDefaultSpec.EnableMetrics = create_opssight_enableMetrics
	opssightDefaultSpec.DefaultCPU = create_opssight_defaultCPU
	opssightDefaultSpec.DefaultMem = create_opssight_defaultMem
	opssightDefaultSpec.LogLevel = create_opssight_logLevel
	opssightDefaultSpec.ConfigMapName = create_opssight_configMapName
	opssightDefaultSpec.SecretName = create_opssight_secretName

	// Add updated spec
	opssight.Spec = *opssightDefaultSpec
}

func populateAlertConfig(alert *alertv1.Alert) {
	// Add Meta Data
	alert.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}

	// Get Default Alert Spec
	alertDefaultSpec := crddefaults.GetAlertDefaultValue()

	// Update values with User input
	alertDefaultSpec.Namespace = namespace
	alertDefaultSpec.Registry = create_alert_registry
	alertDefaultSpec.ImagePath = create_alert_imagePath
	alertDefaultSpec.AlertImageName = create_alert_alertImageName
	alertDefaultSpec.AlertImageVersion = create_alert_alertImageVersion
	alertDefaultSpec.CfsslImageName = create_alert_cfsslImageName
	alertDefaultSpec.CfsslImageVersion = create_alert_cfsslImageVersion
	alertDefaultSpec.BlackduckHost = create_alert_blackduckHost
	alertDefaultSpec.BlackduckUser = create_alert_blackduckUser
	alertDefaultSpec.BlackduckPort = &create_alert_blackduckPort
	alertDefaultSpec.Port = &create_alert_port
	alertDefaultSpec.StandAlone = &create_alert_standAlone
	alertDefaultSpec.AlertMemory = create_alert_alertMemory
	alertDefaultSpec.CfsslMemory = create_alert_cfsslMemory
	alertDefaultSpec.State = create_alert_state

	// Add updated spec
	alert.Spec = *alertDefaultSpec
}
