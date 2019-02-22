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

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create Flags
var createBlackduckSpecType = "persistentStorage"
var createOpsSightSpecType = "disabledBlackduck"
var createAlertSpecType = "spec1"

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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
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
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument - NAME")
		}
		// Check the Spec Type
		switch createBlackduckSpecType {
		case "empty":
			defaultBlackduckSpec = &blackduckv1.BlackduckSpec{}
		case "persistentStorage":
			defaultBlackduckSpec = crddefaults.GetHubDefaultPersistentStorage()
		case "default":
			defaultBlackduckSpec = crddefaults.GetHubDefaultValue()
		default:
			return fmt.Errorf("Blackduck Spec Type %s does not match: empty, persistentStorage, default", createBlackduckSpecType)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating a Blackduck\n")
		// Read Commandline Parameters
		blackduckName := "blackduck"
		if len(args) == 1 {
			blackduckName = args[0]
		}

		// Create namespace for the Blackduck
		DeployCRDNamespace(restconfig, blackduckName)

		// Read Flags Into Default Blackduck Spec
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
		log.Debugf("%+v\n", blackduck)
		_, err := blackduckClient.SynopsysV1().Blackducks(blackduckName).Create(blackduck)
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
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument - NAME")
		}
		// Check the Spec Type
		switch createOpsSightSpecType {
		case "empty":
			defaultOpsSightSpec = &opssightv1.OpsSightSpec{}
		case "disabledBlackduck":
			defaultOpsSightSpec = crddefaults.GetOpsSightDefaultValueWithDisabledHub()
		case "default":
			defaultOpsSightSpec = crddefaults.GetOpsSightDefaultValue()
		default:
			return fmt.Errorf("OpsSight Spec Type %s does not match: empty, disabledBlackduck, default", createOpsSightSpecType)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := "opssight"
		if len(args) == 1 {
			opsSightName = args[0]
		}

		// Create namespace for the OpsSight
		DeployCRDNamespace(restconfig, opsSightName)

		// Read Flags Into Default OpsSight Spec
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
		log.Debugf("%+v\n", opssight)
		_, err := opssightClient.SynopsysV1().OpsSights(opsSightName).Create(opssight)
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
		// Check Number of Arguments
		if len(args) > 1 {
			return fmt.Errorf("This command only accepts up to 1 argument - NAME")
		}
		// Check the Spec Type
		switch createAlertSpecType {
		case "empty":
			defaultAlertSpec = &alertv1.AlertSpec{}
		case "spec1":
			defaultAlertSpec = crddefaults.GetAlertDefaultValue()
		case "spec2":
			defaultAlertSpec = crddefaults.GetAlertDefaultValue2()
		default:
			return fmt.Errorf("Alert Spec Type %s does not match: empty, spec1, spec2", createAlertSpecType)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Creating an Alert\n")
		// Read Commandline Parameters
		alertName := "alert"
		if len(args) == 1 {
			alertName = args[0]
		}

		// Create namespace for the Alert
		DeployCRDNamespace(restconfig, alertName)

		// Read Flags Into Default Alert Spec
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
		log.Debugf("%+v\n", alert)
		_, err := alertClient.SynopsysV1().Alerts(alertName).Create(alert)
		if err != nil {
			fmt.Printf("Error creating the Alert : %s\n", err)
			return
		}
	},
}

func init() {
	createCmd.DisableFlagParsing = true
	rootCmd.AddCommand(createCmd)

	// Add Blackduck Command Flags
	createBlackduckCmd.Flags().StringVar(&createBlackduckSpecType, "spec", createBlackduckSpecType, "TODO")
	// Add Blackduck Spec Flags
	createBlackduckCmd.Flags().StringVar(&blackduckSize, "size", blackduckSize, "Blackduck size - small, medium, large")
	createBlackduckCmd.Flags().StringVar(&blackduckDbPrototype, "db-prototype", blackduckDbPrototype, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresHost, "external-postgres-host", blackduckExternalPostgresPostgresHost, "TODO")
	createBlackduckCmd.Flags().IntVar(&blackduckExternalPostgresPostgresPort, "external-postgres-port", blackduckExternalPostgresPostgresPort, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdmin, "external-postgres-admin", blackduckExternalPostgresPostgresAdmin, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresUser, "external-postgres-user", blackduckExternalPostgresPostgresUser, "TODO")
	createBlackduckCmd.Flags().BoolVar(&blackduckExternalPostgresPostgresSsl, "external-postgres-ssl", blackduckExternalPostgresPostgresSsl, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", blackduckExternalPostgresPostgresAdminPassword, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresUserPassword, "external-postgres-user-password", blackduckExternalPostgresPostgresUserPassword, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckPvcStorageClass, "pvc-storage-class", blackduckPvcStorageClass, "TODO")
	createBlackduckCmd.Flags().BoolVar(&blackduckLivenessProbes, "liveness-probes", blackduckLivenessProbes, "Enable liveness probes")
	createBlackduckCmd.Flags().StringVar(&blackduckScanType, "scan-type", blackduckScanType, "TODO")
	createBlackduckCmd.Flags().BoolVar(&blackduckPersistentStorage, "persistent-storage", blackduckPersistentStorage, "Enable persistent storage")
	createBlackduckCmd.Flags().StringSliceVar(&blackduckPVCJSONSlice, "pvc", blackduckPVCJSONSlice, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckCertificateName, "db-certificate-name", blackduckCertificateName, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckCertificate, "certificate", blackduckCertificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckCertificateKey, "certificate-key", blackduckCertificateKey, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckProxyCertificate, "proxy-certificate", blackduckProxyCertificate, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckType, "type", blackduckType, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckDesiredState, "desired-state", blackduckDesiredState, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&blackduckEnvirons, "environs", blackduckEnvirons, "TODO")
	createBlackduckCmd.Flags().StringSliceVar(&blackduckImageRegistries, "image-registries", blackduckImageRegistries, "List of image registries")
	createBlackduckCmd.Flags().StringSliceVar(&blackduckImageUIDMapJSONSlice, "image-uid-map", blackduckImageUIDMapJSONSlice, "TODO")
	createBlackduckCmd.Flags().StringVar(&blackduckLicenseKey, "license-key", blackduckLicenseKey, "TODO")
	createCmd.AddCommand(createBlackduckCmd)

	// Add OpsSight Command Flags
	createOpsSightCmd.Flags().StringVar(&createOpsSightSpecType, "spec", createOpsSightSpecType, "TODO")
	// Add OpsSight Spec Flags
	createOpsSightCmd.Flags().StringVar(&opssightPerceptorName, "perceptor-name", opssightPerceptorName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceptorImage, "perceptor-image", opssightPerceptorImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorPort, "perceptor-port", opssightPerceptorPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", opssightPerceptorCheckForStalledScansPauseHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", opssightPerceptorStalledScanClientTimeoutHours, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", opssightPerceptorModelMetricsPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", opssightPerceptorUnknownImagePauseMilliseconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", opssightPerceptorClientTimeoutMilliseconds, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodName, "scannerpod-name", opssightScannerPodName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodScannerName, "scannerpod-scanner-name", opssightScannerPodScannerName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodScannerImage, "scannerpod-scanner-image", opssightScannerPodScannerImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightScannerPodScannerPort, "scannerpod-scanner-port", opssightScannerPodScannerPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", opssightScannerPodScannerClientTimeoutSeconds, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeName, "scannerpod-imagefacade-name", opssightScannerPodImageFacadeName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeImage, "scannerpod-imagefacade-image", opssightScannerPodImageFacadeImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightScannerPodImageFacadePort, "scannerpod-imagefacade-port", opssightScannerPodImageFacadePort, "TODO")
	createOpsSightCmd.Flags().StringSliceVar(&opssightScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", opssightScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", opssightScannerPodImageFacadeImagePullerType, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", opssightScannerPodImageFacadeServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightScannerPodReplicaCount, "scannerpod-replica-count", opssightScannerPodReplicaCount, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightScannerPodImageDirectory, "scannerpod-image-directory", opssightScannerPodImageDirectory, "TODO")
	createOpsSightCmd.Flags().BoolVar(&opssightPerceiverEnableImagePerceiver, "enable-image-perceiver", opssightPerceiverEnableImagePerceiver, "TODO")
	createOpsSightCmd.Flags().BoolVar(&opssightPerceiverEnablePodPerceiver, "enable-pod-perceiver", opssightPerceiverEnablePodPerceiver, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverImagePerceiverName, "imageperceiver-name", opssightPerceiverImagePerceiverName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverImagePerceiverImage, "imageperceiver-image", opssightPerceiverImagePerceiverImage, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverName, "podperceiver-name", opssightPerceiverPodPerceiverName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverImage, "podperceiver-image", opssightPerceiverPodPerceiverImage, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", opssightPerceiverPodPerceiverNamespaceFilter, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", opssightPerceiverAnnotationIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", opssightPerceiverDumpIntervalMinutes, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPerceiverServiceAccount, "perceiver-service-account", opssightPerceiverServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPerceiverPort, "perceiver-port", opssightPerceiverPort, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPrometheusName, "prometheus-name", opssightPrometheusName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightPrometheusName, "prometheus-image", opssightPrometheusName, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightPrometheusPort, "prometheus-port", opssightPrometheusPort, "TODO")
	createOpsSightCmd.Flags().BoolVar(&opssightEnableSkyfire, "enable-skyfire", opssightEnableSkyfire, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightSkyfireName, "skyfire-name", opssightSkyfireName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightSkyfireImage, "skyfire-image", opssightSkyfireImage, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfirePort, "skyfire-port", opssightSkyfirePort, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfirePrometheusPort, "skyfire-prometheus-port", opssightSkyfirePrometheusPort, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightSkyfireServiceAccount, "skyfire-service-account", opssightSkyfireServiceAccount, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", opssightSkyfireHubClientTimeoutSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", opssightSkyfireHubDumpPauseSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", opssightSkyfireKubeDumpIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightSkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", opssightSkyfirePerceptorDumpIntervalSeconds, "TODO")
	createOpsSightCmd.Flags().StringSliceVar(&opssightBlackduckHosts, "blackduck-hosts", opssightBlackduckHosts, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightBlackduckUser, "blackduck-user", opssightBlackduckUser, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckPort, "blackduck-port", opssightBlackduckPort, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckConcurrentScanLimit, "blackduck-concurrent-scan-limit", opssightBlackduckConcurrentScanLimit, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckTotalScanLimit, "blackduck-total-scan-limit", opssightBlackduckTotalScanLimit, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightBlackduckPasswordEnvVar, "blackduck-password-environment-variable", opssightBlackduckPasswordEnvVar, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckInitialCount, "blackduck-initial-count", opssightBlackduckInitialCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckMaxCount, "blackduck-max-count", opssightBlackduckMaxCount, "TODO")
	createOpsSightCmd.Flags().IntVar(&opssightBlackduckDeleteHubThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", opssightBlackduckDeleteHubThresholdPercentage, "TODO")
	createOpsSightCmd.Flags().BoolVar(&opssightEnableMetrics, "enable-metrics", opssightEnableMetrics, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightDefaultCPU, "default-cpu", opssightDefaultCPU, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightDefaultMem, "default-mem", opssightDefaultMem, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightLogLevel, "log-level", opssightLogLevel, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightConfigMapName, "config-map-name", opssightConfigMapName, "TODO")
	createOpsSightCmd.Flags().StringVar(&opssightSecretName, "secret-name", opssightSecretName, "TODO")
	createCmd.AddCommand(createOpsSightCmd)

	// Add Alert Command Flags
	createAlertCmd.Flags().StringVar(&createAlertSpecType, "spec", createAlertSpecType, "TODO")
	// Add Alert Spec Flags
	createAlertCmd.Flags().StringVar(&alertRegistry, "alert-registry", alertRegistry, "TODO")
	createAlertCmd.Flags().StringVar(&alertImagePath, "image-path", alertImagePath, "TODO")
	createAlertCmd.Flags().StringVar(&alertAlertImageName, "alert-image-name", alertAlertImageName, "TODO")
	createAlertCmd.Flags().StringVar(&alertAlertImageVersion, "alert-image-version", alertAlertImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&alertCfsslImageName, "cfssl-image-name", alertCfsslImageName, "TODO")
	createAlertCmd.Flags().StringVar(&alertCfsslImageVersion, "cfssl-image-version", alertCfsslImageVersion, "TODO")
	createAlertCmd.Flags().StringVar(&alertBlackduckHost, "blackduck-host", alertBlackduckHost, "TODO")
	createAlertCmd.Flags().StringVar(&alertBlackduckUser, "blackduck-user", alertBlackduckUser, "TODO")
	createAlertCmd.Flags().IntVar(&alertBlackduckPort, "blackduck-port", alertBlackduckPort, "TODO")
	createAlertCmd.Flags().IntVar(&alertPort, "port", alertPort, "TODO")
	createAlertCmd.Flags().BoolVar(&alertStandAlone, "stand-alone", alertStandAlone, "TODO")
	createAlertCmd.Flags().StringVar(&alertAlertMemory, "alert-memory", alertAlertMemory, "TODO")
	createAlertCmd.Flags().StringVar(&alertCfsslMemory, "cfssl-memory", alertCfsslMemory, "TODO")
	createAlertCmd.Flags().StringVar(&alertState, "alert-state", alertState, "TODO")
	createCmd.AddCommand(createAlertCmd)
}

func checkBlackduckFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "size":
			defaultBlackduckSpec.Size = blackduckSize
		case "db-prototype":
			defaultBlackduckSpec.DbPrototype = blackduckDbPrototype
		case "external-postgres-host":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresHost = blackduckExternalPostgresPostgresHost
		case "external-postgres-port":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresPort = blackduckExternalPostgresPostgresPort
		case "external-postgres-admin":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdmin = blackduckExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUser = blackduckExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresSsl = blackduckExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresAdminPassword = blackduckExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.ExternalPostgres.PostgresUserPassword = blackduckExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			if defaultBlackduckSpec.ExternalPostgres == nil {
				defaultBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			defaultBlackduckSpec.PVCStorageClass = blackduckPvcStorageClass
		case "liveness-probes":
			defaultBlackduckSpec.LivenessProbes = blackduckLivenessProbes
		case "scan-type":
			defaultBlackduckSpec.ScanType = blackduckScanType
		case "persistent-storage":
			defaultBlackduckSpec.PersistentStorage = blackduckPersistentStorage
		case "pvc":
			for _, pvcJSON := range blackduckPVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				defaultBlackduckSpec.PVC = append(defaultBlackduckSpec.PVC, *pvc)
			}
		case "db-certificate-name":
			defaultBlackduckSpec.CertificateName = blackduckCertificateName
		case "certificate":
			defaultBlackduckSpec.Certificate = blackduckCertificate
		case "certificate-key":
			defaultBlackduckSpec.CertificateKey = blackduckCertificateKey
		case "proxy-certificate":
			defaultBlackduckSpec.ProxyCertificate = blackduckProxyCertificate
		case "type":
			defaultBlackduckSpec.Type = blackduckType
		case "desired-state":
			defaultBlackduckSpec.DesiredState = blackduckDesiredState
		case "environs":
			defaultBlackduckSpec.Environs = blackduckEnvirons
		case "image-registries":
			defaultBlackduckSpec.ImageRegistries = blackduckImageRegistries
		case "image-uid-map":
			type uid struct {
				Key   string `json:"key"`
				Value int64  `json:"value"`
			}
			defaultBlackduckSpec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range blackduckImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				defaultBlackduckSpec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "license-key":
			defaultBlackduckSpec.LicenseKey = blackduckLicenseKey
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
			defaultOpsSightSpec.Perceptor.Name = opssightPerceptorName
		case "perceptor-image":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.Image = opssightPerceptorImage
		case "perceptor-port":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.Port = opssightPerceptorPort
		case "perceptor-check-scan-hours":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.CheckForStalledScansPauseHours = opssightPerceptorCheckForStalledScansPauseHours
		case "perceptor-scan-client-timeout-hours":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.StalledScanClientTimeoutHours = opssightPerceptorStalledScanClientTimeoutHours
		case "perceptor-metrics-pause-seconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.ModelMetricsPauseSeconds = opssightPerceptorModelMetricsPauseSeconds
		case "perceptor-unknown-image-pause-milliseconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.UnknownImagePauseMilliseconds = opssightPerceptorUnknownImagePauseMilliseconds
		case "perceptor-client-timeout-milliseconds":
			if defaultOpsSightSpec.Perceptor == nil {
				defaultOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			defaultOpsSightSpec.Perceptor.ClientTimeoutMilliseconds = opssightPerceptorClientTimeoutMilliseconds
		case "scannerpod-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.Name = opssightScannerPodName
		case "scannerpod-scanner-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Name = opssightScannerPodScannerName
		case "scannerpod-scanner-image":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Image = opssightScannerPodScannerImage
		case "scannerpod-scanner-port":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.Port = opssightScannerPodScannerPort
		case "scannerpod-scanner-client-timeout-seconds":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.Scanner == nil {
				defaultOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			defaultOpsSightSpec.ScannerPod.Scanner.ClientTimeoutSeconds = opssightScannerPodScannerClientTimeoutSeconds
		case "scannerpod-imagefacade-name":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Name = opssightScannerPodImageFacadeName
		case "scannerpod-imagefacade-image":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Image = opssightScannerPodImageFacadeImage
		case "scannerpod-imagefacade-port":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.Port = opssightScannerPodImageFacadePort
		case "scannerpod-imagefacade-internal-registries":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range opssightScannerPodImageFacadeInternalRegistriesJSONSlice {
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
			defaultOpsSightSpec.ScannerPod.ImageFacade.ImagePullerType = opssightScannerPodImageFacadeImagePullerType
		case "scannerpod-imagefacade-service-account":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if defaultOpsSightSpec.ScannerPod.ImageFacade == nil {
				defaultOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			defaultOpsSightSpec.ScannerPod.ImageFacade.ServiceAccount = opssightScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.ReplicaCount = opssightScannerPodReplicaCount
		case "scannerpod-image-directory":
			if defaultOpsSightSpec.ScannerPod == nil {
				defaultOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			defaultOpsSightSpec.ScannerPod.ImageDirectory = opssightScannerPodImageDirectory
		case "enable-image-perceiver":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.EnableImagePerceiver = opssightPerceiverEnableImagePerceiver
		case "enable-pod-perceiver":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.EnablePodPerceiver = opssightPerceiverEnablePodPerceiver
		case "imageperceiver-name":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.ImagePerceiver == nil {
				defaultOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			defaultOpsSightSpec.Perceiver.ImagePerceiver.Name = opssightPerceiverImagePerceiverName
		case "imageperceiver-image":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.ImagePerceiver == nil {
				defaultOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			defaultOpsSightSpec.Perceiver.ImagePerceiver.Image = opssightPerceiverImagePerceiverImage
		case "podperceiver-name":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.Name = opssightPerceiverPodPerceiverName
		case "podperceiver-image":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.Image = opssightPerceiverPodPerceiverImage
		case "podperceiver-namespace-filter":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if defaultOpsSightSpec.Perceiver.PodPerceiver == nil {
				defaultOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			defaultOpsSightSpec.Perceiver.PodPerceiver.NamespaceFilter = opssightPerceiverPodPerceiverNamespaceFilter
		case "perceiver-annotation-interval-seconds":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.AnnotationIntervalSeconds = opssightPerceiverAnnotationIntervalSeconds
		case "perceiver-dump-interval-minutes":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.DumpIntervalMinutes = opssightPerceiverDumpIntervalMinutes
		case "perceiver-service-account":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.ServiceAccount = opssightPerceiverServiceAccount
		case "perceiver-port":
			if defaultOpsSightSpec.Perceiver == nil {
				defaultOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			defaultOpsSightSpec.Perceiver.Port = opssightPerceiverPort
		case "prometheus-name":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Name = opssightPrometheusName
		case "prometheus-image":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Image = opssightPrometheusImage
		case "prometheus-port":
			if defaultOpsSightSpec.Prometheus == nil {
				defaultOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			defaultOpsSightSpec.Prometheus.Port = opssightPrometheusPort
		case "enable-skyfire":
			defaultOpsSightSpec.EnableSkyfire = opssightEnableSkyfire
		case "skyfire-name":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Name = opssightSkyfireName
		case "skyfire-image":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Image = opssightSkyfireImage
		case "skyfire-port":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.Port = opssightSkyfirePort
		case "skyfire-prometheus-port":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.PrometheusPort = opssightSkyfirePrometheusPort
		case "skyfire-service-account":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.ServiceAccount = opssightSkyfireServiceAccount
		case "skyfire-hub-client-timeout-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.HubClientTimeoutSeconds = opssightSkyfireHubClientTimeoutSeconds
		case "skyfire-hub-dump-pause-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.HubDumpPauseSeconds = opssightSkyfireHubDumpPauseSeconds
		case "skyfire-kube-dump-interval-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.KubeDumpIntervalSeconds = opssightSkyfireKubeDumpIntervalSeconds
		case "skyfire-perceptor-dump-interval-seconds":
			if defaultOpsSightSpec.Skyfire == nil {
				defaultOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			defaultOpsSightSpec.Skyfire.PerceptorDumpIntervalSeconds = opssightSkyfirePerceptorDumpIntervalSeconds
		case "blackduck-hosts":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.Hosts = opssightBlackduckHosts
		case "blackduck-user":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.User = opssightBlackduckUser
		case "blackduck-port":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.Port = opssightBlackduckPort
		case "blackduck-concurrent-scan-limit":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.ConcurrentScanLimit = opssightBlackduckConcurrentScanLimit
		case "blackduck-total-scan-limit":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.TotalScanLimit = opssightBlackduckTotalScanLimit
		case "blackduck-password-environment-variable":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.PasswordEnvVar = opssightBlackduckPasswordEnvVar
		case "blackduck-initial-count":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.InitialCount = opssightBlackduckInitialCount
		case "blackduck-max-count":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.MaxCount = opssightBlackduckMaxCount
		case "blackduck-delete-blackduck-threshold-percentage":
			if defaultOpsSightSpec.Blackduck == nil {
				defaultOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			defaultOpsSightSpec.Blackduck.DeleteHubThresholdPercentage = opssightBlackduckDeleteHubThresholdPercentage
		case "enable-metrics":
			defaultOpsSightSpec.EnableMetrics = opssightEnableMetrics
		case "default-cpu":
			defaultOpsSightSpec.DefaultCPU = opssightDefaultCPU
		case "default-mem":
			defaultOpsSightSpec.DefaultMem = opssightDefaultMem
		case "log-level":
			defaultOpsSightSpec.LogLevel = opssightLogLevel
		case "config-map-name":
			defaultOpsSightSpec.ConfigMapName = opssightConfigMapName
		case "secret-name":
			defaultOpsSightSpec.SecretName = opssightSecretName
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
			defaultAlertSpec.Registry = alertRegistry
		case "image-path":
			defaultAlertSpec.ImagePath = alertImagePath
		case "alert-image-name":
			defaultAlertSpec.AlertImageName = alertAlertImageName
		case "alert-image-version":
			defaultAlertSpec.AlertImageVersion = alertAlertImageVersion
		case "cfssl-image-name":
			defaultAlertSpec.CfsslImageName = alertCfsslImageName
		case "cfssl-image-version":
			defaultAlertSpec.CfsslImageVersion = alertCfsslImageVersion
		case "blackduck-host":
			defaultAlertSpec.BlackduckHost = alertBlackduckHost
		case "blackduck-user":
			defaultAlertSpec.BlackduckUser = alertBlackduckUser
		case "blackduck-port":
			defaultAlertSpec.BlackduckPort = &alertBlackduckPort
		case "port":
			defaultAlertSpec.Port = &alertPort
		case "stand-alone":
			defaultAlertSpec.StandAlone = &alertStandAlone
		case "alert-memory":
			defaultAlertSpec.AlertMemory = alertAlertMemory
		case "cfssl-memory":
			defaultAlertSpec.CfsslMemory = alertCfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
