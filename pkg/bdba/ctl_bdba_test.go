/*
Copyright (C) 2020 Synopsys, Inc.

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

package bdba

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewHelmValuesFromCobraFlags(t *testing.T) {
	assert := assert.New(t)
	bdbaCobraHelper := NewHelmValuesFromCobraFlags()
	assert.Equal(&HelmValuesFromCobraFlags{
		args:     map[string]interface{}{},
		flagTree: FlagTree{},
	}, bdbaCobraHelper)
}

func TestGetArgs(t *testing.T) {
	assert := assert.New(t)
	bdbaCobraHelper := NewHelmValuesFromCobraFlags()
	assert.Equal(map[string]interface{}{}, bdbaCobraHelper.GetArgs())
}

func TestGenerateHelmFlagsFromCobraFlags(t *testing.T) {
	assert := assert.New(t)

	bdbaCobraHelper := NewHelmValuesFromCobraFlags()
	cmd := &cobra.Command{}
	bdbaCobraHelper.AddCobraFlagsToCommand(cmd, true)
	flagset := cmd.Flags()
	// Set flags here...

	bdbaCobraHelper.GenerateHelmFlagsFromCobraFlags(flagset)

	expectedArgs := map[string]interface{}{}

	assert.Equal(expectedArgs, bdbaCobraHelper.GetArgs())

}

func TestSetCRSpecFieldByFlag(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		flagName    string
		initialCtl  *HelmValuesFromCobraFlags
		changedCtl  *HelmValuesFromCobraFlags
		changedArgs map[string]interface{}
	}{
		// case
		{
			flagName:    "version",
			changedCtl:  &HelmValuesFromCobraFlags{},
			changedArgs: map[string]interface{}{},
		},
		// case
		{
			flagName:    "hostname",
			changedCtl:  &HelmValuesFromCobraFlags{},
			changedArgs: map[string]interface{}{},
		},
		// case
		{
			flagName:    "rabbitmq-domain",
			changedCtl:  &HelmValuesFromCobraFlags{},
			changedArgs: map[string]interface{}{},
		},
		// case
		{
			flagName:    "psql-storage-class",
			changedCtl:  &HelmValuesFromCobraFlags{},
			changedArgs: map[string]interface{}{},
		},
		// TODO: More test cases and fix the ones above ^
	}

	// get the flagset
	cmd := &cobra.Command{}
	bdbaCobraHelper := NewHelmValuesFromCobraFlags()
	bdbaCobraHelper.AddCobraFlagsToCommand(cmd, true)
	flagset := cmd.Flags()

	for _, test := range tests {
		// check the Flag exists
		foundFlag := flagset.Lookup(test.flagName)
		if foundFlag == nil {
			t.Errorf("flag %s is not in the spec", test.flagName)
		}
		// test setting the flag
		f := &pflag.Flag{Changed: true, Name: test.flagName}
		bdbaCobraHelper = NewHelmValuesFromCobraFlags() // clear the bdbaCobraHelper
		bdbaCobraHelper.AddHelmValueByCobraFlag(f)
		assert.Equal(test.changedArgs, bdbaCobraHelper.GetArgs())
	}

	// case: nothing set if flag doesn't exist
	bdbaCobraHelper = NewHelmValuesFromCobraFlags()
	f := &pflag.Flag{Changed: true, Name: "bad-flag"}
	bdbaCobraHelper.AddHelmValueByCobraFlag(f)
	assert.Equal(map[string]interface{}{}, bdbaCobraHelper.GetArgs())

}
