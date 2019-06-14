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

package util

import (
	"reflect"
	"sort"
	"testing"
)

func TestMergeEnvMaps(t *testing.T) {
	var tests = []struct {
		description string
		source      map[string]string
		destination map[string]string
		expected    map[string]string
	}{
		// case
		{
			description: "nothing is done for empty",
			source:      map[string]string{},
			destination: map[string]string{},
			expected:    map[string]string{},
		},
		// case
		{
			description: "source value is kept",
			source:      map[string]string{"key1": "val1"},
			destination: map[string]string{},
			expected:    map[string]string{"key1": "val1"},
		},
		// case
		{
			description: "destination is put into source",
			source:      map[string]string{},
			destination: map[string]string{"key1": "val1"},
			expected:    map[string]string{"key1": "val1"},
		},
		// case
		{
			description: "source value is given preference",
			source:      map[string]string{"key1": "valSource"},
			destination: map[string]string{"key1": "valDest"},
			expected:    map[string]string{"key1": "valSource"},
		},
		// case
		{
			description: "source and destination values are merged together",
			source:      map[string]string{"key3": "val3", "key1": "val1", "key2": "val2"},
			destination: map[string]string{"key4": "val4", "key6": "val6", "key5": "val5"},
			expected:    map[string]string{"key1": "val1", "key2": "val2", "key3": "val3", "key4": "val4", "key5": "val5", "key6": "val6"},
		},
	}

	for _, test := range tests {
		observed := MergeEnvMaps(test.source, test.destination)
		if v := reflect.DeepEqual(test.expected, observed); !v {
			t.Errorf("failed to merge slices '%s', expected %+v, got %+v", test.description, test.expected, observed)
		}
	}
}

func TestMergeEnvSlices(t *testing.T) {
	var tests = []struct {
		description string
		source      []string
		destination []string
		expected    []string
	}{
		// case
		{
			description: "nothing is done for empty",
			source:      []string{},
			destination: []string{},
			expected:    []string{},
		},
		// case
		{
			description: "source value is kept",
			source:      []string{"key1:val1"},
			destination: []string{},
			expected:    []string{"key1:val1"},
		},
		// case
		{
			description: "destination is put into source",
			source:      []string{},
			destination: []string{"key1:val1"},
			expected:    []string{"key1:val1"},
		},
		// case
		{
			description: "source value is given preference",
			source:      []string{"key1:valSource"},
			destination: []string{"key1:valDest"},
			expected:    []string{"key1:valSource"},
		},
		// case
		{
			description: "source and destination values are merged together",
			source:      []string{"key3:val3", "key1:val1", "key2:val2"},
			destination: []string{"key4:val4", "key6:val6", "key5:val5"},
			expected:    []string{"key1:val1", "key2:val2", "key3:val3", "key4:val4", "key5:val5", "key6:val6"},
		},
	}

	for _, test := range tests {
		observed := MergeEnvSlices(test.source, test.destination)
		sort.Strings(test.expected)
		sort.Strings(observed)
		if v := SlicesEqual(test.expected, observed); !v {
			t.Errorf("failed to merge slices '%s', expected %+v, got %+v", test.description, test.expected, observed)
		}
	}
}

// SlicesEqual tells whether a and b contain the same elements.
// Elements must be in the same order.
// A nil argument is equivalent to an empty slice.
func SlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
