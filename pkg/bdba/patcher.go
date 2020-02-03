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

package bdba

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func getPatchedRuntimeObjects(yamlManifests string, bdba BDBA) (map[string]runtime.Object, error) {
	yamlManifests = patchYamlManifestPlaceHolders(yamlManifests, bdba)

	// Convert the yaml manifests to Kubernetes Runtime Objects
	mapOfUniqueIDToRuntimeObject, err := util.ConvertYamlFileToRuntimeObjects(yamlManifests)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml manifests to runtime objects: %+v", err)
	}

	patcher := RuntimeObjectPatcher{
		bdba:                         bdba,
		mapOfUniqueIDToRuntimeObject: mapOfUniqueIDToRuntimeObject,
	}

	rtoMap, err := patcher.patch()
	if err != nil {
		return nil, fmt.Errorf("failed to path runtime objects: %+v", err)
	}
	return rtoMap, nil
}

func quoteInt(i int) string {
	return "\"" + strconv.Itoa(i) + "\""
}

func quoteBool(i bool) string {
	return "\"" + strconv.FormatBool(i) + "\""
}

func patchYamlManifestPlaceHolders(yamlManifests string, bdba BDBA) string {
	// Patch the yaml file yamlManifests
	yamlManifests = strings.ReplaceAll(yamlManifests, "${NAME}", bdba.Name)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${NAMESPACE}", bdba.Namespace)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${VERSION}", bdba.Version)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${ROOTURL}", bdba.RootURL)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${INGRESS_HOST}", bdba.IngressHost)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${RABBITMQ_K8S_DOMAIN}", bdba.RabbitMQK8SDomain)

	// Storage
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGSTORAGECLASS}", bdba.PSQLStorageClass)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_STORAGECLASS}", bdba.MinioStorageClass)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${RABBITMQ_STORAGECLASS}", bdba.RabbitMQStorageClass)

	// Web frontend configuration
	yamlManifests = strings.ReplaceAll(yamlManifests, "${SESSION_COOKIE_AGE}", quoteInt(bdba.SessionCookieAge))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${FRONTEND_REPLICAS}", strconv.Itoa(bdba.FrontendReplicas))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${HIDE_LICENSES}", quoteBool(bdba.HideLicenses))

	// SMTP configuration
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_ENABLED}", quoteBool(bdba.EmailEnabled))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_HOST}", bdba.EmailSMTPHost)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_PORT}", quoteInt(bdba.EmailSMTPPort))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_HOST_USER}", bdba.EmailSMTPUser)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_HOST_PASSWORD}", util.EncodeStringToBase64(bdba.EmailSMTPPassword))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_FROM}", bdba.EmailFrom)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_SECURITY}", bdba.EmailSecurity)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EMAIL_VERIFY_CERTIFICATE}", quoteBool(bdba.EmailVerify))

	// LDAP
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_ENABLED}", quoteBool(bdba.LDAPEnabled))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_SERVER_URI}", bdba.LDAPServerURI)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_USER_DN_TEMPLATE}", bdba.LDAPUserDNTemplate)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_BIND_AS_AUTHENTICATING_USER}", quoteBool(bdba.LDAPBindAsAuthenticating))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_BIND_DN}", bdba.LDAPBindDN)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_BIND_PASSWORD}", util.EncodeStringToBase64(bdba.LDAPBindPassword))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_START_TLS}", quoteBool(bdba.LDAPStartTLS))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_VERIFY_CERTIFICATE}", quoteBool(bdba.LDAPVerify))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_ROOT_CA_SECRET}", bdba.LDAPRootCASecret)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_ROOT_CERTIFICATE}", bdba.LDAPRootCAFile)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_REQUIRE_GROUP}", bdba.LDAPRequireGroup)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_USER_SEARCH}", bdba.LDAPUserSearch)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_USER_SEARCH_SCOPE}", bdba.LDAPUserSearchScope)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_GROUP_SEARCH}", bdba.LDAPGroupSearch)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_GROUP_SEARCH_SCOPE}", bdba.LDAPGroupSearchScope)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LDAP_NESTED_SEARCH}", quoteBool(bdba.LDAPNestedSearch))

	// Licensing
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LICENSING_USERNAME}", util.EncodeStringToBase64(bdba.LicensingUsername))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LICENSING_PASSWORD}", util.EncodeStringToBase64(bdba.LicensingPassword))

	// Worker scaling
	yamlManifests = strings.ReplaceAll(yamlManifests, "${WORKER_REPLICAS}", strconv.Itoa(bdba.WorkerReplicas))

	// Networking and security
	yamlManifests = strings.ReplaceAll(yamlManifests, "${HTTP_PROXY}", bdba.HTTPProxy)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${HTTP_NOPROXY}", bdba.HTTPNoProxy)

	// Ingress
	yamlManifests = strings.ReplaceAll(yamlManifests, "${INGRESS_HOST}", bdba.IngressHost)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${TLS_SECRET}", bdba.IngressTLSSecretName)

	yamlManifests = strings.ReplaceAll(yamlManifests, "nginx.ingress.kubernetes.io/proxy-request-buffering: false", "nginx.ingress.kubernetes.io/proxy-request-buffering: \"off\"")
	yamlManifests = strings.ReplaceAll(yamlManifests, "${ADMIN_EMAIL}", bdba.AdminEmail)

	brokerURL := "amqp://bdba:%s@" + bdba.Name + "-rabbitmq"
	yamlManifests = strings.ReplaceAll(yamlManifests, "${BROKER_URL}", brokerURL)

	// External PG
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGHOST}", bdba.PGHost)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGPORT}", bdba.PGPort)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGUSER}", bdba.PGUser)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGDATABASE}", bdba.PGDataBase)

	// Secrets
	yamlManifests = strings.ReplaceAll(yamlManifests, "${DJANGO_SECRET_KEY}", util.EncodeStringToBase64(bdba.DjangoSecretKey))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGPASSWORD}", util.EncodeStringToBase64(bdba.PGPassword))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_ACCESS_KEY_PLAIN}", bdba.MinioAccessKey)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_ACCESS_KEY}", util.EncodeStringToBase64(bdba.MinioAccessKey))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_SECRET_KEY}", util.EncodeStringToBase64(bdba.MinioSecretKey))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${BROKER_PASSWORD}", util.EncodeStringToBase64(bdba.RabbitMQPassword))
	yamlManifests = strings.ReplaceAll(yamlManifests, "${RABBITMQ_ERLANG_COOKIE}", util.EncodeStringToBase64(bdba.RabbitMQErlangCookie))

	return yamlManifests
}

// RuntimeObjectPatcher holds the BDBA run time objects and it is having methods to patch it
type RuntimeObjectPatcher struct {
	bdba                         BDBA
	mapOfUniqueIDToRuntimeObject map[string]runtime.Object
}

func setSecret(secret *corev1.Secret, secretKeyName string, secretKey string) {
	secret.Data[secretKeyName] = []byte(secretKey)
}

func (p *RuntimeObjectPatcher) patch() (map[string]runtime.Object, error) {
	patches := []func() error{
		p.patchNamespace,
		p.patchTLS,
		p.patchLDAP,
		p.patchExternalPG,
	}
	for _, patchFunc := range patches {
		err := patchFunc()
		if err != nil {
			return nil, err
		}
	}
	return p.mapOfUniqueIDToRuntimeObject, nil
}

// patchNamespace will change the resource namespace
func (p *RuntimeObjectPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIDToRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.bdba.Namespace)
	}
	return nil
}

// patchTLS will remove TLS if it's not enabled
func (p *RuntimeObjectPatcher) patchTLS() error {
	var tlsIngress = "true"

	if p.bdba.IngressTLSEnabled != true {
		log.Debug("Patching TLS")
		tlsIngress = "false"
		for k, v := range p.mapOfUniqueIDToRuntimeObject {
			switch v.(type) {
			case *extv1beta1.Ingress:
				p.mapOfUniqueIDToRuntimeObject[k].(*extv1beta1.Ingress).Spec.TLS = nil
			}
		}
	}

	for k, v := range p.mapOfUniqueIDToRuntimeObject {
		switch v.(type) {
		case *corev1.ConfigMap:
			var configMap = p.mapOfUniqueIDToRuntimeObject[k].(*corev1.ConfigMap)
			if strings.Contains(configMap.Name, "bdba-user-configmap") {
				configMap.Data["TLS_INGRESS"] = tlsIngress
			}
		}
	}

	return nil
}

func createLDAPVolumeEntry(p *RuntimeObjectPatcher) corev1.Volume {
	vsource := corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName: p.bdba.LDAPRootCASecret,
		},
	}
	return corev1.Volume{
		Name:         "ldapca-store",
		VolumeSource: vsource,
	}
}

func createLDAPVolumeMountEntry() corev1.VolumeMount {
	return corev1.VolumeMount{
		MountPath: "/ldap/ssl/",
		Name:      "ldapca-store",
		ReadOnly:  true,
	}
}

// patchLDAP
func (p *RuntimeObjectPatcher) patchLDAP() error {
	if p.bdba.LDAPRootCASecret != "" {
		log.Debug("Patching LDAP")
		for k, v := range p.mapOfUniqueIDToRuntimeObject {
			switch v.(type) {
			case *appsv1.Deployment:
				var deployment = p.mapOfUniqueIDToRuntimeObject[k].(*appsv1.Deployment)
				switch deployment.Spec.Template.Name {
				case "bdba-webapp":
					// Create volumes entry
					var specVolumes = &deployment.Spec.Template.Spec.Volumes
					var newVolume = createLDAPVolumeEntry(p)
					*specVolumes = append(*specVolumes, newVolume)
					// Create volumentMounts entry
					for i := range deployment.Spec.Template.Spec.Containers {
						var volumeMounts = &deployment.Spec.Template.Spec.Containers[i].VolumeMounts
						var newVolumeMount = createLDAPVolumeMountEntry()
						*volumeMounts = append(*volumeMounts, newVolumeMount)
					}
				}
			}
		}
	}
	return nil
}

// patchExternalPG
func (p *RuntimeObjectPatcher) patchExternalPG() error {
	if p.bdba.PGHost != "" {
		log.Debug("Patching external PG")
		for k, v := range p.mapOfUniqueIDToRuntimeObject {
			switch v.(type) {
			case *corev1.ConfigMap:
				var configMap = p.mapOfUniqueIDToRuntimeObject[k].(*corev1.ConfigMap)
				if strings.Contains(configMap.Name, "bdba-services-configmap") {
					configMap.Data["PGHOST"] = p.bdba.PGHost
					configMap.Data["PGPORT"] = p.bdba.PGPort
					configMap.Data["PGUSER"] = p.bdba.PGUser
					configMap.Data["PGDATABASE"] = p.bdba.PGDataBase
				}
			}
		}
	}
	return nil
}
