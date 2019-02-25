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

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Gloabal Specs
var globalBlackduckSpec = &blackduckv1.BlackduckSpec{}

// Blackduck Spec Flags
var blackduckSize = ""
var blackduckDbPrototype = ""
var blackduckExternalPostgresPostgresHost = ""
var blackduckExternalPostgresPostgresPort = 0
var blackduckExternalPostgresPostgresAdmin = ""
var blackduckExternalPostgresPostgresUser = ""
var blackduckExternalPostgresPostgresSsl = false
var blackduckExternalPostgresPostgresAdminPassword = ""
var blackduckExternalPostgresPostgresUserPassword = ""
var blackduckPvcStorageClass = ""
var blackduckLivenessProbes = false
var blackduckScanType = ""
var blackduckPersistentStorage = false
var blackduckPVCJSONSlice = []string{}
var blackduckCertificateName = ""
var blackduckCertificate = ""
var blackduckCertificateKey = ""
var blackduckProxyCertificate = ""
var blackduckType = ""
var blackduckDesiredState = ""
var blackduckEnvirons = []string{}
var blackduckImageRegistries = []string{}
var blackduckImageUIDMapJSONSlice = []string{}
var blackduckLicenseKey = ""

func addBlackduckSpecFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&blackduckSize, "size", blackduckSize, "Blackduck size - small, medium, large")
	cmd.Flags().StringVar(&blackduckDbPrototype, "db-prototype", blackduckDbPrototype, "TODO")
	cmd.Flags().StringVar(&blackduckExternalPostgresPostgresHost, "external-postgres-host", blackduckExternalPostgresPostgresHost, "TODO")
	cmd.Flags().IntVar(&blackduckExternalPostgresPostgresPort, "external-postgres-port", blackduckExternalPostgresPostgresPort, "TODO")
	cmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdmin, "external-postgres-admin", blackduckExternalPostgresPostgresAdmin, "TODO")
	cmd.Flags().StringVar(&blackduckExternalPostgresPostgresUser, "external-postgres-user", blackduckExternalPostgresPostgresUser, "TODO")
	cmd.Flags().BoolVar(&blackduckExternalPostgresPostgresSsl, "external-postgres-ssl", blackduckExternalPostgresPostgresSsl, "TODO")
	cmd.Flags().StringVar(&blackduckExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", blackduckExternalPostgresPostgresAdminPassword, "TODO")
	cmd.Flags().StringVar(&blackduckExternalPostgresPostgresUserPassword, "external-postgres-user-password", blackduckExternalPostgresPostgresUserPassword, "TODO")
	cmd.Flags().StringVar(&blackduckPvcStorageClass, "pvc-storage-class", blackduckPvcStorageClass, "TODO")
	cmd.Flags().BoolVar(&blackduckLivenessProbes, "liveness-probes", blackduckLivenessProbes, "Enable liveness probes")
	cmd.Flags().StringVar(&blackduckScanType, "scan-type", blackduckScanType, "TODO")
	cmd.Flags().BoolVar(&blackduckPersistentStorage, "persistent-storage", blackduckPersistentStorage, "Enable persistent storage")
	cmd.Flags().StringSliceVar(&blackduckPVCJSONSlice, "pvc", blackduckPVCJSONSlice, "TODO")
	cmd.Flags().StringVar(&blackduckCertificateName, "db-certificate-name", blackduckCertificateName, "TODO")
	cmd.Flags().StringVar(&blackduckCertificate, "certificate", blackduckCertificate, "TODO")
	cmd.Flags().StringVar(&blackduckCertificateKey, "certificate-key", blackduckCertificateKey, "TODO")
	cmd.Flags().StringVar(&blackduckProxyCertificate, "proxy-certificate", blackduckProxyCertificate, "TODO")
	cmd.Flags().StringVar(&blackduckType, "type", blackduckType, "TODO")
	cmd.Flags().StringVar(&blackduckDesiredState, "desired-state", blackduckDesiredState, "TODO")
	cmd.Flags().StringSliceVar(&blackduckEnvirons, "environs", blackduckEnvirons, "TODO")
	cmd.Flags().StringSliceVar(&blackduckImageRegistries, "image-registries", blackduckImageRegistries, "List of image registries")
	cmd.Flags().StringSliceVar(&blackduckImageUIDMapJSONSlice, "image-uid-map", blackduckImageUIDMapJSONSlice, "TODO")
	cmd.Flags().StringVar(&blackduckLicenseKey, "license-key", blackduckLicenseKey, "TODO")
}

func setBlackduckFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "size":
			globalBlackduckSpec.Size = blackduckSize
		case "db-prototype":
			globalBlackduckSpec.DbPrototype = blackduckDbPrototype
		case "external-postgres-host":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresHost = blackduckExternalPostgresPostgresHost
		case "external-postgres-port":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresPort = blackduckExternalPostgresPostgresPort
		case "external-postgres-admin":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresAdmin = blackduckExternalPostgresPostgresAdmin
		case "external-postgres-user":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresUser = blackduckExternalPostgresPostgresUser
		case "external-postgres-ssl":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresSsl = blackduckExternalPostgresPostgresSsl
		case "external-postgres-admin-password":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresAdminPassword = blackduckExternalPostgresPostgresAdminPassword
		case "external-postgres-user-password":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.ExternalPostgres.PostgresUserPassword = blackduckExternalPostgresPostgresUserPassword
		case "pvc-storage-class":
			if globalBlackduckSpec.ExternalPostgres == nil {
				globalBlackduckSpec.ExternalPostgres = &blackduckv1.PostgresExternalDBConfig{}
			}
			globalBlackduckSpec.PVCStorageClass = blackduckPvcStorageClass
		case "liveness-probes":
			globalBlackduckSpec.LivenessProbes = blackduckLivenessProbes
		case "scan-type":
			globalBlackduckSpec.ScanType = blackduckScanType
		case "persistent-storage":
			globalBlackduckSpec.PersistentStorage = blackduckPersistentStorage
		case "pvc":
			for _, pvcJSON := range blackduckPVCJSONSlice {
				pvc := &blackduckv1.PVC{}
				json.Unmarshal([]byte(pvcJSON), pvc)
				globalBlackduckSpec.PVC = append(globalBlackduckSpec.PVC, *pvc)
			}
		case "db-certificate-name":
			globalBlackduckSpec.CertificateName = blackduckCertificateName
		case "certificate":
			globalBlackduckSpec.Certificate = blackduckCertificate
		case "certificate-key":
			globalBlackduckSpec.CertificateKey = blackduckCertificateKey
		case "proxy-certificate":
			globalBlackduckSpec.ProxyCertificate = blackduckProxyCertificate
		case "type":
			globalBlackduckSpec.Type = blackduckType
		case "desired-state":
			globalBlackduckSpec.DesiredState = blackduckDesiredState
		case "environs":
			globalBlackduckSpec.Environs = blackduckEnvirons
		case "image-registries":
			globalBlackduckSpec.ImageRegistries = blackduckImageRegistries
		case "image-uid-map":
			type uid struct {
				Key   string `json:"key"`
				Value int64  `json:"value"`
			}
			globalBlackduckSpec.ImageUIDMap = make(map[string]int64)
			for _, uidJSON := range blackduckImageUIDMapJSONSlice {
				uidStruct := &uid{}
				json.Unmarshal([]byte(uidJSON), uidStruct)
				globalBlackduckSpec.ImageUIDMap[uidStruct.Key] = uidStruct.Value
			}
		case "license-key":
			globalBlackduckSpec.LicenseKey = blackduckLicenseKey
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	}
	log.Debugf("Flag %s: UNCHANGED\n", f.Name)
}
