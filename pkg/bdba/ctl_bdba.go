/*
 * Copyright (C) 2020 Synopsys, Inc.
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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelmValuesFromCobraFlags is a type for converting synopsysctl flags
// to Helm Chart fields and values
// args: map of helm chart field to value
type HelmValuesFromCobraFlags struct {
	args     map[string]interface{}
	flagTree FlagTree
}

// FlagTree is a set of fields needed to configure the BDBA Helm Chart
type FlagTree struct {
	Version string `json:"version"`

	Hostname          string `json:"hostname"`
	RabbitMQK8SDomain string `json:"rabbitmqK8sDomain"`

	// Storage
	PSQLStorageClass      string `json:"psqlStorageClass"`
	PSQLSize              string `json:"psqlSize"`
	PSQLExistingClaim     string `json:"psqlExistingClaim"`
	MinioStorageClass     string `json:"minioStorageClass"`
	MinioSize             string `json:"minioSize"`
	MinioExistingClaim    string `json:"minioExistingClaim"`
	RabbitMQStorageClass  string `json:"rabbitmqStorageClass"`
	RabbitMQSize          string `json:"rabbitmqSize"`
	RabbitMQExistingClaim string `json:"rabbitmqExistingClaim"`

	// Licensing
	LicensingUsername string `json:"licensingUsername"`
	LicensingPassword string `json:"licensingPassword"`
	LicensingUpstream string `json:"licensingUpstream"`

	// Rabbitmq configuration

	// Web frontend configuration
	SessionCookieAge int    `json:"sessionCookieAge"`
	FrontendReplicas int    `json:"frontendReplicas"`
	HideLicenses     bool   `json:"hideLicenses"`
	OfflineMode      bool   `json:"offlineMode"`
	AdminEmail       string `json:"adminEmail"`
	ErrorAdminEmail  string `json:"errorAdminEmail"`
	RootURL          string `json:"rootURL"`

	// SMTP configuration
	EmailEnabled      bool   `json:"emailEnabled"`
	EmailSMTPHost     string `json:"emailSMTPHost"`
	EmailSMTPPort     int    `json:"emailSMTPPort"`
	EmailSMTPUser     string `json:"emailSMTPUser"`
	EmailSMTPPassword string `json:"emailSMTPPassword"`
	EmailFrom         string `json:"emailFrom"`
	EmailSecurity     string `json:"emailSecurity"`
	EmailVerify       bool   `json:"emailVerify"`

	// LDAP Authentication
	LDAPEnabled              bool   `json:"ldapEnabled"`
	LDAPServerURI            string `json:"ldapServerURI"`
	LDAPUserDNTemplate       string `json:"ldapUserDNTemplate"`
	LDAPBindAsAuthenticating bool   `json:"ldapBindAsAuthenticating"`
	LDAPBindDN               string `json:"ldapBindDN"`
	LDAPBindPassword         string `json:"ldapBindPassword"`
	LDAPStartTLS             bool   `json:"ldapStartTLS"`
	LDAPVerify               bool   `json:"ldapVerify"`
	LDAPRootCASecret         string `json:"ldapRootCASecret"`
	LDAPRootCAFile           string `json:"ldapRootCAFile"`
	LDAPRequireGroup         string `json:"ldapRequireGroup"`
	LDAPUserSearch           string `json:"ldapUserSearch"`
	LDAPUserSearchScope      string `json:"ldapUserSearchScope"`
	LDAPGroupSearch          string `json:"ldapGroupSearch"`
	LDAPGroupSearchScope     string `json:"ldapGroupSearchScope"`
	LDAPNestedSearch         bool   `json:"ldapNestedSearch"`

	// Logging
	FrontendLogging bool `json:"frontendLogging"`
	WorkerLogging   bool `json:"workerLogging"`

	// Worker scaling
	WorkerReplicas    int `json:"workerReplicas"`
	WorkerConcurrency int `json:"workerConcurrency"` // TODO: Patcher

	// Networking and security
	RootCASecret string `json:"rootCASecret"`
	HTTPProxy    string `json:"httpProxy"`
	HTTPNoProxy  string `json:"httpNoProxy"`

	// Ingress
	IngressEnabled       bool   `json:"ingressEnabled"`
	IngressHost          string `json:"ingressHost"`
	IngressTLSEnabled    bool   `json:"ingressTLSEnabled"`
	IngressTLSSecretName string `json:"ingressTLSSecretName"`

	BrokerURL string `json:"brokerURL"`

	// External PG
	PGHost     string `json:"pgHost"`
	PGPort     string `json:"pgPort"`
	PGUser     string `json:"pgUser"`
	PGDataBase string `json:"pgDataBase"`

	// Secrets
	DjangoSecretKey      string `json:"djangoSecretKey"`
	RabbitMQPassword     string `json:"rabbitMQPassword"`
	RabbitMQErlangCookie string `json:"rabbitMQErlangCookie"`
	PGPassword           string `json:"pgpPassword"`
	MinioAccessKey       string `json:"minioAccessKey"`
	MinioSecretKey       string `json:"minioSecretKey"`
}

// NewHelmValuesFromCobraFlags returns an initialized HelmValuesFromCobraFlags
func NewHelmValuesFromCobraFlags() *HelmValuesFromCobraFlags {
	return &HelmValuesFromCobraFlags{
		args:     map[string]interface{}{},
		flagTree: FlagTree{},
	}
}

// GetArgs returns the map of helm chart fields to values
func (ctl *HelmValuesFromCobraFlags) GetArgs() map[string]interface{} {
	return ctl.args
}

// AddCobraFlagsToCommand adds flags for the BDBA helm chart to the cmd
// master=true is used to add all flags for creating an instance
// master=false is used to add a subset of flags for updating an instance
func (ctl *HelmValuesFromCobraFlags) AddCobraFlagsToCommand(cmd *cobra.Command, master bool) {
	// [DEV NOTE:] please organize flags in order of importance and group related flags

	cmd.Flags().StringVar(&ctl.flagTree.Version, "version", "default", "Description")
	// if master {
	// 	cobra.MarkFlagRequired(cmd.Flags(), "version")
	// }

	cmd.Flags().StringVar(&ctl.flagTree.Hostname, "hostname", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQK8SDomain, "rabbitmq-domain", "default", "Description")

	// Storage
	cmd.Flags().StringVar(&ctl.flagTree.PSQLStorageClass, "psql-storage-class", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PSQLSize, "psql-size", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PSQLExistingClaim, "psql-existing-claim", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.MinioStorageClass, "minio-storage-class", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.MinioSize, "minio-size", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.MinioExistingClaim, "minio-existing-claim", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQStorageClass, "rabbitmq-storage-class", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQSize, "rabbitmq-size", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQExistingClaim, "rabbitmq-existing-claim", "default", "Description")

	// Licensing
	cmd.Flags().StringVar(&ctl.flagTree.LicensingUsername, "license-username", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LicensingPassword, "license-password", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LicensingUpstream, "license-upstream", "default", "Description")

	// Rabbitmq configuration

	// Web frontend configuration
	cmd.Flags().IntVar(&ctl.flagTree.SessionCookieAge, "session-cookie-age", 0, "Description")
	cmd.Flags().IntVar(&ctl.flagTree.FrontendReplicas, "frontend-replicas", 0, "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.HideLicenses, "hide-licenses", false, "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.OfflineMode, "enable-offline-mode", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.AdminEmail, "admin-email", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.ErrorAdminEmail, "error-admin-email", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RootURL, "root-url", "default", "Description")

	// SMTP configuration
	cmd.Flags().BoolVar(&ctl.flagTree.EmailEnabled, "enable-email", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.EmailSMTPHost, "email-smtp-host", "default", "Description")
	cmd.Flags().IntVar(&ctl.flagTree.EmailSMTPPort, "email-smtp-port", 0, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.EmailSMTPUser, "email-smtp-user", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.EmailSMTPPassword, "email-smtp-password", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.EmailFrom, "email-from", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.EmailSecurity, "email-security", "default", "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.EmailVerify, "verify-email", false, "Description")

	// LDAP Authentication
	cmd.Flags().BoolVar(&ctl.flagTree.LDAPEnabled, "enable-ldap", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPServerURI, "ldap-server-uri", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPUserDNTemplate, "ldap-user-dn-template", "default", "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.LDAPBindAsAuthenticating, "enable-ldap-bind-as-authenticating", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPBindDN, "ldap-bind-dn", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPBindPassword, "ldap-bind-password", "default", "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.LDAPStartTLS, "ldap-start-tls", false, "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.LDAPVerify, "verify-ldap", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPRootCASecret, "ldap-root-ca-secret", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPRootCAFile, "ldap-root-ca-file", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPRequireGroup, "ldap-require-group", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPUserSearch, "ldap-user-search", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPUserSearchScope, "ldap-user-search-scope", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPGroupSearch, "ldap-group-search", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.LDAPGroupSearchScope, "ldap-group-search-scope", "default", "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.LDAPNestedSearch, "enable-ldap-nested-search", false, "Description")

	// Logging
	cmd.Flags().BoolVar(&ctl.flagTree.FrontendLogging, "enable-frontend-logging", false, "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.WorkerLogging, "enable-worker-logging", false, "Description")

	// Worker scaling
	cmd.Flags().IntVar(&ctl.flagTree.WorkerReplicas, "worker-replicas", 0, "Description")
	cmd.Flags().IntVar(&ctl.flagTree.WorkerConcurrency, "worker-condurrency", 0, "Description")

	// Networking and security
	cmd.Flags().StringVar(&ctl.flagTree.RootCASecret, "root-ca-secret", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.HTTPProxy, "http-proxy", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.HTTPNoProxy, "http-no-proxy", "default", "Description")

	// Ingress
	cmd.Flags().BoolVar(&ctl.flagTree.IngressEnabled, "enable-ingress", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.IngressHost, "ingress-host", "default", "Description")
	cmd.Flags().BoolVar(&ctl.flagTree.IngressTLSEnabled, "enable-ingress-tls", false, "Description")
	cmd.Flags().StringVar(&ctl.flagTree.IngressTLSSecretName, "ingress-tls-secret-name", "default", "Description")

	cmd.Flags().StringVar(&ctl.flagTree.BrokerURL, "broker-url", "default", "Description")

	// External PG
	cmd.Flags().StringVar(&ctl.flagTree.PGHost, "postgres-host", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PGPort, "postgres-port", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PGUser, "postgres-username", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PGDataBase, "postgres-database", "default", "Description")

	// Secrets
	cmd.Flags().StringVar(&ctl.flagTree.DjangoSecretKey, "django-secret-key", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQPassword, "rabbitmq-password", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.RabbitMQErlangCookie, "rabbitmq-erlang-cookie", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.PGPassword, "postgres-password", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.MinioAccessKey, "minio-access-key", "default", "Description")
	cmd.Flags().StringVar(&ctl.flagTree.MinioSecretKey, "minio-secret-key", "default", "Description")

	cmd.Flags().SortFlags = false
}

// CheckValuesFromFlags returns an error if a value set by a flag is invalid
func (ctl *HelmValuesFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	return nil
}

// GenerateHelmFlagsFromCobraFlags checks each flag in synopsysctl and updates the map to
// contain the corresponding helm chart field and value
func (ctl *HelmValuesFromCobraFlags) GenerateHelmFlagsFromCobraFlags(flagset *pflag.FlagSet) (map[string]interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.AddHelmValueByCobraFlag)

	return ctl.args, nil
}

// AddHelmValueByCobraFlag adds the helm chart field and value based on the flag set
// in synopsysctl
func (ctl *HelmValuesFromCobraFlags) AddHelmValueByCobraFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "hostname":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.Hostname)
		case "rabbitmq-domain":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.RabbitMQK8SDomain)
		case "psql-storage-class":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.PSQLStorageClass)
		case "psql-size":
			util.SetHelmValueInMap(ctl.args, []string{"postgresql", "persistence", "size"}, ctl.flagTree.PSQLSize)
		case "psql-existing-claim":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.PSQLExistingClaim)
		case "minio-storage-class":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.MinioStorageClass)
		case "minio-size":
			util.SetHelmValueInMap(ctl.args, []string{"minio", "persistence", "size"}, ctl.flagTree.MinioSize)
		case "minio-existing-claim":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.MinioExistingClaim)
		case "rabbitmq-storage-class":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.RabbitMQStorageClass)
		case "rabbitmq-size":
			util.SetHelmValueInMap(ctl.args, []string{"rabbitmq", "persistence", "size"}, ctl.flagTree.RabbitMQSize)
		case "rabbitmq-existing-claim":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.RabbitMQExistingClaim)
		case "license-username":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LicensingUsername)
		case "license-password":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LicensingPassword)
		case "license-upstream":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LicensingUpstream)
		case "session-cookie-age":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "web", "sessionCookieAge"}, ctl.flagTree.SessionCookieAge)
		case "frontend-replicas":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "web", "replicas"}, ctl.flagTree.FrontendReplicas)
		case "hide-licenses":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "web", "hideLicenses"}, ctl.flagTree.HideLicenses)
		case "enable-offline-mode":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.OfflineMode)
		case "admin-email":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "web", "admin"}, ctl.flagTree.AdminEmail)
		case "error-admin-email":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.ErrorAdminEmail)
		case "root-url":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "web", "rootURL"}, ctl.flagTree.RootURL)
		case "enable-email":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "email", "enabled"}, ctl.flagTree.EmailEnabled)
		case "email-smtp-host":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailSMTPHost)
		case "email-smtp-port":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailSMTPPort)
		case "email-smtp-user":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailSMTPUser)
		case "email-smtp-password":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailSMTPPassword)
		case "email-from":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailFrom)
		case "email-security":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailSecurity)
		case "verify-email":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.EmailVerify)
		case "enable-ldap":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "ldap", "enabled"}, ctl.flagTree.LDAPEnabled)
		case "ldap-server-uri":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPServerURI)
		case "ldap-user-dn-template":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPUserDNTemplate)
		case "enable-ldap-bind-as-authenticating":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPBindAsAuthenticating)
		case "ldap-bind-dn":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPBindDN)
		case "ldap-bind-password":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPBindPassword)
		case "ldap-start-tls":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPStartTLS)
		case "verify-ldap":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPVerify)
		case "ldap-root-ca-secret":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPRootCASecret)
		case "ldap-root-ca-file":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPRootCAFile)
		case "ldap-required-group":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPRequireGroup)
		case "ldap-user-search":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPUserSearch)
		case "ldap-user-search-scope":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPUserSearchScope)
		case "ldap-group-search":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPGroupSearch)
		case "ldap-group-search-scope":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPGroupSearchScope)
		case "enable-ldap-nested-search":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.LDAPNestedSearch)
		case "enable-frontend-logging":
			util.SetHelmValueInMap(ctl.args, []string{"frontend", "applicationLogging", "enabled"}, ctl.flagTree.FrontendLogging)
		case "enable-worker-logging":
			util.SetHelmValueInMap(ctl.args, []string{"worker", "applicationLogging", "enabled"}, ctl.flagTree.WorkerLogging)
		case "worker-replicas":
			util.SetHelmValueInMap(ctl.args, []string{"worker", "replicas"}, ctl.flagTree.WorkerReplicas)
		case "worker-concurrency":
			util.SetHelmValueInMap(ctl.args, []string{"worker", "concurrency"}, ctl.flagTree.WorkerConcurrency)
		case "root-ca-secret":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.RootCASecret)
		case "http-proxy":
			util.SetHelmValueInMap(ctl.args, []string{"httpProxy"}, ctl.flagTree.HTTPProxy)
		case "http-no-proxy":
			util.SetHelmValueInMap(ctl.args, []string{"httpNoProxy"}, ctl.flagTree.HTTPNoProxy)
		case "enable-ingress":
			util.SetHelmValueInMap(ctl.args, []string{"ingress", "enabled"}, ctl.flagTree.IngressEnabled)
		case "ingress-host":
			util.SetHelmValueInMap(ctl.args, []string{"ingress", "host"}, ctl.flagTree.IngressHost)
		case "enable-ingress-tls":
			util.SetHelmValueInMap(ctl.args, []string{"ingress", "tls", "enabled"}, ctl.flagTree.IngressTLSEnabled)
		case "ingress-tls-secret-name":
			util.SetHelmValueInMap(ctl.args, []string{"ingress", "tls", "secretName"}, ctl.flagTree.IngressTLSSecretName)
		case "broker-url":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.BrokerURL)
		case "postgres-host":
			util.SetHelmValueInMap(ctl.args, []string{}, ctl.flagTree.PGHost)
		case "postgres-port":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.PGPort)
		case "postgres-username":
			util.SetHelmValueInMap(ctl.args, []string{"postgresql", "postgresqlUsername"}, ctl.flagTree.PGUser)
		case "postgres-database":
			util.SetHelmValueInMap(ctl.args, []string{"postgresql", "postgresqlDatabase"}, ctl.flagTree.PGDataBase)
		case "django-secret-key":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.DjangoSecretKey)
		case "rabbitmq-password":
			util.SetHelmValueInMap(ctl.args, []string{"rabbitmq", "rabbitmq", "password"}, ctl.flagTree.RabbitMQPassword)
		case "rabbitmq-erlang-cookie":
			util.SetHelmValueInMap(ctl.args, []string{"rabbitmq", "rabbitmq", "erlangCookie"}, ctl.flagTree.RabbitMQErlangCookie)
		case "postgres-password":
			util.SetHelmValueInMap(ctl.args, []string{"postgresql", "postgresqlPassword"}, ctl.flagTree.PGPassword)
		case "minio-access-key":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.MinioAccessKey)
		case "minio-secret-key":
			util.SetHelmValueInMap(ctl.args, []string{""}, ctl.flagTree.MinioSecretKey)
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
