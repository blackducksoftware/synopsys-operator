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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create an BDBA CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making an BDBA Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	spec    BDBA
	Version string

	Name string
	Hostname string

	MinioAccessKey string
	MinioSecretKey string
	RabbitMQK8SDomain string

	DjangoSecretKey string
	RabbitMQPassword string

	// Storage
	PSQLStorageClass string
	PSQLSize string
	PSQLExistingClaim string
	MinioStorageClass string
	MinioSize string
	MinioExistingClaim string
	RabbitMQStorageClass string
	RabbitMQSize string
	RabbitMQExistingClaim string

	// Web frontend configuration
	SessionCookieAge int
	FrontendReplicas int
	HideLicenses bool
	OfflineMode bool
	AdminEmail string
	ErrorAdminEmail string
	RootURL string

	// SMTP configuration
	EmailEnabled bool
	EmailSMTPHost string
	EmailSMTPPort int
	EmailSMTPUser string
	EmailSMTPPassword string
	EmailFrom string
	EmailSecurity string
	EmailVerify bool

	// LDAP
	LDAPEnabled bool
	LDAPServerURI string
	LDAPUserDNTemplate string
	LDAPBindAsAuthenticating bool
	LDAPBindDN string
	LDAPBindPassword string
	LDAPStartTLS bool
	LDAPVerify bool
	LDAPRootCASecret string
	LDAPRootCAFile string
	LDAPRequireGroup string
	LDAPUserSearch string
	LDAPUserSearchScope string
	LDAPGroupSearch string
	LDAPGroupSearchScope string
	LDAPNestedSearch bool

	// Licensing
	LicensingUsername string
	LicensingPassword string
	LicensingUpstream string

	// Logging
	FrontendLogging bool
	WorkerLogging bool

	// Worker scaling
	WorkerReplicas int
	WorkerConcurrency int

	// Networking and security
	RootCASecret string
	HTTPProxy string
	HTTPNoProxy string

	// Ingress
	IngressEnabled bool
	IngressHost string
	IngressTLSEnabled bool
	IngressTLSSecretName string

	BrokerURL string

	PGPassword string

	// External PG
	PGHost string
	PGPort string
	PGUser string
	PGDataBase string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		spec: BDBA{},
	}
}

// GetCRSpec returns a pointer to the BDBASpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return ctl.spec
}

// SetCRSpec sets the BDBASpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(BDBA)
	if !ok {
		return fmt.Errorf("error setting BDBA spec")
	}

	ctl.spec = convertedSpec
	return nil
}

// SetPredefinedCRSpec sets the Spec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	ctl.spec = *GetBDBADefault()
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of BDBA")
	cmd.Flags().StringVar(&ctl.Name, "name", ctl.Name, "TODO")

	cmd.Flags().StringVar(&ctl.MinioAccessKey, "minio-acceskey", ctl.MinioAccessKey, "TODO")
	cmd.Flags().StringVar(&ctl.MinioSecretKey, "minio-secret-key", ctl.MinioSecretKey, "TODO")
	cmd.Flags().StringVar(&ctl.RabbitMQK8SDomain, "rabbitmq-k8s-domain", ctl.RabbitMQK8SDomain, "TODO")

	cmd.Flags().StringVar(&ctl.DjangoSecretKey, "django-secret-key", ctl.DjangoSecretKey, "TODO")
	cmd.Flags().StringVar(&ctl.RabbitMQPassword, "rabbitmq-password", ctl.RabbitMQPassword, "TODO")

	// Storage
	cmd.Flags().StringVar(&ctl.PSQLStorageClass, "psql-storage-class", ctl.PSQLStorageClass, "StorageClass for postgresql")
	cmd.Flags().StringVar(&ctl.PSQLSize, "psql-size", ctl.PSQLSize, "Size of postgresql claim")
	cmd.Flags().StringVar(&ctl.PSQLExistingClaim, "psql-existing-claim", ctl.PSQLExistingClaim, "Existing claim to use for postgresql")
	cmd.Flags().StringVar(&ctl.MinioStorageClass, "minio-storage-class", ctl.MinioStorageClass, "StorageClass for minio")
	cmd.Flags().StringVar(&ctl.MinioSize, "minio-size", ctl.MinioSize, "Size of minio claim")
	cmd.Flags().StringVar(&ctl.MinioExistingClaim, "minio-existing-claim", ctl.MinioExistingClaim, "Existing claim to use for minio")
	cmd.Flags().StringVar(&ctl.RabbitMQStorageClass, "rabbitmq-storage-class", ctl.RabbitMQStorageClass, "StorageClass for rabbitmq")
	cmd.Flags().StringVar(&ctl.RabbitMQSize, "rabbitmq-size", ctl.RabbitMQSize, "Size of rabbitmq claim")
	cmd.Flags().StringVar(&ctl.RabbitMQExistingClaim, "rabbitmq-existing-claim", ctl.RabbitMQExistingClaim, "Existing claim to use for rabbitmq")

	// Web frontend configuration
	cmd.Flags().IntVar(&ctl.SessionCookieAge, "session-cookie-age", ctl.SessionCookieAge, "Session cookie age")
	cmd.Flags().IntVar(&ctl.FrontendReplicas, "frontend-replicas", ctl.FrontendReplicas, "Number of frontend instances")
	cmd.Flags().BoolVar(&ctl.HideLicenses, "hide-licenses", ctl.HideLicenses, "Hide licensing information from scan")
	cmd.Flags().BoolVar(&ctl.OfflineMode, "offline-mode", ctl.OfflineMode, "Do not make network request to Internet")
	cmd.Flags().StringVar(&ctl.AdminEmail, "admin-email", ctl.AdminEmail, "Admin user's email address")
	cmd.Flags().StringVar(&ctl.ErrorAdminEmail, "error-admin-email", ctl.ErrorAdminEmail, "Error report email receiver ")
	cmd.Flags().StringVar(&ctl.RootURL, "root-url", ctl.RootURL, "Root URL of web service for mails")

	// SMTP configuration
	cmd.Flags().BoolVar(&ctl.EmailEnabled, "email-enabled", ctl.EmailEnabled, "Enable sending email")
	cmd.Flags().StringVar(&ctl.EmailSMTPHost, "email-host", ctl.EmailSMTPHost, "Email SMTP host")
	cmd.Flags().IntVar(&ctl.EmailSMTPPort, "email-host-port", ctl.EmailSMTPPort, "Email SMTP port")
	cmd.Flags().StringVar(&ctl.EmailSMTPUser, "email-host-user", ctl.EmailSMTPUser, "Email SMTP hostname")
	cmd.Flags().StringVar(&ctl.EmailSMTPPassword, "email-host-password", ctl.EmailSMTPPassword, "Email SMTP password")
	cmd.Flags().StringVar(&ctl.EmailFrom, "email-from", ctl.EmailFrom, "Sender of email")
	cmd.Flags().StringVar(&ctl.EmailSecurity, "email-security", ctl.EmailSecurity, "Email security mode. \"none\", \"ssl\", or \"starttls\"")
	cmd.Flags().BoolVar(&ctl.EmailVerify, "email-verify", ctl.EmailVerify, "Verify Email certificate")

	// LDAP
	cmd.Flags().BoolVar(&ctl.LDAPEnabled, "ldap-enabled", ctl.LDAPEnabled, "Enable LDAP authentication")
	cmd.Flags().StringVar(&ctl.LDAPServerURI, "ldap-server-uri", ctl.LDAPServerURI, "LDAP server URI")
	cmd.Flags().StringVar(&ctl.LDAPUserDNTemplate, "ldap-user-dn-template", ctl.LDAPUserDNTemplate, "LDAP dn template for user")
	cmd.Flags().BoolVar(&ctl.LDAPBindAsAuthenticating, "ldap-bind-as-authenticating", ctl.LDAPBindAsAuthenticating, "Bind as authenticating user")
	cmd.Flags().StringVar(&ctl.LDAPBindDN, "ldap-bind-dn", ctl.LDAPBindDN, "LDAP bind DN (generic bind, optional)")
	cmd.Flags().StringVar(&ctl.LDAPBindPassword, "ldap-bind-password", ctl.LDAPBindPassword, "LDAP bind password (generic bind)")
	cmd.Flags().BoolVar(&ctl.LDAPStartTLS, "ldap-starttls", ctl.LDAPStartTLS, "User StartTLS for securing LDAP")
	cmd.Flags().BoolVar(&ctl.LDAPVerify, "ldap-verify", ctl.LDAPVerify, "Verify LDAP server certificate")
	cmd.Flags().StringVar(&ctl.LDAPRootCASecret, "ldap-root-ca-secret", ctl.LDAPRootCASecret, "Secret for LDAP root certificate")
	cmd.Flags().StringVar(&ctl.LDAPRootCAFile, "ldap-root-ca-file", ctl.LDAPRootCAFile, "Secret for LDAP root CA file in Secrets")
	cmd.Flags().StringVar(&ctl.LDAPRequireGroup, "ldap-require-group", ctl.LDAPRequireGroup, "LDAP group required for access")
	cmd.Flags().StringVar(&ctl.LDAPUserSearch, "ldap-user-search", ctl.LDAPUserSearch, "LDAP user search DN template")
	cmd.Flags().StringVar(&ctl.LDAPUserSearchScope, "ldap-user-search-scope", ctl.LDAPUserSearchScope, "LDAP user search scope")
	cmd.Flags().StringVar(&ctl.LDAPGroupSearch, "ldap-group-search", ctl.LDAPGroupSearch, "LDAP group search DN template")
	cmd.Flags().StringVar(&ctl.LDAPGroupSearchScope, "ldap-group-search-scope", ctl.LDAPGroupSearchScope, "LDAP group search scope")
	cmd.Flags().BoolVar(&ctl.LDAPNestedSearch, "ldap-nested-search", ctl.LDAPNestedSearch, "User nested group search")

	// Licensing
	cmd.Flags().StringVar(&ctl.LicensingUsername, "licensing-username", ctl.LicensingUsername, "Username for licensing server")
	cmd.Flags().StringVar(&ctl.LicensingPassword, "licensing-password", ctl.LicensingPassword, "Password for licensing server")
	cmd.Flags().StringVar(&ctl.LicensingUpstream, "licensing-upstream", ctl.LicensingUpstream, "Upstream server for data updates")

	// Logging
	cmd.Flags().BoolVar(&ctl.FrontendLogging, "frontend-logging", ctl.FrontendLogging, "Enable application logging for webapp pods")
	cmd.Flags().BoolVar(&ctl.WorkerLogging, "worker-logging", ctl.WorkerLogging, "Enable application logging for worker pods")

	// Worker scaling
	cmd.Flags().IntVar(&ctl.WorkerReplicas, "worker-replicas", ctl.WorkerReplicas, "Number of scanner instances")
	cmd.Flags().IntVar(&ctl.WorkerConcurrency, "worker-concurrency", ctl.WorkerReplicas, "Number of concurrent scanners in scanner pods")

	// Networking and security
	cmd.Flags().StringVar(&ctl.RootCASecret, "root-ca-secret", ctl.RootCASecret, "Kubernetes Secret for root CA")
	cmd.Flags().StringVar(&ctl.HTTPProxy, "http-proxy", ctl.HTTPProxy, "Proxy URL")
	cmd.Flags().StringVar(&ctl.HTTPNoProxy, "http-no-proxy", ctl.HTTPNoProxy, "No proxy list")

	// Ingress
	cmd.Flags().BoolVar(&ctl.IngressEnabled, "ingress-enabled", ctl.IngressEnabled, "Enable ingress")
	cmd.Flags().StringVar(&ctl.IngressHost, "ingress-host", ctl.IngressHost, "Hostname for ingress")
	cmd.Flags().BoolVar(&ctl.IngressTLSEnabled, "ingress-tls-enabled", ctl.IngressTLSEnabled, "Enable TLS")
	cmd.Flags().StringVar(&ctl.IngressTLSSecretName, "ingress-tls-secret-name", ctl.IngressTLSSecretName, "TLS secret for certificate")

	cmd.Flags().StringVar(&ctl.PGPassword, "pg-password", ctl.PGPassword, "TODO")

	// External PG
	cmd.Flags().StringVar(&ctl.PGHost, "pg-host", ctl.PGHost, "TODO")
	cmd.Flags().StringVar(&ctl.PGPort, "pg-port", ctl.PGPort, "TODO")
	cmd.Flags().StringVar(&ctl.PGUser, "pg-user", ctl.PGUser, "TODO")
	cmd.Flags().StringVar(&ctl.PGDataBase, "pg-database", ctl.PGDataBase, "TODO")
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the spec
func (ctl *CRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	return nil
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the spec with the value that's stored
// in the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}

	// Need to check kubeconfig and name flags before anything else as they are needed in the VisitAll below
	kubeConfig := flagset.Lookup("kubeconfig")
	name := flagset.Lookup("name").Value.String()

	if kubeConfig != nil {
		ctl.spec.KubeConfig = kubeConfig.Value.String()
	} else {
		ctl.spec.KubeConfig = ""
	}

	if name != "" {
		ctl.spec.Name = name
	}

	// DjangoSecretKey and RabbitMQErlangCookie are autogenerated here outside VisitAll
	// because there's no flag to set them manually by user.
	client, _ := CreateKubeClient(ctl.spec.KubeConfig)
	existingSecret := GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-bdba-django-secrets")
	if existingSecret == nil {
		randomString, _ := GenerateRandomString(50)
		ctl.spec.DjangoSecretKey = randomString
	}

	existingSecret = GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-rabbitmq")
	if existingSecret == nil {
		randomString, _ := GenerateRandomString(32)
		ctl.spec.RabbitMQErlangCookie = randomString
	}

	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return ctl.spec, nil
}

// SetCRSpecFieldByFlag updates a field in the spec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	client, _ := CreateKubeClient(ctl.spec.KubeConfig)

	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.spec.Version = ctl.Version
		case "root-url":
			ctl.spec.RootURL = ctl.RootURL
		case "hostname":
			ctl.spec.Hostname = ctl.Hostname

		// Storage
		case "psql-storage-class":
			ctl.spec.PSQLStorageClass = ctl.PSQLStorageClass
		case "minio-storage-class":
			ctl.spec.MinioStorageClass = ctl.MinioStorageClass
		case "rabbitmq-storage-class":
			ctl.spec.RabbitMQStorageClass = ctl.RabbitMQStorageClass

		// Web frontend configuration
		case "session-cookie-age":
			ctl.spec.SessionCookieAge = ctl.SessionCookieAge
		case "admin-email":
			ctl.spec.AdminEmail = ctl.AdminEmail
		case "hide-licenses":
			ctl.spec.HideLicenses = ctl.HideLicenses

		// SMTP configuration
		case "email-enabled":
			ctl.spec.EmailEnabled = ctl.EmailEnabled
		case "email-host":
			ctl.spec.EmailSMTPHost = ctl.EmailSMTPHost
		case "email-host-port":
			ctl.spec.EmailSMTPPort = ctl.EmailSMTPPort
		case "email-host-user":
			ctl.spec.EmailSMTPUser = ctl.EmailSMTPUser
		case "email-host-password":
			ctl.spec.EmailSMTPPassword = ctl.EmailSMTPPassword
		case "email-from":
			ctl.spec.EmailFrom = ctl.EmailFrom
		case "email-security":
			ctl.spec.EmailSecurity = ctl.EmailSecurity
		case "email-verify":
			ctl.spec.EmailVerify = ctl.EmailVerify

		// LDAP
		case "ldap-enabled":
	 		ctl.spec.LDAPEnabled = ctl.LDAPEnabled
		case "ldap-server-uri":
			ctl.spec.LDAPServerURI = ctl.LDAPServerURI
		case "ldap-user-dn-template":
			ctl.spec.LDAPUserDNTemplate = ctl.LDAPUserDNTemplate
		case "ldap-bind-as-authenticating":
			ctl.spec.LDAPBindAsAuthenticating = ctl.LDAPBindAsAuthenticating
		case "ldap-bind-dn":
			ctl.spec.LDAPBindDN = ctl.LDAPBindDN
		case "ldap-bind-password":
			ctl.spec.LDAPBindPassword = ctl.LDAPBindPassword
		case "ldap-starttls":
			ctl.spec.LDAPStartTLS = ctl.LDAPStartTLS
		case "ldap-verify":
			ctl.spec.LDAPVerify = ctl.LDAPVerify
		case "ldap-root-ca-secret":
			ctl.spec.LDAPRootCASecret = ctl.LDAPRootCASecret
		case "ldap-root-ca-file":
			ctl.spec.LDAPRootCAFile = ctl.LDAPRootCAFile
		case "ldap-require-group":
			ctl.spec.LDAPRequireGroup = ctl.LDAPRequireGroup
		case "ldap-user-search":
			ctl.spec.LDAPUserSearch = ctl.LDAPUserSearch
		case "ldap-user-search-scope":
			ctl.spec.LDAPUserSearchScope = ctl.LDAPUserSearchScope
		case "ldap-group-search":
			ctl.spec.LDAPGroupSearch = ctl.LDAPGroupSearch
		case "ldap-group-search-scope":
			ctl.spec.LDAPGroupSearchScope = ctl.LDAPGroupSearchScope
		case "ldap-nested-search":
			ctl.spec.LDAPNestedSearch = ctl.LDAPNestedSearch

		// Licensing
		case "licensing-password":
			ctl.spec.LicensingPassword = ctl.LicensingPassword
		case "licensing-username":
			ctl.spec.LicensingUsername = ctl.LicensingUsername

		// Worker scaling
		case "worker-replicas":
			ctl.spec.WorkerReplicas = ctl.WorkerReplicas

		// Networking and security
		case "http-proxy":
			ctl.spec.HTTPProxy = ctl.HTTPProxy
		case "http-no-proxy":
			ctl.spec.HTTPNoProxy = ctl.HTTPNoProxy

		// Ingress
		case "ingress-enabled":
			ctl.spec.IngressEnabled = ctl.IngressEnabled
		case "ingress-host":
			ctl.spec.IngressHost = ctl.IngressHost
		case "ingress-tls-enabled":
			ctl.spec.IngressTLSEnabled = ctl.IngressTLSEnabled
		case "ingress-tls-secret-name":
			ctl.spec.IngressTLSSecretName = ctl.IngressTLSSecretName

		case "rabbitmq-k8s-domain":
			ctl.spec.RabbitMQK8SDomain = ctl.RabbitMQK8SDomain

		// External PG
		case "pg-host":
			ctl.spec.PGHost = ctl.PGHost
		case "pg-port":
			ctl.spec.PGPort = ctl.PGPort
		case "pg-user":
			ctl.spec.PGUser = ctl.PGUser
		case "pg-database":
			ctl.spec.PGDataBase = ctl.PGDataBase

		// Secrets
		case "rabbitmq-password":
			ctl.spec.RabbitMQPassword = ctl.RabbitMQPassword
		case "pg-password":
			ctl.spec.PGPassword = ctl.PGPassword
		case "minio-acceskey":
			ctl.spec.MinioAccessKey = ctl.MinioAccessKey
		case "minio-secret-key":
			ctl.spec.MinioSecretKey = ctl.MinioSecretKey

		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
		switch f.Name {
		// Autogenerated secrets if not manually set and they don't exist yet in the namespace
		case "rabbitmq-password":
			existingSecret := GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-rabbitmq")
			if existingSecret == nil {
				randomString, _ := GenerateRandomString(32)
				ctl.spec.RabbitMQPassword = randomString
			}
		case "pg-password":
			existingSecret := GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-postgresql")
			if existingSecret == nil {
				randomString, _ := GenerateRandomString(32)
				ctl.spec.PGPassword = randomString
			}
		case "minio-acceskey":
			existingSecret := GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-minio")
			if existingSecret == nil {
				randomString, _ := GenerateRandomString(32) // Generate aws key AKIA + 16. A-Z 2-7
				ctl.spec.MinioAccessKey = randomString
			}
		case "minio-secret-key":
			existingSecret := GetExistingSecret(client, ctl.spec.Namespace, ctl.spec.Name + "-minio")
			if existingSecret == nil {
				randomString, _ := GenerateRandomString(40)
				ctl.spec.MinioSecretKey = randomString
			}
		}
	}
}
