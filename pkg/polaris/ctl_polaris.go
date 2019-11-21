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

	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create an Polaris CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making an Polaris Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	spec              Polaris
	Version           string
	EnvironmentName   string
	EnvironmentDNS    string
	ImagePullSecrets  string
	StorageClass      string
	GCPServiceAccount string
	IngressClass      string
	Registry          string

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

	UploadServerSize   string
	EventstoreSize     string
	MongoDBSize        string
	DownloadServerSize string
	ReportStorageSize  string

	EnableReporting bool

	polarisLicensePath                           string
	coverityLicensePath                          string
	organizationProvisionOrganizationDescription string
	organizationProvisionOrganizationName        string
	organizationProvisionAdminName               string
	organizationProvisionAdminUsername           string
	organizationProvisionAdminEmail              string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		spec: Polaris{
			PolarisDBSpec: &PolarisDBSpec{},
			PolarisSpec:   &PolarisSpec{},
		},
	}
}

// GetCRSpec returns a pointer to the PolarisSpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return ctl.spec
}

// SetCRSpec sets the PolarisSpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(Polaris)
	if !ok {
		return fmt.Errorf("error setting Polaris spec")
	}

	ctl.spec = convertedSpec
	return nil
}

// SetPredefinedCRSpec sets the Spec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	ctl.spec = *GetPolarisDefault()
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - Set to true for create and false for update
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {

	// [DEV NOTE:] please organize flags in order of importance
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Polaris you want to install [Example: \"2019.11.1\"]\n")

	// domain-name specific flags
	cmd.Flags().StringVar(&ctl.IngressClass, "ingress-class", GetPolarisDefault().IngressClass, "Name of ingress class")
	cmd.Flags().StringVar(&ctl.EnvironmentDNS, "fqdn", ctl.EnvironmentDNS, "Fully qualified domain name [Example: \"example.polaris.synopsys.com\"]\n")

	// license related flags
	if master {
		// licenses are not allowed to be changed during update
		cmd.Flags().StringVar(&ctl.GCPServiceAccount, "gcp-service-account-path", ctl.GCPServiceAccount, "Absolute path to given Google Cloud Platform service account for pulling images")
		cmd.Flags().StringVarP(&ctl.polarisLicensePath, "polaris-license-path", "", ctl.polarisLicensePath, "Absolute path to given Polaris Platform license")
		cmd.Flags().StringVarP(&ctl.coverityLicensePath, "coverity-license-path", "", ctl.coverityLicensePath, "Absolute path to given Coverity license\n")
	}

	// smtp related flags
	cmd.Flags().StringVar(&ctl.SMTPHost, "smtp-host", ctl.SMTPHost, "SMTP host")
	cmd.Flags().IntVar(&ctl.SMTPPort, "smtp-port", ctl.SMTPPort, "SMTP port")
	cmd.Flags().StringVar(&ctl.SMTPUsername, "smtp-username", ctl.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.SMTPPassword, "smtp-password", ctl.SMTPPassword, "SMTP password")
	cmd.Flags().StringVar(&ctl.SMTPSenderEmail, "smtp-sender-email", ctl.SMTPSenderEmail, "SMTP sender email")
	cmd.Flags().StringVar(&ctl.SMTPTlsMode, "smtp-tls-mode", string(GetPolarisDefault().PolarisDBSpec.SMTPDetails.TLSMode), "SMTP TLS mode [disable|try-starttls|require-starttls|require-tls]")
	cmd.Flags().StringVar(&ctl.SMTPTlsTrustedHosts, "smtp-trusted-hosts", ctl.SMTPTlsTrustedHosts, "Whitespace separated list of trusted hosts")
	cmd.Flags().BoolVar(&ctl.SMTPTlsIgnoreInvalidCert, "insecure-skip-smtp-tls-verify", false, "SMTP server's certificates won't be validated\n")

	// postgres specific flags
	// these flags are specific for an external managed postgres
	cmd.Flags().StringVar(&ctl.PostgresHost, "postgres-host", ctl.PostgresHost, "Postgres host")
	cmd.Flags().IntVar(&ctl.PostgresPort, "postgres-port", ctl.PostgresPort, "Postgres port")
	cmd.Flags().StringVar(&ctl.PostgresSSLMode, "postgres-ssl-mode", string(GetPolarisDefault().PolarisDBSpec.PostgresDetails.SSLMode), "Postgres ssl mode [disable|require]")
	cmd.Flags().StringVar(&ctl.PostgresUsername, "postgres-username", ctl.PostgresUsername, "Postgres username")
	// if using in-cluster containerized Postgres, then currently we require "enable-postgres-container", "postgres-password" and optionally "postgres-size"
	// [TODO: make the above point clear to customers]
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "Postgres password")
	cmd.Flags().BoolVar(&ctl.PostgresInternal, "enable-postgres-container", GetPolarisDefault().PolarisDBSpec.PostgresDetails.IsInternal, "If true, synopsysctl will deploy a postgres container backed by persistent volume (Not recommended for production usage)\n")

	// organization settings are not allowed to be changed during update
	if master {
		cmd.Flags().StringVarP(&ctl.organizationProvisionOrganizationDescription, "organization-description", "", ctl.organizationProvisionOrganizationDescription, "Organization description")
		cmd.Flags().StringVarP(&ctl.organizationProvisionAdminEmail, "organization-admin-email", "", ctl.organizationProvisionAdminEmail, "Organization admin email")
		cmd.Flags().StringVarP(&ctl.organizationProvisionAdminName, "organization-admin-name", "", ctl.organizationProvisionAdminName, "Organization admin name")
		cmd.Flags().StringVarP(&ctl.organizationProvisionAdminUsername, "organization-admin-username", "", ctl.organizationProvisionAdminUsername, "Organization admin username\n")
	}

	// size parameters are not allowed to change during update because of Kubernetes not allowing storage to be decreased (although note that it does allow it to be increased, see https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)
	if master {
		cmd.Flags().StringVar(&ctl.EventstoreSize, "eventstore-size", GetPolarisDefault().PolarisDBSpec.EventstoreDetails.Storage.StorageSize, "Persistent volume claim size for eventstore")
		cmd.Flags().StringVar(&ctl.MongoDBSize, "mongodb-size", GetPolarisDefault().PolarisDBSpec.MongoDBDetails.Storage.StorageSize, "Persistent volume claim size for mongodb")
		cmd.Flags().StringVar(&ctl.DownloadServerSize, "downloadserver-size", GetPolarisDefault().PolarisSpec.DownloadServerDetails.Storage.StorageSize, "Persistent volume claim size for downloadserver")
		cmd.Flags().StringVar(&ctl.UploadServerSize, "uploadserver-size", GetPolarisDefault().PolarisDBSpec.UploadServerDetails.Storage.StorageSize, "Persistent volume claim size for uploadserver")
		cmd.Flags().StringVar(&ctl.PostgresSize, "postgres-size", GetPolarisDefault().PolarisDBSpec.PostgresDetails.Storage.StorageSize, "Persistent volume claim size to use for postgres. Only applicable if --enable-postgres-container is set to true")
		cmd.Flags().StringVar(&ctl.StorageClass, "storage-class", ctl.StorageClass, "Set the storage class to use for all persistent volume claims\n")
	}

	// reporting related flags
	cmd.Flags().BoolVar(&ctl.EnableReporting, "enable-reporting", GetPolarisDefault().EnableReporting, "Enable Reporting Platform")
	cmd.Flags().StringVar(&ctl.ReportStorageSize, "reportstorage-size", GetPolarisDefault().ReportingSpec.ReportStorageDetails.Storage.StorageSize, "Persistent volume claim size for reportstorage. Only applicable if --enable-reporting is set to true")

	// flags that are add-ons (helpful, but not required to set up an environment
	cmd.Flags().StringVar(&ctl.Registry, "registry", ctl.Registry, "Docker registry e.g. docker.io/myuser")
	cmd.Flags().StringVar(&ctl.ImagePullSecrets, "pull-secret", ctl.ImagePullSecrets, "Pull secret when using a private registry\n")

	cmd.Flags().SortFlags = false
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
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return ctl.spec, nil
}

// SetCRSpecFieldByFlag updates a field in the spec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.spec.Version = ctl.Version
		case "fqdn":
			ctl.spec.EnvironmentDNS = ctl.EnvironmentDNS
		case "enable-reporting":
			ctl.spec.EnableReporting = ctl.EnableReporting
		case "pull-secret":
			ctl.spec.ImagePullSecrets = ctl.ImagePullSecrets
		case "ingress-class":
			ctl.spec.IngressClass = ctl.IngressClass
		case "storage-class":
			ctl.spec.StorageClass = ctl.StorageClass
		case "registry":
			ctl.spec.Registry = ctl.Registry
		case "gcp-service-account-path":
			data, err := util.ReadFileData(ctl.GCPServiceAccount)
			if err != nil {
				log.Fatalf("failed to read gcp service account file at path: %s, error: %+v", ctl.GCPServiceAccount, err)
			}
			ctl.spec.GCPServiceAccount = data
		case "postgres-host":
			ctl.spec.PolarisDBSpec.PostgresDetails.Host = ctl.PostgresHost
		case "postgres-port":
			ctl.spec.PolarisDBSpec.PostgresDetails.Port = ctl.PostgresPort
		case "postgres-username":
			ctl.spec.PolarisDBSpec.PostgresDetails.Username = ctl.PostgresUsername
		case "postgres-password":
			ctl.spec.PolarisDBSpec.PostgresDetails.Password = ctl.PostgresPassword
		case "postgres-size":
			ctl.spec.PolarisDBSpec.PostgresDetails.Storage.StorageSize = ctl.PostgresSize
		case "enable-postgres-container":
			ctl.spec.PolarisDBSpec.PostgresDetails.IsInternal = ctl.PostgresInternal
			if ctl.PostgresInternal {
				ctl.spec.PolarisDBSpec.PostgresDetails.SSLMode = PostgresSSLModeDisable
			}
		case "postgres-ssl-mode":
			switch PostgresSSLMode(ctl.PostgresSSLMode) {
			case PostgresSSLModeDisable:
				ctl.spec.PolarisDBSpec.PostgresDetails.SSLMode = PostgresSSLModeDisable
			//case PostgresSSLModeAllow:
			//	ctl.spec.PolarisDBSpec.PostgresDetails.SSLMode = PostgresSSLModeAllow
			//case PostgresSSLModePrefer:
			//	ctl.spec.PolarisDBSpec.PostgresDetails.SSLMode = PostgresSSLModePrefer
			case PostgresSSLModeRequire:
				ctl.spec.PolarisDBSpec.PostgresDetails.SSLMode = PostgresSSLModeRequire
			default:
				log.Fatalf("%s is an invalid value for --postgres-ssl-mode", ctl.PostgresSSLMode)
			}
		case "uploadserver-size":
			ctl.spec.PolarisDBSpec.UploadServerDetails.Storage.StorageSize = ctl.UploadServerSize
		case "eventstore-size":
			ctl.spec.PolarisDBSpec.EventstoreDetails.Storage.StorageSize = ctl.EventstoreSize
		case "mongodb-size":
			ctl.spec.PolarisDBSpec.MongoDBDetails.Storage.StorageSize = ctl.MongoDBSize
		case "downloadserver-size":
			ctl.spec.PolarisSpec.DownloadServerDetails.Storage.StorageSize = ctl.DownloadServerSize
		case "reportstorage-size":
			if ctl.EnableReporting {
				ctl.spec.ReportingSpec.ReportStorageDetails.Storage.StorageSize = ctl.ReportStorageSize
			}
		case "smtp-host":
			ctl.spec.PolarisDBSpec.SMTPDetails.Host = ctl.SMTPHost
		case "smtp-port":
			ctl.spec.PolarisDBSpec.SMTPDetails.Port = ctl.SMTPPort
		case "smtp-username":
			ctl.spec.PolarisDBSpec.SMTPDetails.Username = ctl.SMTPUsername
		case "smtp-password":
			ctl.spec.PolarisDBSpec.SMTPDetails.Password = ctl.SMTPPassword
		case "smtp-sender-email":
			ctl.spec.PolarisDBSpec.SMTPDetails.SenderEmail = ctl.SMTPSenderEmail
		case "smtp-tls-mode":
			switch SMTPTLSMode(ctl.SMTPTlsMode) {
			case SMTPTLSModeDisable:
				ctl.spec.PolarisDBSpec.SMTPDetails.TLSMode = SMTPTLSModeDisable
			case SMTPTLSModeTryStartTLS:
				ctl.spec.PolarisDBSpec.SMTPDetails.TLSMode = SMTPTLSModeTryStartTLS
			case SMTPTLSModeRequireStartTLS:
				ctl.spec.PolarisDBSpec.SMTPDetails.TLSMode = SMTPTLSModeRequireStartTLS
			case SMTPTLSModeRequireTLS:
				ctl.spec.PolarisDBSpec.SMTPDetails.TLSMode = SMTPTLSModeRequireTLS
			default:
				log.Fatalf("%s is an invalid value for --smtp-tls-mode", ctl.SMTPTlsMode)
			}
		case "smtp-trusted-hosts":
			ctl.spec.PolarisDBSpec.SMTPDetails.TLSTrustedHosts = ctl.SMTPTlsTrustedHosts
		case "insecure-skip-smtp-tls-verify":
			ctl.spec.PolarisDBSpec.SMTPDetails.TLSCheckServerIdentity = !ctl.SMTPTlsIgnoreInvalidCert
		case "organization-description":
			ctl.spec.OrganizationDetails.OrganizationProvisionOrganizationDescription = ctl.organizationProvisionOrganizationDescription
		case "organization-admin-name":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminName = ctl.organizationProvisionAdminName
		case "organization-admin-username":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminUsername = ctl.organizationProvisionAdminUsername
		case "organization-admin-email":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminEmail = ctl.organizationProvisionAdminEmail
		case "coverity-license-path":
			data, err := util.ReadFileData(ctl.coverityLicensePath)
			if err != nil {
				log.Fatalf("failed to read coverity license file at path: %s, error: %+v", ctl.coverityLicensePath, err)
			}
			ctl.spec.Licenses.Coverity = data
		case "polaris-license-path":
			data, err := util.ReadFileData(ctl.polarisLicensePath)
			if err != nil {
				log.Fatalf("failed to read polaris platform license file at path: %s, error: %+v", ctl.polarisLicensePath, err)
			}
			ctl.spec.Licenses.Polaris = data
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}

// GetPolarisDefault returns PolarisDB default configuration
func GetPolarisDefault() *Polaris {
	return &Polaris{
		ImagePullSecrets:    "gcr-json-key",
		IngressClass:        "nginx",
		Licenses:            &Licenses{},
		OrganizationDetails: &OrganizationDetails{},
		EnableReporting:     false,
		PolarisSpec: &PolarisSpec{
			DownloadServerDetails: DownloadServerDetails{
				Storage: Storage{
					StorageSize: DOWNLOAD_SERVER_PV_SIZE,
				},
			},
		},
		ReportingSpec: &ReportingSpec{
			ReportStorageDetails: ReportStorageDetails{
				Storage: Storage{
					StorageSize: REPORT_STORAGE_PV_SIZE,
				},
			},
		},
		PolarisDBSpec: &PolarisDBSpec{
			SMTPDetails: SMTPDetails{
				TLSCheckServerIdentity: true,
				TLSMode:                SMTPTLSModeDisable,
			},
			PostgresDetails: PostgresDetails{
				Host:       "postgresql",
				Username:   "postgres",
				Port:       5432,
				IsInternal: false,
				SSLMode:    PostgresSSLModeRequire,
				Storage: Storage{
					StorageSize: POSTGRES_PV_SIZE,
				},
			},
			MongoDBDetails: MongoDBDetails{
				Storage: Storage{
					StorageSize: MONGODB_PV_SIZE,
				},
			},
			EventstoreDetails: EventstoreDetails{
				Replicas: util.IntToInt32(3),
				Storage: Storage{
					StorageSize: EVENTSTORE_PV_SIZE,
				},
			},
			UploadServerDetails: UploadServerDetails{
				Replicas: util.IntToInt32(1),
				Storage: Storage{
					StorageSize: UPLOAD_SERVER_PV_SIZE,
				},
			},
		},
	}
}
