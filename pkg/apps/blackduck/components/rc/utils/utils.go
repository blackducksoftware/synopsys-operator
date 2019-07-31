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

package utils

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetDBSecretVolume get database secret volume
func GetDBSecretVolume(name string) *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: utils.GetResourceName(name, util.BlackDuckName, "db-creds"),
		Items: []horizonapi.KeyPath{
			{Key: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Path: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: util.IntToInt32(420)},
			{Key: "HUB_POSTGRES_USER_PASSWORD_FILE", Path: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}

// GetProxyVolume get proxy certificate secret volume
func GetProxyVolume(name string) *components.Volume {
	return components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "blackduck-proxy-certificate",
		MapOrSecretName: utils.GetResourceName(name, util.BlackDuckName, "proxy-certificate"),
		Items: []horizonapi.KeyPath{
			{Key: "HUB_PROXY_CERT_FILE", Path: "HUB_PROXY_CERT_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})
}

// GetBlackDuckConfigEnv get Black Duck environment configuration
func GetBlackDuckConfigEnv(name string) *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: utils.GetResourceName(name, util.BlackDuckName, "config")}
}

// GetBlackDuckDBConfigEnv get Black Duck database environment configuration
func GetBlackDuckDBConfigEnv(name string) *horizonapi.EnvConfig {
	return &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: utils.GetResourceName(name, util.BlackDuckName, "db-config")}
}

var affTypeMap = map[string]horizonapi.AffinityType{
	"Hard": horizonapi.AffinityHard,
	"Soft": horizonapi.AffinitySoft,
}

var nodeOperatorMap = map[string]horizonapi.NodeOperator{
	"In":           horizonapi.NodeOperatorIn,
	"NotIn":        horizonapi.NodeOperatorNotIn,
	"Exists":       horizonapi.NodeOperatorExists,
	"DoesNotExist": horizonapi.NodeOperatorDoesNotExist,
	"Gt":           horizonapi.NodeOperatorGt,
	"Lt":           horizonapi.NodeOperatorLt,
}

// GetNodeAffinityConfigs get node affinity configuration
func GetNodeAffinityConfigs(podName string, bdspec *blackduckapi.BlackduckSpec) map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig {

	// make an empty NodeAffinityMap
	nodeAffinityMap := make(map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig)

	for _, affinity := range bdspec.NodeAffinities[podName] {
		nodeAffinityMap[affTypeMap[affinity.AffinityType]] = append(nodeAffinityMap[affTypeMap[affinity.AffinityType]],
			&horizonapi.NodeAffinityConfig{
				Expressions: []horizonapi.NodeExpression{
					{
						Key:    affinity.Key,
						Op:     nodeOperatorMap[affinity.Op],
						Values: affinity.Values,
					},
				},
			},
		)
	}

	return nodeAffinityMap
}

// GetPVCName get PVC name
func GetPVCName(name string, blackduck *blackduckapi.Blackduck) string {
	if blackduck.Annotations["synopsys.com/created.by"] == "pre-2019.6.0" {
		return fmt.Sprintf("blackduck-%s", name)
	}
	return utils.GetResourceName(blackduck.Name, util.BlackDuckName, name)
}
