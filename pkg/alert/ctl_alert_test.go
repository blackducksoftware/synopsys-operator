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
		Spec: &alertv1.AlertSpec{},
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
		expected *alertv1.AlertSpec
	}{
		{input: "empty", expected: &alertv1.AlertSpec{}},
		{input: "spec1", expected: crddefaults.GetAlertDefaultValue()},
		{input: "spec2", expected: crddefaults.GetAlertDefaultValue2()},
	}

	// test cases: "empty", "spec1", "spec2"
	for _, test := range tests {
		assert.Nil(alertCtl.SwitchSpec(test.input))
		assert.Equal(*test.expected, alertCtl.GetSpec())
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
	ctl.AddSpecFlags(actualCmd)

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&ctl.Registry, "alert-registry", ctl.Registry, "Registry with the Alert Image")
	cmd.Flags().StringVar(&ctl.ImagePath, "image-path", ctl.ImagePath, "Path to the Alert Image")
	cmd.Flags().StringVar(&ctl.AlertImageName, "alert-image-name", ctl.AlertImageName, "Name of the Alert Image")
	cmd.Flags().StringVar(&ctl.AlertImageVersion, "alert-image-version", ctl.AlertImageVersion, "Version of the Alert Image")
	cmd.Flags().StringVar(&ctl.CfsslImageName, "cfssl-image-name", ctl.CfsslImageName, "Name of Cfssl Image")
	cmd.Flags().StringVar(&ctl.CfsslImageVersion, "cfssl-image-version", ctl.CfsslImageVersion, "Version of Cffsl Image")
	cmd.Flags().StringVar(&ctl.BlackduckHost, "blackduck-host", ctl.BlackduckHost, "Host url of Blackduck")
	cmd.Flags().StringVar(&ctl.BlackduckUser, "blackduck-user", ctl.BlackduckUser, "Username for Blackduck")
	cmd.Flags().IntVar(&ctl.BlackduckPort, "blackduck-port", ctl.BlackduckPort, "Port for Blackduck")
	cmd.Flags().IntVar(&ctl.Port, "port", ctl.Port, "Port for Alert")
	cmd.Flags().BoolVar(&ctl.StandAlone, "stand-alone", ctl.StandAlone, "Enable Stand Alone mode")
	cmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation for the Alert")
	cmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation for the Cfssl")
	cmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of the Alert")

	assert.Equal(cmd.Flags(), actualCmd.Flags())
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
		changedSpec *alertv1.AlertSpec
	}{
		// case
		{flagName: "alert-registry",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Registry: "changed",
			},
			changedSpec: &alertv1.AlertSpec{Registry: "changed"},
		},
		// case
		{flagName: "image-path",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				ImagePath: "changed",
			},
			changedSpec: &alertv1.AlertSpec{ImagePath: "changed"},
		},
		// case
		{flagName: "alert-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				AlertImageName: "changed",
			},
			changedSpec: &alertv1.AlertSpec{AlertImageName: "changed"},
		},
		// case
		{flagName: "alert-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				AlertImageVersion: "changed",
			},
			changedSpec: &alertv1.AlertSpec{AlertImageVersion: "changed"},
		},
		// case
		{flagName: "cfssl-image-name",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				CfsslImageName: "changed",
			},
			changedSpec: &alertv1.AlertSpec{CfsslImageName: "changed"},
		},
		// case
		{flagName: "cfssl-image-version",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				CfsslImageVersion: "changed",
			},
			changedSpec: &alertv1.AlertSpec{CfsslImageVersion: "changed"},
		},
		// case
		{flagName: "blackduck-host",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				BlackduckHost: "changed",
			},
			changedSpec: &alertv1.AlertSpec{BlackduckHost: "changed"},
		},
		// case
		{flagName: "blackduck-user",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				BlackduckUser: "changed",
			},
			changedSpec: &alertv1.AlertSpec{BlackduckUser: "changed"},
		},
		// case
		{flagName: "blackduck-port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				BlackduckPort: 10,
			},
			changedSpec: &alertv1.AlertSpec{BlackduckPort: crddefaults.IntToPtr(10)},
		},
		// case
		{flagName: "port",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				Port: 10,
			},
			changedSpec: &alertv1.AlertSpec{Port: crddefaults.IntToPtr(10)},
		},
		// case
		{flagName: "stand-alone",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				StandAlone: true,
			},
			changedSpec: &alertv1.AlertSpec{StandAlone: crddefaults.BoolToPtr(true)},
		},
		// case
		{flagName: "alert-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				AlertMemory: "changed",
			},
			changedSpec: &alertv1.AlertSpec{AlertMemory: "changed"},
		},
		// case
		{flagName: "cfssl-memory",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				CfsslMemory: "changed",
			},
			changedSpec: &alertv1.AlertSpec{CfsslMemory: "changed"},
		},
		// case
		{flagName: "alert-desired-state",
			initialCtl: NewAlertCtl(),
			changedCtl: &Ctl{Spec: &alertv1.AlertSpec{},
				DesiredState: "changed",
			},
			changedSpec: &alertv1.AlertSpec{DesiredState: "changed"},
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
