/*
 * Copyright (C) 2020 Synopsys, Inc.
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

package polarisreporting

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFlag(t *testing.T) {
	type test struct {
		testDesc   string
		startMap   map[string]interface{}
		keyString  []string
		finalValue interface{}
		finalMap   map[string]interface{}
	}
	tests := []test{
		{
			testDesc:   "empty key string and empty startMap",
			startMap:   map[string]interface{}{},
			keyString:  []string{},
			finalValue: "finalValue",
			finalMap:   map[string]interface{}{},
		},
		{
			testDesc: "empty key string and non-empty startMap",
			startMap: map[string]interface{}{
				"first": "otherValue",
			},
			keyString:  []string{},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first": "otherValue",
			},
		},
		{
			testDesc:   "place value into empty map with one key",
			startMap:   map[string]interface{}{},
			keyString:  []string{"first"},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first": "finalValue",
			},
		},
		{
			testDesc: "override value in map",
			startMap: map[string]interface{}{
				"first": "oldValue",
			},
			keyString:  []string{"first"},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first": "finalValue",
			},
		},
		{
			testDesc:   "place value into empty map with multiple keys",
			startMap:   map[string]interface{}{},
			keyString:  []string{"first", "second"},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first": map[string]interface{}{
					"second": "finalValue",
				},
			},
		},
		{
			testDesc: "place value into non-empty map with one key",
			startMap: map[string]interface{}{
				"first": "otherValue",
			},
			keyString:  []string{"second"},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first":  "otherValue",
				"second": "finalValue",
			},
		},
		{
			testDesc: "place value into non-empty map with nested keys",
			startMap: map[string]interface{}{
				"first": map[string]interface{}{
					"second": "otherValue",
				},
			},
			keyString:  []string{"first", "third"},
			finalValue: "finalValue",
			finalMap: map[string]interface{}{
				"first": map[string]interface{}{
					"second": "otherValue",
					"third":  "finalValue",
				},
			},
		},
		{
			testDesc:   "value is non-string",
			startMap:   map[string]interface{}{},
			keyString:  []string{"first", "second"},
			finalValue: 1234,
			finalMap: map[string]interface{}{
				"first": map[string]interface{}{
					"second": 1234,
				},
			},
		},
	}
	for _, tt := range tests {
		SetHelmValueInMap(tt.startMap, tt.keyString, tt.finalValue)
		assert := assert.New(t)
		assert.Equal(tt.startMap, tt.finalMap, fmt.Sprintf("failed case: %s\nGot: %+v\nWanted: %+v", tt.testDesc, tt.startMap, tt.finalMap))
	}
}
