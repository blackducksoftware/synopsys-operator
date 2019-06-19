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
	log "github.com/sirupsen/logrus"
)

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

// GetNodeAffinityConfigs takes in a podName, and returns all associated []*horizonapi.NodeAffinityConfig based on what the user provided
func (c *Creater) GetNodeAffinityConfigs(podName string) map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig {

	// make an empty NodeAffinityMap
	nodeAffinityMap := make(map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig)

	for _, affinity := range c.blackDuck.Spec.NodeAffinities[podName] {
		log.Debugf("Adding affinity: %v to pod: %v\n", affinity, podName)
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
