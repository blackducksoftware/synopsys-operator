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

// BDBA configures all BDBA specifications
type BDBA struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`

	Hostname    string `json:"hostname"`
	RabbitMQK8SDomain string `json:"rabbitmqK8sDomain"`

	// Storage
	PSQLStorageClass string `json:"psqlStorageClass"`
	PSQLSize string `json:"psqlSize"`
	PSQLExistingClaim string `json:"psqlExistingClaim"`
	MinioStorageClass string `json:"minioStorageClass"`
	MinioSize string `json:"minioSize"`
	MinioExistingClaim string `json:"minioExistingClaim"`
	RabbitMQStorageClass string `json:"rabbitmqStorageClass"`
	RabbitMQSize string `json:"rabbitmqSize"`
	RabbitMQExistingClaim string `json:"rabbitmqExistingClaim"`

	// Licensing
	LicensingUsername string `json:"licensingUsername"`
	LicensingPassword string `json:"licensingPassword"`
	LicensingUpstream string `json:"licensingUpstream"`

	// Rabbitmq configuration

	// Web frontend configuration
	SessionCookieAge int `json:"sessionCookieAge"`
	FrontendReplicas int `json:"frontendReplicas"`
	HideLicenses      bool `json:"hideLicenses"`
	OfflineMode bool `json:"offlineMode"`
	AdminEmail string `json:"adminEmail"`
	ErrorAdminEmail string `json:"errorAdminEmail"`
	RootURL string `json:"rootURL"`

	// SMTP configuration
	EmailEnabled bool `json:"emailEnabled"`
	EmailSMTPHost string `json:"emailSMTPHost"`
	EmailSMTPPort int `json:"emailSMTPPort"`
	EmailSMTPUser string `json:"emailSMTPUser"`
	EmailSMTPPassword string `json:"emailSMTPPassword"`
	EmailFrom string `json:"emailFrom"`
	EmailSecurity string `json:"emailSecurity"`
	EmailVerify bool `json:"emailVerify"`

	// LDAP Authentication
	LDAPEnabled bool `json:"ldapEnabled"`
	LDAPServerURI string `json:"ldapServerURI"`
	LDAPUserDNTemplate string `json:"ldapUserDNTemplate"`
	LDAPBindAsAuthenticating bool `json:"ldapBindAsAuthenticating"`
	LDAPBindDN string `json:"ldapBindDN"`
	LDAPBindPassword string `json:"ldapBindPassword"`
	LDAPStartTLS bool `json:"ldapStartTLS"`
	LDAPVerify bool `json:"ldapVerify"`
	LDAPRootCASecret string `json:"ldapRootCASecret"`
	LDAPRootCAFile string `json:"ldapRootCAFile"`
	LDAPRequireGroup string `json:"ldapRequireGroup"`
	LDAPUserSearch string `json:"ldapUserSearch"`
	LDAPUserSearchScope string `json:"ldapUserSearchScope"`
	LDAPGroupSearch string `json:"ldapGroupSearch"`
	LDAPGroupSearchScope string `json:"ldapGroupSearchScope"`
	LDAPNestedSearch bool `json:"ldapNestedSearch"`

	// Logging
	FrontendLogging bool `json:"frontendLogging"`
	WorkerLogging bool `json:"workerLogging"`

	// Worker scaling
	WorkerReplicas int `json:"workerReplicas"`
	WorkerConcurrency int `json:"workerConcurrency"` // TODO: Patcher

	// Networking and security
	RootCASecret string `json:"rootCASecret"`
	HTTPProxy string `json:"httpProxy"`
	HTTPNoProxy string `json:"httpNoProxy"`

	// Ingress
	IngressEnabled bool `json:"ingressEnabled"`
	IngressHost string `json:"ingressHost"`
	IngressTLSEnabled bool `json:"ingressTLSEnabled"`
	IngressTLSSecretName string `json:"ingressTLSSecretName"`

	BrokerURL string `json:"brokerURL"`

	// External PG
	PGHost string `json:"pgHost"`
	PGPort string `json:"pgPort"`
	PGUser string `json:"pgUser"`
	PGDataBase string `json:"pgDataBase"`

	// Secrets
	DjangoSecretKey string `json:"djangoSecretKey"`
	RabbitMQPassword string `json:"rabbitMQPassword"`
	RabbitMQErlangCookie string `json:"rabbitMQErlangCookie"`
	PGPassword string `json:"pgpPassword"`
	MinioAccessKey string `json:"minioAccessKey"`
	MinioSecretKey string `json:"minioSecretKey"`

	KubeConfig string `json:"kubeConfig"`
}
