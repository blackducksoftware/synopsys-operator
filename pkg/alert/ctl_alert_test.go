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
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewAlertCtl(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	assert.Equal(&Ctl{
		Spec: &alertapi.AlertSpec{},
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
	specToSet := alertapi.AlertSpec{Namespace: "test", Version: "test"}
	alertCtl.SetSpec(specToSet)
	assert.Equal(specToSet, alertCtl.GetSpec())

	// check for error
	assert.Error(alertCtl.SetSpec(""))
}

func TestCheckSpecFlags(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	cmd := &cobra.Command{}
	specFlags := alertCtl.CheckSpecFlags(cmd.Flags())
	assert.Nil(specFlags)
}

func TestSwitchSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	defaultSpec := *crddefaults.GetAlertDefault()
	defaultSpec.StandAlone = crddefaults.BoolToPtr(true)
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
		assert.Nil(alertCtl.SwitchSpec(test.input))
		assert.Equal(test.expected, alertCtl.GetSpec())
	}

	// test cases: default
	createAlertSpecType := ""
	assert.Error(alertCtl.SwitchSpec(createAlertSpecType))

}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	// test case: Only Non-Master Flags are added
	ctl := NewAlertCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd, false)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Alert")
	cmd.Flags().StringVar(&ctl.AlertImage, "alert-image", ctl.AlertImage, "Url of Alert's Image")
	cmd.Flags().StringVar(&ctl.CfsslImage, "cfssl-image", ctl.CfsslImage, "Url of Cfssl's Image")
	cmd.Flags().StringVar(&ctl.StandAlone, "stand-alone", ctl.StandAlone, "If true, Alert runs in stand alone mode [true|false]")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type to expose Alert's user interface [NODEPORT|LOADBALANCER|OPENSHIFT]")
	cmd.Flags().Int32Var(&ctl.Port, "port", ctl.Port, "Port of Alert")
	cmd.Flags().StringVar(&ctl.EncryptionPassword, "encryption-password", ctl.EncryptionPassword, "Encryption Password for Alert")
	cmd.Flags().StringVar(&ctl.EncryptionGlobalSalt, "encryption-global-salt", ctl.EncryptionGlobalSalt, "Encryption Global Salt for Alert")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "Environment variables of Alert")
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Alert has persistent storage [true|false]")
	cmd.Flags().StringVar(&ctl.PVCName, "pvc-name", ctl.PVCName, "Name of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCStorageClass, "pvc-storage-class", ctl.PVCStorageClass, "StorageClass for the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCSize, "pvc-size", ctl.PVCSize, "Memory allocation of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation of Alert")
	cmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation of the Cfssl")
	cmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of Alert")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("alert-desired-state", "alert-desired-state flag is deprecated and will be removed by the next release")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewAlertCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, true)
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
		{flagName: "version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Version: "changed",
			},
			changedSpec: &alertapi.AlertSpec{Version: "changed"},
		},
		// case
		{flagName: "alert-image",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				AlertImage: "changed",
			},
			changedSpec: &alertapi.AlertSpec{AlertImage: "changed"},
		},
		// case
		{flagName: "cfssl-image",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				CfsslImage: "changed",
			},
			changedSpec: &alertapi.AlertSpec{CfsslImage: "changed"},
		},
		// case
		{flagName: "stand-alone",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				StandAlone: "true",
			},
			changedSpec: &alertapi.AlertSpec{StandAlone: crddefaults.BoolToPtr(true)},
		},
		// case
		{flagName: "stand-alone",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				StandAlone: "false",
			},
			changedSpec: &alertapi.AlertSpec{StandAlone: crddefaults.BoolToPtr(false)},
		},
		// case
		{flagName: "expose-ui",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				ExposeService: "changed",
			},
			changedSpec: &alertapi.AlertSpec{ExposeService: "changed"},
		},
		// case
		{flagName: "port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Port: 1234,
			},
			changedSpec: &alertapi.AlertSpec{Port: crddefaults.IntToInt32(1234)},
		},
		// case
		{flagName: "encryption-password",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				EncryptionPassword: "changedEncryptionPassword",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionPassword: "changedEncryptionPassword"},
		},
		// case
		{flagName: "encryption-global-salt",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				EncryptionGlobalSalt: "changedEncryptionGlobalSalt",
			},
			changedSpec: &alertapi.AlertSpec{EncryptionGlobalSalt: "changedEncryptionGlobalSalt"},
		},
		// case
		{flagName: "environs",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				Environs: []string{"changedEnviron:Number1", "changedEnviron:Number2"},
			},
			changedSpec: &alertapi.AlertSpec{Environs: []string{"changedEnviron:Number1", "changedEnviron:Number2"}},
		},
		// case
		{flagName: "persistent-storage",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				PersistentStorage: "true",
			},
			changedSpec: &alertapi.AlertSpec{PersistentStorage: true},
		},
		// case
		{flagName: "persistent-storage",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				PersistentStorage: "false",
			},
			changedSpec: &alertapi.AlertSpec{PersistentStorage: false},
		},
		// case
		{flagName: "pvc-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				PVCName: "changedPVCName",
			},
			changedSpec: &alertapi.AlertSpec{PVCName: "changedPVCName"},
		},
		// case
		{flagName: "pvc-storage-class",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				PVCStorageClass: "changedStorageClass",
			},
			changedSpec: &alertapi.AlertSpec{PVCStorageClass: "changedStorageClass"},
		},
		// case
		{flagName: "pvc-size",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				PVCSize: "changedStorageSize",
			},
			changedSpec: &alertapi.AlertSpec{PVCSize: "changedStorageSize"},
		},
		// case
		{flagName: "alert-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				AlertMemory: "changed",
			},
			changedSpec: &alertapi.AlertSpec{AlertMemory: "changed"},
		},
		// case
		{flagName: "cfssl-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				CfsslMemory: "changed",
			},
			changedSpec: &alertapi.AlertSpec{CfsslMemory: "changed"},
		},
		// case
		{flagName: "alert-desired-state",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertapi.AlertSpec{},
				DesiredState: "changed",
			},
			changedSpec: &alertapi.AlertSpec{DesiredState: "changed"},
		},
	}

	// get the Ctl's flags
	cmd := &cobra.Command{}
	actualCtl := NewAlertCtl()
	actualCtl.AddSpecFlags(cmd, true)
	flagset := cmd.Flags()

	for _, test := range tests {
		actualCtl = NewAlertCtl()
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
	actualCtl = NewAlertCtl()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	actualCtl.SetFlag(f)
	assert.Equal(&alertapi.AlertSpec{}, actualCtl.Spec)

}
