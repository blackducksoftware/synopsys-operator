/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package soperator

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater stores the configuration and clients to create specific versions of Synopsys Operator
type Creater struct {
	DryRun     bool
	KubeConfig *rest.Config
	KubeClient *kubernetes.Clientset
}

// NewCreater returns this Alert Creater
func NewCreater(dryRun bool, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset) *Creater {
	return &Creater{DryRun: dryRun, KubeConfig: kubeConfig, KubeClient: kubeClient}
}

//// GetComponents returns the resource components for an Alert
//func (sc *Creater) GetComponents(specConfig SpecConfig) (*api.ComponentList, error) {
//	return specConfig.GetComponents()
//}

// Versions is an Interface function that returns the versions supported by this Creater
func (sc *Creater) Versions() []string {
	return SOperatorCRDVersionMap.GetVersions()
}

//// EnsureSynopsysOperator updates the Synopsys Operator's Kubernetes componenets and changes
//// all CRDs to versions that the Operator can use
//func (sc *Creater) EnsureSynopsysOperator(namespace string, blackduckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset, alertClient *alertclientset.Clientset,
//	oldOperatorSpec *SpecConfig, newOperatorSpec *SpecConfig) error {
//
//	// Get CRD Version Data
//	newOperatorImageVersion, err := operatorutil.GetImageTag(newOperatorSpec.Image)
//	if err != nil {
//		return fmt.Errorf("failed to get version of the new Synopsys Operator image: %s", err)
//	}
//	oldOperatorImageVersion, err := operatorutil.GetImageTag(oldOperatorSpec.Image)
//	if err != nil {
//		return fmt.Errorf("failed to get version of the old Synopsys Operator image: %s", err)
//	}
//
//	newCrdData := SOperatorCRDVersionMap.GetCRDVersions(newOperatorImageVersion)
//	oldCrdData := SOperatorCRDVersionMap.GetCRDVersions(oldOperatorImageVersion)
//
//	// Get CRDs that need to be updated (specs have new version set)
//	log.Debugf("Getting CRDs that need new versions")
//	var oldBlackducks = []blackduckapi.Blackduck{}
//
//	apiExtensionClient, err := apiextensionsclient.NewForConfig(sc.KubeConfig)
//	if err != nil {
//		return fmt.Errorf("error creating the api extension client due to %+v", err)
//	}
//
//	if newCrdData.Blackduck.APIVersion != oldCrdData.Blackduck.APIVersion {
//		_, err := operatorutil.GetCustomResourceDefinition(apiExtensionClient, operatorutil.BlackDuckCRDName)
//		if err == nil {
//			oldBlackducks, err = GetBlackduckVersionsToRemove(blackduckClient, newCrdData.Blackduck.APIVersion)
//			if err != nil {
//				return fmt.Errorf("failed to get Blackduck's to update: %s", err)
//			}
//			err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.Blackduck.CRDName)
//			if err != nil {
//				return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.Blackduck.CRDName, err)
//			}
//			log.Debugf("updating %d Black Ducks", len(oldBlackducks))
//		}
//	}
//
//	var oldOpsSights = []opssightapi.OpsSight{}
//	if newCrdData.OpsSight.APIVersion != oldCrdData.OpsSight.APIVersion {
//		_, err := operatorutil.GetCustomResourceDefinition(apiExtensionClient, operatorutil.OpsSightCRDName)
//		if err == nil {
//			oldOpsSights, err = GetOpsSightVersionsToRemove(opssightClient, newCrdData.OpsSight.APIVersion)
//			if err != nil {
//				return fmt.Errorf("failed to get OpsSights to update: %s", err)
//			}
//			err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.OpsSight.CRDName)
//			if err != nil {
//				return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.OpsSight.CRDName, err)
//			}
//			log.Debugf("updating %d OpsSights", len(oldOpsSights))
//		}
//	}
//
//	var oldAlerts = []alertapi.Alert{}
//	if newCrdData.Alert.APIVersion != oldCrdData.Alert.APIVersion {
//		_, err := operatorutil.GetCustomResourceDefinition(apiExtensionClient, operatorutil.AlertCRDName)
//		if err == nil {
//			oldAlerts, err = GetAlertVersionsToRemove(alertClient, newCrdData.Alert.APIVersion)
//			if err != nil {
//				return fmt.Errorf("failed to get Alerts to update%s", err)
//			}
//			err = operatorutil.DeleteCustomResourceDefinition(apiExtensionClient, oldCrdData.Alert.CRDName)
//			if err != nil {
//				return fmt.Errorf("unable to delete the %s crd because %s", oldCrdData.Alert.CRDName, err)
//			}
//			log.Debugf("updating %d Alerts", len(oldAlerts))
//		}
//	}
//
//	// Update the Synopsys Operator's Components
//	log.Debugf("updating Synopsys Operator's Components")
//	newOperatorSpec.ClusterType = GetClusterType(sc.KubeClient)
//	err = sc.UpdateSOperatorComponents(newOperatorSpec)
//	if err != nil {
//		return fmt.Errorf("failed to update Synopsys Operator components: %s", err)
//	}
//
//	// Update the CRDs in the cluster with the new versions
//	// loop to wait for kuberentes to register new CRDs
//	log.Debugf("updating CRDs to new Versions")
//	for i := 1; i <= 10; i++ {
//		if err = operatorutil.UpdateBlackducks(blackduckClient, oldBlackducks); err == nil {
//			break
//		}
//		if i >= 10 {
//			return fmt.Errorf("failed to update Black Ducks: %s", err)
//		}
//		log.Debugf("attempt %d to update Black Ducks", i)
//		time.Sleep(1 * time.Second)
//	}
//	for i := 1; i <= 10; i++ {
//		if err = operatorutil.UpdateOpsSights(opssightClient, oldOpsSights); err == nil {
//			break
//		}
//		if i >= 10 {
//			return fmt.Errorf("failed to update OpsSights: %s", err)
//		}
//		log.Debugf("attempt %d to update OpsSights", i)
//		time.Sleep(1 * time.Second)
//	}
//	for i := 1; i <= 10; i++ {
//		if err = operatorutil.UpdateAlerts(alertClient, oldAlerts); err == nil {
//			break
//		}
//		if i >= 10 {
//			return fmt.Errorf("failed to update Alerts: %s", err)
//		}
//		log.Debugf("attempt %d to update Alerts", i)
//		time.Sleep(1 * time.Second)
//	}
//
//	return nil
//}

// UpdateSOperatorComponents updates Kubernetes resources for the Synopsys Operator
func (sc *Creater) UpdateSOperatorComponents(specConfig *SpecConfig) error {
	//sOperatorComponents, err := specConfig.GetComponents()
	//if err != nil {
	//	return fmt.Errorf("failed to get Synopsys Operator components: %s", err)
	//}
	//sOperatorCommonConfig := crdupdater.NewCRUDComponents(sc.KubeConfig, sc.KubeClient, false, false, specConfig.Namespace, "", sOperatorComponents, "app=synopsys-operator,component=operator", true)
	//_, errs := sOperatorCommonConfig.CRUDComponents()
	//if errs != nil {
	//	return fmt.Errorf("failed to update Synopsys Operator components: %+v", errs)
	//}

	// TODO find a cleaner way to create / update the operator deployment now that we no longer use CRUD updater
	if _, err := sc.KubeClient.CoreV1().ServiceAccounts(specConfig.Namespace).Create(specConfig.getOperatorServiceAccount()); err != nil {
		return err
	}

	if specConfig.IsClusterScoped {
		if _, err := sc.KubeClient.RbacV1beta1().ClusterRoles().Create(specConfig.getOperatorClusterRole()); err != nil {
			return err
		}

		if _, err := sc.KubeClient.RbacV1beta1().ClusterRoleBindings().Create(specConfig.getOperatorClusterRoleBinding()); err != nil {
			return err
		}
	} else {
		if _, err := sc.KubeClient.RbacV1beta1().Roles(specConfig.Namespace).Create(specConfig.getOperatorRole()); err != nil {
			return err
		}

		if _, err := sc.KubeClient.RbacV1beta1().RoleBindings(specConfig.Namespace).Create(specConfig.getOperatorRoleBinding()); err != nil {
			return err
		}
	}

	if _, err := sc.KubeClient.CoreV1().Secrets(specConfig.Namespace).Create(specConfig.getOperatorSecret()); err != nil {
		return err
	}

	if _, err := sc.KubeClient.CoreV1().Secrets(specConfig.Namespace).Create(specConfig.getTLSCertificateSecret()); err != nil {
		return err
	}

	operatorCm, err := specConfig.GetOperatorConfigMap()
	if err != nil {
		return err
	}
	if _, err := sc.KubeClient.CoreV1().ConfigMaps(specConfig.Namespace).Create(operatorCm); err != nil {
		return err
	}

	operatorDeployment, err := specConfig.getOperatorDeployment()
	if err != nil {
		return err
	}

	if _, err := sc.KubeClient.AppsV1().Deployments(specConfig.Namespace).Create(operatorDeployment); err != nil {
		return err
	}

	for _, service := range specConfig.getOperatorService() {
		if _, err := sc.KubeClient.CoreV1().Services(specConfig.Namespace).Create(service); err != nil {
			return err
		}
	}

	// TODO deploy prometheus

	return nil
}

// UpdatePrometheus updates Kubernetes resources for Prometheus
//func (sc *Creater) UpdatePrometheus(specConfig *PrometheusSpecConfig) error {
//	prometheusComponents, err := specConfig.GetComponents()
//	if err != nil {
//		return fmt.Errorf("failed to get Prometheus components: %s", err)
//	}
//	prometheusCommonConfig := crdupdater.NewCRUDComponents(sc.KubeConfig, sc.KubeClient, false, false, specConfig.Namespace, "", prometheusComponents, "app=synopsys-operator,component=prometheus", true)
//	_, errs := prometheusCommonConfig.CRUDComponents()
//	if errs != nil {
//		return fmt.Errorf("failed to update Prometheus components: %+v", errs)
//	}
//	return nil
//}
