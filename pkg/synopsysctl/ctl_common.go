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

package synopsysctl

import (
	"fmt"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ResourceCtl interface is used in place of any ctl
type ResourceCtl interface {
	CheckSpecFlags() error                  // returns an error if a flag format is invalid
	GetSpec() (interface{}, error)          // returns spec for the resource
	SetSpec(interface{}) error              // sets the spec
	SwitchSpec(string) error                // change the spec for the resource
	AddSpecFlags(*cobra.Command, bool)      // add flags for the resource spec
	NumSpecFlagsChanged(*pflag.FlagSet) int // returns number of spec flags that were changed in the flagset
	SetChangedFlags(*pflag.FlagSet)         // calls setFlag on each flag in flagset
	SetFlag(*pflag.Flag)                    // updates the spec value for the flag
	SpecIsValid() (bool, error)             // verifies the spec has necessary fields to deploy
	CanUpdate() (bool, error)               // checks if a user has permission to modify based on the spec
}

// CommonCtl is an embedded type for Ctl structs to provide common functionality
type CommonCtl struct {
	Ctl interface{}
}

// GetSpec returns the Spec for the resource
func (c *CommonCtl) GetSpec() (interface{}, error) {
	if ctl, ok := c.Ctl.(*AlertCtl); ok {
		return *ctl.Spec, nil
	}
	if ctl, ok := c.Ctl.(*BlackDuckCtl); ok {
		return *ctl.Spec, nil
	}
	if ctl, ok := c.Ctl.(*OpsSightCtl); ok {
		return *ctl.Spec, nil
	}
	return nil, fmt.Errorf("unable to get resource spec")
}

// SetSpec sets the Spec for the resource
func (c *CommonCtl) SetSpec(spec interface{}) error {
	if ctl, ok := c.Ctl.(*AlertCtl); ok {
		convertedSpec, ok := spec.(alertapi.AlertSpec)
		if !ok {
			return fmt.Errorf("'%+v' is not a valid spec", spec)
		}
		ctl.Spec = &convertedSpec
		return nil
	}
	if ctl, ok := c.Ctl.(*BlackDuckCtl); ok {
		convertedSpec, ok := spec.(blackduckapi.BlackduckSpec)
		if !ok {
			return fmt.Errorf("'%+v' is not a valid spec", spec)
		}
		ctl.Spec = &convertedSpec
		return nil
	}
	if ctl, ok := c.Ctl.(*OpsSightCtl); ok {
		convertedSpec, ok := spec.(opssightapi.OpsSightSpec)
		if !ok {
			return fmt.Errorf("'%+v' is not a valid spec", spec)
		}
		ctl.Spec = &convertedSpec
		return nil
	}
	return fmt.Errorf("Error setting Alert Spec")
}

// NumSpecFlagsChanged returns the number of spec flags that were set
func (c *CommonCtl) NumSpecFlagsChanged(flagset *pflag.FlagSet) int {
	if ctl, ok := c.Ctl.(*AlertCtl); ok {
		return GetNumSpecFlagsChanged(ctl, flagset)
	}
	if ctl, ok := c.Ctl.(*BlackDuckCtl); ok {
		return GetNumSpecFlagsChanged(ctl, flagset)
	}
	if ctl, ok := c.Ctl.(*OpsSightCtl); ok {
		return GetNumSpecFlagsChanged(ctl, flagset)
	}
	log.Errorf("Error getting Num Spec Flags")
	return 0
}

var nSpecFlagCmd cobra.Command // command for getting flags for this spec
var specFlags *pflag.FlagSet   // set of flags for this spec
var numSpecFlagsChanged int    // number of flags that were set for the spec

// GetNumSpecFlagsChanged returns the number of spec flags that were set
func GetNumSpecFlagsChanged(ctl ResourceCtl, flagset *pflag.FlagSet) int {
	// Initialize variables for comparing flags
	nSpecFlagCmd = cobra.Command{}
	ctl.AddSpecFlags(&nSpecFlagCmd, true)
	specFlags = nSpecFlagCmd.Flags()
	numSpecFlagsChanged = 0
	// Count changed flags
	flagset.VisitAll(incrementNumSpecFlagsChanged)
	return numSpecFlagsChanged
}

// incrementNumSpecFlagsChanged increments numSpecFlagsChanged if the flag relates to
// the resource's spec and it has changed
func incrementNumSpecFlagsChanged(flag *pflag.Flag) {
	isFlagInSpec := specFlags.Lookup(flag.Name) != nil // check if the flag is in the Spec's Flags
	if isFlagInSpec {
		if flag.Changed {
			numSpecFlagsChanged = numSpecFlagsChanged + 1
		}
	}
}
