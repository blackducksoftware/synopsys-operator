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

package crdupdater

import (
	"reflect"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type label struct {
	operator string
	value    []string
}

// getLabelsMap convert the label selector string to kubernetes label format
func getLabelsMap(labelSelectors string) map[string]label {
	labelSelectorArr := strings.SplitN(labelSelectors, ",", 2)
	expectedLabels := make(map[string]label, len(labelSelectorArr))
	for _, labelSelector := range labelSelectorArr {
		if strings.Contains(labelSelector, "!=") {
			labels := strings.Split(labelSelector, "!=")
			if len(labels) == 2 {
				expectedLabels[labels[0]] = label{operator: "!=", value: []string{labels[1]}}
			}
		} else if strings.Contains(labelSelector, "=") {
			labels := strings.Split(labelSelector, "=")
			if len(labels) == 2 {
				expectedLabels[labels[0]] = label{operator: "=", value: []string{labels[1]}}
			}
		} else if strings.Contains(labelSelector, " in ") {
			labels := strings.Split(labelSelector, " in (")
			if len(labels) == 2 {
				values := []string{}
				valueArr := strings.Split(labels[1][:len(labels[1])-1], ",")
				for _, value := range valueArr {
					values = append(values, value)
				}
				expectedLabels[labels[0]] = label{operator: "in", value: values}
			}
		} else if strings.Contains(labelSelector, " notin ") {
			labels := strings.Split(labelSelector, " notin (")
			if len(labels) == 2 {
				values := []string{}
				valueArr := strings.Split(labels[1][:len(labels[1])-1], ",")
				for _, value := range valueArr {
					values = append(values, value)
				}
				expectedLabels[labels[0]] = label{operator: "notin", value: values}
			}
		}

	}
	// fmt.Printf("expected Labels: %+v\n", expectedLabels)
	return expectedLabels
}

// isLabelsExist checks for the expected labels to match with the actual labels
func isLabelsExist(expectedLabels map[string]label, actualLabels map[string]string) bool {
	for key, expectedValue := range expectedLabels {
		actualValue, ok := actualLabels[key]
		if !ok {
			return false
		}
		switch expectedValue.operator {
		case "=":
			if !strings.EqualFold(expectedValue.value[0], actualValue) {
				return false
			}
		case "!=":
			if strings.EqualFold(expectedValue.value[0], actualValue) {
				return false
			}
		case "in":
			isExist := false
			for _, value := range expectedValue.value {
				if strings.EqualFold(value, actualValue) {
					isExist = true
					break
				}
			}
			if !isExist {
				return false
			}
		case "notin":
			isExist := false
			for _, value := range expectedValue.value {
				if strings.EqualFold(value, actualValue) {
					isExist = true
					break
				}
			}
			if isExist {
				return false
			}
		}
	}
	return true
}

func sortEnvs(envs []corev1.EnvVar) []corev1.EnvVar {
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })
	return envs
}

func sortVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	sort.Slice(volumeMounts, func(i, j int) bool { return volumeMounts[i].Name < volumeMounts[j].Name })
	return volumeMounts
}

func compareVolumes(oldVolume []corev1.Volume, newVolume []corev1.Volume) bool {
	for i, volume := range oldVolume {
		if volume.Secret != nil && !reflect.DeepEqual(volume.Secret.SecretName, newVolume[i].Secret.SecretName) && !reflect.DeepEqual(volume.Secret.Items, newVolume[i].Secret.Items) {
			return false
		} else if volume.ConfigMap != nil && !reflect.DeepEqual(volume.ConfigMap.Name, newVolume[i].ConfigMap.Name) && !reflect.DeepEqual(volume.ConfigMap.Items, newVolume[i].ConfigMap.Items) {
			return false
		} else if volume.Secret == nil && volume.ConfigMap == nil && !reflect.DeepEqual(oldVolume, newVolume) {
			return false
		}
	}
	return true
}

func sortVolumes(volumes []corev1.Volume) []corev1.Volume {
	for _, volume := range volumes {
		if volume.Secret != nil {
			sort.Slice(volume.Secret.Items, func(i, j int) bool { return volume.Secret.Items[i].Key < volume.Secret.Items[j].Key })
		}
		if volume.ConfigMap != nil {
			sort.Slice(volume.ConfigMap.Items, func(i, j int) bool { return volume.ConfigMap.Items[i].Key < volume.ConfigMap.Items[j].Key })
		}
	}
	sort.Slice(volumes, func(i, j int) bool { return volumes[i].Name < volumes[j].Name })
	return volumes
}
