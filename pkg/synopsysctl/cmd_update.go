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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update Command ResourceCtlSpecBuilders
var updateAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updateBlackDuckCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updateOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface

var updateMockFormat = "json"

// updateCmd provides functionality to update/upgrade features of
// Synopsys resources
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

/*
Update Operator Commands
*/

// getUpdatedOperator returns a SpecConfig for Synopsys Operator with the updates provided by the user
func getUpdatedOperator(currOperatorSpec *soperator.SpecConfig, cmd *cobra.Command) (*soperator.SpecConfig, []string, error) {

	newCrds := make([]string, 0)
	namespace := currOperatorSpec.Namespace

	// convert crds to CRD map for easy comparison
	crdMap := make(map[string]string, 0)

	crds := currOperatorSpec.Crds
	for _, crd := range crds {
		crdMap[strings.TrimSpace(crd)] = strings.TrimSpace(crd)
	}

	newOperatorSpec := soperator.SpecConfig{}

	// update Spec with changes from user's flags
	if cmd.Flag("synopsys-operator-image").Changed {
		log.Debugf("updating Synopsys Operator's image to '%s'", synopsysOperatorImage)
		// check Synopsys Operator image
		if _, err := util.ValidateImageString(synopsysOperatorImage); err != nil {
			return nil, nil, err
		}
		newOperatorSpec.Image = synopsysOperatorImage
	}
	if cmd.Flag("expose-ui").Changed {
		log.Debugf("updating expose ui")
		isValid := util.IsExposeServiceValid(exposeUI)
		if !isValid {
			cmd.Help()
			return nil, nil, fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
		newOperatorSpec.Expose = exposeUI
	}
	if cmd.Flag("postgres-restart-in-minutes").Changed {
		log.Debugf("updating postgres restart in minutes")
		newOperatorSpec.PostgresRestartInMins = postgresRestartInMins
	}
	if cmd.Flag("pod-wait-timeout-in-seconds").Changed {
		log.Debugf("updating pod wait timeout in seconds")
		newOperatorSpec.PodWaitTimeoutSeconds = podWaitTimeoutSeconds
	}
	if cmd.Flag("resync-interval-in-seconds").Changed {
		log.Debugf("updating resync interval in seconds")
		newOperatorSpec.ResyncIntervalInSeconds = resyncIntervalInSeconds
	}
	if cmd.Flag("postgres-termination-grace-period").Changed {
		log.Debugf("updating postgres termination grace period")
		newOperatorSpec.TerminationGracePeriodSeconds = terminationGracePeriodSeconds
	}
	if cmd.Flag("dry-run").Changed {
		log.Debugf("updating dry run")
		newOperatorSpec.DryRun = (strings.ToUpper(dryRun) == "TRUE")
	}
	if cmd.Flag("log-level").Changed {
		log.Debugf("updating log level")
		newOperatorSpec.LogLevel = logLevel
	}
	if cmd.Flag("no-of-threads").Changed {
		log.Debugf("updating no of threads")
		newOperatorSpec.Threadiness = threadiness
	}

	// validate whether Alert CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.Flags().Lookup("enable-alert").Changed {
		log.Debugf("updating enable Alert")
		_, ok := crdMap[util.AlertCRDName]
		if ok && isEnabledAlert {
			log.Errorf("Custom Resource Definition '%s' already exists...", util.AlertCRDName)
		} else if !ok && isEnabledAlert {
			// create CRD
			crds = append(crds, util.AlertCRDName)
			newCrds = append(newCrds, util.AlertCRDName)
		} else {
			// check whether the CRD can be deleted
			err := isDeleteCrd(util.AlertCRDName, namespace)
			if err != nil {
				log.Warn(err)
			} else {
				// delete CRD
				deleteCrd(util.AlertCRDName, namespace)
			}
			// remove it from crds, so that the CRD controller won't run
			crds = util.RemoveFromStringSlice(crds, util.AlertCRDName)
		}
	}

	// validate whether Black Duck CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.Flags().Lookup("enable-blackduck").Changed {
		log.Debugf("updating enable Black Duck")
		_, ok := crdMap[util.BlackDuckCRDName]
		if ok && isEnabledBlackDuck {
			log.Errorf("Custom Resource Definition '%s' already exists...", util.BlackDuckCRDName)
		} else if !ok && isEnabledBlackDuck {
			// create CRD
			crds = append(crds, util.BlackDuckCRDName)
			newCrds = append(newCrds, util.BlackDuckCRDName)
		} else {
			// check whether the CRD can be deleted
			err := isDeleteCrd(util.BlackDuckCRDName, namespace)
			if err != nil {
				log.Warn(err)
			} else {
				// delete CRD
				deleteCrd(util.BlackDuckCRDName, namespace)
			}
			// remove it from crds, so that the CRD controller won't run
			crds = util.RemoveFromStringSlice(crds, util.BlackDuckCRDName)
		}
	}

	// validate whether OpsSight CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.Flags().Lookup("enable-opssight").Changed {
		log.Debugf("updating enable OpsSight")
		_, ok := crdMap[util.OpsSightCRDName]
		if ok && isEnabledOpsSight {
			log.Errorf("Custom Resource Definition '%s' already exists...", util.OpsSightCRDName)
		} else if !ok && isEnabledOpsSight {
			// create CRD
			crds = append(crds, util.OpsSightCRDName)
			newCrds = append(newCrds, util.OpsSightCRDName)
		} else {
			// check whether the CRD can be deleted
			err := isDeleteCrd(util.OpsSightCRDName, namespace)
			if err != nil {
				log.Warn(err)
			} else {
				// delete CRD
				deleteCrd(util.OpsSightCRDName, namespace)
			}
			// remove it from crds, so that the CRD controller won't run
			crds = util.RemoveFromStringSlice(crds, util.OpsSightCRDName)
		}
	}

	newOperatorSpec.Crds = crds
	// HACK because of mergo merge issue
	if len(newOperatorSpec.Crds) == 0 {
		currOperatorSpec.Crds = newOperatorSpec.Crds
	}
	newOperatorSpec.IsClusterScoped = currOperatorSpec.IsClusterScoped

	// merge old and new data
	err := mergo.Merge(&newOperatorSpec, currOperatorSpec)
	if err != nil {
		return nil, newCrds, fmt.Errorf("unable to merge old and new Synopsys Operator's info due to %+v", err)
	}

	return &newOperatorSpec, newCrds, err
}

// updateOperatorCmd updates Synopsys Operator
var updateOperatorCmd = &cobra.Command{
	Use:           "operator",
	Example:       "synopsysctl update operator --synopsys-operator-image docker.io/new_image_url\nsynopsysctl update operator --enable-blackduck\nsynopsysctl update operator --enable-blackduck -n <namespace>\nsynopsysctl update operator --expose-ui OPENSHIFT",
	Short:         "Update Synopsys Operator",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cmd.Help()
			return fmt.Errorf("this command doesn't take any arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		var isClusterScoped bool
		// Set namespace if one wasn't provided
		if !cmd.Flags().Lookup("namespace").Changed {
			// set existing Synopsys Operator namespace else use default
			isClusterScoped = util.GetClusterScope(apiExtensionClient)
			if isClusterScoped {
				namespaces, err := util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
				if err != nil {
					return err
				}
				if len(namespaces) > 1 {
					return fmt.Errorf("more than 1 Synopsys Operator found in your cluster. please pass the namespace of the Synopsys Operator that you want to update")
				}
				namespace = namespaces[0]
			} else {
				namespace = DefaultOperatorNamespace
			}
		}

		currOperatorSpec, err := soperator.GetOldOperatorSpec(restconfig, kubeClient, namespace)
		if err != nil {
			return err
		}

		newOperatorSpec, newCrds, err := getUpdatedOperator(currOperatorSpec, cmd)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating the updated Spec for Synopsys Operator in namespace '%s'...", operatorNamespace)
			newOperatorSpec.RestConfig = nil // assigning the rest config to nil to run in mock mode. Getting weird issue if it is not nil
			return PrintResource(newOperatorSpec, mockFormat, false)
		}

		log.Infof("updating Synopsys Operator in namespace '%s'...", namespace)
		// create custom resource definitions
		isClusterScoped = newOperatorSpec.IsClusterScoped
		crdConfigs, err := getCrdConfigs(namespace, isClusterScoped, newCrds)
		if err != nil {
			return err
		}
		if len(crdConfigs) > 0 {
			err = deployCrds(namespace, isClusterScoped, crdConfigs)
			if err != nil {
				return err
			}
		}

		sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
		// update Synopsys Operator
		err = sOperatorCreater.EnsureSynopsysOperator(namespace, blackDuckClient, opsSightClient, alertClient, currOperatorSpec, newOperatorSpec)
		if err != nil {
			return fmt.Errorf("unable to update Synopsys Operator due to %+v", err)
		}
		log.Debugf("updating Prometheus in namespace '%s'", namespace)
		// Create new Prometheus SpecConfig
		currPrometheusSpec, err := soperator.GetOldPrometheusSpec(restconfig, kubeClient, namespace)
		if err != nil {
			return fmt.Errorf("error in updating Prometheus due to %+v", err)
		}
		// check for changes
		newPrometheusSpec := soperator.PrometheusSpecConfig{}
		if cmd.Flag("metrics-image").Changed {
			log.Debugf("updating Prometheus' image to '%s'", metricsImage)
			newPrometheusSpec.Image = metricsImage
		}
		if cmd.Flag("expose-metrics").Changed {
			log.Debugf("updating expose metrics")
			isValid := util.IsExposeServiceValid(exposeMetrics)
			if !isValid {
				cmd.Help()
				return fmt.Errorf("expose metrics must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
			}
			newPrometheusSpec.Expose = exposeMetrics
		}
		// merge old and new data
		err = mergo.Merge(&newPrometheusSpec, currPrometheusSpec)
		if err != nil {
			return fmt.Errorf("unable to merge old and new Prometheus' info due to %+v", err)
		}
		// update prometheus
		err = sOperatorCreater.UpdatePrometheus(&newPrometheusSpec)
		if err != nil {
			return fmt.Errorf("unable to update Prometheus due to %+v", err)
		}
		log.Infof("successfully submitted updates to Synopsys Operator in namespace '%s'", namespace)
		return nil
	},
}

// updateOperatorNativeCmd prints the Kubernetes resources with updates to a Synopsys Operator instance
var updateOperatorNativeCmd = &cobra.Command{
	Use:           "native",
	Example:       "synopsysctl update operator native --synopsys-operator-image docker.io/new_image_url\nsynopsysctl update operator native --enable-blackduck\nsynopsysctl update operator native --enable-blackduck -n <namespace>\nsynopsysctl update operator native --expose-ui OPENSHIFT\nsynopsysctl update operator native -o yaml",
	Short:         "Print the Kubernetes resources with updates to a Synopsys Operator instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cmd.Help()
			return fmt.Errorf("this command doesn't take any arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var isClusterScoped bool
		// Set namespace if one wasn't provided
		if !cmd.Flags().Lookup("namespace").Changed {
			// set existing Synopsys Operator namespace else use default
			isClusterScoped = util.GetClusterScope(apiExtensionClient)
			if isClusterScoped {
				namespaces, err := util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
				if err != nil {
					return err
				}
				if len(namespaces) > 1 {
					return fmt.Errorf("more than 1 Synopsys Operator found in your cluster. please pass the namespace of the Synopsys Operator that you want to update")
				}
				namespace = namespaces[0]
			} else {
				namespace = DefaultOperatorNamespace
			}
		}

		currOperatorSpec, err := soperator.GetOldOperatorSpec(restconfig, kubeClient, namespace)
		if err != nil {
			return err
		}

		newOperatorSpec, _, err := getUpdatedOperator(currOperatorSpec, cmd)
		if err != nil {
			return err
		}

		log.Debugf("generating the updated Kubernetes resources for Synopsys Operator in namespace '%s'...", operatorNamespace)
		return PrintResource(*newOperatorSpec, nativeFormat, true)
	},
}

/*
Update Alert Commands
*/

func updateAlert(alt *alertapi.Alert, flagset *pflag.FlagSet) (*alertapi.Alert, error) {
	updateAlertCobraHelper.SetCRSpec(alt.Spec)
	alertInterface, err := updateAlertCobraHelper.GenerateCRSpecFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	newSpec := alertInterface.(alertapi.AlertSpec)
	newSpec.Environs = util.MergeEnvSlices(newSpec.Environs, alt.Spec.Environs)
	alt.Spec = newSpec
	return alt, nil
}

// updateAlertCmd updates an Alert instance
var updateAlertCmd = &cobra.Command{
	Use:           "alert NAME",
	Example:       "synopsysctl update alert <name> --port 80\nsynopsysctl update alert <name> -n <namespace> --port 80\nsynopsysctl update alert <name> --mock json",
	Short:         "Update an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		alertName, alertNamespace, crdnamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, args[0])
		if err != nil {
			return err
		}
		currAlert, err := util.GetAlert(alertClient, crdnamespace, alertName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
		}
		newAlert, err := updateAlert(currAlert, cmd.Flags())
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
			return PrintResource(*currAlert, mockFormat, false)
		}

		log.Infof("updating Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		// update the namespace label if the version of the app got changed
		_, err = util.CheckAndUpdateNamespace(kubeClient, util.AlertName, alertNamespace, alertName, newAlert.Spec.Version, false)
		if err != nil {
			return err
		}
		// Update the Alert
		_, err = util.UpdateAlert(alertClient, crdnamespace, newAlert)
		if err != nil {
			return fmt.Errorf("error updating Alert '%s' due to %+v", newAlert.Name, err)
		}
		log.Infof("successfully submitted updates to Alert '%s' in namespace '%s'", alertName, alertNamespace)
		return nil
	},
}

// updateAlertNativeCmd prints the Kubernetes resources with updates to an Alert instance
var updateAlertNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsysctl update alert native <name> --port 80\nsynopsysctl update alert native <name> -n <namespace> --port 80\nsynopsysctl update alert native <name> -o yaml",
	Short:         "Print the Kubernetes resources with updates to an Alert instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName, alertNamespace, crdnamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, args[0])
		if err != nil {
			return err
		}
		currAlert, err := util.GetAlert(alertClient, crdnamespace, alertName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
		}
		newAlert, err := updateAlert(currAlert, cmd.Flags())
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for Alert '%s' in namespace '%s'...", alertName, alertNamespace)
		return PrintResource(*newAlert, nativeFormat, true)
	},
}

/*
Update Black Duck Commands
*/

func updateBlackDuck(bd *blackduckapi.Blackduck, flagset *pflag.FlagSet) (*blackduckapi.Blackduck, error) {
	updateBlackDuckCobraHelper.SetCRSpec(bd.Spec)
	blackDuckInterface, err := updateBlackDuckCobraHelper.GenerateCRSpecFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	newSpec := blackDuckInterface.(blackduckapi.BlackduckSpec)
	// merge environs
	newSpec.Environs = util.MergeEnvSlices(newSpec.Environs, bd.Spec.Environs)
	bd.Spec = newSpec
	return bd, nil
}

// updateBlackDuckCmd updates a Black Duck instance
var updateBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsyctl update blackduck <name> --size medium\nsynopsyctl update blackduck <name> -n <namespace> --size medium\nsynopsyctl update blackduck <name> -n <namespace> --size medium --mock json",
	Short:         "Update a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		currBlackDuck, err = updateBlackDuck(currBlackDuck, cmd.Flags())
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			return PrintResource(*currBlackDuck, mockFormat, false)
		}

		log.Infof("updating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		// update the namespace label if the version of the app got changed
		_, err = util.CheckAndUpdateNamespace(kubeClient, util.BlackDuckName, blackDuckNamespace, blackDuckName, currBlackDuck.Spec.Version, false)
		if err != nil {
			return err
		}
		// Update Black Duck
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck)
		if err != nil {
			return fmt.Errorf("error updating Black Duck '%s' due to %+v", currBlackDuck.Name, err)
		}
		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckNativeCmd prints the Kubernetes resources with updates to a Black Duck instance
var updateBlackDuckNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsyctl update blackduck native <name> --size medium\nsynopsyctl update blackduck native <name> -n <namespace> --size medium\nsynopsyctl update blackduck native <name> --size medium -o yaml",
	Short:         "Print the Kubernetes resources with updates to a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		newBlackDuck, err := updateBlackDuck(currBlackDuck, cmd.Flags())
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		return PrintResource(*newBlackDuck, nativeFormat, true)
	},
}

// updateBlackDuckRootKeyCmd create new Black Duck root key for source code upload in the cluster
var updateBlackDuckRootKeyCmd = &cobra.Command{
	Use:           "masterkey NEW_SEAL_KEY STORED_MASTER_KEY_FILE_PATH",
	Example:       "synopsysctl update blackduck masterkey <new seal key> <file path of the stored master key>",
	Short:         "Update the master key to all Black Duck instances for source code upload functionality",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Help()
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, crdnamespace, crdScope, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}

		operatorNamespace, err := util.GetOperatorNamespaceByCRDScope(kubeClient, util.BlackDuckCRDName, crdScope, namespace)
		if err != nil {
			return fmt.Errorf("unable to find the Synopsys Operator instance due to %+v", err)
		}

		newSealKey := args[0]
		filePath := args[1]

		blackducks, err := util.ListBlackduck(blackDuckClient, crdnamespace, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances in namespace '%s' due to %+v", operatorNamespace, err)
		}

		secret, err := util.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			return fmt.Errorf("unable to find Synopsys Operator's blackduck-secret in namespace '%s' due to %+v", operatorNamespace, err)
		}

		for _, blackduck := range blackducks.Items {
			blackDuckName := blackduck.Name
			blackDuckNamespace := blackduck.Spec.Namespace
			log.Infof("updating Black Duck '%s's master key in namespace '%s'...", blackDuckName, blackDuckNamespace)

			fileName := filepath.Join(filePath, fmt.Sprintf("%s-%s.key", blackDuckNamespace, blackDuckName))
			masterKey, err := ioutil.ReadFile(fileName)
			if err != nil {
				return fmt.Errorf("error reading the master key from file '%s' due to %+v", fileName, err)
			}

			// Filter the upload cache pod to get the root key using the seal key
			uploadCachePod, err := util.FilterPodByNamePrefixInNamespace(kubeClient, blackDuckNamespace, util.GetResourceName(blackDuckName, util.BlackDuckName, "uploadcache"))
			if err != nil {
				return fmt.Errorf("unable to filter the upload cache pod in namespace '%s' due to %+v", blackDuckNamespace, err)
			}

			// Create the exec into Kubernetes pod request
			req := util.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
			// TODO: changed the upload cache service name to authentication until the HUB-20412 is fixed. once it if fixed, changed the name to use GetResource method
			_, err = util.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://uploadcache:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
			if err != nil {
				return fmt.Errorf("unable to exec into upload cache pod in namespace '%s' due to %+v", blackDuckNamespace, err)
			}

			log.Infof("successfully submitted updates to Black Duck '%s's master key in namespace '%s'", blackDuckName, operatorNamespace)
		}

		secret.Data["SEAL_KEY"] = []byte(newSealKey)

		_, err = util.UpdateSecret(kubeClient, operatorNamespace, secret)
		if err != nil {
			return fmt.Errorf("unable to update Synopsys Operator's blackduck-secret in namespace '%s' due to %+v", operatorNamespace, err)
		}

		return nil
	},
}

func updateBlackDuckAddEnviron(bd *blackduckapi.Blackduck, environ string) (*blackduckapi.Blackduck, error) {
	bd.Spec.Environs = util.MergeEnvSlices(strings.Split(environ, ","), bd.Spec.Environs)
	return bd, nil
}

// updateBlackDuckAddEnvironCmd adds an Environment Variable to a Black Duck instance
var updateBlackDuckAddEnvironCmd = &cobra.Command{
	Use:           "addenviron BLACK_DUCK_NAME (ENVIRON_NAME:ENVIRON_VALUE)",
	Example:       "synopsysctl update blackduck addenviron <name> USE_ALERT:1\nsynopsysctl update blackduck addenviron <name> USE_ALERT:1 -n <namespace>\nsynopsysctl update blackduck addenviron <name> USE_ALERT:1 --mock json",
	Short:         "Add an Environment Variable to a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Help()
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		newBlackDuck, err := updateBlackDuckAddEnviron(currBlackDuck, args[1])
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			return PrintResource(*newBlackDuck, mockFormat, false)
		}

		log.Infof("updating Black Duck '%s' with environ '%s' in namespace '%s'...", blackDuckName, args[1], blackDuckNamespace)
		_, err = util.UpdateBlackduck(blackDuckClient, newBlackDuck)
		if err != nil {
			return fmt.Errorf("error updating Black Duck '%s' due to %+v", newBlackDuck.Name, err)
		}
		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddEnvironCmd prints the Kubernetes resources with updates from adding an Environment Variable to a Black Duck instance
var updateBlackDuckAddEnvironNativeCmd = &cobra.Command{
	Use:           "native BLACK_DUCK_NAME (ENVIRON_NAME:ENVIRON_VALUE)",
	Example:       "synopsysctl update blackduck addenviron native <name> USE_ALERT:1\nsynopsysctl update blackduck addenviron native <name> USE_ALERT:1 -n <namespace>\nsynopsysctl update blackduck addenviron native <name> USE_ALERT:1 -o yaml",
	Short:         "Print the Kubernetes resources with updates from adding an Environment Variable to a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Help()
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		newBlackDuck, err := updateBlackDuckAddEnviron(currBlackDuck, args[1])
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		return PrintResource(*newBlackDuck, nativeFormat, true)
	},
}

func updateBlackDuckSetImageRegistry(bd *blackduckapi.Blackduck, imageRegistry string) (*blackduckapi.Blackduck, error) {
	// Get the name of the container
	baseContainerName, err := util.GetImageName(imageRegistry)
	if err != nil {
		return nil, err
	}
	// Add Registry to Spec
	var found bool
	for i, imageReg := range bd.Spec.ImageRegistries {
		existingBaseContainerName, err := util.GetImageName(imageReg)
		if err != nil {
			return nil, err
		}
		found = strings.EqualFold(existingBaseContainerName, baseContainerName)
		if found {
			bd.Spec.ImageRegistries[i] = imageRegistry // replace existing imageReg
			break
		}
	}
	if !found { // if didn't already exist, add new imageReg
		bd.Spec.ImageRegistries = append(bd.Spec.ImageRegistries, imageRegistry)
	}
	return bd, nil
}

// updateBlackDuckSetImageRegistryCmd adds an image to a Black Duck instance
var updateBlackDuckSetImageRegistryCmd = &cobra.Command{
	Use:           "setimage BLACK_DUCK_NAME (REGISTRY/IMAGE:TAG)",
	Example:       "synopsysctl update blackduck setimage <name> docker.io/blackducksoftware/blackduck-cfssl:2019.6.0\nsynopsysctl update blackduck setimage <name> docker.io/blackducksoftware/blackduck-cfssl:2019.6.0 -n <namespace>",
	Short:         "Set the registry location for an image used by a specific Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Help()
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		newBlackDuck, err := updateBlackDuckSetImageRegistry(currBlackDuck, args[1])
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			return PrintResource(*newBlackDuck, mockFormat, false)
		}

		log.Infof("updating Black Duck '%s' with image registry in namespace '%s'...", blackDuckName, blackDuckNamespace)
		_, err = util.UpdateBlackduck(blackDuckClient, newBlackDuck)
		if err != nil {
			return fmt.Errorf("error updating Black Duck '%s' due to %+v", newBlackDuck.Name, err)
		}
		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckSetImageRegistryNativeCmd prints the Kubernetes resources with updates from adding an Image Registry to a Black Duck instance
var updateBlackDuckSetImageRegistryNativeCmd = &cobra.Command{
	Use:           "addregistry BLACK_DUCK_NAME REGISTRY",
	Example:       "synopsysctl update blackduck addregistry native <name> docker.io\nsynopsysctl update blackduck addregistry native <name> docker.io -n <namespace>\nsynopsysctl update blackduck addregistry native <name> docker.io -o yaml",
	Short:         "Print the Kubernetes resources with updates from adding an Image Registry to a Black Duck instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmd.Help()
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, args[0])
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		newBlackDuck, err := updateBlackDuckSetImageRegistry(currBlackDuck, args[1])
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
		return PrintResource(*newBlackDuck, nativeFormat, true)
	},
}

/*
Update OpsSight Commands
*/

func updateOpsSight(ops *opssightapi.OpsSight, flagset *pflag.FlagSet) (*opssightapi.OpsSight, error) {
	updateOpsSightCobraHelper.SetCRSpec(ops.Spec)
	opsSightInterface, err := updateOpsSightCobraHelper.GenerateCRSpecFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	newSpec := opsSightInterface.(opssightapi.OpsSightSpec)
	ops.Spec = newSpec
	return ops, nil
}

// updateOpsSightCmd updates an OpsSight instance
var updateOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME",
	Example:       "synopsyctl update opssight <name> --blackduck-max-count 2\nsynopsyctl update opssight <name> --blackduck-max-count 2 -n <namespace>\nsynopsyctl update opssight <name> --blackduck-max-count 2 --mock json",
	Short:         "Update an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSight(currOpsSight, cmd.Flags())
		if err != nil {
			return err
		}

		// update the namespace label if the version of the app got changed
		// TODO: when opssight versioning PR is merged, the hard coded 2.2.4 version to be replaced with OpsSight
		_, err = util.CheckAndUpdateNamespace(kubeClient, util.OpsSightName, opsSightNamespace, opsSightName, "2.2.4", false)
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*newOpsSight, mockFormat, false)
		}

		log.Infof("updating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		_, err = util.UpdateOpsSight(opsSightClient, crdnamespace, newOpsSight)
		if err != nil {
			return fmt.Errorf("error updating OpsSight '%s' due to %+v", newOpsSight.Name, err)
		}
		log.Infof("successfully submitted updates to OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// updateOpsSightNativeCmd prints the Kubernetes resources with updates to an OpsSight instance
var updateOpsSightNativeCmd = &cobra.Command{
	Use:           "native NAME",
	Example:       "synopsyctl update opssight native <name> --blackduck-max-count 2\nsynopsyctl update opssight native <name> --blackduck-max-count 2 -n <namespace>\nsynopsyctl update opssight native <name> --blackduck-max-count 2 -o yaml",
	Short:         "Print the Kubernetes resources with updates to an OpsSight instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSight(currOpsSight, cmd.Flags())
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*newOpsSight, nativeFormat, true)
	},
}

func updateOpsSightExternalHost(ops *opssightapi.OpsSight, scheme, domain, port, user, pass, scanLimit string) (*opssightapi.OpsSight, error) {
	hostPort, err := strconv.ParseInt(port, 0, 64)
	if err != nil {
		return nil, err
	}
	hostScanLimit, err := strconv.ParseInt(scanLimit, 0, 64)
	if err != nil {
		return nil, err
	}
	newHost := opssightapi.Host{
		Scheme:              scheme,
		Domain:              domain,
		Port:                int(hostPort),
		User:                user,
		Password:            pass,
		ConcurrentScanLimit: int(hostScanLimit),
	}
	ops.Spec.Blackduck.ExternalHosts = append(ops.Spec.Blackduck.ExternalHosts, &newHost)
	return ops, nil
}

// updateOpsSightExternalHostCmd updates an external host for an OpsSight intance's component
var updateOpsSightExternalHostCmd = &cobra.Command{
	Use:           "externalhost NAME SCHEME DOMAIN PORT USER PASSWORD SCANLIMIT",
	Example:       "synopsysctl update opssight externalhost <name> scheme domain 80 user pass 50\nsynopsysctl update opssight externalhost <name> scheme domain 80 user pass 50 -n <namespace>",
	Short:         "Update an external host for an OpsSight intance's component",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 7 {
			cmd.Help()
			return fmt.Errorf("this command takes 7 arguments")
		}
		// Check Host Port
		_, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid port number: '%s'", err)
		}
		// Check Host Scan Limit
		_, err = strconv.ParseInt(args[6], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid concurrent scan limit: %s", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSightExternalHost(currOpsSight, args[1], args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*newOpsSight, mockFormat, false)
		}

		log.Infof("updating OpsSight '%s' with an external host in namespace '%s'...", opsSightName, opsSightNamespace)
		_, err = util.UpdateOpsSight(opsSightClient, crdnamespace, newOpsSight)
		if err != nil {
			return fmt.Errorf("error updating OpsSight '%s' due to %+v", newOpsSight.Name, err)
		}
		log.Infof("successfully submitted updates to OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// updateOpsSightExternalHostNativeCmd prints the Kubernetes resources with updates to an external host for an OpsSight intance's component
var updateOpsSightExternalHostNativeCmd = &cobra.Command{
	Use:           "externalhost NAME SCHEME DOMAIN PORT USER PASSWORD SCANLIMIT",
	Example:       "synopsysctl update opssight externalhost native <name> scheme domain 80 user pass 50\nsynopsysctl update opssight externalhost native <name> scheme domain 80 user pass 50 -n <namespace>",
	Short:         "Print the Kubernetes resources with updates to an external host for an OpsSight intance's component",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 7 {
			cmd.Help()
			return fmt.Errorf("this command takes 7 arguments")
		}
		// Check Host Port
		_, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid port number: '%s'", err)
		}
		// Check Host Scan Limit
		_, err = strconv.ParseInt(args[6], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid concurrent scan limit: %s", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSightExternalHost(currOpsSight, args[1], args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*newOpsSight, nativeFormat, true)
	},
}

func updateOpsSightAddRegistry(ops *opssightapi.OpsSight, url, user, pass string) (*opssightapi.OpsSight, error) {
	newReg := opssightapi.RegistryAuth{
		URL:      url,
		User:     user,
		Password: pass,
	}
	ops.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ops.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
	return ops, nil
}

// updateOpsSightAddRegistryCmd adds an internal registry to an OpsSight instance's ImageFacade
var updateOpsSightAddRegistryCmd = &cobra.Command{
	Use:           "registry NAME URL USER PASSWORD",
	Example:       "synopsysctl update opssight registry <name> reg_url reg_username reg_password\nsynopsysctl update opssight registry <name> reg_url reg_username reg_password -n <namespace>",
	Short:         "Add an internal registry to an OpsSight instance's ImageFacade",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			cmd.Help()
			return fmt.Errorf("this command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mockMode := cmd.Flags().Lookup("mock").Changed
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSightAddRegistry(currOpsSight, args[1], args[2], args[3])
		if err != nil {
			return err
		}

		// If mock mode, return and don't create resources
		if mockMode {
			log.Debugf("generating updates to the CRD for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			return PrintResource(*newOpsSight, mockFormat, false)
		}

		log.Infof("updating OpsSight '%s' with internal registry in namespace '%s'...", opsSightName, opsSightNamespace)
		_, err = util.UpdateOpsSight(opsSightClient, crdnamespace, newOpsSight)
		if err != nil {
			return fmt.Errorf("error updating OpsSight '%s' due to %+v", newOpsSight.Name, err)
		}
		log.Infof("successfully submitted updates to OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		return nil
	},
}

// updateOpsSightAddRegistryNativeCmd prints the Kubernetes resources with updates from adding an internal registry to an OpsSight instance's ImageFacade
var updateOpsSightAddRegistryNativeCmd = &cobra.Command{
	Use:           "native NAME URL USER PASSWORD",
	Example:       "synopsysctl update opssight registry native <name> reg_url reg_username reg_password\nsynopsysctl update opssight registry native <name> reg_url reg_username reg_password -n <namespace>",
	Short:         "Print the Kubernetes resources with updates from adding an internal registry to an OpsSight instance's ImageFacade",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			cmd.Help()
			return fmt.Errorf("this command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName, opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, args[0])
		if err != nil {
			return err
		}
		currOpsSight, err := util.GetOpsSight(opsSightClient, crdnamespace, opsSightName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
		}
		newOpsSight, err := updateOpsSightAddRegistry(currOpsSight, args[1], args[2], args[3])
		if err != nil {
			return err
		}

		log.Debugf("generating updates to the Kubernetes resources for OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
		return PrintResource(*newOpsSight, nativeFormat, true)
	},
}

func addUpdateOperatorFlags(cmd *cobra.Command) {
	// Add Operator Commands
	cmd.Flags().BoolVarP(&isEnabledAlert, "enable-alert", "a", isEnabledAlert, "Enable/Disable Alert Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().BoolVarP(&isEnabledBlackDuck, "enable-blackduck", "b", isEnabledBlackDuck, "Enable/Disable Black Duck Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().BoolVarP(&isEnabledOpsSight, "enable-opssight", "s", isEnabledOpsSight, "Enable/Disable OpsSight Custom Resource Definition (CRD) in your cluster")
	// cmd.Flags().BoolVarP(&isEnabledPrm, "enable-prm", "p", isEnabledPrm, "Enable/Disable Polaris Reporting Module Custom Resource Definition (CRD) in your cluster")
	cmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "Image URL of Synopsys Operator")
	cmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "x", exposeMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().StringVarP(&metricsImage, "metrics-image", "m", metricsImage, "Image URL of Synopsys Operator's metrics pod")
	cmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "q", postgresRestartInMins, "Minutes to check for restarting postgres")
	cmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	cmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	cmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	cmd.Flags().StringVarP(&dryRun, "dry-run", "d", dryRun, "If true, Synopsys Operator runs without being connected to a cluster [true|false]")
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	cmd.Flags().IntVarP(&threadiness, "no-of-threads", "t", threadiness, "Number of threads to process the custom resources")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the Synopsys Operator instance")
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackDuckCobraHelper = blackduck.NewCRSpecBuilderFromCobraFlags()
	updateOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	updateAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()

	rootCmd.AddCommand(updateCmd)

	/* Update Operator Comamnds */

	// updateOperatorCmd
	addUpdateOperatorFlags(updateOperatorCmd)
	addMockFlag(updateOperatorCmd)
	updateCmd.AddCommand(updateOperatorCmd)

	addUpdateOperatorFlags(updateOperatorNativeCmd)
	addNativeFormatFlag(updateOperatorNativeCmd)
	updateOperatorCmd.AddCommand(updateOperatorNativeCmd)

	/* Update Alert Comamnds */

	// updateAlertCmd
	updateAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updateAlertCobraHelper.AddCRSpecFlagsToCommand(updateAlertCmd, false)
	addMockFlag(updateAlertCmd)
	updateCmd.AddCommand(updateAlertCmd)

	updateAlertCobraHelper.AddCRSpecFlagsToCommand(updateAlertNativeCmd, false)
	addNativeFormatFlag(updateAlertNativeCmd)
	updateAlertCmd.AddCommand(updateAlertNativeCmd)

	/* Update Black Duck Comamnds */

	// updateBlackDuckCmd
	updateBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updateBlackDuckCobraHelper.AddCRSpecFlagsToCommand(updateBlackDuckCmd, false)
	addMockFlag(updateBlackDuckCmd)
	updateCmd.AddCommand(updateBlackDuckCmd)

	updateBlackDuckCobraHelper.AddCRSpecFlagsToCommand(updateBlackDuckNativeCmd, false)
	addNativeFormatFlag(updateBlackDuckNativeCmd)
	updateBlackDuckCmd.AddCommand(updateBlackDuckNativeCmd)

	// updateBlackDuckRootKeyCmd
	updateBlackDuckCmd.AddCommand(updateBlackDuckRootKeyCmd)

	// updateBlackDuckAddEnvironCmd
	addMockFlag(updateBlackDuckAddEnvironCmd)
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddEnvironCmd)

	// updateBlackDuckSetImageRegistryCmd
	addMockFlag(updateBlackDuckSetImageRegistryCmd)
	updateBlackDuckCmd.AddCommand(updateBlackDuckSetImageRegistryCmd)

	addNativeFormatFlag(updateBlackDuckSetImageRegistryNativeCmd)
	updateBlackDuckSetImageRegistryCmd.AddCommand(updateBlackDuckSetImageRegistryNativeCmd)

	/* Update OpsSight Comamnds */

	// updateOpsSightCmd
	updateOpsSightCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updateOpsSightCobraHelper.AddCRSpecFlagsToCommand(updateOpsSightCmd, false)
	addMockFlag(updateOpsSightCmd)
	updateCmd.AddCommand(updateOpsSightCmd)

	updateOpsSightCobraHelper.AddCRSpecFlagsToCommand(updateOpsSightNativeCmd, false)
	addNativeFormatFlag(updateOpsSightNativeCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightNativeCmd)

	// updateOpsSightExternalHostCmd
	addMockFlag(updateOpsSightExternalHostCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightExternalHostCmd)

	addNativeFormatFlag(updateOpsSightExternalHostNativeCmd)
	updateOpsSightExternalHostCmd.AddCommand(updateOpsSightExternalHostNativeCmd)

	// updateOpsSightAddRegistryCmd
	addMockFlag(updateOpsSightAddRegistryCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightAddRegistryCmd)

	addNativeFormatFlag(updateOpsSightAddRegistryNativeCmd)
	updateOpsSightAddRegistryCmd.AddCommand(updateOpsSightAddRegistryNativeCmd)
}
