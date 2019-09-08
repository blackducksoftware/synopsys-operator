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

package v1

// PVC will contain the specifications of the different PVC.
// This will overwrite the default claim configuration
type PVC struct {
	Name         string `json:"name"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	VolumeName   string `json:"volumeName,omitempty"`
}

// RegistryConfiguration contains the registry configuration
type RegistryConfiguration struct {
	Registry    string   `json:"registry,omitempty"`
	PullSecrets []string `json:"pullSecrets,omitempty"`
}

// Environs will hold the list of Environment variables
type Environs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NodeAffinity will contain the specifications of a node affinity
// TODO: currently, keeping it simple, but can be modified in the future to take in complex scenarios
type NodeAffinity struct {
	AffinityType string   `json:"affinityType"`
	Key          string   `json:"key"`
	Op           string   `json:"op"`
	Values       []string `json:"values"`
}
