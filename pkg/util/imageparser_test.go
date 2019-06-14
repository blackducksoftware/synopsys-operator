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
	"strconv"
	"testing"
)

func TestParseImageVersion(t *testing.T) {
	version := "1.0.6"
	versions, err := ValidateImageVersion(version)
	t.Logf("versions: %+v", versions)
	if err != nil {
		t.Errorf("unable to parse image version: %+v", err)
	}
	if len(versions) == 4 {
		version1, _ := strconv.Atoi(versions[1])
		version3, _ := strconv.Atoi(versions[3])
		if err == nil && version1 >= 1 && version3 > 3 {
			t.Logf("/opt/blackduck/hub/blackduck-upload-cache")
		} else {
			t.Logf("/opt/blackduck/hub/hub-upload-cache")
		}
	}

	version = "2019.1.0"
	_, err = ValidateImageVersion(version)
	if err != nil {
		t.Errorf("unable to parse image version: %+v", err)
	}

	version = "2019.a.b.c"
	versions, err = ValidateImageVersion(version)
	if err == nil {
		t.Errorf("unable to get image version: %+v", versions)
	}

	version = "2019.1.0-SNAPSHOT"
	versions, err = ValidateImageVersion(version)
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
		imageSubstringSubmatch, err := ValidateImageString(image)
		length := len(imageSubstringSubmatch)

		if length != tc.expectedLength {
			t.Errorf("expected length %d got %d, err %+v", tc.expectedLength, length, err)
		}
	}
}

func TestGetImageVersion(t *testing.T) {
	type args struct {
		image string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				image: "docker.io/blackducksoftware/synopsys-operator:2019.4.2",
			},
			want:    "2019.4.2",
			wantErr: false,
		},
		{
			name: "edge",
			args: args{
				image: "artifactory.test.lab:8321/blackducksoftware/synopsys-operator:2019.4.2",
			},
			want:    "2019.4.2",
			wantErr: false,
		},
		{
			name: "no version tag fed",
			args: args{
				image: "docker.io/blackducksoftware/synopsys-operator",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "no version tag, but still two or more splits; also testing weird tag",
			args: args{
				image: "artifactory.test.lab:8321/blackducksoftware/synopsys-operator",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImageTag(tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetImageTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
