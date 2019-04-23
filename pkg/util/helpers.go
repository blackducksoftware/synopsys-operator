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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
)

// Container defines the configuration for a container
type Container struct {
	ContainerConfig       *horizonapi.ContainerConfig
	EnvConfigs            []*horizonapi.EnvConfig
	VolumeMounts          []*horizonapi.VolumeMountConfig
	PortConfig            []*horizonapi.PortConfig
	ActionConfig          *horizonapi.ActionConfig
	ReadinessProbeConfigs []*horizonapi.ProbeConfig
	LivenessProbeConfigs  []*horizonapi.ProbeConfig
	PreStopConfig         *horizonapi.ActionConfig
}

// MergeEnvMaps will merge the source and destination environs. If the same value exist in both, destination environ will given more preference
func MergeEnvMaps(source, destination map[string]string) map[string]string {
	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for key, value := range source {
		if len(value) == 0 {
			delete(destination, key)
		} else {
			destination[key] = value
		}
	}
	return destination
}

// MergeEnvSlices will merge the source and destination environs. If the same value exist in both, destination environ will given more preference
func MergeEnvSlices(source, destination []string) []string {
	// create a destination map
	destinationMap := make(map[string]string)
	for _, value := range destination {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for _, value := range source {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapValue) == 0 {
				delete(destinationMap, mapKey)
			} else {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// convert destination map to string array
	mergedValues := []string{}
	for key, value := range destinationMap {
		mergedValues = append(mergedValues, fmt.Sprintf("%s:%s", key, value))
	}
	return mergedValues
}
