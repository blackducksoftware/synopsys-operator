/*
Copyright (C) 2018 Synopsys, Inc.

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
	"fmt"
	"testing"
)

func TestParseImageVersion(t *testing.T) {
	version := "2019.1.0"
	_, err := ParseImageVersion(version)
	if err != nil {
		t.Errorf("unable to parse image version: %+v", err)
	}

	version = "2019.a.b.c"
	versions, err := ParseImageVersion(version)
	if err == nil {
		t.Errorf("unable to get image version: %+v", versions)
	}

	version = "2019.1.0-SNAPSHOT"
	versions, err = ParseImageVersion(version)
	if err == nil {
		t.Errorf("unable to get image version: %+v", versions)
	}
}

func TestParseImageString(t *testing.T) {
	testcases := []struct {
		description    string
		repo           string
		tag            string
		expectedLength int
	}{
		{
			description:    "repo with path and tag",
			repo:           "url.com/imagename",
			tag:            "latest",
			expectedLength: 4,
		},
		{
			description:    "repo with path and tag",
			repo:           "url.com/projectname/imagename",
			tag:            "5.0.0",
			expectedLength: 4,
		},
		{
			description:    "repo with path without tag",
			repo:           "url.com/imagename",
			tag:            "",
			expectedLength: 0,
		},
		{
			description:    "repo with path and port and tag",
			repo:           "url.com:80/imagename",
			tag:            "latest",
			expectedLength: 4,
		},
		{
			description:    "repo with path and port without tag",
			repo:           "url.com:80/imagename",
			tag:            "",
			expectedLength: 0,
		},
		{
			description:    "image name only with tag",
			repo:           "imagename",
			tag:            "1.2.3",
			expectedLength: 0,
		},
		{
			description:    "image name only without tag",
			repo:           "imagename",
			tag:            "",
			expectedLength: 0,
		},
	}

	for _, tc := range testcases {
		var image string
		if len(tc.tag) > 0 {
			image = fmt.Sprintf("%s:%s", tc.repo, tc.tag)
		} else {
			image = tc.repo
		}
		length, err := ParseImageString(image)

		if length != tc.expectedLength {
			t.Errorf("expected length %d got %d, err %+v", tc.expectedLength, length, err)
		}
	}
}
