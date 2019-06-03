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
	"testing"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewBlackDuckCtl(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackDuckCtl()
	assert.Equal(&Ctl{
		Spec: &blackduckv1.BlackduckSpec{},
	}, blackduckCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewBlackDuckCtl()
	assert.Equal(blackduckv1.BlackduckSpec{}, blackDuckCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewBlackDuckCtl()
	specToSet := blackduckv1.BlackduckSpec{Namespace: "test"}
	blackDuckCtl.SetSpec(specToSet)
	assert.Equal(specToSet, blackDuckCtl.GetSpec())

	// check for error
	assert.Error(blackDuckCtl.SetSpec(""))
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)

	// default case
	blackduckCtl := NewBlackDuckCtl()
	cmd := &cobra.Command{}
	assert.Nil(blackduckCtl.CheckSpecFlags(cmd.Flags()))

	var tests = []struct {
		input *Ctl
	}{
		// case
		{input: &Ctl{
			Spec: &blackduckv1.BlackduckSpec{},
			Size: "notValid",
		}},
		// case
		{input: &Ctl{
			Spec:     &blackduckv1.BlackduckSpec{},
			Environs: []string{"invalid"},
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
		expected *blackduckv1.BlackduckSpec
	}{
		{input: EmptySpec, expected: &blackduckv1.BlackduckSpec{}},
		{input: PersistentStorageLatestSpec, expected: crddefaults.GetBlackDuckDefaultPersistentStorageLatest()},
		{input: PersistentStorageV1Spec, expected: crddefaults.GetBlackDuckDefaultPersistentStorageV1()},
		{input: ExternalPersistentStorageLatestSpec, expected: crddefaults.GetBlackDuckDefaultExternalPersistentStorageLatest()},
		{input: ExternalPersistentStorageV1Spec, expected: crddefaults.GetBlackDuckDefaultExternalPersistentStorageV1()},
		{input: BDBASpec, expected: crddefaults.GetBlackDuckDefaultBDBA()},
		{input: EphemeralSpec, expected: crddefaults.GetBlackDuckDefaultEphemeral()},
		{input: EphemeralCustomAuthCASpec, expected: crddefaults.GetBlackDuckDefaultEphemeralCustomAuthCA()},
		{input: ExternalDBSpec, expected: crddefaults.GetBlackDuckDefaultExternalDB()},
		{input: IPV6DisabledSpec, expected: crddefaults.GetBlackDuckDefaultIPV6Disabled()},
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
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Black duck has persistent storage [true|false]")
	cmd.Flags().StringVar(&ctl.PVCFilePath, "pvc-file-path", ctl.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "Size of Black Duck [small|medium|large|x-large]")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Black Duck")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type of Black Duck Webserver's user interface [LOADBALANCER|NODEPORT|OPENSHIFT]")
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

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewBlackDuckCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, true)
	actualCtl.SetChangedFlags(cmd.Flags())

	expCtl := NewBlackDuckCtl()
	expCtl.Spec.Environs = []string{
		"USE_BINARY_UPLOADS:0",
		"ENABLE_SOURCE_UPLOADS:0",
	}

	assert.Equal(expCtl.Spec, actualCtl.Spec)

}

func TestSetFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *Ctl
		changedCtl  *Ctl
		changedSpec *blackduckv1.BlackduckSpec
	}{
		// case
		{
			flagName:   "size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{},
				Size: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Size: "changed"},
		},
		// case
		{
			flagName:   "version",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:    &blackduckv1.BlackduckSpec{},
				Version: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Version: "changed"},
		},
		// case
		{
			flagName:   "expose-ui",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:          &blackduckv1.BlackduckSpec{},
				ExposeService: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{ExposeService: "changed"},
		},
		// case
		{
			flagName:   "db-prototype",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:        &blackduckv1.BlackduckSpec{},
				DbPrototype: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{DbPrototype: "changed"},
		},
		// case
		{
			flagName:   "external-postgres-host",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{},
				ExternalPostgresHost: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresHost: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-port",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{},
				ExternalPostgresPort: 10,
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresPort: 10}},
		},
		// case
		{
			flagName:   "external-postgres-admin",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                  &blackduckv1.BlackduckSpec{},
				ExternalPostgresAdmin: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresAdmin: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-user",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{},
				ExternalPostgresUser: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresUser: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                &blackduckv1.BlackduckSpec{},
				ExternalPostgresSsl: "false",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresSsl: false}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                &blackduckv1.BlackduckSpec{},
				ExternalPostgresSsl: "true",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresSsl: true}},
		},
		// case
		{
			flagName:   "external-postgres-admin-password",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                          &blackduckv1.BlackduckSpec{},
				ExternalPostgresAdminPassword: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresAdminPassword: crddefaults.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "external-postgres-user-password",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                         &blackduckv1.BlackduckSpec{},
				ExternalPostgresUserPassword: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresUserPassword: crddefaults.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "pvc-storage-class",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				PvcStorageClass: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVCStorageClass: "changed"},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckv1.BlackduckSpec{},
				LivenessProbes: "false",
			},
			changedSpec: &blackduckv1.BlackduckSpec{LivenessProbes: false},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckv1.BlackduckSpec{},
				LivenessProbes: "true",
			},
			changedSpec: &blackduckv1.BlackduckSpec{LivenessProbes: true},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{},
				PersistentStorage: "false",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PersistentStorage: false},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{},
				PersistentStorage: "true",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PersistentStorage: true},
		},
		// case: add postgres-claim with size if PVC doesn't exist
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: append postgres-claim with size if PVC doesn't exist
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "other-pvc", Size: "other-size"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "other-pvc", Size: "other-size"}, {Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: update postgres-claim with size if PVC exists
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "blackduck-postgres", Size: "unchanged"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case
		{
			flagName:   "certificate-name",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				CertificateName: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{CertificateName: "changed"},
		},
		// case
		{
			flagName:   "type",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{},
				Type: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Type: "changed"},
		},
		// case
		{
			flagName:   "desired-state",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:         &blackduckv1.BlackduckSpec{},
				DesiredState: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{DesiredState: "changed"},
		},
		// case
		{
			// TODO: add a check in environs for correct input string format
			flagName:   "environs",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:     &blackduckv1.BlackduckSpec{},
				Environs: []string{"changed"},
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"changed"}},
		},
		// case
		{
			// TODO: add a check for name:Val
			flagName:   "image-registries",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				ImageRegistries: []string{"changed"},
			},
			changedSpec: &blackduckv1.BlackduckSpec{ImageRegistries: []string{"changed"}},
		},
		// case
		{
			flagName:   "license-key",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:       &blackduckv1.BlackduckSpec{},
				LicenseKey: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{LicenseKey: "changed"},
		},
		// case : set binary analysis to disabled by default
		{
			flagName:   "enable-binary-analysis",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{Environs: []string{}},
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:0"}},
		},
		// case : set binary analysis to enabled
		{
			flagName:   "enable-binary-analysis",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:0"}},
				EnableBinaryAnalysis: true,
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:1"}},
		},
		// case : set binary analysis to disabled
		{
			flagName:   "enable-binary-analysis",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:1"}},
				EnableBinaryAnalysis: false,
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"USE_BINARY_UPLOADS:0"}},
		},
		// case : set source code upload to disabled by default
		{
			flagName:   "enable-source-code-upload",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{Environs: []string{}},
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:0"}},
		},
		// case : set source code upload to enabled
		{
			flagName:   "enable-source-code-upload",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                   &blackduckv1.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:0"}},
				EnableSourceCodeUpload: true,
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:1"}},
		},
		// case : set source code upload to disabled
		{
			flagName:   "enable-source-code-upload",
			initialCtl: NewBlackDuckCtl(),
			changedCtl: &Ctl{
				Spec:                   &blackduckv1.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:1"}},
				EnableSourceCodeUpload: false,
			},
			changedSpec: &blackduckv1.BlackduckSpec{Environs: []string{"ENABLE_SOURCE_UPLOADS:0"}},
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
	assert.Equal(&blackduckv1.BlackduckSpec{}, actualCtl.Spec)

}
