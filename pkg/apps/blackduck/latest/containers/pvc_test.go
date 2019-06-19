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
	"fmt"
	"strings"
	"testing"

	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPVCs(t *testing.T) {
	// Default protoform Config in DryRun mode for this test
	pc := protoform.Config{}
	pc.SelfSetDefaults()
	pc.DryRun = true
	pc.IsClusterScoped = true
	// Black Duck flavor for this test
	flavor := GetContainersFlavor("small")
	// Binary Analysis state for this test
	ba := false
	// Default PVCs for this test
	c := NewCreater(&pc, nil, nil, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: blackduckapi.BlackduckSpec{PersistentStorage: true}}, flavor, ba)
	defaultPVCs := c.GetPVCs()

	// Case: No PVCs specified - defaults
	specPVCs := []blackduckapi.PVC{}
	blackDuckSpec := blackduckapi.BlackduckSpec{
		PersistentStorage: true,
		PVC:               specPVCs,
	}
	c = NewCreater(&pc, nil, nil, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: blackDuckSpec}, flavor, ba)
	createdPVCs := c.GetPVCs()
	err := checkPVCs(defaultPVCs, specPVCs, createdPVCs)
	if err != nil {
		t.Errorf("%s", err)
	}

	// Case: PVC name and storage class are changed
	specPVCs = []blackduckapi.PVC{
		{
			Name:         strings.Split(defaultPVCs[0].GetName(), "test-")[1],
			Size:         "10Gi",
			StorageClass: "testStorageClass",
		},
	}
	blackDuckSpec = blackduckapi.BlackduckSpec{
		PersistentStorage: true,
		PVCStorageClass:   "globalStorageClass",
		PVC:               specPVCs,
	}
	c = NewCreater(&pc, nil, nil, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: blackDuckSpec}, flavor, ba)
	createdPVCs = c.GetPVCs()
	err = checkPVCs(defaultPVCs, specPVCs, createdPVCs)
	if err != nil {
		t.Errorf("%s", err)
	}

	// Case: Invalid PVC Name isn't added
	specPVCs = []blackduckapi.PVC{
		{
			Name:         "badName",
			Size:         "10Gi",
			StorageClass: "testStorageClass",
		},
	}
	blackDuckSpec = blackduckapi.BlackduckSpec{
		PersistentStorage: true,
		PVCStorageClass:   "globalStorageClass",
		PVC:               specPVCs,
	}
	c = NewCreater(&pc, nil, nil, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: blackDuckSpec}, flavor, ba)
	createdPVCs = c.GetPVCs()
	err = checkPVCs(defaultPVCs, []blackduckapi.PVC{}, createdPVCs) // a bad PVC is like it doesn't exist
	if err != nil {
		t.Errorf("%s", err)
	}

	// Case: Bad PVC is ignored and Good PVC is added
	specPVCs = []blackduckapi.PVC{
		{
			Name:         strings.Split(defaultPVCs[0].GetName(), "test-")[1],
			Size:         "10Gi",
			StorageClass: "testStorageClass",
		},
		{
			Name:         "badName",
			Size:         "10Gi",
			StorageClass: "testStorageClass",
		},
	}
	blackDuckSpec = blackduckapi.BlackduckSpec{
		PersistentStorage: true,
		PVCStorageClass:   "globalStorageClass",
		PVC:               specPVCs,
	}
	c = NewCreater(&pc, nil, nil, &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: blackDuckSpec}, flavor, ba)
	createdPVCs = c.GetPVCs()
	err = checkPVCs(defaultPVCs, []blackduckapi.PVC{{Name: strings.Split(defaultPVCs[0].GetName(), "test-")[1], Size: "10Gi", StorageClass: "testStorageClass"}}, createdPVCs)
	if err != nil {
		t.Errorf("%s", err)
	}
}

func checkPVCs(defaultPVCs []*horizoncomponents.PersistentVolumeClaim, expectedPVCs []blackduckapi.PVC, observedPVCs []*horizoncomponents.PersistentVolumeClaim) error {
	// Create PVC mappings
	defaultPVCMap := make(map[string]*horizoncomponents.PersistentVolumeClaim)
	for _, claim := range defaultPVCs {
		defaultPVCMap[claim.GetName()] = claim
	}

	expectedPVCMap := make(map[string]blackduckapi.PVC)
	for _, claim := range expectedPVCs {
		expectedPVCMap[claim.Name] = claim
	}

	observedPVCMap := make(map[string]*horizoncomponents.PersistentVolumeClaim)
	for _, claim := range observedPVCs {
		observedPVCMap[claim.GetName()] = claim
	}

	// Check the number of observed PVCs is the same as default PVCs
	if len(defaultPVCs) != len(observedPVCs) {
		return fmt.Errorf("invalid number of PVCs - expected %+v, got %+v", len(defaultPVCs), len(observedPVCs))
	}

	// Check the observed PVCs have correct default PVC names
	for observedPVCName := range observedPVCMap {
		if _, exists := defaultPVCMap[observedPVCName]; !exists {
			return fmt.Errorf("created PVC %+v is not in the default PVCs for Black Duck", observedPVCName)
		}
	}

	// Check the expected PVC names are in the observed PVC names
	for expectedPVCName := range expectedPVCMap {
		if _, exists := observedPVCMap[fmt.Sprintf("test-%s", expectedPVCName)]; !exists {
			return fmt.Errorf("expected PVC %+v is not in the created PVCs", fmt.Sprintf("test-%s", expectedPVCName))
		}
	}

	// Check the expected PVC values are correctly set in observed PVCs
	for expectedPVCName, expectedPVC := range expectedPVCMap {
		observedPVCObj := observedPVCMap[fmt.Sprintf("test-%s", expectedPVCName)].PersistentVolumeClaim
		defaultPVCObj := defaultPVCMap[fmt.Sprintf("test-%s", expectedPVCName)].PersistentVolumeClaim
		if len(expectedPVC.Size) > 0 { // use the specified size
			observedSize := observedPVCObj.Spec.Resources.Requests[v1.ResourceStorage]
			expectedSize, _ := resource.ParseQuantity(expectedPVC.Size)
			if expectedSize != observedSize {
				return fmt.Errorf("invalid set storage size - expected %+v, got %+v", expectedSize, observedSize)
			}
		} else { // use the default size
			observedSize := observedPVCObj.Spec.Resources.Requests[v1.ResourceStorage]
			defaultSize := defaultPVCObj.Spec.Resources.Requests[v1.ResourceStorage]
			if observedSize != defaultSize {
				return fmt.Errorf("invalid default storage size - expected %+v, got %+v", defaultSize, observedSize)
			}
		}
		if len(expectedPVC.StorageClass) > 0 {
			if expectedPVC.StorageClass != *observedPVCObj.Spec.StorageClassName {
				return fmt.Errorf("invalid storageClass - expected %+v, got %+v", expectedPVC.StorageClass, *observedPVCObj.Spec.StorageClassName)
			}
		}
	}

	return nil
}
