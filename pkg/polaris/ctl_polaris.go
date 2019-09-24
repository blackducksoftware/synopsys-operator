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
	"strings"

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
type PolarisCRSpecBuilderFromCobraFlags struct {
	spec             Polaris
	Version          string
	EnvironmentName  string
	EnvironmentDNS   string
	ImagePullSecrets string
	StorageClass     string

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

	UploadServerSize string
	EventstoreSize   string
	Reporting        string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewPolarisCRSpecBuilderFromCobraFlags() *PolarisCRSpecBuilderFromCobraFlags {
	return &PolarisCRSpecBuilderFromCobraFlags{
		spec: Polaris{
			PolarisDBSpec: &PolarisDBSpec{},
			PolarisSpec:   &PolarisSpec{},
		},
	}
}

// GetCRSpec returns a pointer to the PolarisSpec as an interface{}
func (ctl *PolarisCRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return ctl.spec
}

// SetCRSpec sets the PolarisSpec in the struct
func (ctl *PolarisCRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(Polaris)
	if !ok {
		return fmt.Errorf("error setting Polaris spec")
	}

	ctl.spec = convertedSpec
	return nil
}

// SetPredefinedCRSpec sets the Spec to a predefined spec
func (ctl *PolarisCRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	ctl.spec = *GetPolarisDefault()
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *PolarisCRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Polaris")
	cmd.Flags().StringVar(&ctl.EnvironmentDNS, "environment-dns", ctl.EnvironmentDNS, "Environment DNS")
	cmd.Flags().StringVar(&ctl.ImagePullSecrets, "pull-secret", ctl.ImagePullSecrets, "Pull secret")
	cmd.Flags().StringVar(&ctl.StorageClass, "storage-class", ctl.StorageClass, "Storage class")
	cmd.Flags().StringVar(&ctl.Reporting, "reporting", ctl.Reporting, "Enable reporting [true|false]")
	//cmd.Flags().StringVar(&ctl.PostgresHost, "postgres-host", ctl.PostgresHost, "")
	//cmd.Flags().IntVar(&ctl.PostgresPort, "postgres-port", ctl.PostgresPort, "")
	cmd.Flags().StringVar(&ctl.PostgresUsername, "postgres-username", ctl.PostgresUsername, "Postgres username")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "Postgres password")

	cmd.Flags().StringVar(&ctl.PostgresSize, "postgres-size", ctl.PostgresSize, "PVC size to use for postgres. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.UploadServerSize, "uploadserver-size", ctl.UploadServerSize, "PVC size to use for uploadserver. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.EventstoreSize, "eventstore-size", ctl.EventstoreSize, "PVC size to use for eventstore. e.g. 100Gi")

	cmd.Flags().StringVar(&ctl.SMTPHost, "smtp-host", ctl.SMTPHost, "SMTP host")
	cmd.Flags().IntVar(&ctl.SMTPPort, "smtp-port", ctl.SMTPPort, "SMTP port")
	cmd.Flags().StringVar(&ctl.SMTPUsername, "smtp-username", ctl.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.SMTPPassword, "smtp-password", ctl.SMTPPassword, "SMTP password")
	cmd.Flags().StringVar(&ctl.SMTPSenderEmail, "smtp-sender-email", ctl.SMTPSenderEmail, "SMTP sender email")
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the spec
func (ctl *PolarisCRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	return nil
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the spec with the value that's stored
// in the corresponding struct field
func (ctl *PolarisCRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return ctl.spec, nil
}

// SetCRSpecFieldByFlag updates a field in the spec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *PolarisCRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.spec.Version = ctl.Version
		case "environment-dns":
			ctl.spec.EnvironmentDNS = ctl.EnvironmentDNS
		case "reporting":
			ctl.spec.EnableReporting = strings.ToUpper(ctl.Reporting) == "TRUE"
		case "pull-secret":
			ctl.spec.ImagePullSecrets = ctl.ImagePullSecrets
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
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}

// GetPolarisDBDefault returns PolarisDB default configuration
func GetPolarisDefault() *Polaris {
	return &Polaris{
		EnableReporting: false,
		PolarisSpec: &PolarisSpec{
			DownloadServerDetails: &DownloadServerDetails{
				Storage: &Storage{
					StorageSize: DOWNLOAD_SERVER_PV_SIZE,
				},
			},
		},
		ReportingSpec: &ReportingSpec{
			ReportStorageDetails: &ReportStorageDetails{
				Storage: &Storage{
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
