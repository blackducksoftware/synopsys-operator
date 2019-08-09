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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CRSpecBuilderFromCobraFlagsInterface requires Cobra commands, Cobra flags and other
// values to create a Spec for a Custom Resource.
type CRSpecBuilderFromCobraFlagsInterface interface {
	GetCRSpec() interface{}                                      // returns the spec for the resource
	SetCRSpec(interface{}) error                                 // sets the spec
	SetPredefinedCRSpec(string) error                            // sets the spec to a predefined spec
	AddCRSpecFlagsToCommand(*cobra.Command, bool)                // Adds flags for the resource's spec to a Cobra command
	CheckValuesFromFlags(*pflag.FlagSet) error                   // returns an error if a value for the spec is invalid
	GenerateCRSpecFromFlags(*pflag.FlagSet) (interface{}, error) // calls SetSpecFieldByFlag on each flag in flagset
	SetCRSpecFieldByFlag(*pflag.Flag)                            // updates the resource's spec with the value from a flag
}
