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

package util

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GenPVC will generate the list of PVCs
func GenPVC(defaultPVC map[string]string, blackDuck *blackduckapi.Blackduck) ([]*components.PersistentVolumeClaim, error) {
	var pvcs []*components.PersistentVolumeClaim
	if blackDuck.Spec.PersistentStorage {
		pvcMap := make(map[string]blackduckapi.PVC)
		for _, claim := range blackDuck.Spec.PVC {
			pvcMap[claim.Name] = claim
		}

		for name, size := range defaultPVC {
			var claim blackduckapi.PVC

			if _, ok := pvcMap[name]; ok {
				claim = pvcMap[name]
				if len(claim.StorageClass) == 0 {
					claim.StorageClass = blackDuck.Spec.PVCStorageClass
				}
			} else {
				claim = blackduckapi.PVC{
					Name:         name,
					Size:         size,
					StorageClass: blackDuck.Spec.PVCStorageClass,
				}
			}

			// Set the claim name to be app specific if the PVC was not created by an operator version prior to
			// 2019.6.0
			if blackDuck.Annotations["synopsys.com/created.by"] != "pre-2019.6.0" {
				claim.Name = util.GetResourceName(blackDuck.Name, "", name)
			}

			pvc, err := createPVC(claim, horizonapi.ReadWriteOnce, getLabel(blackDuck.Name, "pvc"), blackDuck.Spec.Namespace)
			if err != nil {
				return nil, fmt.Errorf("failed to create PVC object '%s': %+v", name, err)
			}
			pvcs = append(pvcs, pvc)
		}
	}
	return pvcs, nil
}

func createPVC(claim blackduckapi.PVC, accessMode horizonapi.PVCAccessModeType, label map[string]string, namespace string) (*components.PersistentVolumeClaim, error) {
	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string

	if len(claim.StorageClass) > 0 {
		class = &claim.StorageClass
	} else {
		class = nil
	}

	var size string
	_, err := resource.ParseQuantity(claim.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PVC size: %+v", err)
	}
	size = claim.Size

	config := horizonapi.PVCConfig{
		Name:      claim.Name,
		Namespace: namespace,
		Size:      size,
		Class:     class,
	}

	if len(claim.VolumeName) > 0 {
		// Needed so that it doesn't use the default storage class
		var tmp = ""
		config.Class = &tmp
		config.VolumeName = claim.VolumeName
	}

	pvc, err := components.NewPersistentVolumeClaim(config)
	if err != nil {
		return nil, err
	}

	pvc.AddAccessMode(accessMode)
	pvc.AddLabels(label)

	return pvc, nil
}

// getLabel will return the label
func getLabel(name string, componentName string) map[string]string {
	return map[string]string{
		"app":       util.BlackDuckName,
		"name":      name,
		"component": componentName,
	}
}
