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

// CRSpecBuilderFromCobraFlags uses Cobra commands, Cobra flags and other
// values to create an Alert CR's Spec.
//
// The fields in the CRSpecBuilderFromCobraFlags represent places where the values of the Cobra flags are stored.
//
// Usage: Use CRSpecBuilderFromCobraFlags to add flags to your Cobra Command for making an Alert Spec.
// When flags are used the correspoding value in this struct will by set. You can then
// generate the spec by telling CRSpecBuilderFromCobraFlags what flags were changed.
type CRSpecBuilderFromCobraFlags struct {
	alertSpec            *alertapi.AlertSpec
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

// NewCRSpecBuilderFromCobraFlags creates a new CRSpecBuilderFromCobraFlags type
func NewCRSpecBuilderFromCobraFlags() *CRSpecBuilderFromCobraFlags {
	return &CRSpecBuilderFromCobraFlags{
		alertSpec: &alertapi.AlertSpec{},
	}
}

// GetCRSpec returns a pointer to the AlertSpec as an interface{}
func (ctl *CRSpecBuilderFromCobraFlags) GetCRSpec() interface{} {
	return *ctl.alertSpec
}

// SetCRSpec sets the alertSpec in the struct
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpec(spec interface{}) error {
	convertedSpec, ok := spec.(alertapi.AlertSpec)
	if !ok {
		return fmt.Errorf("error setting Alert spec")
	}
	ctl.alertSpec = &convertedSpec
	return nil
}

// Constants for predefined specs
const (
	EmptySpec   string = "empty"
	DefaultSpec string = "default"
)

// SetPredefinedCRSpec sets the alertSpec to a predefined spec
func (ctl *CRSpecBuilderFromCobraFlags) SetPredefinedCRSpec(specType string) error {
	switch specType {
	case EmptySpec:
		ctl.alertSpec = &alertapi.AlertSpec{}
	case DefaultSpec:
		ctl.alertSpec = util.GetAlertDefault()
		ctl.alertSpec.PersistentStorage = true
		ctl.alertSpec.StandAlone = util.BoolToPtr(true)
	default:
		return fmt.Errorf("Alert spec type '%s' is not valid", specType)
	}
	return nil
}

// AddCRSpecFlagsToCommand adds flags to a Cobra Command that are need for Alert's Spec.
// The flags map to fields in the CRSpecBuilderFromCobraFlags struct.
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *CRSpecBuilderFromCobraFlags) AddCRSpecFlagsToCommand(cmd *cobra.Command, master bool) {
	cmd.Flags().StringVar(&ctl.Version, "version", ctl.Version, "Version of Alert")
	cmd.Flags().StringVar(&ctl.AlertImage, "alert-image", ctl.AlertImage, "URL of Alert's Image")
	cmd.Flags().StringVar(&ctl.CfsslImage, "cfssl-image", ctl.CfsslImage, "URL of CFSSL's Image")
	cmd.Flags().StringVar(&ctl.StandAlone, "standalone", ctl.StandAlone, "If true, Alert runs in standalone mode (true|false)")
	if master {
		cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", util.NONE, "Service type to expose Alert's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	} else {
		cmd.Flags().StringVar(&ctl.ExposeService, "expose-ui", ctl.ExposeService, "Service type to expose Alert's user interface (NODEPORT|LOADBALANCER|OPENSHIFT|NONE)")
	}
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
}

// CheckValuesFromFlags returns an error if a value stored in the struct will not be able to be
// used in the AlertSpec
func (ctl *CRSpecBuilderFromCobraFlags) CheckValuesFromFlags(flagset *pflag.FlagSet) error {
	if FlagWasSet(flagset, "encryption-password") {
		encryptPassLength := len(ctl.EncryptionPassword)
		if encryptPassLength > 0 && encryptPassLength < 16 {
			return fmt.Errorf("flag EncryptionPassword is %d characters. Must be 16 or more characters", encryptPassLength)
		}
	}
	if FlagWasSet(flagset, "encryption-global-salt") {
		globalSaltLength := len(ctl.EncryptionGlobalSalt)
		if globalSaltLength > 0 && globalSaltLength < 16 {
			return fmt.Errorf("flag EncryptionGlobalSalt is %d characters. Must be 16 or more characters", globalSaltLength)
		}
	}
	if FlagWasSet(flagset, "expose-ui") {
		isValid := util.IsExposeServiceValid(ctl.ExposeService)
		if !isValid {
			return fmt.Errorf("expose ui must be '%s', '%s', '%s' or '%s'", util.NODEPORT, util.LOADBALANCER, util.OPENSHIFT, util.NONE)
		}
	}
	return nil
}

// FlagWasSet returns true if a flag was changed and it exists, otherwise it returns false
func FlagWasSet(flagset *pflag.FlagSet, flagName string) bool {
	if flagset.Lookup(flagName) != nil && flagset.Lookup(flagName).Changed {
		return true
	}
	return false
}

// GenerateCRSpecFromFlags checks if a flag was changed and updates the alertSpec with the value that's stored
// in the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) GenerateCRSpecFromFlags(flagset *pflag.FlagSet) (interface{}, error) {
	err := ctl.CheckValuesFromFlags(flagset)
	if err != nil {
		return nil, err
	}
	flagset.VisitAll(ctl.SetCRSpecFieldByFlag)
	return *ctl.alertSpec, nil
}

// SetCRSpecFieldByFlag updates a field in the alertSpec if the flag was set by the user. It gets the
// value from the corresponding struct field
func (ctl *CRSpecBuilderFromCobraFlags) SetCRSpecFieldByFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("flag '%s': CHANGED", f.Name)
		switch f.Name {
		case "version":
			ctl.alertSpec.Version = ctl.Version
		case "alert-image":
			ctl.alertSpec.AlertImage = ctl.AlertImage
		case "cfssl-image":
			ctl.alertSpec.CfsslImage = ctl.CfsslImage
		case "standalone":
			standAloneVal := strings.ToUpper(ctl.StandAlone) == "TRUE"
			ctl.alertSpec.StandAlone = &standAloneVal
		case "expose-ui":
			ctl.alertSpec.ExposeService = ctl.ExposeService
		case "port":
			ctl.alertSpec.Port = &ctl.Port
		case "encryption-password":
			ctl.alertSpec.EncryptionPassword = ctl.EncryptionPassword
		case "encryption-global-salt":
			ctl.alertSpec.EncryptionGlobalSalt = ctl.EncryptionGlobalSalt
		case "persistent-storage":
			ctl.alertSpec.PersistentStorage = strings.ToUpper(ctl.PersistentStorage) == "TRUE"
		case "pvc-name":
			ctl.alertSpec.PVCName = ctl.PVCName
		case "pvc-storage-class":
			ctl.alertSpec.PVCStorageClass = ctl.PVCStorageClass
		case "pvc-size":
			ctl.alertSpec.PVCSize = ctl.PVCSize
		case "alert-memory":
			ctl.alertSpec.AlertMemory = ctl.AlertMemory
		case "cfssl-memory":
			ctl.alertSpec.CfsslMemory = ctl.CfsslMemory
		case "environs":
			ctl.alertSpec.Environs = ctl.Environs
		case "alert-desired-state":
			ctl.alertSpec.DesiredState = ctl.DesiredState
		default:
			log.Debugf("flag '%s': NOT FOUND", f.Name)
		}
	} else {
		log.Debugf("flag '%s': UNCHANGED", f.Name)
	}
}
