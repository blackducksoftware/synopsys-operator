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

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// BlackduckCtl type provides functionality for a Blackduck
// for the Synopsysctl tool
type BlackduckCtl struct {
	Spec                                           *blackduckv1.BlackduckSpec
	BlackduckSize                                  string
	BlackduckDbPrototype                           string
	BlackduckExternalPostgresPostgresHost          string
	BlackduckExternalPostgresPostgresPort          int
	BlackduckExternalPostgresPostgresAdmin         string
	BlackduckExternalPostgresPostgresUser          string
	BlackduckExternalPostgresPostgresSsl           bool
	BlackduckExternalPostgresPostgresAdminPassword string
	BlackduckExternalPostgresPostgresUserPassword  string
	BlackduckPvcStorageClass                       string
	BlackduckLivenessProbes                        bool
	BlackduckScanType                              string
	BlackduckPersistentStorage                     bool
	BlackduckPVCJSONSlice                          []string
	BlackduckCertificateName                       string
	BlackduckCertificate                           string
	BlackduckCertificateKey                        string
	BlackduckProxyCertificate                      string
	BlackduckType                                  string
	BlackduckDesiredState                          string
	BlackduckEnvirons                              []string
	BlackduckImageRegistries                       []string
	BlackduckImageUIDMapJSONSlice                  []string
	BlackduckLicenseKey                            string
}

// NewBlackduckCtl creates a new BlackduckCtl struct
func NewBlackduckCtl() *BlackduckCtl {
	return &BlackduckCtl{
		Spec:                                           &blackduckv1.BlackduckSpec{},
		BlackduckSize:                                  "",
		BlackduckDbPrototype:                           "",
		BlackduckExternalPostgresPostgresHost:          "",
		BlackduckExternalPostgresPostgresPort:          0,
		BlackduckExternalPostgresPostgresAdmin:         "",
		BlackduckExternalPostgresPostgresUser:          "",
		BlackduckExternalPostgresPostgresSsl:           false,
		BlackduckExternalPostgresPostgresAdminPassword: "",
		BlackduckExternalPostgresPostgresUserPassword:  "",
		BlackduckPvcStorageClass:                       "",
		BlackduckLivenessProbes:                        false,
		BlackduckScanType:                              "",
		BlackduckPersistentStorage:                     false,
		BlackduckPVCJSONSlice:                          []string{},
		BlackduckCertificateName:                       "",
		BlackduckCertificate:                           "",
		BlackduckCertificateKey:                        "",
		BlackduckProxyCertificate:                      "",
		BlackduckType:                                  "",
		BlackduckDesiredState:                          "",
		BlackduckEnvirons:                              []string{},
		BlackduckImageRegistries:                       []string{},
		BlackduckImageUIDMapJSONSlice:                  []string{},
		BlackduckLicenseKey:                            "",
	}
}

// GetSpec returns the Spec for the resource
func (ctl *BlackduckCtl) GetSpec() blackduckv1.BlackduckSpec {
	return *ctl.Spec
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
	cmd.Flags().StringVar(&ctl.BlackduckSize, "size", ctl.BlackduckSize, "Blackduck size - small, medium, large")
	cmd.Flags().StringVar(&ctl.BlackduckDbPrototype, "db-prototype", ctl.BlackduckDbPrototype, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckExternalPostgresPostgresHost, "external-postgres-host", ctl.BlackduckExternalPostgresPostgresHost, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckExternalPostgresPostgresPort, "external-postgres-port", ctl.BlackduckExternalPostgresPostgresPort, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckExternalPostgresPostgresAdmin, "external-postgres-admin", ctl.BlackduckExternalPostgresPostgresAdmin, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckExternalPostgresPostgresUser, "external-postgres-user", ctl.BlackduckExternalPostgresPostgresUser, "TODO")
	cmd.Flags().BoolVar(&ctl.BlackduckExternalPostgresPostgresSsl, "external-postgres-ssl", ctl.BlackduckExternalPostgresPostgresSsl, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", ctl.BlackduckExternalPostgresPostgresAdminPassword, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckExternalPostgresPostgresUserPassword, "external-postgres-user-password", ctl.BlackduckExternalPostgresPostgresUserPassword, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckPvcStorageClass, "pvc-storage-class", ctl.BlackduckPvcStorageClass, "TODO")
	cmd.Flags().BoolVar(&ctl.BlackduckLivenessProbes, "liveness-probes", ctl.BlackduckLivenessProbes, "Enable liveness probes")
	cmd.Flags().StringVar(&ctl.BlackduckScanType, "scan-type", ctl.BlackduckScanType, "TODO")
	cmd.Flags().BoolVar(&ctl.BlackduckPersistentStorage, "persistent-storage", ctl.BlackduckPersistentStorage, "Enable persistent storage")
	cmd.Flags().StringSliceVar(&ctl.BlackduckPVCJSONSlice, "pvc", ctl.BlackduckPVCJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckCertificateName, "db-certificate-name", ctl.BlackduckCertificateName, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckCertificate, "certificate", ctl.BlackduckCertificate, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckCertificateKey, "certificate-key", ctl.BlackduckCertificateKey, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckProxyCertificate, "proxy-certificate", ctl.BlackduckProxyCertificate, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckType, "type", ctl.BlackduckType, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckDesiredState, "desired-state", ctl.BlackduckDesiredState, "TODO")
	cmd.Flags().StringSliceVar(&ctl.BlackduckEnvirons, "environs", ctl.BlackduckEnvirons, "TODO")
	cmd.Flags().StringSliceVar(&ctl.BlackduckImageRegistries, "image-registries", ctl.BlackduckImageRegistries, "List of image registries")
	cmd.Flags().StringSliceVar(&ctl.BlackduckImageUIDMapJSONSlice, "image-uid-map", ctl.BlackduckImageUIDMapJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.BlackduckLicenseKey, "license-key", ctl.BlackduckLicenseKey, "TODO")
}

// SetFlags sets the Blackduck's Spec if a flag was changed
func (ctl *BlackduckCtl) SetFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "size":
			ctl.Spec.Size = ctl.BlackduckSize
		case "db-prototype":
			ctl.Spec.DbPrototype = ctl.BlackduckDbPrototype
		case "external-postgres-host":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresHost = ctl.BlackduckExternalPostgresPostgresHost
		case "external-postgres-port":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresPort = ctl.BlackduckExternalPostgresPostgresPort
		case "external-postgres-admin":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdmin = ctl.BlackduckExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUser = ctl.BlackduckExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresSsl = ctl.BlackduckExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresAdminPassword = ctl.BlackduckExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.ExternalPostgres.PostgresUserPassword = ctl.BlackduckExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			if ctl.Spec.ExternalPostgres == nil {
				ctl.Spec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			ctl.Spec.PVCStorageClass = ctl.BlackduckPvcStorageClass
		case "liveness-probes":
			ctl.Spec.LivenessProbes = ctl.BlackduckLivenessProbes
		case "scan-type":
			ctl.Spec.ScanType = ctl.BlackduckScanType
		case "persistent-storage":
			ctl.Spec.PersistentStorage = ctl.BlackduckPersistentStorage
		case "pvc":
			for _, pvcJSON := range ctl.BlackduckPVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				ctl.Spec.PVC = append(ctl.Spec.PVC, *pvc)
			}
		case "db-certificate-name":
			ctl.Spec.CertificateName = ctl.BlackduckCertificateName
		case "certificate":
			ctl.Spec.Certificate = ctl.BlackduckCertificate
		case "certificate-key":
			ctl.Spec.CertificateKey = ctl.BlackduckCertificateKey
		case "proxy-certificate":
			ctl.Spec.ProxyCertificate = ctl.BlackduckProxyCertificate
		case "type":
			ctl.Spec.Type = ctl.BlackduckType
		case "desired-state":
			ctl.Spec.DesiredState = ctl.BlackduckDesiredState
		case "environs":
			ctl.Spec.Environs = ctl.BlackduckEnvirons
		case "image-registries":
			ctl.Spec.ImageRegistries = ctl.BlackduckImageRegistries
		case "image-uid-map":
			type uid struct {
				Key   string `json:"key"`
				Value int64  `json:"value"`
			}
			ctl.Spec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range ctl.BlackduckImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				ctl.Spec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "license-key":
			ctl.Spec.LicenseKey = ctl.BlackduckLicenseKey
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
