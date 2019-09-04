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
	synopsysV1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/utils"
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
	spec             polarisMultiCRSpec
	Version          string
	EnvironmentName  string
	EnvironmentDNS   string
	ImagePullSecrets string
	StorageClass     string

	PostgresHost     string
	PostgresPort     int32
	PostgresUsername string
	PostgresPassword string
	PostgresSize     string

	SMTPHost     string
	SMTPPort     int32
	SMTPUsername string
	SMTPPassword string

	UploadServerSize string
	EventstoreSize   string
}

type polarisMultiCRSpec struct {
	authSpec      *synopsysV1.AuthServerSpec
	polarisDBSpec *synopsysV1.PolarisDBSpec
	polarisSpec   *synopsysV1.PolarisSpec
}

type polarisMultiCR struct {
	auth      *synopsysV1.AuthServer
	polarisDB *synopsysV1.PolarisDB
	polaris   *synopsysV1.Polaris
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewPolarisCRSpecBuilderFromCobraFlags() *PolarisCRSpecBuilderFromCobraFlags {
	return &PolarisCRSpecBuilderFromCobraFlags{
		spec: struct {
			authSpec      *synopsysV1.AuthServerSpec
			polarisDBSpec *synopsysV1.PolarisDBSpec
			polarisSpec   *synopsysV1.PolarisSpec
		}{authSpec: &synopsysV1.AuthServerSpec{}, polarisDBSpec: &synopsysV1.PolarisDBSpec{}, polarisSpec: &synopsysV1.PolarisSpec{}},
	}
}

// GetCRSpec returns a pointer to the PolarisSpec as an interface{}
func (ctl *PolarisCRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return ctl.spec
}

// SetCRSpec sets the PolarisSpec in the struct
func (ctl *PolarisCRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(polarisMultiCRSpec)
	if !ok {
		return fmt.Errorf("error setting Polaris spec")
	}

	ctl.spec = convertedSpec
	return nil
}

// SetPredefinedCRSpec sets the Spec to a predefined spec
func (ctl *PolarisCRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	ctl.spec.polarisDBSpec = utils.GetPolarisDBDefault()
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *PolarisCRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Polaris")
	cmd.Flags().StringVar(&ctl.EnvironmentDNS, "environment-dns", ctl.EnvironmentDNS, "Environment DNS")
	cmd.Flags().StringVar(&ctl.EnvironmentName, "environment-name", ctl.EnvironmentName, "Environment name")
	cmd.Flags().StringVar(&ctl.ImagePullSecrets, "pull-secret", ctl.ImagePullSecrets, "Pull secret")
	cmd.Flags().StringVar(&ctl.StorageClass, "storage-class", ctl.StorageClass, "Storage class")

	//cmd.Flags().StringVar(&ctl.PostgresHost, "postgres-host", ctl.PostgresHost, "")
	//cmd.Flags().Int32Var(&ctl.PostgresPort, "postgres-port",  ctl.PostgresPort, "")
	cmd.Flags().StringVar(&ctl.PostgresUsername, "postgres-username", ctl.PostgresUsername, "Postgres username")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "Postgres password")

	cmd.Flags().StringVar(&ctl.PostgresSize, "postgres-size", ctl.PostgresSize, "PVC size to use for postgres. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.UploadServerSize, "uploadserver-size", ctl.UploadServerSize, "PVC size to use for uploadserver. e.g. 100Gi")
	cmd.Flags().StringVar(&ctl.EventstoreSize, "eventstore-size", ctl.EventstoreSize, "PVC size to use for eventstore. e.g. 100Gi")

	cmd.Flags().StringVar(&ctl.SMTPHost, "smtp-host", ctl.SMTPHost, "SMTP host")
	cmd.Flags().Int32Var(&ctl.SMTPPort, "smtp-port", ctl.SMTPPort, "SMTP port")
	cmd.Flags().StringVar(&ctl.SMTPUsername, "smtp-username", ctl.SMTPUsername, "SMTP username")
	cmd.Flags().StringVar(&ctl.SMTPPassword, "smtp-password", ctl.SMTPPassword, "SMTP password")
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
			ctl.spec.authSpec.Version = ctl.Version
			ctl.spec.polarisDBSpec.Version = ctl.Version
			ctl.spec.polarisSpec.Version = ctl.Version
		case "environment-dns":
			ctl.spec.authSpec.EnvironmentDNS = ctl.EnvironmentDNS
			ctl.spec.polarisDBSpec.EnvironmentDNS = ctl.EnvironmentDNS
			ctl.spec.polarisSpec.EnvironmentDNS = ctl.EnvironmentDNS
		case "environment-name":
			ctl.spec.authSpec.EnvironmentName = ctl.EnvironmentName
			ctl.spec.polarisDBSpec.EnvironmentName = ctl.EnvironmentName
			ctl.spec.polarisSpec.EnvironmentName = ctl.EnvironmentName
		case "pull-secret":
			ctl.spec.authSpec.ImagePullSecrets = ctl.ImagePullSecrets
			ctl.spec.polarisDBSpec.ImagePullSecrets = ctl.ImagePullSecrets
			ctl.spec.polarisSpec.ImagePullSecrets = ctl.ImagePullSecrets
		case "storage-class":
			ctl.spec.polarisDBSpec.PostgresStorageDetails.StorageClass = &ctl.StorageClass
			ctl.spec.polarisDBSpec.UploadServerDetails.Storage.StorageClass = &ctl.StorageClass
		case "postgres-host":
			ctl.spec.polarisDBSpec.PostgresDetails.Host = ctl.PostgresHost
		case "postgres-port":
			ctl.spec.polarisDBSpec.PostgresDetails.Port = &ctl.PostgresPort
		case "postgres-username":
			ctl.spec.polarisDBSpec.PostgresDetails.Username = ctl.PostgresUsername
		case "postgres-password":
			ctl.spec.polarisDBSpec.PostgresDetails.Password = ctl.PostgresPassword
		case "postgres-size":
			ctl.spec.polarisDBSpec.PostgresStorageDetails.StorageSize = ctl.PostgresSize
		case "uploadserver-size":
			ctl.spec.polarisDBSpec.UploadServerDetails.Storage.StorageSize = ctl.UploadServerSize
		case "eventstore-size":
			ctl.spec.polarisDBSpec.EventstoreDetails.StorageSize = ctl.EventstoreSize
		case "smtp-host":
			ctl.spec.polarisDBSpec.SMTPDetails.Host = ctl.SMTPHost
		case "smtp-port":
			ctl.spec.polarisDBSpec.SMTPDetails.Port = &ctl.SMTPPort
		case "smtp-username":
			ctl.spec.polarisDBSpec.SMTPDetails.Username = ctl.SMTPUsername
		case "smtp-password":
			ctl.spec.polarisDBSpec.SMTPDetails.Password = ctl.SMTPPassword
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
