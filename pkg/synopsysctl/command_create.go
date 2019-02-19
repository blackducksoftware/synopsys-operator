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
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	bdutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Commands
//var createCmd *cobra.Command

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
		blackduck.ObjectMeta = metav1.ObjectMeta{
			Name: namespace,
		}
		defaultBlackduckSpec := crddefaults.GetHubDefaultValue()
		flagset := cmd.Flags()
		flagset.VisitAll(checkBlackduckFlags)
		blackduck.Spec = *defaultBlackduckSpec

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
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceptor_name, "perceptor-name", create_opssight_perceptor_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceptor_image, "perceptor-image", create_opssight_perceptor_image, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_port, "perceptor-port", create_opssight_perceptor_port, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_checkForStalledScansPauseHours, "perceptor-check-scan-hours", create_opssight_perceptor_checkForStalledScansPauseHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_stalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", create_opssight_perceptor_stalledScanClientTimeoutHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_modelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", create_opssight_perceptor_modelMetricsPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_unknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", create_opssight_perceptor_unknownImagePauseMilliseconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceptor_clientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", create_opssight_perceptor_clientTimeoutMilliseconds, "TODO")
	//TODO - var create_opssight_scannerPod = &opssightv1.ScannerPod{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_name, "scannerpod-name", create_opssight_scannerPod_name, "TODO")
	//TODO - var create_opssight_scannerPod_scanner = &opssightv1.Scanner{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_scanner_name, "scannerpod-scanner-name", create_opssight_scannerPod_scanner_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_scanner_image, "scannerpod-scanner-image", create_opssight_scannerPod_scanner_image, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_scannerPod_scanner_port, "scannerpod-scanner-port", create_opssight_scannerPod_scanner_port, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_scannerPod_scanner_clientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", create_opssight_scannerPod_scanner_clientTimeoutSeconds, "TODO")
	//TODO - var create_opssight_scannerPod_imageFacade = &opssightv1.ImageFacade{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_imageFacade_name, "scannerpod-imagefacade-name", create_opssight_scannerPod_imageFacade_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_imageFacade_image, "scannerpod-imagefacade-image", create_opssight_scannerPod_imageFacade_image, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_scannerPod_imageFacade_port, "scannerpod-imagefacade-port", create_opssight_scannerPod_imageFacade_port, "TODO")
	//TODO - var create_opssight_scannerPod_imageFacade_internalRegistries = []opssightv1.RegistryAuth{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_imageFacade_imagePullerType, "scannerpod-imagefacade-image-puller-type", create_opssight_scannerPod_imageFacade_imagePullerType, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_imageFacade_serviceAccount, "scannerpod-imagefacade-service-account", create_opssight_scannerPod_imageFacade_serviceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_scannerPod_replicaCount, "scannerpod-replica-count", create_opssight_scannerPod_replicaCount, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_scannerPod_imageDirectory, "scannerpod-image-directory", create_opssight_scannerPod_imageDirectory, "TODO")
	//TODO - var create_opssight_perceiver = &opssightv1.Perceiver{}
	createOpsSightCmd.Flags().BoolVar(&create_opssight_perceiver_enableImagePerceiver, "enable-image-perceiver", create_opssight_perceiver_enableImagePerceiver, "TODO")
	createOpsSightCmd.Flags().BoolVar(&create_opssight_perceiver_enablePodPerceiver, "enable-pod-perceiver", create_opssight_perceiver_enablePodPerceiver, "TODO")
	//TODO - var create_opssight_perceiver_imagePerceiver = &opssightv1.ImagePerceiver{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_imagePerceiver_name, "imageperceiver-name", create_opssight_perceiver_imagePerceiver_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_imagePerceiver_image, "imageperceiver-image", create_opssight_perceiver_imagePerceiver_image, "TODO")
	//TODO - var create_opssight_perceiver_podPerceiver = &opssightv1.PodPerceiver{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_podPerceiver_name, "podperceiver-name", create_opssight_perceiver_podPerceiver_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_podPerceiver_image, "podperceiver-image", create_opssight_perceiver_podPerceiver_image, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_podPerceiver_namespaceFilter, "podperceiver-namespace-filter", create_opssight_perceiver_podPerceiver_namespaceFilter, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceiver_annotationIntervalSeconds, "perceiver-annotation-interval-seconds", create_opssight_perceiver_annotationIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceiver_dumpIntervalMinutes, "perceiver-dump-interval-minutes", create_opssight_perceiver_dumpIntervalMinutes, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_perceiver_serviceAccount, "perceiver-service-account", create_opssight_perceiver_serviceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_perceiver_port, "perceiver-port", create_opssight_perceiver_port, "TODO")
	//TODO - var create_opssight_prometheus = &opssightv1.Prometheus{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_prometheus_name, "prometheus-name", create_opssight_prometheus_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_prometheus_name, "prometheus-image", create_opssight_prometheus_name, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_prometheus_port, "prometheus-port", create_opssight_prometheus_port, "TODO")
	createOpsSightCmd.Flags().BoolVar(&create_opssight_enableSkyfire, "enable-skyfire", create_opssight_enableSkyfire, "TODO")
	//TODO - var create_opssight_skyfire = &opssightv1.Skyfire{}
	createOpsSightCmd.Flags().StringVar(&create_opssight_skyfire_name, "skyfire-name", create_opssight_skyfire_name, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_skyfire_image, "skyfire-image", create_opssight_skyfire_image, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_port, "skyfire-port", create_opssight_skyfire_port, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_prometheusPort, "skyfire-prometheus-port", create_opssight_skyfire_prometheusPort, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_skyfire_serviceAccount, "skyfire-service-account", create_opssight_skyfire_serviceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_hubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", create_opssight_skyfire_hubClientTimeoutSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_hubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", create_opssight_skyfire_hubDumpPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_kubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", create_opssight_skyfire_kubeDumpIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_skyfire_perceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", create_opssight_skyfire_perceptorDumpIntervalSeconds, "TODO")
	//TODO - var create_opssight_blackduck = &opssightv1.Blackduck{}
	createOpsSightCmd.Flags().StringSliceVar(&create_opssight_blackduck_hosts, "blackduck-hosts", create_opssight_blackduck_hosts, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_blackduck_user, "blackduck-user", create_opssight_blackduck_user, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_port, "blackduck-port", create_opssight_blackduck_port, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_concurrentScanLimit, "blackduck-concurrent-scan-limit", create_opssight_blackduck_concurrentScanLimit, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_totalScanLimit, "blackduck-total-scan-limit", create_opssight_blackduck_totalScanLimit, "TODO")
	createOpsSightCmd.Flags().StringVar(&create_opssight_blackduck_passwordEnvVar, "blackduck-password-environment-variable", create_opssight_blackduck_passwordEnvVar, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_initialCount, "blackduck-initial-count", create_opssight_blackduck_initialCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_maxCount, "blackduck-max-count", create_opssight_blackduck_maxCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&create_opssight_blackduck_deleteHubThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", create_opssight_blackduck_deleteHubThresholdPercentage, "TODO")
	//TODO - var create_opssight_blackduck_blackduckSpec = &blackduckv1.BlackduckSpec{}

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

func checkBlackduckFlags(f *pflag.Flag) {
	if f.Changed {
		fmt.Printf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "namespace":
			defaultBlackduckSpec.Namespace = namespace
		case "size":
			defaultBlackduckSpec.Size = create_blackduck_size
		case "db-prototype":
			defaultBlackduckSpec.DbPrototype = create_blackduck_dbPrototype
		case "external-postgres-host":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresHost = create_blackduck_externalPostgres_postgresHost
		case "external-postgres-port":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresPort = create_blackduck_externalPostgres_postgresPort
		case "external-postgres-admin":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdmin = create_blackduck_externalPostgres_postgresAdmin
		case "external-postgres-user":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUser = create_blackduck_externalPostgres_postgresUser
		case "external-postgres-ssl":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresSsl = create_blackduck_externalPostgres_postgresSsl
		case "external-postgres-admin-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdminPassword = create_blackduck_externalPostgres_postgresAdminPassword
		case "external-postgres-user-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUserPassword = create_blackduck_externalPostgres_postgresUserPassword
		case "pvc-storage-class":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.PVCStorageClass = create_blackduck_pvcStorageClass
		case "liveness-probes":
			defaultBlackduckSpec.LivenessProbes = create_blackduck_livenessProbes
		case "scan-type":
			defaultBlackduckSpec.ScanType = create_blackduck_scanType
		case "persistent-storage":
			defaultBlackduckSpec.PersistentStorage = create_blackduck_persistentStorage
		case "db-certificate-name":
			defaultBlackduckSpec.CertificateName = create_blackduck_certificateName
		case "certificate":
			defaultBlackduckSpec.Certificate = create_blackduck_certificate
		case "certificate-key":
			defaultBlackduckSpec.CertificateKey = create_blackduck_certificateKey
		case "proxy-certificate":
			defaultBlackduckSpec.ProxyCertificate = create_blackduck_proxyCertificate
		case "type":
			defaultBlackduckSpec.Type = create_blackduck_type
		case "desired-state":
			defaultBlackduckSpec.DesiredState = create_blackduck_desiredState
		case "environs":
			defaultBlackduckSpec.Environs = create_blackduck_environs
		case "image-registries":
			defaultBlackduckSpec.ImageRegistries = create_blackduck_imageRegistries
		case "license-key":
			defaultBlackduckSpec.LicenseKey = create_blackduck_licenseKey
		default:
			fmt.Printf("Flag Not Found: %s\n", f.Name)
		}
	}
	fmt.Printf("Flag %s: UNCHANGED\n", f.Name)
}

func populateOpssightConfig(opssight *opssightv1.OpsSight) {
	// Add Meta Data
	opssight.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
	}

	// Get Default OpsSight Spec
	opssightDefaultSpec := bdutil.GetOpsSightDefaultValueWithDisabledHub()

	// Update values with User input
	opssightDefaultSpec.Namespace = namespace
	opssightDefaultSpec.Perceptor = create_opssight_perceptor
	opssightDefaultSpec.Perceptor.Name = create_opssight_perceptor_name
	opssightDefaultSpec.Perceptor.Image = create_opssight_perceptor_image
	opssightDefaultSpec.Perceptor.Port = create_opssight_perceptor_port
	opssightDefaultSpec.Perceptor.CheckForStalledScansPauseHours = create_opssight_perceptor_checkForStalledScansPauseHours
	opssightDefaultSpec.Perceptor.StalledScanClientTimeoutHours = create_opssight_perceptor_stalledScanClientTimeoutHours
	opssightDefaultSpec.Perceptor.ModelMetricsPauseSeconds = create_opssight_perceptor_modelMetricsPauseSeconds
	opssightDefaultSpec.Perceptor.UnknownImagePauseMilliseconds = create_opssight_perceptor_unknownImagePauseMilliseconds
	opssightDefaultSpec.Perceptor.ClientTimeoutMilliseconds = create_opssight_perceptor_clientTimeoutMilliseconds
	opssightDefaultSpec.ScannerPod = create_opssight_scannerPod
	opssightDefaultSpec.ScannerPod.Name = create_opssight_scannerPod_name
	opssightDefaultSpec.ScannerPod.Scanner = create_opssight_scannerPod_scanner
	opssightDefaultSpec.ScannerPod.Scanner.Name = create_opssight_scannerPod_scanner_name
	opssightDefaultSpec.ScannerPod.Scanner.Image = create_opssight_scannerPod_scanner_image
	opssightDefaultSpec.ScannerPod.Scanner.Port = create_opssight_scannerPod_scanner_port
	opssightDefaultSpec.ScannerPod.Scanner.ClientTimeoutSeconds = create_opssight_scannerPod_scanner_clientTimeoutSeconds
	opssightDefaultSpec.ScannerPod.ImageFacade = create_opssight_scannerPod_imageFacade
	opssightDefaultSpec.ScannerPod.ImageFacade.Name = create_opssight_scannerPod_imageFacade_name
	opssightDefaultSpec.ScannerPod.ImageFacade.Image = create_opssight_scannerPod_imageFacade_image
	opssightDefaultSpec.ScannerPod.ImageFacade.Port = create_opssight_scannerPod_imageFacade_port
	opssightDefaultSpec.ScannerPod.ImageFacade.InternalRegistries = create_opssight_scannerPod_imageFacade_internalRegistries
	opssightDefaultSpec.ScannerPod.ImageFacade.ImagePullerType = create_opssight_scannerPod_imageFacade_imagePullerType
	opssightDefaultSpec.ScannerPod.ImageFacade.ServiceAccount = create_opssight_scannerPod_imageFacade_serviceAccount
	opssightDefaultSpec.ScannerPod.ReplicaCount = create_opssight_scannerPod_replicaCount
	opssightDefaultSpec.ScannerPod.ImageDirectory = create_opssight_scannerPod_imageDirectory
	opssightDefaultSpec.Perceiver = create_opssight_perceiver
	opssightDefaultSpec.Perceiver.EnableImagePerceiver = create_opssight_perceiver_enableImagePerceiver
	opssightDefaultSpec.Perceiver.EnablePodPerceiver = create_opssight_perceiver_enablePodPerceiver
	opssightDefaultSpec.Perceiver.ImagePerceiver = create_opssight_perceiver_imagePerceiver
	opssightDefaultSpec.Perceiver.ImagePerceiver.Name = create_opssight_perceiver_imagePerceiver_name
	opssightDefaultSpec.Perceiver.ImagePerceiver.Image = create_opssight_perceiver_imagePerceiver_image
	opssightDefaultSpec.Perceiver.PodPerceiver = create_opssight_perceiver_podPerceiver
	opssightDefaultSpec.Perceiver.PodPerceiver.Name = create_opssight_perceiver_podPerceiver_name
	opssightDefaultSpec.Perceiver.PodPerceiver.Image = create_opssight_perceiver_podPerceiver_image
	opssightDefaultSpec.Perceiver.PodPerceiver.NamespaceFilter = create_opssight_perceiver_podPerceiver_namespaceFilter
	opssightDefaultSpec.Perceiver.AnnotationIntervalSeconds = create_opssight_perceiver_annotationIntervalSeconds
	opssightDefaultSpec.Perceiver.DumpIntervalMinutes = create_opssight_perceiver_dumpIntervalMinutes
	opssightDefaultSpec.Perceiver.ServiceAccount = create_opssight_perceiver_serviceAccount
	opssightDefaultSpec.Perceiver.Port = create_opssight_perceiver_port
	opssightDefaultSpec.Prometheus = create_opssight_prometheus
	opssightDefaultSpec.Prometheus.Name = create_opssight_prometheus_name
	opssightDefaultSpec.Prometheus.Image = create_opssight_prometheus_image
	opssightDefaultSpec.Prometheus.Port = create_opssight_prometheus_port
	opssightDefaultSpec.EnableSkyfire = create_opssight_enableSkyfire
	opssightDefaultSpec.Skyfire = create_opssight_skyfire
	opssightDefaultSpec.Skyfire.Name = create_opssight_skyfire_name
	opssightDefaultSpec.Skyfire.Image = create_opssight_skyfire_image
	opssightDefaultSpec.Skyfire.Port = create_opssight_skyfire_port
	opssightDefaultSpec.Skyfire.PrometheusPort = create_opssight_skyfire_prometheusPort
	opssightDefaultSpec.Skyfire.ServiceAccount = create_opssight_skyfire_serviceAccount
	opssightDefaultSpec.Skyfire.HubClientTimeoutSeconds = create_opssight_skyfire_hubClientTimeoutSeconds
	opssightDefaultSpec.Skyfire.HubDumpPauseSeconds = create_opssight_skyfire_hubDumpPauseSeconds
	opssightDefaultSpec.Skyfire.KubeDumpIntervalSeconds = create_opssight_skyfire_kubeDumpIntervalSeconds
	opssightDefaultSpec.Skyfire.PerceptorDumpIntervalSeconds = create_opssight_skyfire_perceptorDumpIntervalSeconds
	opssightDefaultSpec.Blackduck = create_opssight_blackduck
	opssightDefaultSpec.Blackduck.Hosts = create_opssight_blackduck_hosts
	opssightDefaultSpec.Blackduck.User = create_opssight_blackduck_user
	opssightDefaultSpec.Blackduck.Port = create_opssight_blackduck_port
	opssightDefaultSpec.Blackduck.ConcurrentScanLimit = create_opssight_blackduck_concurrentScanLimit
	opssightDefaultSpec.Blackduck.TotalScanLimit = create_opssight_blackduck_totalScanLimit
	opssightDefaultSpec.Blackduck.PasswordEnvVar = create_opssight_blackduck_passwordEnvVar
	opssightDefaultSpec.Blackduck.InitialCount = create_opssight_blackduck_initialCount
	opssightDefaultSpec.Blackduck.MaxCount = create_opssight_blackduck_maxCount
	opssightDefaultSpec.Blackduck.DeleteHubThresholdPercentage = create_opssight_blackduck_deleteHubThresholdPercentage
	opssightDefaultSpec.Blackduck.BlackduckSpec = create_opssight_blackduck_blackduckSpec
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
	alertDefaultSpec := bdutil.GetAlertDefaultValue()

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
