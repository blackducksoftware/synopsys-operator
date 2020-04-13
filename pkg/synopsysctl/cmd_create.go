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

	alertctl "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	"github.com/blackducksoftware/synopsys-operator/pkg/bdba"
	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	polarisreporting "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"
	polarisreportingctl "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"

	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Create Command CRSpecBuilderFromCobraFlagsInterface
var createAlertCobraHelper alertctl.HelmValuesFromCobraFlags
var createBlackDuckCobraHelper blackduck.HelmValuesFromCobraFlags
var createOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface
var createPolarisCobraHelper polaris.HelmValuesFromCobraFlags
var createPolarisReportingCobraHelper polarisreporting.HelmValuesFromCobraFlags
var createBDBACobraHelper bdba.HelmValuesFromCobraFlags

// Default Base Specs for Create
var baseAlertSpec string
var baseBlackDuckSpec string
var baseOpsSightSpec string

var namespace string

var alertNativePVC bool

// createCmd creates a Synopsys resource in the cluster
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Synopsys resource in your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

/*
Create Alert Commands
*/

// createCmd creates an Alert instance
var createAlertCmd = &cobra.Command{
	Use:           "alert NAME",
	Example:       "synopsysctl create alert <name>\nsynopsysctl create alert <name> -n <namespace>\nsynopsysctl create alert <name> --mock json",
	Short:         "Create an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		alertName := fmt.Sprintf("%s%s", args[0], AlertPostSuffix)

		// Get the flags to set Helm values
		helmValuesMap, err := createAlertCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			alertChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				alertChartRepository = fmt.Sprintf("%s/charts/alert-helmchart-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(alertName, namespace, alertChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf(strings.Replace(fmt.Sprintf("failed to create Alert resources: %+v", err), fmt.Sprintf("release '%s' ", alertName), fmt.Sprintf("release '%s' ", args[0]), 0))
		}

		// Create secrets for Alert
		customCertificateSecret := map[string]runtime.Object{}
		javaKeystoreSecret := map[string]runtime.Object{}
		certificateFlag := cmd.Flag("certificate-file-path")
		certificateKeyFlag := cmd.Flag("certificate-key-file-path")
		if certificateFlag.Changed && certificateKeyFlag.Changed {
			certificateData, err := util.ReadFileData(certificateFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read certificate file: %+v", err)
			}

			certificateKeyData, err := util.ReadFileData(certificateKeyFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read certificate file: %+v", err)
			}
			customCertificateSecretName := "alert-custom-certificate"
			customCertificateSecret, err = alertctl.GetAlertCustomCertificateSecret(namespace, customCertificateSecretName, certificateData, certificateKeyData)
			util.SetHelmValueInMap(helmValuesMap, []string{"webserverCustomCertificatesSecretName"}, customCertificateSecretName)
		}

		javaKeystoreFlag := cmd.Flag("java-keystore-file-path")
		if javaKeystoreFlag.Changed {
			javaKeystoreData, err := util.ReadFileData(javaKeystoreFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read Java Keystore file: %+v", err)
			}
			javaKeystoreSecretName := "alert-java-keystore"
			javaKeystoreSecret, err = alertctl.GetAlertJavaKeystoreSecret(namespace, javaKeystoreSecretName, javaKeystoreData)
			util.SetHelmValueInMap(helmValuesMap, []string{"javaKeystoreSecretName"}, javaKeystoreSecretName)
		}

		// If mock mode, return and don't create resources
		if mockMode {
			_, err = PrintComponent(helmValuesMap, "YAML")
			return err
		}

		// Deploy the Secrets
		if len(customCertificateSecret) > 0 {
			err = KubectlApplyRuntimeObjects(customCertificateSecret)
			if err != nil {
				return fmt.Errorf("failed to deploy the customCertificateSecret Secrets: %s", err)
			}
		}
		if len(javaKeystoreSecret) > 0 {
			err = KubectlApplyRuntimeObjects(javaKeystoreSecret)
			if err != nil {
				return fmt.Errorf("failed to deploy the javaKeystoreSecret Secrets: %s", err)
			}
		}

		// Deploy Alert Resources
		err = util.CreateWithHelm3(alertName, namespace, alertChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf(strings.Replace(fmt.Sprintf("failed to create Alert resources: %+v", err), fmt.Sprintf("release '%s' ", alertName), fmt.Sprintf("release '%s' ", args[0]), 0))
		}

		log.Infof("Alert has been successfully Created!")
		return nil
	},
}

// createAlertNativeCmd prints the Kubernetes resources for creating an Alert instance
var createAlertNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create alert native <name>\nsynopsysctl create alert native <name> -n <namespace>\nsynopsysctl create alert native <name> -o yaml",
	Short:         "Print the Kubernetes resources for creating an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName := fmt.Sprintf("%s%s", args[0], AlertPostSuffix)

		// Get the flags to set Helm values
		helmValuesMap, err := createAlertCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			alertChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				alertChartRepository = fmt.Sprintf("%s/charts/synopsys-alert-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Get secrets for Alert
		customCertificateSecret := map[string]runtime.Object{}
		javaKeystoreSecret := map[string]runtime.Object{}
		certificateFlag := cmd.Flag("certificate-file-path")
		certificateKeyFlag := cmd.Flag("certificate-key-file-path")
		if certificateFlag.Changed && certificateKeyFlag.Changed {
			certificateData, err := util.ReadFileData(certificateFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read certificate file: %+v", err)
			}

			certificateKeyData, err := util.ReadFileData(certificateKeyFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read certificate file: %+v", err)
			}
			customCertificateSecretName := "alert-custom-certificate"
			customCertificateSecret, err = alertctl.GetAlertCustomCertificateSecret(namespace, customCertificateSecretName, certificateData, certificateKeyData)
			util.SetHelmValueInMap(helmValuesMap, []string{"webserverCustomCertificatesSecretName"}, customCertificateSecretName)
		}

		javaKeystoreFlag := cmd.Flag("java-keystore-file-path")
		if javaKeystoreFlag.Changed {
			javaKeystoreData, err := util.ReadFileData(javaKeystoreFlag.Value.String())
			if err != nil {
				log.Fatalf("failed to read Java Keystore file: %+v", err)
			}
			javaKeystoreSecretName := "alert-java-keystore"
			javaKeystoreSecret, err = alertctl.GetAlertJavaKeystoreSecret(namespace, javaKeystoreSecretName, javaKeystoreData)
			util.SetHelmValueInMap(helmValuesMap, []string{"javaKeystoreSecretName"}, javaKeystoreSecretName)
		}
		// Print the Secrets
		if len(customCertificateSecret) > 0 {
			_, err = PrintComponent(customCertificateSecret, "YAML")
			return err
		}
		if len(javaKeystoreSecret) > 0 {
			_, err = PrintComponent(customCertificateSecret, "YAML")
			return err
		}

		// Deploy Alert Resources
		err = util.TemplateWithHelm3(alertName, namespace, alertChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to create Alert resources: %+v", err)
		}

		return nil
	},
}

/*
Create Black Duck Commands
*/

func checkPasswords(flagset *pflag.FlagSet) {
	if flagset.Lookup("admin-password").Changed ||
		flagset.Lookup("user-password").Changed {
		// user is explicitly required to set the postgres passwords for: 'admin', 'postgres', and 'user'
		cobra.MarkFlagRequired(flagset, "admin-password")
		cobra.MarkFlagRequired(flagset, "user-password")
	} else {
		// require all external-postgres parameters
		cobra.MarkFlagRequired(flagset, "external-postgres-host")
		cobra.MarkFlagRequired(flagset, "external-postgres-port")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin")
		cobra.MarkFlagRequired(flagset, "external-postgres-user")
		cobra.MarkFlagRequired(flagset, "external-postgres-ssl")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin-password")
		cobra.MarkFlagRequired(flagset, "external-postgres-user-password")
	}
}

func checkSealKey(flagset *pflag.FlagSet) {
	cobra.MarkFlagRequired(flagset, "seal-key")
}

// createBlackDuckCmd creates a Black Duck instance
var createBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME -n NAMESPACE",
	Example:       "synopsysctl create blackduck <name> -n <namespace>",
	Short:         "Create a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument, but got %+v", args)
		}
		checkPasswords(cmd.Flags())
		cobra.MarkFlagRequired(cmd.Flags(), "certificate-file-path")
		cobra.MarkFlagRequired(cmd.Flags(), "certificate-key-file-path")
		checkSealKey(cmd.Flags())
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		helmValuesMap, err := createBlackDuckCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			blackduckChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				blackduckChartRepository = fmt.Sprintf("%s/charts/blackduck-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		secrets, err := blackduck.GetCertsFromFlagsAndSetHelmValue(args[0], namespace, cmd.Flags(), helmValuesMap)
		if err != nil {
			return err
		}
		for _, v := range secrets {
			if _, err := kubeClient.CoreV1().Secrets(namespace).Create(&v); err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create certifacte secret: %+v", err)
			}
		}

		var extraFiles []string
		size, found := helmValuesMap["size"]
		if found {
			extraFiles = append(extraFiles, fmt.Sprintf("%s.yaml", size.(string)))
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(args[0], namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath, true, extraFiles...)
		if err != nil {
			return fmt.Errorf("failed to create Blackduck resources: %+v", err)
		}

		// Deploy Resources
		err = util.CreateWithHelm3(args[0], namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath, false, extraFiles...)
		if err != nil {
			return fmt.Errorf("failed to create Blackduck resources: %+v", err)
		}
		return nil
	},
}

// createBlackDuckNativeCmd prints the Kubernetes resources for creating a Black Duck instance
var createBlackDuckNativeCmd = &cobra.Command{
	Use:           "native NAME -n NAMESPACE",
	Example:       "synopsysctl create blackduck native <name> -n <namespace>",
	Short:         "Print the Kubernetes resources for creating a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument, but got %+v", args)
		}
		checkPasswords(cmd.Flags())
		cobra.MarkFlagRequired(cmd.Flags(), "certificate-file-path")
		cobra.MarkFlagRequired(cmd.Flags(), "certificate-key-file-path")
		checkSealKey(cmd.Flags())
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		helmValuesMap, err := createBlackDuckCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			blackduckChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				blackduckChartRepository = fmt.Sprintf("%s/charts/blackduck-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		secrets, err := blackduck.GetCertsFromFlagsAndSetHelmValue(args[0], namespace, cmd.Flags(), helmValuesMap)
		if err != nil {
			return err
		}
		for _, v := range secrets {
			PrintComponent(v, "YAML") // helm only supports yaml
		}

		// Check Dry Run before deploying any resources
		err = util.TemplateWithHelm3(args[0], namespace, blackduckChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to create Blackduck resources: %+v", err)
		}

		return nil
	},
}

/*
Create OpsSight Commands
*/

var createOpsSightPreRun = func(cmd *cobra.Command, args []string) error {
	// Set the base spec
	if !cmd.Flags().Lookup("template").Changed {
		baseOpsSightSpec = defaultBaseOpsSightSpec
	}
	log.Debugf("setting OpsSight's base spec to '%s'", baseOpsSightSpec)
	err := createOpsSightCobraHelper.SetPredefinedCRSpec(baseOpsSightSpec)
	if err != nil {
		cmd.Help()
		return err
	}
	return nil
}

func updateOpsSightSpecWithFlags(cmd *cobra.Command, opsSightName string, opsSightNamespace string) (*opssightv1.OpsSight, error) {
	// Update Spec with user's flags
	log.Debugf("updating spec with user's flags")
	opsSightInterface, err := createOpsSightCobraHelper.GenerateCRSpecFromFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Set Namespace in Spec
	opsSightSpec, _ := opsSightInterface.(opssightv1.OpsSightSpec)
	opsSightSpec.Namespace = opsSightNamespace

	// Create and Deploy OpsSight CRD
	opsSight := &opssightv1.OpsSight{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opsSightName,
			Namespace: opsSightNamespace,
		},
		Spec: opsSightSpec,
	}
	opsSight.Kind = "OpsSight"
	opsSight.APIVersion = "synopsys.com/v1"
	return opsSight, nil
}

// createOpsSightCmd creates an OpsSight instance
var createOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME",
	Example:       "synopsysctl create opssight <name>\nsynopsysctl create opssight <name> -n <namespace>\nsynopsysctl create opssight <name> --mock json",
	Short:         "Create an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 arguments")
		}
		return nil
	},
	PreRunE: createOpsSightPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		opsSightName := args[0]
		opsSightNamespace, crdNamespace, _, err := getInstanceInfo(mockMode, util.OpsSightCRDName, "", namespace, opsSightName)
		if err != nil {
			return err
		}
		opsSight, err := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*opsSight, mockFormat, false)
		}

		log.Infof("creating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

		// Deploy the OpsSight instance
		_, err = util.CreateOpsSight(opsSightClient, crdNamespace, opsSight)
		if err != nil {
			return fmt.Errorf("error creating the OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		log.Infof("successfully submitted OpsSight '%s' into namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// createOpsSightNativeCmd prints the Kubernetes resources for creating an OpsSight instance
var createOpsSightNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl create opssight native <name>\nsynopsysctl create opssight native <name> -n <namespace>\nsynopsysctl create opssight native <name> -o yaml",
	Short:         "Print the Kubernetes resources for creating an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	PreRunE: createOpsSightPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
		opsSightNamespace, _, _, err := getInstanceInfo(true, util.OpsSightCRDName, "", namespace, opsSightName)
		if err != nil {
			return err
		}
		opsSight, err := updateOpsSightSpecWithFlags(cmd, opsSightName, opsSightNamespace)
		if err != nil {
			return err
		}

		log.Debugf("generating Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*opsSight, nativeFormat, true)
	},
}

// createPolarisCmd creates a Polaris instance
var createPolarisCmd = &cobra.Command{
	Use:           "polaris -n NAMESPACE",
	Short:         "Create a Polaris instance. (Please make sure you have read and understand prerequisites before installing Polaris: https://sig-confluence.internal.synopsys.com/display/DD/Polaris+on-premises])",
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: "\nRequried flags for setup with external database:\n\n 	synopsysctl create polaris --namespace 'onprem' --version '2020.03' --gcp-service-account-path '<PATH>/gcp-service-account-file.json' --coverity-license-path '<PATH>/coverity-license-file.xml' --fqdn 'example.polaris.com' --smtp-host 'example.smtp.com' --smtp-port 25 --smtp-username 'example' --smtp-password 'example' --smtp-sender-email 'example.email.com' --postgres-host 'example.postgres.com' --postgres-port 5432 --postgres-username 'example' --postgres-password 'example' ",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		// TODO: allow user to specify --version and --chart-location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf("failed to create Polaris resources: %+v", err)
		}

		// Deploy Polaris Resources
		err = util.CreateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf("failed to create Polaris resources: %+v", err)
		}

		log.Infof("Polaris has been successfully Created!")
		return nil
	},
}

// createPolarisNativeCmd prints the Kubernetes resources for creating a Polaris instance
var createPolarisNativeCmd = &cobra.Command{
	Use:           "native -n NAMESPACE",
	Short:         "Print Kubernetes resources for creating a Polaris instance (Please make sure you have read and understand prerequisites before installing Polaris: https://sig-confluence.internal.synopsys.com/display/DD/Polaris+on-premises])",
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: "\nRequried flags for setup with external database:\n\n 	synopsysctl create polaris native --namespace 'onprem' --version '2020.04' --gcp-service-account-path '<PATH>/gcp-service-account-file.json' --coverity-license-path '<PATH>/coverity-license-file.xml' --fqdn 'example.polaris.com' --smtp-host 'example.smtp.com' --smtp-port 25 --smtp-username 'example' --smtp-password 'example' --smtp-sender-email 'example.email.com' --postgres-host 'example.postgres.com' --postgres-port 5432 --postgres-username 'example' --postgres-password 'example' ",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Print Polaris Resources
		err = util.TemplateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to generate Polaris resources: %+v", err)
		}

		return nil
	},
}

func polarisPostgresCheck(flagset *pflag.FlagSet) error {
	usingPostgresContainer, _ := flagset.GetBool("enable-postgres-container")
	if usingPostgresContainer {
		if flagset.Lookup("postgres-host").Changed || flagset.Lookup("postgres-port").Changed || flagset.Lookup("postgres-username").Changed {
			return fmt.Errorf("cannot change the host, port and username when using the postgres container")
		}
		if flagset.Lookup("postgres-ssl-mode").Changed {
			return fmt.Errorf("cannot enable SSL when using postgres container")
		}
	} else {
		if flagset.Lookup("postgres-size").Changed {
			return fmt.Errorf("cannot configure the postgresql size when using an external database")
		}
		// External DB. Host, port and username are mandatory
		cobra.MarkFlagRequired(flagset, "postgres-host")
		cobra.MarkFlagRequired(flagset, "postgres-port")
		cobra.MarkFlagRequired(flagset, "postgres-username")
	}

	return nil
}

// createPolarisReportingCmd creates a Polaris-Reporting instance
var createPolarisReportingCmd = &cobra.Command{
	Use:           "polaris-reporting -n NAMESPACE",
	Example:       "synopsysctl create polaris-reporting -n <namespace>",
	Short:         "Create a Polaris-Reporting instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisReportingCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		// TODO: allow user to specify --version and --chart-location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisReportingChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf("failed to create Polaris-Reporting resources: %+v", err)
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreportingctl.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to create GCP Service Account Secrets: %+v", err)
		}

		// Deploy the Secret
		err = KubectlApplyRuntimeObjects(gcpServiceAccountSecrets)
		if err != nil {
			return fmt.Errorf("failed to deploy the gcpServiceAccount Secrets: %s", err)
		}

		// Deploy Polaris-Reporting Resources
		err = util.CreateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf("failed to create Polaris-Reporting resources: %+v", err)
		}

		log.Infof("Polaris-Reporting has been successfully Created!")
		return nil
	},
}

// createPolarisReportingNativeCmd prints Polaris-Reporting resources
var createPolarisReportingNativeCmd = &cobra.Command{
	Use:           "native -n NAMESPACE",
	Example:       "synopsysctl create polaris-reporting native -n <namespace>",
	Short:         "Print Kubernetes resources for creating a Polaris-Reporting instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createPolarisReportingCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			polarisReportingChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreportingctl.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to create GCP Service Account Secrets: %+v", err)
		}

		// Print the Secret
		for _, obj := range gcpServiceAccountSecrets {
			PrintComponent(obj, "YAML") // helm only supports yaml
		}

		// Print Polaris-Reporting Resources
		err = util.TemplateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to generate Polaris-Reporting resources: %+v", err)
		}

		return nil
	},
}

// createBDBACmd creates a BDBA instance
var createBDBACmd = &cobra.Command{
	Use:           "bdba -n NAMESPACE",
	Example:       "synopsysctl create bdba -n <namespace>",
	Short:         "Create a BDBA instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 arguments, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createBDBACobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		// TODO: allow user to specify --version and --chart-location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			bdbaChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				bdbaChartRepository = fmt.Sprintf("%s/charts/bdba-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Check Dry Run before deploying any resources
		err = util.CreateWithHelm3(bdbaName, namespace, bdbaChartRepository, helmValuesMap, kubeConfigPath, true)
		if err != nil {
			return fmt.Errorf("failed to create BDBA resources: %+v", err)
		}

		// Deploy Resources
		err = util.CreateWithHelm3(bdbaName, namespace, bdbaChartRepository, helmValuesMap, kubeConfigPath, false)
		if err != nil {
			return fmt.Errorf("failed to create BDBA resources: %+v", err)
		}

		log.Infof("BDBA has been successfully Created!")
		return nil
	},
}

// createBDBANativeCmd prints BDBA resources
var createBDBANativeCmd = &cobra.Command{
	Use:           "native -n NAMESPACE",
	Example:       "synopsysctl create bdba -n <namespace>",
	Short:         "Print Kubernetes resources for creating a BDBA instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check the Number of Arguments
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 0 argument, but got %+v", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the flags to set Helm values
		helmValuesMap, err := createBDBACobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			bdbaChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				bdbaChartRepository = fmt.Sprintf("%s/charts/bdba-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Print Resources
		err = util.TemplateWithHelm3(bdbaName, namespace, bdbaChartRepository, helmValuesMap)
		if err != nil {
			return fmt.Errorf("failed to generate BDBA resources: %+v", err)
		}

		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	createBlackDuckCobraHelper = *blackduck.NewHelmValuesFromCobraFlags()
	createAlertCobraHelper = *alertctl.NewHelmValuesFromCobraFlags()
	createOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	createPolarisCobraHelper = *polaris.NewHelmValuesFromCobraFlags()
	createPolarisReportingCobraHelper = *polarisreporting.NewHelmValuesFromCobraFlags()
	createBDBACobraHelper = *bdba.NewHelmValuesFromCobraFlags()

	rootCmd.AddCommand(createCmd)

	// Add Alert Command
	createAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createAlertCmd.PersistentFlags(), "namespace")
	createAlertCobraHelper.AddCobraFlagsToCommand(createAlertCmd, true)
	addChartLocationPathFlag(createAlertCmd)
	addMockFlag(createAlertCmd)
	createCmd.AddCommand(createAlertCmd)

	createAlertCobraHelper.AddCobraFlagsToCommand(createAlertNativeCmd, true)
	addChartLocationPathFlag(createAlertNativeCmd)
	createAlertCmd.AddCommand(createAlertNativeCmd)

	// Add Black Duck Command
	createBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createBlackDuckCmd.PersistentFlags(), "namespace")
	addChartLocationPathFlag(createBlackDuckCmd)
	createBlackDuckCobraHelper.AddCRSpecFlagsToCommand(createBlackDuckCmd, true)
	createCmd.AddCommand(createBlackDuckCmd)

	createBlackDuckCobraHelper.AddCRSpecFlagsToCommand(createBlackDuckNativeCmd, true)
	addChartLocationPathFlag(createBlackDuckNativeCmd)
	createBlackDuckCmd.AddCommand(createBlackDuckNativeCmd)

	// Add OpsSight Command
	createOpsSightCmd.PersistentFlags().StringVar(&baseOpsSightSpec, "template", baseOpsSightSpec, "Base resource configuration to modify with flags [empty|upstream|default|disabledBlackDuck]")
	createOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightCmd, true)
	addMockFlag(createOpsSightCmd)
	createCmd.AddCommand(createOpsSightCmd)

	createOpsSightCobraHelper.AddCRSpecFlagsToCommand(createOpsSightNativeCmd, true)
	addNativeFormatFlag(createOpsSightNativeCmd)
	createOpsSightCmd.AddCommand(createOpsSightNativeCmd)

	// Add Polaris commands
	createPolarisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createPolarisCmd.PersistentFlags(), "namespace")
	createPolarisCobraHelper.AddCobraFlagsToCommand(createPolarisCmd, true)
	addChartLocationPathFlag(createPolarisCmd)
	createCmd.AddCommand(createPolarisCmd)

	createPolarisCobraHelper.AddCobraFlagsToCommand(createPolarisNativeCmd, true)
	addChartLocationPathFlag(createPolarisNativeCmd)
	createPolarisCmd.AddCommand(createPolarisNativeCmd)

	// Add Polaris-Reporting commands
	createPolarisReportingCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createPolarisReportingCmd.PersistentFlags(), "namespace")
	createPolarisReportingCobraHelper.AddCobraFlagsToCommand(createPolarisReportingCmd, true)
	addChartLocationPathFlag(createPolarisReportingCmd)
	createCmd.AddCommand(createPolarisReportingCmd)

	createPolarisReportingCobraHelper.AddCobraFlagsToCommand(createPolarisReportingNativeCmd, true)
	addChartLocationPathFlag(createPolarisReportingNativeCmd)
	createPolarisReportingCmd.AddCommand(createPolarisReportingNativeCmd)

	// Add BDBA commands
	createBDBACmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(createBDBACmd.PersistentFlags(), "namespace")
	createBDBACobraHelper.AddCobraFlagsToCommand(createBDBACmd, true)
	addChartLocationPathFlag(createBDBACmd)
	createCmd.AddCommand(createBDBACmd)

	createBDBACobraHelper.AddCobraFlagsToCommand(createBDBANativeCmd, true)
	addChartLocationPathFlag(createBDBANativeCmd)
	createBDBACmd.AddCommand(createBDBANativeCmd)

}
