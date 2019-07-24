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

package alert

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"reflect"
	"testing"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSpecConfig_getAlertPersistentVolumeClaim(t *testing.T) {
	tests := []struct {
		name          string
		altSpecConfig *SpecConfig
		wantErr       bool
	}{
		{
			name: "base test",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:       "alert",
						PVCSize:         "1G",
						PVCStorageClass: "standard",
					},
				},
				isClusterScope: false,
			},
			wantErr: false,
		},
		{
			name: "invalid size",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:       "alert",
						PVCSize:         "1GB",
						PVCStorageClass: "standard",
					},
				},
				isClusterScope: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.altSpecConfig.getAlertPersistentVolumeClaim()
			if (err != nil) != tt.wantErr {
				t.Errorf("SpecConfig.getAlertPersistentVolumeClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				size, _ := resource.ParseQuantity(tt.altSpecConfig.alert.Spec.PVCSize)
				want := &v1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PersistentVolumeClaim",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      utils.GetResourceName(tt.altSpecConfig.alert.Name, util.AlertName, tt.altSpecConfig.alert.Spec.PVCName),
						Namespace: tt.altSpecConfig.alert.Spec.Namespace,
						Labels:    map[string]string{"app": util.AlertName, "component": "alert", "name": tt.altSpecConfig.alert.Name},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: size,
							},
						},
						StorageClassName: &tt.altSpecConfig.alert.Spec.PVCStorageClass,
					},
				}
				if !reflect.DeepEqual(got.PersistentVolumeClaim, want) {
					t.Errorf("SpecConfig.getAlertPersistentVolumeClaim() = %v, want %v", got, want)
				}
			}
		})
	}
}
