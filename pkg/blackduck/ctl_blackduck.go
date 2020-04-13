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

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelmValuesFromCobraFlags is a type for converting synopsysctl flags
// to Helm Chart fields and values
// args: map of helm chart field to value
type HelmValuesFromCobraFlags struct {
	args     map[string]interface{}
	flagTree FlagTree
}

// FlagTree is a set of fields needed to configure the Blackduck Helm Chart
type FlagTree struct {
	Size                          string
	Version                       string
	ExposeService                 string
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
	MigrationMode                 bool
	Environs                      []string
	AdminPassword                 string
	UserPassword                  string
	EnableBinaryAnalysis          bool
	EnableSourceCodeUpload        bool
	NodeAffinityFilePath          string
	SecurityContextFilePath       string
	Registry                      string
	RegistryNamespace             string
	PullSecrets                   []string
	SealKey                       string
}

// NewHelmValuesFromCobraFlags creates a new HelmValuesFromCobraFlags type
func NewHelmValuesFromCobraFlags() *HelmValuesFromCobraFlags {
	return &HelmValuesFromCobraFlags{
		args: make(map[string]interface{}, 0),
	}
}

// GenerateHelmFlagsFromCobraFlags checks each flag in synopsysctl and updates the map to
// contain the corresponding helm chart field and value
func (ctl *HelmValuesFromCobraFlags) GenerateHelmFlagsFromCobraFlags(flagset *pflag.FlagSet) (map[string]interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.AddHelmValueByCobraFlag)

	return ctl.args, nil
}

func (ctl *HelmValuesFromCobraFlags) SetArgs(args map[string]interface{}) {
	for key, value := range args {
		ctl.args[key] = value
	}
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

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for BlackDuck's Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *HelmValuesFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.PvcStorageClass, "pvc-storage-class", ctl.flagTree.PvcStorageClass, "Name of Storage Class for the PVC")
		cmd.Flags().StringVar(&ctl.flagTree.PersistentStorage, "persistent-storage", "true", "If true, Black Duck has persistent storage [true|false]")
		cmd.Flags().StringVar(&ctl.flagTree.PVCFilePath, "pvc-file-path", ctl.flagTree.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	}
	cmd.Flags().StringVar(&ctl.flagTree.Size, "size", ctl.flagTree.Size, "Size of Black Duck [small|medium|large|x-large]")
	cmd.Flags().StringVar(&ctl.flagTree.Version, "version", "2020.4.0", "Version of Black Duck")
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.ExposeService, "expose-ui", util.NONE, "Service type of Black Duck webserver's user interface [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	} else {
		cmd.Flags().StringVar(&ctl.flagTree.ExposeService, "expose-ui", ctl.flagTree.ExposeService, "Service type of Black Duck webserver's user interface [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	}

	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresHost, "external-postgres-host", ctl.flagTree.ExternalPostgresHost, "Host of external Postgres")
	cmd.Flags().IntVar(&ctl.flagTree.ExternalPostgresPort, "external-postgres-port", 5432, "Port of external Postgres")
	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresAdmin, "external-postgres-admin", ctl.flagTree.ExternalPostgresAdmin, "Name of 'admin' of external Postgres database")
	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresUser, "external-postgres-user", "blackduck_user", "Name of 'user' of external Postgres database")
	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresSsl, "external-postgres-ssl", "true", "If true, Black Duck uses SSL for external Postgres connection [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresAdminPassword, "external-postgres-admin-password", ctl.flagTree.ExternalPostgresAdminPassword, "'admin' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.flagTree.ExternalPostgresUserPassword, "external-postgres-user-password", ctl.flagTree.ExternalPostgresUserPassword, "'user' password of external Postgres database")
	cmd.Flags().StringVar(&ctl.flagTree.LivenessProbes, "liveness-probes", ctl.flagTree.LivenessProbes, "If true, Black Duck uses liveness probes [true|false]")
	cmd.Flags().StringVar(&ctl.flagTree.PostgresClaimSize, "postgres-claim-size", "150Gi", "Size of the blackduck-postgres PVC")
	cmd.Flags().StringVar(&ctl.flagTree.CertificateName, "certificate-name", ctl.flagTree.CertificateName, "Name of Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.flagTree.CertificateFilePath, "certificate-file-path", ctl.flagTree.CertificateFilePath, "Absolute path to a file for the Black Duck nginx certificate")
	cmd.Flags().StringVar(&ctl.flagTree.CertificateKeyFilePath, "certificate-key-file-path", ctl.flagTree.CertificateKeyFilePath, "Absolute path to a file for the Black Duck nginx certificate key")
	cmd.Flags().StringVar(&ctl.flagTree.ProxyCertificateFilePath, "proxy-certificate-file-path", ctl.flagTree.ProxyCertificateFilePath, "Absolute path to a file for the Black Duck proxy server’s Certificate Authority (CA)")
	cmd.Flags().StringVar(&ctl.flagTree.AuthCustomCAFilePath, "auth-custom-ca-file-path", ctl.flagTree.AuthCustomCAFilePath, "Absolute path to a file for the Custom Auth CA for Black Duck")

	if !strings.Contains(cmd.CommandPath(), "native") {
		cmd.Flags().BoolVar(&ctl.flagTree.MigrationMode, "migration-mode", ctl.flagTree.MigrationMode, "Create Black Duck in the database-migration state")
	}
	cmd.Flags().StringSliceVar(&ctl.flagTree.Environs, "environs", ctl.flagTree.Environs, "List of environment variables")

	cmd.Flags().StringVar(&ctl.flagTree.AdminPassword, "admin-password", ctl.flagTree.AdminPassword, "'admin' password of Postgres database")
	cmd.Flags().StringVar(&ctl.flagTree.UserPassword, "user-password", ctl.flagTree.UserPassword, "'user' password of Postgres database")
	cmd.Flags().BoolVar(&ctl.flagTree.EnableBinaryAnalysis, "enable-binary-analysis", false, "If true, enable binary analysis by setting the environment variable (this takes priority over environs flag values)")
	cmd.Flags().BoolVar(&ctl.flagTree.EnableSourceCodeUpload, "enable-source-code-upload", false, "If true, enable source code upload by setting the environment variable (this takes priority over environs flag values)")
	cmd.Flags().StringVar(&ctl.flagTree.NodeAffinityFilePath, "node-affinity-file-path", ctl.flagTree.NodeAffinityFilePath, "Absolute path to a file containing a list of node affinities")
	cmd.Flags().StringVar(&ctl.flagTree.SecurityContextFilePath, "security-context-file-path", ctl.flagTree.SecurityContextFilePath, "Absolute path to a file containing a map of pod names to security contexts runAsUser, fsGroup, and runAsGroup")
	cmd.Flags().StringVar(&ctl.flagTree.Registry, "registry", "docker.io/blackducksoftware", "Name of the registry to use for images e.g. docker.io/blackducksoftware")
	cmd.Flags().StringSliceVar(&ctl.flagTree.PullSecrets, "pull-secret-name", ctl.flagTree.PullSecrets, "Only if the registry requires authentication")
	if master {
		cmd.Flags().StringVar(&ctl.flagTree.SealKey, "seal-key", ctl.flagTree.SealKey, "Seal key to encrypt the master key when Source code upload is enabled and it should be of length 32")
	}
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
func (ctl *HelmValuesFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	if FlagWasSet(flagset, "size") {
		if !isValidSize(ctl.flagTree.Size) {
			return fmt.Errorf("size must be 'small', 'medium', 'large' or 'x-large'")
		}
	}
	if FlagWasSet(flagset, "expose-ui") {
		isValid := util.IsExposeServiceValid(ctl.flagTree.ExposeService)
		if !isValid {
			return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	if FlagWasSet(flagset, "environs") {
		for _, environ := range ctl.flagTree.Environs {
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
	if FlagWasSet(flagset, "seal-key") {
		if len(ctl.flagTree.SealKey) != 32 {
			return fmt.Errorf("seal key should be of length 32")
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

// AddHelmValueByCobraFlag updates a field in the blackDuckSpec if the flag was set by the user. It gets the
// value from the corresponding struct field.
// Note: It should only handle values with a 1 to 1 mapping - struct-field to spec
func (ctl *HelmValuesFromCobraFlags) AddHelmValueByCobraFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "size":
			util.SetHelmValueInMap(ctl.args, []string{"size"}, ctl.flagTree.Size)
		//case "expose-ui":
		//	ctl.blackDuckSpec.ExposeService = ctl.ExposeService
		case "environs":
			for _, value := range ctl.flagTree.Environs {
				values := strings.SplitN(value, ":", 2)
				if len(values) != 2 {
					panic(fmt.Errorf("invalid environ configuration for %s", value))
				}
				util.SetHelmValueInMap(ctl.args, []string{"environs", values[0]}, values[1])
			}
		case "enable-binary-analysis":
			util.SetHelmValueInMap(ctl.args, []string{"enableBinaryScanner"}, ctl.flagTree.EnableBinaryAnalysis)
		case "enable-source-code-upload":
			util.SetHelmValueInMap(ctl.args, []string{"enableSourceCodeUpload"}, ctl.flagTree.EnableSourceCodeUpload)
		case "external-postgres-host":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "host"}, ctl.flagTree.ExternalPostgresHost)
		case "external-postgres-port":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "port"}, ctl.flagTree.ExternalPostgresPort)
		case "external-postgres-admin":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "adminUserName"}, ctl.flagTree.ExternalPostgresAdmin)
		case "external-postgres-user":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "userUserName"}, ctl.flagTree.ExternalPostgresUser)
		case "external-postgres-ssl":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "ssl"}, strings.ToUpper(ctl.flagTree.ExternalPostgresSsl) == "TRUE")
		case "external-postgres-admin-password":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "adminPassword"}, ctl.flagTree.ExternalPostgresUser)
		case "external-postgres-user-password":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "userPassword"}, ctl.flagTree.ExternalPostgresUserPassword)
		case "pvc-storage-class":
			util.SetHelmValueInMap(ctl.args, []string{"storageClass"}, ctl.flagTree.PvcStorageClass)
		case "liveness-probes":
			util.SetHelmValueInMap(ctl.args, []string{"enableLivenessProbe"}, strings.ToUpper(ctl.flagTree.LivenessProbes) == "TRUE")
		case "persistent-storage":
			util.SetHelmValueInMap(ctl.args, []string{"enablePersistentStorage"}, strings.ToUpper(ctl.flagTree.PersistentStorage) == "TRUE")
		//case "pvc-file-path":
		//	data, err := util.ReadFileData(ctl.PVCFilePath)
		//	if err != nil {
		//		log.Fatalf("failed to read pvc file: %+v", err)
		//	}
		//	pvcs := []blackduckv1.PVC{}
		//	err = json.Unmarshal([]byte(data), &pvcs)
		//	if err != nil {
		//		log.Fatalf("failed to unmarshal pvc structs: %+v", err)
		//	}
		//	for _, newPVC := range pvcs {
		//		found := false
		//		for i, currPVC := range ctl.blackDuckSpec.PVC {
		//			if newPVC.Name == currPVC.Name {
		//				ctl.blackDuckSpec.PVC[i] = newPVC
		//				found = true
		//				break
		//			}
		//		}
		//		if !found {
		//			ctl.blackDuckSpec.PVC = append(ctl.blackDuckSpec.PVC, newPVC)
		//		}
		//	}
		case "node-affinity-file-path":
			data, err := util.ReadFileData(ctl.flagTree.NodeAffinityFilePath)
			if err != nil {
				log.Fatalf("failed to read node affinity file: %+v", err)
			}
			nodeAffinities := map[string][]blackduckv1.NodeAffinity{}
			err = json.Unmarshal([]byte(data), &nodeAffinities)
			if err != nil {
				log.Fatalf("failed to unmarshal node affinities: %+v", err)
			}

			for k, v := range nodeAffinities {
				util.SetHelmValueInMap(ctl.args, []string{k, "affinity"}, OperatorAffinityTok8sAffinity(v))
			}

		case "security-context-file-path":
			data, err := util.ReadFileData(ctl.flagTree.SecurityContextFilePath)
			if err != nil {
				log.Errorf("failed to read security context file: %+v", err)
				return
			}
			SecurityContexts := map[string]api.SecurityContext{}
			err = json.Unmarshal([]byte(data), &SecurityContexts)
			if err != nil {
				log.Errorf("failed to unmarshal security contexts: %+v", err)
				return
			}
			for k, v := range SecurityContexts {
				util.SetHelmValueInMap(ctl.args, []string{k, "securityContext"}, OperatorSecurityContextTok8sAffinity(v))
			}
		case "postgres-claim-size":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "claimSize"}, ctl.flagTree.PostgresClaimSize)
		case "admin-password":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "adminPassword"}, ctl.flagTree.AdminPassword)
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "isExternal"}, false)
		case "user-password":
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "userPassword"}, ctl.flagTree.UserPassword)
			util.SetHelmValueInMap(ctl.args, []string{"postgres", "isExternal"}, false)
		case "registry":
			util.SetHelmValueInMap(ctl.args, []string{"registry"}, ctl.flagTree.Registry)
		case "pull-secret-name":
			var pullSecrets []corev1.LocalObjectReference
			for _, v := range ctl.flagTree.PullSecrets {
				pullSecrets = append(pullSecrets, corev1.LocalObjectReference{Name: v})
			}
			util.SetHelmValueInMap(ctl.args, []string{"imagePullSecrets"}, pullSecrets)
		case "seal-key":
			util.SetHelmValueInMap(ctl.args, []string{"sealKey"}, ctl.flagTree.SealKey)
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
