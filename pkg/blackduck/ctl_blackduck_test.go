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

func TestNewCRSpecBuilderFromCobraFlags(t *testing.T) {
	assert := assert.New(t)
	blackDuckCobraHelper := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(&CRSpecBuilderFromCobraFlags{
		blackDuckSpec: &blackduckapi.BlackduckSpec{},
	}, blackDuckCobraHelper)
}

func TestGetCRSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(blackduckapi.BlackduckSpec{}, blackDuckCtl.GetCRSpec())
}

func TestSetCRSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewCRSpecBuilderFromCobraFlags()
	specToSet := blackduckapi.BlackduckSpec{Namespace: "test"}
	blackDuckCtl.SetCRSpec(specToSet)
	assert.Equal(specToSet, blackDuckCtl.GetCRSpec())

	// check for error
	assert.Error(blackDuckCtl.SetCRSpec(""))
}

func TestCheckValuesFromFlags(t *testing.T) {
	// default case
	blackDuckCobraHelper := NewCRSpecBuilderFromCobraFlags()
	blackDuckCobraHelper.ExposeService = util.NONE
	cmd := &cobra.Command{}
	blackDuckCobraHelper.AddCRSpecFlagsToCommand(cmd, true)
	err := blackDuckCobraHelper.CheckValuesFromFlags(cmd.Flags())
	if err != nil {
		t.Errorf("expected nil error, got: %+v", err)
	}

	var tests = []struct {
		input          *CRSpecBuilderFromCobraFlags
		flagNameToTest string
		flagValue      string
	}{
		// case
		{input: &CRSpecBuilderFromCobraFlags{
			blackDuckSpec: &blackduckapi.BlackduckSpec{},
			Size:          "notValid",
		},
			flagNameToTest: "size",
			flagValue:      "notValid",
		},
		// case
		{input: &CRSpecBuilderFromCobraFlags{
			blackDuckSpec: &blackduckapi.BlackduckSpec{},
			Environs:      []string{"invalid"},
		},
			flagNameToTest: "environs",
			flagValue:      "invalid",
		},
		{input: &CRSpecBuilderFromCobraFlags{
			blackDuckSpec: &blackduckapi.BlackduckSpec{},
			ExposeService: "",
		},
			flagNameToTest: "expose-ui",
			flagValue:      "",
		},
	}

	for _, test := range tests {
		cmd := &cobra.Command{}
		blackDuckCobraHelper.AddCRSpecFlagsToCommand(cmd, true)
		flagset := cmd.Flags()
		flagset.Set(test.flagNameToTest, test.flagValue)
		err := test.input.CheckValuesFromFlags(flagset)
		if err == nil {
			t.Errorf("Expected an error but got nil, test: %+v", test)
		}
	}
}
func TestSetPredefinedCRSpec(t *testing.T) {
	assert := assert.New(t)
	blackDuckCtl := NewCRSpecBuilderFromCobraFlags()
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
		assert.Nil(blackDuckCtl.SetPredefinedCRSpec(test.input))
		assert.Equal(*test.expected, blackDuckCtl.GetCRSpec())
	}

	// test cases: ""
	createBlackDuckSpecType := ""
	assert.Error(blackDuckCtl.SetPredefinedCRSpec(createBlackDuckSpecType))
}

func TestAddCRSpecFlagsToCommand(t *testing.T) {
	assert := assert.New(t)

	ctl := NewCRSpecBuilderFromCobraFlags()
	actualCmd := &cobra.Command{}
	ctl.AddCRSpecFlagsToCommand(actualCmd, true)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "Name of Storage Class for the PVC")
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Black Duck has persistent storage [true|false]")
	cmd.Flags().StringVar(&ctl.PVCFilePath, "pvc-file-path", ctl.PVCFilePath, "Absolute path to a file containing a list of PVC json structs")
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "Size of Black Duck [small|medium|large|x-large]")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Black Duck")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type of Black Duck webserver's user interface [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
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
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of environment variables (NAME:VALUE,NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key of Black Duck")
	cmd.Flags().StringVar(&ctl.AdminPassword, "admin-password", ctl.AdminPassword, "'admin' password of Postgres database")
	cmd.Flags().StringVar(&ctl.PostgresPassword, "postgres-password", ctl.PostgresPassword, "'postgres' password of Postgres database")
	cmd.Flags().StringVar(&ctl.UserPassword, "user-password", ctl.UserPassword, "'user' password of Postgres database")
	cmd.Flags().BoolVar(&ctl.EnableBinaryAnalysis, "enable-binary-analysis", ctl.EnableBinaryAnalysis, "If true, enable binary analysis by setting the environment variable (this takes priority over environs flag values)")
	cmd.Flags().BoolVar(&ctl.EnableSourceCodeUpload, "enable-source-code-upload", ctl.EnableSourceCodeUpload, "If true, enable source code upload by setting the environment variable (this takes priority over environs flag values)")
	cmd.Flags().StringVar(&ctl.NodeAffinityFilePath, "node-affinity-file-path", ctl.NodeAffinityFilePath, "Absolute path to a file containing a list of node affinities")
	cmd.Flags().StringVar(&ctl.SecurityContextFilePath, "security-context-file-path", ctl.SecurityContextFilePath, "Absolute path to a file containing a map of pod names to security contexts runAsUser, fsGroup, and runAsGroup")
	cmd.Flags().StringVar(&ctl.Registry, "registry", ctl.Registry, "Name of the registry to use for images e.g. docker.io/blackducksoftware")
	cmd.Flags().StringSliceVar(&ctl.PullSecrets, "pull-secret-name", ctl.PullSecrets, "Only if the registry requires authentication")
	cmd.Flags().StringVar(&ctl.SealKey, "seal-key", ctl.SealKey, "Seal key to encrypt the master key when Source code upload is enabled and it should be of length 32")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("desired-state", "desired-state flag is deprecated and will be removed by the next release")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestGenerateCRSpecFromFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewCRSpecBuilderFromCobraFlags()
	cmd := &cobra.Command{}
	actualCtl.AddCRSpecFlagsToCommand(cmd, true)
	actualCtl.GenerateCRSpecFromFlags(cmd.Flags())

	expCtl := NewCRSpecBuilderFromCobraFlags()
	sort.Strings(expCtl.blackDuckSpec.Environs)
	sort.Strings(actualCtl.blackDuckSpec.Environs)

	assert.Equal(expCtl.blackDuckSpec, actualCtl.blackDuckSpec)

}

func TestAddEnvironsFlagValues(t *testing.T) {
	// enable-binary-analysis  - USE_BINARY_UPLOADS:0
	// enable-source-code-upload - ENABLE_SOURCE_UPLOADS:false
	assert := assert.New(t)

	var tests = []struct {
		enableBinaryAnalysisFlagChanged   bool
		enableBinaryAnalysisFlagValue     string
		enableSourceCodeUploadFlagChanged bool
		enableSourceCodeUploadFlagValue   string
		environsFlagValue                 string
		environsInitiallyInSpec           []string
		expectedEnvirons                  []string
	}{
		{ // case : setting binary analysis to true
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "true",
			enableSourceCodeUploadFlagChanged: false,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"USE_BINARY_UPLOADS:1"},
		},
		{ // case : setting binary analysis to false
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "false",
			enableSourceCodeUploadFlagChanged: false,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"USE_BINARY_UPLOADS:0"},
		},
		{ // case : setting source code upload to true
			enableBinaryAnalysisFlagChanged:   false,
			enableBinaryAnalysisFlagValue:     "false",
			enableSourceCodeUploadFlagChanged: true,
			enableSourceCodeUploadFlagValue:   "true",
			environsFlagValue:                 "",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"ENABLE_SOURCE_UPLOADS:true"},
		},
		{ // case : setting source code upload to false
			enableBinaryAnalysisFlagChanged:   false,
			enableBinaryAnalysisFlagValue:     "false",
			enableSourceCodeUploadFlagChanged: true,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"ENABLE_SOURCE_UPLOADS:false"},
		},
		{ // case : adding environs
			enableBinaryAnalysisFlagChanged:   false,
			enableBinaryAnalysisFlagValue:     "false",
			enableSourceCodeUploadFlagChanged: false,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "a:a,z:z",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"a:a", "z:z"},
		},
		{ // case : adding environs and additional environ flags
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "true",
			enableSourceCodeUploadFlagChanged: true,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "a:a,z:z",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"a:a", "USE_BINARY_UPLOADS:1", "ENABLE_SOURCE_UPLOADS:false", "z:z"},
		},
		{ // case : adding environs and additional environ flags, to flags already in the spec
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "true",
			enableSourceCodeUploadFlagChanged: true,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "a:a",
			environsInitiallyInSpec:           []string{"z:z"},
			expectedEnvirons:                  []string{"a:a", "USE_BINARY_UPLOADS:1", "ENABLE_SOURCE_UPLOADS:false", "z:z"},
		},
		{ // case : additional environ flags take priority over --enirons flag
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "false",
			enableSourceCodeUploadFlagChanged: false,
			enableSourceCodeUploadFlagValue:   "false",
			environsFlagValue:                 "USE_BINARY_UPLOADS:1",
			environsInitiallyInSpec:           []string{},
			expectedEnvirons:                  []string{"USE_BINARY_UPLOADS:0"},
		},
		{ // case : flags take priority over values already in the spec
			enableBinaryAnalysisFlagChanged:   true,
			enableBinaryAnalysisFlagValue:     "true",
			enableSourceCodeUploadFlagChanged: true,
			enableSourceCodeUploadFlagValue:   "true",
			environsFlagValue:                 "a:z",
			environsInitiallyInSpec:           []string{"a:a", "USE_BINARY_UPLOADS:0", "ENABLE_SOURCE_UPLOADS:false"},
			expectedEnvirons:                  []string{"a:z", "USE_BINARY_UPLOADS:1", "ENABLE_SOURCE_UPLOADS:true"},
		},
	}

	for _, test := range tests {
		CRSpecBuilder := NewCRSpecBuilderFromCobraFlags()

		// Get a flagset
		cmd := &cobra.Command{}
		CRSpecBuilder.AddCRSpecFlagsToCommand(cmd, true)
		flagset := cmd.Flags()

		// Set the flagset values based on the test
		if test.enableBinaryAnalysisFlagChanged {
			flagset.Set("enable-binary-analysis", test.enableBinaryAnalysisFlagValue)
		}
		if test.enableSourceCodeUploadFlagChanged {
			flagset.Set("enable-source-code-upload", test.enableSourceCodeUploadFlagValue)
		}
		if test.environsFlagValue != "" {
			flagset.Set("environs", test.environsFlagValue)
		}

		// Set the initial values in the CRSpecBuilder's BlackDuck Spec
		if len(test.environsInitiallyInSpec) > 0 {
			CRSpecBuilder.blackDuckSpec.Environs = test.environsInitiallyInSpec
		}

		// Run the command being tested
		CRSpecBuilder.addEnvironsFlagValues(flagset)

		// Verify the test passed
		sort.Strings(CRSpecBuilder.blackDuckSpec.Environs)
		sort.Strings(test.expectedEnvirons)
		assert.Equal(CRSpecBuilder.blackDuckSpec.Environs, test.expectedEnvirons)

	}
}

func TestSetCRSpecFieldByFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *CRSpecBuilderFromCobraFlags
		changedCtl  *CRSpecBuilderFromCobraFlags
		changedSpec *blackduckapi.BlackduckSpec
	}{
		// case
		{
			flagName:   "size",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				Size:          "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Size: "changed"},
		},
		// case
		{
			flagName:   "version",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				Version:       "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Version: "changed"},
		},
		// case
		{
			flagName:   "expose-ui",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				ExposeService: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{ExposeService: "changed"},
		},
		// case
		{
			flagName:   "db-prototype",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				DbPrototype:   "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{DbPrototype: "changed"},
		},
		// case
		{
			flagName:   "external-postgres-host",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:        &blackduckapi.BlackduckSpec{},
				ExternalPostgresHost: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresHost: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-port",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:        &blackduckapi.BlackduckSpec{},
				ExternalPostgresPort: 10,
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresPort: 10}},
		},
		// case
		{
			flagName:   "external-postgres-admin",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:         &blackduckapi.BlackduckSpec{},
				ExternalPostgresAdmin: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresAdmin: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-user",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:        &blackduckapi.BlackduckSpec{},
				ExternalPostgresUser: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresUser: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:       &blackduckapi.BlackduckSpec{},
				ExternalPostgresSsl: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresSsl: false}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:       &blackduckapi.BlackduckSpec{},
				ExternalPostgresSsl: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresSsl: true}},
		},
		// case
		{
			flagName:   "external-postgres-admin-password",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:                 &blackduckapi.BlackduckSpec{},
				ExternalPostgresAdminPassword: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresAdminPassword: util.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "external-postgres-user-password",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:                &blackduckapi.BlackduckSpec{},
				ExternalPostgresUserPassword: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{
				ExternalPostgres: &blackduckapi.PostgresExternalDBConfig{
					PostgresUserPassword: util.Base64Encode([]byte("changed"))}},
		},
		// case
		{
			flagName:   "pvc-storage-class",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:   &blackduckapi.BlackduckSpec{},
				PvcStorageClass: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVCStorageClass: "changed"},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:  &blackduckapi.BlackduckSpec{},
				LivenessProbes: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LivenessProbes: false},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:  &blackduckapi.BlackduckSpec{},
				LivenessProbes: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LivenessProbes: true},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:     &blackduckapi.BlackduckSpec{},
				PersistentStorage: "false",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PersistentStorage: false},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:     &blackduckapi.BlackduckSpec{},
				PersistentStorage: "true",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PersistentStorage: true},
		},
		// case
		{
			flagName:   "pvc-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				PVCFilePath:   "../../examples/synopsysctl/pvc.json",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{
				{Name: "blackduck-postgres", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-cfssl", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-registration", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-zookeeper", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-authentication", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-webapp", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-logstash", Size: "1Gi", StorageClass: "", VolumeName: ""},
				{Name: "blackduck-uploadcache-data", Size: "1Gi", StorageClass: "", VolumeName: ""},
			}},
		},
		// case
		{
			flagName:   "node-affinity-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:        &blackduckapi.BlackduckSpec{},
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
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:     &blackduckapi.BlackduckSpec{},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: append postgres-claim with size if PVC doesn't exist
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:     &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "other-pvc", Size: "other-size"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "other-pvc", Size: "other-size"}, {Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case: update postgres-claim with size if PVC exists
		{
			flagName:   "postgres-claim-size",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:     &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "unchanged"}}},
				PostgresClaimSize: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{PVC: []blackduckapi.PVC{{Name: "blackduck-postgres", Size: "changed"}}},
		},
		// case
		{
			flagName:   "certificate-name",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:   &blackduckapi.BlackduckSpec{},
				CertificateName: "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{CertificateName: "changed"},
		},
		// case
		{
			flagName:   "certificate-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:       &blackduckapi.BlackduckSpec{},
				CertificateFilePath: "../../examples/synopsysctl/certificate.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Certificate: "CERTIFICATE"},
		},
		// case
		{
			flagName:   "certificate-key-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:          &blackduckapi.BlackduckSpec{},
				CertificateKeyFilePath: "../../examples/synopsysctl/certificateKey.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{CertificateKey: "CERTIFICATE_KEY=CERTIFICATE_KEY_DATA"},
		},
		// case
		{
			flagName:   "proxy-certificate-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:            &blackduckapi.BlackduckSpec{},
				ProxyCertificateFilePath: "../../examples/synopsysctl/proxyCertificate.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{ProxyCertificate: "PROXY_CERTIFICATE"},
		},
		// case
		{
			flagName:   "auth-custom-ca-file-path",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:        &blackduckapi.BlackduckSpec{},
				AuthCustomCAFilePath: "../../examples/synopsysctl/authCustomCA.txt",
			},
			changedSpec: &blackduckapi.BlackduckSpec{AuthCustomCA: "AUTH_CUSTOM_CA"},
		},
		// case
		{
			flagName:   "type",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				Type:          "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{Type: "changed"},
		},
		// case
		{
			flagName:   "desired-state",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				DesiredState:  "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{DesiredState: "changed"},
		},
		// case
		{
			flagName:   "migration-mode",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				MigrationMode: true,
			},
			changedSpec: &blackduckapi.BlackduckSpec{DesiredState: "DbMigrate"},
		},
		// case
		{
			// TODO: add a check for name:Val
			flagName:   "image-registries",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec:   &blackduckapi.BlackduckSpec{},
				ImageRegistries: []string{"changed"},
			},
			changedSpec: &blackduckapi.BlackduckSpec{ImageRegistries: []string{"changed"}},
		},
		// case
		{
			flagName:   "license-key",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				LicenseKey:    "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{LicenseKey: "changed"},
		},
		// case
		{
			flagName:   "seal-key",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{
				blackDuckSpec: &blackduckapi.BlackduckSpec{},
				SealKey:       "changed",
			},
			changedSpec: &blackduckapi.BlackduckSpec{SealKey: util.Base64Encode([]byte("changed"))},
		},
	}

	// get the CRSpecBuilderFromCobraFlags's flags
	cmd := &cobra.Command{}
	actualCtl := NewCRSpecBuilderFromCobraFlags()
	actualCtl.AddCRSpecFlagsToCommand(cmd, true)
	flagset := cmd.Flags()

	for _, test := range tests {
		actualCtl = NewCRSpecBuilderFromCobraFlags()
		// check the Flag exists
		foundFlag := flagset.Lookup(test.flagName)
		if foundFlag == nil {
			t.Errorf("flag %s is not in the spec", test.flagName)
		}
		// check the correct CRSpecBuilderFromCobraFlags is used
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		// test setting a flag
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetCRSpecFieldByFlag(f)
		assert.Equal(test.changedSpec, actualCtl.blackDuckSpec)
	}

	// case: nothing set if flag doesn't exist
	actualCtl = NewCRSpecBuilderFromCobraFlags()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	actualCtl.SetCRSpecFieldByFlag(f)
	assert.Equal(&blackduckapi.BlackduckSpec{}, actualCtl.blackDuckSpec)

}
