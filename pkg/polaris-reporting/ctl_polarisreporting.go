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

package polarisreporting

import (
	"fmt"
	"strconv"

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
	Version                  string
	FQDN                     string
	GCPServiceAccount        string
	IngressClass             string
	StorageClass             string
	ReportStorageSize        string
	EventstoreSize           string
	PostgresInternal         string
	PostgresHost             string
	PostgresPort             int
	PostgresUsername         string
	PostgresPassword         string
	PostgresSize             string
	PostgresSSLMode          string
	SMTPHost                 string
	SMTPPort                 int
	SMTPUsername             string
	SMTPPassword             string
	SMTPSenderEmail          string
	SMTPTlsMode              string
	SMTPTlsIgnoreInvalidCert string
	SMTPTlsTrustedHosts      string
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

// AddCobraFlagsToCommand adds flags for the Polaris-Reporting helm chart to the cmd
// master=true is used to add all flags for creating an instance
// master=false is used to add a subset of flags for updating an instance
func (ctl *HelmValuesFromCobraFlags) AddCobraFlagsToCommand(cmd *cobra.Command, master bool) {
	// [DEV NOTE:] please organize flags in order of importance and group related flags
	cmd.Flags().StringVar(&ctl.flagTree.Version, "version", "0.0.69", "Version of Polaris-Reporting you want to install [Example: \"1.0.0\"]\n") // TODO: Put a real version here

	// domain specific flags
	cmd.Flags().StringVar(&ctl.flagTree.FQDN, "fqdn", "nginx", "Fully qualified domain name [Example: \"example.polaris.synopsys.com\"]")
	cmd.Flags().StringVar(&ctl.flagTree.IngressClass, "ingress-class", "", "Name of ingress class\n")

	if master {
		cobra.MarkFlagRequired(cmd.Flags(), "fqdn")
	}

	// license related flags
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.GCPServiceAccount, "gcp-service-account-path", "", "Absolute path to given Google Cloud Platform service account for pulling images\n")

		cobra.MarkFlagRequired(cmd.Flags(), "gcp-service-account-path")
	}

	// storage related flags
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.ReportStorageSize, "reportstorage-size", "5Gi", "Persistent Volume Claim size for reportstorage")
		cmd.Flags().StringVar(&ctl.flagTree.EventstoreSize, "eventstore-size", "50Gi", "Persistent Volume Claim size for eventstore")
	}
	cmd.Flags().StringVar(&ctl.flagTree.StorageClass, "storage-class", ctl.flagTree.StorageClass, "Storage Class for all Polaris-Reporting's storage\n")

	// smtp related flags
	cmd.Flags().StringVar(&ctl.flagTree.SMTPHost, "smtp-host", ctl.flagTree.SMTPHost, "SMTP host")
	cmd.Flags().IntVar(&ctl.flagTree.SMTPPort, "smtp-port", ctl.flagTree.SMTPPort, "SMTP port")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPUsername, "smtp-username", ctl.flagTree.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPPassword, "smtp-password", ctl.flagTree.SMTPPassword, "SMTP password")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPSenderEmail, "smtp-sender-email", ctl.flagTree.SMTPSenderEmail, "SMTP sender email")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPTlsMode, "smtp-tls-mode", ctl.flagTree.SMTPTlsMode, "SMTP TLS mode [disable|try-starttls|require-starttls|require-tls]")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPTlsTrustedHosts, "smtp-trusted-hosts", ctl.flagTree.SMTPTlsTrustedHosts, "Whitespace separated list of trusted hosts")
	cmd.Flags().StringVar(&ctl.flagTree.SMTPTlsIgnoreInvalidCert, "insecure-skip-smtp-tls-verify", "false", "SMTP server's certificates won't be validated\n")

	if master {
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-host")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-port")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-username")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-password")
		cobra.MarkFlagRequired(cmd.Flags(), "smtp-sender-email")
	}

	// postgres specific flags
	cmd.Flags().StringVar(&ctl.flagTree.PostgresInternal, "enable-postgres-container", "false", "If true, synopsysctl will deploy a postgres container backed by persistent volume (Not recommended for production usage)")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresHost, "postgres-host", ctl.flagTree.PostgresHost, "Postgres host")
	cmd.Flags().IntVar(&ctl.flagTree.PostgresPort, "postgres-port", 5432, "Postgres port")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresUsername, "postgres-username", ctl.flagTree.PostgresUsername, "Postgres username")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresPassword, "postgres-password", ctl.flagTree.PostgresPassword, "Postgres password")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresSize, "postgres-size", "50Gi", "Persistent volume claim size to use for postgres. Only applicable if --enable-postgres-container is set to true\n")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresSSLMode, "postgres-ssl-mode", "require", "Postgres ssl mode [disable|require]")

	if master {
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-host")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-username")
		cobra.MarkFlagRequired(cmd.Flags(), "postgres-password")
	}

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
	ctl.args["global"] = map[string]interface{}{"environment": "onprem"}
	flagset.VisitAll(ctl.AddHelmValueByCobraFlag)

	return ctl.args, nil
}

// AddHelmValueByCobraFlag adds the helm chart field and value based on the flag set
// in synopsysctl
func (ctl *HelmValuesFromCobraFlags) AddHelmValueByCobraFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "fqdn":
			globalVal := ctl.args["global"].(map[string]interface{})
			globalVal["rootDomain"] = ctl.flagTree.FQDN
			ctl.args["global"] = globalVal
		case "ingress-class":
			ctl.args["ingressClass"] = ctl.flagTree.IngressClass
		case "storage-class":
			// postgres storage class
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["storageClass"] = ctl.flagTree.StorageClass
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}
			}

			// event store storage class
			if val, ok := ctl.args["eventstore"]; ok && val != nil {
				eventstoreVal := val.(map[string]interface{})
				if val, ok = eventstoreVal["persistence"]; ok && val != nil {
					persistenceVal := val.(map[string]interface{})
					persistenceVal["storageClass"] = ctl.flagTree.StorageClass
					eventstoreVal["persistence"] = persistenceVal
				} else {
					eventstoreVal["persistence"] = map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}
				}
				ctl.args["eventstore"] = eventstoreVal
			} else {
				ctl.args["eventstore"] = map[string]interface{}{"persistence": map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}}
			}

			// report-storage storage class
			if val, ok := ctl.args["rp-storage-service"]; ok && val != nil {
				reportServiceVal := val.(map[string]interface{})
				if val, ok = reportServiceVal["report-storage"]; ok && val != nil {
					storageServiceVal := val.(map[string]interface{})
					if val, ok = storageServiceVal["volume"]; ok && val != nil {
						volumeServiceVal := val.(map[string]interface{})
						volumeServiceVal["storageClass"] = ctl.flagTree.StorageClass
						storageServiceVal["volume"] = volumeServiceVal
					} else {
						storageServiceVal["volume"] = map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}
					}
					reportServiceVal["report-storage"] = storageServiceVal
				} else {
					reportServiceVal["report-storage"] = map[string]interface{}{"volume": map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}}
				}
				ctl.args["rp-storage-service"] = reportServiceVal
			} else {
				ctl.args["rp-storage-service"] = map[string]interface{}{"report-storage": map[string]interface{}{"volume": map[string]interface{}{"storageClass": ctl.flagTree.StorageClass}}}
			}
		case "eventstore-size":
			if val, ok := ctl.args["eventstore"]; ok && val != nil {
				eventstoreVal := val.(map[string]interface{})
				if val, ok = eventstoreVal["persistence"]; ok && val != nil {
					persistenceVal := val.(map[string]interface{})
					persistenceVal["size"] = ctl.flagTree.EventstoreSize
					eventstoreVal["persistence"] = persistenceVal
				} else {
					eventstoreVal["persistence"] = map[string]interface{}{"size": ctl.flagTree.EventstoreSize}
				}
				ctl.args["eventstore"] = eventstoreVal
			} else {
				ctl.args["eventstore"] = map[string]interface{}{"persistence": map[string]interface{}{"size": ctl.flagTree.EventstoreSize}}
			}
		case "reportstorage-size":
			if val, ok := ctl.args["rp-storage-service"]; ok && val != nil {
				reportServiceVal := val.(map[string]interface{})
				if val, ok = reportServiceVal["report-storage"]; ok && val != nil {
					storageServiceVal := val.(map[string]interface{})
					if val, ok = storageServiceVal["volume"]; ok && val != nil {
						volumeServiceVal := val.(map[string]interface{})
						volumeServiceVal["size"] = ctl.flagTree.ReportStorageSize
						storageServiceVal["volume"] = volumeServiceVal
					} else {
						storageServiceVal["volume"] = map[string]interface{}{"size": ctl.flagTree.ReportStorageSize}
					}
					reportServiceVal["report-storage"] = storageServiceVal
				} else {
					reportServiceVal["report-storage"] = map[string]interface{}{"volume": map[string]interface{}{"size": ctl.flagTree.ReportStorageSize}}
				}
				ctl.args["rp-storage-service"] = reportServiceVal
			} else {
				ctl.args["rp-storage-service"] = map[string]interface{}{"report-storage": map[string]interface{}{"volume": map[string]interface{}{"size": ctl.flagTree.ReportStorageSize}}}
			}
		case "smtp-host":
			if val, ok := ctl.args["smtp"]; ok && val != nil {
				smtpVal := val.(map[string]interface{})
				smtpVal["host"] = ctl.flagTree.SMTPHost
				ctl.args["smtp"] = smtpVal
			} else {
				ctl.args["smtp"] = map[string]interface{}{"host": ctl.flagTree.SMTPHost}
			}
		case "smtp-port":
			if val, ok := ctl.args["smtp"]; ok && val != nil {
				smtpVal := val.(map[string]interface{})
				smtpVal["port"] = fmt.Sprintf("%d", ctl.flagTree.SMTPPort)
				ctl.args["smtp"] = smtpVal
			} else {
				ctl.args["smtp"] = map[string]interface{}{"port": fmt.Sprintf("%d", ctl.flagTree.SMTPPort)}
			}
		case "smtp-username":
			if val, ok := ctl.args["smtp"]; ok && val != nil {
				smtpVal := val.(map[string]interface{})
				smtpVal["user"] = ctl.flagTree.SMTPUsername
				ctl.args["smtp"] = smtpVal
			} else {
				ctl.args["smtp"] = map[string]interface{}{"user": ctl.flagTree.SMTPUsername}
			}
		case "smtp-password":
			if val, ok := ctl.args["smtp"]; ok && val != nil {
				smtpVal := val.(map[string]interface{})
				smtpVal["password"] = ctl.flagTree.SMTPPassword
				ctl.args["smtp"] = smtpVal
			} else {
				ctl.args["smtp"] = map[string]interface{}{"password": ctl.flagTree.SMTPPassword}
			}
		case "smtp-sender-email":
			if val, ok := ctl.args["onprem-auth-service"]; ok && val != nil {
				onpremAuthVal := val.(map[string]interface{})
				if val, ok = onpremAuthVal["smtp"]; ok && val != nil {
					smtpVal := val.(map[string]interface{})
					smtpVal["sender_email"] = ctl.flagTree.SMTPSenderEmail
					onpremAuthVal["smtp"] = smtpVal
				} else {
					onpremAuthVal["smtp"] = map[string]interface{}{"sender_email": ctl.flagTree.SMTPSenderEmail}
				}
				ctl.args["onprem-auth-service"] = onpremAuthVal
			} else {
				ctl.args["onprem-auth-service"] = map[string]interface{}{"smtp": map[string]interface{}{"sender_email": ctl.flagTree.SMTPSenderEmail}}
			}
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
				log.Fatalf("%s is an invalid value for --smtp-tls-mode", ctl.flagTree.SMTPTlsMode)
			}
			if val, ok := ctl.args["onprem-auth-service"]; ok && val != nil {
				onpremAuthVal := val.(map[string]interface{})
				if val, ok = onpremAuthVal["smtp"]; ok && val != nil {
					smtpVal := val.(map[string]interface{})
					smtpVal["tls_mode"] = tlsMode
					onpremAuthVal["smtp"] = smtpVal
				} else {
					onpremAuthVal["smtp"] = map[string]interface{}{"tls_mode": tlsMode}
				}
				ctl.args["onprem-auth-service"] = onpremAuthVal
			} else {
				ctl.args["onprem-auth-service"] = map[string]interface{}{"smtp": map[string]interface{}{"tls_mode": tlsMode}}
			}
		case "smtp-trusted-hosts":
			if val, ok := ctl.args["onprem-auth-service"]; ok && val != nil {
				onpremAuthVal := val.(map[string]interface{})
				if val, ok = onpremAuthVal["smtp"]; ok && val != nil {
					smtpVal := val.(map[string]interface{})
					smtpVal["tls_trusted_hosts"] = ctl.flagTree.SMTPTlsTrustedHosts
					onpremAuthVal["smtp"] = smtpVal
				} else {
					onpremAuthVal["smtp"] = map[string]interface{}{"tls_trusted_hosts": ctl.flagTree.SMTPTlsTrustedHosts}
				}
				ctl.args["onprem-auth-service"] = onpremAuthVal
			} else {
				ctl.args["onprem-auth-service"] = map[string]interface{}{"smtp": map[string]interface{}{"tls_trusted_hosts": ctl.flagTree.SMTPTlsTrustedHosts}}
			}
		case "insecure-skip-smtp-tls-verify":
			b, _ := strconv.ParseBool(ctl.flagTree.SMTPTlsIgnoreInvalidCert)
			if val, ok := ctl.args["onprem-auth-service"]; ok && val != nil {
				onpremAuthVal := val.(map[string]interface{})
				if val, ok = onpremAuthVal["smtp"]; ok && val != nil {
					smtpVal := val.(map[string]interface{})
					smtpVal["tls_check_server_identity"] = !b
					onpremAuthVal["smtp"] = smtpVal
				} else {
					onpremAuthVal["smtp"] = map[string]interface{}{"tls_check_server_identity": !b}
				}
				ctl.args["onprem-auth-service"] = onpremAuthVal
			} else {
				ctl.args["onprem-auth-service"] = map[string]interface{}{"smtp": map[string]interface{}{"tls_check_server_identity": !b}}
			}
		case "enable-postgres-container":
			b, _ := strconv.ParseBool(ctl.flagTree.PostgresInternal)
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["isExternal"] = !b
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"isExternal": !b}
			}
		case "postgres-host":
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["host"] = ctl.flagTree.PostgresHost
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"host": ctl.flagTree.PostgresHost}
			}
		case "postgres-port":
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["port"] = fmt.Sprintf("%d", ctl.flagTree.PostgresPort)
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"port": fmt.Sprintf("%d", ctl.flagTree.PostgresPort)}
			}
		case "postgres-username":
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["user"] = ctl.flagTree.PostgresUsername
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"user": ctl.flagTree.PostgresUsername}
			}
		case "postgres-password":
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["password"] = ctl.flagTree.PostgresPassword
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"password": ctl.flagTree.PostgresPassword}
			}
		case "postgres-size":
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["size"] = ctl.flagTree.PostgresSize
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"size": ctl.flagTree.PostgresSize}
			}
			ctl.args["postgres"].(map[string]interface{})["size"] = ctl.flagTree.PostgresSize
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
				log.Fatalf("%s is an invalid value for --postgres-ssl-mode", ctl.flagTree.PostgresSSLMode)
			}
			if val, ok := ctl.args["postgres"]; ok && val != nil {
				postgresVal := val.(map[string]interface{})
				postgresVal["sslMode"] = sslMode
				ctl.args["postgres"] = postgresVal
			} else {
				ctl.args["postgres"] = map[string]interface{}{"sslMode": sslMode}
			}
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
