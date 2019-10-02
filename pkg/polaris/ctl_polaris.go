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

	PostgresHost     string
	PostgresPort     int
	PostgresUsername string
	PostgresPassword string
	PostgresSize     string

	SMTPHost        string
	SMTPPort        int
	SMTPUsername    string
	SMTPPassword    string
	SMTPSenderEmail string

	UploadServerSize   string
	EventstoreSize     string
	MongoDBSize        string
	DownloadServerSize string
	ReportStorageSize  string

	EnableReporting bool

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
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Polaris")
	cmd.Flags().StringVar(&ctl.EnvironmentDNS, "environment-dns", ctl.EnvironmentDNS, "Environment DNS")
	cmd.Flags().StringVar(&ctl.ImagePullSecrets, "pull-secret", ctl.ImagePullSecrets, "Pull secret")
	cmd.Flags().StringVar(&ctl.StorageClass, "storage-class", ctl.StorageClass, "Storage class")
	cmd.Flags().BoolVar(&ctl.EnableReporting, "enable-reporting", false, "Send this flag if you wish to enable ReportingPlatform")
	cmd.Flags().StringVar(&ctl.GCPServiceAccount, "gcp-service-account-path", ctl.GCPServiceAccount, "Google Cloud Service account")
	//cmd.Flags().StringVar(&ctl.PostgresHost, "postgres-host", ctl.PostgresHost, "")
	//cmd.Flags().IntVar(&ctl.PostgresPort, "postgres-port", ctl.PostgresPort, "")
	cmd.Flags().StringVar(&ctl.PostgresUsername, "postgres-username", ctl.PostgresUsername, "Postgres username")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "Postgres password")

	cmd.Flags().StringVar(&ctl.PostgresSize, "postgres-size", ctl.PostgresSize, "PVC size to use for postgres. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.UploadServerSize, "uploadserver-size", ctl.UploadServerSize, "PVC size to use for uploadserver. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.EventstoreSize, "eventstore-size", ctl.EventstoreSize, "PVC size to use for eventstore. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.MongoDBSize, "mongodb-size", ctl.MongoDBSize, "PVC size to use for mongodb. e.g.100Gi")
	cmd.Flags().StringVar(&ctl.DownloadServerSize, "downloadserver-size", ctl.DownloadServerSize, "PVC size to use for downloadserver. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.ReportStorageSize, "reportstorage-size", ctl.ReportStorageSize, "PVC size to use for reportstorage. e.g. 100Gi, applicable only if --enable-reporting is set to true.")

	cmd.Flags().StringVar(&ctl.SMTPHost, "smtp-host", ctl.SMTPHost, "SMTP host")
	cmd.Flags().IntVar(&ctl.SMTPPort, "smtp-port", ctl.SMTPPort, "SMTP port")
	cmd.Flags().StringVar(&ctl.SMTPUsername, "smtp-username", ctl.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.SMTPPassword, "smtp-password", ctl.SMTPPassword, "SMTP password")
	cmd.Flags().StringVar(&ctl.SMTPSenderEmail, "smtp-sender-email", ctl.SMTPSenderEmail, "SMTP sender email")

	cmd.Flags().StringVarP(&ctl.organizationProvisionOrganizationDescription, "organization-description", "", ctl.organizationProvisionOrganizationDescription, "Organization description")
	cmd.Flags().StringVarP(&ctl.organizationProvisionOrganizationName, "organization-name", "", ctl.organizationProvisionOrganizationName, "Organization name")
	cmd.Flags().StringVarP(&ctl.organizationProvisionAdminName, "organization-admin-name", "", ctl.organizationProvisionAdminName, "Organization admin name")
	cmd.Flags().StringVarP(&ctl.organizationProvisionAdminUsername, "organization-admin-username", "", ctl.organizationProvisionAdminUsername, "Organization admin username")
	cmd.Flags().StringVarP(&ctl.organizationProvisionAdminEmail, "organization-admin-email", "", ctl.organizationProvisionAdminEmail, "Organization admin username")
	cmd.Flags().StringVarP(&ctl.coverityLicensePath, "coverity-license-path", "", ctl.coverityLicensePath, "Path to the coverity license")
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
		case "environment-dns":
			ctl.spec.EnvironmentDNS = ctl.EnvironmentDNS
		case "enable-reporting":
			ctl.spec.EnableReporting = ctl.EnableReporting
		case "pull-secret":
			ctl.spec.ImagePullSecrets = ctl.ImagePullSecrets
		case "gcp-service-account-path":
			data, err := util.ReadFileData(ctl.GCPServiceAccount)
			if err != nil {
				panic(err)
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
		case "organization-description":
			ctl.spec.OrganizationDetails.OrganizationProvisionOrganizationDescription = ctl.organizationProvisionOrganizationDescription
		case "organization-name":
			ctl.spec.OrganizationDetails.OrganizationProvisionOrganizationName = ctl.organizationProvisionOrganizationName
		case "organization-admin-name":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminName = ctl.organizationProvisionAdminName
		case "organization-admin-username":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminUsername = ctl.organizationProvisionAdminUsername
		case "organization-admin-email":
			ctl.spec.OrganizationDetails.OrganizationProvisionAdminEmail = ctl.organizationProvisionAdminEmail
		case "coverity-license-path":
			data, err := util.ReadFileData(ctl.coverityLicensePath)
			if err != nil {
				panic(err)
			}
			ctl.spec.Licenses.Coverity = data

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
		ImagePullSecrets: "gcr-json-key",
		Licenses:         &Licenses{},
		OrganizationDetails: &OrganizationDetails{
			OrganizationProvisionLicenseSeatCount:   "100",
			OrganizationProvisionLicenseType:        "PAID",
			OrganizationProvisionResultsStartDate:   "2019-02-22",
			OrganizationProvisionResultsEndDate:     "2030-10-01",
			OrganizationProvisionRetentionStartDate: "2019-02-22",
			OrganizationProvisionRetentionEndDate:   "2031-10-01",
		},
		EnableReporting: false,
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
			SMTPDetails:          SMTPDetails{},
			PostgresInstanceType: "internal",
			PostgresDetails: PostgresDetails{
				Host: "postgresql",
				Port: 5432,
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
