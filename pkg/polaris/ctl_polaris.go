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

package polaris

import (
	"fmt"
	"regexp"

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

// FlagTree is a set of fields needed to configure the Polaris Reporting Helm Chart
type FlagTree struct {
	Version                   string
	EnvironmentName           string
	FQDN                      string
	StorageClass              string
	GCPServiceAccountFilePath string
	IngressClass              string

	PostgresHost     string
	PostgresPort     int
	PostgresUsername string
	PostgresPassword string
	PostgresSize     string
	PostgresSSLMode  string
	PostgresInternal bool

	SMTPHost                 string
	SMTPPort                 int
	SMTPUsername             string
	SMTPPassword             string
	SMTPSenderEmail          string
	SMTPTlsMode              string
	SMTPTlsIgnoreInvalidCert bool
	SMTPTlsTrustedHosts      string

	DownloadServerSize string
	UploadServerSize   string
	EventstoreSize     string
	MongoDBSize        string
	ReportStorageSize  string

	EnableReporting bool

	coverityLicensePath string
}

// NewHelmValuesFromCobraFlags returns an initialized HelmValuesFromCobraFlags
func NewHelmValuesFromCobraFlags() *HelmValuesFromCobraFlags {
	return &HelmValuesFromCobraFlags{
		args:     make(map[string]interface{}, 0),
		flagTree: FlagTree{},
	}
}

// GetArgs returns the map of helm chart fields to values
func (ctl *HelmValuesFromCobraFlags) GetArgs() map[string]interface{} {
	return ctl.args
}

// SetArgs set the map to values
func (ctl *HelmValuesFromCobraFlags) SetArgs(args map[string]interface{}) {
	for key, value := range args {
		ctl.args[key] = value
	}
}

// AddCobraFlagsToCommand adds flags for the Polaris helm chart to the cmd
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - Set to true for create and false for update
func (ctl *HelmValuesFromCobraFlags) AddCobraFlagsToCommand(cmd *cobra.Command, master bool) {

	// [DEV NOTE:] please organize flags in order of importance
	cmd.Flags().StringVar(&ctl.flagTree.Version, "version", ctl.flagTree.Version, "Version of Polaris you want to install [Example: \"2019.11\"]\n")

	// domain-name specific flags
	cmd.Flags().StringVar(&ctl.flagTree.IngressClass, "ingress-class", "nginx", "Name of ingress class")
	cmd.Flags().StringVar(&ctl.flagTree.FQDN, "fqdn", ctl.flagTree.FQDN, "Fully qualified domain name [Example: \"example.polaris.synopsys.com\"]\n")

	// license related flags
	if master {
		// licenses are not allowed to be changed during update
		cmd.Flags().StringVar(&ctl.flagTree.GCPServiceAccountFilePath, "gcp-service-account-path", ctl.flagTree.GCPServiceAccountFilePath, "Absolute path to given Google Cloud Platform service account for pulling images")
		cmd.Flags().StringVar(&ctl.flagTree.coverityLicensePath, "coverity-license-path", ctl.flagTree.coverityLicensePath, "Absolute path to given Coverity license\n")
	}

	// smtp related flags
	cmd.Flags().StringVar(&ctl.flagTree.SMTPHost, "smtp-host", ctl.flagTree.SMTPHost, "SMTP host")
	cmd.Flags().IntVar(&ctl.flagTree.SMTPPort, "smtp-port", 2525, "SMTP port")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPUsername, "smtp-username", ctl.flagTree.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPPassword, "smtp-password", ctl.flagTree.SMTPPassword, "SMTP password")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPTlsMode, "smtp-tls-mode", "require-starttls", "SMTP TLS mode [disable|try-starttls|require-starttls|require-tls]")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPTlsTrustedHosts, "smtp-trusted-hosts", "*", "Whitespace separated list of trusted hosts")
	cmd.Flags().BoolVar(&ctl.flagTree.SMTPTlsIgnoreInvalidCert, "insecure-skip-smtp-tls-verify", false, "SMTP server's certificates won't be validated")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPSenderEmail, "smtp-sender-email", ctl.flagTree.SMTPSenderEmail, "SMTP sender email\n")

	if master {
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-sender-email")
	}

	// postgres specific flags
	// these flags are specific for an external managed postgres
	cmd.Flags().BoolVar(&ctl.flagTree.PostgresInternal, "enable-postgres-container", false, "If true, synopsysctl will deploy a postgres container backed by persistent volume (Not recommended for production usage)")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresHost, "postgres-host", ctl.flagTree.PostgresHost, "Postgres host. If --enable-postgres-container=true, the defualt is \"postgres\"")
	cmd.Flags().IntVar(&ctl.flagTree.PostgresPort, "postgres-port", 5432, "Postgres port")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresSSLMode, "postgres-ssl-mode", ctl.flagTree.PostgresSSLMode, "Postgres ssl mode [disable|require]")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresUsername, "postgres-username", ctl.flagTree.PostgresUsername, "Postgres username. If --enable-postgres-container=true, the defualt is \"postgres\"")
	// if using in-cluster containerized Postgres, then currently we require "enable-postgres-container", "postgres-password" and optionally "postgres-size"
	// [TODO: make the above point clear to customers]
	cmd.Flags().StringVar(&ctl.flagTree.PostgresPassword, "postgres-password", ctl.flagTree.PostgresPassword, "Postgres password\n")

	// size parameters are not allowed to change during update because of Kubernetes not allowing storage to be decreased (although note that it does allow it to be increased, see https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.EventstoreSize, "eventstore-size", EVENTSTORE_PV_SIZE, "Persistent volume claim size for eventstore")
		cmd.Flags().StringVar(&ctl.flagTree.MongoDBSize, "mongodb-size", MONGODB_PV_SIZE, "Persistent volume claim size for mongodb")
		cmd.Flags().StringVar(&ctl.flagTree.DownloadServerSize, "downloadserver-size", DOWNLOAD_SERVER_PV_SIZE, "Persistent volume claim size for download server")
		cmd.Flags().StringVar(&ctl.flagTree.UploadServerSize, "uploadserver-size", UPLOAD_SERVER_PV_SIZE, "Persistent volume claim size for upload server")
		cmd.Flags().StringVar(&ctl.flagTree.PostgresSize, "postgres-size", POSTGRES_PV_SIZE, "Persistent volume claim size to use for postgres. Only applicable if --enable-postgres-container is set to true")
		cmd.Flags().StringVar(&ctl.flagTree.StorageClass, "storage-class", ctl.flagTree.StorageClass, "Set the storage class to use for all persistent volume claims\n")
	}

	// reporting related flags
	cmd.Flags().BoolVar(&ctl.flagTree.EnableReporting, "enable-reporting", false, "Enable Reporting Platform")
	cmd.Flags().StringVar(&ctl.flagTree.ReportStorageSize, "reportstorage-size", REPORT_STORAGE_PV_SIZE, "Persistent volume claim size for reportstorage. Only applicable if --enable-reporting is set to true")

	cmd.Flags().SortFlags = false
}

// CheckValuesFromFlags returns an error if a value set by a flag is invalid
func (ctl *HelmValuesFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	return nil
}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]+$`)
	return Re.MatchString(email)
}

func validateFQDN(fqdn string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\-]+([.][a-z0-9\-]+)*\.[a-z]+$`)
	return Re.MatchString(fqdn)
}

// GenerateHelmFlagsFromCobraFlags checks each flag in synopsysctl and updates the map to
// contain the corresponding helm chart field and value
func (ctl *HelmValuesFromCobraFlags) GenerateHelmFlagsFromCobraFlags(flagset *pflag.FlagSet) (map[string]interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	var isErrorExist bool
	util.SetHelmValueInMap(ctl.args, []string{"global", "environment"}, "onprem")
	flagset.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			log.Debugf("flag '%s': CHANGED", f.Name)
			switch f.Name {
			case "fqdn":
				// Hosts
				if !validateFQDN(ctl.flagTree.FQDN) {
					log.Errorf("%s is not a valid FQDN", ctl.flagTree.FQDN)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"global", "rootDomain"}, ctl.flagTree.FQDN)
			case "gcp-service-account-path":
				data, err := util.ReadFileData(ctl.flagTree.GCPServiceAccountFilePath)
				if err != nil {
					log.Errorf("failed to read gcp service account file at path: %s, error: %+v", ctl.flagTree.GCPServiceAccountFilePath, err)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"imageCredentials", "password"}, data)
			case "coverity-license-path":
				data, err := util.ReadFileData(ctl.flagTree.coverityLicensePath)
				if err != nil {
					log.Errorf("failed to read coverity license file at path: %s, error: %+v", ctl.flagTree.coverityLicensePath, err)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"coverity", "license"}, data)
			case "enable-reporting":
				util.SetHelmValueInMap(ctl.args, []string{"enableReporting"}, ctl.flagTree.EnableReporting)
			case "ingress-class":
				util.SetHelmValueInMap(ctl.args, []string{"ingressClass"}, ctl.flagTree.IngressClass)
			case "storage-class":
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "storageClass"}, ctl.flagTree.StorageClass)
				util.SetHelmValueInMap(ctl.args, []string{"eventstore", "persistence", "storageClass"}, ctl.flagTree.StorageClass)
				util.SetHelmValueInMap(ctl.args, []string{"polaris-helmchart-minio", "downloadServer", "persistence", "storageClass"}, ctl.flagTree.StorageClass)
				util.SetHelmValueInMap(ctl.args, []string{"polaris-helmchart-minio", "uploadServer", "persistence", "storageClass"}, ctl.flagTree.StorageClass)
				util.SetHelmValueInMap(ctl.args, []string{"mongodb", "persistence", "storageClass"}, ctl.flagTree.StorageClass)
				util.SetHelmValueInMap(ctl.args, []string{"rp-storage-service", "report-storage", "volume", "storageClass"}, ctl.flagTree.StorageClass)
			case "eventstore-size":
				util.SetHelmValueInMap(ctl.args, []string{"eventstore", "persistence", "size"}, ctl.flagTree.EventstoreSize)
			case "downloadserver-size":
				util.SetHelmValueInMap(ctl.args, []string{"polaris-helmchart-minio", "downloadServer", "persistence", "size"}, ctl.flagTree.EventstoreSize)
			case "uploadserver-size":
				util.SetHelmValueInMap(ctl.args, []string{"polaris-helmchart-minio", "uploadServer", "persistence", "size"}, ctl.flagTree.EventstoreSize)
			case "mongodb-size":
				util.SetHelmValueInMap(ctl.args, []string{"mongodb", "persistence", "size"}, ctl.flagTree.MongoDBSize)
			case "reportstorage-size":
				util.SetHelmValueInMap(ctl.args, []string{"rp-storage-service", "report-storage", "volume", "size"}, ctl.flagTree.ReportStorageSize)
			case "smtp-host":
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "smtp", "host"}, ctl.flagTree.SMTPHost)
			case "smtp-port":
				// Ports
				port := ctl.flagTree.SMTPPort
				if port < 1 || port > 65535 {
					log.Errorf("%d is not a valid SMTP port", port)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "smtp", "port"}, port)
			case "smtp-username":
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "smtp", "user"}, ctl.flagTree.SMTPUsername)
			case "smtp-password":
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "smtp", "password"}, ctl.flagTree.SMTPPassword)
			case "smtp-sender-email":
				// Emails
				if !validateEmail(ctl.flagTree.SMTPSenderEmail) {
					log.Errorf("%s is not a valid SMTP sender email address", ctl.flagTree.SMTPSenderEmail)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "smtp", "sender_email"}, ctl.flagTree.SMTPSenderEmail)
				util.SetHelmValueInMap(ctl.args, []string{"swip-onprem", "smtp", "sender_email"}, ctl.flagTree.SMTPSenderEmail)
			case "smtp-tls-mode":
				var tlsMode SMTPTLSMode
				switch SMTPTLSMode(ctl.flagTree.SMTPTlsMode) {
				case SMTPTLSModeDisable:
					tlsMode = SMTPTLSModeDisable
				case SMTPTLSModeTryStartTLS:
					tlsMode = SMTPTLSModeTryStartTLS
				case SMTPTLSModeRequireStartTLS:
					tlsMode = SMTPTLSModeRequireStartTLS
				case SMTPTLSModeRequireTLS:
					tlsMode = SMTPTLSModeRequireTLS
				default:
					log.Errorf("%s is an invalid value for --smtp-tls-mode", ctl.flagTree.SMTPTlsMode)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "auth-server", "smtp", "tls_mode"}, tlsMode)
			case "smtp-trusted-hosts":
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "auth-server", "smtp", "tls_trusted_hosts"}, ctl.flagTree.SMTPTlsTrustedHosts)
			case "insecure-skip-smtp-tls-verify":
				util.SetHelmValueInMap(ctl.args, []string{"onprem-auth-service", "auth-server", "smtp", "tls_check_server_identity"}, !ctl.flagTree.SMTPTlsIgnoreInvalidCert)
			case "enable-postgres-container":
				// If using external postgres, host and username must be set
				if ctl.flagTree.PostgresInternal == false {
					if len(ctl.flagTree.PostgresHost) == 0 {
						log.Errorf("you must set external postgres database postgres-host")
						isErrorExist = true
					}
					if len(ctl.flagTree.PostgresUsername) == 0 {
						log.Errorf("you must set external postgres database postgres-host")
						isErrorExist = true
					}
				}
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "isExternal"}, !ctl.flagTree.PostgresInternal)
			case "postgres-host":
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "host"}, ctl.flagTree.PostgresHost)
			case "postgres-port":
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "port"}, fmt.Sprintf("%d", ctl.flagTree.PostgresPort))
			case "postgres-username":
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "user"}, ctl.flagTree.PostgresUsername)
			case "postgres-password":
				if len(ctl.flagTree.PostgresPassword) == 0 {
					log.Errorf("you must set postgres-password")
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "password"}, ctl.flagTree.PostgresPassword)
			case "postgres-size":
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "size"}, ctl.flagTree.PostgresSize)
			case "postgres-ssl-mode":
				var sslMode PostgresSSLMode
				switch PostgresSSLMode(ctl.flagTree.PostgresSSLMode) {
				case PostgresSSLModeDisable:
					sslMode = PostgresSSLModeDisable
				//case PostgresSSLModeAllow:
				//  ctl.args["postgres.sslMode"] = fmt.Sprintf("%s", PostgresSSLModeAllow)
				//case PostgresSSLModePrefer:
				//  ctl.args["postgres.sslMode"] = fmt.Sprintf("%s", PostgresSSLModePrefer)
				case PostgresSSLModeRequire:
					sslMode = PostgresSSLModeRequire
				default:
					log.Errorf("%s is an invalid value for --postgres-ssl-mode", ctl.flagTree.PostgresSSLMode)
					isErrorExist = true
				}
				util.SetHelmValueInMap(ctl.args, []string{"postgres", "sslMode"}, sslMode)
			default:
				log.Debugf("flag '%s': NOT FOUND", f.Name)
			}
		} else {
			log.Debugf("flag '%s': UNCHANGED", f.Name)
		}
	})

	if isErrorExist {
		log.Fatalf("please fix all the above errors to continue")
	}

	return ctl.args, nil
}
