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
	"time"
)

func TestMergeEnvMaps(t *testing.T) {
	var tests = []struct {
		description string
		source      map[string]string
		destination map[string]string
		expected    map[string]string
	}{
		{
			description: "nothing is done for empty",
			source:      map[string]string{},
			destination: map[string]string{},
			expected:    map[string]string{},
		},
		{
			description: "source value is kept",
			source:      map[string]string{"key1": "val1"},
			destination: map[string]string{},
			expected:    map[string]string{"key1": "val1"},
		},
		{
			description: "destination is put into source",
			source:      map[string]string{},
			destination: map[string]string{"key1": "val1"},
			expected:    map[string]string{"key1": "val1"},
		},
		{
			description: "source value is given preference",
			source:      map[string]string{"key1": "valSource"},
			destination: map[string]string{"key1": "valDest"},
			expected:    map[string]string{"key1": "valSource"},
		},
		{
			description: "source value is empty, delete it from destination",
			source:      map[string]string{"key1": "", "key2": "val2"},
			destination: map[string]string{"key1": "val1", "key3": "val3"},
			expected:    map[string]string{"key2": "val2", "key3": "val3"},
		},
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
		{
			description: "nothing is done for empty",
			source:      []string{},
			destination: []string{},
			expected:    []string{},
		},
		{
			description: "source value is kept",
			source:      []string{"key1:val1"},
			destination: []string{},
			expected:    []string{"key1:val1"},
		},
		{
			description: "destination is put into source",
			source:      []string{},
			destination: []string{"key1:val1"},
			expected:    []string{"key1:val1"},
		},
		{
			description: "source value is given preference",
			source:      []string{"key1:valSource"},
			destination: []string{"key1:valDest"},
			expected:    []string{"key1:valSource"},
		},
		{
			description: "source value is empty",
			source:      []string{"key1:", "key2:val2"},
			destination: []string{"key1:val1", "key3:val3"},
			expected:    []string{"key2:val2", "key3:val3"},
		},
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
		if v := reflect.DeepEqual(test.expected, observed); !v {
			t.Errorf("failed to merge slices '%s', expected %+v, got %+v", test.description, test.expected, observed)
		}
	}
}

func TestGetResourceName(t *testing.T) {
	type args struct {
		name        string
		appName     string
		defaultName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no appName",
			args: args{
				name:        "name",
				appName:     "",
				defaultName: "defaultName",
			},
			want: "name-defaultName",
		},
		{
			name: "appName, no defaultName",
			args: args{
				name:        "name",
				appName:     "appName",
				defaultName: "",
			},
			want: "name-appName",
		},
		{
			name: "appName, defaultName",
			args: args{
				name:        "name",
				appName:     "appName",
				defaultName: "defaultName",
			},
			want: "name-appName-defaultName",
		},
		// now not covered
		{
			name: "no appName, no defaultName",
			args: args{
				name:        "name",
				appName:     "",
				defaultName: "",
			},
			want: "name-",
		},
		{
			name: "all empty",
			args: args{
				name:        "",
				appName:     "",
				defaultName: "",
			},
			want: "-",
		},
		{
			name: "just defaultName",
			args: args{
				name:        "",
				appName:     "",
				defaultName: "defaultName",
			},
			want: "-defaultName",
		},
		{
			name: "just appName",
			args: args{
				name:        "",
				appName:     "appName",
				defaultName: "",
			},
			want: "-appName",
		},
		{
			name: "no name",
			args: args{
				name:        "",
				appName:     "appName",
				defaultName: "defaultName",
			},
			want: "-appName-defaultName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourceName(tt.args.name, tt.args.appName, tt.args.defaultName); got != tt.want {
				t.Errorf("GetResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVersionGreaterThanOrEqualTo(t *testing.T) {
	type args struct {
		version      string
		majorRelease int
		minorRelease time.Month
		dotRelease   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "less version",
			args: args{
				version:      "2019.8.0",
				majorRelease: 2019,
				minorRelease: time.September,
				dotRelease:   0,
			},
			want: false,
		},
		{
			name: "equal version",
			args: args{
				version:      "2019.8.0",
				majorRelease: 2019,
				minorRelease: time.August,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater minor version",
			args: args{
				version:      "2019.10.0",
				majorRelease: 2019,
				minorRelease: time.August,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater dot version",
			args: args{
				version:      "2019.12.0",
				majorRelease: 2019,
				minorRelease: time.October,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater major version",
			args: args{
				version:      "2020.1.0",
				majorRelease: 2019,
				minorRelease: time.December,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater major version comparing against greater minor version",
			args: args{
				version:      "2020.1.0",
				majorRelease: 2019,
				minorRelease: time.August,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater minor version comparing against greater dot release",
			args: args{
				version:      "2019.10.0",
				majorRelease: 2019,
				minorRelease: time.June,
				dotRelease:   1,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := IsVersionGreaterThanOrEqualTo(tt.args.version, tt.args.majorRelease, tt.args.minorRelease, tt.args.dotRelease); got != tt.want {
				t.Errorf("IsVersionGreaterThanOrEqualTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotDefaultVersionGreaterThanOrEqualTo(t *testing.T) {
	type args struct {
		version      string
		majorRelease int
		minorRelease int
		dotRelease   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "less version",
			args: args{
				version:      "4.2.0",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   0,
			},
			want: false,
		},
		{
			name: "equal version",
			args: args{
				version:      "5.0.0",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater minor version",
			args: args{
				version:      "5.2.0",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater dot version",
			args: args{
				version:      "5.0.1",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater major version",
			args: args{
				version:      "6.0.0",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater major version comparing against greater minor version",
			args: args{
				version:      "6.0.0",
				majorRelease: 5,
				minorRelease: 1,
				dotRelease:   0,
			},
			want: true,
		},
		{
			name: "greater minor version comparing against greater dot release",
			args: args{
				version:      "5.1.0",
				majorRelease: 5,
				minorRelease: 0,
				dotRelease:   1,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := IsNotDefaultVersionGreaterThanOrEqualTo(tt.args.version, tt.args.majorRelease, tt.args.minorRelease, tt.args.dotRelease); got != tt.want {
				t.Errorf("IsNotDefaultVersionGreaterThanOrEqualTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
