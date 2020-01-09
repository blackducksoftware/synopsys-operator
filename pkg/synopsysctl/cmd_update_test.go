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
	"sort"
	"testing"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	blackduckctl "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUpdatingEnvironValues(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		environsFlagValue       string
		environsInitiallyInSpec []string
		expectedEnvirons        []string
	}{
		{ // case : remove single environ
			environsFlagValue:       "a:",
			environsInitiallyInSpec: []string{"a:a"},
			expectedEnvirons:        []string{},
		},
		{ // case : remove environ from a list
			environsFlagValue:       "a:a,b:",
			environsInitiallyInSpec: []string{"b:"},
			expectedEnvirons:        []string{"a:a"},
		},
		{ // case : remove multiple environs
			environsFlagValue:       "a:,b:,c:",
			environsInitiallyInSpec: []string{"a:a", "b:b", "c:c"},
			expectedEnvirons:        []string{},
		},
		{ // case : remove an environ and change other's values
			environsFlagValue:       "a:,b:c,c:b",
			environsInitiallyInSpec: []string{"a:a", "b:b", "c:c"},
			expectedEnvirons:        []string{"b:c", "c:b"},
		},
	}

	for _, test := range tests {
		CRSpecBuilder := blackduckctl.NewCRSpecBuilderFromCobraFlags()

		// Get a flagset
		cmd := &cobra.Command{}
		CRSpecBuilder.AddCRSpecFlagsToCommand(cmd, true)
		flagset := cmd.Flags()

		// Set the flagset values based on the test
		flagset.Set("environs", test.environsFlagValue)
		bd := blackduckapi.Blackduck{}
		bd.Spec.Environs = test.environsInitiallyInSpec
		updatedBd, err := updateBlackDuckSpec(&bd, flagset)
		if err != nil {
			t.Errorf("%+v", err)
		}

		// Verify the test passed
		sort.Strings(updatedBd.Spec.Environs)
		sort.Strings(test.expectedEnvirons)
		assert.Equal(updatedBd.Spec.Environs, test.expectedEnvirons)
	}
}
