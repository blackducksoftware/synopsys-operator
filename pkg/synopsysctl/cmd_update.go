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
	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource Ctl for edit
var updateBlackduckCtl ResourceCtl
var updateOpsSightCtl ResourceCtl
var updateAlertCtl ResourceCtl

// Update Defaults
var updateSynopsysOperatorImage = ""
var updatePrometheusImage = ""
var updateSecretAdminPassword = ""
var updateSecretPostgresPassword = ""
var updateSecretUserPassword = ""
var updateSecretBlackduckPassword = ""
var updateExposeUI = ""
var updateExposePrometheusMetrics = ""
var updateTerminationGracePeriodSeconds int64
var updateOperatorTimeBombInSeconds int64
var updateLogLevel = ""
var updateThreadiness int
var updatePostgresRestartInMins int64
var updatePodWaitTimeoutSeconds int64
var updateResyncIntervalInSeconds int64

// updateCmd provides functionality to update/upgrade features of
// Synopsys resources
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Synopsys Resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Must specify a sub-command")
	},
}

// updateOperatorCmd lets the user update the Synopsys-Operator
var updateOperatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Update the Synopsys-Operator",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check Number of Arguments
		if len(args) != 0 {
			return fmt.Errorf("this command takes 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, err := GetOperatorNamespace()
		if err != nil {
			log.Errorf("unable to find synopsys operator in the cluster because %+v", err)
			return nil
		}

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
			log.Debugf("updating synopsys operator image to %s", updateSynopsysOperatorImage)
			// check image tag
			imageHasTag := len(strings.Split(updateSynopsysOperatorImage, ":")) == 2
			if !imageHasTag {
				log.Errorf("synopsys operator's image does not have a tag: %s", updateSynopsysOperatorImage)
				return nil
			}
			newOperatorSpec.Image = updateSynopsysOperatorImage
		}
		if cmd.Flag("admin-password").Changed {
			log.Debugf("updating admin password")
			newOperatorSpec.AdminPassword = updateSecretAdminPassword
		}
		if cmd.Flag("postgres-password").Changed {
			log.Debugf("updating postgres password")
			newOperatorSpec.PostgresPassword = updateSecretPostgresPassword
		}
		if cmd.Flag("user-password").Changed {
			log.Debugf("updating user password")
			newOperatorSpec.UserPassword = updateSecretUserPassword
		}
		if cmd.Flag("blackduck-password").Changed {
			log.Debugf("updating blackduck password")
			newOperatorSpec.BlackduckPassword = updateSecretBlackduckPassword
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
			log.Errorf("unable to merge old and new synopsys operator info because %+v", err)
			return nil
		}

		// update synopsys operator
		err = newOperatorSpec.UpdateSynopsysOperator(restconfig, kubeClient, namespace, blackduckClient, opssightClient, alertClient, oldOperatorSpec)
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
		err = newPrometheusSpec.UpdatePrometheus()
		if err != nil {
			log.Errorf("unable to update Prometheus because %+v", err)
			return nil
		}

		routeClient, err := routeclient.NewForConfig(restconfig)
		// expose the routes
		if strings.ToUpper(updateExposeUI) == OPENSHIFT {
			if err != nil {
				log.Errorf("unable to create the route client due to %+v", err)
				return nil
			}
			_, err = operatorutil.GetOpenShiftRoutes(routeClient, namespace, "synopsys-operator-ui")
			if err != nil {
				log.Infof("creating openshift routes for the synopsys operator user interface")
				_, err = operatorutil.CreateOpenShiftRoutes(routeClient, namespace, "synopsys-operator-ui", "Service", "synopsys-operator", "synopsys-operator-ui", routev1.TLSTerminationEdge)
				if err != nil {
					log.Warnf("could not create route (possible reason: kubernetes doesn't support routes) due to %+v", err)
				}
			}
		}

		// expose the metrics routes
		if strings.ToUpper(updateExposePrometheusMetrics) == OPENSHIFT {
			if err != nil {
				log.Errorf("unable to create the route client due to %+v", err)
				return nil
			}
			_, err = operatorutil.GetOpenShiftRoutes(routeClient, namespace, "synopsys-operator-prometheus")
			if err != nil {
				log.Infof("creating openshift routes for the synopsys operator prometheus metrics")
				_, err = operatorutil.CreateOpenShiftRoutes(routeClient, namespace, "synopsys-operator-prometheus", "Service", "prometheus", "prometheus", routev1.TLSTerminationEdge)
				if err != nil {
					log.Warnf("could not create route (possible reason: kubernetes doesn't support routes) due to %+v", err)
				}
			}
		}
		log.Infof("successfully updated the synopsys operator in '%s' namespace", namespace)
		return nil
	},
}

// updateBlackduckCmd lets the user update a BlackDuck instance
var updateBlackduckCmd = &cobra.Command{
	Use:   "blackduck NAMESPACE",
	Short: "Update an instance of Blackduck",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		blackduckNamespace := args[0]
		log.Infof("updating BlackDuck %s...", blackduckNamespace)

		// Get the Blackuck
		currBlackduck, err := operatorutil.GetHub(blackduckClient, blackduckNamespace, blackduckNamespace)
		if err != nil {
			log.Errorf("error getting Blackduck: %s", err)
			return nil
		}
		// Check if it can be updated
		updateBlackduckCtl.SetSpec(currBlackduck.Spec)
		canUpdate, err := updateBlackduckCtl.CanUpdate()
		if err != nil {
			log.Errorf("cannot Update: %s", err)
			return nil
		}
		if canUpdate {
			// Make changes to Spec
			flagset := cmd.Flags()
			updateBlackduckCtl.SetChangedFlags(flagset)
			newSpec := updateBlackduckCtl.GetSpec().(blackduckapi.BlackduckSpec)
			// Create new Blackduck CRD
			newBlackduck := *currBlackduck //make copy
			newBlackduck.Spec = newSpec
			// Update Blackduck
			_, err = operatorutil.UpdateBlackduck(blackduckClient, newBlackduck.Spec.Namespace, &newBlackduck)
			if err != nil {
				log.Errorf("error updating the BlackDuck: %s", err)
				return nil
			}
			log.Infof("successfully updated BlackDuck: '%s'", blackduckNamespace)
		}
		return nil
	},
}

// updateBlackduckRootKeyCmd create new Black Duck root key for source code upload in the cluster
var updateBlackduckRootKeyCmd = &cobra.Command{
	Use:   "rootKey BLACK_DUCK_NAME NEW_SEAL_KEY MASTER_KEY_FILE_PATH",
	Short: "Update the root key of Black Duck for source code upload",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := args[0]
		newSealKey := args[1]
		filePath := args[2]

		log.Infof("updating BlackDuck %s Root Key...", namespace)

		_, err := operatorutil.GetHub(blackduckClient, metav1.NamespaceDefault, namespace)
		if err != nil {
			log.Errorf("unable to find Black Duck %s instance because %+v", namespace, err)
			return nil
		}
		operatorNamespace, err := GetOperatorNamespace()
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator instance because %+v", err)
			return nil
		}

		fileName := filepath.Join(filePath, fmt.Sprintf("%s.key", namespace))
		masterKey, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorf("error reading the master key from %s because %+v", fileName, err)
			return nil
		}

		// Filter the upload cache pod to get the root key using the seal key
		uploadCachePod, err := operatorutil.FilterPodByNamePrefixInNamespace(kubeClient, namespace, "uploadcache")
		if err != nil {
			log.Errorf("unable to filter the upload cache pod of %s because %+v", namespace, err)
			return nil
		}

		// Create the exec into kubernetes pod request
		req := operatorutil.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
		_, err = operatorutil.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://uploadcache:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
		if err != nil {
			log.Errorf("unable to exec into upload cache pod in %s because %+v", namespace, err)
			return nil
		}

		secret, err := operatorutil.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator blackduck-secret in %s namespace because %+v", operatorNamespace, err)
			return nil
		}
		secret.Data["SEAL_KEY"] = []byte(newSealKey)

		err = operatorutil.UpdateSecret(kubeClient, operatorNamespace, secret)
		if err != nil {
			log.Errorf("unable to update the Synopsys Operator blackduck-secret in %s namespace because %+v", operatorNamespace, err)
			return nil
		}
		log.Infof("successfully updated BlackDuck %s's Root Key", namespace)
		return nil
	},
}

// updateOpsSightCmd lets the user update an OpsSight instance
var updateOpsSightCmd = &cobra.Command{
	Use:   "opssight NAMESPACE",
	Short: "Update an instance of OpsSight",
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
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightNamespace, opsSightNamespace)
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
			_, err = operatorutil.UpdateOpsSight(opssightClient, newOpsSight.Spec.Namespace, &newOpsSight)
			if err != nil {
				log.Errorf("Error updating the OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight: '%s'", opsSightNamespace)
		}
		return nil
	},
}

// updateOpsSightImageCmd lets the user update an image in an OpsSight instance
var updateOpsSightImageCmd = &cobra.Command{
	Use:   "image NAMESPACE COMPONENT IMAGE",
	Short: "Update an image for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("this command takes 3 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
		componentName := args[1]
		componentImage := args[2]

		log.Infof("updating OpsSight %s's Image...", opsSightName)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
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
			// Update the Spec with new Image
			switch componentName {
			case "Perceptor":
				currOpsSight.Spec.Perceptor.Image = componentImage
			case "Scanner":
				currOpsSight.Spec.ScannerPod.Scanner.Image = componentImage
			case "ImageFacade":
				currOpsSight.Spec.ScannerPod.ImageFacade.Image = componentImage
			case "ImagePerceiver":
				currOpsSight.Spec.Perceiver.ImagePerceiver.Image = componentImage
			case "PodPerceiver":
				currOpsSight.Spec.Perceiver.PodPerceiver.Image = componentImage
			case "Skyfire":
				currOpsSight.Spec.Skyfire.Image = componentImage
			case "Prometheus":
				currOpsSight.Spec.Prometheus.Image = componentImage
			default:
				log.Errorf("%s is not a valid COMPONENT", componentName)
				log.Errorf("Valid Components: Perceptor, Scanner, ImageFacade, ImagePerceiver, PodPerceiver, Skyfire, Prometheus")
				return fmt.Errorf("invalid Component Name")
			}
			// Update OpsSight with New Image
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("Error updating the OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's Image", opsSightName)
		}
		return nil
	},
}

// updateOpsSightExternalHostCmd lets the user update an OpsSight with an External Host
var updateOpsSightExternalHostCmd = &cobra.Command{
	Use:   "externalHost NAMESPACE SCHEME DOMAIN PORT USER PASSWORD SCANLIMIT",
	Short: "Update an external host for a component of OpsSight",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 7 {
			return fmt.Errorf("this command takes 7 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
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

		log.Infof("adding External Host to OpsSight %s...", opsSightName)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
		if err != nil {
			log.Errorf("error getting the OpsSight: %s", err)
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
			// Add External Host to Spec
			newHost := opssightapi.Host{
				Scheme:              hostScheme,
				Domain:              hostDomain,
				Port:                int(hostPort),
				User:                hostUser,
				Password:            hostPassword,
				ConcurrentScanLimit: int(hostScanLimit),
			}
			currOpsSight.Spec.Blackduck.ExternalHosts = append(currOpsSight.Spec.Blackduck.ExternalHosts, &newHost)
			// Update OpsSight with External Host
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("error updating the OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's External Host", opsSightName)
		}
		return nil
	},
}

// updateOpsSightAddRegistryCmd lets the user update and OpsSight by
// adding a registry for the ImageFacade
var updateOpsSightAddRegistryCmd = &cobra.Command{
	Use:   "registry NAMESPACE URL USER PASSWORD",
	Short: "Add an Internal Registry to OpsSight's ImageFacade",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return fmt.Errorf("this command takes 4 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opsSightName := args[0]
		regURL := args[1]
		regUser := args[2]
		regPass := args[3]

		log.Infof("adding Internal Registry to OpsSight %s...", opsSightName)

		// Get OpsSight Spec
		currOpsSight, err := operatorutil.GetOpsSight(opssightClient, opsSightName, opsSightName)
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
			// Add Internal Registry to Spec
			newReg := opssightapi.RegistryAuth{
				URL:      regURL,
				User:     regUser,
				Password: regPass,
			}
			currOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries = append(currOpsSight.Spec.ScannerPod.ImageFacade.InternalRegistries, &newReg)
			// Update OpsSight with Internal Registry
			_, err = operatorutil.UpdateOpsSight(opssightClient, opsSightName, currOpsSight)
			if err != nil {
				log.Errorf("error adding Internal Registry with updating OpsSight: %s", err)
				return nil
			}
			log.Infof("successfully updated OpsSight %s's Registry", opsSightName)
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

		log.Infof("updating Alert %s...", alertNamespace)

		// Get the Alert
		currAlert, err := operatorutil.GetAlert(alertClient, alertNamespace, alertNamespace)
		if err != nil {
			log.Errorf("error Updaing Alert while getting the Alert: %s", err)
			return nil
		}
		updateAlertCtl.SetSpec(currAlert.Spec)

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
			// Create new Alert CRD
			newAlert := *currAlert //make copy
			newAlert.Spec = newSpec
			// Update Alert
			_, err = operatorutil.UpdateAlert(alertClient, newAlert.Spec.Namespace, &newAlert)
			if err != nil {
				log.Errorf("error in updating the alert due to %+v", err)
				return nil
			}
			log.Infof("successfully updated Alert: '%s'", alertNamespace)
		}
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackduckCtl = blackduck.NewBlackduckCtl()
	updateOpsSightCtl = opssight.NewOpsSightCtl()
	updateAlertCtl = alert.NewAlertCtl()

	rootCmd.AddCommand(updateCmd)

	// Add Operator Commands
	updateOperatorCmd.Flags().StringVarP(&updateExposeUI, "expose-ui", "e", updateExposeUI, "expose the synopsys operator's user interface. possible values are [NODEPORT/LOADBALANCER/OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&updateSynopsysOperatorImage, "synopsys-operator-image", "i", updateSynopsysOperatorImage, "synopsys operator image URL")
	updateOperatorCmd.Flags().StringVarP(&updateExposePrometheusMetrics, "expose-prometheus-metrics", "m", updateExposePrometheusMetrics, "expose the synopsys operator's prometheus metrics. possible values are [NODEPORT/LOADBALANCER/OPENSHIFT]")
	updateOperatorCmd.Flags().StringVarP(&updatePrometheusImage, "prometheus-image", "k", updatePrometheusImage, "prometheus image URL")
	updateOperatorCmd.Flags().StringVarP(&updateSecretAdminPassword, "admin-password", "a", updateSecretAdminPassword, "postgres admin password")
	updateOperatorCmd.Flags().StringVarP(&updateSecretPostgresPassword, "postgres-password", "p", updateSecretPostgresPassword, "postgres password")
	updateOperatorCmd.Flags().StringVarP(&updateSecretUserPassword, "user-password", "u", updateSecretUserPassword, "postgres user password")
	updateOperatorCmd.Flags().StringVarP(&updateSecretBlackduckPassword, "blackduck-password", "b", updateSecretBlackduckPassword, "blackduck password for 'sysadmin' account")
	updateOperatorCmd.Flags().Int64VarP(&updateOperatorTimeBombInSeconds, "operator-time-bomb-in-seconds", "o", updateOperatorTimeBombInSeconds, "termination grace period in seconds for shutting down crds")
	updateOperatorCmd.Flags().Int64VarP(&updatePostgresRestartInMins, "postgres-restart-in-minutes", "n", updatePostgresRestartInMins, "check for postgres restart in minutes")
	updateOperatorCmd.Flags().Int64VarP(&updatePodWaitTimeoutSeconds, "pod-wait-timeout-in-seconds", "w", updatePodWaitTimeoutSeconds, "wait for pod to be running in seconds")
	updateOperatorCmd.Flags().Int64VarP(&updateResyncIntervalInSeconds, "resync-interval-in-seconds", "r", updateResyncIntervalInSeconds, "custom resources resync time period in seconds")
	updateOperatorCmd.Flags().Int64VarP(&updateTerminationGracePeriodSeconds, "postgres-termination-grace-period", "g", updateTerminationGracePeriodSeconds, "termination grace period in seconds for shutting down postgres")
	updateOperatorCmd.Flags().StringVarP(&updateLogLevel, "log-level", "l", updateLogLevel, "log level of synopsys operator")
	updateOperatorCmd.Flags().IntVarP(&updateThreadiness, "no-of-threads", "c", updateThreadiness, "number of threads to process the custom resources")
	updateCmd.AddCommand(updateOperatorCmd)

	// Add Bladuck Commands
	updateBlackduckCtl.AddSpecFlags(updateBlackduckCmd, false)
	updateCmd.AddCommand(updateBlackduckCmd)
	updateBlackduckCmd.AddCommand(updateBlackduckRootKeyCmd)

	// Add OpsSight Commands
	updateOpsSightCtl.AddSpecFlags(updateOpsSightCmd, false)
	updateCmd.AddCommand(updateOpsSightCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightImageCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightExternalHostCmd)
	updateOpsSightCmd.AddCommand(updateOpsSightAddRegistryCmd)

	// Add Alert Commands
	updateAlertCtl.AddSpecFlags(updateAlertCmd, false)
	updateCmd.AddCommand(updateAlertCmd)
}
