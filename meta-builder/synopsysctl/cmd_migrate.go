///*
// * Copyright (C) 2019 Synopsys, Inc.
// *
// *  Licensed to the Apache Software Foundation (ASF) under one
// * or more contributor license agreements. See the NOTICE file
// * distributed with this work for additional information
// * regarding copyright ownership. The ASF licenses this file
// * to you under the Apache License, Version 2.0 (the
// * "License"); you may not use this file except in compliance
// *  with the License. You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on an
// * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// * KIND, either express or implied. See the License for the
// * specific language governing permissions and limitations
// *  under the License.
// */
//
package synopsysctl

//import (
//	"encoding/json"
//	"fmt"
//	"github.com/blackducksoftware/synopsys-operator/meta-builder/utils"
//	"github.com/blackducksoftware/synopsys-operator/meta-builder/soperator"
//
//	"strings"
//	"time"
//
//
//	log "github.com/sirupsen/logrus"
//	"github.com/spf13/cobra"
//	batchv1 "k8s.io/api/batch/v1"
//	corev1 "k8s.io/api/core/v1"
//	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//var busyBoxImage = defaultBusyBoxImage
//
//// migrateCmd migrates a resource before upgrading Synopsys Operator
//var migrateCmd = &cobra.Command{
//	Use:           "migrate",
//	Example:       "synopsysctl migrate <from>\nsynopsysctl migrate 2019.4.2\nsynopsysctl migrate <from> -n <namespace>",
//	Short:         "Migrate a Synopsys resource before upgrading the operator",
//	SilenceUsage:  true,
//	SilenceErrors: true,
//	Args: func(cmd *cobra.Command, args []string) error {
//		// Check number of arguments
//		if len(args) != 1 {
//			cmd.Help()
//			return fmt.Errorf("this command takes 1 argument")
//		}
//		// validate Synopsys Operator image
//		if _, err := utils.ValidateImageString(synopsysOperatorImage); err != nil {
//			return err
//		}
//		// validate busy box image
//		if _, err := utils.ValidateImageString(busyBoxImage); err != nil {
//			return err
//		}
//		return nil
//	},
//	RunE: func(cmd *cobra.Command, args []string) error {
//		// Check if a namespace was provided, else determine the namespace from the cluster
//		namespaceToMigrate := ""
//		if len(namespace) > 0 {
//			namespaceToMigrate = namespace
//		} else {
//			isClusterScoped := utils.GetClusterScope(apiExtensionClient)
//			if isClusterScoped {
//				namespaces, err := utils.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
//				if err != nil {
//					return err
//				}
//				if len(namespaces) > 1 {
//					return fmt.Errorf("more than 1 Synopsys Operator found in your cluster. please pass the namespace of the Synopsys Operator that you want to migrate")
//				}
//				namespaceToMigrate = namespaces[0]
//			} else {
//				return fmt.Errorf("namespace of Synopsys Operator must be provided in namespace scoped mode")
//			}
//		}
//		// Migrate the CRDs
//		err := migrate(namespaceToMigrate)
//		if err != nil {
//			return err
//		}
//
//		// Update the Operator Image
//		currOperatorSpec, err := soperator.GetOldOperatorSpec(restconfig, kubeClient, namespaceToMigrate) // Get current Synopsys Operator Spec
//		if err != nil {
//			return err
//		}
//		newOperatorSpec := *currOperatorSpec          // Make copy
//		newOperatorSpec.Image = synopsysOperatorImage // Set new image
//		sOperatorCreater := soperator.NewCreater(false, restconfig, kubeClient)
//		err = sOperatorCreater.EnsureSynopsysOperator(namespaceToMigrate, blackDuckClient, opsSightClient, alertClient, currOperatorSpec, &newOperatorSpec) // this will scale up the deployment
//		if err != nil {
//			return fmt.Errorf("unable to update Synopsys Operator due to %+v", err)
//		}
//		return nil
//	},
//}
//
//func migrate(namespace string) error {
//	err := scaleDownDeployment(namespace, util.OperatorName)
//	if err != nil {
//		return err
//	}
//	err = migrateSize(namespace)
//	if err != nil {
//		return err
//	}
//	err = migrateCRD(namespace)
//	if err != nil {
//		return err
//	}
//	err = migrateOperator(namespace)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func scaleDownDeployment(namespace string, name string) error {
//	log.Infof("scaling down %s deployment in namespace '%s'", name, namespace)
//	deployment, err := util.GetDeployment(kubeClient, namespace, name)
//	if err != nil {
//		return fmt.Errorf("unable to find the %s deployment in namespace '%s' due to %+v", name, namespace, err)
//	}
//	replicas := util.IntToInt32(0)
//	_, err = util.PatchDeploymentForReplicas(kubeClient, deployment, replicas)
//	if err != nil {
//		return fmt.Errorf("unable to scale down the %s deployment in namespace '%s' due to %+v", name, namespace, err)
//	}
//	log.Infof("successfully scaled down %s deployment in namespace '%s'", name, namespace)
//	return nil
//}
//
//func scaleDownRC(namespace string, name string) error {
//	log.Infof("scaling down %s replication controller in namespace '%s'", name, namespace)
//	rc, err := util.GetReplicationController(kubeClient, namespace, name)
//	if err != nil {
//		return fmt.Errorf("unable to find the %s replication controller in namespace '%s' due to %+v", name, namespace, err)
//	}
//	replicas := util.IntToInt32(0)
//	_, err = util.PatchReplicationControllerForReplicas(kubeClient, rc, replicas)
//	if err != nil {
//		return fmt.Errorf("unable to scale down the %s replication controller in namespace '%s' due to %+v", name, namespace, err)
//	}
//	log.Infof("successfully scaled down %s replication controller in namespace '%s'", name, namespace)
//	return nil
//}
//
//// migrateOperator adds CRDNames and IsClusterScope parameter to Synopsys Operator config map
//func migrateOperator(namespace string) error {
//	log.Infof("migrating Synopsys Operator resources in namespace '%s'", namespace)
//
//	isClusterScoped = util.GetClusterScope(apiExtensionClient)
//	// list the existing CRD's and convert them to map with both key and value as name
//	var crdList *apiextensions.CustomResourceDefinitionList
//	crdList, err := util.ListCustomResourceDefinitions(apiExtensionClient, "app=synopsys-operator")
//	if err != nil {
//		return fmt.Errorf("unable to list Custom Resource Definitions due to %+v", err)
//	}
//
//	ns, err := util.GetNamespace(kubeClient, namespace)
//	if err != nil {
//		return fmt.Errorf("unable to find Synopsys Operator in namespace %s due to %+v", namespace, err)
//	}
//
//	if _, ok := ns.Labels["owner"]; !ok {
//		ns.Labels = util.InitLabels(ns.Labels)
//		ns.Labels["owner"] = util.OperatorName
//		_, err = util.UpdateNamespace(kubeClient, ns)
//		if err != nil {
//			return fmt.Errorf("unable to update Synopsys Operator in namespace %s due to %+v", namespace, err)
//		}
//	}
//
//	cm, err := util.GetConfigMap(kubeClient, namespace, util.OperatorName)
//	if err != nil {
//		return fmt.Errorf("error getting the Synopsys Operator config map in namespace %s due to %+v", namespace, err)
//	}
//	data := cm.Data["config.json"]
//	var cmData map[string]interface{}
//	err = json.Unmarshal([]byte(data), &cmData)
//	if err != nil {
//		log.Errorf("unable to unmarshal config map data due to %+v", err)
//	}
//	crds := make([]string, 0)
//	if _, ok := cmData["CrdNames"]; !ok {
//		for _, crd := range crdList.Items {
//			crds = append(crds, crd.Name)
//		}
//		cmData["CrdNames"] = strings.Join(crds, ",")
//		cmData["IsClusterScoped"] = isClusterScoped
//	}
//
//	if val, ok := cm.Data["Expose"]; (ok && len(string(val)) == 0) || !ok {
//		cmData["Expose"] = util.NONE
//	}
//
//	bytes, err := json.Marshal(cmData)
//	if err != nil {
//		return fmt.Errorf("unable to marshal config map data due to %+v", err)
//	}
//
//	cm.Data["config.json"] = string(bytes)
//
//	_, err = util.UpdateConfigMap(kubeClient, namespace, cm)
//	if err != nil {
//		return fmt.Errorf("unable to update the Synopsys Operator config map in namespace %s due to %+v", namespace, err)
//	}
//
//	cm, err = util.GetConfigMap(kubeClient, namespace, "prometheus")
//	if err != nil {
//		return fmt.Errorf("error getting the Prometheus config map in namespace %s due to %+v", namespace, err)
//	}
//	isUpdated := false
//	if val, ok := cm.Data["Expose"]; (ok && len(val) == 0) || !ok {
//		cm.Data["Expose"] = util.NONE
//		isUpdated = true
//	}
//
//	if isUpdated {
//		_, err = util.UpdateConfigMap(kubeClient, namespace, cm)
//		if err != nil {
//			return fmt.Errorf("unable to update the Synopsys Operator config map in namespace %s due to %+v", namespace, err)
//		}
//	}
//
//	log.Infof("successfully migrated Synopsys Operator resources in namespace '%s'", namespace)
//
//	return nil
//}
//
//// migrateCRD adds the labels to the custom resource definitions for the existing operator
//func migrateCRD(namespace string) error {
//	crdNames := []string{util.AlertCRDName, util.BlackDuckCRDName, util.OpsSightCRDName, util.SizeCRDName}
//	for _, crdName := range crdNames {
//		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdName)
//		if err != nil {
//			log.Errorf("error getting %s custom resource defintion due to %+v", crdName, err)
//			continue
//		}
//
//		// if crd labels doesn't contain app, then updates
//		if _, ok := crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)]; !ok {
//			crd.Labels = util.InitLabels(crd.Labels)
//			crd.Labels["app"] = "synopsys-operator"
//			crd.Labels["component"] = "operator"
//			crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)] = namespace
//			_, err = util.UpdateCustomResourceDefinition(apiExtensionClient, crd)
//			if err != nil {
//				return fmt.Errorf("unable to update %s custom resource defintion due to %+v", crdName, err)
//			}
//		}
//		log.Infof("successfully migrated '%s' custom resource definition", crd.GetName())
//		err = migrateCR(namespace, crdName)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// migrateCR add the labels to the existing custom resource instances
//func migrateCR(namespace string, crdName string) error {
//	switch crdName {
//	case util.AlertCRDName:
//		err := migrateAlert(namespace)
//		if err != nil {
//			return err
//		}
//	case util.BlackDuckCRDName:
//		err := migrateBlackDuck(namespace)
//		if err != nil {
//			return err
//		}
//	case util.OpsSightCRDName:
//		err := migrateOpsSight(namespace)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// migrateAlert migrates the existing Alert instances
//func migrateAlert(namespace string) error {
//	alerts, err := util.ListAlerts(alertClient, namespace)
//	if err != nil {
//		return fmt.Errorf("failed to list Alert instances in namespace '%s' due to %+v", namespace, err)
//	}
//	for _, alert := range alerts.Items {
//		alertName := alert.Name
//		alertNamespace := alert.Spec.Namespace
//		log.Infof("migrating Alert '%s' in namespace '%s'...", alertName, alertNamespace)
//
//		// update annotations
//		if _, ok := alert.Annotations["synopsys.com/created.by"]; !ok {
//			alert.Annotations = util.InitAnnotations(alert.Annotations)
//			alert.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
//			if len(alert.Spec.ExposeService) == 0 {
//				alert.Spec.ExposeService = util.NONE
//			}
//			_, err := alertClient.SynopsysV1().Alerts(alertNamespace).Update(&alert)
//			if err != nil {
//				return fmt.Errorf("error migrating Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
//			}
//		}
//
//		// add synopsys labels to namespace
//		err = addNamespaceLabels(alertNamespace, alertName, util.AlertName, alert.Spec.Version)
//		if err != nil {
//			return err
//		}
//
//		// include name in all resources
//		err = addNameLabels(alertNamespace, alertName, util.AlertName)
//		if err != nil {
//			return err
//		}
//		log.Infof("successfully migrated Alert '%s' in namespace '%s'", alertName, alertNamespace)
//	}
//	return nil
//}
//
//// migrateBlackDuck migrates the existing Black Duck instances
//func migrateBlackDuck(namespace string) error {
//	blackDucks, err := util.ListBlackDucks(blackDuckClient, namespace)
//	if err != nil {
//		return fmt.Errorf("failed to list Black Duck instances in namespace '%s' due to %+v", namespace, err)
//	}
//	for _, blackDuck := range blackDucks.Items {
//		blackDuckName := blackDuck.Name
//		blackDuckNamespace := blackDuck.Spec.Namespace
//		log.Infof("migrating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
//
//		// ASSUMING ALL PASSWORDS HAVE REMAINED THE SAME, no need to pull from secret
//		defaultPassword := util.Base64Encode([]byte("blackduck"))
//
//		blackDuck.Spec.AdminPassword = defaultPassword
//		blackDuck.Spec.UserPassword = defaultPassword
//		blackDuck.Spec.PostgresPassword = defaultPassword
//
//		// update annotations
//		if _, ok := blackDuck.Annotations["synopsys.com/created.by"]; !ok {
//			blackDuck.Annotations = util.InitAnnotations(blackDuck.Annotations)
//			blackDuck.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
//			if len(blackDuck.Spec.ExposeService) == 0 {
//				blackDuck.Spec.ExposeService = util.NONE
//			}
//			_, err := blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Update(&blackDuck)
//			if err != nil {
//				return fmt.Errorf("error migrating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
//			}
//		}
//
//		// add synopsys labels to namespace
//		err = addNamespaceLabels(blackDuckNamespace, blackDuckName, util.BlackDuckName, blackDuck.Spec.Version)
//		if err != nil {
//			return err
//		}
//
//		// include name in all resources
//		err = addNameLabels(blackDuckNamespace, blackDuckName, util.BlackDuckName)
//		if err != nil {
//			return err
//		}
//
//		if blackDuck.Spec.PersistentStorage {
//			var rabbitmqRCName, zookeeperRCName, uploadCacheRCName, uploadCacheKeyPVCName, uploadCacheDataPVCName string
//			if value, ok := blackDuck.Annotations["synopsys.com/created.by"]; ok && "pre-2019.6.0" == value {
//				rabbitmqRCName = "rabbitmq"
//				zookeeperRCName = "zookeeper"
//				uploadCacheRCName = "uploadcache"
//				uploadCacheKeyPVCName = "blackduck-uploadcache-key"
//				uploadCacheDataPVCName = "blackduck-uploadcache-data"
//			} else {
//				rabbitmqRCName = utils.GetResourceName(blackDuckName, util.BlackDuckName, "rabbitmq")
//				zookeeperRCName = utils.GetResourceName(blackDuckName, util.BlackDuckName, "zookeeper")
//				uploadCacheRCName = utils.GetResourceName(blackDuckName, util.BlackDuckName, "uploadcache")
//				uploadCacheKeyPVCName = fmt.Sprintf("%s-blackduck-uploadcache-key", blackDuckName)
//				uploadCacheDataPVCName = fmt.Sprintf("%s-blackduck-uploadcache-data", blackDuckName)
//			}
//			// scale down zookeeper
//			err = scaleDownRC(blackDuckNamespace, zookeeperRCName)
//			if err != nil {
//				return err
//			}
//			// scale down upload cache
//			err = scaleDownRC(blackDuckNamespace, uploadCacheRCName)
//			if err != nil {
//				return err
//			}
//
//			if isSourceCodeEnabled(blackDuck.Spec.Environs) {
//				err = migrateUploadCachePVCJob(blackDuckNamespace, blackDuckName, uploadCacheKeyPVCName, uploadCacheDataPVCName)
//				if err != nil {
//					return err
//				}
//			}
//
//			pvcs := []string{"blackduck-rabbitmq"}
//			if value, ok := blackDuck.Annotations["synopsys.com/created.by"]; ok && "pre-2019.6.0" == value {
//				pvcs = append(pvcs, "zookeeper-data", "zookeeper-datalog")
//			} else {
//				pvcs = append(pvcs, utils.GetResourceName(blackDuckName, util.BlackDuckName, "zookeeper-data"), utils.GetResourceName(blackDuckName, util.BlackDuckName, "zookeeper-datalog"))
//			}
//			for _, pvc := range pvcs {
//				// check for an existance of PVC
//				_, err := util.GetPVC(kubeClient, blackDuckNamespace, pvc)
//				if err == nil {
//					if "blackduck-rabbitmq" == pvc {
//						// scale down rabbitmq
//						err = scaleDownRC(blackDuckNamespace, rabbitmqRCName)
//						if err != nil {
//							return err
//						}
//					}
//					err = removePVC(blackDuckNamespace, pvc)
//					if err != nil {
//						return err
//					}
//				}
//			}
//		}
//		log.Infof("successfully migrated Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
//	}
//	return nil
//}
//
//func isSourceCodeEnabled(environs []string) bool {
//	for _, value := range environs {
//		if strings.Contains(value, "ENABLE_SOURCE_UPLOADS") {
//			values := strings.SplitN(value, ":", 2)
//			if len(values) == 2 {
//				mapValue := strings.ToLower(strings.TrimSpace(values[1]))
//				if strings.EqualFold(mapValue, "true") {
//					return true
//				}
//			}
//			return false
//		}
//	}
//	return false
//}
//
//// removePVC removes the PVC
//func removePVC(namespace string, name string) error {
//	log.Infof("removing %s PVC from namespace '%s'", name, namespace)
//	err := util.DeletePVC(kubeClient, namespace, name)
//	if err == nil {
//		log.Infof("removed %s PVC successfully from namespace '%s'", name, namespace)
//	}
//	return err
//}
//
//// migrateUploadCachePVCJob create a Kube job to migrate the upload cache key data to upload cache data PVC
//func migrateUploadCachePVCJob(namespace string, name string, uploadCacheKeyVolumeName string, uploadCacheDataVolumeName string) error {
//	log.Infof("migrating upload cache key persistent volume to upload cache data persistent volume for Black Duck %s in namespace '%s'", name, namespace)
//	uploadCacheKeyVolume := components.NewPVCVolume(horizonapi.PVCVolumeConfig{PVCName: uploadCacheKeyVolumeName, VolumeName: "dir-uploadcache-key"})
//	uploadCacheDataVolume := components.NewPVCVolume(horizonapi.PVCVolumeConfig{PVCName: uploadCacheDataVolumeName, VolumeName: "dir-uploadcache-data"})
//	migrateJob := &batchv1.Job{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "migrate-upload-cache",
//		},
//		Spec: batchv1.JobSpec{
//			Template: corev1.PodTemplateSpec{
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{
//						{
//							Name:    "migrate-key",
//							Image:   busyBoxImage,
//							Command: []string{"sh", "-c", "mkdir -p /opt/blackduck/hub/blackduck-upload-cache/keys && mkdir -p /opt/blackduck/hub/blackduck-upload-cache/uploads/bdio && mkdir -p /opt/blackduck/hub/blackduck-upload-cache/uploads/sources && chmod 775 /opt/blackduck/hub/blackduck-upload-cache/keys && chmod 775 /opt/blackduck/hub/blackduck-upload-cache/uploads/bdio && chmod 775 /opt/blackduck/hub/blackduck-upload-cache/uploads/sources && if [ ! \"$(ls -A /opt/blackduck/hub/blackduck-upload-cache/keys)\" ]; then cp -pr /tmp/keys/MASTER_KEY_ENCRYPTED /opt/blackduck/hub/blackduck-upload-cache/keys; cp -pr /tmp/keys/MASTER_KEY_HASHED /opt/blackduck/hub/blackduck-upload-cache/keys; fi && if [ ! \"$(ls -A /opt/blackduck/hub/blackduck-upload-cache/uploads/bdio)\" ] && [ -d /opt/blackduck/hub/blackduck-upload-cache/bdio ]; then cp -pr /opt/blackduck/hub/blackduck-upload-cache/bdio /opt/blackduck/hub/blackduck-upload-cache/uploads; fi && if [ ! \"$(ls -A /opt/blackduck/hub/blackduck-upload-cache/uploads/sources)\" ] && [ -d /opt/blackduck/hub/blackduck-upload-cache/sources ]; then cp -pr /opt/blackduck/hub/blackduck-upload-cache/sources /opt/blackduck/hub/blackduck-upload-cache/uploads; fi && if [ -d /opt/blackduck/hub/blackduck-upload-cache/sources ]; then rm -rf /opt/blackduck/hub/blackduck-upload-cache/sources; fi && if [ -d /opt/blackduck/hub/blackduck-upload-cache/bdio ]; then rm -rf /opt/blackduck/hub/blackduck-upload-cache/bdio; fi"},
//							VolumeMounts: []corev1.VolumeMount{
//								{Name: "dir-uploadcache-key", MountPath: "/tmp/keys"},
//								{Name: "dir-uploadcache-data", MountPath: "/opt/blackduck/hub/blackduck-upload-cache"},
//							},
//						},
//					},
//					RestartPolicy: corev1.RestartPolicyNever,
//					Volumes: []corev1.Volume{
//						{Name: uploadCacheKeyVolume.Name, VolumeSource: uploadCacheKeyVolume.VolumeSource},
//						{Name: uploadCacheDataVolume.Name, VolumeSource: uploadCacheDataVolume.VolumeSource},
//					},
//				},
//			},
//		},
//	}
//
//	job, err := kubeClient.BatchV1().Jobs(namespace).Create(migrateJob)
//	if err != nil {
//		return err
//	}
//
//	timeout := time.NewTimer(30 * time.Minute)
//	ticker := time.NewTicker(10 * time.Second)
//	defer ticker.Stop()
//	defer timeout.Stop()
//
//	for {
//		select {
//		case <-timeout.C:
//			return fmt.Errorf("the migration of upload cache key to data is timed out for Black Duck %s in namespace '%s'", name, namespace)
//
//		case <-ticker.C:
//			job, err = kubeClient.BatchV1().Jobs(job.Namespace).Get(job.Name, metav1.GetOptions{})
//			if err != nil {
//				return err
//			}
//			if job.Status.Succeeded > 0 {
//				log.Infof("successfully migrated upload cache key persistent volume to upload cache data persistent volume for Black Duck %s in namespace '%s'", name, namespace)
//				kubeClient.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{})
//				return nil
//			}
//		}
//	}
//}
//
//// migrateOpsSight migrates the existing OpsSight instances
//func migrateOpsSight(namespace string) error {
//	opsSights, err := util.ListOpsSights(opsSightClient, namespace)
//	if err != nil {
//		return fmt.Errorf("failed to list OpsSight instances in namespace '%s' due to %+v", namespace, err)
//	}
//	for _, opsSight := range opsSights.Items {
//		opsSightName := opsSight.Name
//		opsSightNamespace := opsSight.Spec.Namespace
//		log.Infof("migrating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
//
//		// set the desired state to empty which is equivalent to START
//		opsSight.Spec.DesiredState = ""
//
//		// ASSUMING ALL PASSWORDS HAVE REMAINED THE SAME, no need to pull from secret
//		defaultPassword := util.Base64Encode([]byte("blackduck"))
//		opsSight.Spec.Blackduck.BlackduckPassword = defaultPassword
//		opsSight.Spec.Blackduck.BlackduckSpec.AdminPassword = defaultPassword
//		opsSight.Spec.Blackduck.BlackduckSpec.UserPassword = defaultPassword
//		opsSight.Spec.Blackduck.BlackduckSpec.PostgresPassword = defaultPassword
//		if strings.HasPrefix(opsSight.Spec.Perceptor.Name, "opssight-") {
//			opsSight.Spec.Perceptor.Name = "core"
//			opsSight.Spec.ScannerPod.Name = "scanner"
//			opsSight.Spec.ScannerPod.Scanner.Name = "scanner"
//			opsSight.Spec.ScannerPod.ImageFacade.Name = "image-getter"
//			opsSight.Spec.ScannerPod.ImageFacade.ServiceAccount = "scanner"
//			opsSight.Spec.Perceiver.PodPerceiver.Name = "pod-processor"
//			opsSight.Spec.Perceiver.ImagePerceiver.Name = "image-processor"
//			opsSight.Spec.Perceiver.ServiceAccount = "processor"
//		}
//
//		// update annotations
//		if _, ok := opsSight.Annotations["synopsys.com/created.by"]; !ok {
//			opsSight.Annotations = util.InitAnnotations(opsSight.Annotations)
//			opsSight.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
//			if len(opsSight.Spec.Perceptor.Expose) == 0 {
//				opsSight.Spec.Perceptor.Expose = util.NONE
//			}
//			if len(opsSight.Spec.Prometheus.Expose) == 0 {
//				opsSight.Spec.Prometheus.Expose = util.NONE
//			}
//			_, err := opsSightClient.SynopsysV1().OpsSights(opsSightNamespace).Update(&opsSight)
//			if err != nil {
//				return fmt.Errorf("error migrating OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
//			}
//		}
//
//		// add synopsys labels to namespace
//		err = addNamespaceLabels(opsSightNamespace, opsSightName, util.OpsSightName, "2.2.3")
//		if err != nil {
//			return err
//		}
//
//		// include name in all resources
//		err = addNameLabels(opsSightNamespace, opsSightName, util.OpsSightName)
//		if err != nil {
//			return err
//		}
//		log.Infof("successfully migrated OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
//	}
//	return nil
//}
//
//// migrateSize will ensure that the Size CRD exists as well as the default sizes
//func migrateSize(namespace string) error {
//	isClusterScoped := util.GetClusterScope(apiExtensionClient)
//	if _, err := util.GetCustomResourceDefinition(apiExtensionClient, util.SizeCRDName); err != nil {
//		crd, err := getCrdConfigs(namespace, isClusterScoped, []string{util.SizeCRDName})
//		if err != nil {
//			return err
//		}
//		if err := deployCrds(namespace, isClusterScoped, crd); err != nil {
//			return err
//		}
//		if err := util.WaitForCRD(util.SizeCRDName, time.Second, time.Minute*3, apiExtensionClient); err != nil {
//			return err
//		}
//	}
//
//	for _, v := range size.GetAllDefaultSizes() {
//		if _, err := sizeClient.SynopsysV1().Sizes(namespace).Get(v.Name, metav1.GetOptions{}); err != nil {
//			if _, err := sizeClient.SynopsysV1().Sizes(namespace).Create(v); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//// addNamespaceLabels adds a synopsys label to the namespace label
//func addNamespaceLabels(namespace string, name string, appName string, version string) error {
//	ns, err := util.GetNamespace(kubeClient, namespace)
//	if err != nil {
//		return fmt.Errorf("error getting %s namespace due to %+v", namespace, err)
//	}
//
//	// update labels in namespace
//	if _, ok := ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", appName, name)]; !ok {
//		ns.Labels = util.InitLabels(ns.Labels)
//		ns.Labels["owner"] = util.OperatorName
//		ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", appName, name)] = version
//		_, err = util.UpdateNamespace(kubeClient, ns)
//		if err != nil {
//			return fmt.Errorf("error updating %s namespace due to %+v", namespace, err)
//		}
//	}
//	return nil
//}
//
//// addNameLabels adds a name label to all its resources
//func addNameLabels(namespace string, name string, appName string) error {
//	deployments, err := util.ListDeployments(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list deployments for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, deployment := range deployments.Items {
//		if _, ok := deployment.Labels["name"]; !ok || appName == util.OpsSightName {
//			deployment.Labels = util.InitLabels(deployment.Labels)
//			deployment.Labels["name"] = name
//			deployment.Spec.Template.Labels = util.InitLabels(deployment.Spec.Template.Labels)
//			deployment.Spec.Template.Labels["name"] = name
//			_, err = util.UpdateDeployment(kubeClient, namespace, &deployment)
//			if err != nil {
//				return fmt.Errorf("unable to update %s deployment in namespace %s due to %+v", deployment.GetName(), namespace, err)
//			}
//		}
//	}
//
//	rcs, err := util.ListReplicationControllers(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list replication controllers for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, rc := range rcs.Items {
//		if _, ok := rc.Labels["name"]; !ok || appName == util.OpsSightName {
//			rc.Labels = util.InitLabels(rc.Labels)
//			rc.Labels["name"] = name
//			rc.Spec.Template.Labels = util.InitLabels(rc.Spec.Template.Labels)
//			rc.Spec.Template.Labels["name"] = name
//			if appName == util.OpsSightName {
//				rc.Spec.Selector = util.InitLabels(rc.Spec.Selector)
//				rc.Spec.Selector["name"] = name
//			}
//			_, err = util.UpdateReplicationController(kubeClient, namespace, &rc)
//			if err != nil {
//				return fmt.Errorf("unable to update %s replication controller in namespace %s due to %+v", rc.GetName(), namespace, err)
//			}
//		}
//	}
//
//	// delete pods
//	if appName == util.OpsSightName {
//		pods, err := util.ListPodsWithLabels(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//		if err != nil {
//			return fmt.Errorf("unable to list pods for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//		}
//
//		for _, pod := range pods.Items {
//			err = util.DeletePod(kubeClient, namespace, pod.GetName())
//			if err != nil {
//				return fmt.Errorf("unable to delete pod %s in namespace %s due to %+v", pod.GetName(), namespace, err)
//			}
//		}
//	}
//
//	services, err := util.ListServices(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list services for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, svc := range services.Items {
//		if _, ok := svc.Labels["name"]; !ok || appName == util.OpsSightName {
//			svc.Labels = util.InitLabels(svc.Labels)
//			svc.Labels["name"] = name
//			_, err = util.UpdateService(kubeClient, namespace, &svc)
//			if err != nil {
//				return fmt.Errorf("unable to update %s service in namespace %s due to %+v", svc.GetName(), namespace, err)
//			}
//		}
//	}
//
//	configmaps, err := util.ListConfigMaps(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list config maps for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, cm := range configmaps.Items {
//		if _, ok := cm.Labels["name"]; !ok || appName == util.OpsSightName {
//			cm.Labels = util.InitLabels(cm.Labels)
//			cm.Labels["name"] = name
//			_, err = util.UpdateConfigMap(kubeClient, namespace, &cm)
//			if err != nil {
//				return fmt.Errorf("unable to update %s config map in namespace %s due to %+v", cm.GetName(), namespace, err)
//			}
//		}
//	}
//
//	secrets, err := util.ListSecrets(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list secrets for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, secret := range secrets.Items {
//		if _, ok := secret.Labels["name"]; !ok || appName == util.OpsSightName {
//			secret.Labels = util.InitLabels(secret.Labels)
//			secret.Labels["name"] = name
//			_, err = util.UpdateSecret(kubeClient, namespace, &secret)
//			if err != nil {
//				return fmt.Errorf("unable to update %s secret in namespace %s due to %+v", secret.GetName(), namespace, err)
//			}
//		}
//	}
//
//	serviceAccounts, err := util.ListServiceAccounts(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list service accounts for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, serviceAccount := range serviceAccounts.Items {
//		if _, ok := serviceAccount.Labels["name"]; !ok || appName == util.OpsSightName {
//			serviceAccount.Labels = util.InitLabels(serviceAccount.Labels)
//			serviceAccount.Labels["name"] = name
//			_, err = util.UpdateServiceAccount(kubeClient, namespace, &serviceAccount)
//			if err != nil {
//				return fmt.Errorf("unable to update %s service account in namespace %s due to %+v", serviceAccount.GetName(), namespace, err)
//			}
//		}
//	}
//
//	clusterRoles, err := util.ListClusterRoles(kubeClient, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list cluster role for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, clusterRole := range clusterRoles.Items {
//		if _, ok := clusterRole.Labels["name"]; !ok || appName == util.OpsSightName {
//			clusterRole.Labels = util.InitLabels(clusterRole.Labels)
//			clusterRole.Labels["name"] = name
//			_, err = util.UpdateClusterRole(kubeClient, &clusterRole)
//			if err != nil {
//				return fmt.Errorf("unable to update %s cluster role due to %+v", clusterRole.GetName(), err)
//			}
//		}
//	}
//
//	clusterRoleBindings, err := util.ListClusterRoleBindings(kubeClient, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list cluster role bindings for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, crb := range clusterRoleBindings.Items {
//		if _, ok := crb.Labels["name"]; !ok || appName == util.OpsSightName {
//			crb.Labels = util.InitLabels(crb.Labels)
//			crb.Labels["name"] = name
//			_, err = util.UpdateClusterRoleBinding(kubeClient, &crb)
//			if err != nil {
//				return fmt.Errorf("unable to update %s cluster role bindings due to %+v", crb.GetName(), err)
//			}
//		}
//	}
//
//	pvcs, err := util.ListPVCs(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
//	if err != nil {
//		return fmt.Errorf("unable to list persistent volume claims for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//	}
//
//	for _, pvc := range pvcs.Items {
//		if _, ok := pvc.Labels["name"]; !ok {
//			pvc.Labels = util.InitLabels(pvc.Labels)
//			pvc.Labels["name"] = name
//			_, err = util.UpdatePVC(kubeClient, namespace, &pvc)
//			if err != nil {
//				return fmt.Errorf("unable to update %s persistent volume claim in namespace %s due to %+v", pvc.GetName(), namespace, err)
//			}
//		}
//	}
//
//	if util.IsOpenshift(kubeClient) {
//		routeClient := util.GetRouteClient(restconfig)
//		if routeClient != nil {
//			routes, err := util.ListRoutes(routeClient, namespace, fmt.Sprintf("app=%s", appName))
//			if err != nil {
//				return fmt.Errorf("unable to list routes for %s %s in namespace %s due to %+v", appName, name, namespace, err)
//			}
//
//			for _, route := range routes.Items {
//				if _, ok := route.Labels["name"]; !ok || appName == util.OpsSightName {
//					route.Labels = util.InitLabels(route.Labels)
//					route.Labels["name"] = name
//					_, err = util.UpdateRoute(routeClient, namespace, &route)
//					if err != nil {
//						return fmt.Errorf("unable to update %s route in namespace %s due to %+v", route.GetName(), namespace, err)
//					}
//				}
//			}
//		}
//	}
//
//	return nil
//}
//
//// migrateCleanupCmd cleanup the unused resources
//var migrateCleanupCmd = &cobra.Command{
//	Use:           "cleanup",
//	Example:       "synopsysctl migrate cleanup <from>\nsynopsysctl migrate cleanup 2019.4.2\nsynopsysctl migrate cleanup <from> -n <namespace>",
//	Short:         "Cleanup any unused resources after a Synopsys Operator migration. This should only be done after the user has verified full functionality. This can not be undone",
//	SilenceUsage:  true,
//	SilenceErrors: true,
//	Args: func(cmd *cobra.Command, args []string) error {
//		// Check number of arguments
//		if len(args) != 1 {
//			cmd.Help()
//			return fmt.Errorf("this command takes 1 argument")
//		}
//		return nil
//	},
//	RunE: func(cmd *cobra.Command, args []string) error {
//		if len(namespace) > 0 {
//			return cleanup(namespace)
//		}
//
//		// get operator namespace
//		isClusterScoped := util.GetClusterScope(apiExtensionClient)
//		if isClusterScoped {
//			namespaces, err := util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
//			if err != nil {
//				return err
//			}
//
//			if len(namespaces) > 1 {
//				return fmt.Errorf("more than 1 Synopsys Operator found in your cluster. please pass the namespace of the Synopsys Operator that you want to cleanup")
//			}
//			return cleanup(namespaces[0])
//		}
//		return fmt.Errorf("namespace of the Synopsys Operator need to be provided")
//	},
//}
//
//// cleanup will cleanup the resources
//func cleanup(namespace string) error {
//	blackDucks, err := util.ListBlackDucks(blackDuckClient, namespace)
//	if err != nil {
//		return fmt.Errorf("failed to list Black Duck instances in namespace '%s' due to %+v", namespace, err)
//	}
//	for _, blackDuck := range blackDucks.Items {
//		blackDuckName := blackDuck.Name
//		blackDuckNamespace := blackDuck.Spec.Namespace
//		log.Infof("cleaning up the Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
//		if blackDuck.Spec.PersistentStorage {
//			pods, err := util.ListPodsWithLabels(kubeClient, namespace, "job-name=migrate-upload-cache")
//			if err != nil {
//				return fmt.Errorf("failed to find the upload-cache migration pod in namespace '%s' due to %+v", namespace, err)
//			}
//
//			for _, pod := range pods.Items {
//				err = util.DeletePod(kubeClient, namespace, pod.Name)
//				if err != nil {
//					return fmt.Errorf("unable to delete pod %s in namespace '%s' due to %+v", pod.Name, namespace, err)
//				}
//			}
//
//			var uploadCacheKeyPVCName string
//			if value, ok := blackDuck.Annotations["synopsys.com/created.by"]; ok && "pre-2019.6.0" == value {
//				uploadCacheKeyPVCName = "blackduck-uploadcache-key"
//			} else {
//				uploadCacheKeyPVCName = fmt.Sprintf("%s-blackduck-uploadcache-key", blackDuckName)
//			}
//			// check for an existance of PVC
//			_, err = util.GetPVC(kubeClient, blackDuckNamespace, uploadCacheKeyPVCName)
//			if err == nil {
//				err = removePVC(blackDuckNamespace, uploadCacheKeyPVCName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//		log.Infof("successfully cleaned up the Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
//	}
//	return nil
//}
//
//func init() {
//	// Add Migrate Commands
//	rootCmd.AddCommand(migrateCmd)
//	migrateCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
//	migrateCmd.Flags().StringVarP(&busyBoxImage, "busybox-image", "b", busyBoxImage, "Image URL of Busybox")
//	migrateCmd.Flags().StringVarP(&synopsysOperatorImage, "update-image", "i", synopsysOperatorImage, "Image to migrate the Synopsys Operator instance to")
//	// Add Migrate Cleanup command to Migrate command
//	migrateCmd.AddCommand(migrateCleanupCmd)
//	migrateCleanupCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
//}
