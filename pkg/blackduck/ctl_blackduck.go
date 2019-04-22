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

type uid struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

// Ctl type provides functionality for a Blackduck
// for the Synopsysctl tool
type Ctl struct {
	Spec                                  *blackduckv1.BlackduckSpec
	Size                                  string
	Version                               string
	ExposeService                         string
	DbPrototype                           string
	ExternalPostgresPostgresHost          string
	ExternalPostgresPostgresPort          int
	ExternalPostgresPostgresAdmin         string
	ExternalPostgresPostgresUser          string
	ExternalPostgresPostgresSsl           bool
	ExternalPostgresPostgresAdminPassword string
	ExternalPostgresPostgresUserPassword  string
	PvcStorageClass                       string
	LivenessProbes                        bool
	ScanType                              string
	PersistentStorage                     bool
	PVCFile                               string
	CertificateName                       string
	Certificate                           string
	CertificateKey                        string
	ProxyCertificate                      string
	AuthCustomCA                          string
	Type                                  string
	DesiredState                          string
	Environs                              []string
	ImageRegistries                       []string
	ImageUIDMapFile                       string
	LicenseKey                            string
}

// NewBlackduckCtl creates a new Ctl struct
func NewBlackduckCtl() *Ctl {
	return &Ctl{
		Spec:                                  &blackduckv1.BlackduckSpec{},
		Size:                                  "",
		Version:                               "",
		ExposeService:                         "",
		DbPrototype:                           "",
		ExternalPostgresPostgresHost:          "",
		ExternalPostgresPostgresPort:          0,
		ExternalPostgresPostgresAdmin:         "",
		ExternalPostgresPostgresUser:          "",
		ExternalPostgresPostgresSsl:           false,
		ExternalPostgresPostgresAdminPassword: "",
		ExternalPostgresPostgresUserPassword:  "",
		PvcStorageClass:                       "",
		LivenessProbes:                        false,
		ScanType:                              "",
		PersistentStorage:                     false,
		PVCFile:                               "",
		CertificateName:                       "",
		Certificate:                           "",
		CertificateKey:                        "",
		ProxyCertificate:                      "",
		AuthCustomCA:                          "",
		Type:                                  "",
		DesiredState:                          "",
		Environs:                              []string{},
		ImageRegistries:                       []string{},
		ImageUIDMapFile:                       "",
		LicenseKey:                            "",
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
		return fmt.Errorf("Error setting Blackduck Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

func isValidSize(size string) bool {
	switch size {
	case
		"",
		"small",
		"medium",
		"large",
		"xlarge":
		return true
	}
	return false
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags() error {
	if !isValidSize(ctl.Size) {
		return fmt.Errorf("Size must be 'small', 'medium', 'large', or 'xlarge'")
	}
	for _, environ := range ctl.Environs {
		if !strings.Contains(environ, ":") {
			return fmt.Errorf("invalid Environ Format - NAME:VALUE")
		}
	}
	return nil
}

// Constants for Default Specs
const (
	EmptySpec                           string = "empty"
	TemplateSpec                        string = "template"
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

// SwitchSpec switches the Blackduck's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(createBlackduckSpecType string) error {
	switch createBlackduckSpecType {
	case EmptySpec:
		ctl.Spec = &blackduckv1.BlackduckSpec{}
	case TemplateSpec:
		ctl.Spec = crddefaults.GetBlackDuckTemplate()
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
		return fmt.Errorf("Blackduck Spec Type %s is not valid", createBlackduckSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for the OpsSight's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "size - small, medium, large, xlarge")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Blackduck Version")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-service", ctl.ExposeService, "Expose webserver service [LOADBALANCER/NODEPORT/OPENSHIFT]")
	cmd.Flags().StringVar(&ctl.DbPrototype, "db-prototype", ctl.DbPrototype, "Black Duck name to clone the database")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresHost, "external-postgres-host", ctl.ExternalPostgresPostgresHost, "Host for Postgres")
	cmd.Flags().IntVar(&ctl.ExternalPostgresPostgresPort, "external-postgres-port", ctl.ExternalPostgresPostgresPort, "Port for Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdmin, "external-postgres-admin", ctl.ExternalPostgresPostgresAdmin, "Name of Admin for Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUser, "external-postgres-user", ctl.ExternalPostgresPostgresUser, "Username for Postgres")
	cmd.Flags().BoolVar(&ctl.ExternalPostgresPostgresSsl, "external-postgres-ssl", ctl.ExternalPostgresPostgresSsl, "Use SSL for Postgres connection")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", ctl.ExternalPostgresPostgresAdminPassword, "Password for the Postgres Admin")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUserPassword, "external-postgres-user-password", ctl.ExternalPostgresPostgresUserPassword, "Password for a Postgres User")
	cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "Name of Storage Class for the PVC")
	cmd.Flags().BoolVar(&ctl.LivenessProbes, "liveness-probes", ctl.LivenessProbes, "Enable liveness probes")
	cmd.Flags().StringVar(&ctl.ScanType, "scan-type", ctl.ScanType, "Type of Scan artifact. Possible values are Artifacts/Images/Custom")
	cmd.Flags().BoolVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "Enable persistent storage")
	cmd.Flags().StringVar(&ctl.PVCFile, "pvc-file", ctl.PVCFile, "File containing a list of PVC json structs")
	cmd.Flags().StringVar(&ctl.CertificateName, "db-certificate-name", ctl.CertificateName, "Name of Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.Certificate, "certificate-file", ctl.Certificate, "File for the Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.CertificateKey, "certificate-key-file", ctl.CertificateKey, "File for the Black Duck nginx certificate key")
	cmd.Flags().StringVar(&ctl.ProxyCertificate, "proxy-certificate-file", ctl.ProxyCertificate, "File for the Black Duck proxy certificate")
	cmd.Flags().StringVar(&ctl.AuthCustomCA, "auth-custom-ca-file", ctl.AuthCustomCA, "File for the Custom Auth CA for Black Duck")
	cmd.Flags().StringVar(&ctl.Type, "type", ctl.Type, "Type of Blackduck")
	cmd.Flags().StringVar(&ctl.DesiredState, "desired-state", ctl.DesiredState, "Desired state of Blackduck")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of Environment Variables (NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringVar(&ctl.ImageUIDMapFile, "image-uid-map-file", ctl.ImageUIDMapFile, "File containing a map of Container UIDs to Tags")
	cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key for the Knowledge Base")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	flagset.VisitAll(ctl.SetFlag)
}

// SetFlag sets a Blackduck's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED", f.Name)
		switch f.Name {
		case "size":
			ctl.Spec.Size = ctl.Size
		case "version":
			ctl.Spec.Version = ctl.Version
		case "expose-service":
			ctl.Spec.ExposeService = ctl.ExposeService
		case "db-prototype":
			ctl.Spec.DbPrototype = ctl.DbPrototype
		case "external-postgres-host":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresHost = ctl.ExternalPostgresPostgresHost
		case "external-postgres-port":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresPort = ctl.ExternalPostgresPostgresPort
		case "external-postgres-admin":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdmin = ctl.ExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUser = ctl.ExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresSsl = ctl.ExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdminPassword = ctl.ExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUserPassword = ctl.ExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			ctl.Spec.PVCStorageClass = ctl.PvcStorageClass
		case "liveness-probes":
			ctl.Spec.LivenessProbes = ctl.LivenessProbes
		case "scan-type":
			ctl.Spec.ScanType = ctl.ScanType
		case "persistent-storage":
			ctl.Spec.PersistentStorage = ctl.PersistentStorage
		case "pvc-file":
			data, err := util.ReadFileData(ctl.PVCFile)
			if err != nil {
				log.Errorf("failed to read pvc file: %s", err)
			}
			type PVCStructs struct {
				Data []blackduckv1.PVC
			}
			pvcStructs := PVCStructs{Data: []blackduckv1.PVC{}}
			err = json.Unmarshal([]byte(data), &pvcStructs)
			if err != nil {
				log.Errorf("failed to unmarshal pvc structs: %s", err)
				return
			}
			ctl.Spec.PVC = []blackduckv1.PVC{} // clear old values
			for _, pvc := range pvcStructs.Data {
				ctl.Spec.PVC = append(ctl.Spec.PVC, pvc)
			}
		case "db-certificate-name":
			ctl.Spec.CertificateName = ctl.CertificateName
		case "certificate-file":
			data, err := util.ReadFileData(ctl.Certificate)
			if err != nil {
				log.Errorf("failed to read certificate file: %s", err)
			}
			ctl.Spec.Certificate = data
		case "certificate-key-file":
			data, err := util.ReadFileData(ctl.CertificateKey)
			if err != nil {
				log.Errorf("failed to read certificate file: %s", err)
			}
			ctl.Spec.CertificateKey = data
		case "proxy-certificate-file":
			data, err := util.ReadFileData(ctl.ProxyCertificate)
			if err != nil {
				log.Errorf("failed to read certificate file: %s", err)
			}
			ctl.Spec.ProxyCertificate = data
		case "auth-custom-ca-file":
			data, err := util.ReadFileData(ctl.AuthCustomCA)
			if err != nil {
				log.Errorf("failed to read authCustomCA file: %s", err)
			}
			ctl.Spec.AuthCustomCA = data
		case "type":
			ctl.Spec.Type = ctl.Type
		case "desired-state":
			ctl.Spec.DesiredState = ctl.DesiredState
		case "environs":
			ctl.Spec.Environs = ctl.Environs
		case "image-registries":
			ctl.Spec.ImageRegistries = ctl.ImageRegistries
		case "image-uid-map-file":
			data, err := util.ReadFileData(ctl.ImageUIDMapFile)
			if err != nil {
				log.Errorf("failed to read image UID map file: %s", err)
			}
			type UIDStructs struct {
				Data []uid
			}
			uidStructs := UIDStructs{Data: []uid{}}
			err = json.Unmarshal([]byte(data), &uidStructs)
			if err != nil {
				log.Errorf("failed to unmarshal UID Map structs: %s", err)
				return
			}
			ctl.Spec.ImageUIDMap = make(map[string]int64)
			for _, mapStruct := range uidStructs.Data {
				ctl.Spec.ImageUIDMap[mapStruct.Key] = mapStruct.Value
			}
		case "license-key":
			ctl.Spec.LicenseKey = ctl.LicenseKey
		default:
			log.Debugf("Flag %s: Not Found", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED", f.Name)
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
