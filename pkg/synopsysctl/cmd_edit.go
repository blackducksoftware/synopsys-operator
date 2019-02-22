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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Gloabal Specs
var editBlackduckSpec = &blackduckv1.BlackduckSpec{}
var editOpsSightSpec = &opssightv1.OpsSightSpec{}
var editAlertSpec = &alertv1.AlertSpec{}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Allows you to directly edit the API resource",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "--help" {
			return fmt.Errorf("Help Called")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Non-Synopsys Resource\n")
		kubeCmdArgs := append([]string{"edit"}, args...)
		out, err := RunKubeCmd(kubeCmdArgs...)
		if err != nil {
			log.Errorf("Error Editing the Resource with KubeCmd: %s", out)
		} else {
			fmt.Printf("%+v", out)
		}
	},
}

var editBlackduckCmd = &cobra.Command{
	Use:   "blackduck",
	Short: "Edit an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing Blackduck\n")
		// Read Commandline Parameters
		blackduckName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			bd, err := getBlackduckSpec(blackduckName)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			editBlackduckSpec = &bd.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editBlackduckFlags)
			// Update Blackduck with Updates
			err = updateBlackduckSpec(bd)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "blackduck", blackduckName, "-n", blackduckName)
			if err != nil {
				fmt.Printf("Error Editing the Blackduck: %s\n", err)
			}
		}
	},
}

var editBlackduckAddPVCCmd = &cobra.Command{
	Use:   "addPVC",
	Short: "Add a PVC to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding PVC to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddEnvironCmd = &cobra.Command{
	Use:   "addEnviron",
	Short: "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Environ to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry",
	Short: "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image Registry to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editBlackduckAddUIDCmd = &cobra.Command{
	Use:   "addUID",
	Short: "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding an Image UID to Blackduck\n")
		// Read Commandline Parameters
		//blackduckNamespace := args[0]
	},
}

var editOpsSightCmd = &cobra.Command{
	Use:   "opssight",
	Short: "Edit an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an OpsSight\n")
		// Read Commandline Parameters
		opsSightName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			ops, err := getOpsSightSpec(opsSightName)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			editOpsSightSpec = &ops.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editOpsSightFlags)
			// Update OpsSight with Updates
			err = updateOpsSightSpec(ops)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "opssight", opsSightName, "-n", opsSightName)
			if err != nil {
				fmt.Printf("Error Editing the OpsSight: %s\n", err)
			}
		}
	},
}

var editOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "addRegistry",
	Short: "Add an Internal Registry to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Internal Registryto OpsSight\n")
		// Read Commandline Parameters
		//opSightNamespace := args[0]
		return
	},
}

var editOpsSightAddHostCmd = &cobra.Command{
	Use:   "addHost",
	Short: "Add a Blackduck Host to OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("This command only accepts 2 arguments - NAME  BLACKDUCK_HOST")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Adding Blackduck Host to OpsSight\n")
		// Read Commandline Parameters
		//opSightNamespace := args[0]
		return
	},
}

var editAlertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Edit an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("This command only accepts 1 argument - NAME")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Editing an Alert\n")
		// Read Commandline Parameters
		alertName := args[0]

		// Update spec with flags or pipe to KubeCmd
		flagset := cmd.Flags()
		if flagset.NFlag() != 0 {
			alt, err := getAlertSpec(alertName)
			if err != nil {
				fmt.Printf("Get Spec: %s\n", err)
				return
			}
			editAlertSpec = &alt.Spec
			// Update Spec with Changes from Flags
			flagset.VisitAll(editAlertFlags)
			// Update Alert with Updates
			err = updateAlertSpec(alt)
			if err != nil {
				fmt.Printf("Update Spec: %s\n", err)
				return
			}
		} else {
			err := RunKubeEditorCmd("edit", "alert", alertName, "-n", alertName)
			if err != nil {
				fmt.Printf("Error Editing the Alert: %s\n", err)
			}
		}
	},
}

func init() {
	editCmd.DisableFlagParsing = true
	rootCmd.AddCommand(editCmd)

	// Add Blackduck Spec Flags
	editBlackduckCmd.Flags().StringVar(&blackduckSize, "size", blackduckSize, "Blackduck size - small, medium, large")
	editBlackduckCmd.Flags().StringVar(&blackduckDbPrototype, "db-prototype", blackduckDbPrototype, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresHost, "external-postgres-host", blackduckExternalPostgresPostgresHost, "TODO")
	editBlackduckCmd.Flags().IntVar(&blackduckExternalPostgresPostgresPort, "external-postgres-port", blackduckExternalPostgresPostgresPort, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdmin, "external-postgres-admin", blackduckExternalPostgresPostgresAdmin, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresUser, "external-postgres-user", blackduckExternalPostgresPostgresUser, "TODO")
	editBlackduckCmd.Flags().BoolVar(&blackduckExternalPostgresPostgresSsl, "external-postgres-ssl", blackduckExternalPostgresPostgresSsl, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", blackduckExternalPostgresPostgresAdminPassword, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckExternalPostgresPostgresUserPassword, "external-postgres-user-password", blackduckExternalPostgresPostgresUserPassword, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckPvcStorageClass, "pvc-storage-class", blackduckPvcStorageClass, "TODO")
	editBlackduckCmd.Flags().BoolVar(&blackduckLivenessProbes, "liveness-probes", blackduckLivenessProbes, "Enable liveness probes")
	editBlackduckCmd.Flags().StringVar(&blackduckScanType, "scan-type", blackduckScanType, "TODO")
	editBlackduckCmd.Flags().BoolVar(&blackduckPersistentStorage, "persistent-storage", blackduckPersistentStorage, "Enable persistent storage")
	editBlackduckCmd.Flags().StringSliceVar(&blackduckPVCJSONSlice, "pvc", blackduckPVCJSONSlice, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckCertificateName, "db-certificate-name", blackduckCertificateName, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckCertificate, "certificate", blackduckCertificate, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckCertificateKey, "certificate-key", blackduckCertificateKey, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckProxyCertificate, "proxy-certificate", blackduckProxyCertificate, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckType, "type", blackduckType, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckDesiredState, "desired-state", blackduckDesiredState, "TODO")
	editBlackduckCmd.Flags().StringSliceVar(&blackduckEnvirons, "environs", blackduckEnvirons, "TODO")
	editBlackduckCmd.Flags().StringSliceVar(&blackduckImageRegistries, "image-registries", blackduckImageRegistries, "List of image registries")
	editBlackduckCmd.Flags().StringSliceVar(&blackduckImageUIDMapJSONSlice, "image-uid-map", blackduckImageUIDMapJSONSlice, "TODO")
	editBlackduckCmd.Flags().StringVar(&blackduckLicenseKey, "license-key", blackduckLicenseKey, "TODO")
	editCmd.AddCommand(editBlackduckCmd)

	// Add Blackduck Commands
	editBlackduckCmd.AddCommand(editBlackduckAddPVCCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddEnvironCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddRegistryCmd)
	editBlackduckCmd.AddCommand(editBlackduckAddUIDCmd)

	// Add OpsSight Spec Flags
	editOpsSightCmd.Flags().StringVar(&opssightPerceptorName, "perceptor-name", opssightPerceptorName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceptorImage, "perceptor-image", opssightPerceptorImage, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorPort, "perceptor-port", opssightPerceptorPort, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", opssightPerceptorCheckForStalledScansPauseHours, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", opssightPerceptorStalledScanClientTimeoutHours, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", opssightPerceptorModelMetricsPauseSeconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", opssightPerceptorUnknownImagePauseMilliseconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", opssightPerceptorClientTimeoutMilliseconds, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodName, "scannerpod-name", opssightScannerPodName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodScannerName, "scannerpod-scanner-name", opssightScannerPodScannerName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodScannerImage, "scannerpod-scanner-image", opssightScannerPodScannerImage, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightScannerPodScannerPort, "scannerpod-scanner-port", opssightScannerPodScannerPort, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", opssightScannerPodScannerClientTimeoutSeconds, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeName, "scannerpod-imagefacade-name", opssightScannerPodImageFacadeName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeImage, "scannerpod-imagefacade-image", opssightScannerPodImageFacadeImage, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightScannerPodImageFacadePort, "scannerpod-imagefacade-port", opssightScannerPodImageFacadePort, "TODO")
	editOpsSightCmd.Flags().StringSliceVar(&opssightScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", opssightScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", opssightScannerPodImageFacadeImagePullerType, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", opssightScannerPodImageFacadeServiceAccount, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightScannerPodReplicaCount, "scannerpod-replica-count", opssightScannerPodReplicaCount, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightScannerPodImageDirectory, "scannerpod-image-directory", opssightScannerPodImageDirectory, "TODO")
	editOpsSightCmd.Flags().BoolVar(&opssightPerceiverEnableImagePerceiver, "enable-image-perceiver", opssightPerceiverEnableImagePerceiver, "TODO")
	editOpsSightCmd.Flags().BoolVar(&opssightPerceiverEnablePodPerceiver, "enable-pod-perceiver", opssightPerceiverEnablePodPerceiver, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverImagePerceiverName, "imageperceiver-name", opssightPerceiverImagePerceiverName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverImagePerceiverImage, "imageperceiver-image", opssightPerceiverImagePerceiverImage, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverName, "podperceiver-name", opssightPerceiverPodPerceiverName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverImage, "podperceiver-image", opssightPerceiverPodPerceiverImage, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", opssightPerceiverPodPerceiverNamespaceFilter, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", opssightPerceiverAnnotationIntervalSeconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", opssightPerceiverDumpIntervalMinutes, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPerceiverServiceAccount, "perceiver-service-account", opssightPerceiverServiceAccount, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPerceiverPort, "perceiver-port", opssightPerceiverPort, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPrometheusName, "prometheus-name", opssightPrometheusName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightPrometheusName, "prometheus-image", opssightPrometheusName, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightPrometheusPort, "prometheus-port", opssightPrometheusPort, "TODO")
	editOpsSightCmd.Flags().BoolVar(&opssightEnableSkyfire, "enable-skyfire", opssightEnableSkyfire, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightSkyfireName, "skyfire-name", opssightSkyfireName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightSkyfireImage, "skyfire-image", opssightSkyfireImage, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfirePort, "skyfire-port", opssightSkyfirePort, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfirePrometheusPort, "skyfire-prometheus-port", opssightSkyfirePrometheusPort, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightSkyfireServiceAccount, "skyfire-service-account", opssightSkyfireServiceAccount, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", opssightSkyfireHubClientTimeoutSeconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", opssightSkyfireHubDumpPauseSeconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", opssightSkyfireKubeDumpIntervalSeconds, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightSkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", opssightSkyfirePerceptorDumpIntervalSeconds, "TODO")
	editOpsSightCmd.Flags().StringSliceVar(&opssightBlackduckHosts, "blackduck-hosts", opssightBlackduckHosts, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightBlackduckUser, "blackduck-user", opssightBlackduckUser, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckPort, "blackduck-port", opssightBlackduckPort, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckConcurrentScanLimit, "blackduck-concurrent-scan-limit", opssightBlackduckConcurrentScanLimit, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckTotalScanLimit, "blackduck-total-scan-limit", opssightBlackduckTotalScanLimit, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightBlackduckPasswordEnvVar, "blackduck-password-environment-variable", opssightBlackduckPasswordEnvVar, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckInitialCount, "blackduck-initial-count", opssightBlackduckInitialCount, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckMaxCount, "blackduck-max-count", opssightBlackduckMaxCount, "TODO")
	editOpsSightCmd.Flags().IntVar(&opssightBlackduckDeleteHubThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", opssightBlackduckDeleteHubThresholdPercentage, "TODO")
	editOpsSightCmd.Flags().BoolVar(&opssightEnableMetrics, "enable-metrics", opssightEnableMetrics, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightDefaultCPU, "default-cpu", opssightDefaultCPU, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightDefaultMem, "default-mem", opssightDefaultMem, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightLogLevel, "log-level", opssightLogLevel, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightConfigMapName, "config-map-name", opssightConfigMapName, "TODO")
	editOpsSightCmd.Flags().StringVar(&opssightSecretName, "secret-name", opssightSecretName, "TODO")
	editCmd.AddCommand(editOpsSightCmd)

	// Add OpsSight Commands
	editOpsSightCmd.AddCommand(editOpsSightAddRegistryCmd)
	editOpsSightCmd.AddCommand(editOpsSightAddHostCmd)

	// Add Alert Spec Flags
	editAlertCmd.Flags().StringVar(&alertRegistry, "alert-registry", alertRegistry, "TODO")
	editAlertCmd.Flags().StringVar(&alertImagePath, "image-path", alertImagePath, "TODO")
	editAlertCmd.Flags().StringVar(&alertAlertImageName, "alert-image-name", alertAlertImageName, "TODO")
	editAlertCmd.Flags().StringVar(&alertAlertImageVersion, "alert-image-version", alertAlertImageVersion, "TODO")
	editAlertCmd.Flags().StringVar(&alertCfsslImageName, "cfssl-image-name", alertCfsslImageName, "TODO")
	editAlertCmd.Flags().StringVar(&alertCfsslImageVersion, "cfssl-image-version", alertCfsslImageVersion, "TODO")
	editAlertCmd.Flags().StringVar(&alertBlackduckHost, "blackduck-host", alertBlackduckHost, "TODO")
	editAlertCmd.Flags().StringVar(&alertBlackduckUser, "blackduck-user", alertBlackduckUser, "TODO")
	editAlertCmd.Flags().IntVar(&alertBlackduckPort, "blackduck-port", alertBlackduckPort, "TODO")
	editAlertCmd.Flags().IntVar(&alertPort, "port", alertPort, "TODO")
	editAlertCmd.Flags().BoolVar(&alertStandAlone, "stand-alone", alertStandAlone, "TODO")
	editAlertCmd.Flags().StringVar(&alertAlertMemory, "alert-memory", alertAlertMemory, "TODO")
	editAlertCmd.Flags().StringVar(&alertCfsslMemory, "cfssl-memory", alertCfsslMemory, "TODO")
	editAlertCmd.Flags().StringVar(&alertState, "alert-state", alertState, "TODO")
	editCmd.AddCommand(editAlertCmd)
}

func editBlackduckFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "size":
			editBlackduckSpec.Size = blackduckSize
		case "db-prototype":
			editBlackduckSpec.DbPrototype = blackduckDbPrototype
		case "external-postgres-host":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresHost = blackduckExternalPostgresPostgresHost
		case "external-postgres-port":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresPort = blackduckExternalPostgresPostgresPort
		case "external-postgres-admin":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresAdmin = blackduckExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresUser = blackduckExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresSsl = blackduckExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresAdminPassword = blackduckExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.ExternalPostgres.PostgresUserPassword = blackduckExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			if editBlackduckSpec.ExternalPostgres == nil {
				editBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			editBlackduckSpec.PVCStorageClass = blackduckPvcStorageClass
		case "liveness-probes":
			editBlackduckSpec.LivenessProbes = blackduckLivenessProbes
		case "scan-type":
			editBlackduckSpec.ScanType = blackduckScanType
		case "persistent-storage":
			editBlackduckSpec.PersistentStorage = blackduckPersistentStorage
		case "pvc":
			for _, pvcJSON := range blackduckPVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				editBlackduckSpec.PVC = append(editBlackduckSpec.PVC, *pvc)
			}
		case "db-certificate-name":
			editBlackduckSpec.CertificateName = blackduckCertificateName
		case "certificate":
			editBlackduckSpec.Certificate = blackduckCertificate
		case "certificate-key":
			editBlackduckSpec.CertificateKey = blackduckCertificateKey
		case "proxy-certificate":
			editBlackduckSpec.ProxyCertificate = blackduckProxyCertificate
		case "type":
			editBlackduckSpec.Type = blackduckType
		case "desired-state":
			editBlackduckSpec.DesiredState = blackduckDesiredState
		case "environs":
			editBlackduckSpec.Environs = blackduckEnvirons
		case "image-registries":
			editBlackduckSpec.ImageRegistries = blackduckImageRegistries
		case "image-uid-map":
			type uid struct {
				Key   string `json:"key"`
				Value int64  `json:"value"`
			}
			editBlackduckSpec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range blackduckImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				editBlackduckSpec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "license-key":
			editBlackduckSpec.LicenseKey = blackduckLicenseKey
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}

func editOpsSightFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "perceptor-name":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.Name = opssightPerceptorName
		case "perceptor-image":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.Image = opssightPerceptorImage
		case "perceptor-port":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.Port = opssightPerceptorPort
		case "perceptor-check-scan-hours":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.CheckForStalledScansPauseHours = opssightPerceptorCheckForStalledScansPauseHours
		case "perceptor-scan-client-timeout-hours":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.StalledScanClientTimeoutHours = opssightPerceptorStalledScanClientTimeoutHours
		case "perceptor-metrics-pause-seconds":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.ModelMetricsPauseSeconds = opssightPerceptorModelMetricsPauseSeconds
		case "perceptor-unknown-image-pause-milliseconds":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.UnknownImagePauseMilliseconds = opssightPerceptorUnknownImagePauseMilliseconds
		case "perceptor-client-timeout-milliseconds":
			if editOpsSightSpec.Perceptor == nil {
				editOpsSightSpec.Perceptor = &opssightv1.Perceptor{}
			}
			editOpsSightSpec.Perceptor.ClientTimeoutMilliseconds = opssightPerceptorClientTimeoutMilliseconds
		case "scannerpod-name":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			editOpsSightSpec.ScannerPod.Name = opssightScannerPodName
		case "scannerpod-scanner-name":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.Scanner == nil {
				editOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			editOpsSightSpec.ScannerPod.Scanner.Name = opssightScannerPodScannerName
		case "scannerpod-scanner-image":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.Scanner == nil {
				editOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			editOpsSightSpec.ScannerPod.Scanner.Image = opssightScannerPodScannerImage
		case "scannerpod-scanner-port":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.Scanner == nil {
				editOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			editOpsSightSpec.ScannerPod.Scanner.Port = opssightScannerPodScannerPort
		case "scannerpod-scanner-client-timeout-seconds":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.Scanner == nil {
				editOpsSightSpec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			editOpsSightSpec.ScannerPod.Scanner.ClientTimeoutSeconds = opssightScannerPodScannerClientTimeoutSeconds
		case "scannerpod-imagefacade-name":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			editOpsSightSpec.ScannerPod.ImageFacade.Name = opssightScannerPodImageFacadeName
		case "scannerpod-imagefacade-image":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			editOpsSightSpec.ScannerPod.ImageFacade.Image = opssightScannerPodImageFacadeImage
		case "scannerpod-imagefacade-port":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			editOpsSightSpec.ScannerPod.ImageFacade.Port = opssightScannerPodImageFacadePort
		case "scannerpod-imagefacade-internal-registries":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range opssightScannerPodImageFacadeInternalRegistriesJSONSlice {
				registry := &opssightv1.RegistryAuth{}
				json.Unmarshal([]byte(registryJSON), registry)
				editOpsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(editOpsSightSpec.ScannerPod.ImageFacade.InternalRegistries, *registry)
			}
		case "scannerpod-imagefacade-image-puller-type":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			editOpsSightSpec.ScannerPod.ImageFacade.ImagePullerType = opssightScannerPodImageFacadeImagePullerType
		case "scannerpod-imagefacade-service-account":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if editOpsSightSpec.ScannerPod.ImageFacade == nil {
				editOpsSightSpec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			editOpsSightSpec.ScannerPod.ImageFacade.ServiceAccount = opssightScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			editOpsSightSpec.ScannerPod.ReplicaCount = opssightScannerPodReplicaCount
		case "scannerpod-image-directory":
			if editOpsSightSpec.ScannerPod == nil {
				editOpsSightSpec.ScannerPod = &opssightv1.ScannerPod{}
			}
			editOpsSightSpec.ScannerPod.ImageDirectory = opssightScannerPodImageDirectory
		case "enable-image-perceiver":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.EnableImagePerceiver = opssightPerceiverEnableImagePerceiver
		case "enable-pod-perceiver":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.EnablePodPerceiver = opssightPerceiverEnablePodPerceiver
		case "imageperceiver-name":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if editOpsSightSpec.Perceiver.ImagePerceiver == nil {
				editOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			editOpsSightSpec.Perceiver.ImagePerceiver.Name = opssightPerceiverImagePerceiverName
		case "imageperceiver-image":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if editOpsSightSpec.Perceiver.ImagePerceiver == nil {
				editOpsSightSpec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			editOpsSightSpec.Perceiver.ImagePerceiver.Image = opssightPerceiverImagePerceiverImage
		case "podperceiver-name":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if editOpsSightSpec.Perceiver.PodPerceiver == nil {
				editOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			editOpsSightSpec.Perceiver.PodPerceiver.Name = opssightPerceiverPodPerceiverName
		case "podperceiver-image":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if editOpsSightSpec.Perceiver.PodPerceiver == nil {
				editOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			editOpsSightSpec.Perceiver.PodPerceiver.Image = opssightPerceiverPodPerceiverImage
		case "podperceiver-namespace-filter":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			if editOpsSightSpec.Perceiver.PodPerceiver == nil {
				editOpsSightSpec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			editOpsSightSpec.Perceiver.PodPerceiver.NamespaceFilter = opssightPerceiverPodPerceiverNamespaceFilter
		case "perceiver-annotation-interval-seconds":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.AnnotationIntervalSeconds = opssightPerceiverAnnotationIntervalSeconds
		case "perceiver-dump-interval-minutes":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.DumpIntervalMinutes = opssightPerceiverDumpIntervalMinutes
		case "perceiver-service-account":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.ServiceAccount = opssightPerceiverServiceAccount
		case "perceiver-port":
			if editOpsSightSpec.Perceiver == nil {
				editOpsSightSpec.Perceiver = &opssightv1.Perceiver{}
			}
			editOpsSightSpec.Perceiver.Port = opssightPerceiverPort
		case "prometheus-name":
			if editOpsSightSpec.Prometheus == nil {
				editOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			editOpsSightSpec.Prometheus.Name = opssightPrometheusName
		case "prometheus-image":
			if editOpsSightSpec.Prometheus == nil {
				editOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			editOpsSightSpec.Prometheus.Image = opssightPrometheusImage
		case "prometheus-port":
			if editOpsSightSpec.Prometheus == nil {
				editOpsSightSpec.Prometheus = &opssightv1.Prometheus{}
			}
			editOpsSightSpec.Prometheus.Port = opssightPrometheusPort
		case "enable-skyfire":
			editOpsSightSpec.EnableSkyfire = opssightEnableSkyfire
		case "skyfire-name":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.Name = opssightSkyfireName
		case "skyfire-image":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.Image = opssightSkyfireImage
		case "skyfire-port":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.Port = opssightSkyfirePort
		case "skyfire-prometheus-port":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.PrometheusPort = opssightSkyfirePrometheusPort
		case "skyfire-service-account":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.ServiceAccount = opssightSkyfireServiceAccount
		case "skyfire-hub-client-timeout-seconds":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.HubClientTimeoutSeconds = opssightSkyfireHubClientTimeoutSeconds
		case "skyfire-hub-dump-pause-seconds":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.HubDumpPauseSeconds = opssightSkyfireHubDumpPauseSeconds
		case "skyfire-kube-dump-interval-seconds":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.KubeDumpIntervalSeconds = opssightSkyfireKubeDumpIntervalSeconds
		case "skyfire-perceptor-dump-interval-seconds":
			if editOpsSightSpec.Skyfire == nil {
				editOpsSightSpec.Skyfire = &opssightv1.Skyfire{}
			}
			editOpsSightSpec.Skyfire.PerceptorDumpIntervalSeconds = opssightSkyfirePerceptorDumpIntervalSeconds
		case "blackduck-hosts":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.Hosts = opssightBlackduckHosts
		case "blackduck-user":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.User = opssightBlackduckUser
		case "blackduck-port":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.Port = opssightBlackduckPort
		case "blackduck-concurrent-scan-limit":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.ConcurrentScanLimit = opssightBlackduckConcurrentScanLimit
		case "blackduck-total-scan-limit":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.TotalScanLimit = opssightBlackduckTotalScanLimit
		case "blackduck-password-environment-variable":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.PasswordEnvVar = opssightBlackduckPasswordEnvVar
		case "blackduck-initial-count":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.InitialCount = opssightBlackduckInitialCount
		case "blackduck-max-count":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.MaxCount = opssightBlackduckMaxCount
		case "blackduck-delete-blackduck-threshold-percentage":
			if editOpsSightSpec.Blackduck == nil {
				editOpsSightSpec.Blackduck = &opssightv1.Blackduck{}
			}
			editOpsSightSpec.Blackduck.DeleteHubThresholdPercentage = opssightBlackduckDeleteHubThresholdPercentage
		case "enable-metrics":
			editOpsSightSpec.EnableMetrics = opssightEnableMetrics
		case "default-cpu":
			editOpsSightSpec.DefaultCPU = opssightDefaultCPU
		case "default-mem":
			editOpsSightSpec.DefaultMem = opssightDefaultMem
		case "log-level":
			editOpsSightSpec.LogLevel = opssightLogLevel
		case "config-map-name":
			editOpsSightSpec.ConfigMapName = opssightConfigMapName
		case "secret-name":
			editOpsSightSpec.SecretName = opssightSecretName
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)

}

func editAlertFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "alert-registry":
			editAlertSpec.Registry = alertRegistry
		case "image-path":
			editAlertSpec.ImagePath = alertImagePath
		case "alert-image-name":
			editAlertSpec.AlertImageName = alertAlertImageName
		case "alert-image-version":
			editAlertSpec.AlertImageVersion = alertAlertImageVersion
		case "cfssl-image-name":
			editAlertSpec.CfsslImageName = alertCfsslImageName
		case "cfssl-image-version":
			editAlertSpec.CfsslImageVersion = alertCfsslImageVersion
		case "blackduck-host":
			editAlertSpec.BlackduckHost = alertBlackduckHost
		case "blackduck-user":
			editAlertSpec.BlackduckUser = alertBlackduckUser
		case "blackduck-port":
			editAlertSpec.BlackduckPort = &alertBlackduckPort
		case "port":
			editAlertSpec.Port = &alertPort
		case "stand-alone":
			editAlertSpec.StandAlone = &alertStandAlone
		case "alert-memory":
			editAlertSpec.AlertMemory = alertAlertMemory
		case "cfssl-memory":
			editAlertSpec.CfsslMemory = alertCfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
