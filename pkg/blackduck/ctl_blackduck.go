/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package blackduck

import (
	"encoding/json"
	"fmt"
	"strings"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create a Black Duck CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making a Black Duck Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	blackDuckSpec                 *blackduckv1.BlackduckSpec
	Size                          string
	Version                       string
	ExposeService                 string
	DbPrototype                   string
	ExternalPostgresHost          string
	ExternalPostgresPort          int
	ExternalPostgresAdmin         string
	ExternalPostgresUser          string
	ExternalPostgresSsl           string
	ExternalPostgresAdminPassword string
	ExternalPostgresUserPassword  string
	PvcStorageClass               string
	LivenessProbes                string
	PersistentStorage             string
	PVCFilePath                   string
	PostgresClaimSize             string
	CertificateName               string
	CertificateFilePath           string
	CertificateKeyFilePath        string
	ProxyCertificateFilePath      string
	AuthCustomCAFilePath          string
	Type                          string
	DesiredState                  string
	MigrationMode                 bool
	Environs                      []string
	ImageRegistries               []string
	LicenseKey                    string
	AdminPassword                 string
	PostgresPassword              string
	UserPassword                  string
	EnableBinaryAnalysis          bool
	EnableSourceCodeUpload        bool
	NodeAffinityFilePath          string
}

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		blackDuckSpec: &blackduckv1.BlackduckSpec{},
	}
}

// GetCRSpec returns a pointer to the BlackDuckSpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return *ctl.blackDuckSpec
}

// SetCRSpec sets the blackDuckSpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(blackduckv1.BlackduckSpec)
	if !ok {
		return fmt.Errorf("error in Black Duck spec conversion")
	}
	ctl.blackDuckSpec = &convertedSpec
	return nil
}

// Constants for predefined specs
const (
	EmptySpec                           string = "empty"
	PersistentStorageLatestSpec         string = "persistentStorageLatest"
	PersistentStorageV1Spec             string = "persistentStorageV1"
	ExternalPersistentStorageLatestSpec string = "externalPersistentStorageLatest"
	ExternalPersistentStorageV1Spec     string = "externalPersistentStorageV1"
	BDBASpec                            string = "bdba"
	EphemeralSpec                       string = "ephemeral"
	EphemeralCustomAuthCASpec           string = "ephemeralCustomAuthCA"
	ExternalDBSpec                      string = "externalDB"
	IPV6DisabledSpec                    string = "IPV6Disabled"
)

// SetPredefinedCRSpec sets the blackDuckSpec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	switch specType {
	case EmptySpec:
		ctl.blackDuckSpec = &blackduckv1.BlackduckSpec{}
	case PersistentStorageLatestSpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultPersistentStorageLatest()
	case PersistentStorageV1Spec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultPersistentStorageV1()
	case ExternalPersistentStorageLatestSpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultExternalPersistentStorageLatest()
	case ExternalPersistentStorageV1Spec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultExternalPersistentStorageV1()
	case BDBASpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultBDBA()
	case EphemeralSpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultEphemeral()
	case EphemeralCustomAuthCASpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultEphemeralCustomAuthCA()
	case ExternalDBSpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultExternalDB()
	case IPV6DisabledSpec:
		ctl.blackDuckSpec = util.GetBlackDuckDefaultIPV6Disabled()
	default:
		return fmt.Errorf("Black Duck spec type '%s' is not valid", specType)
	}
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for BlackDuck's Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	if master {
		cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "Name of Storage Class for the PVC")
		cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Black Duck has persistent storage (true|false)")
		cmd.Flags().StringVar(&ctl.PVCFilePath, "pvc-file-path", ctl.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	}
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "Size of Black Duck (small|medium|large|x-large)")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Black Duck")
	if master {
		cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", util.NONE, "Service type of Black Duck webserver's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	} else {
		cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type of Black Duck webserver's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	}
	if !strings.Contains(cmd.CommandPath(), "native") {
		cmd.Flags().StringVar(&ctl.DbPrototype, "db-prototype", ctl.DbPrototype, "Black Duck name to clone the database")
	}
	cmd.Flags().StringVar(&ctl.ExternalPostgresHost, "external-postgres-host", ctl.ExternalPostgresHost, "Host of external Postgres")
	cmd.Flags().IntVar(&ctl.ExternalPostgresPort, "external-postgres-port", ctl.ExternalPostgresPort, "Port of external Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresAdmin, "external-postgres-admin", ctl.ExternalPostgresAdmin, "Name of 'admin' of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresUser, "external-postgres-user", ctl.ExternalPostgresUser, "Name of 'user' of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresSsl, "external-postgres-ssl", ctl.ExternalPostgresSsl, "If true, Black Duck uses SSL for external Postgres connection (true|false)")
	cmd.Flags().StringVar(&ctl.ExternalPostgresAdminPassword, "external-postgres-admin-password", ctl.ExternalPostgresAdminPassword, "'admin' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresUserPassword, "external-postgres-user-password", ctl.ExternalPostgresUserPassword, "'user' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.LivenessProbes, "liveness-probes", ctl.LivenessProbes, "If true, Black Duck uses liveness probes (true|false)")
	cmd.Flags().StringVar(&ctl.PostgresClaimSize, "postgres-claim-size", ctl.PostgresClaimSize, "Size of the blackduck-postgres PVC")
	cmd.Flags().StringVar(&ctl.CertificateName, "certificate-name", ctl.CertificateName, "Name of Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.CertificateFilePath, "certificate-file-path", ctl.CertificateFilePath, "Absolute path to a file for the Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.CertificateKeyFilePath, "certificate-key-file-path", ctl.CertificateKeyFilePath, "Absolute path to a file for the Black Duck nginx certificate key")
	cmd.Flags().StringVar(&ctl.ProxyCertificateFilePath, "proxy-certificate-file-path", ctl.ProxyCertificateFilePath, "Absolute path to a file for the Black Duck proxy server’s Certificate Authority (CA)")
	cmd.Flags().StringVar(&ctl.AuthCustomCAFilePath, "auth-custom-ca-file-path", ctl.AuthCustomCAFilePath, "Absolute path to a file for the Custom Auth CA for Black Duck")
	if !strings.Contains(cmd.CommandPath(), "native") {
		cmd.Flags().StringVar(&ctl.Type, "type", ctl.Type, "Type of Black Duck")
	}
	cmd.Flags().StringVar(&ctl.DesiredState, "desired-state", ctl.DesiredState, "Desired state of Black Duck")
	if !strings.Contains(cmd.CommandPath(), "native") {
		cmd.Flags().BoolVar(&ctl.MigrationMode, "migration-mode", ctl.MigrationMode, "Create Black Duck in the database-migration state")
	}
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of Environment Variables (NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	if !strings.Contains(cmd.CommandPath(), "native") {
		cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key of Black Duck")
	}
	cmd.Flags().StringVar(&ctl.AdminPassword, "admin-password", ctl.AdminPassword, "'admin' password of Postgres database")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "'postgres' password of Postgres database")
	cmd.Flags().StringVar(&ctl.UserPassword, "user-password", ctl.UserPassword, "'user' password of Postgres database")
	cmd.Flags().BoolVar(&ctl.EnableBinaryAnalysis, "enable-binary-analysis", ctl.EnableBinaryAnalysis, "If true, enable binary analysis")
	cmd.Flags().BoolVar(&ctl.EnableSourceCodeUpload, "enable-source-code-upload", ctl.EnableSourceCodeUpload, "If true, enable source code upload")
	cmd.Flags().StringVar(&ctl.NodeAffinityFilePath, "node-affinity-file-path", ctl.NodeAffinityFilePath, "Absolute path to a file containing a list of node affinities")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("desired-state", "desired-state flag is deprecated and will be removed by the next release")
}

func isValidSize(size string) bool {
	switch strings.ToLower(size) {
	case
		"",
		"small",
		"medium",
		"large",
		"x-large":
		return true
	}
	return false
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the blackDuckSpec
func (ctl *CRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	if FlagWasSet(flagset, "size") {
		if !isValidSize(ctl.Size) {
			return fmt.Errorf("size must be 'small', 'medium', 'large' or 'x-large'")
		}
	}
	if FlagWasSet(flagset, "expose-ui") {
		isValid := util.IsExposeServiceValid(ctl.ExposeService)
		if !isValid {
			return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "environs") {
		for _, environ := range ctl.Environs {
			if !strings.Contains(environ, ":") {
				return fmt.Errorf("invalid environ format - NAME:VALUE")
			}
		}
	}
	if FlagWasSet(flagset, "migration-mode") {
		if val, _ := flagset.GetBool("migration-mode"); !val {
			return fmt.Errorf("--migration-mode cannot be set to false")
		}
	}
	return nil
}

// FlagWasSet returns true if a flag was changed and it exists, otherwise it returns false
func FlagWasSet(flagset *pflag.FlagSet, flagName string) bool {
	if flagset.Lookup(flagName) != nil && flagset.Lookup(flagName).Changed {
		return true
	}
	return false
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the blackDuckSpec with the value that's stored
// in the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return *ctl.blackDuckSpec, nil
}

// SetCRSpecFieldByFlag updates a field in the blackDuckSpec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "size":
			ctl.blackDuckSpec.Size = ctl.Size
		case "version":
			ctl.blackDuckSpec.Version = ctl.Version
		case "expose-ui":
			ctl.blackDuckSpec.ExposeService = ctl.ExposeService
		case "db-prototype":
			ctl.blackDuckSpec.DbPrototype = ctl.DbPrototype
		case "external-postgres-host":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresHost = ctl.ExternalPostgresHost
		case "external-postgres-port":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresPort = ctl.ExternalPostgresPort
		case "external-postgres-admin":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresAdmin = ctl.ExternalPostgresAdmin
		case "external-postgres-user":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresUser = ctl.ExternalPostgresUser
		case "external-postgres-ssl":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresSsl = strings.ToUpper(ctl.ExternalPostgresSsl) == "TRUE"
		case "external-postgres-admin-password":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresAdminPassword = util.Base64Encode([]byte(ctl.ExternalPostgresAdminPassword))
		case "external-postgres-user-password":
			if ctl.blackDuckSpec.ExternalPostgres == nil {
				ctl.blackDuckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.blackDuckSpec.ExternalPostgres.PostgresUserPassword = util.Base64Encode([]byte(ctl.ExternalPostgresUserPassword))
		case "pvc-storage-class":
			ctl.blackDuckSpec.PVCStorageClass = ctl.PvcStorageClass
		case "liveness-probes":
			ctl.blackDuckSpec.LivenessProbes = strings.ToUpper(ctl.LivenessProbes) == "TRUE"
		case "persistent-storage":
			ctl.blackDuckSpec.PersistentStorage = strings.ToUpper(ctl.PersistentStorage) == "TRUE"
		case "pvc-file-path":
			data, err := util.ReadFileData(ctl.PVCFilePath)
			if err != nil {
				log.Errorf("failed to read pvc file: %+v", err)
				return
			}
			pvcs := []blackduckv1.PVC{}
			err = json.Unmarshal([]byte(data), &pvcs)
			if err != nil {
				log.Errorf("failed to unmarshal pvc structs: %+v", err)
				return
			}
			ctl.blackDuckSpec.PVC = pvcs
		case "node-affinity-file-path":
			data, err := util.ReadFileData(ctl.NodeAffinityFilePath)
			if err != nil {
				log.Errorf("failed to read node affinity file: %+v", err)
				return
			}
			nodeAffinities := map[string][]blackduckv1.NodeAffinity{}
			err = json.Unmarshal([]byte(data), &nodeAffinities)
			if err != nil {
				log.Errorf("failed to unmarshal node affinities: %+v", err)
				return
			}
			ctl.blackDuckSpec.NodeAffinities = nodeAffinities
		case "postgres-claim-size":
			for i := range ctl.blackDuckSpec.PVC {
				if ctl.blackDuckSpec.PVC[i].Name == "blackduck-postgres" { // update claim size and return
					ctl.blackDuckSpec.PVC[i].Size = ctl.PostgresClaimSize
					return
				}
			}
			ctl.blackDuckSpec.PVC = append(ctl.blackDuckSpec.PVC, blackduckv1.PVC{Name: "blackduck-postgres", Size: ctl.PostgresClaimSize}) // add postgres PVC if doesn't exist
		case "certificate-name":
			ctl.blackDuckSpec.CertificateName = ctl.CertificateName
		case "certificate-file-path":
			data, err := util.ReadFileData(ctl.CertificateFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.blackDuckSpec.Certificate = data
		case "certificate-key-file-path":
			data, err := util.ReadFileData(ctl.CertificateKeyFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.blackDuckSpec.CertificateKey = data
		case "proxy-certificate-file-path":
			data, err := util.ReadFileData(ctl.ProxyCertificateFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.blackDuckSpec.ProxyCertificate = data
		case "auth-custom-ca-file-path":
			data, err := util.ReadFileData(ctl.AuthCustomCAFilePath)
			if err != nil {
				log.Errorf("failed to read authCustomCA file: %+v", err)
				return
			}
			ctl.blackDuckSpec.AuthCustomCA = data
		case "type":
			ctl.blackDuckSpec.Type = ctl.Type
		case "desired-state":
			ctl.blackDuckSpec.DesiredState = ctl.DesiredState
		case "migration-mode":
			if ctl.MigrationMode {
				ctl.blackDuckSpec.DesiredState = "DbMigrate"
			}
		case "environs":
			ctl.blackDuckSpec.Environs = ctl.Environs
		case "image-registries":
			ctl.blackDuckSpec.ImageRegistries = ctl.ImageRegistries
		case "license-key":
			ctl.blackDuckSpec.LicenseKey = ctl.LicenseKey
		case "admin-password":
			ctl.blackDuckSpec.AdminPassword = util.Base64Encode([]byte(ctl.AdminPassword))
		case "postgres-password":
			ctl.blackDuckSpec.PostgresPassword = util.Base64Encode([]byte(ctl.PostgresPassword))
		case "user-password":
			ctl.blackDuckSpec.UserPassword = util.Base64Encode([]byte(ctl.UserPassword))
		case "enable-binary-analysis":
			if ctl.EnableBinaryAnalysis {
				ctl.blackDuckSpec.Environs = util.MergeEnvSlices([]string{"USE_BINARY_UPLOADS:1"}, ctl.blackDuckSpec.Environs)
			} else {
				ctl.blackDuckSpec.Environs = util.MergeEnvSlices([]string{"USE_BINARY_UPLOADS:0"}, ctl.blackDuckSpec.Environs)
			}
		case "enable-source-code-upload":
			if ctl.EnableSourceCodeUpload {
				ctl.blackDuckSpec.Environs = util.MergeEnvSlices([]string{"ENABLE_SOURCE_UPLOADS:true"}, ctl.blackDuckSpec.Environs)
			} else {
				ctl.blackDuckSpec.Environs = util.MergeEnvSlices([]string{"ENABLE_SOURCE_UPLOADS:false"}, ctl.blackDuckSpec.Environs)
			}
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
