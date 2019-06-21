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
	"reflect"
	"testing"

	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSpecConfig_getAlertSecret(t *testing.T) {
	tests := []struct {
		name          string
		altSpecConfig *SpecConfig
		wantErr       bool
	}{
		{
			name: "base test, namespace scope",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:            "alert",
						EncryptionPassword:   "1234567890123456",
						EncryptionGlobalSalt: "1234567890123456",
					},
				},
				isClusterScope: false,
			},
			wantErr: false,
		},
		{
			name: "base test, cluster scope",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:            "alert",
						EncryptionPassword:   "1234567890123456",
						EncryptionGlobalSalt: "1234567890123456",
					},
				},
				isClusterScope: true,
			},
			wantErr: false,
		},
		{
			name: "encryption password not enough characters",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:          "alert",
						EncryptionPassword: "123456789012345",
					},
				},
				isClusterScope: false,
			},
			wantErr: true,
		},
		{
			name: "encryption global salt not enough characters",
			altSpecConfig: &SpecConfig{
				alert: &alertapi.Alert{
					Spec: alertapi.AlertSpec{
						Namespace:            "alert",
						EncryptionGlobalSalt: "123456789012345",
					},
				},
				isClusterScope: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.altSpecConfig.getAlertSecret()
			if (err != nil) != tt.wantErr {
				t.Errorf("altSpecConfig.GetAlertSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				want := &v1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      util.GetResourceName(tt.altSpecConfig.alert.Name, util.AlertName, "secret"),
						Namespace: tt.altSpecConfig.alert.Spec.Namespace,
						Labels:    map[string]string{"app": util.AlertName, "component": "alert", "name": tt.altSpecConfig.alert.Name},
					},
					Data: map[string][]byte{
						"ALERT_ENCRYPTION_GLOBAL_SALT": []byte(tt.altSpecConfig.alert.Spec.EncryptionGlobalSalt), "ALERT_ENCRYPTION_PASSWORD": []byte(tt.altSpecConfig.alert.Spec.EncryptionPassword),
					},
					Type: "Opaque",
				}
				if !reflect.DeepEqual(got.Secret, want) {
					t.Errorf("altSpecConfig.GetAlertSecret() = %v, want %v", got, want)
				}
			}
		})
	}
}
