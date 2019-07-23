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

package alert

import (
	"testing"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewCRSpecBuilderFromCobraFlags(t *testing.T) {
	assert := assert.New(t)
	alertCobraHelper := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(&CRSpecBuilderFromCobraFlags{
		alertSpec: &alertapi.AlertSpec{},
	}, alertCobraHelper)
}

func TestGetCRSpec(t *testing.T) {
	assert := assert.New(t)
	alertCobraHelper := NewCRSpecBuilderFromCobraFlags()
	assert.Equal(alertapi.AlertSpec{}, alertCobraHelper.GetCRSpec())
}

func TestSetCRSpec(t *testing.T) {
	assert := assert.New(t)
	alertCobraHelper := NewCRSpecBuilderFromCobraFlags()
	specToSet := alertapi.AlertSpec{Namespace: "test", Version: "test"}
	alertCobraHelper.SetCRSpec(specToSet)
	assert.Equal(specToSet, alertCobraHelper.GetCRSpec())

	// check for error
	assert.Error(alertCobraHelper.SetCRSpec(""))
}

func TestCheckValuesFromFlags(t *testing.T) {
	assert := assert.New(t)
	alertCobraHelper := NewCRSpecBuilderFromCobraFlags()
	alertCobraHelper.ExposeService = util.NONE
	cmd := &cobra.Command{}
	specFlags := alertCobraHelper.CheckValuesFromFlags(cmd.Flags())
	assert.Nil(specFlags)

	var tests = []struct {
		input          *CRSpecBuilderFromCobraFlags
		flagNameToTest string
		flagValue      string
	}{
		// invalid expose case
		{input: &CRSpecBuilderFromCobraFlags{
			alertSpec:     &alertapi.AlertSpec{},
			ExposeService: "",
		},
			flagNameToTest: "expose-ui",
			flagValue:      "",
		},
	}

	for _, test := range tests {
		cmd := &cobra.Command{}
		alertCobraHelper.AddCRSpecFlagsToCommand(cmd, true)
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
	alertCobraHelper := NewCRSpecBuilderFromCobraFlags()
	defaultSpec := *util.GetAlertDefault()
	defaultSpec.StandAlone = util.BoolToPtr(true)
	defaultSpec.PersistentStorage = true

	var tests = []struct {
		input    string
		expected alertapi.AlertSpec
	}{
		{input: EmptySpec, expected: alertapi.AlertSpec{}},
		{input: DefaultSpec, expected: defaultSpec},
	}

	// test cases: "default"
	for _, test := range tests {
		assert.Nil(alertCobraHelper.SetPredefinedCRSpec(test.input))
		assert.Equal(test.expected, alertCobraHelper.GetCRSpec())
	}

	// test cases: default
	createAlertSpecType := ""
	assert.Error(alertCobraHelper.SetPredefinedCRSpec(createAlertSpecType))

}

func TestAddCRSpecFlagsToCommand(t *testing.T) {
	assert := assert.New(t)

	// test case: Only Non-Master Flags are added
	ctl := NewCRSpecBuilderFromCobraFlags()
	actualCmd := &cobra.Command{}
	ctl.AddCRSpecFlagsToCommand(actualCmd, false)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Alert")
	cmd.Flags().StringVar(&ctl.AlertImage, "alert-image", ctl.AlertImage, "URL of Alert's Image")
	cmd.Flags().StringVar(&ctl.CfsslImage, "cfssl-image", ctl.CfsslImage, "URL of CFSSL's Image")
	cmd.Flags().StringVar(&ctl.StandAlone, "standalone", ctl.StandAlone, "If true, Alert runs in standalone mode (true|false)")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type to expose Alert's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	cmd.Flags().Int32Var(&ctl.Port, "port", ctl.Port, "Port of Alert")
	cmd.Flags().StringVar(&ctl.EncryptionPassword, "encryption-password", ctl.EncryptionPassword, "Encryption Password for Alert")
	cmd.Flags().StringVar(&ctl.EncryptionGlobalSalt, "encryption-global-salt", ctl.EncryptionGlobalSalt, "Encryption Global Salt for Alert")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "Environment variables of Alert")
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Alert has persistent storage (true|false)")
	cmd.Flags().StringVar(&ctl.PVCName, "pvc-name", ctl.PVCName, "Name of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCStorageClass, "pvc-storage-class", ctl.PVCStorageClass, "Storage class for the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCSize, "pvc-size", ctl.PVCSize, "Memory allocation of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation of Alert")
	cmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation of CFSSL")
	cmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of Alert")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("alert-desired-state", "alert-desired-state flag is deprecated and will be removed by the next release")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestGenerateCRSpecFromFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewCRSpecBuilderFromCobraFlags()
	cmd := &cobra.Command{}
	actualCtl.AddCRSpecFlagsToCommand(cmd, true)
	actualCtl.GenerateCRSpecFromFlags(cmd.Flags())

	expCtl := NewCRSpecBuilderFromCobraFlags()

	assert.Equal(expCtl.alertSpec, actualCtl.alertSpec)

}

func TestSetCRSpecFieldByFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *CRSpecBuilderFromCobraFlags
		changedCtl  *CRSpecBuilderFromCobraFlags
		changedSpec *alertapi.AlertSpec
	}{
		// case
		{flagName: "version",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				Version: "changed",
			},
			changedSpec: &alertapi.AlertSpec{Version: "changed"},
		},
		// case
		{flagName: "alert-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				AlertImage: "changed",
			},
			changedSpec: &alertapi.AlertSpec{AlertImage: "changed"},
		},
		// case
		{flagName: "cfssl-image",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				CfsslImage: "changed",
			},
			changedSpec: &alertapi.AlertSpec{CfsslImage: "changed"},
		},
		// case
		{flagName: "standalone",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				StandAlone: "true",
			},
			changedSpec: &alertapi.AlertSpec{StandAlone: util.BoolToPtr(true)},
		},
		// case
		{flagName: "standalone",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				StandAlone: "false",
			},
			changedSpec: &alertapi.AlertSpec{StandAlone: util.BoolToPtr(false)},
		},
		// case
		{flagName: "expose-ui",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				ExposeService: "changed",
			},
			changedSpec: &alertapi.AlertSpec{ExposeService: "changed"},
		},
		// case
		{flagName: "port",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				Port: 1234,
			},
			changedSpec: &alertapi.AlertSpec{Port: util.IntToInt32(1234)},
		},
		// case
		{flagName: "encryption-password",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				EncryptionPassword: "changedEncryptionPassword",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionPassword: "changedEncryptionPassword"},
		},
		// case
		{flagName: "encryption-global-salt",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				EncryptionGlobalSalt: "changedEncryptionGlobalSalt",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionGlobalSalt: "changedEncryptionGlobalSalt"},
		},
		// case
		{flagName: "environs",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				Environs: []string{"changedEnviron:Number1", "changedEnviron:Number2"},
			},
			changedSpec: &alertapi.AlertSpec{Environs: []string{"changedEnviron:Number1", "changedEnviron:Number2"}},
		},
		// case
		{flagName: "persistent-storage",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				PersistentStorage: "true",
			},
			changedSpec: &alertapi.AlertSpec{PersistentStorage: true},
		},
		// case
		{flagName: "persistent-storage",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				PersistentStorage: "false",
			},
			changedSpec: &alertapi.AlertSpec{PersistentStorage: false},
		},
		// case
		{flagName: "pvc-name",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				PVCName: "changedPVCName",
			},
			changedSpec: &alertapi.AlertSpec{PVCName: "changedPVCName"},
		},
		// case
		{flagName: "pvc-storage-class",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				PVCStorageClass: "changedStorageClass",
			},
			changedSpec: &alertapi.AlertSpec{PVCStorageClass: "changedStorageClass"},
		},
		// case
		{flagName: "pvc-size",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				PVCSize: "changedStorageSize",
			},
			changedSpec: &alertapi.AlertSpec{PVCSize: "changedStorageSize"},
		},
		// case
		{flagName: "alert-memory",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				AlertMemory: "changed",
			},
			changedSpec: &alertapi.AlertSpec{AlertMemory: "changed"},
		},
		// case
		{flagName: "cfssl-memory",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				CfsslMemory: "changed",
			},
			changedSpec: &alertapi.AlertSpec{CfsslMemory: "changed"},
		},
		// case
		{flagName: "alert-desired-state",
			initialCtl: NewCRSpecBuilderFromCobraFlags(),
			changedCtl: &CRSpecBuilderFromCobraFlags{alertSpec: &alertapi.AlertSpec{},
				DesiredState: "changed",
			},
			changedSpec: &alertapi.AlertSpec{DesiredState: "changed"},
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
		assert.Equal(test.changedSpec, actualCtl.alertSpec)
	}

	// case: nothing set if flag doesn't exist
	actualCtl = NewCRSpecBuilderFromCobraFlags()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	actualCtl.SetCRSpecFieldByFlag(f)
	assert.Equal(&alertapi.AlertSpec{}, actualCtl.alertSpec)

}
