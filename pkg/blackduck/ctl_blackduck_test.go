/*
Copyright (C) 2019 Synopsys, Inc.

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
	"sort"
	"testing"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewBlackDuckCtl(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackDuckCtl()
	assert.Equal(&Ctl{
		Spec: &blackduckapi.BlackduckSpec{},
	}, blackduckCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewBlackDuckCtl()
	assert.Equal(blackduckapi.BlackduckSpec{}, blackDuckCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewBlackDuckCtl()
	specToSet := blackduckapi.BlackduckSpec{Namespace: "test"}
	blackDuckCtl.SetSpec(specToSet)
	assert.Equal(specToSet, blackDuckCtl.GetSpec())

	// check for error
	assert.Error(blackDuckCtl.SetSpec(""))
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)

	// default case
	blackduckCtl := NewBlackDuckCtl()
	blackduckCtl.ExposeService = util.NONE
	cmd := &cobra.Command{}
	blackduckCtl.AddSpecFlags(cmd, true)
	err := blackduckCtl.CheckSpecFlags(cmd.Flags())
	if err != nil {
		t.Errorf("expected nil error, got: %+v", err)
	}

	var tests = []struct {
		input *Ctl
	}{
		// case
		{input: &Ctl{
			Spec: &blackduckapi.BlackduckSpec{},
			Size: "notValid",
		}},
		// case
		{input: &Ctl{
			Spec:     &blackduckapi.BlackduckSpec{},
			Environs: []string{"invalid"},
		}},
		{input: &Ctl{
			Spec:          &blackduckapi.BlackduckSpec{},
			ExposeService: "",
		}},
	}

	for _, test := range tests {
		assert.Error(test.input.CheckSpecFlags(cmd.Flags()))
	}
}
func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewBlackDuckCtl()
	var tests = []struct {
		input    string
		expected *blackduckapi.BlackduckSpec
	}{
		{input: EmptySpec, expected: &blackduckapi.BlackduckSpec{}},
		{input: PersistentStorageLatestSpec, expected: util.GetBlackDuckDefaultPersistentStorageLatest()},
		{input: PersistentStorageV1Spec, expected: util.GetBlackDuckDefaultPersistentStorageV1()},
		{input: ExternalPersistentStorageLatestSpec, expected: util.GetBlackDuckDefaultExternalPersistentStorageLatest()},
		{input: ExternalPersistentStorageV1Spec, expected: util.GetBlackDuckDefaultExternalPersistentStorageV1()},
		{input: BDBASpec, expected: util.GetBlackDuckDefaultBDBA()},
		{input: EphemeralSpec, expected: util.GetBlackDuckDefaultEphemeral()},
		{input: EphemeralCustomAuthCASpec, expected: util.GetBlackDuckDefaultEphemeralCustomAuthCA()},
		{input: ExternalDBSpec, expected: util.GetBlackDuckDefaultExternalDB()},
		{input: IPV6DisabledSpec, expected: util.GetBlackDuckDefaultIPV6Disabled()},
	}

	// test cases: "empty", "persistentStorage", "default"
	for _, test := range tests {
		assert.Nil(blackDuckCtl.SwitchSpec(test.input))
		assert.Equal(*test.expected, blackDuckCtl.GetSpec())
	}

	// test cases: ""
	createBlackDuckSpecType := ""
	assert.Error(blackDuckCtl.SwitchSpec(createBlackDuckSpecType))
}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	ctl := NewBlackDuckCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd, true)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "Name of Storage Class for the PVC")
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Black Duck has persistent storage (true|false)")
	cmd.Flags().StringVar(&ctl.PVCFilePath, "pvc-file-path", ctl.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "Size of Black Duck (small|medium|large|x-large)")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Black Duck")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type of Black Duck webserver's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	cmd.Flags().StringVar(&ctl.DbPrototype, "db-prototype", ctl.DbPrototype, "Black Duck name to clone the database")
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
	cmd.Flags().StringVar(&ctl.Type, "type", ctl.Type, "Type of Black Duck")
	cmd.Flags().StringVar(&ctl.DesiredState, "desired-state", ctl.DesiredState, "Desired state of Black Duck")
	cmd.Flags().BoolVar(&ctl.MigrationMode, "migration-mode", ctl.MigrationMode, "Create Black Duck in the database-migration state")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of Environment Variables (NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key of Black Duck")
	cmd.Flags().StringVar(&ctl.AdminPassword, "admin-password", ctl.AdminPassword, "'admin' password of Postgres database")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "'postgres' password of Postgres database")
	cmd.Flags().StringVar(&ctl.UserPassword, "user-password", ctl.UserPassword, "'user' password of Postgres database")
	cmd.Flags().BoolVar(&ctl.EnableBinaryAnalysis, "enable-binary-analysis", ctl.EnableBinaryAnalysis, "If true, enable binary analysis")
	cmd.Flags().BoolVar(&ctl.EnableSourceCodeUpload, "enable-source-code-upload", ctl.EnableSourceCodeUpload, "If true, enable source code upload")
	cmd.Flags().StringVar(&ctl.NodeAffinityFilePath, "node-affinity-file-path", ctl.NodeAffinityFilePath, "Absolute path to a file containing a list of node affinities")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("desired-state", "desired-state flag is deprecated and will be removed by the next release")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewBlackDuckCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, true)
	actualCtl.SetChangedFlags(cmd.Flags())

	expCtl := NewBlackDuckCtl()
	sort.Strings(expCtl.Spec.Environs)
	sort.Strings(actualCtl.Spec.Environs)

	assert.Equal(expCtl.Spec, actualCtl.Spec)

}

func TestSetFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *Ctl
		changedCtl  *Ctl
		changedSpec *blackduckapi.BlackduckSpec
	}{
		// case
		{
			flagName:   "size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckapi.BlackduckSpec{},
				Size: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Size: "changed"},
		},
		// case
		{
			flagName:   "version",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:    &blackduckapi.BlackduckSpec{},
				Version: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Version: "changed"},
		},
		// case
		{
			flagName:   "expose-ui",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:          &blackduckapi.BlackduckSpec{},
				ExposeService: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{ExposeService: "changed"},
		},
		// case
		{
			flagName:   "db-prototype",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:        &blackduckapi.BlackduckSpec{},
				DbPrototype: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{DbPrototype: "changed"},
		},
		// case
		{
			flagName:   "external-postgres-host",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{},
				ExternalPostgresHost: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresHost: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-port",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{},
				ExternalPostgresPort: 10,
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresPort: 10}},
		},
		// case
		{
			flagName:   "external-postgres-admin",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                  &blackduckapi.BlackduckSpec{},
				ExternalPostgresAdmin: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresAdmin: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-user",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{},
				ExternalPostgresUser: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresUser: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                &blackduckapi.BlackduckSpec{},
				ExternalPostgresSsl: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresSsl: false}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                &blackduckapi.BlackduckSpec{},
				ExternalPostgresSsl: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresSsl: true}},
		},
		// case
		{
			flagName:   "external-postgres-admin-password",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                          &blackduckapi.BlackduckSpec{},
				ExternalPostgresAdminPassword: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresAdminPassword: util.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "external-postgres-user-password",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                         &blackduckapi.BlackduckSpec{},
				ExternalPostgresUserPassword: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresUserPassword: util.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "pvc-storage-class",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckapi.BlackduckSpec{},
				PvcStorageClass: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVCStorageClass: "changed"},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckapi.BlackduckSpec{},
				LivenessProbes: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LivenessProbes: false},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckapi.BlackduckSpec{},
				LivenessProbes: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LivenessProbes: true},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckapi.BlackduckSpec{},
				PersistentStorage: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PersistentStorage: false},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckapi.BlackduckSpec{},
				PersistentStorage: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PersistentStorage: true},
		},
		// case
		{
			flagName:   "pvc-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:        &blackduckapi.BlackduckSpec{},
				PVCFilePath: "../../examples/synopsysctl/pvc.json",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "name1", Size: "size1", StorageClass: "storageclass1"}, {Name: "name2", Size: "size2", StorageClass: "storageclass2"}}},
		},
		// case
		{
			flagName:   "node-affinity-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{},
				NodeAffinityFilePath: "../../examples/synopsysctl/nodeAffinity.json",
			},
			changedSpec: &blackduckapi.BlackduckSpec{NodeAffinities: map[string][]blackduckapi.NodeAffinity{
				"affinity1": {
					{
						AffinityType: "type1",
						Key:          "key1",
						Op:           "op1",
						Values:       []string{"val1.1", "val1.2"},
					},
				},
			},
			},
		},
		// case: add postgres-claim with size if PVC doesn't exist
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckapi.BlackduckSpec{},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: append postgres-claim with size if PVC doesn't exist
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "other-pvc", Size: "other-size"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "other-pvc", Size: "other-size"}, {Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: update postgres-claim with size if PVC exists
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "unchanged"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case
		{
			flagName:   "certificate-name",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckapi.BlackduckSpec{},
				CertificateName: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{CertificateName: "changed"},
		},
		// case
		{
			flagName:   "certificate-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                &blackduckapi.BlackduckSpec{},
				CertificateFilePath: "../../examples/synopsysctl/certificate.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Certificate: "CERTIFICATE"},
		},
		// case
		{
			flagName:   "certificate-key-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                   &blackduckapi.BlackduckSpec{},
				CertificateKeyFilePath: "../../examples/synopsysctl/certificateKey.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{CertificateKey: "CERTIFICATE_KEY=CERTIFICATE_KEY_DATA"},
		},
		// case
		{
			flagName:   "proxy-certificate-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                     &blackduckapi.BlackduckSpec{},
				ProxyCertificateFilePath: "../../examples/synopsysctl/proxyCertificate.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{ProxyCertificate: "PROXY_CERTIFICATE"},
		},
		// case
		{
			flagName:   "auth-custom-ca-file-path",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{},
				AuthCustomCAFilePath: "../../examples/synopsysctl/authCustomCA.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{AuthCustomCA: "AUTH_CUSTOM_CA"},
		},
		// case
		{
			flagName:   "type",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckapi.BlackduckSpec{},
				Type: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Type: "changed"},
		},
		// case
		{
			flagName:   "desired-state",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:         &blackduckapi.BlackduckSpec{},
				DesiredState: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{DesiredState: "changed"},
		},
		// case
		{
			flagName:   "migration-mode",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:          &blackduckapi.BlackduckSpec{},
				MigrationMode: true,
			},
			changedSpec: &blackduckapi.BlackduckSpec{DesiredState: "DbMigrate"},
		},
		// case
		{
			// TODO: add a check in environs for correct input string format
			flagName:   "environs",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:     &blackduckapi.BlackduckSpec{},
				Environs: []string{"changed"},
			},
			changedSpec: &blackduckapi.BlackduckSpec{Environs: []string{"changed"}},
		},
		// case
		{
			// TODO: add a check for name:Val
			flagName:   "image-registries",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckapi.BlackduckSpec{},
				ImageRegistries: []string{"changed"},
			},
			changedSpec: &blackduckapi.BlackduckSpec{ImageRegistries: []string{"changed"}},
		},
		// case
		{
			flagName:   "license-key",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:       &blackduckapi.BlackduckSpec{},
				LicenseKey: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LicenseKey: "changed"},
		},
		// case : set binary analysis to enabled
		{
			flagName:   "enable-binary-analysis",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:0"}},
				EnableBinaryAnalysis: true,
			},
			changedSpec: &blackduckapi.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:1"}},
		},
		// case : set binary analysis to disabled
		{
			flagName:   "enable-binary-analysis",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckapi.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:1"}},
				EnableBinaryAnalysis: false,
			},
			changedSpec: &blackduckapi.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:0"}},
		},
		// case : set source code upload to enabled
		{
			flagName:   "enable-source-code-upload",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                   &blackduckapi.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:false"}},
				EnableSourceCodeUpload: true,
			},
			changedSpec: &blackduckapi.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:true"}},
		},
		// case : set source code upload to disabled
		{
			flagName:   "enable-source-code-upload",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                   &blackduckapi.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:true"}},
				EnableSourceCodeUpload: false,
			},
			changedSpec: &blackduckapi.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:false"}},
		},
	}

	// get the Ctl's flags
	cmd := &cobra.Command{}
	actualCtl := NewBlackDuckCtl()
	actualCtl.AddSpecFlags(cmd, true)
	flagset := cmd.Flags()

	for _, test := range tests {
		actualCtl = NewBlackDuckCtl()
		// check the Flag exists
		foundFlag := flagset.Lookup(test.flagName)
		if foundFlag == nil {
			t.Errorf("flag %s is not in the spec", test.flagName)
		}
		// check the correct Ctl is used
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		// test setting a flag
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetFlag(f)
		assert.Equal(test.changedSpec, actualCtl.Spec)
	}

	// case: nothing set if flag doesn't exist
	actualCtl = NewBlackDuckCtl()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	actualCtl.SetFlag(f)
	assert.Equal(&blackduckapi.BlackduckSpec{}, actualCtl.Spec)

}
