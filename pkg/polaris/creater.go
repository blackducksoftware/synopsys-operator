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

package polaris

import (
	"fmt"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"strings"
)

func GetComponents(baseUrl string, polaris Polaris) (map[string]runtime.Object, error) {
	components := make(map[string]runtime.Object)

	dbComponents, err := GetPolarisDBComponents(baseUrl, polaris)
	if err != nil {
		return nil, err
	}
	polarisComponents, err := GetPolarisComponents(baseUrl, polaris)
	if err != nil {
		return nil, err
	}

	for k, v := range dbComponents {
		components[k] = v
	}
	for k, v := range polarisComponents {
		components[k] = v
	}

	if polaris.EnableReporting {
		reportingComponents, err := GetPolarisReportingComponents(baseUrl, polaris)
		if err != nil {
			return nil, err
		}
		for k, v := range reportingComponents {
			components[k] = v
		}
	}

	provisionComponents, err := GetPolarisProvisionComponents(baseUrl, polaris)
	if err != nil {
		return nil, err
	}
	for k, v := range provisionComponents {
		components[k] = v
	}

	return components, nil
}

// ConvertYamlFileToRuntimeObjects converts the yaml file string to map of runtime object
func ConvertYamlFileToRuntimeObjects(stringContent string) map[string]runtime.Object {
	routev1.AddToScheme(scheme.Scheme)
	securityv1.AddToScheme(scheme.Scheme)

	listOfSingleK8sResourceYaml := strings.Split(stringContent, "---")
	mapOfUniqueIDToDesiredRuntimeObject := make(map[string]runtime.Object, 0)

	for _, singleYaml := range listOfSingleK8sResourceYaml {
		if singleYaml == "\n" || singleYaml == "" {
			// ignore empty cases
			//log.V(1).Info("Got empty", "here", singleYaml)
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		runtimeObject, groupVersionKind, err := decode([]byte(singleYaml), nil, nil)
		if err != nil {
			//log.V(1).Info("unable to decode a single yaml object, skipping", "singleYaml", singleYaml, "error", err)
			continue
		}

		accessor := meta.NewAccessor()
		runtimeObjectKind := groupVersionKind.Kind
		runtimeObjectName, err := accessor.Name(runtimeObject)
		if err != nil {
			//log.V(1).Info("Failed to get runtimeObject's name", "err", err)
			continue
		}
		uniqueID := fmt.Sprintf("%s.%s", runtimeObjectKind, runtimeObjectName)
		//log.V(1).Info("creating runtime object label", "uniqueId", uniqueID)
		mapOfUniqueIDToDesiredRuntimeObject[uniqueID] = runtimeObject
	}
	return mapOfUniqueIDToDesiredRuntimeObject
}
