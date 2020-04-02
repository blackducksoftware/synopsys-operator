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
	"sync"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	alert "github.com/blackducksoftware/synopsys-operator/pkg/alert"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/bdba"
	polarisreporting "github.com/blackducksoftware/synopsys-operator/pkg/polaris-reporting"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	// bdappsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"

	blackduck "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	opssight "github.com/blackducksoftware/synopsys-operator/pkg/opssight"
	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update Command ResourceCtlSpecBuilders
var updateAlertCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updateBlackDuckCobraHelper blackduck.HelmValuesFromCobraFlags
var updateOpsSightCobraHelper CRSpecBuilderFromCobraFlagsInterface
var updatePolarisCobraHelper polaris.HelmValuesFromCobraFlags
var updatePolarisReportingCobraHelper polarisreporting.HelmValuesFromCobraFlags
var updateBDBACobraHelper bdba.HelmValuesFromCobraFlags

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

// updateBlackDuckCmd updates a Black Duck instance
var updateBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME",
	Example:       "synopsyctl update blackduck <name> -n <namespace> --size medium",
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
		// Update the Helm Chart Location
		chartLocationFlag := cmd.Flag("chart-location-path")
		if chartLocationFlag.Changed {
			blackduckChartRepository = chartLocationFlag.Value.String()
		} else {
			versionFlag := cmd.Flag("version")
			if versionFlag.Changed {
				blackduckChartRepository = fmt.Sprintf("https://artifactory.internal.synopsys.com/artifactory/bds-hub-helm-snapshot-local/blackduck/blackduck-%s.tgz", versionFlag.Value.String())
			}
		}
		// TODO verity we can download the chart
		isOperatorBased := false
		instance, err := util.GetWithHelm3(args[0], namespace, kubeConfigPath)
		if err != nil {
			isOperatorBased = true
		}

		if !isOperatorBased && instance != nil {
			helmValuesMap, err := updateBlackDuckCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
			if err != nil {
				return err
			}

			secrets, err := blackduck.GetCertsFromFlagsAndSetHelmValue(args[0], namespace, cmd.Flags(), helmValuesMap)
			if err != nil {
				return err
			}
			for _, v := range secrets {
				if _, err := kubeClient.CoreV1().Secrets(namespace).Create(&v); err != nil {
					if k8serrors.IsAlreadyExists(err) {
						if _, err := kubeClient.CoreV1().Secrets(namespace).Update(&v); err != nil {
							return fmt.Errorf("failed to update certificate secret: %+v", err)
						}
					} else {
						return fmt.Errorf("failed to create certificate secret: %+v", err)
					}
				}
			}

			var extraFiles []string
			size, found := instance.Config["size"]
			if found {
				extraFiles = append(extraFiles, fmt.Sprintf("%s.yaml", size.(string)))
			}

			updateBlackDuckCobraHelper.SetArgs(instance.Config)
			if err := util.UpdateWithHelm3(args[0], namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath, extraFiles...); err != nil {
				return err
			}
		} else if isOperatorBased {
			if !cmd.Flag("version").Changed { // TODO fill in the blackduck version
				return fmt.Errorf("you must upgrade this Blackduck version with --version to use this synopsysctl binary - modifying Alert versions before XXXXX are not supported with this binary")
			}
			ok, err := util.IsVersionGreaterThanOrEqualTo(cmd.Flag("version").Value.String(), 2019, time.April, 0)
			if err != nil {
				return err
			}

			if !ok {
				return fmt.Errorf("migration is only suported for version 2019.4.0 and above")
			}

			operatorNamespace := namespace
			isClusterScoped := util.GetClusterScope(apiExtensionClient)
			if isClusterScoped {
				opNamespace, err := util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
				if err != nil {
					return err
				}
				if len(opNamespace) > 1 {
					return fmt.Errorf("more than 1 Synopsys Operator found in your cluster")
				}
				operatorNamespace = opNamespace[0]
			}

			blackDuckName := args[0]
			crdNamespace := namespace
			if isClusterScoped {
				crdNamespace = metav1.NamespaceAll
			}

			currBlackDuck, err := util.GetBlackduck(blackDuckClient, crdNamespace, blackDuckName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, crdNamespace, err)
			}
			if err := migrate(currBlackDuck, operatorNamespace, cmd.Flags()); err != nil {
				// TODO restart operator if migration failed?
				return err
			}
		}

		return nil
	},
}

// setBlackDuckFileOwnershipJob that sets the Owner of the files
func setBlackDuckFileOwnershipJob(namespace string, name string, pvcName string, ownership int64, wg *sync.WaitGroup) error {
	busyBoxImage := defaultBusyBoxImage
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
		_, err := util.GetWithHelm3(args[0], namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("couldn't find instance %s in namespace %s", args[0], namespace)
		}
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

	_, err = util.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -X PUT --header "X-SEAL-KEY:%s" -H "X-MASTER-KEY:%s" https://localhost:9444/api/internal/recovery --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(newSealKey)), masterKey)})
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
		helmValuesMap := make(map[string]interface{})
		util.SetHelmValueInMap(helmValuesMap, []string{"sealKey"}, newSealKey)
		if err := util.UpdateWithHelm3(name, namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath); err != nil {
			return err
		}

		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'. Wait for upload cache pod to restart to resume the source code upload", name, namespace)
	}
	return nil
}

// updateBlackDuckAddEnvironCmd adds an Environment Variable to a Black Duck instance
var updateBlackDuckAddEnvironCmd = &cobra.Command{
	Use:           "addenviron BLACK_DUCK_NAME (ENVIRON_NAME:ENVIRON_VALUE)",
	Example:       "synopsysctl update blackduck addenviron <name> USE_ALERT:1\nsynopsysctl update blackduck addenviron <name> USE_ALERT:1 -n <namespace>",
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
		_, err := util.GetWithHelm3(args[0], namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("couldn't find instance %s in namespace %s", args[0], namespace)
		}

		vals := strings.Split(args[1], ":")
		if len(vals) != 2 {
			return fmt.Errorf("%s is not valid - expecting NAME:VALUE", args[0])
		}
		log.Infof("updating Black Duck '%s' with environ '%s' in namespace '%s'...", args[0], args[1], namespace)

		helmValuesMap := make(map[string]interface{})
		util.SetHelmValueInMap(helmValuesMap, []string{"environs", vals[0]}, vals[1])

		if err := util.UpdateWithHelm3(args[0], namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath); err != nil {
			return err
		}

		log.Infof("successfully submitted updates to Black Duck '%s' in namespace '%s'", args[0], namespace)
		return nil
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

// updatePolarisCmd updates a Polaris instance
var updatePolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsyctl update polaris -n <namespace>",
	Short:         "Update a Polaris instance. (Please make sure you have read and understand prerequisites before installing Polaris: https://sig-confluence.internal.synopsys.com/display/DD/Polaris+on-premises])",
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
		helmRelease, err := util.GetWithHelm3(polarisName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to get previous user defined values: %+v", err)
		}
		updatePolarisCobraHelper.SetArgs(helmRelease.Config)
		// Get the flags to set Helm values
		helmValuesMap, err := updatePolarisCobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
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

		// Deploy Polaris Resources
		err = util.UpdateWithHelm3(polarisName, namespace, polarisChartRepository, helmValuesMap, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to update Polaris resources due to %+v", err)
		}

		log.Infof("Polaris has been successfully Updated in namespace '%s'!", namespace)
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
		helmRelease, err := util.GetWithHelm3(polarisReportingName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to get previous user defined values: %+v", err)
		}
		updatePolarisReportingCobraHelper.SetArgs(helmRelease.Config)

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
				polarisReportingChartRepository = fmt.Sprintf("%s/charts/polaris-helmchart-reporting-%s.tgz", baseChartRepository, versionFlag.Value.String())
			}
		}

		// Update Polaris-Reporting Resources
		err = util.UpdateWithHelm3(polarisReportingName, namespace, polarisReportingChartRepository, helmValuesMap, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to update Polaris-Reporting resources due to %+v", err)
		}

		log.Infof("Polaris-Reporting has been successfully Updated in namespace '%s'!", namespace)
		return nil
	},
}

// updateBDBACmd updates a BDBA instance
var updateBDBACmd = &cobra.Command{
	Use:           "bdba",
	Example:       "",
	Short:         "Update a BDBA instance",
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
		helmRelease, err := util.GetWithHelm3(bdbaName, namespace, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to get previous user defined values: %+v", err)
		}
		updateBDBACobraHelper.SetArgs(helmRelease.Config)

		// Get the flags to set Helm values
		helmValuesMap, err := updateBDBACobraHelper.GenerateHelmFlagsFromCobraFlags(cmd.Flags())
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

		// Update Resources
		err = util.UpdateWithHelm3(bdbaName, namespace, bdbaChartRepository, helmValuesMap, kubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to update BDBA resources due to %+v", err)
		}

		log.Infof("BDBA has been successfully Updated in namespace '%s'!", namespace)
		return nil
	},
}

func init() {
	// initialize global resource ctl structs for commands to use
	updateBlackDuckCobraHelper = *blackduck.NewHelmValuesFromCobraFlags()
	updateOpsSightCobraHelper = opssight.NewCRSpecBuilderFromCobraFlags()
	updateAlertCobraHelper = alert.NewCRSpecBuilderFromCobraFlags()
	updatePolarisCobraHelper = *polaris.NewHelmValuesFromCobraFlags()
	updatePolarisReportingCobraHelper = *polarisreporting.NewHelmValuesFromCobraFlags()
	updateBDBACobraHelper = *bdba.NewHelmValuesFromCobraFlags()

	rootCmd.AddCommand(updateCmd)

	// updateAlertCmd
	updateAlertCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	updateAlertCobraHelper.AddCRSpecFlagsToCommand(updateAlertCmd, false)
	addMockFlag(updateAlertCmd)
	updateCmd.AddCommand(updateAlertCmd)

	/* Update Black Duck Comamnds */

	// updateBlackDuckCmd
	updateBlackDuckCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(updateBlackDuckCmd.Flags(), "namespace")
	addChartLocationPathFlag(updateBlackDuckCmd)
	updateBlackDuckCobraHelper.AddCRSpecFlagsToCommand(updateBlackDuckCmd, false)
	updateCmd.AddCommand(updateBlackDuckCmd)

	// updateBlackDuckMasterKeyCmd
	updateBlackDuckCmd.AddCommand(updateBlackDuckMasterKeyCmd)

	// updateBlackDuckMasterKeyNativeCmd
	updateBlackDuckMasterKeyCmd.AddCommand(updateBlackDuckMasterKeyNativeCmd)

	// updateBlackDuckAddEnvironCmd
	updateBlackDuckCmd.AddCommand(updateBlackDuckAddEnvironCmd)

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
	cobra.MarkFlagRequired(updatePolarisCmd.PersistentFlags(), "namespace")
	updatePolarisCobraHelper.AddCobraFlagsToCommand(updatePolarisCmd, false)
	addChartLocationPathFlag(updatePolarisCmd)
	updateCmd.AddCommand(updatePolarisCmd)

	// Polaris-Reporting
	updatePolarisReportingCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(updatePolarisReportingCmd.PersistentFlags(), "namespace")
	updatePolarisReportingCobraHelper.AddCobraFlagsToCommand(updatePolarisReportingCmd, false)
	addChartLocationPathFlag(updatePolarisReportingCmd)
	updateCmd.AddCommand(updatePolarisReportingCmd)

	// BDBA
	updateBDBACmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	cobra.MarkFlagRequired(updateBDBACmd.PersistentFlags(), "namespace")
	updateBDBACobraHelper.AddCobraFlagsToCommand(updateBDBACmd, false)
	addChartLocationPathFlag(updateBDBACmd)
	updateCmd.AddCommand(updateBDBACmd)
}
