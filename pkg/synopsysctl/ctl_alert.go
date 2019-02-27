// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"fmt"

	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AlertCtl type provides functionality for an Alert
// for the Synopsysctl tool
type AlertCtl struct {
	Spec              *alertv1.AlertSpec
	Registry          string
	ImagePath         string
	AlertImageName    string
	AlertImageVersion string
	CfsslImageName    string
	CfsslImageVersion string
	BlackduckHost     string
	BlackduckUser     string
	BlackduckPort     int
	Port              int
	StandAlone        bool
	AlertMemory       string
	CfsslMemory       string
	State             string
}

// NewAlertCtl creates a new AlertCtl struct
func NewAlertCtl() *AlertCtl {
	return &AlertCtl{
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
		State:             "",
	}
}

// GetSpec returns the Spec for the resource
func (ctl *AlertCtl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *AlertCtl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(alertv1.AlertSpec)
	if !ok {
		return fmt.Errorf("Error setting Alert Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *AlertCtl) CheckSpecFlags() error {
	return nil
}

// SwitchSpec switches the Alert's Spec to a different predefined spec
func (ctl *AlertCtl) SwitchSpec(createAlertSpecType string) error {
	switch createAlertSpecType {
	case "empty":
		ctl.Spec = &alertv1.AlertSpec{}
	case "spec1":
		ctl.Spec = crddefaults.GetAlertDefaultValue()
	case "spec2":
		ctl.Spec = crddefaults.GetAlertDefaultValue2()
	default:
		return fmt.Errorf("Alert Spec Type %s does not match: empty, spec1, spec2", createAlertSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for the Alert's Spec to the command
func (ctl *AlertCtl) AddSpecFlags(cmd *cobra.Command) {
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
	cmd.Flags().StringVar(&ctl.State, "alert-state", ctl.State, "State of the Alert")
}

// SetFlags sets the Alert's Spec if a flag was changed
func (ctl *AlertCtl) SetFlags(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s:   CHANGED\n", f.Name)
		switch f.Name {
		case "alert-registry":
			ctl.Spec.Registry = ctl.Registry
		case "image-path":
			ctl.Spec.ImagePath = ctl.ImagePath
		case "alert-image-name":
			ctl.Spec.AlertImageName = ctl.AlertImageName
		case "alert-image-version":
			ctl.Spec.AlertImageVersion = ctl.AlertImageVersion
		case "cfssl-image-name":
			ctl.Spec.CfsslImageName = ctl.CfsslImageName
		case "cfssl-image-version":
			ctl.Spec.CfsslImageVersion = ctl.CfsslImageVersion
		case "blackduck-host":
			ctl.Spec.BlackduckHost = ctl.BlackduckHost
		case "blackduck-user":
			ctl.Spec.BlackduckUser = ctl.BlackduckUser
		case "blackduck-port":
			fmt.Printf("Shouldn't be here\n")
			ctl.Spec.BlackduckPort = &ctl.BlackduckPort
		case "port":
			ctl.Spec.Port = &ctl.Port
		case "stand-alone":
			ctl.Spec.StandAlone = &ctl.StandAlone
		case "alert-memory":
			ctl.Spec.AlertMemory = ctl.AlertMemory
		case "cfssl-memory":
			ctl.Spec.CfsslMemory = ctl.CfsslMemory
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}
