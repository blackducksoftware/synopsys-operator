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

package components

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	corev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/core/v1"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
)

// PVC holds common Persistent Volume Claim definitions
type PVC struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
}

// NewPVC returns a new PVC object
func NewPVC(config *protoform.Config, client *kubernetes.Clientset) PVC {
	return PVC{config: config, kubeClient: client}
}

// Generate will craete the pvcs from the Black Duck spec
func (p *PVC) Generate(defaultPVC map[string]string, userPVCs []corev1.PVC, defaultSC string, appName string, ns string, createdByAnnotation string) ([]*components.PersistentVolumeClaim, error) {
	var pvcs []*components.PersistentVolumeClaim
	pvcMap := make(map[string]corev1.PVC)
	for _, claim := range userPVCs {
		pvcMap[claim.Name] = claim
	}

	for name, size := range defaultPVC {
		var claim corev1.PVC

		if _, ok := pvcMap[name]; ok {
			claim = pvcMap[name]
		} else {
			claim = corev1.PVC{
				Name:         name,
				Size:         size,
				StorageClass: defaultSC,
			}
		}

		// Set the claim name to be app specific if the PVC was not created by an operator version prior to
		// 2019.6.0
		if createdByAnnotation != "pre-2019.6.0" {
			claim.Name = apputils.GetResourceName(appName, "", name)
		}

		pvc, err := p.createPVC(claim, horizonapi.ReadWriteOnce, apputils.GetLabel("pvc", appName), ns)
		if err != nil {
			return nil, err
		}
		pvcs = append(pvcs, pvc)
	}
	return pvcs, nil
}

func (p *PVC) createPVC(claim corev1.PVC, accessMode horizonapi.PVCAccessModeType, label map[string]string, namespace string) (*components.PersistentVolumeClaim, error) {
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
		return nil, err
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
