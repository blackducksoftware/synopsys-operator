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
	"strings"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
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
	StandAlone           string
	ExposeService        string
	Port                 int32
	EncryptionPassword   string
	EncryptionGlobalSalt string
	Environs             []string
	PersistentStorage    string
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
		Spec: &alertapi.AlertSpec{},
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
		return fmt.Errorf("error setting Alert spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags(flagset *pflag.FlagSet) error {
	encryptPassLength := len(ctl.EncryptionPassword)
	if encryptPassLength > 0 && encryptPassLength < 16 {
		return fmt.Errorf("flag EncryptionPassword is %d characters. Must be 16 or more characters", encryptPassLength)
	}
	globalSaltLength := len(ctl.EncryptionGlobalSalt)
	if globalSaltLength > 0 && globalSaltLength < 16 {
		return fmt.Errorf("flag EncryptionGlobalSalt is %d characters. Must be 16 or more characters", globalSaltLength)
	}
	isValid := util.IsExposeServiceValid(ctl.ExposeService)
	if !isValid {
		return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
	}
	return nil
}

// Constants for Default Specs
const (
	EmptySpec   string = "empty"
	DefaultSpec string = "default"
)

// SwitchSpec switches Alert's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(specType string) error {
	switch specType {
	case EmptySpec:
		ctl.Spec = &alertapi.AlertSpec{}
	case DefaultSpec:
		ctl.Spec = util.GetAlertDefault()
		ctl.Spec.PersistentStorage = true
		ctl.Spec.StandAlone = util.BoolToPtr(true)
	default:
		return fmt.Errorf("Alert spec type '%s' is not valid", specType)
	}
	return nil
}

// AddSpecFlags adds flags for Alert's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Alert")
	cmd.Flags().StringVar(&ctl.AlertImage, "alert-image", ctl.AlertImage, "URL of Alert's Image")
	cmd.Flags().StringVar(&ctl.CfsslImage, "cfssl-image", ctl.CfsslImage, "URL of CFSSL's Image")
	cmd.Flags().StringVar(&ctl.StandAlone, "standalone", ctl.StandAlone, "If true, Alert runs in standalone mode [true|false]")
	cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type to expose Alert's user interface [NODEPORT|LOADBALANCER|OPENSHIFT|NONE]")
	cmd.Flags().Int32Var(&ctl.Port, "port", ctl.Port, "Port of Alert")
	cmd.Flags().StringVar(&ctl.EncryptionPassword, "encryption-password", ctl.EncryptionPassword, "Encryption Password for Alert")
	cmd.Flags().StringVar(&ctl.EncryptionGlobalSalt, "encryption-global-salt", ctl.EncryptionGlobalSalt, "Encryption Global Salt for Alert")
	cmd.Flags().StringSliceVar(&ctl.Environs, "environs", ctl.Environs, "Environment variables of Alert")
	cmd.Flags().StringVar(&ctl.PersistentStorage, "persistent-storage", ctl.PersistentStorage, "If true, Alert has persistent storage [true|false]")
	cmd.Flags().StringVar(&ctl.PVCName, "pvc-name", ctl.PVCName, "Name of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCStorageClass, "pvc-storage-class", ctl.PVCStorageClass, "Storage class for the persistent volume claim")
	cmd.Flags().StringVar(&ctl.PVCSize, "pvc-size", ctl.PVCSize, "Memory allocation of the persistent volume claim")
	cmd.Flags().StringVar(&ctl.AlertMemory, "alert-memory", ctl.AlertMemory, "Memory allocation of Alert")
	cmd.Flags().StringVar(&ctl.CfsslMemory, "cfssl-memory", ctl.CfsslMemory, "Memory allocation of CFSSL")
	cmd.Flags().StringVar(&ctl.DesiredState, "alert-desired-state", ctl.DesiredState, "State of Alert")

	// TODO: Remove this flag in next release
	cmd.Flags().MarkDeprecated("alert-desired-state", "alert-desired-state flag is deprecated and will be removed by the next release")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	flagset.VisitAll(ctl.SetFlag)
}

// SetFlag sets an Alert's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.Spec.Version = ctl.Version
		case "alert-image":
			ctl.Spec.AlertImage = ctl.AlertImage
		case "cfssl-image":
			ctl.Spec.CfsslImage = ctl.CfsslImage
		case "standalone":
			standAloneVal := strings.ToUpper(ctl.StandAlone) == "TRUE"
			ctl.Spec.StandAlone = &standAloneVal
		case "expose-ui":
			ctl.Spec.ExposeService = ctl.ExposeService
		case "port":
			ctl.Spec.Port = &ctl.Port
		case "encryption-password":
			ctl.Spec.EncryptionPassword = ctl.EncryptionPassword
		case "encryption-global-salt":
			ctl.Spec.EncryptionGlobalSalt = ctl.EncryptionGlobalSalt
		case "persistent-storage":
			ctl.Spec.PersistentStorage = strings.ToUpper(ctl.PersistentStorage) == "TRUE"
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
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
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
