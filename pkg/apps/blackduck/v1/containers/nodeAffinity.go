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
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
)

var affTypeMap = map[string]horizonapi.AffinityType{
	"AffinityHard": horizonapi.AffinityHard,
	"AffinitySoft": horizonapi.AffinitySoft,
}

var nodeOperatorMap = map[string]horizonapi.NodeOperator{
	"In":           horizonapi.NodeOperatorIn,
	"NotIn":        horizonapi.NodeOperatorNotIn,
	"Exists":       horizonapi.NodeOperatorExists,
	"DoesNotExist": horizonapi.NodeOperatorDoesNotExist,
	"Gt":           horizonapi.NodeOperatorGt,
	"Lt":           horizonapi.NodeOperatorLt,
}

var affinityMap = make(map[string][]blackduckapi.NodeAffinity)

// getAffinitiesForSpecificPod iterates once through the user provided NodeAffinities, internally caches them into a map of format "podName":[]blackduckapi.NodeAffinity, and returns []blackduckapi.NodeAffinity for the given "podName"
func (c *Creater) getAffinitiesForSpecificPod(podName string) []blackduckapi.NodeAffinity {
	if affinitiesForSpecificPod, ok := affinityMap[podName]; ok {
		return affinitiesForSpecificPod
	}

	affinitiesForSpecificPod := []blackduckapi.NodeAffinity{}
	for _, affinity := range c.hubSpec.NodeAffinities {
		affinityMap[affinity.PodName] = append(affinitiesForSpecificPod, affinity)
	}
	return affinitiesForSpecificPod
}

// GetNodeAffinityConfigs takes in a podName, and returns all associated []*horizonapi.NodeAffinityConfig based on what the user provided
func (c *Creater) GetNodeAffinityConfigs(podName string) map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig {

	// make an empty NodeAffinityMap
	nodeAffinityMap := make(map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig)

	for _, affinity := range c.getAffinitiesForSpecificPod(podName) {
		nodeAffinityMap[affTypeMap[affinity.AffinityType]] = append(nodeAffinityMap[affTypeMap[affinity.AffinityType]],
			&horizonapi.NodeAffinityConfig{
				Expressions: []horizonapi.NodeExpression{
					{
						Key:    affinity.Key,
						Op:     nodeOperatorMap[affinity.Op],
						Values: []string{affinity.Value},
					},
				},
			},
		)
	}

	return nodeAffinityMap
}
