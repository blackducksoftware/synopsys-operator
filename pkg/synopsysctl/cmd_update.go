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
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update Command Resource Ctls
var updateAlertCtl ResourceCtl
var updateBlackDuckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl

// updateCmd provides functionality to update/upgrade features of
// Synopsys resources
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Synopsys resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Must specify a sub-command")
	},
}

// updateOperatorCmd lets the user update Synopsys Operator
var updateOperatorCmd = &cobra.Command{
	Use:     "operator",
	Example: "synopsysctl update operator --synopsys-operator-image docker.io/new_image_url\nsynopsysctl update operator --enable-blackduck\nsynopsysctl update operator --enable-blackduck -n <namespace>\nsynopsysctl update operator --expose-ui OPENSHIFT",
	Short:   "Update Synopsys Operator",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read Commandline Parameters
		if len(namespace) > 0 {
			updateOperator(namespace, cmd)
		} else {
			namespace := DefaultOperatorNamespace
			var err error
			if !cmd.LocalFlags().Lookup("mock").Changed && !cmd.LocalFlags().Lookup("mock-kube").Changed {
				isClusterScoped := util.GetClusterScope(apiExtensionClient)
				if isClusterScoped {
					namespace, err = util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
					if err != nil {
						log.Error(err)
						return nil
					}
				}
			}
			updateOperator(namespace, cmd)
		}
		return nil
	},
}

func updateOperator(namespace string, cmd *cobra.Command) error {
	log.Infof("updating the synopsys operator in '%s' namespace...", namespace)
	// Create new Synopsys-Operator SpecConfig
	oldOperatorSpec, err := soperator.GetOldOperatorSpec(restconfig, kubeClient, namespace)
	if err != nil {
		log.Errorf("unable to update the synopsys operator because %+v", err)
		return nil
	}
	newOperatorSpec := soperator.SpecConfig{}
	// Update Spec with changed values
	if cmd.Flag("synopsys-operator-image").Changed {
		log.Debugf("updating synopsys operator image to %s", synopsysOperatorImage)
		// check image tag
		imageHasTag := len(strings.Split(synopsysOperatorImage, ":")) == 2
		if !imageHasTag {
			log.Errorf("synopsys operator's image does not have a tag: %s", synopsysOperatorImage)
			return nil
		}
		newOperatorSpec.Image = synopsysOperatorImage
	}
	if cmd.Flag("expose-ui").Changed {
		log.Debugf("updating expose ui")
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

	// list the existing CRD's and convert them to map with key as name and scope as value
	crds := []string{}
	crdMap := make(map[string]string)
	var crdList *apiextensions.CustomResourceDefinitionList
	if !cmd.LocalFlags().Lookup("mock").Changed && !cmd.LocalFlags().Lookup("mock-kube").Changed {
		crdList, err = util.ListCustomResourceDefinitions(apiExtensionClient, "app=synopsys-operator")
		if err != nil {
			log.Errorf("unable to list the custom resource definitions due to %+v", err)
			return nil
		}
		for _, crd := range crdList.Items {
			crds = append(crds, crd.Name)
			crdMap[crd.Name] = crd.Name
		}
	}

	deleteCrds := []string{}
	// validate whether Alert CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.LocalFlags().Lookup("enable-alert").Changed {
		log.Debugf("updating enable Alert")
		_, ok := crdMap[operatorutil.AlertCRDName]
		if ok && isEnabledAlert {
			log.Errorf("%s custom resource definition already exists...", util.AlertCRDName)
		} else if !ok && isEnabledAlert {
			crds = append(crds, util.AlertCRDName)
		} else {
			deleteCrds = append(deleteCrds, operatorutil.AlertCRDName)
			crds = util.RemoveFromStringSlice(crds, util.AlertCRDName)
		}
	}

	// validate whether Black Duck CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.LocalFlags().Lookup("enable-blackduck").Changed {
		log.Debugf("updating enable Black Duck")
		_, ok := crdMap[operatorutil.BlackDuckCRDName]
		if ok && isEnabledBlackDuck {
			log.Errorf("%s custom resource definition already exists...", util.BlackDuckCRDName)
		} else if !ok && isEnabledBlackDuck {
			crds = append(crds, util.BlackDuckCRDName)
		} else {
			deleteCrds = append(deleteCrds, operatorutil.BlackDuckCRDName)
			crds = util.RemoveFromStringSlice(crds, util.BlackDuckCRDName)
		}
	}

	// validate whether OpsSight CRD enable parameter is enabled/disabled and add/remove them from the cluster
	if cmd.LocalFlags().Lookup("enable-opssight").Changed {
		log.Debugf("updating enable OpsSight")
		_, ok := crdMap[operatorutil.OpsSightCRDName]
		if ok && isEnabledOpsSight {
			log.Errorf("%s custom resource definition already exists...", util.OpsSightCRDName)
		} else if !ok && isEnabledOpsSight {
			crds = append(crds, util.OpsSightCRDName)
		} else {
			deleteCrds = append(deleteCrds, operatorutil.OpsSightCRDName)
			crds = util.RemoveFromStringSlice(crds, util.OpsSightCRDName)
		}
	}

	newOperatorSpec.Crds = crds

	newOperatorSpec.IsClusterScoped = util.GetClusterScope(apiExtensionClient)

	// merge old and new data
	err = mergo.Merge(&newOperatorSpec, oldOperatorSpec)
	if err != nil {
		log.Errorf("unable to merge old and new synopsys operator info because %+v", err)
		return nil
	}

	// update synopsys operator
	if cmd.LocalFlags().Lookup("mock").Changed {
		log.Debugf("running mock mode")
		// assigning the rest config to nil to run in mock mode. Getting weird issue if it is not nil
		newOperatorSpec.RestConfig = nil
		err := PrintResource(newOperatorSpec, mockFormat, false)
		if err != nil {
			log.Errorf("%s", err)
		}
	} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
		log.Debugf("running kube mock mode")
		err := PrintResource(newOperatorSpec, mockKubeFormat, true)
		if err != nil {
			log.Errorf("%s", err)
		}
	} else {
		sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
		// update Synopsys Operator
		err = sOperatorCreater.EnsureSynopsysOperator(namespace, blackDuckClient, opsSightClient, alertClient, oldOperatorSpec, &newOperatorSpec)
		if err != nil {
			log.Errorf("unable to update the synopsys operator because %+v", err)
			return nil
		}

		log.Debugf("updating Prometheus in namespace %s", namespace)
		// Create new Prometheus SpecConfig
		oldPrometheusSpec, err := soperator.GetOldPrometheusSpec(restconfig, kubeClient, namespace)
		if err != nil {
			log.Errorf("error in updating the prometheus because %+v", err)
			return nil
		}

		// check for changes
		newPrometheusSpec := soperator.PrometheusSpecConfig{}
		if cmd.Flag("metrics-image").Changed {
			log.Debugf("updating Prometheus Image to %s", metricsImage)
			newPrometheusSpec.Image = metricsImage
		}
		if cmd.Flag("expose-metrics").Changed {
			log.Debugf("updating expose metrics")
			newPrometheusSpec.Expose = exposeMetrics
		}

		// merge old and new data
		err = mergo.Merge(&newPrometheusSpec, oldPrometheusSpec)
		if err != nil {
			log.Errorf("unable to merge old and new prometheus info because %+v", err)
			return nil
		}

		// update prometheus
		err = sOperatorCreater.UpdatePrometheus(&newPrometheusSpec)
		if err != nil {
			log.Errorf("unable to update Prometheus because %+v", err)
			return nil
		}

		log.Infof("successfully updated the synopsys operator in '%s' namespace", namespace)
	}
	return nil
}

// updateAlertCmd lets the user update an Alert Instance
var updateAlertCmd = &cobra.Command{
	Use:     "alert NAME",
	Example: "synopsysctl update alert <name> --port 80\nsynopsysctl update alert <name> -n <namespace> --port 80",
	Short:   "Describe an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertName, alertNamespace, _, err := getInstanceInfo(cmd, args, util.AlertCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("updating Alert '%s' instance in '%s' namespace...", alertName, alertNamespace)

		// Get Alert
		currAlert, err := operatorutil.GetAlert(alertClient, alertNamespace, alertName)
		if err != nil {
			log.Errorf("error getting an Alert %s instance in %s namespace due to %+v", alertName, alertNamespace, err)
			return nil
		}
		err = updateAlertCtl.SetSpec(currAlert.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Alert instance in %s namespace to spec due to %+v", alertName, alertNamespace, err)
			return nil
		}

		// Check if it can be updated
		canUpdate, err := updateAlertCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update Alert: %s", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateAlertCtl.SetChangedFlags(flagset)
			newSpec := updateAlertCtl.GetSpec().(alertapi.AlertSpec)
			// merge environs
			newSpec.Environs = operatorutil.MergeEnvSlices(newSpec.Environs, currAlert.Spec.Environs)
			// Create new Alert CRD
			newAlert := *currAlert //make copy
			newAlert.Spec = newSpec
			// Update Alert
			err = ctlUpdateResource(newAlert, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Alert instance in %s namespace due to %+v", alertName, alertNamespace, err)
				return nil
			}
			log.Infof("successfully updated the '%s' Alert instance in '%s' namespace", alertName, alertNamespace)
		}
		return nil
	},
}

// updateBlackDuckCmd lets the user update a Black Duck instance
var updateBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAME",
	Example: "synopsyctl update blackduck <name> --size medium\nsynopsyctl update blackduck <name> -n <namespace> --size medium",
	Short:   "Update a Black Duck instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("updating Black Duck '%s' instance in '%s' namespace...", blackDuckName, blackDuckNamespace)

		// Get Black Duck
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance in %s namespace to spec due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateBlackDuckCtl.SetChangedFlags(flagset)
			newSpec := updateBlackDuckCtl.GetSpec().(blackduckapi.BlackduckSpec)
			// merge environs
			newSpec.Environs = operatorutil.MergeEnvSlices(newSpec.Environs, currBlackDuck.Spec.Environs)
			// Create new Black Duck CRD
			newBlackDuck := *currBlackDuck //make copy
			newBlackDuck.Spec = newSpec
			// Update Black Duck
			err = ctlUpdateResource(newBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
			log.Infof("successfully updated the '%s' Black Duck instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		}
		return nil
	},
}

// updateBlackDuckRootKeyCmd create new Black Duck root key for source code upload in the cluster
var updateBlackDuckRootKeyCmd = &cobra.Command{
	Use:     "masterkey NEW_SEAL_KEY STORED_MASTER_KEY_FILE_PATH",
	Example: "synopsysctl update blackduck masterkey <new seal key> <file path of the stored master key>",
	Short:   "Update the root key to all Black Duck instance for source code upload functionality",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var crdScope apiextensions.ResourceScope
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.BlackDuckCRDName)
		if err != nil {
			log.Errorf("unable to get the %s custom resource definition in your cluster due to %+v", util.BlackDuckCRDName, err)
			return nil
		}
		crdScope = crd.Spec.Scope

		// Check Number of Arguments
		if crdScope != apiextensions.ClusterScoped && len(namespace) == 0 {
			log.Errorf("namespace to update the Black Duck instance need to be provided")
			return nil
		}

		var operatorNamespace string
		if crdScope == apiextensions.ClusterScoped {
			operatorNamespace, err = getOperatorNamespace(metav1.NamespaceAll)
			if err != nil {
				log.Errorf("unable to find the Synopsys Operator instance due to %+v", err)
				return nil
			}
		} else {
			operatorNamespace = namespace
		}

		newSealKey := args[1]
		filePath := args[2]

		blackducks, err := operatorutil.ListHubs(blackduckClient, operatorNamespace)
		if err != nil {
			log.Errorf("unable to list Black Duck instances in %s namespace because %+v", operatorNamespace, err)
			return nil
		}

		secret, err := operatorutil.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
			return nil
		}

		for _, blackduck := range blackducks.Items {
			blackDuckName := blackduck.Name
			blackDuckNamespace := blackduck.Namespace
			log.Infof("updating %s Black Duck instance master key in %s namespace...", blackDuckName, blackDuckNamespace)

			fileName := filepath.Join(filePath, fmt.Sprintf("%s-%s.key", blackDuckNamespace, blackDuckName))
			masterKey, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Errorf("error reading the master key from %s because %+v", fileName, err)
				return nil
			}

			// Filter the upload cache pod to get the root key using the seal key
			uploadCachePod, err := operatorutil.FilterPodByNamePrefixInNamespace(kubeClient, blackDuckNamespace, util.GetResourceName(blackDuckName, util.BlackDuckName, "uploadcache", crdScope == apiextensions.ClusterScoped))
			if err != nil {
				log.Errorf("unable to filter the upload cache pod of %s because %+v", blackDuckNamespace, err)
				return nil
			}

			// Create the exec into kubernetes pod request
			req := operatorutil.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
			// TODO: changed the upload cache service name to authentication until the HUB-20412 is fixed. once it if fixed, changed the name to use GetResource method
			_, err = operatorutil.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://uploadcache:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
			if err != nil {
				log.Errorf("unable to exec into upload cache pod in %s because %+v", blackDuckNamespace, err)
				return nil
			}

			log.Infof("successfully updated %s Black Duck instance master key in %s namespace", blackDuckName, operatorNamespace)
		}

		secret.Data["SEAL_KEY"] = []byte(newSealKey)

		err = operatorutil.UpdateSecret(kubeClient, operatorNamespace, secret)
		if err != nil {
			log.Errorf("unable to update the Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
			return nil
		}

		return nil
	},
}

var blackDuckPVCSize = "2Gi"
var blackDuckPVCStorageClass = ""

// updateBlackDuckAddPVCCmd adds a PVC to a Black Duck
var updateBlackDuckAddPVCCmd = &cobra.Command{
	Use:     "addpvc BLACK_DUCK_NAME PVC_NAME",
	Example: "synopsysctl update blackduck addpvc bdname mypvc --size 2Gi --storage-class standard\nsynopsysctl update blackduck addpvc bdname mypvc --size 2Gi --storage-class standard -n bdnamespace",
	Short:   "Add a Persistent Volume Claim to a Black Duck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}

		pvcName := args[1]

		log.Infof("adding PVC to Black Duck %s instance in %s namespace...", blackDuckName, blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance in %s namespace to spec due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Add PVC to Spec
			newPVC := blackduckapi.PVC{
				Name:         pvcName,
				Size:         blackDuckPVCSize,
				StorageClass: blackDuckPVCStorageClass,
			}
			currBlackDuck.Spec.PVC = append(currBlackDuck.Spec.PVC, newPVC)
			// Update Black Duck with PVC
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddEnvironCmd adds an environ to a Blackduck
var updateBlackDuckAddEnvironCmd = &cobra.Command{
	Use:     "addenviron BLACK_DUCK_NAME (ENVIRON_NAME:ENVIRON_VALUE)",
	Example: "synopsysctl update blackduck addenviron bdnamespace USE_ALERT:1",
	Short:   "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		environ := args[1]

		log.Infof("adding Environ to Black Duck %s instance in %s namespace...", blackDuckName, blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance in %s namespace to spec due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Merge Environ to Spec
			currBlackDuck.Spec.Environs = operatorutil.MergeEnvSlices(strings.Split(environ, ","), currBlackDuck.Spec.Environs)
			// Update Black Duck with Environ
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance in '%s' namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddRegistryCmd adds an Image Registry to a Blackduck
var updateBlackDuckAddRegistryCmd = &cobra.Command{
	Use:     "addregistry BLACK_DUCK_NAME REGISTRY",
	Example: "synopsysctl update blackduck addregistry bdnamespace docker.io",
	Short:   "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		registry := args[1]

		log.Infof("adding an Image Registry to Black Duck %s instance in %s namespace...", blackDuckName, blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance in %s namespace to spec due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Add Registry to Spec
			currBlackDuck.Spec.ImageRegistries = append(currBlackDuck.Spec.ImageRegistries, registry)
			// Update Black Duck with Environ
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance in %s namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddUIDCmd adds a UID mapping to a Blackduck
var updateBlackDuckAddUIDCmd = &cobra.Command{
	Use:     "adduid BLACK_DUCK_NAME UID_KEY UID_VALUE",
	Example: "synopsysctl update blackduck adduid bdnamespace uidname 80",
	Short:   "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckName, blackDuckNamespace, _, err := getInstanceInfo(cmd, args, util.BlackDuckCRDName, namespace)
		if err != nil {
			log.Error(err)
			return nil
		}
		uidKey := args[1]
		uidVal := args[2]

		log.Debugf("adding an Image UID to Black Duck %s in %s namespace...", blackDuckName, blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckName)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance in %s namespace to spec due to %+v", blackDuckName, blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Add UID Mapping to Spec
			intUIDVal, err := strconv.ParseInt(uidVal, 0, 64)
			if err != nil {
				log.Errorf("Couldn't convert UID_VAL to int: %s", err)
			}
			if currBlackDuck.Spec.ImageUIDMap == nil {
				currBlackDuck.Spec.ImageUIDMap = make(map[string]int64)
			}
			currBlackDuck.Spec.ImageUIDMap[uidKey] = intUIDVal
			// Update Black Duck with UID mapping
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance in %s namespace due to %+v", blackDuckName, blackDuckNamespace, err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance in %s namespace", blackDuckName, blackDuckNamespace)
		return nil
	},
}

// updateOpsSightCmd lets the user update an OpsSight instance
var updateOpsSightCmd = &cobra.Command{
	Use:     "opssight NAMESPACE",
	Example: "synopsyctl update opssight opsnamespace --blackduck-max-count 2",
	Short:   "Update an instance of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		log.Infof("updating OpsSight %s...", opsSightNamespace)

		// Get the current OpsSight
		currOpsSight, err := operatorutil.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("error getting OpsSight: %s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateOpsSightCtl.SetChangedFlags(flagset)
			newSpec := updateOpsSightCtl.GetSpec().(opssightapi.OpsSightSpec)
			// Create new OpsSight CRD
			newOpsSight := *currOpsSight //make copy
			newOpsSight.Spec = newSpec
			// Update OpsSight
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("failed to update OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight: '%s'", opsSightNamespace)
		}
		return nil
	},
}

// updateOpsSightImageCmd lets the user update an image in an OpsSight instance
var updateOpsSightImageCmd = &cobra.Command{
	Use:     "image NAMESPACE OPSSIGHTCORE|SCANNER|IMAGEGETTER|IMAGEPROCESSOR|PODPROCESSOR|METRICS IMAGE",
	Example: "synopsysctl update opssight image opsnamespace SCANNER docker.io/new_scanner_image_url",
	Short:   "Update an image of an OpsSight component",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		componentName := args[1]
		componentImage := args[2]

		log.Infof("updating OpsSight %s's Image...", opsSightNamespace)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update OpsSight: %s", err)
			return nil
		}
		if canUpdate {
			newOpsSight := *currOpsSight //make copy
			// Update the Spec with new Image
			switch strings.ToUpper(componentName) {
			case "OPSSIGHTCORE":
				currOpsSight.Spec.Perceptor.Image = componentImage
			case "SCANNER":
				currOpsSight.Spec.ScannerPod.Scanner.Image = componentImage
			case "IMAGEGETTER":
				currOpsSight.Spec.ScannerPod.ImageFacade.Image = componentImage
			case "IMAGEPROCESSOR":
				currOpsSight.Spec.Perceiver.ImagePerceiver.Image = componentImage
			case "PODPROCESSOR":
				currOpsSight.Spec.Perceiver.PodPerceiver.Image = componentImage
			case "METRICS":
				currOpsSight.Spec.Prometheus.Image = componentImage
			default:
				return fmt.Errorf("'%s' is not a valid component", componentName)
			}
			// Update OpsSight with New Image
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("failed to update OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's Image", opsSightNamespace)
		}
		return nil
	},
}

// updateOpsSightExternalHostCmd lets the user update an OpsSight with an External Host
var updateOpsSightExternalHostCmd = &cobra.Command{
	Use:     "externalhost NAMESPACE SCHEME DOMAIN PORT USER PASSWORD SCANLIMIT",
	Example: "synopsysctl update opssight externalhost opsnamespace scheme domain 80 user pass 50",
	Short:   "Update an external host for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 7 {
			return fmt.Errorf("this command takes 7 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		hostScheme := args[1]
		hostDomain := args[2]
		hostPort, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			log.Errorf("invalid Port Number: '%s'", err)
		}
		hostUser := args[4]
		hostPassword := args[5]
		hostScanLimit, err := strconv.ParseInt(args[6], 0, 64)
		if err != nil {
			log.Errorf("invalid Concurrent Scan Limit: %s", err)
		}

		log.Infof("adding External Host to OpsSight %s...", opsSightNamespace)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("error getting OpsSight: %s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update OpsSight: %s", err)
			return nil
		}
		if canUpdate {
			newOpsSight := *currOpsSight //make copy
			// Add External Host to Spec
			newHost := opssightapi.Host{
				Scheme:              hostScheme,
				Domain:              hostDomain,
				Port:                int(hostPort),
				User:                hostUser,
				Password:            hostPassword,
				ConcurrentScanLimit: int(hostScanLimit),
			}
			newOpsSight.Spec.Blackduck.ExternalHosts = append(newOpsSight.Spec.Blackduck.ExternalHosts, &newHost)
			// Update OpsSight with External Host
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("failed to update OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's External Host", opsSightNamespace)
		}
		return nil
	},
}

// updateOpsSightAddRegistryCmd lets the user update and OpsSight by
// adding a registry for the ImageFacade
var updateOpsSightAddRegistryCmd = &cobra.Command{
	Use:     "registry NAMESPACE URL USER PASSWORD",
	Example: "synopsysctl update opssight registry opsnamespace reg_url reg_username reg_password",
	Short:   "Add an Internal Registry to OpsSight's ImageFacade",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return fmt.Errorf("this command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightNamespace := args[0]
		regURL := args[1]
		regUser := args[2]
		regPass := args[3]

		log.Infof("adding Internal Registry to OpsSight %s...", opsSightNamespace)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opsSightClient, opsSightNamespace, opsSightNamespace)
		if err != nil {
			log.Errorf("error adding Internal Registry while getting OpsSight: %s", err)
			return nil
		}
		// Check if it can be updated
		updateOpsSightCtl.SetSpec(currOpsSight.Spec)
		canUpdate, err := updateOpsSightCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update OpsSight: %s", err)
			return nil
		}
		if canUpdate {
			newOpsSight := *currOpsSight //make copy
			// Add Internal Registry to Spec
			newReg := opssightapi.RegistryAuth{
				URL:      regURL,
				User:     regUser,
				Password: regPass,
			}
			newOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries = append(newOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
			// Update OpsSight with Internal Registry
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, mockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, mockKubeFormat)
			if err != nil {
				log.Errorf("failed to update OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's Registry", opsSightNamespace)
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackDuckCtl = blackduck.NewBlackDuckCtl()
	updateOpsSightCtl = opssight.NewOpsSightCtl()
	updateAlertCtl = alert.NewAlertCtl()

	rootCmd.AddCommand(updateCmd)

	// Add Operator Commands
	updateOperatorCmd.Flags().BoolVarP(&isEnabledAlert, "enable-alert", "a", isEnabledAlert, "Enable/Disable Alert Custom Resource Definition (CRD) in your cluster")
	updateOperatorCmd.Flags().BoolVarP(&isEnabledBlackDuck, "enable-blackduck", "b", isEnabledBlackDuck, "Enable/Disable Black Duck Custom Resource Definition (CRD) in your cluster")
	updateOperatorCmd.Flags().BoolVarP(&isEnabledOpsSight, "enable-opssight", "s", isEnabledOpsSight, "Enable/Disable OpsSight Custom Resource Definition (CRD) in your cluster")
	// updateOperatorCmd.Flags().BoolVarP(&isEnabledPrm, "enable-prm", "p", isEnabledPrm, "Enable/Disable Polaris Reporting Module Custom Resource Definition (CRD) in your cluster")
	updateOperatorCmd.Flags().StringVarP(&exposeUI, "expose-ui", "e", exposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&synopsysOperatorImage, "synopsys-operator-image", "i", synopsysOperatorImage, "Image URL of Synopsys Operator")
	updateOperatorCmd.Flags().StringVarP(&exposeMetrics, "expose-metrics", "x", exposeMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&metricsImage, "metrics-image", "m", metricsImage, "Image URL of Synopsys Operator's metrics pod")
	updateOperatorCmd.Flags().Int64VarP(&postgresRestartInMins, "postgres-restart-in-minutes", "q", postgresRestartInMins, "Minutes to check for restarting postgres")
	updateOperatorCmd.Flags().Int64VarP(&podWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", podWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	updateOperatorCmd.Flags().Int64VarP(&resyncIntervalInSeconds, "resync-interval-in-seconds", "r", resyncIntervalInSeconds, "Seconds for resyncing custom resources")
	updateOperatorCmd.Flags().Int64VarP(&terminationGracePeriodSeconds, "postgres-termination-grace-period", "g", terminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	updateOperatorCmd.Flags().StringVarP(&dryRun, "dry-run", "d", dryRun, "If true, Synopsys Operator runs without being connected to a cluster [true|false]")
	updateOperatorCmd.Flags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level of Synopsys Operator")
	updateOperatorCmd.Flags().IntVarP(&threadiness, "no-of-threads", "t", threadiness, "Number of threads to process the custom resources")
	updateOperatorCmd.Flags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Prints the Synopsys Operator spec in the specified format instead of creating it [json|yaml]")
	updateOperatorCmd.Flags().StringVarP(&mockKubeFormat, "mock-kube", "k", mockKubeFormat, "Prints the Synopsys Operator's kubernetes resource specs in the specified format instead of creating it [json|yaml]")
	updateOperatorCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to create the resource(s)")
	updateCmd.AddCommand(updateOperatorCmd)

	// Add Alert Commands
	updateAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateAlertCmd.PersistentFlags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateAlertCmd.PersistentFlags().StringVarP(&mockKubeFormat, "mock-kube", "k", mockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateAlertCtl.AddSpecFlags(updateAlertCmd, false)
	updateCmd.AddCommand(updateAlertCmd)

	// Add Black Duck Commands
	updateBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckCmd.Flags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateBlackDuckCmd.Flags().StringVarP(&mockKubeFormat, "mock-kube", "k", mockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateBlackDuckCtl.AddSpecFlags(updateBlackDuckCmd, false)
	updateCmd.AddCommand(updateBlackDuckCmd)

	updateBlackDuckRootKeyCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckCmd.AddCommand(updateBlackDuckRootKeyCmd)

	updateBlackDuckAddPVCCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckAddPVCCmd.Flags().StringVar(&blackDuckPVCSize, "size", blackDuckPVCSize, "Size of the PVC")
	updateBlackDuckAddPVCCmd.Flags().StringVar(&blackDuckPVCStorageClass, "storage-class", blackDuckPVCStorageClass, "Storage Class name")
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddPVCCmd)

	updateBlackDuckAddEnvironCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddEnvironCmd)

	updateBlackDuckAddRegistryCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddRegistryCmd)

	updateBlackDuckAddUIDCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to update the resource(s)")
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddUIDCmd)

	// Add OpsSight Commands
	updateOpsSightCmd.PersistentFlags().StringVarP(&mockFormat, "mock", "o", mockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateOpsSightCmd.PersistentFlags().StringVarP(&mockKubeFormat, "mock-kube", "k", mockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateOpsSightCtl.AddSpecFlags(updateOpsSightCmd, false)
	updateCmd.AddCommand(updateOpsSightCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightImageCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightExternalHostCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightAddRegistryCmd)
}
