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

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Ctl type provides functionality for an Alert
// for the Synopsysctl tool
type Ctl struct {
	Spec                 *alertapi.AlertSpec
	Version              string
	AlertImage           string
	CfsslImage           string
	StandAlone           bool
	ExposeService        string
	Port                 int
	EncryptionPassword   string
	EncryptionGlobalSalt string
	Environs             []string
	PersistentStorage    bool
	PVCName              string
	PVCStorageClass      string
	PVCSize              string
	AlertMemory          string
	CfsslMemory          string
	DesiredState         string
}

// NewAlertCtl creates a new AlertCtl struct
func NewAlertCtl() *Ctl {
	return &Ctl{
		Spec:                 &alertapi.AlertSpec{},
		Version:              "",
		AlertImage:           "",
		CfsslImage:           "",
		StandAlone:           false,
		ExposeService:        "",
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
	}
}

// GetSpec returns the Spec for the resource
func (ctl *Ctl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *Ctl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(alertapi.AlertSpec)
	if !ok {
		return fmt.Errorf("Error setting Alert Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags() error {
	encryptPassLength := len(ctl.EncryptionPassword)
	if encryptPassLength > 0 && encryptPassLength < 16 {
		return fmt.Errorf("flag EncryptionPassword is %d characters. Must be 16 or more characters", encryptPassLength)
	}
	globalSaltLength := len(ctl.EncryptionGlobalSalt)
	if globalSaltLength > 0 && globalSaltLength < 16 {
		return fmt.Errorf("flag EncryptionGlobalSalt is %d characters. Must be 16 or more characters", globalSaltLength)
	}
	return nil
}

// Constants for Default Specs
const (
	EmptySpec    string = "empty"
	TemplateSpec string = "template"
	DefaultSpec  string = "default"
)

// SwitchSpec switches the Alert's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(specType string) error {
	switch specType {
	case EmptySpec:
		ctl.Spec = &alertapi.AlertSpec{}
	case TemplateSpec:
		ctl.Spec = crddefaults.GetAlertTemplate()
	case DefaultSpec:
		ctl.Spec = crddefaults.GetAlertDefault()
	default:
		return fmt.Errorf("Alert Spec Type %s is not valid", specType)
	}
	return nil
}

// AddSpecFlags adds flags for the Alert's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of the Alert")
	cmd.Flags().StringVar(&ctl.AlertImage, "alert-image", ctl.AlertImage, "Url of the Alert Image")
	cmd.Flags().StringVar(&ctl.CfsslImage, "cfssl-image", ctl.CfsslImage, "Url of Cfssl Image")
	cmd.Flags().BoolVar(&ctl.StandAlone, "stand-alone", ctl.StandAlone, "Enable Stand Alone mode")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-service", ctl.ExposeService, "Type of Service to Expose")
	cmd.Flags().IntVar(&ctl.Port, "port", ctl.Port, "Port for Alert")
	cmd.Flags().StringVar(&ctl.EncryptionPassword, "encryption-password", ctl.EncryptionPassword, "Encryption Password for the Alert")
	cmd.Flags().StringVar(&ctl.EncryptionGlobalSalt, "encryption-global-salt", ctl.EncryptionGlobalSalt, "Encryption Global Salt for the Alert")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "Environment variables for the Alert")
	cmd.Flags().BoolVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "Enable persistent storage")
	cmd.Flags().StringVar(&ctl.PVCName, "pvc-name", ctl.PVCName, "Name for the PVC")
	cmd.Flags().StringVar(&ctl.PVCStorageClass, "pvc-storage-class", ctl.PVCStorageClass, "StorageClass for the PVC")
	cmd.Flags().StringVar(&ctl.PVCSize, "pvc-size", ctl.PVCSize, "Memory allocation for the PVC")
	cmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation for the Alert")
	cmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation for the Cfssl")
	cmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of the Alert")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	flagset.VisitAll(ctl.SetFlag)
}

// SetFlag sets an Alert's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "version":
			ctl.Spec.Version = ctl.Version
		case "alert-image":
			ctl.Spec.AlertImage = ctl.AlertImage
		case "cfssl-image":
			ctl.Spec.CfsslImage = ctl.CfsslImage
		case "stand-alone":
			ctl.Spec.StandAlone = &ctl.StandAlone
		case "expose-service":
			ctl.Spec.ExposeService = ctl.ExposeService
		case "port":
			ctl.Spec.Port = &ctl.Port
		case "encryption-password":
			ctl.Spec.EncryptionPassword = ctl.EncryptionPassword
		case "encryption-global-salt":
			ctl.Spec.EncryptionGlobalSalt = ctl.EncryptionGlobalSalt
		case "persistent-storage":
			ctl.Spec.PersistentStorage = ctl.PersistentStorage
		case "pvc-name":
			ctl.Spec.PVCName = ctl.PVCName
		case "pvc-storage-class":
			ctl.Spec.PVCStorageClass = ctl.PVCStorageClass
		case "pvc-size":
			ctl.Spec.PVCSize = ctl.PVCSize
		case "alert-memory":
			ctl.Spec.AlertMemory = ctl.AlertMemory
		case "cfssl-memory":
			ctl.Spec.CfsslMemory = ctl.CfsslMemory
		case "environs":
			ctl.Spec.Environs = ctl.Environs
		case "alert-desired-state":
			ctl.Spec.DesiredState = ctl.DesiredState
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}

// SpecIsValid verifies the spec has necessary fields to deploy
func (ctl *Ctl) SpecIsValid() (bool, error) {
	return true, nil
}

// CanUpdate checks if a user has permission to modify based on the spec
func (ctl *Ctl) CanUpdate() (bool, error) {
	return true, nil
}
