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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	polarisreporting "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"
	"github.com/pkg/errors"

	// bdappsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	appsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update Command ResourceCtlSpecBuilders
var updateAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updateBlackDuckCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updateOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updatePolarisCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updatePolarisReportingCobraHelper polarisreporting.HelmValuesFromCobraFlags

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
	// check whether the update Alert version is greater than or equal to 5.0.0
	isGreaterThanOrEqualTo, err := util.IsNotDefaultVersionGreaterThanOrEqualTo(newSpec.Version, 5, 0, 0)
	if err != nil {
		return nil, err
	}

	// if greater than or equal to 5.0.0, then copy PUBLIC_HUB_WEBSERVER_HOST to ALERT_HOSTNAME and PUBLIC_HUB_WEBSERVER_PORT to ALERT_SERVER_PORT
	// and delete PUBLIC_HUB_WEBSERVER_HOST and PUBLIC_HUB_WEBSERVER_PORT from the environs. In future, we need to request the customer to use the new params
	if isGreaterThanOrEqualTo {
		isChanged := false
		maps := util.StringArrayToMapSplitBySeparator(newSpec.Environs, ":")
		if _, ok := maps["PUBLIC_HUB_WEBSERVER_HOST"]; ok {
			if _, ok1 := maps["ALERT_HOSTNAME"]; !ok1 {
				maps["ALERT_HOSTNAME"] = maps["PUBLIC_HUB_WEBSERVER_HOST"]
				isChanged = true
			}
			delete(maps, "PUBLIC_HUB_WEBSERVER_HOST")
		}

		if _, ok := maps["PUBLIC_HUB_WEBSERVER_PORT"]; ok {
			if _, ok1 := maps["ALERT_SERVER_PORT"]; !ok1 {
				maps["ALERT_SERVER_PORT"] = maps["PUBLIC_HUB_WEBSERVER_PORT"]
				isChanged = true
			}
			delete(maps, "PUBLIC_HUB_WEBSERVER_PORT")
		}

		if isChanged {
			newSpec.Environs = util.MapToStringArrayJoinBySeparator(maps, ":")
		}
	}
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
		alertName := args[0]
		alertNamespace, crdnamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, alertName)
		if err != nil {
			return err
		}
		currAlert, err := util.GetAlert(alertClient, crdnamespace, alertName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
		}
		currAlert, err = updateAlert(currAlert, cmd.Flags())
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
		_, err = util.CheckAndUpdateNamespace(kubeClient, util.AlertName, alertNamespace, alertName, currAlert.Spec.Version, false)
		if err != nil {
			return err
		}
		// Update the Alert
		_, err = util.UpdateAlert(alertClient, crdnamespace, currAlert)
		if err != nil {
			return fmt.Errorf("error updating Alert '%s' due to %+v", currAlert.Name, err)
		}
		log.Infof("successfully submitted updates to Alert '%s' in namespace '%s'", alertName, alertNamespace)
		return nil
	},
}

/*
Update Black Duck Commands
*/

func updateBlackDuckSpec(bd *blackduckapi.Blackduck, flagset *pflag.FlagSet) (*blackduckapi.Blackduck, error) {
	updateBlackDuckCobraHelper.SetCRSpec(bd.Spec)
	blackDuckInterface, err := updateBlackDuckCobraHelper.GenerateCRSpecFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	newSpec := blackDuckInterface.(blackduckapi.BlackduckSpec)
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
		blackDuckName := args[0]
		blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
		}
		oldBlackDuck := *currBlackDuck
		currBlackDuck, err = updateBlackDuckSpec(currBlackDuck, cmd.Flags())
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

		// Update the File Owernship in Persistent Volumes if Security Context changes are needed
		oldVersion := oldBlackDuck.Spec.Version
		oldState := oldBlackDuck.Spec.DesiredState
		oldVersionIsGreaterThanOrEqualv2019x12x0, err := util.IsVersionGreaterThanOrEqualTo(oldVersion, 2019, time.December, 0)
		if err != nil {
			return err
		}
		newVersionIsGreaterThanOrEqualv2019x12x0, err := util.IsVersionGreaterThanOrEqualTo(currBlackDuck.Spec.Version, 2019, time.December, 0)
		if err != nil {
			return err
		}
		if !newVersionIsGreaterThanOrEqualv2019x12x0 && cmd.Flags().Changed("security-context-file-path") {
			return fmt.Errorf("security contexts from --security-context-file-path cannot be set for versions before 2019.12.0, you're using version %s", currBlackDuck.Spec.Version)
		}
		if util.IsOpenshift(kubeClient) && cmd.Flags().Changed("security-context-file-path") {
			return fmt.Errorf("cannot set security contexts with --security-context-file-path in an Openshift environment")
		}
		bdUpdatedToHaveSecurityContexts := cmd.Flags().Lookup("version").Changed && (!oldVersionIsGreaterThanOrEqualv2019x12x0 && newVersionIsGreaterThanOrEqualv2019x12x0) // case: Security Contexts are set in an old version and then upgrade to a version that requires changes
		bdUpdatedToHaveSecurityContextsAndNoPersistentStorage := bdUpdatedToHaveSecurityContexts && !currBlackDuck.Spec.PersistentStorage                                   // case: Black Duck will be restarted during update and no changes to PVs are needed
		bdSecurityContextsWereChanged := cmd.Flags().Lookup("security-context-file-path").Changed && newVersionIsGreaterThanOrEqualv2019x12x0                               // case: Security Contexts are set and the version requires changes
		if (bdUpdatedToHaveSecurityContexts || bdSecurityContextsWereChanged) && !bdUpdatedToHaveSecurityContextsAndNoPersistentStorage && !util.IsOpenshift(kubeClient) {
			log.Infof("stopping Black Duck to apply Security Context changes")
			if currBlackDuck.Spec.DesiredState != "STOP" {
				stoppedBlackDuck := oldBlackDuck
				stoppedBlackDuck.Spec.DesiredState = "STOP"
				_, err = util.UpdateBlackduck(blackDuckClient, &stoppedBlackDuck)
				if err != nil {
					return errors.Wrap(err, "failed to get Black Duck while setting file owernship")
				}
				log.Infof("waiting for Black Duck to stop...")
				waitCount := 0
				for {
					// ... wait for the Black Duck to stop
					pods, err := util.ListPodsWithLabels(kubeClient, blackDuckNamespace, fmt.Sprintf("app=blackduck,name=%s", blackDuckName))
					if err != nil {
						return errors.Wrap(err, "failed to list pods to stop BlackDuck for setting group ownership")
					}
					if len(pods.Items) == 0 {
						break
					}
					time.Sleep(time.Second * 5)
					waitCount = waitCount + 1
					if waitCount%5 == 0 {
						log.Debugf("waiting for Black Duck to stop - %d pods remaining", len(pods.Items))
					}
				}
			}
			// Get a list of Persistent Volumes based on Persistent Volume Claims
			pvcList, err := util.ListPVCs(kubeClient, blackDuckNamespace, fmt.Sprintf("app=blackduck,component=pvc,name=%s", blackDuckName))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to list PVCs to update the group ownership"))
			}
			// Map the Persistent Volume to the respective Security Context file Ownership value - if security contexts are not provided then this map will be empty
			pvcNameToFileOwnershipMap := map[string]int64{}
			pvcNameToSecurityContextNameMap := map[string]string{
				"blackduck-postgres":         "blackduck-postgres",
				"blackduck-cfssl":            "blackduck-cfssl",
				"blackduck-registration":     "blackduck-registration",
				"blackduck-zookeeper":        "blackduck-zookeeper",
				"blackduck-authentication":   "blackduck-authentication",
				"blackduck-webapp":           "blackduck-webapp",
				"blackduck-logstash":         "blackduck-webapp",
				"blackduck-uploadcache-data": "blackduck-uploadcache",
			}
			for _, pvc := range pvcList.Items {
				r, _ := regexp.Compile("blackduck-.*")
				pvcNameKey := r.FindString(pvc.Name) // removes the "<blackduckName>-" from the PvcName
				sc := appsutil.GetSecurityContext(currBlackDuck.Spec.SecurityContexts, pvcNameToSecurityContextNameMap[pvcNameKey])
				if sc != nil {
					if sc.RunAsUser != nil {
						pvcNameToFileOwnershipMap[pvc.Name] = *sc.RunAsUser
					}
				}
			}
			// If security contexts were provided, update the Persistent Volumes that have a file Ownership value set
			if len(pvcNameToFileOwnershipMap) > 0 {
				log.Infof("updating file ownership in Persistent Volumes...")
				// Create Jobs to set the file owernship in each Persistent Volume
				log.Infof("creating jobs to set the file owernship in each Persistent Volume")
				var wg sync.WaitGroup
				wg.Add(len(pvcNameToFileOwnershipMap))
				for pvcName, ownership := range pvcNameToFileOwnershipMap {
					log.Infof("creating file owernship job to set ownership value to '%d' in PV '%s'", ownership, pvcName)
					go setBlackDuckFileOwnershipJob(blackDuckNamespace, blackDuckName, pvcName, ownership, &wg)
				}
				log.Infof("waiting for file owernship jobs to finish...")
				wg.Wait()
				if len(pvcNameToFileOwnershipMap) != len(pvcList.Items) {
					log.Warnf("a Job was not created for each Persistent Volume")
				}
			}
			// Get new BlackDuck (after update to STOP), set Desired State to the original value, and reapply the user's updates
			log.Debugf("restarting Black Duck...")
			currBlackDuck, err = util.GetBlackduck(blackDuckClient, crdnamespace, blackDuckName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
			}
			currBlackDuck.Spec.DesiredState = oldState
			currBlackDuck, err = updateBlackDuckSpec(currBlackDuck, cmd.Flags())
			if err != nil {
				return err
			}
		}
		// Update Black Duck with User's Changes
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error updating Black Duck '%s'", currBlackDuck.Name))
		}
		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// setBlackDuckFileOwnershipJob that sets the Owner of the files
func setBlackDuckFileOwnershipJob(namespace string, name string, pvcName string, ownership int64, wg *sync.WaitGroup) error {
	volumeClaim := components.NewPVCVolume(horizonapi.PVCVolumeConfig{PVCName: pvcName})
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("set-file-ownership-%s", pvcName),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "set-file-ownership-container",
							Image:   busyBoxImage,
							Command: []string{"chown", "-R", fmt.Sprintf("%d", ownership), "/setfileownership"},
							VolumeMounts: []corev1.VolumeMount{
								{Name: pvcName, MountPath: "/setfileownership"},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{Name: pvcName, VolumeSource: volumeClaim.VolumeSource},
					},
				},
			},
		},
	}
	defer wg.Done()

	job, err := kubeClient.BatchV1().Jobs(namespace).Create(job)
	if err != nil {
		panic(fmt.Sprintf("failed to create job for setting group ownership due to %s", err))
	}

	timeout := time.NewTimer(30 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("failed to set the group ownership of files for PV '%s' in namespace '%s'", pvcName, namespace)

		case <-ticker.C:
			job, err = kubeClient.BatchV1().Jobs(job.Namespace).Get(job.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if job.Status.Succeeded > 0 {
				log.Infof("successfully set the group ownership of files for PV '%s' in namespace '%s'", pvcName, namespace)
				kubeClient.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{})
				return nil
			}
		}
	}
}

// updateBlackDuckMasterKeyCmd create new Black Duck master key for source code upload in the cluster
var updateBlackDuckMasterKeyCmd = &cobra.Command{
	Use:           "masterkey BLACK_DUCK_NAME DIRECTORY_PATH_OF_STORED_MASTER_KEY NEW_SEAL_KEY -n NAMESPACE",
	Example:       "synopsysctl update blackduck masterkey <Black Duck name> <directory path of the stored master key> <new seal key> -n <namespace>",
	Short:         "Update the master key of the Black Duck instance that is used for source code upload",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			cmd.Help()
			return fmt.Errorf("this command takes 3 arguments")
		}

		if len(args[2]) != 32 {
			return fmt.Errorf("new seal key should be of length 32")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := updateMasterKey(namespace, args[0], args[1], args[2], false); err != nil {
			return err
		}
		return nil
	},
}

// updateBlackDuckMasterKeyNativeCmd create new Black Duck master key for source code upload in the cluster
var updateBlackDuckMasterKeyNativeCmd = &cobra.Command{
	Use:           "native BLACK_DUCK_NAME DIRECTORY_PATH_OF_STORED_MASTER_KEY NEW_SEAL_KEY -n NAMESPACE",
	Example:       "synopsysctl update blackduck masterkey native <Black Duck name> <directory path of the stored master key> <new seal key> -n <namespace>",
	Short:         "Update the master key of the Black Duck instance that is used for source code upload",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			cmd.Help()
			return fmt.Errorf("this command takes 3 arguments")
		}

		if len(args[2]) != 32 {
			return fmt.Errorf("new seal key should be of length 32")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := updateMasterKey(namespace, args[0], args[1], args[2], true); err != nil {
			return err
		}
		return nil
	},
}

// updateMasterKey updates the master key and encoded with new seal key
func updateMasterKey(namespace string, name string, oldMasterKeyFilePath string, newSealKey string, isNative bool) error {

	// getting the seal key secret to retrieve the seal key
	secret, err := util.GetSecret(kubeClient, namespace, fmt.Sprintf("%s-blackduck-upload-cache", name))
	if err != nil {
		return fmt.Errorf("unable to find Seal key secret (%s-blackduck-upload-cache) in namespace '%s' due to %+v", name, namespace, err)
	}

	// retrieve the Black Duck configmap
	cm, err := util.GetConfigMap(kubeClient, namespace, fmt.Sprintf("%s-blackduck-config", name))
	if err != nil {
		return fmt.Errorf("unable to find Black Duck config map (%s-blackduck-config) in namespace '%s' due to %+v", name, namespace, err)
	}

	log.Infof("updating Black Duck '%s's master key in namespace '%s'...", name, namespace)

	// read the old master key
	fileName := filepath.Join(oldMasterKeyFilePath, fmt.Sprintf("%s-%s.key", namespace, name))
	masterKey, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("error reading the master key from file '%s' due to %+v", fileName, err)
	}

	// Filter the upload cache pod to get the root key using the seal key
	uploadCachePod, err := util.FilterPodByNamePrefixInNamespace(kubeClient, namespace, util.GetResourceName(name, util.BlackDuckName, "uploadcache"))
	if err != nil {
		return fmt.Errorf("unable to filter the upload cache pod in namespace '%s' due to %+v", namespace, err)
	}

	// Create the exec into Kubernetes pod request
	req := util.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
	uploadCache := "uploadcache"
	if isVersionGreaterThanorEqualTo, err := util.IsVersionGreaterThanOrEqualTo(cm.Data["HUB_VERSION"], 2019, time.August, 0); err == nil && isVersionGreaterThanorEqualTo {
		uploadCache = util.GetResourceName(name, util.BlackDuckName, "uploadcache")
	}

	_, err = util.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://%s:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey, uploadCache)})
	if err != nil {
		return fmt.Errorf("unable to exec into upload cache pod in namespace '%s' due to %+v", namespace, err)
	}

	log.Infof("successfully updated the master key in the upload cache container of Black Duck '%s' in namespace '%s'", name, namespace)

	if isNative {
		// update the new seal key
		secret.Data["SEAL_KEY"] = []byte(newSealKey)
		_, err = util.UpdateSecret(kubeClient, namespace, secret)
		if err != nil {
			return fmt.Errorf("unable to update Seal key secret (%s-blackduck-upload-cache) in namespace '%s' due to %+v", name, namespace, err)
		}

		log.Infof("successfully updated the seal key secret for Black Duck '%s' in namespace '%s'", name, namespace)

		// delete the upload cache pod
		err = util.DeletePod(kubeClient, namespace, uploadCachePod.Name)
		if err != nil {
			return fmt.Errorf("unable to delete an upload cache pod in namespace '%s' due to %+v", namespace, err)
		}

		log.Infof("successfully deleted an upload cache pod for Black Duck '%s' in namespace '%s' to reflect the new seal key. Wait for upload cache pod to restart to resume the source code upload", name, namespace)
	} else {
		_, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, name)
		if err != nil {
			return err
		}
		currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdnamespace, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", name, namespace, err)
		}
		currBlackDuck.Spec.SealKey = util.Base64Encode([]byte(newSealKey))
		_, err = util.UpdateBlackduck(blackDuckClient, currBlackDuck)
		if err != nil {
			return fmt.Errorf("error updating Black Duck '%s' in namespace '%s' due to %+v", name, namespace, err)
		}
		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'. Wait for upload cache pod to restart to resume the source code upload", name, namespace)
	}
	return nil
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
		blackDuckName := args[0]
		blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
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
		blackDuckName := args[0]
		blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
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
		blackDuckName := args[0]
		blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
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
		blackDuckName := args[0]
		blackDuckNamespace, crdnamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
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
		opsSightName := args[0]
		opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
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
		// TODO: when opssight versioning PR is merged, the hard coded 2.2.5 version to be replaced with OpsSight
		_, err = util.CheckAndUpdateNamespace(kubeClient, util.OpsSightName, opsSightNamespace, opsSightName, "2.2.5", false)
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
		opsSightName := args[0]
		opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
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
		opsSightName := args[0]
		opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
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
		opsSightName := args[0]
		opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
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
		opsSightName := args[0]
		opsSightNamespace, crdnamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
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
	cmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "Image URL of Synopsys Operator")
	cmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "q", postgresRestartInMins, "Minutes to check for restarting postgres")
	cmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	cmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	cmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	cmd.Flags().StringVarP(&dryRun, "dry-run", "d", dryRun, "If true, Synopsys Operator runs without being connected to a cluster [true|false]")
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	cmd.Flags().IntVarP(&threadiness, "no-of-threads", "t", threadiness, "Number of threads to process the custom resources")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the Synopsys Operator instance")
}

func updatePolaris(polarisObj polaris.Polaris, flagset *pflag.FlagSet) (*polaris.Polaris, error) {
	if err := updatePolarisCobraHelper.SetCRSpec(polarisObj); err != nil {
		return nil, err
	}

	spec, err := updatePolarisCobraHelper.GenerateCRSpecFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	newSpec := spec.(polaris.Polaris)

	// Unmarshal the platform license and set the organization name according to the license
	var plaformLicense *polaris.PlatformLicense
	if err := json.Unmarshal([]byte(newSpec.Licenses.Polaris), &plaformLicense); err != nil {
		return nil, fmt.Errorf("the Polaris license is in an invalid format and has to have a IssuedTo field: %s", newSpec.Licenses.Polaris)
	}

	if strings.Compare(newSpec.OrganizationDetails.OrganizationProvisionOrganizationName, plaformLicense.License.IssuedTo) != 0 {
		return nil, fmt.Errorf("the Polaris license provided is not valid for organizationd: %s", newSpec.OrganizationDetails.OrganizationProvisionOrganizationName)
	}

	if err := validatePolaris(newSpec); err != nil {
		return nil, err
	}
	return &newSpec, nil
}

// updatePolarisCmd updates a Polaris instance
var updatePolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsyctl update polaris -n <namespace>",
	Short:         "Update a Polaris instance \n\nSee detailed documentation about updating Polaris: [https://synopsys.atlassian.net/wiki/spaces/POP/overview]",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("this command takes 0 arguments")
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		polarisObj, err := getPolarisFromSecret()
		if err != nil {
			return err
		}
		reportingIsAlreadyEnabled := polarisObj.EnableReporting
		if cmd.Flags().Changed("reportstorage-size") && reportingIsAlreadyEnabled {
			return fmt.Errorf("reportstorage-size cannot be changed, you can only set it if you are enabling reporting for the first time")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		polarisObj, err := getPolarisFromSecret()
		if err != nil {
			return err
		}

		if polarisObj == nil {
			return fmt.Errorf("either namespace does not exist or secret does not exist because this instance of polaris was not created via synopsysctl")
		}

		newPolaris, err := updatePolaris(*polarisObj, cmd.Flags())
		if err != nil {
			return err
		}

		if err := CheckVersionExists(baseURL, newPolaris.Version); err != nil {
			return err
		}
		if err := ensurePolaris(newPolaris, true); err != nil {
			return err
		}
		return nil
	},
}

// updatePolarisReportingCmd updates a Polaris-Reporting instance
var updatePolarisReportingCmd = &cobra.Command{
	Use:           "polaris-reporting",
	Example:       "",
	Short:         "Update a Polaris-Reporting instance",
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
		helmValuesMap, err := updatePolarisReportingCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
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
				polarisReportingChartRepository = fmt.Sprintf("https://chartmuseum.polaris-cc-staging.sig-clops.synopsys.com/charts/polaris-helmchart-reporting-%s.tgz", versionFlag.Value.String())
			}
		}

		// Get Secret For the GCP Key
		gcpServiceAccountPath := cmd.Flag("gcp-service-account-path").Value.String()
		gcpServiceAccountData, err := util.ReadFileData(gcpServiceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read gcp service account file at location: '%s', error: %+v", gcpServiceAccountPath, err)
		}
		gcpServiceAccountSecrets, err := polarisreporting.GetPolarisReportingSecrets(namespace, gcpServiceAccountData)
		if err != nil {
			return fmt.Errorf("failed to generate GCP Service Account Secrets: %+v", err)
		}

		// Deploy the Secret
		err = KubectlApplyRuntimeObjects(gcpServiceAccountSecrets)
		if err != nil {
			return fmt.Errorf("failed to update the gcpServiceAccount Secrets: %s", err)
		}

		// Deploy Polaris-Reporting Resources
		// out, err := util.RunHelm3("upgrade", []string{polarisReportingName, polarisReportingChartRepository, "-n", namespace, "--reuse-values"}, helmValuesMap)
		// if err != nil {
		// 	return fmt.Errorf("failed to update Polaris-Reporting resources: %+v", out)
		// }
		_, _ = util.RunHelm3("install", polarisReportingName, namespace, polarisReportingChartRepository, []string{}, helmValuesMap) // remove this line

		log.Infof("Polaris-Reporting has been successfully Updated in namespace '%s'!", namespace)
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackDuckCobraHelper = blackduck.NewCRSpecBuilderFromCobraFlags()
	updateOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	updateAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()
	updatePolarisCobraHelper = polaris.NewCRSpecBuilderFromCobraFlags()
	updatePolarisReportingCobraHelper = *polarisreporting.NewHelmValuesFromCobraFlags()

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

	/* Update Black Duck Comamnds */

	// updateBlackDuckCmd
	updateBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updateBlackDuckCobraHelper.AddCRSpecFlagsToCommand(updateBlackDuckCmd, false)
	addMockFlag(updateBlackDuckCmd)
	updateCmd.AddCommand(updateBlackDuckCmd)

	// updateBlackDuckMasterKeyCmd
	updateBlackDuckCmd.AddCommand(updateBlackDuckMasterKeyCmd)

	// updateBlackDuckMasterKeyNativeCmd
	updateBlackDuckMasterKeyCmd.AddCommand(updateBlackDuckMasterKeyNativeCmd)

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

	// Polaris
	updatePolarisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updatePolarisCobraHelper.AddCRSpecFlagsToCommand(updatePolarisCmd, false)
	addbaseURLFlag(updatePolarisCmd)
	updateCmd.AddCommand(updatePolarisCmd)

	// Polaris-Reporting
	updatePolarisReportingCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(updatePolarisReportingCmd.PersistentFlags(), "namespace")
	updatePolarisReportingCobraHelper.AddCobraFlagsToCommand(updatePolarisReportingCmd, false)
	addChartLocationPathFlag(updatePolarisReportingCmd)
	updateCmd.AddCommand(updatePolarisReportingCmd)
}
