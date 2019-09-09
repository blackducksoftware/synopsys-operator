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
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	"sigs.k8s.io/yaml"
)

// PrintFormat represents the format to print the struct
type PrintFormat string

// Constants for the PrintFormats
const (
	JSON PrintFormat = "JSON"
	YAML PrintFormat = "YAML"
)

// PrintComponents outputs components for a CRD in the correct format for 'kubectl create -f <file>'
func PrintComponents(objs []runtime.Object, format string) error {
	for i, obj := range objs {
		_, err := PrintComponent(obj, format)
		if err != nil {
			return fmt.Errorf("failed to print components: %s", err)
		}
		if i != len(objs)-1 && format == "yaml" {
			fmt.Printf("---\n")
		}
	}
	return nil
}

// PrintComponent will print the interface in either json or yaml format
func PrintComponent(v interface{}, format string) (string, error) {
	var b []byte
	var err error
	switch {
	case strings.ToUpper(format) == string(JSON):
		b, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to convert struct to json. err: %+v. struct: %+v", err, v)
		}
		fmt.Println(string(b))
	case strings.ToUpper(format) == string(YAML):
		b, err = yaml.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to convert struct to yaml. err: %+v. struct: %+v", err, v)
		}
		fmt.Println(string(b))
	default:
		return "", fmt.Errorf("'%s' is an invalid format", format)
	}
	return string(b), nil
}
