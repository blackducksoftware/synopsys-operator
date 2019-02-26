// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"encoding/json"
	"fmt"
	"strings"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type uid struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

// BlackduckCtl type provides functionality for a Blackduck
// for the Synopsysctl tool
type BlackduckCtl struct {
	Spec                                  *blackduckv1.BlackduckSpec
	Size                                  string
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
	PVCJSONSlice                          []string
	CertificateName                       string
	Certificate                           string
	CertificateKey                        string
	ProxyCertificate                      string
	Type                                  string
	DesiredState                          string
	Environs                              []string
	ImageRegistries                       []string
	ImageUIDMapJSONSlice                  []string
	LicenseKey                            string
}

// NewBlackduckCtl creates a new BlackduckCtl struct
func NewBlackduckCtl() *BlackduckCtl {
	return &BlackduckCtl{
		Spec:                                  &blackduckv1.BlackduckSpec{},
		Size:                                  "",
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
		PVCJSONSlice:                          []string{},
		CertificateName:                       "",
		Certificate:                           "",
		CertificateKey:                        "",
		ProxyCertificate:                      "",
		Type:                                  "",
		DesiredState:                          "",
		Environs:                              []string{},
		ImageRegistries:                       []string{},
		ImageUIDMapJSONSlice:                  []string{},
		LicenseKey:                            "",
	}
}

// GetSpec returns the Spec for the resource
func (ctl *BlackduckCtl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *BlackduckCtl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(blackduckv1.BlackduckSpec)
	if !ok {
		return fmt.Errorf("Error setting Blackduck Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *BlackduckCtl) CheckSpecFlags() error {
	if ctl.Size != "" && ctl.Size != "small" && ctl.Size != "medium" && ctl.Size != "large" {
		return fmt.Errorf("Size must be 'small', 'medium', or 'large'")
	}
	for _, pvcJSON := range ctl.PVCJSONSlice {
		pvc := &blackduckv1.PVC{}
		err := json.Unmarshal([]byte(pvcJSON), pvc)
		if err != nil {
			return fmt.Errorf("Invalid format for PVC")
		}
	}
	for _, environ := range ctl.Environs {
		if !strings.Contains(environ, ":") {
			return fmt.Errorf("Invalid Environ Format - NAME:VALUE")
		}
	}
	for _, uidJSON := range ctl.ImageUIDMapJSONSlice {
		uidStruct := &uid{}
		err := json.Unmarshal([]byte(uidJSON), uidStruct)
		if err != nil {
			return fmt.Errorf("Invalid format for Image UID")
		}
	}
	return nil
}

// SwitchSpec switches the Blackduck's Spec to a different predefined spec
func (ctl *BlackduckCtl) SwitchSpec(createBlackduckSpecType string) error {
	switch createBlackduckSpecType {
	case "empty":
		ctl.Spec = &blackduckv1.BlackduckSpec{}
	case "persistentStorage":
		ctl.Spec = crddefaults.GetHubDefaultPersistentStorage()
	case "default":
		ctl.Spec = crddefaults.GetHubDefaultValue()
	default:
		return fmt.Errorf("Blackduck Spec Type %s does not match: empty, persistentStorage, default", createBlackduckSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for the OpsSight's Spec to the command
func (ctl *BlackduckCtl) AddSpecFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctl.Size, "spec-size", ctl.Size, " size - small, medium, large")
	cmd.Flags().StringVar(&ctl.DbPrototype, "spec-db-prototype", ctl.DbPrototype, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresHost, "spec-external-postgres-host", ctl.ExternalPostgresPostgresHost, "TODO")
	cmd.Flags().IntVar(&ctl.ExternalPostgresPostgresPort, "spec-external-postgres-port", ctl.ExternalPostgresPostgresPort, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdmin, "spec-external-postgres-admin", ctl.ExternalPostgresPostgresAdmin, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUser, "spec-external-postgres-user", ctl.ExternalPostgresPostgresUser, "TODO")
	cmd.Flags().BoolVar(&ctl.ExternalPostgresPostgresSsl, "spec-external-postgres-ssl", ctl.ExternalPostgresPostgresSsl, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdminPassword, "spec-external-postgres-admin-password", ctl.ExternalPostgresPostgresAdminPassword, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUserPassword, "spec-external-postgres-user-password", ctl.ExternalPostgresPostgresUserPassword, "TODO")
	cmd.Flags().StringVar(&ctl.PvcStorageClass, "spec-pvc-storage-class", ctl.PvcStorageClass, "TODO")
	cmd.Flags().BoolVar(&ctl.LivenessProbes, "spec-liveness-probes", ctl.LivenessProbes, "Enable liveness probes")
	cmd.Flags().StringVar(&ctl.ScanType, "spec-scan-type", ctl.ScanType, "TODO")
	cmd.Flags().BoolVar(&ctl.PersistentStorage, "spec-persistent-storage", ctl.PersistentStorage, "Enable persistent storage")
	cmd.Flags().StringSliceVar(&ctl.PVCJSONSlice, "spec-pvc", ctl.PVCJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.CertificateName, "spec-db-certificate-name", ctl.CertificateName, "TODO")
	cmd.Flags().StringVar(&ctl.Certificate, "spec-certificate", ctl.Certificate, "TODO")
	cmd.Flags().StringVar(&ctl.CertificateKey, "spec-certificate-key", ctl.CertificateKey, "TODO")
	cmd.Flags().StringVar(&ctl.ProxyCertificate, "spec-proxy-certificate", ctl.ProxyCertificate, "TODO")
	cmd.Flags().StringVar(&ctl.Type, "spec-type", ctl.Type, "TODO")
	cmd.Flags().StringVar(&ctl.DesiredState, "spec-desired-state", ctl.DesiredState, "TODO")
	cmd.Flags().StringSliceVar(&ctl.Environs, "spec-environs", ctl.Environs, "TODO")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "spec-image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringSliceVar(&ctl.ImageUIDMapJSONSlice, "spec-image-uid-map", ctl.ImageUIDMapJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.LicenseKey, "spec-license-key", ctl.LicenseKey, "TODO")
}

// SetFlags sets the Blackduck's Spec if a flag was changed
func (ctl *BlackduckCtl) SetFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s:   CHANGED\n", f.Name)
		switch f.Name {
		case "spec-size":
			ctl.Spec.Size = ctl.Size
		case "spec-db-prototype":
			ctl.Spec.DbPrototype = ctl.DbPrototype
		case "spec-external-postgres-host":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresHost = ctl.ExternalPostgresPostgresHost
		case "spec-external-postgres-port":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresPort = ctl.ExternalPostgresPostgresPort
		case "spec-external-postgres-admin":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdmin = ctl.ExternalPostgresPostgresAdmin
		case "spec-external-postgres-user":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUser = ctl.ExternalPostgresPostgresUser
		case "spec-external-postgres-ssl":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresSsl = ctl.ExternalPostgresPostgresSsl
		case "spec-external-postgres-admin-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdminPassword = ctl.ExternalPostgresPostgresAdminPassword
		case "spec-external-postgres-user-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUserPassword = ctl.ExternalPostgresPostgresUserPassword
		case "spec-pvc-storage-class":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.PVCStorageClass = ctl.PvcStorageClass
		case "spec-liveness-probes":
			ctl.Spec.LivenessProbes = ctl.LivenessProbes
		case "spec-scan-type":
			ctl.Spec.ScanType = ctl.ScanType
		case "spec-persistent-storage":
			ctl.Spec.PersistentStorage = ctl.PersistentStorage
		case "spec-pvc":
			for _, pvcJSON := range ctl.PVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				ctl.Spec.PVC = append(ctl.Spec.PVC, *pvc)
			}
		case "spec-db-certificate-name":
			ctl.Spec.CertificateName = ctl.CertificateName
		case "spec-certificate":
			ctl.Spec.Certificate = ctl.Certificate
		case "spec-certificate-key":
			ctl.Spec.CertificateKey = ctl.CertificateKey
		case "spec-proxy-certificate":
			ctl.Spec.ProxyCertificate = ctl.ProxyCertificate
		case "spec-type":
			ctl.Spec.Type = ctl.Type
		case "spec-desired-state":
			ctl.Spec.DesiredState = ctl.DesiredState
		case "spec-environs":
			ctl.Spec.Environs = ctl.Environs
		case "spec-image-registries":
			ctl.Spec.ImageRegistries = ctl.ImageRegistries
		case "spec-image-uid-map":
			ctl.Spec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range ctl.ImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				ctl.Spec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "spec-license-key":
			ctl.Spec.LicenseKey = ctl.LicenseKey
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}
