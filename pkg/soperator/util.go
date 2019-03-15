/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package soperator

type OperatorVersions struct {
	Blackduck CrdData
	OpsSight  CrdData
	Alert     CrdData
}

type CrdData struct {
	Name    string
	Version string
}

// GetCrdDataList returns a list of CrdData for a version that can be iterated over
func GetCrdDataList(version string) []CrdData {
	data := OperatorVersionLookup[version]
	return []CrdData{data.Blackduck, data.OpsSight, data.Alert}
}

// OperatorVersionLookup is alookup table for crd versions that are compatible with operator verions
var OperatorVersionLookup = map[string]OperatorVersions{
	"2019.0.0": OperatorVersions{
		Blackduck: CrdData{Name: "hub.synopsys.com", Version: "v1"},
		OpsSight:  CrdData{Name: "opssights.synopsys.com", Version: "v1"},
		Alert:     CrdData{Name: "alerts.synopsys.com", Version: "v1"},
	},
	"2019.1.1": OperatorVersions{
		Blackduck: CrdData{Name: "blackducks.synopsys.com", Version: "v1"},
		OpsSight:  CrdData{Name: "opssights.synopsys.com", Version: "v1"},
		Alert:     CrdData{Name: "alerts.synopsys.com", Version: "v1"},
	},
}
