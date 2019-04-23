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

package containers

import (
	"strings"
)

const envOptions = `
IPV4_ONLY=0
USE_ALERT=0
USE_BINARY_UPLOADS=0`

// GetHubKnobs ...
func GetHubKnobs() (env map[string]string, images []string) {
	env = map[string]string{}
	images = []string{}
	for _, val := range strings.Split(envOptions, "\n") {
		if strings.Contains(val, "=") {
			keyval := strings.Split(val, "=")
			env[keyval[0]] = keyval[1]
		} else if strings.Contains(val, "image") {
			fullImage := strings.Split(val, ": ")
			images = append(images, fullImage[1])
		}
	}
	return env, images
}
