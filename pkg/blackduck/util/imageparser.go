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
	"regexp"
)

var imageTagRegexp = regexp.MustCompile(`([0-9a-zA-Z-_:\\.]*)/([0-9a-zA-Z-_:\\.]*):([a-zA-Z0-9-\\._]+)$`)

// ParseImageString will take a docker image string and return the repo and tag parts
func ParseImageString(image string) (int, error) {
	tagMatch := imageTagRegexp.FindStringSubmatch(image)
	if len(tagMatch) >= 4 {
		return len(tagMatch), nil
	}

	return 0, fmt.Errorf("unable to parse the input image %s for regex %+v", image, imageTagRegexp)
}
