/*
Copyright (C) 2019 Synopsys, Inc.

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

// SOperatorCRDVersionMap is a global lookup table in the package. It maps versions of the operator
// to the resource versions it is compatible with
var SOperatorCRDVersionMap = operatorCRDVersionMap{
	versionMap: map[string]operatorVersions{
		"latest": {
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
		"2019.4.0": {
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
		"2019.1.0": {
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
		"2018.12.0": {
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
	},
}

// operatorCRDVersionMap stores the version map
type operatorCRDVersionMap struct {
	versionMap map[string]operatorVersions
}

// operatorVersions holds data for each resource
// Pointers are used so that fields can be nil if the operator version
// cannot handle a specific resource
type operatorVersions struct {
	Blackduck *crdVersionData
	OpsSight  *crdVersionData
	Alert     *crdVersionData
}

// crdVersionData holds the name of the crd and the version
type crdVersionData struct {
	CRDName    string
	APIVersion string
}

// GetVersions returns a list of strings for the supported Operator Versions
func (m *operatorCRDVersionMap) GetVersions() []string {
	versions := []string{}
	for v := range m.versionMap {
		versions = append(versions, v)
	}
	return versions
}

// GetCRDVersions returns CRDVersionData for an Operator's Version. If the Operator's
// version doesn't exist then it assumes master
func (m *operatorCRDVersionMap) GetCRDVersions(operatorVersion string) operatorVersions {
	versions, ok := m.versionMap[operatorVersion]
	if !ok {
		return m.versionMap["latest"]
	}
	return versions
}

// GetIterableAPIVersions returns a list of CrdData for a version that can be iterated over
func (m *operatorCRDVersionMap) GetIterableCRDData(operatorVersion string) []crdVersionData {
	data := m.GetCRDVersions(operatorVersion)
	CrdDataList := []crdVersionData{}
	if data.Blackduck != nil { // skips resources the operator version cannot handle
		CrdDataList = append(CrdDataList, *data.Blackduck)
	}
	if data.OpsSight != nil {
		CrdDataList = append(CrdDataList, *data.OpsSight)
	}
	if data.Alert != nil {
		CrdDataList = append(CrdDataList, *data.Alert)
	}
	return CrdDataList
}
