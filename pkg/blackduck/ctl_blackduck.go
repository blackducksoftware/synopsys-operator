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
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Ctl type provides functionality for a Black Duck
// for the Synopsysctl tool
type Ctl struct {
	Spec                          *blackduckv1.BlackduckSpec
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
	ImageUIDMapFilePath           string
	LicenseKey                    string
	AdminPassword                 string
	PostgresPassword              string
	UserPassword                  string
	EnableBinaryAnalysis          bool
	EnableSourceCodeUpload        bool
	NodeAffinityFilePath          string
}

// NewBlackDuckCtl creates a new Ctl struct
func NewBlackDuckCtl() *Ctl {
	return &Ctl{
		Spec: &blackduckv1.BlackduckSpec{},
	}
}

// GetSpec returns the Spec for the resource
func (ctl *Ctl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *Ctl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(blackduckv1.BlackduckSpec)
	if !ok {
		return fmt.Errorf("error in Black Duck spec conversion")
	}
	ctl.Spec = &convertedSpec
	return nil
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

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags(flagset *pflag.FlagSet) error {
	if !isValidSize(ctl.Size) {
		return fmt.Errorf("size must be 'small', 'medium', 'large', or 'x-large'")
	}
	for _, environ := range ctl.Environs {
		if !strings.Contains(environ, ":") {
			return fmt.Errorf("invalid environ format - NAME:VALUE")
		}
	}

	if ctl.ExternalPostgresHost == "" {
		// user is explicitly required to set the postgres passwords for: 'admin', 'postgres', and 'user'
		cobra.MarkFlagRequired(flagset, "admin-password")
		cobra.MarkFlagRequired(flagset, "postgres-password")
		cobra.MarkFlagRequired(flagset, "user-password")
	} else {
		// require all external-postgres parameters
		cobra.MarkFlagRequired(flagset, "external-postgres-host")
		cobra.MarkFlagRequired(flagset, "external-postgres-port")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin")
		cobra.MarkFlagRequired(flagset, "external-postgres-user")
		cobra.MarkFlagRequired(flagset, "external-postgres-ssl")
		cobra.MarkFlagRequired(flagset, "external-postgres-admin-password")
		cobra.MarkFlagRequired(flagset, "external-postgres-user-password")
	}

	setStateToDbMigrate := flagset.Lookup("migration-mode").Changed
	if val, _ := flagset.GetBool("migration-mode"); !val && setStateToDbMigrate {
		return fmt.Errorf("--migration-mode cannot be set to false")
	}

	return nil
}

// Constants for Default Specs
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

// SwitchSpec switches the Black Duck's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(createBlackDuckSpecType string) error {
	switch createBlackDuckSpecType {
	case EmptySpec:
		ctl.Spec = &blackduckv1.BlackduckSpec{}
	case PersistentStorageLatestSpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultPersistentStorageLatest()
	case PersistentStorageV1Spec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultPersistentStorageV1()
	case ExternalPersistentStorageLatestSpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultExternalPersistentStorageLatest()
	case ExternalPersistentStorageV1Spec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultExternalPersistentStorageV1()
	case BDBASpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultBDBA()
	case EphemeralSpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultEphemeral()
	case EphemeralCustomAuthCASpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultEphemeralCustomAuthCA()
	case ExternalDBSpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultExternalDB()
	case IPV6DisabledSpec:
		ctl.Spec = crddefaults.GetBlackDuckDefaultIPV6Disabled()
	default:
		return fmt.Errorf("Black Duck spec type '%s' is not valid", createBlackDuckSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for Black Duck's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	if master {
		cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "Name of Storage Class for the PVC")
		cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Black Duck has persistent storage [true|false]")
		cmd.Flags().StringVar(&ctl.PVCFilePath, "pvc-file-path", ctl.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	}
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "Size of Black Duck [small|medium|large|x-large]")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Black Duck")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type of Black Duck webserver's user interface [LOADBALANCER|NODEPORT|OPENSHIFT]")
	cmd.Flags().StringVar(&ctl.DbPrototype, "db-prototype", ctl.DbPrototype, "Black Duck name to clone the database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresHost, "external-postgres-host", ctl.ExternalPostgresHost, "Host of external Postgres")
	cmd.Flags().IntVar(&ctl.ExternalPostgresPort, "external-postgres-port", ctl.ExternalPostgresPort, "Port of external Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresAdmin, "external-postgres-admin", ctl.ExternalPostgresAdmin, "Name of 'admin' of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresUser, "external-postgres-user", ctl.ExternalPostgresUser, "Name of 'user' of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresSsl, "external-postgres-ssl", ctl.ExternalPostgresSsl, "If true, Black Duck uses SSL for external Postgres connection [true|false]")
	cmd.Flags().StringVar(&ctl.ExternalPostgresAdminPassword, "external-postgres-admin-password", ctl.ExternalPostgresAdminPassword, "'admin' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresUserPassword, "external-postgres-user-password", ctl.ExternalPostgresUserPassword, "'user' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.LivenessProbes, "liveness-probes", ctl.LivenessProbes, "If true, Black Duck uses liveness probes [true|false]")
	cmd.Flags().StringVar(&ctl.PostgresClaimSize, "postgres-claim-size", ctl.PostgresClaimSize, "Size of the blackduck-postgres PVC")
	cmd.Flags().StringVar(&ctl.CertificateName, "certificate-name", ctl.CertificateName, "Name of Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.CertificateFilePath, "certificate-file-path", ctl.CertificateFilePath, "Absolute path to a file for the Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.CertificateKeyFilePath, "certificate-key-file-path", ctl.CertificateKeyFilePath, "Absolute path to a file for the Black Duck nginx certificate key")
	cmd.Flags().StringVar(&ctl.ProxyCertificateFilePath, "proxy-certificate-file-path", ctl.ProxyCertificateFilePath, "Absolute path to a file for the Black Duck proxy server’s Certificate Authority (CA)")
	cmd.Flags().StringVar(&ctl.AuthCustomCAFilePath, "auth-custom-ca-file-path", ctl.AuthCustomCAFilePath, "Absolute path to a file for the Custom Auth CA for Black Duck")
	cmd.Flags().StringVar(&ctl.Type, "type", ctl.Type, "Type of Black Duck")
	cmd.Flags().StringVar(&ctl.DesiredState, "desired-state", ctl.DesiredState, "Desired state of Black Duck")
	cmd.Flags().BoolVar(&ctl.MigrationMode, "migration-mode", ctl.MigrationMode, "Create Black Duck in the database-migration state")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of Environment Variables (NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringVar(&ctl.ImageUIDMapFilePath, "image-uid-map-file-path", ctl.ImageUIDMapFilePath, "Absolute path to a file containing a map of Container UIDs to Tags")
	cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key of Black Duck")
	cmd.Flags().StringVar(&ctl.AdminPassword, "admin-password", ctl.AdminPassword, "'admin' password of Postgres database")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "'postgres' password of Postgres database")
	cmd.Flags().StringVar(&ctl.UserPassword, "user-password", ctl.UserPassword, "'user' password of Postgres database")
	cmd.Flags().BoolVar(&ctl.EnableBinaryAnalysis, "enable-binary-analysis", ctl.EnableBinaryAnalysis, "If true, enable binary analysis")
	cmd.Flags().BoolVar(&ctl.EnableSourceCodeUpload, "enable-source-code-upload", ctl.EnableSourceCodeUpload, "If true, enable source code upload")
	cmd.Flags().StringVar(&ctl.NodeAffinityFilePath, "node-affinity-file-path", ctl.NodeAffinityFilePath, "Absolute path to a file containing a list of node affinities")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("desired-state", "desired-state flag is deprecated and will be removed by the next release")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	// Update spec fields with flags
	flagset.VisitAll(ctl.SetFlag)
}

// SetFlag sets a Black Duck's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "size":
			ctl.Spec.Size = ctl.Size
		case "version":
			ctl.Spec.Version = ctl.Version
		case "expose-ui":
			ctl.Spec.ExposeService = ctl.ExposeService
		case "db-prototype":
			ctl.Spec.DbPrototype = ctl.DbPrototype
		case "external-postgres-host":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresHost = ctl.ExternalPostgresHost
		case "external-postgres-port":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresPort = ctl.ExternalPostgresPort
		case "external-postgres-admin":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdmin = ctl.ExternalPostgresAdmin
		case "external-postgres-user":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUser = ctl.ExternalPostgresUser
		case "external-postgres-ssl":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresSsl = strings.ToUpper(ctl.ExternalPostgresSsl) == "TRUE"
		case "external-postgres-admin-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdminPassword = crddefaults.Base64Encode([]byte(ctl.ExternalPostgresAdminPassword))
		case "external-postgres-user-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUserPassword = crddefaults.Base64Encode([]byte(ctl.ExternalPostgresUserPassword))
		case "pvc-storage-class":
			ctl.Spec.PVCStorageClass = ctl.PvcStorageClass
		case "liveness-probes":
			ctl.Spec.LivenessProbes = strings.ToUpper(ctl.LivenessProbes) == "TRUE"
		case "persistent-storage":
			ctl.Spec.PersistentStorage = strings.ToUpper(ctl.PersistentStorage) == "TRUE"
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
			ctl.Spec.PVC = pvcs
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
			ctl.Spec.NodeAffinities = nodeAffinities
		case "postgres-claim-size":
			for i := range ctl.Spec.PVC {
				if ctl.Spec.PVC[i].Name == "blackduck-postgres" { // update claim size and return
					ctl.Spec.PVC[i].Size = ctl.PostgresClaimSize
					return
				}
			}
			ctl.Spec.PVC = append(ctl.Spec.PVC, blackduckv1.PVC{Name: "blackduck-postgres", Size: ctl.PostgresClaimSize}) // add postgres PVC if doesn't exist
		case "certificate-name":
			ctl.Spec.CertificateName = ctl.CertificateName
		case "certificate-file-path":
			data, err := util.ReadFileData(ctl.CertificateFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.Spec.Certificate = data
		case "certificate-key-file-path":
			data, err := util.ReadFileData(ctl.CertificateKeyFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.Spec.CertificateKey = data
		case "proxy-certificate-file-path":
			data, err := util.ReadFileData(ctl.ProxyCertificateFilePath)
			if err != nil {
				log.Errorf("failed to read certificate file: %+v", err)
				return
			}
			ctl.Spec.ProxyCertificate = data
		case "auth-custom-ca-file-path":
			data, err := util.ReadFileData(ctl.AuthCustomCAFilePath)
			if err != nil {
				log.Errorf("failed to read authCustomCA file: %+v", err)
				return
			}
			ctl.Spec.AuthCustomCA = data
		case "type":
			ctl.Spec.Type = ctl.Type
		case "desired-state":
			ctl.Spec.DesiredState = ctl.DesiredState
		case "migration-mode":
			if ctl.MigrationMode {
				ctl.Spec.DesiredState = "DbMigrate"
			}
		case "environs":
			ctl.Spec.Environs = ctl.Environs
		case "image-registries":
			ctl.Spec.ImageRegistries = ctl.ImageRegistries
		case "image-uid-map-file-path":
			data, err := util.ReadFileData(ctl.ImageUIDMapFilePath)
			if err != nil {
				log.Errorf("failed to read image UID map file: %+v", err)
				return
			}
			uidMap := map[string]int64{}
			err = json.Unmarshal([]byte(data), &uidMap)
			if err != nil {
				log.Errorf("failed to unmarshal UID Map structs: %+v", err)
				return
			}
			ctl.Spec.ImageUIDMap = uidMap
		case "license-key":
			ctl.Spec.LicenseKey = ctl.LicenseKey
		case "admin-password":
			ctl.Spec.AdminPassword = crddefaults.Base64Encode([]byte(ctl.AdminPassword))
		case "postgres-password":
			ctl.Spec.PostgresPassword = crddefaults.Base64Encode([]byte(ctl.PostgresPassword))
		case "user-password":
			ctl.Spec.UserPassword = crddefaults.Base64Encode([]byte(ctl.UserPassword))
		case "enable-binary-analysis":
			if ctl.EnableBinaryAnalysis {
				ctl.Spec.Environs = util.MergeEnvSlices([]string{"USE_BINARY_UPLOADS:1"}, ctl.Spec.Environs)
			} else {
				ctl.Spec.Environs = util.MergeEnvSlices([]string{"USE_BINARY_UPLOADS:0"}, ctl.Spec.Environs)
			}
		case "enable-source-code-upload":
			if ctl.EnableSourceCodeUpload {
				ctl.Spec.Environs = util.MergeEnvSlices([]string{"ENABLE_SOURCE_UPLOADS:true"}, ctl.Spec.Environs)
			} else {
				ctl.Spec.Environs = util.MergeEnvSlices([]string{"ENABLE_SOURCE_UPLOADS:false"}, ctl.Spec.Environs)
			}
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}

// SpecIsValid verifies the spec has necessary fields to deploy
func (ctl *Ctl) SpecIsValid() (bool, error) {
	return true, nil
}

// CanUpdate checks if a user has permission to modify based on the spec
func (ctl *Ctl) CanUpdate() (bool, error) {
	return true, nil
}
