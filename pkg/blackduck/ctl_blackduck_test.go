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
	"fmt"
	"testing"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewBlackduckCtl(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackduckCtl()
	assert.Equal(&Ctl{
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
	}, blackduckCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackduckCtl()
	assert.Equal(blackduckv1.BlackduckSpec{}, blackduckCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackduckCtl()
	specToSet := blackduckv1.BlackduckSpec{Namespace: "test"}
	blackduckCtl.SetSpec(specToSet)
	assert.Equal(specToSet, blackduckCtl.GetSpec())

	// check for error
	assert.EqualError(blackduckCtl.SetSpec(""), "Error setting Blackduck Spec")
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)

	// default case
	blackduckCtl := NewBlackduckCtl()
	assert.Nil(blackduckCtl.CheckSpecFlags())

	var tests = []struct {
		input    *Ctl
		expected string
	}{
		// case
		{input: &Ctl{
			Spec: &blackduckv1.BlackduckSpec{},
			Size: "notValid",
		}, expected: "Size must be 'small', 'medium', 'large', or 'xlarge'"},
		// case
		{input: &Ctl{
			Spec:         &blackduckv1.BlackduckSpec{},
			PVCJSONSlice: []string{"invalid:"},
		}, expected: "Invalid format for PVC"},
		// case
		{input: &Ctl{
			Spec:     &blackduckv1.BlackduckSpec{},
			Environs: []string{"invalid"},
		}, expected: "Invalid Environ Format - NAME:VALUE"},
		// case
		{input: &Ctl{
			Spec:                 &blackduckv1.BlackduckSpec{},
			ImageUIDMapJSONSlice: []string{"invalid:"},
			LicenseKey:           "",
		}, expected: "Invalid format for Image UID"},
	}

	for _, test := range tests {
		assert.EqualError(test.input.CheckSpecFlags(), test.expected)
	}
}

func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	blackduckCtl := NewBlackduckCtl()

	var tests = []struct {
		input    string
		expected *blackduckv1.BlackduckSpec
	}{
		{input: "empty", expected: &blackduckv1.BlackduckSpec{}},
		{input: "persistentStorage", expected: crddefaults.GetHubDefaultPersistentStorage()},
		{input: "default", expected: crddefaults.GetHubDefaultValue()},
	}

	// test cases: "empty", "persistentStorage", "default"
	for _, test := range tests {
		assert.Nil(blackduckCtl.SwitchSpec(test.input))
		assert.Equal(*test.expected, blackduckCtl.GetSpec())
	}

	// test cases: ""
	createBlackduckSpecType := ""
	assert.EqualError(blackduckCtl.SwitchSpec(createBlackduckSpecType), fmt.Sprintf("Blackduck Spec Type %s does not match: empty, persistentStorage, default", createBlackduckSpecType))

}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	ctl := NewBlackduckCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd, true)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.Size, "size", ctl.Size, "size - small, medium, large")
	cmd.Flags().StringVar(&ctl.DbPrototype, "db-prototype", ctl.DbPrototype, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresHost, "external-postgres-host", ctl.ExternalPostgresPostgresHost, "Host for Postgres")
	cmd.Flags().IntVar(&ctl.ExternalPostgresPostgresPort, "external-postgres-port", ctl.ExternalPostgresPostgresPort, "Port for Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdmin, "external-postgres-admin", ctl.ExternalPostgresPostgresAdmin, "Name of Admin for Postgres")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUser, "external-postgres-user", ctl.ExternalPostgresPostgresUser, "Username for Postgres")
	cmd.Flags().BoolVar(&ctl.ExternalPostgresPostgresSsl, "external-postgres-ssl", ctl.ExternalPostgresPostgresSsl, "TODO")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresAdminPassword, "external-postgres-admin-password", ctl.ExternalPostgresPostgresAdminPassword, "Password for the Postgres Admin")
	cmd.Flags().StringVar(&ctl.ExternalPostgresPostgresUserPassword, "external-postgres-user-password", ctl.ExternalPostgresPostgresUserPassword, "Password for a Postgres User")
	cmd.Flags().StringVar(&ctl.PvcStorageClass, "pvc-storage-class", ctl.PvcStorageClass, "TODO")
	cmd.Flags().BoolVar(&ctl.LivenessProbes, "liveness-probes", ctl.LivenessProbes, "Enable liveness probes")
	cmd.Flags().StringVar(&ctl.ScanType, "scan-type", ctl.ScanType, "TODO")
	cmd.Flags().BoolVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "Enable persistent storage")
	cmd.Flags().StringSliceVar(&ctl.PVCJSONSlice, "pvc", ctl.PVCJSONSlice, "List of PVC json structs")
	cmd.Flags().StringVar(&ctl.CertificateName, "db-certificate-name", ctl.CertificateName, "TODO")
	cmd.Flags().StringVar(&ctl.Certificate, "certificate", ctl.Certificate, "TODO")
	cmd.Flags().StringVar(&ctl.CertificateKey, "certificate-key", ctl.CertificateKey, "TODO")
	cmd.Flags().StringVar(&ctl.ProxyCertificate, "proxy-certificate", ctl.ProxyCertificate, "TODO")
	cmd.Flags().StringVar(&ctl.Type, "type", ctl.Type, "Type of Blackduck")
	cmd.Flags().StringVar(&ctl.DesiredState, "desired-state", ctl.DesiredState, "Desired state of Blackduck")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "List of Environment Variables (NAME:VALUE)")
	cmd.Flags().StringSliceVar(&ctl.ImageRegistries, "image-registries", ctl.ImageRegistries, "List of image registries")
	cmd.Flags().StringSliceVar(&ctl.ImageUIDMapJSONSlice, "image-uid-map", ctl.ImageUIDMapJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.LicenseKey, "license-key", ctl.LicenseKey, "License Key for the Knowledge Base")
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Blackduck Version")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-service", ctl.ExposeService, "Expose service type [Loadbalancer/Nodeport]")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewBlackduckCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, true)
	actualCtl.SetChangedFlags(cmd.Flags())

	expCtl := NewBlackduckCtl()

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
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{},
				Size: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Size: "changed"},
		},
		// case
		{
			flagName:   "db-prototype",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:        &blackduckv1.BlackduckSpec{},
				DbPrototype: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{DbPrototype: "changed"},
		},
		// case
		{
			flagName:   "external-postgres-host",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                         &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresHost: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresHost: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-port",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                         &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresPort: 10,
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresPort: 10}},
		},
		// case
		{
			flagName:   "external-postgres-admin",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                          &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresAdmin: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresAdmin: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-user",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                         &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresUser: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresUser: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-ssl",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                        &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresSsl: true,
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresSsl: true}},
		},
		// case
		{
			flagName:   "external-postgres-admin-password",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                                  &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresAdminPassword: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresAdminPassword: "changed"}},
		},
		// case
		{
			flagName:   "external-postgres-user-password",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                                 &blackduckv1.BlackduckSpec{},
				ExternalPostgresPostgresUserPassword: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{
				ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
					PostgresUserPassword: "changed"}},
		},
		// case
		{
			flagName:   "pvc-storage-class",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				PvcStorageClass: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVCStorageClass: "changed"},
		},
		// case
		{
			flagName:   "liveness-probes",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckv1.BlackduckSpec{},
				LivenessProbes: true,
			},
			changedSpec: &blackduckv1.BlackduckSpec{LivenessProbes: true},
		},
		// case
		{
			flagName:   "scan-type",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:     &blackduckv1.BlackduckSpec{},
				ScanType: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{ScanType: "changed"},
		},
		// case
		{
			flagName:   "persistent-storage",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:              &blackduckv1.BlackduckSpec{},
				PersistentStorage: true,
			},
			changedSpec: &blackduckv1.BlackduckSpec{PersistentStorage: true},
		},
		// case
		{
			flagName:   "pvc",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:         &blackduckv1.BlackduckSpec{},
				PVCJSONSlice: []string{"{\"name\": \"changed\", \"size\": \"1G\"}"},
			},
			changedSpec: &blackduckv1.BlackduckSpec{PVC: []blackduckv1.PVC{{Name: "changed", Size: "1G"}}},
		},
		// case
		{
			flagName:   "db-certificate-name",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				CertificateName: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{CertificateName: "changed"},
		},
		// case
		{
			flagName:   "certificate",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:        &blackduckv1.BlackduckSpec{},
				Certificate: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Certificate: "changed"},
		},
		// case
		{
			flagName:   "certificate-key",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:           &blackduckv1.BlackduckSpec{},
				CertificateKey: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{CertificateKey: "changed"},
		},
		// case
		{
			flagName:   "proxy-certificate",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:             &blackduckv1.BlackduckSpec{},
				ProxyCertificate: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{ProxyCertificate: "changed"},
		},
		// case
		{
			flagName:   "type",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{},
				Type: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{Type: "changed"},
		},
		// case
		{
			flagName:   "desired-state",
			initialCtl: NewBlackduckCtl(),
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
			initialCtl: NewBlackduckCtl(),
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
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:            &blackduckv1.BlackduckSpec{},
				ImageRegistries: []string{"changed"},
			},
			changedSpec: &blackduckv1.BlackduckSpec{ImageRegistries: []string{"changed"}},
		},
		// case
		{
			flagName:   "image-uid-map",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:                 &blackduckv1.BlackduckSpec{},
				ImageUIDMapJSONSlice: []string{"{\"Key\": \"changed\", \"Value\": 1}"},
			},
			changedSpec: &blackduckv1.BlackduckSpec{ImageUIDMap: map[string]int64{"changed": 1}},
		},
		// case
		{
			flagName:   "license-key",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec:       &blackduckv1.BlackduckSpec{},
				LicenseKey: "changed",
			},
			changedSpec: &blackduckv1.BlackduckSpec{LicenseKey: "changed"},
		},
		// case
		{
			flagName:   "",
			initialCtl: NewBlackduckCtl(),
			changedCtl: &Ctl{
				Spec: &blackduckv1.BlackduckSpec{},
			},
			changedSpec: &blackduckv1.BlackduckSpec{},
		},
	}

	for _, test := range tests {
		actualCtl := NewBlackduckCtl()
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetFlag(f)
		assert.Equal(test.changedSpec, actualCtl.Spec)
	}

}
