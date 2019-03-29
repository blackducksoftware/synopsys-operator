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
	"fmt"
	"testing"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewAlertCtl(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	assert.Equal(&Ctl{
		Spec:                 &alertapi.AlertSpec{},
		Registry:             "",
		ImagePath:            "",
		AlertImageName:       "",
		AlertImageVersion:    "",
		CfsslImageName:       "",
		CfsslImageVersion:    "",
		StandAlone:           false,
		Port:                 0,
		Environs:             []string{},
		EncryptionPassword:   "",
		EncryptionGlobalSalt: "",
		PersistentStorage:    false,
		PVCName:              "",
		PVCStorageClass:      "",
		PVCSize:              "",
		AlertMemory:          "",
		CfsslMemory:          "",
		DesiredState:         "",
	}, alertCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	assert.Equal(alertapi.AlertSpec{}, alertCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	specToSet := alertapi.AlertSpec{Namespace: "test", Registry: "test"}
	alertCtl.SetSpec(specToSet)
	assert.Equal(specToSet, alertCtl.GetSpec())

	// check for error
	assert.EqualError(alertCtl.SetSpec(""), "Error setting Alert Spec")
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	specFlags := alertCtl.CheckSpecFlags()
	assert.Nil(specFlags)
}

func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()

	var tests = []struct {
		input    string
		expected alertapi.AlertSpec
	}{
		{input: "empty", expected: alertapi.AlertSpec{}},
		{input: "default", expected: *crddefaults.GetAlertDefaultValue()},
	}

	// test cases: "empty", "spec1", "spec2"
	for _, test := range tests {
		assert.Nil(alertCtl.SwitchSpec(test.input))
		assert.Equal(test.expected, alertCtl.GetSpec())
	}

	// test cases: default
	createAlertSpecType := ""
	assert.EqualError(alertCtl.SwitchSpec(createAlertSpecType),
		fmt.Sprintf("Alert Spec Type %s does not match: default or empty", createAlertSpecType))

}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	ctl := NewAlertCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd)

	expectedCmd := &cobra.Command{}
	expectedCmd.Flags().StringVar(&ctl.Registry, "alert-registry", ctl.Registry, "Registry with the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.ImagePath, "image-path", ctl.ImagePath, "Path to the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.AlertImageName, "alert-image-name", ctl.AlertImageName, "Name of the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.AlertImageVersion, "alert-image-version", ctl.AlertImageVersion, "Version of the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.CfsslImageName, "cfssl-image-name", ctl.CfsslImageName, "Name of Cfssl Image")
	expectedCmd.Flags().StringVar(&ctl.CfsslImageVersion, "cfssl-image-version", ctl.CfsslImageVersion, "Version of Cffsl Image")
	expectedCmd.Flags().BoolVar(&ctl.StandAlone, "stand-alone", ctl.StandAlone, "Enable Stand Alone mode")
	expectedCmd.Flags().IntVar(&ctl.Port, "port", ctl.Port, "Port for Alert")
	expectedCmd.Flags().StringVar(&ctl.EncryptionPassword, "encryption-password", ctl.EncryptionPassword, "Encryption Password for the Alert")
	expectedCmd.Flags().StringVar(&ctl.EncryptionGlobalSalt, "encryption-global-salt", ctl.EncryptionGlobalSalt, "Encryption Global Salt for the Alert")
	expectedCmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "Environment variables for the Alert")
	expectedCmd.Flags().BoolVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "Enable persistent storage")
	expectedCmd.Flags().StringVar(&ctl.PVCName, "pvc-name", ctl.PVCName, "Name for the PVC")
	expectedCmd.Flags().StringVar(&ctl.PVCStorageClass, "pvc-storage-class", ctl.PVCStorageClass, "StorageClass for the PVC")
	expectedCmd.Flags().StringVar(&ctl.PVCSize, "pvc-size", ctl.PVCSize, "Memory allocation for the PVC")
	expectedCmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation for the Alert")
	expectedCmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation for the Cfssl")
	expectedCmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of the Alert")

	assert.Equal(expectedCmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewAlertCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd)
	actualCtl.SetChangedFlags(cmd.Flags())

	expCtl := NewAlertCtl()

	assert.Equal(expCtl.Spec, actualCtl.Spec)

}

func TestSetFlag(t *testing.T) {

	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *Ctl
		changedCtl  *Ctl
		changedSpec *alertapi.AlertSpec
	}{
		// case
		{flagName: "alert-registry",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "changedRegistry",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{Registry: "changedRegistry"},
		},
		// case
		{flagName: "image-path",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "changedImagePath",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{ImagePath: "changedImagePath"},
		},
		// case
		{flagName: "alert-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "changedAlertImageName",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{AlertImageName: "changedAlertImageName"},
		},
		// case
		{flagName: "alert-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "changedAlertImageVersion",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{AlertImageVersion: "changedAlertImageVersion"},
		},
		// case
		{flagName: "cfssl-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "changedCfsslImageName",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{CfsslImageName: "changedCfsslImageName"},
		},
		// case
		{flagName: "cfssl-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "changedCfsslImageVersion",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{CfsslImageVersion: "changedCfsslImageVersion"},
		},
		// case
		{flagName: "stand-alone",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           true,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{StandAlone: crddefaults.BoolToPtr(true)},
		},
		// case
		{flagName: "port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 10,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{Port: crddefaults.IntToPtr(10)},
		},
		// case
		{flagName: "encryption-password",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "changedEncryptionPassword",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionPassword: "changedEncryptionPassword"},
		},
		// case
		{flagName: "encryption-global-salt",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "changedEncryptionGlobalSalt",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionGlobalSalt: "changedEncryptionGlobalSalt"},
		},
		// case
		{flagName: "environs",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{"changedEnviron:Number1", "changedEnviron:Number2"},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{Environs: []string{"changedEnviron:Number1", "changedEnviron:Number2"}},
		},
		// case
		{flagName: "persistent-storage",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    true,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{PersistentStorage: true},
		},
		// case
		{flagName: "pvc-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "changedPVCName",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{PVCName: "changedPVCName"},
		},
		// case
		{flagName: "storage-class",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "changedStorageClass",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{PVCStorageClass: "changedStorageClass"},
		},
		// case
		{flagName: "pvc-size",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "changedStorageSize",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{PVCSize: "changedStorageSize"},
		},
		// case
		{flagName: "alert-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "changedAlertMemory",
				CfsslMemory:          "",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{AlertMemory: "changedAlertMemory"},
		},
		// case
		{flagName: "cfssl-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "changedCfsslMemory",
				DesiredState:         "",
			},
			changedSpec: &alertapi.AlertSpec{CfsslMemory: "changedCfsslMemory"},
		},
		// case
		{flagName: "alert-desired-state",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Registry:             "",
				ImagePath:            "",
				AlertImageName:       "",
				AlertImageVersion:    "",
				CfsslImageName:       "",
				CfsslImageVersion:    "",
				StandAlone:           false,
				Port:                 0,
				EncryptionPassword:   "",
				EncryptionGlobalSalt: "",
				Environs:             []string{},
				PersistentStorage:    false,
				PVCName:              "",
				PVCStorageClass:      "",
				PVCSize:              "",
				AlertMemory:          "",
				CfsslMemory:          "",
				DesiredState:         "changedState",
			},
			changedSpec: &alertapi.AlertSpec{DesiredState: "changedState"},
		},
	}

	for _, test := range tests {
		actualCtl := NewAlertCtl()
		assert.Equal(test.initialCtl, actualCtl)
		actualCtl = test.changedCtl
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		actualCtl.SetFlag(f)
		assert.Equal(test.changedSpec, actualCtl.Spec)
	}
}
