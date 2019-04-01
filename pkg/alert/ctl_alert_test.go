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

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewAlertCtl(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	assert.Equal(&Ctl{
		Spec:              &alertv1.AlertSpec{},
		Registry:          "",
		ImagePath:         "",
		AlertImageName:    "",
		AlertImageVersion: "",
		CfsslImageName:    "",
		CfsslImageVersion: "",
		BlackduckHost:     "",
		BlackduckUser:     "",
		BlackduckPort:     0,
		Port:              0,
		StandAlone:        false,
		AlertMemory:       "",
		CfsslMemory:       "",
		DesiredState:      "",
	}, alertCtl)
}

func TestGetSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	assert.Equal(alertv1.AlertSpec{}, alertCtl.GetSpec())
}

func TestSetSpec(t *testing.T) {
	assert := assert.New(t)
	alertCtl := NewAlertCtl()
	specToSet := alertv1.AlertSpec{Namespace: "test", Registry: "test"}
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
		expected alertv1.AlertSpec
	}{
		{input: "empty", expected: alertv1.AlertSpec{}},
		{input: "spec1", expected: *crddefaults.GetAlertDefaultValue()},
		{input: "spec2", expected: *crddefaults.GetAlertDefaultValue2()},
	}

	// test cases: "empty", "spec1", "spec2"
	for _, test := range tests {
		assert.Nil(alertCtl.SwitchSpec(test.input))
		assert.Equal(test.expected, alertCtl.GetSpec())
	}

	// test cases: default
	createAlertSpecType := ""
	assert.EqualError(alertCtl.SwitchSpec(createAlertSpecType),
		fmt.Sprintf("Alert Spec Type %s does not match: empty, spec1, spec2", createAlertSpecType))

}

func TestAddSpecFlags(t *testing.T) {
	assert := assert.New(t)

	ctl := NewAlertCtl()
	actualCmd := &cobra.Command{}
	ctl.AddSpecFlags(actualCmd, false)

	expectedCmd := &cobra.Command{}
	expectedCmd.Flags().StringVar(&ctl.Registry, "alert-registry", ctl.Registry, "Registry with the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.ImagePath, "image-path", ctl.ImagePath, "Path to the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.AlertImageName, "alert-image-name", ctl.AlertImageName, "Name of the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.AlertImageVersion, "alert-image-version", ctl.AlertImageVersion, "Version of the Alert Image")
	expectedCmd.Flags().StringVar(&ctl.CfsslImageName, "cfssl-image-name", ctl.CfsslImageName, "Name of Cfssl Image")
	expectedCmd.Flags().StringVar(&ctl.CfsslImageVersion, "cfssl-image-version", ctl.CfsslImageVersion, "Version of Cffsl Image")
	expectedCmd.Flags().StringVar(&ctl.BlackduckHost, "blackduck-host", ctl.BlackduckHost, "Host url of Blackduck")
	expectedCmd.Flags().StringVar(&ctl.BlackduckUser, "blackduck-user", ctl.BlackduckUser, "Username for Blackduck")
	expectedCmd.Flags().IntVar(&ctl.BlackduckPort, "blackduck-port", ctl.BlackduckPort, "Port for Blackduck")
	expectedCmd.Flags().IntVar(&ctl.Port, "port", ctl.Port, "Port for Alert")
	expectedCmd.Flags().BoolVar(&ctl.StandAlone, "stand-alone", ctl.StandAlone, "Enable Stand Alone mode")
	expectedCmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation for the Alert")
	expectedCmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation for the Cfssl")
	expectedCmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of the Alert")

	assert.Equal(expectedCmd.Flags(), actualCmd.Flags())
}

func TestSetChangedFlags(t *testing.T) {
	assert := assert.New(t)

	actualCtl := NewAlertCtl()
	cmd := &cobra.Command{}
	actualCtl.AddSpecFlags(cmd, false)
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
		changedSpec *alertv1.AlertSpec
	}{
		// case
		{flagName: "alert-registry",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "changedRegistry",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{Registry: "changedRegistry"},
		},
		// case
		{flagName: "image-path",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "changedImagePath",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{ImagePath: "changedImagePath"},
		},
		// case
		{flagName: "alert-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "changedAlertImageName",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{AlertImageName: "changedAlertImageName"},
		},
		// case
		{flagName: "alert-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "changedAlertImageVersion",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{AlertImageVersion: "changedAlertImageVersion"},
		},
		// case
		{flagName: "cfssl-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "changedCfsslImageName",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{CfsslImageName: "changedCfsslImageName"},
		},
		// case
		{flagName: "cfssl-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "changedCfsslImageVersion",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{CfsslImageVersion: "changedCfsslImageVersion"},
		},
		// case
		{flagName: "blackduck-host",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "changedBlackduckHost",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{BlackduckHost: "changedBlackduckHost"},
		},
		// case
		{flagName: "blackduck-user",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "changedBlackduckUser",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{BlackduckUser: "changedBlackduckUser"},
		},
		// case
		{flagName: "blackduck-port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     10,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{BlackduckPort: crddefaults.IntToPtr(10)},
		},
		// case
		{flagName: "port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              10,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{Port: crddefaults.IntToPtr(10)},
		},
		// case
		{flagName: "stand-alone",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        true,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{StandAlone: crddefaults.BoolToPtr(true)},
		},
		// case
		{flagName: "alert-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "changedAlertMemory",
				CfsslMemory:       "",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{AlertMemory: "changedAlertMemory"},
		},
		// case
		{flagName: "cfssl-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "changedCfsslMemory",
				DesiredState:      "",
			},
			changedSpec: &alertv1.AlertSpec{CfsslMemory: "changedCfsslMemory"},
		},
		// case
		{flagName: "alert-desired-state",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry:          "",
				ImagePath:         "",
				AlertImageName:    "",
				AlertImageVersion: "",
				CfsslImageName:    "",
				CfsslImageVersion: "",
				BlackduckHost:     "",
				BlackduckUser:     "",
				BlackduckPort:     0,
				Port:              0,
				StandAlone:        false,
				AlertMemory:       "",
				CfsslMemory:       "",
				DesiredState:      "changedState",
			},
			changedSpec: &alertv1.AlertSpec{DesiredState: "changedState"},
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
