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

package synopsysctl

import (
	"fmt"
	"strings"

	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func migrate(bd *v1.Blackduck, operatorNamespace string, flags *pflag.FlagSet) error {
	// TODO ensure operator is installed and running a recent version that doesn't require additional migration

	log.Info("stopping Synopsys Operator")
	soOperatorDeploy, err := util.GetDeployment(kubeClient, operatorNamespace, "synopsys-operator")
	if err != nil {
		return err
	}

	soOperatorDeploy, err = util.PatchDeploymentForReplicas(kubeClient, soOperatorDeploy, util.IntToInt32(0))
	if err != nil {
		return err
	}

	// Generate Helm configuration
	helmValuesMap, err := BlackduckV1ToHelm(bd, operatorNamespace)
	if err != nil {
		return err
	}

	helmValuesMapFromFlag, err := updateBlackDuckCobraHelper.GenerateHelmFlagsFromCobraFlags(flags)
	if err != nil {
		return err
	}

	if err := mergo.Merge(&helmValuesMap, helmValuesMapFromFlag, mergo.WithOverride); err != nil {
		return err
	}

	log.Info("deleting existing Black Duck resources")
	// TODO wait for resources to be deleted
	if err := deleteComponents(bd.Spec.Namespace, bd.Name, util.BlackDuckName); err != nil {
		return err
	}

	log.Info("upgrading Black Duck using Helm based deployment")
	// Update the Helm Chart Location

	chartLocationFlag := flags.Lookup("chart-location-path")
	if chartLocationFlag.Changed {
		blackduckChartRepository = chartLocationFlag.Value.String()
	} else {
		versionFlag := flags.Lookup("version")
		if versionFlag.Changed {
			blackduckChartRepository = fmt.Sprintf("%s/charts/blackduck-%s.tgz", baseChartRepository, versionFlag.Value.String())
		}
	}

	var extraFiles []string
	size, found := helmValuesMap["size"]
	if found {
		extraFiles = append(extraFiles, fmt.Sprintf("%s.yaml", size.(string)))
	}

	err = util.CreateWithHelm3(bd.Name, bd.Spec.Namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath, true, extraFiles...)
	if err != nil {
		return fmt.Errorf("failed to create Blackduck resources: %+v", err)
	}

	// Deploy Resources
	err = util.CreateWithHelm3(bd.Name, bd.Spec.Namespace, blackduckChartRepository, helmValuesMap, kubeConfigPath, false, extraFiles...)
	if err != nil {
		return fmt.Errorf("failed to create Blackduck resources: %+v", err)
	}

	log.Info("removing Black Duck custom resource")
	if err := util.DeleteBlackduck(blackDuckClient, bd.Name, bd.Namespace, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	log.Info("starting Synopsys Operator")
	if _, err := util.PatchDeploymentForReplicas(kubeClient, soOperatorDeploy, util.IntToInt32(1)); err != nil {
		return err
	}

	return nil
}

func isFeatureEnabled(environs []string, featureName string, expectedValue string) bool {
	for _, value := range environs {
		if strings.Contains(value, featureName) {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				mapValue := strings.ToLower(strings.TrimSpace(values[1]))
				if strings.EqualFold(mapValue, expectedValue) {
					return true
				}
			}
			return false
		}
	}
	return false
}

func BlackduckV1ToHelm(bd *v1.Blackduck, operatorNamespace string) (map[string]interface{}, error) {
	helmConfig := make(map[string]interface{})

	// Seal key
	if len(bd.Spec.SealKey) > 0 {
		decodedSealKey, err := util.Base64Decode(bd.Spec.SealKey)
		if err != nil {
			return nil, err
		}
		util.SetHelmValueInMap(helmConfig, []string{"sealKey"}, decodedSealKey)
	} else {
		// TODO handle ClusterScope
		sealSecret, err := kubeClient.CoreV1().Secrets(operatorNamespace).Get("blackduck-secret", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if _, ok := sealSecret.Data["SEAL_KEY"]; !ok {
			return nil, fmt.Errorf("couldn't find SEAL_KEY in %s/blackduck-secret", namespace)
		}
		util.SetHelmValueInMap(helmConfig, []string{"sealKey"}, string(sealSecret.Data["SEAL_KEY"]))
	}

	// Webserver
	webserverSecretName := util.GetResourceName(bd.Name, util.BlackDuckName, "webserver-certificate")
	var webserverSecret *corev1.Secret
	var err error
	if len(bd.Spec.Certificate) > 0 && len(bd.Spec.CertificateKey) > 0 {
		webserverSecret, err = blackduck.GetCertificateSecret(webserverSecretName, bd.Spec.Namespace, []byte(bd.Spec.Certificate), []byte(bd.Spec.CertificateKey))
		if err != nil {
			return nil, err
		}
	} else {
		currentSecret, err := kubeClient.CoreV1().Secrets(bd.Spec.Namespace).Get(util.GetResourceName(bd.Name, util.BlackDuckName, "webserver-certificate"), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		webserverSecret, err = blackduck.GetCertificateSecret(webserverSecretName, bd.Spec.Namespace, currentSecret.Data["WEBSERVER_CUSTOM_CERT_FILE"], currentSecret.Data["WEBSERVER_CUSTOM_KEY_FILE"])
		if err != nil {
			return nil, err
		}
	}
	if _, err := kubeClient.CoreV1().Secrets(bd.Spec.Namespace).Create(webserverSecret); err != nil && !k8serrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create secret: %+v", err)
	}
	util.SetHelmValueInMap(helmConfig, []string{"tlsCertSecretName"}, webserverSecretName)

	// Auth CA
	if len(bd.Spec.AuthCustomCA) > 0 {
		authSecretName := util.GetResourceName(bd.Name, util.BlackDuckName, "auth-custom-ca")
		authSecret, err := blackduck.GetAuthCertificateSecret(authSecretName, bd.Spec.Namespace, []byte(bd.Spec.AuthCustomCA))
		if err != nil {
			return nil, err
		}

		if _, err := kubeClient.CoreV1().Secrets(bd.Spec.Namespace).Create(authSecret); err != nil && !k8serrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create secret: %+v", err)
		}
		util.SetHelmValueInMap(helmConfig, []string{"certAuthCACertSecretName"}, authSecretName)
	}

	// Proxy Cert
	if len(bd.Spec.ProxyCertificate) > 0 {
		proxySecretName := util.GetResourceName(bd.Name, util.BlackDuckName, "proxy-certificate")
		proxySecret, err := blackduck.GetProxyCertificateSecret(proxySecretName, bd.Spec.Namespace, []byte(bd.Spec.ProxyCertificate))
		if err != nil {
			return nil, err
		}

		if _, err := kubeClient.CoreV1().Secrets(bd.Spec.Namespace).Create(proxySecret); err != nil && !k8serrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create secret: %+v", err)
		}
		util.SetHelmValueInMap(helmConfig, []string{"proxyCertSecretName"}, proxySecretName)
	}

	// Postgres
	if bd.Spec.ExternalPostgres != nil {
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "host"}, bd.Spec.ExternalPostgres.PostgresHost)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "port"}, bd.Spec.ExternalPostgres.PostgresPort)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "adminUserName"}, bd.Spec.ExternalPostgres.PostgresAdmin)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "userUserName"}, bd.Spec.ExternalPostgres.PostgresUser)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "ssl"}, bd.Spec.ExternalPostgres.PostgresUser)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "adminPassword"}, bd.Spec.ExternalPostgres.PostgresAdminPassword)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "userPassword"}, bd.Spec.ExternalPostgres.PostgresUserPassword)
	} else {
		adminPassword, err := util.Base64Decode(bd.Spec.AdminPassword)
		if err != nil {
			return nil, err
		}
		userPassword, err := util.Base64Decode(bd.Spec.UserPassword)
		if err != nil {
			return nil, err
		}
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "adminUserName"}, "blackduck")
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "userUserName"}, "blackduck_user")
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "adminPassword"}, adminPassword)
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "userPassword"}, userPassword)

		util.SetHelmValueInMap(helmConfig, []string{"postgres", "isExternal"}, false)
	}

	if len(bd.Spec.PVCStorageClass) > 0 {
		util.SetHelmValueInMap(helmConfig, []string{"storageClass"}, bd.Spec.PVCStorageClass)
	}
	util.SetHelmValueInMap(helmConfig, []string{"enablePersistentStorage"}, bd.Spec.PersistentStorage)
	util.SetHelmValueInMap(helmConfig, []string{"enableLivenessProbe"}, bd.Spec.LivenessProbes)

	if len(bd.Spec.DesiredState) > 0 {
		util.SetHelmValueInMap(helmConfig, []string{"status"}, bd.Spec.DesiredState)
	} else {
		util.SetHelmValueInMap(helmConfig, []string{"status"}, "Running")
	}

	if bd.Spec.RegistryConfiguration != nil {
		util.SetHelmValueInMap(helmConfig, []string{"registry"}, bd.Spec.RegistryConfiguration.Registry)
		util.SetHelmValueInMap(helmConfig, []string{"imagePullSecrets"}, bd.Spec.RegistryConfiguration.PullSecrets)
	}

	if isFeatureEnabled(bd.Spec.Environs, "ENABLE_SOURCE_UPLOADS", "true") {
		util.SetHelmValueInMap(helmConfig, []string{"enableSourceCodeUpload"}, true)
	}

	if isFeatureEnabled(bd.Spec.Environs, "USE_BINARY_UPLOADS", "1") {
		util.SetHelmValueInMap(helmConfig, []string{"enableBinaryScanner"}, true)
	}

	// NodeAffinities
	for k, v := range bd.Spec.NodeAffinities {
		util.SetHelmValueInMap(helmConfig, []string{k, "affinity"}, blackduck.OperatorAffinityTok8sAffinity(v))
	}

	//SecurityContexts
	for k, v := range bd.Spec.SecurityContexts {
		util.SetHelmValueInMap(helmConfig, []string{k, "securityContext"}, blackduck.OperatorSecurityContextTok8sAffinity(v))
	}

	// Environs
	for _, v := range bd.Spec.Environs {
		values := strings.SplitN(v, ":", 2)
		if len(values) == 2 {
			util.SetHelmValueInMap(helmConfig, []string{"environs", values[0]}, values[1])
		}
	}

	// Map existing PVCs
	if bd.Spec.PersistentStorage {
		util.SetHelmValueInMap(helmConfig, []string{"postgres", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-postgres", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"authentication", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-authentication", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"cfssl", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-cfssl", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"logstash", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-logstash", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"registration", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-registration", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"uploadcache", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-uploadcache-data", bd.Name))
		util.SetHelmValueInMap(helmConfig, []string{"webapp", "persistentVolumeClaimName"}, fmt.Sprintf("%s-blackduck-webapp", bd.Name))
	}

	util.SetHelmValueInMap(helmConfig, []string{"size"}, bd.Spec.Size)
	return helmConfig, nil
}

func deleteComponents(namespace string, name string, app string) error {
	deploy, err := kubeClient.AppsV1().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range deploy.Items {
		propagationPolicy := metav1.DeletePropagationBackground
		if err := kubeClient.AppsV1().Deployments(namespace).Delete(v.Name, &metav1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		}); err != nil {
			return err
		}
	}

	rc, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range rc.Items {
		propagationPolicy := metav1.DeletePropagationBackground
		if err := kubeClient.CoreV1().ReplicationControllers(namespace).Delete(v.Name, &metav1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		}); err != nil {
			return err
		}
	}

	svc, err := kubeClient.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range svc.Items {
		if strings.Contains(v.Name, "-exposed") {
			if err := kubeClient.CoreV1().Services(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}

	cm, err := kubeClient.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range cm.Items {
		if err := kubeClient.CoreV1().ConfigMaps(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	secret, err := kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s, component!=secret", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range secret.Items {
		if err := kubeClient.CoreV1().Secrets(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	serviceAccounts, err := kubeClient.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s, name=%s", app, name),
	})
	if err != nil {
		return err
	}
	for _, v := range serviceAccounts.Items {
		if err := kubeClient.CoreV1().ServiceAccounts(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}
