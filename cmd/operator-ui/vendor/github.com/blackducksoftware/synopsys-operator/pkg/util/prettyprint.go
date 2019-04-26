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
	"encoding/json"
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// PrintFormat represents the format to print the struct
type PrintFormat string

// Constants for the PrintFormats
const (
	JSON PrintFormat = "json"
	YAML PrintFormat = "yaml"
)

// PrettyPrint will print the interface in string format
func PrettyPrint(v interface{}, format string) (string, error) {
	var b []byte
	var err error
	switch {
	case format == string(JSON):
		b, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("Failed to convert struct to yaml. Struct: %+v", v)
		}
		fmt.Println(string(b))
	case format == string(YAML):
		b, err = yaml.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("Failed to convert struct to yaml. Struct: %+v", v)
		}
		fmt.Println(string(b))
	default:
		return "", fmt.Errorf("%s is an invalid format for PrettyPrint", format)
	}
	return string(b), nil
}
