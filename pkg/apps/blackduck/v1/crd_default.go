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

package blackduck

import (
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
)

// GetDefault creates a BlackDuck crd configuration object with defaults
func (hc *Creater) GetDefault(bd *blackduckapi.BlackduckSpec) *blackduckapi.BlackduckSpec {
	spec := &blackduckapi.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Size:              "small",
		PVCStorageClass:   "",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC: []blackduckapi.PVC{
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
		},
		CertificateName: "default",
		Type:            "Artifacts",
		Environs:        []string{},
		ImageRegistries: []string{},
		LicenseKey:      "",
	}
	if bd.ExternalPostgres != nil {
		postPVC := blackduckapi.PVC{
			Name: "blackduck-postgres", // only external db is disabled
			Size: "200Gi",
		}
		spec.PVC = append(spec.PVC, postPVC)
	}
	if hc.isBinaryAnalysisEnabled(spec) {
		// add rabbitmq and upload cache
		uploadcacheDataPVC := blackduckapi.PVC{
			Name: "blackduck-uploadcache-data",
			Size: "100Gi",
		}
		uploadcacheKeyPVC := blackduckapi.PVC{
			Name: "blackduck-uploadcache-key",
			Size: "2Gi",
		}
		rabbitmqPVC := blackduckapi.PVC{
			Name: "blackduck-rabbitmq",
			Size: "5Gi",
		}
		spec.PVC = append(spec.PVC, uploadcacheDataPVC)
		spec.PVC = append(spec.PVC, uploadcacheKeyPVC)
		spec.PVC = append(spec.PVC, rabbitmqPVC)
	}
	return spec
}
