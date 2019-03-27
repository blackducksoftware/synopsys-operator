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

// SOperatorCRDVersionMap is alookup table for crd versions that are compatible with the operator verions
var SOperatorCRDVersionMap = operatorCRDVersionMap{
	versionMap: map[string]operatorVersions{
		"master": operatorVersions{
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
		"2019.1.0": operatorVersions{
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
		"2018.12.0": operatorVersions{
			Blackduck: &crdVersionData{CRDName: "blackducks.synopsys.com", APIVersion: "v1"},
			OpsSight:  &crdVersionData{CRDName: "opssights.synopsys.com", APIVersion: "v1"},
			Alert:     &crdVersionData{CRDName: "alerts.synopsys.com", APIVersion: "v1"},
		},
	},
}

type operatorCRDVersionMap struct {
	versionMap map[string]operatorVersions
}

type operatorVersions struct {
	Blackduck *crdVersionData
	OpsSight  *crdVersionData
	Alert     *crdVersionData
}

type crdVersionData struct {
	CRDName    string
	APIVersion string
}

// GetCRDVersions returns CRD Version Data for the Operator's Version. If the Operator's
// version doesn't exist then it assumes master
func (m *operatorCRDVersionMap) GetCRDVersions(operatorVersion string) operatorVersions {
	versions, ok := m.versionMap[operatorVersion]
	if !ok {
		return m.versionMap["master"]
	}
	return versions
}

// GetIterableAPIVersions returns a list of CrdData for a version that can be iterated over
func (m *operatorCRDVersionMap) GetIterableCRDData(operatorVersion string) []crdVersionData {
	data := m.GetCRDVersions(operatorVersion)
	CrdDataList := []crdVersionData{}
	if data.Blackduck != nil {
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
