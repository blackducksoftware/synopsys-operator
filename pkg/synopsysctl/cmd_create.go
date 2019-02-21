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
	"encoding/json"
	"fmt"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Default Specs
var defaultBlackduckSpec = &blackduckv1.BlackduckSpec{}
var defaultOpsSightSpec = &opssightv1.OpsSightSpec{}
var defaultAlertSpec = &alertv1.AlertSpec{}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys Resource in your cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		numArgs := 1
		if len(args) < numArgs {
			return fmt.Errorf("Not enough arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating a Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"create"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			fmt.Printf("Error Creating the Resource with KubeCmd: %s\n", err)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

// createCmd represents the create command for Blackduck
var createBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Create an instance of a Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating a Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the Blackduck
		DeployCRDNamespace(restconfig, blackduckName)

		// Read Flags Into Default Blackduck Spec
		defaultBlackduckSpec = crddefaults.GetHubDefaultPersistentStorage()
		flagset := cmd.Flags()
		flagset.VisitAll(checkBlackduckFlags)

		// Set Namespace in Spec
		defaultBlackduckSpec.Namespace = blackduckName

		// Create and Deploy Blackduck CRD
		blackduck := &blackduckv1.Blackduck{
			ObjectMeta: metav1.ObjectMeta{
				Name: blackduckName,
			},
			Spec: *defaultBlackduckSpec,
		}
		blackduckClient, err := blackduckclientset.NewForConfig(restconfig)
		_, err = blackduckClient.SynopsysV1().Blackducks(blackduckName).Create(blackduck)
		if err != nil {
			fmt.Printf("Error creating the Blackduck : %s\n", err)
			return
		}
	},
}

// createCmd represents the create command for OpsSight
var createOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Create an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the OpsSight
		DeployCRDNamespace(restconfig, opsSightName)

		// Read Flags Into Default OpsSight Spec
		defaultOpsSightSpec = crddefaults.GetOpsSightDefaultValueWithDisabledHub()
		flagset := cmd.Flags()
		flagset.VisitAll(checkOpsSightFlags)

		// Set Namespace in Spec
		defaultOpsSightSpec.Namespace = opsSightName

		// Create and Deploy OpsSight CRD
		opssight := &opssightv1.OpsSight{
			ObjectMeta: metav1.ObjectMeta{
				Name: opsSightName,
			},
			Spec: *defaultOpsSightSpec,
		}
		opssightClient, err := opssightclientset.NewForConfig(restconfig)
		_, err = opssightClient.SynopsysV1().OpsSights(opsSightName).Create(opssight)
		if err != nil {
			fmt.Printf("Error creating the OpsSight : %s\n", err)
			return
		}
	},
}

// createCmd represents the create command for Alert
var createAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Create an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating an Alert\n")
		// Read Commandline Parameters
		alertName := args[0]

		// Get Kubernetes Rest Config
		restconfig := getKubeRestConfig()

		// Create namespace for the Alert
		DeployCRDNamespace(restconfig, alertName)

		// Read Flags Into Default Alert Spec
		defaultAlertSpec = crddefaults.GetAlertDefaultValue()
		flagset := cmd.Flags()
		flagset.VisitAll(checkAlertFlags)

		// Set Namespace in Spec
		defaultAlertSpec.Namespace = alertName

		// Create and Deploy Alert CRD
		alert := &alertv1.Alert{
			ObjectMeta: metav1.ObjectMeta{
				Name: alertName,
			},
			Spec: *defaultAlertSpec,
		}
		alertClient, err := alertclientset.NewForConfig(restconfig)
		_, err = alertClient.SynopsysV1().Alerts(alertName).Create(alert)
		if err != nil {
			fmt.Printf("Error creating the Alert : %s\n", err)
			return
		}
	},
}

func init() {
	createCmd.DisableFlagParsing = true
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Flags
	createBlackduckCmd.Flags().StringVar(&createBlackduckSize, "size", createBlackduckSize, "Blackduck size - small, medium, large")
	createBlackduckCmd.Flags().StringVar(&createBlackduckDbPrototype, "db-prototype", createBlackduckDbPrototype, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckExternalPostgresPostgresHost, "external-postgres-host", createBlackduckExternalPostgresPostgresHost, "TODO")
	createBlackduckCmd.Flags().IntVar(&createBlackduckExternalPostgresPostgresPort, "external-postgres-port", createBlackduckExternalPostgresPostgresPort, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckExternalPostgresPostgresAdmin, "external-postgres-admin", createBlackduckExternalPostgresPostgresAdmin, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckExternalPostgresPostgresUser, "external-postgres-user", createBlackduckExternalPostgresPostgresUser, "TODO")
	createBlackduckCmd.Flags().BoolVar(&createBlackduckExternalPostgresPostgresSsl, "external-postgres-ssl", createBlackduckExternalPostgresPostgresSsl, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", createBlackduckExternalPostgresPostgresAdminPassword, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckExternalPostgresPostgresUserPassword, "external-postgres-user-password", createBlackduckExternalPostgresPostgresUserPassword, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckPvcStorageClass, "pvc-storage-class", createBlackduckPvcStorageClass, "TODO")
	createBlackduckCmd.Flags().BoolVar(&createBlackduckLivenessProbes, "liveness-probes", createBlackduckLivenessProbes, "Enable liveness probes")
	createBlackduckCmd.Flags().StringVar(&createBlackduckScanType, "scan-type", createBlackduckScanType, "TODO")
	createBlackduckCmd.Flags().BoolVar(&createBlackduckPersistentStorage, "persistent-storage", createBlackduckPersistentStorage, "Enable persistent storage")
	createBlackduckCmd.Flags().StringSliceVar(&createBlackduckPVCJSONSlice, "pvc", createBlackduckPVCJSONSlice, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckCertificateName, "db-certificate-name", createBlackduckCertificateName, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckCertificate, "certificate", createBlackduckCertificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckCertificateKey, "certificate-key", createBlackduckCertificateKey, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckProxyCertificate, "proxy-certificate", createBlackduckProxyCertificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckType, "type", createBlackduckType, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckDesiredState, "desired-state", createBlackduckDesiredState, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&createBlackduckEnvirons, "environs", createBlackduckEnvirons, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&createBlackduckImageRegistries, "image-registries", createBlackduckImageRegistries, "List of image registries")
	createBlackduckCmd.Flags().StringSliceVar(&createBlackduckImageUIDMapJSONSlice, "image-uid-map", createBlackduckImageUIDMapJSONSlice, "TODO")
	createBlackduckCmd.Flags().StringVar(&createBlackduckLicenseKey, "license-key", createBlackduckLicenseKey, "TODO")
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Flags
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceptorName, "perceptor-name", createOpssightPerceptorName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceptorImage, "perceptor-image", createOpssightPerceptorImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorPort, "perceptor-port", createOpssightPerceptorPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", createOpssightPerceptorCheckForStalledScansPauseHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", createOpssightPerceptorStalledScanClientTimeoutHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", createOpssightPerceptorModelMetricsPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", createOpssightPerceptorUnknownImagePauseMilliseconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", createOpssightPerceptorClientTimeoutMilliseconds, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodName, "scannerpod-name", createOpssightScannerPodName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodScannerName, "scannerpod-scanner-name", createOpssightScannerPodScannerName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodScannerImage, "scannerpod-scanner-image", createOpssightScannerPodScannerImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightScannerPodScannerPort, "scannerpod-scanner-port", createOpssightScannerPodScannerPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", createOpssightScannerPodScannerClientTimeoutSeconds, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodImageFacadeName, "scannerpod-imagefacade-name", createOpssightScannerPodImageFacadeName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodImageFacadeImage, "scannerpod-imagefacade-image", createOpssightScannerPodImageFacadeImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightScannerPodImageFacadePort, "scannerpod-imagefacade-port", createOpssightScannerPodImageFacadePort, "TODO")
	createOpsSightCmd.Flags().StringSliceVar(&createOpssightScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", createOpssightScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", createOpssightScannerPodImageFacadeImagePullerType, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", createOpssightScannerPodImageFacadeServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightScannerPodReplicaCount, "scannerpod-replica-count", createOpssightScannerPodReplicaCount, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightScannerPodImageDirectory, "scannerpod-image-directory", createOpssightScannerPodImageDirectory, "TODO")
	createOpsSightCmd.Flags().BoolVar(&createOpssightPerceiverEnableImagePerceiver, "enable-image-perceiver", createOpssightPerceiverEnableImagePerceiver, "TODO")
	createOpsSightCmd.Flags().BoolVar(&createOpssightPerceiverEnablePodPerceiver, "enable-pod-perceiver", createOpssightPerceiverEnablePodPerceiver, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverImagePerceiverName, "imageperceiver-name", createOpssightPerceiverImagePerceiverName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverImagePerceiverImage, "imageperceiver-image", createOpssightPerceiverImagePerceiverImage, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverPodPerceiverName, "podperceiver-name", createOpssightPerceiverPodPerceiverName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverPodPerceiverImage, "podperceiver-image", createOpssightPerceiverPodPerceiverImage, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", createOpssightPerceiverPodPerceiverNamespaceFilter, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", createOpssightPerceiverAnnotationIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", createOpssightPerceiverDumpIntervalMinutes, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPerceiverServiceAccount, "perceiver-service-account", createOpssightPerceiverServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPerceiverPort, "perceiver-port", createOpssightPerceiverPort, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPrometheusName, "prometheus-name", createOpssightPrometheusName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightPrometheusName, "prometheus-image", createOpssightPrometheusName, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightPrometheusPort, "prometheus-port", createOpssightPrometheusPort, "TODO")
	createOpsSightCmd.Flags().BoolVar(&createOpssightEnableSkyfire, "enable-skyfire", createOpssightEnableSkyfire, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightSkyfireName, "skyfire-name", createOpssightSkyfireName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightSkyfireImage, "skyfire-image", createOpssightSkyfireImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfirePort, "skyfire-port", createOpssightSkyfirePort, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfirePrometheusPort, "skyfire-prometheus-port", createOpssightSkyfirePrometheusPort, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightSkyfireServiceAccount, "skyfire-service-account", createOpssightSkyfireServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", createOpssightSkyfireHubClientTimeoutSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", createOpssightSkyfireHubDumpPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", createOpssightSkyfireKubeDumpIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightSkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", createOpssightSkyfirePerceptorDumpIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().StringSliceVar(&createOpssightBlackduckHosts, "blackduck-hosts", createOpssightBlackduckHosts, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightBlackduckUser, "blackduck-user", createOpssightBlackduckUser, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckPort, "blackduck-port", createOpssightBlackduckPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckConcurrentScanLimit, "blackduck-concurrent-scan-limit", createOpssightBlackduckConcurrentScanLimit, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckTotalScanLimit, "blackduck-total-scan-limit", createOpssightBlackduckTotalScanLimit, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightBlackduckPasswordEnvVar, "blackduck-password-environment-variable", createOpssightBlackduckPasswordEnvVar, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckInitialCount, "blackduck-initial-count", createOpssightBlackduckInitialCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckMaxCount, "blackduck-max-count", createOpssightBlackduckMaxCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&createOpssightBlackduckDeleteHubThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", createOpssightBlackduckDeleteHubThresholdPercentage, "TODO")
	createOpsSightCmd.Flags().BoolVar(&createOpssightEnableMetrics, "enable-metrics", createOpssightEnableMetrics, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightDefaultCPU, "default-cpu", createOpssightDefaultCPU, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightDefaultMem, "default-mem", createOpssightDefaultMem, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightLogLevel, "log-level", createOpssightLogLevel, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightConfigMapName, "config-map-name", createOpssightConfigMapName, "TODO")
	createOpsSightCmd.Flags().StringVar(&createOpssightSecretName, "secret-name", createOpssightSecretName, "TODO")
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Flags
	createAlertCmd.Flags().StringVar(&createAlertRegistry, "alert-registry", createAlertRegistry, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertImagePath, "image-path", createAlertImagePath, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertAlertImageName, "alert-image-name", createAlertAlertImageName, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertAlertImageVersion, "alert-image-version", createAlertAlertImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertCfsslImageName, "cfssl-image-name", createAlertCfsslImageName, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertCfsslImageVersion, "cfssl-image-version", createAlertCfsslImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertBlackduckHost, "blackduck-host", createAlertBlackduckHost, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertBlackduckUser, "blackduck-user", createAlertBlackduckUser, "TODO")
	createAlertCmd.Flags().IntVar(&createAlertBlackduckPort, "blackduck-port", createAlertBlackduckPort, "TODO")
	createAlertCmd.Flags().IntVar(&createAlertPort, "port", createAlertPort, "TODO")
	createAlertCmd.Flags().BoolVar(&createAlertStandAlone, "stand-alone", createAlertStandAlone, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertAlertMemory, "alert-memory", createAlertAlertMemory, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertCfsslMemory, "cfssl-memory", createAlertCfsslMemory, "TODO")
	createAlertCmd.Flags().StringVar(&createAlertState, "alert-state", createAlertState, "TODO")
	createCmd.AddCommand(createAlertCmd)
}

func checkBlackduckFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "size":
			defaultBlackduckSpec.Size = createBlackduckSize
		case "db-prototype":
			defaultBlackduckSpec.DbPrototype = createBlackduckDbPrototype
		case "external-postgres-host":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresHost = createBlackduckExternalPostgresPostgresHost
		case "external-postgres-port":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresPort = createBlackduckExternalPostgresPostgresPort
		case "external-postgres-admin":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdmin = createBlackduckExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUser = createBlackduckExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresSsl = createBlackduckExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdminPassword = createBlackduckExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUserPassword = createBlackduckExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.PVCStorageClass = createBlackduckPvcStorageClass
		case "liveness-probes":
			defaultBlackduckSpec.LivenessProbes = createBlackduckLivenessProbes
		case "scan-type":
			defaultBlackduckSpec.ScanType = createBlackduckScanType
		case "persistent-storage":
			defaultBlackduckSpec.PersistentStorage = createBlackduckPersistentStorage
		case "pvc":
			for _, pvcJSON := range createBlackduckPVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				defaultBlackduckSpec.PVC = append(defaultBlackduckSpec.PVC, *pvc)
			}
		case "db-certificate-name":
			defaultBlackduckSpec.CertificateName = createBlackduckCertificateName
		case "certificate":
			defaultBlackduckSpec.Certificate = createBlackduckCertificate
		case "certificate-key":
			defaultBlackduckSpec.CertificateKey = createBlackduckCertificateKey
		case "proxy-certificate":
			defaultBlackduckSpec.ProxyCertificate = createBlackduckProxyCertificate
		case "type":
			defaultBlackduckSpec.Type = createBlackduckType
		case "desired-state":
			defaultBlackduckSpec.DesiredState = createBlackduckDesiredState
		case "environs":
			defaultBlackduckSpec.Environs = createBlackduckEnvirons
		case "image-registries":
			defaultBlackduckSpec.ImageRegistries = createBlackduckImageRegistries
		case "image-uid-map":
			type uid struct {
				Key   string `json:"key"`
				Value int64  `json:"value"`
			}
			defaultBlackduckSpec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range createBlackduckImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				defaultBlackduckSpec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "license-key":
			defaultBlackduckSpec.LicenseKey = createBlackduckLicenseKey
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}

func checkOpsSightFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "perceptor-name":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.Name = createOpssightPerceptorName
		case "perceptor-image":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.Image = createOpssightPerceptorImage
		case "perceptor-port":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.Port = createOpssightPerceptorPort
		case "perceptor-check-scan-hours":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.CheckForStalledScansPauseHours = createOpssightPerceptorCheckForStalledScansPauseHours
		case "perceptor-scan-client-timeout-hours":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.StalledScanClientTimeoutHours = createOpssightPerceptorStalledScanClientTimeoutHours
		case "perceptor-metrics-pause-seconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.ModelMetricsPauseSeconds = createOpssightPerceptorModelMetricsPauseSeconds
		case "perceptor-unknown-image-pause-milliseconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.UnknownImagePauseMilliseconds = createOpssightPerceptorUnknownImagePauseMilliseconds
		case "perceptor-client-timeout-milliseconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.ClientTimeoutMilliseconds = createOpssightPerceptorClientTimeoutMilliseconds
		case "scannerpod-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.Name = createOpssightScannerPodName
		case "scannerpod-scanner-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Name = createOpssightScannerPodScannerName
		case "scannerpod-scanner-image":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Image = createOpssightScannerPodScannerImage
		case "scannerpod-scanner-port":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Port = createOpssightScannerPodScannerPort
		case "scannerpod-scanner-client-timeout-seconds":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.ClientTimeoutSeconds = createOpssightScannerPodScannerClientTimeoutSeconds
		case "scannerpod-imagefacade-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Name = createOpssightScannerPodImageFacadeName
		case "scannerpod-imagefacade-image":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Image = createOpssightScannerPodImageFacadeImage
		case "scannerpod-imagefacade-port":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Port = createOpssightScannerPodImageFacadePort
		case "scannerpod-imagefacade-internal-registries":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range createOpssightScannerPodImageFacadeInternalRegistriesJSONSlice {
				registry := &opssightv1.RegistryAuth{}
				json.Unmarshal([]byte(registryJSON), registry)
				defaultOpsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(defaultOpsSightSpec.ScannerPod.ImageFacade.InternalRegistries, *registry)
			}
		case "scannerpod-imagefacade-image-puller-type":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.ImagePullerType = createOpssightScannerPodImageFacadeImagePullerType
		case "scannerpod-imagefacade-service-account":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.ServiceAccount = createOpssightScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.ReplicaCount = createOpssightScannerPodReplicaCount
		case "scannerpod-image-directory":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.ImageDirectory = createOpssightScannerPodImageDirectory
		case "enable-image-perceiver":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.EnableImagePerceiver = createOpssightPerceiverEnableImagePerceiver
		case "enable-pod-perceiver":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.EnablePodPerceiver = createOpssightPerceiverEnablePodPerceiver
		case "imageperceiver-name":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.ImagePerceiver == nil {
				defaultOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			defaultOpsSightSpec.Perceiver.ImagePerceiver.Name = createOpssightPerceiverImagePerceiverName
		case "imageperceiver-image":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.ImagePerceiver == nil {
				defaultOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			defaultOpsSightSpec.Perceiver.ImagePerceiver.Image = createOpssightPerceiverImagePerceiverImage
		case "podperceiver-name":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.Name = createOpssightPerceiverPodPerceiverName
		case "podperceiver-image":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.Image = createOpssightPerceiverPodPerceiverImage
		case "podperceiver-namespace-filter":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.NamespaceFilter = createOpssightPerceiverPodPerceiverNamespaceFilter
		case "perceiver-annotation-interval-seconds":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.AnnotationIntervalSeconds = createOpssightPerceiverAnnotationIntervalSeconds
		case "perceiver-dump-interval-minutes":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.DumpIntervalMinutes = createOpssightPerceiverDumpIntervalMinutes
		case "perceiver-service-account":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.ServiceAccount = createOpssightPerceiverServiceAccount
		case "perceiver-port":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.Port = createOpssightPerceiverPort
		case "prometheus-name":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Name = createOpssightPrometheusName
		case "prometheus-image":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Image = createOpssightPrometheusImage
		case "prometheus-port":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Port = createOpssightPrometheusPort
		case "enable-skyfire":
			defaultOpsSightSpec.EnableSkyfire = createOpssightEnableSkyfire
		case "skyfire-name":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Name = createOpssightSkyfireName
		case "skyfire-image":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Image = createOpssightSkyfireImage
		case "skyfire-port":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Port = createOpssightSkyfirePort
		case "skyfire-prometheus-port":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.PrometheusPort = createOpssightSkyfirePrometheusPort
		case "skyfire-service-account":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.ServiceAccount = createOpssightSkyfireServiceAccount
		case "skyfire-hub-client-timeout-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.HubClientTimeoutSeconds = createOpssightSkyfireHubClientTimeoutSeconds
		case "skyfire-hub-dump-pause-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.HubDumpPauseSeconds = createOpssightSkyfireHubDumpPauseSeconds
		case "skyfire-kube-dump-interval-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.KubeDumpIntervalSeconds = createOpssightSkyfireKubeDumpIntervalSeconds
		case "skyfire-perceptor-dump-interval-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.PerceptorDumpIntervalSeconds = createOpssightSkyfirePerceptorDumpIntervalSeconds
		case "blackduck-hosts":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.Hosts = createOpssightBlackduckHosts
		case "blackduck-user":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.User = createOpssightBlackduckUser
		case "blackduck-port":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.Port = createOpssightBlackduckPort
		case "blackduck-concurrent-scan-limit":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.ConcurrentScanLimit = createOpssightBlackduckConcurrentScanLimit
		case "blackduck-total-scan-limit":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.TotalScanLimit = createOpssightBlackduckTotalScanLimit
		case "blackduck-password-environment-variable":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.PasswordEnvVar = createOpssightBlackduckPasswordEnvVar
		case "blackduck-initial-count":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.InitialCount = createOpssightBlackduckInitialCount
		case "blackduck-max-count":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.MaxCount = createOpssightBlackduckMaxCount
		case "blackduck-delete-blackduck-threshold-percentage":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.DeleteHubThresholdPercentage = createOpssightBlackduckDeleteHubThresholdPercentage
		case "enable-metrics":
			defaultOpsSightSpec.EnableMetrics = createOpssightEnableMetrics
		case "default-cpu":
			defaultOpsSightSpec.DefaultCPU = createOpssightDefaultCPU
		case "default-mem":
			defaultOpsSightSpec.DefaultMem = createOpssightDefaultMem
		case "log-level":
			defaultOpsSightSpec.LogLevel = createOpssightLogLevel
		case "config-map-name":
			defaultOpsSightSpec.ConfigMapName = createOpssightConfigMapName
		case "secret-name":
			defaultOpsSightSpec.SecretName = createOpssightSecretName
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)

}

func checkAlertFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "alert-registry":
			defaultAlertSpec.Registry = createAlertRegistry
		case "image-path":
			defaultAlertSpec.ImagePath = createAlertImagePath
		case "alert-image-name":
			defaultAlertSpec.AlertImageName = createAlertAlertImageName
		case "alert-image-version":
			defaultAlertSpec.AlertImageVersion = createAlertAlertImageVersion
		case "cfssl-image-name":
			defaultAlertSpec.CfsslImageName = createAlertCfsslImageName
		case "cfssl-image-version":
			defaultAlertSpec.CfsslImageVersion = createAlertCfsslImageVersion
		case "blackduck-host":
			defaultAlertSpec.BlackduckHost = createAlertBlackduckHost
		case "blackduck-user":
			defaultAlertSpec.BlackduckUser = createAlertBlackduckUser
		case "blackduck-port":
			defaultAlertSpec.BlackduckPort = &createAlertBlackduckPort
		case "port":
			defaultAlertSpec.Port = &createAlertPort
		case "stand-alone":
			defaultAlertSpec.StandAlone = &createAlertStandAlone
		case "alert-memory":
			defaultAlertSpec.AlertMemory = createAlertAlertMemory
		case "cfssl-memory":
			defaultAlertSpec.CfsslMemory = createAlertCfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
