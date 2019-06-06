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
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update Command Resource Ctls
var updateAlertCtl ResourceCtl
var updateBlackDuckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl

// Update Comamnd Defaults
var updateSynopsysOperatorImage = ""
var updatePrometheusImage = ""
var updateExposeUI = ""
var updateExposePrometheusMetrics = ""
var updateTerminationGracePeriodSeconds int64
var updateOperatorTimeBombInSeconds int64
var updateLogLevel = ""
var updateThreadiness int
var updatePostgresRestartInMins int64
var updatePodWaitTimeoutSeconds int64
var updateResyncIntervalInSeconds int64

// Flags for using mock mode - doesn't deploy
var updateMockFormat string
var updateMockKubeFormat string

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
	Example: "synopsysctl update operator --synopsys-operator-image docker.io/new_image_url\nsynopsysctl update operator --expose-ui OPENSHIFT",
	Short:   "Update Synopsys Operator",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 0 {
			return fmt.Errorf("this command takes 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, err := operatorutil.GetOperatorNamespace(kubeClient)
		if err != nil {
			log.Errorf("unable to find Synopsys Operator in the cluster because %+v", err)
			return nil
		}

		log.Infof("updating Synopsys Operator in '%s' namespace...", namespace)

		// Create new Synopsys Operator SpecConfig
		oldOperatorSpec, err := soperator.GetOldOperatorSpec(kubeClient, namespace)
		if err != nil {
			log.Errorf("unable to update Synopsys Operator because %+v", err)
			return nil
		}
		newOperatorSpec := soperator.SpecConfig{}
		// Update Spec with changed values
		if cmd.Flag("synopsys-operator-image").Changed {
			log.Debugf("updating Synopsys Operator image to %s", updateSynopsysOperatorImage)
			// check image tag
			imageHasTag := len(strings.Split(updateSynopsysOperatorImage, ":")) == 2
			if !imageHasTag {
				log.Errorf("Synopsys Operator's image does not have a tag: %s", updateSynopsysOperatorImage)
				return nil
			}
			newOperatorSpec.Image = updateSynopsysOperatorImage
		}
		if cmd.Flag("expose-ui").Changed {
			log.Debugf("updating expose ui")
			newOperatorSpec.Expose = updateExposeUI
		}
		if cmd.Flag("operator-time-bomb-in-seconds").Changed {
			log.Debugf("updating operator time bomb in seconds")
			newOperatorSpec.OperatorTimeBombInSeconds = updateOperatorTimeBombInSeconds
		}
		if cmd.Flag("postgres-restart-in-minutes").Changed {
			log.Debugf("updating postgres restart in minutes")
			newOperatorSpec.PostgresRestartInMins = updatePostgresRestartInMins
		}
		if cmd.Flag("pod-wait-timeout-in-seconds").Changed {
			log.Debugf("updating pod wait timeout in seconds")
			newOperatorSpec.PodWaitTimeoutSeconds = updatePodWaitTimeoutSeconds
		}
		if cmd.Flag("resync-interval-in-seconds").Changed {
			log.Debugf("updating resync interval in seconds")
			newOperatorSpec.ResyncIntervalInSeconds = updateResyncIntervalInSeconds
		}
		if cmd.Flag("postgres-termination-grace-period").Changed {
			log.Debugf("updating postgres termination grace period")
			newOperatorSpec.TerminationGracePeriodSeconds = updateTerminationGracePeriodSeconds
		}
		if cmd.Flag("log-level").Changed {
			log.Debugf("updating log level")
			newOperatorSpec.LogLevel = updateLogLevel
		}
		if cmd.Flag("no-of-threads").Changed {
			log.Debugf("updating no of threads")
			newOperatorSpec.Threadiness = updateThreadiness
		}

		// merge old and new data
		err = mergo.Merge(&newOperatorSpec, oldOperatorSpec)
		if err != nil {
			log.Errorf("unable to merge old and new Synopsys Operator info because %+v", err)
			return nil
		}

		if cmd.LocalFlags().Lookup("mock").Changed {
			log.Debugf("running mock mode")
			err := PrintResource(newOperatorSpec, updateMockFormat, false)
			if err != nil {
				log.Errorf("%s", err)
			}
		} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
			log.Debugf("running kube mock mode")
			err := PrintResource(newOperatorSpec, updateMockKubeFormat, true)
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
			if cmd.Flag("prometheus-image").Changed {
				log.Debugf("updating PrometheusImage to %s", updatePrometheusImage)
				newPrometheusSpec.Image = updatePrometheusImage
			}
			if cmd.Flag("expose-prometheus-metrics").Changed {
				log.Debugf("updating expose prometheus metrics")
				newPrometheusSpec.Expose = updateExposePrometheusMetrics
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
	},
}

// updateAlertCmd lets the user update an Alert Instance
var updateAlertCmd = &cobra.Command{
	Use:   "alert NAMESPACE",
	Short: "Describe an instance of Alert",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		alertNamespace := args[0]

		log.Infof("updating Alert %s instance...", alertNamespace)

		// Get the Alert
		currAlert, err := operatorutil.GetAlert(alertClient, alertNamespace, alertNamespace)
		if err != nil {
			log.Errorf("error getting an Alert %s instance due to %+v", alertNamespace, err)
			return nil
		}
		err = updateAlertCtl.SetSpec(currAlert.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Alert instance to spec due to %+v", alertNamespace, err)
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
			err = ctlUpdateResource(newAlert, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Alert: %s", err)
				return nil
			}
			log.Infof("successfully updated the '%s' Alert instance", alertNamespace)
		}
		return nil
	},
}

// updateBlackDuckCmd lets the user update a Black Duck instance
var updateBlackDuckCmd = &cobra.Command{
	Use:     "blackduck NAMESPACE",
	Example: "synopsyctl update blackduck bdnamespace --size medium",
	Short:   "Update a Black Duck instance",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		log.Infof("updating Black Duck %s instance...", blackDuckNamespace)

		// Get Black Duck
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
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
			err = ctlUpdateResource(newBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Black Duck: %s", err)
				return nil
			}
			log.Infof("successfully updated the '%s' Black Duck instance", blackDuckNamespace)
		}
		return nil
	},
}

// updateBlackDuckRootKeyCmd create new Black Duck root key for source code upload in the cluster
var updateBlackDuckRootKeyCmd = &cobra.Command{
	Use:     "rootkey NAMESPACE NEW_SEAL_KEY MASTER_KEY_FILE_PATH",
	Example: "synopsysctl update blackduck rootkey bdnamespace newsealkey ~/master/key/filepath",
	Short:   "Update the root key of Black Duck for source code upload",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		newSealKey := args[1]
		filePath := args[2]

		log.Infof("updating Black Duck %s Root Key...", blackDuckNamespace)

		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, metav1.NamespaceDefault, blackDuckNamespace)
		if err != nil {
			log.Errorf("unable to find Black Duck %s instance because %+v", blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
			return nil
		}
		// Check if it can be updated
		canUpdate, err := updateBlackDuckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			operatorNamespace, err := operatorutil.GetOperatorNamespace(kubeClient)
			if err != nil {
				log.Errorf("unable to find Synopsys Operator instance because %+v", err)
				return nil
			}

			fileName := filepath.Join(filePath, fmt.Sprintf("%s.key", blackDuckNamespace))
			masterKey, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Errorf("error reading the master key from %s because %+v", fileName, err)
				return nil
			}

			// Filter the upload cache pod to get the root key using the seal key
			uploadCachePod, err := operatorutil.FilterPodByNamePrefixInNamespace(kubeClient, blackDuckNamespace, "uploadcache")
			if err != nil {
				log.Errorf("unable to filter the upload cache pod of %s because %+v", blackDuckNamespace, err)
				return nil
			}

			// Create the exec into kubernetes pod request
			req := operatorutil.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
			_, err = operatorutil.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://uploadcache:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
			if err != nil {
				log.Errorf("unable to exec into upload cache pod in %s because %+v", blackDuckNamespace, err)
				return nil
			}

			secret, err := operatorutil.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
			if err != nil {
				log.Errorf("unable to find Synopsys Operator blackduck-secret in %s namespace because %+v", operatorNamespace, err)
				return nil
			}
			secret.Data["SEAL_KEY"] = []byte(newSealKey)

			if cmd.LocalFlags().Lookup("mock").Changed {
				log.Debugf("running mock mode")
				PrintComponent(secret, updateMockFormat)
			} else if cmd.LocalFlags().Lookup("mock-kube").Changed {
				log.Debugf("running kube mock mode")
				PrintComponent(secret, updateMockKubeFormat)
			} else {
				err = operatorutil.UpdateSecret(kubeClient, operatorNamespace, secret)
				if err != nil {
					log.Errorf("unable to update the Synopsys Operator blackduck-secret in %s namespace because %+v", operatorNamespace, err)
					return nil
				}
			}
		}
		log.Infof("successfully updated Black Duck %s's Root Key", blackDuckNamespace)
		return nil
	},
}

var blackDuckPVCSize = "2Gi"
var blackDuckPVCStorageClass = ""

// updateBlackDuckAddPVCCmd adds a PVC to a Black Duck
var updateBlackDuckAddPVCCmd = &cobra.Command{
	Use:     "addpvc NAMESPACE PVC_NAME",
	Example: "synopsysctl update blackduck addpvc bdnamespace mypvc --size 2Gi --storage-class standard",
	Short:   "Add a Persistent Volume Claim to a Black Duck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		pvcName := args[1]

		log.Infof("adding PVC to Black Duck %s instance...", blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
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
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Black Duck: %s", err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance", blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddEnvironCmd adds an environ to a Blackduck
var updateBlackDuckAddEnvironCmd = &cobra.Command{
	Use:     "addenviron NAMESPACE (ENVIRON_NAME:ENVIRON_VALUE)",
	Example: "synopsysctl update blackduck addenviron bdnamespace USE_ALERT:1",
	Short:   "Add an Environment Variable to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		environ := args[1]

		log.Infof("adding Environ to Black Duck %s instance...", blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
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
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Black Duck: %s", err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance", blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddRegistryCmd adds an Image Registry to a Blackduck
var updateBlackDuckAddRegistryCmd = &cobra.Command{
	Use:     "addregistry NAMESPACE REGISTRY",
	Example: "synopsysctl update blackduck addregistry bdnamespace docker.io",
	Short:   "Add an Image Registry to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		registry := args[1]

		log.Infof("adding an Image Registry to Black Duck %s instance...", blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("error getting %s Black Duck instance due to %+v", blackDuckNamespace, err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
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
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Black Duck: %s", err)
				return nil
			}
		}
		log.Infof("successfully updated the '%s' Black Duck instance", blackDuckNamespace)
		return nil
	},
}

// updateBlackDuckAddUIDCmd adds a UID mapping to a Blackduck
var updateBlackDuckAddUIDCmd = &cobra.Command{
	Use:     "adduid NAMESPACE UID_KEY UID_VALUE",
	Example: "synopsysctl update blackduck adduid bdnamespace uidname 80",
	Short:   "Add an Image UID to Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackDuckNamespace := args[0]
		uidKey := args[1]
		uidVal := args[2]

		log.Debugf("adding an Image UID to Black Duck %s...", blackDuckNamespace)

		// Get Black Duck Spec
		currBlackDuck, err := operatorutil.GetHub(blackDuckClient, blackDuckNamespace, blackDuckNamespace)
		if err != nil {
			log.Errorf("%s", err)
			return nil
		}
		// Set the existing Black Duck to the spec
		err = updateBlackDuckCtl.SetSpec(currBlackDuck.Spec)
		if err != nil {
			log.Errorf("cannot set an existing %s Black Duck instance to spec due to %+v", blackDuckNamespace, err)
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
			err = ctlUpdateResource(currBlackDuck, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
			if err != nil {
				log.Errorf("failed to update Black Duck: %s", err)
				return nil
			}
		}
		log.Infof("successfully updated Black Duck: '%s'", blackDuckNamespace)
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
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
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
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
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
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
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
			err = ctlUpdateResource(newOpsSight, cmd.LocalFlags().Lookup("mock").Changed, updateMockFormat, cmd.LocalFlags().Lookup("mock-kube").Changed, updateMockKubeFormat)
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
	updateOperatorCmd.Flags().StringVarP(&updateExposeUI, "expose-ui", "e", updateExposeUI, "Service type to expose Synopsys Operator's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&updateSynopsysOperatorImage, "synopsys-operator-image", "i", updateSynopsysOperatorImage, "Image URL of Synopsys Operator")
	updateOperatorCmd.Flags().StringVarP(&updateExposePrometheusMetrics, "expose-metrics", "x", updateExposePrometheusMetrics, "Service type to expose Synopsys Operator's metrics application [NODEPORT|LOADBALANCER|OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&updatePrometheusImage, "metrics-image", "m", updatePrometheusImage, "Image URL of Synopsys Operator's metrics pod")
	updateOperatorCmd.Flags().Int64VarP(&updateOperatorTimeBombInSeconds, "operator-time-bomb-in-seconds", "t", updateOperatorTimeBombInSeconds, "Termination grace period in seconds for shutting down crds")
	updateOperatorCmd.Flags().Int64VarP(&updatePostgresRestartInMins, "postgres-restart-in-minutes", "n", updatePostgresRestartInMins, "Minutes to check for restarting postgres")
	updateOperatorCmd.Flags().Int64VarP(&updatePodWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", updatePodWaitTimeoutSeconds, "Seconds to wait for pods to be running")
	updateOperatorCmd.Flags().Int64VarP(&updateResyncIntervalInSeconds, "resync-interval-in-seconds", "r", updateResyncIntervalInSeconds, "Seconds for resyncing custom resources")
	updateOperatorCmd.Flags().Int64VarP(&updateTerminationGracePeriodSeconds, "postgres-termination-grace-period", "g", updateTerminationGracePeriodSeconds, "Termination grace period in seconds for shutting down postgres")
	updateOperatorCmd.Flags().StringVarP(&updateLogLevel, "log-level", "l", updateLogLevel, "Log level of Synopsys Operator")
	updateOperatorCmd.Flags().IntVarP(&updateThreadiness, "no-of-threads", "c", updateThreadiness, "Number of threads to process the custom resources")
	updateOperatorCmd.Flags().StringVarP(&updateMockFormat, "mock", "o", updateMockFormat, "Prints the Synopsys Operator spec in the specified format instead of creating it [json|yaml]")
	updateOperatorCmd.Flags().StringVarP(&updateMockKubeFormat, "mock-kube", "k", updateMockKubeFormat, "Prints the Synopsys Operator's kubernetes resource specs in the specified format instead of creating it [json|yaml]")

	updateCmd.AddCommand(updateOperatorCmd)

	// Add Alert Commands
	updateAlertCmd.PersistentFlags().StringVarP(&updateMockFormat, "mock", "o", updateMockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateAlertCmd.PersistentFlags().StringVarP(&updateMockKubeFormat, "mock-kube", "k", updateMockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateAlertCtl.AddSpecFlags(updateAlertCmd, false)
	updateCmd.AddCommand(updateAlertCmd)

	// Add Black Duck Commands
	updateBlackDuckCmd.Flags().StringVar(&updateMockFormat, "mock", updateMockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateBlackDuckCmd.Flags().StringVar(&updateMockKubeFormat, "mock-kube", updateMockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateBlackDuckCtl.AddSpecFlags(updateBlackDuckCmd, false)
	updateCmd.AddCommand(updateBlackDuckCmd)
	updateBlackDuckCmd.AddCommand(updateBlackDuckRootKeyCmd)

	updateBlackDuckAddPVCCmd.Flags().StringVar(&blackDuckPVCSize, "size", blackDuckPVCSize, "Size of the PVC")
	updateBlackDuckAddPVCCmd.Flags().StringVar(&blackDuckPVCStorageClass, "storage-class", blackDuckPVCStorageClass, "Storage Class name")
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddPVCCmd)

	updateBlackDuckCmd.AddCommand(updateBlackDuckAddEnvironCmd)

	updateBlackDuckCmd.AddCommand(updateBlackDuckAddRegistryCmd)

	updateBlackDuckCmd.AddCommand(updateBlackDuckAddUIDCmd)

	// Add OpsSight Commands
	updateOpsSightCmd.PersistentFlags().StringVarP(&updateMockFormat, "mock", "o", updateMockFormat, "Prints the new CRD resource spec in the specified format instead of editing it [json|yaml]")
	updateOpsSightCmd.PersistentFlags().StringVarP(&updateMockKubeFormat, "mock-kube", "k", updateMockKubeFormat, "Prints the new Kubernetes resource specs in the specified format instead of editing them [json|yaml]")
	updateOpsSightCtl.AddSpecFlags(updateOpsSightCmd, false)
	updateCmd.AddCommand(updateOpsSightCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightImageCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightExternalHostCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightAddRegistryCmd)
}
