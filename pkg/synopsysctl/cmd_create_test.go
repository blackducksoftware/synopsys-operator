/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package synopsysctl

import (
	"testing"
	//blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"reflect"

	// "github.com/blackducksoftware/synopsys-operator/pkg/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/cobra"
)

func TestAddPVCValuesToBlackDuckSpec(t *testing.T) {

	testCases := []struct {
		name             string
		expectedPVCMap   map[string]string
		blackDuckVersion string
	}{
		{
			name: "BlackDuck v2 with solr",
			expectedPVCMap: map[string]string{
				"blackduck-authentication":   "2Gi",
				"blackduck-cfssl":            "2Gi",
				"blackduck-registration":     "2Gi",
				"blackduck-solr":             "2Gi",
				"blackduck-webapp":           "2Gi",
				"blackduck-logstash":         "20Gi",
				"blackduck-zookeeper":        "4Gi",
				"blackduck-uploadcache-data": "100Gi",
				"blackduck-postgres":         "150Gi", // Needed because the Black Duck Default below has postgres enabled
			},
			blackDuckVersion: "2019.4.0",
		},
		{
			name: "BlackDuck v2 no solr",
			expectedPVCMap: map[string]string{
				"blackduck-authentication":   "2Gi",
				"blackduck-cfssl":            "2Gi",
				"blackduck-registration":     "2Gi",
				"blackduck-webapp":           "2Gi",
				"blackduck-logstash":         "20Gi",
				"blackduck-zookeeper":        "4Gi",
				"blackduck-uploadcache-data": "100Gi",
				"blackduck-postgres":         "150Gi",
			},
			blackDuckVersion: "2019.6.0",
		},
		{
			name: "BlackDuck v1 with solr",
			expectedPVCMap: map[string]string{
				"blackduck-authentication": "2Gi",
				"blackduck-cfssl":          "2Gi",
				"blackduck-registration":   "2Gi",
				"blackduck-solr":           "2Gi",
				"blackduck-webapp":         "2Gi",
				"blackduck-logstash":       "20Gi",
				"blackduck-zookeeper":      "4Gi",
				"blackduck-postgres":       "150Gi",
			},
			blackDuckVersion: "2018.12.0",
		},
		{
			name: "BlackDuck latest with no solr",
			expectedPVCMap: map[string]string{
				"blackduck-authentication":   "2Gi",
				"blackduck-cfssl":            "2Gi",
				"blackduck-registration":     "2Gi",
				"blackduck-webapp":           "2Gi",
				"blackduck-logstash":         "20Gi",
				"blackduck-zookeeper":        "4Gi",
				"blackduck-uploadcache-data": "100Gi",
				"blackduck-postgres":         "150Gi",
			},
			blackDuckVersion: "2019.8.1",
		},
	}

	cmd := &cobra.Command{}
	createBlackDuckCobraHelper.AddCRSpecFlagsToCommand(cmd, true)
	blackDuckSpec := util.GetBlackDuckDefaultPersistentStorageLatest()
	for _, test := range testCases {
		// Set the Version of BlackDuck
		blackDuckSpec.Version = test.blackDuckVersion
		createBlackDuckCobraHelper.SetCRSpec(*blackDuckSpec)
		cmd.Flags().Set("version", test.blackDuckVersion)

		// Execute the function being tested
		finalBlackDuckSpec, err := addPVCValuesToBlackDuckSpec(cmd, "", "", blackDuckSpec)
		if err != nil {
			t.Errorf("failed to add PVC values: %s", err)
		}

		// Check PVC values and convert to a map for comparing with expectedPVCMap
		observedPVCMap := map[string]string{}
		observedPVCs := finalBlackDuckSpec.PVC
		for _, pvc := range observedPVCs {
			if val, ok := test.expectedPVCMap[pvc.Name]; ok {
				if val != pvc.Size {
					t.Errorf("Case %s: invalid PVC size; expected %+v - got: %+v", test.name, val, pvc.Size)
				}
			} else {
				t.Errorf("Case %s: unexpected PVC Name: %+v", test.name, pvc.Name)
			}
			observedPVCMap[pvc.Name] = pvc.Size
		}

		// Compare results to the expectedPVCMap
		eq := reflect.DeepEqual(test.expectedPVCMap, observedPVCMap)
		if !eq {
			t.Errorf("Case %s: got = %v, want %v", test.name, observedPVCMap, test.expectedPVCMap)
		}
	}
}
