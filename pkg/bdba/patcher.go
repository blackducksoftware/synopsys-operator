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

package bdba

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func getPatchedRuntimeObjects(yamlManifests string, bdba BDBA) (map[string]runtime.Object, error) {
	yamlManifests = patchYamlManifestPlaceHolders(yamlManifests, bdba)

	// Convert the yaml manifests to Kubernetes Runtime Objects
	mapOfUniqueIDToRuntimeObject, err := util.ConvertYamlFileToRuntimeObjects(yamlManifests)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml manifests to runtime objects: %+v", err)
	}

	patcher := RuntimeObjectPatcher{
		bdba:                         bdba,
		mapOfUniqueIDToRuntimeObject: mapOfUniqueIDToRuntimeObject,
	}

	rtoMap, err := patcher.patch()
	if err != nil {
		return nil, fmt.Errorf("failed to path runtime objects: %+v", err)
	}
	return rtoMap, nil
}

func patchYamlManifestPlaceHolders(yamlManifests string, bdba BDBA) string {
	// Patch the yaml file yamlManifests
	yamlManifests = strings.ReplaceAll(yamlManifests, "${NAME}", bdba.Name)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${NAMESPACE}", bdba.Namespace)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${VERSION}", bdba.Version)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${HOSTNAME}", bdba.Hostname)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${INGRESS_HOST}", bdba.IngressHost)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_ACCESS_KEY}", bdba.MinioAccessKey)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${MINIO_SECRET_KEY}", bdba.MinioSecretKey)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${WORKER_REPLICAS}", strconv.Itoa(bdba.WorkerReplicas))
	yamlManifests = strings.ReplaceAll(yamlManifests, "nginx.ingress.kubernetes.io/proxy-request-buffering: false", "nginx.ingress.kubernetes.io/proxy-request-buffering: \"false\"")

	yamlManifests = strings.ReplaceAll(yamlManifests, "${ADMIN_EMAIL}", bdba.AdminEmail)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${BROKER_URL}", bdba.BrokerURL)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${PGPASSWORD}", bdba.PGPPassword)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${RABBITMQ_ULIMIT_NOFILES}", bdba.RabbitMQULimitNoFiles)
	//"${RABBITMQ_PASSWORD//\\/\\\\}"

	yamlManifests = strings.ReplaceAll(yamlManifests, "${HIDE_LICENSES}", bdba.HideLicenses)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LICENSING_PASSWORD}", bdba.LicensingPassword)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${LICENSING_USERNAME}", bdba.LicensingUsername)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${INSECURE_COOKIES}", bdba.InsecureCookies)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${SESSION_COOKIE_AGE}", bdba.SessionCookieAge)

	yamlManifests = strings.ReplaceAll(yamlManifests, "${URL}", bdba.URL)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${ACTUAL}", bdba.Actual)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${EXPECTED}", bdba.Expected)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${START_FLAG}", bdba.StartFlag)
	yamlManifests = strings.ReplaceAll(yamlManifests, "${RESULT}", bdba.Result)

	return yamlManifests
}

// RuntimeObjectPatcher holds the BDBA run time objects and it is having methods to patch it
type RuntimeObjectPatcher struct {
	bdba                         BDBA
	mapOfUniqueIDToRuntimeObject map[string]runtime.Object
}

func (p *RuntimeObjectPatcher) patch() (map[string]runtime.Object, error) {
	patches := []func() error{
		p.patchNamespace,
	}
	for _, patchFunc := range patches {
		err := patchFunc()
		if err != nil {
			return nil, err
		}
	}
	return p.mapOfUniqueIDToRuntimeObject, nil
}

// patchNamespace will change the resource namespace
func (p *RuntimeObjectPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.mapOfUniqueIDToRuntimeObject {
		accessor.SetNamespace(runtimeObject, p.bdba.Namespace)
	}
	return nil
}
