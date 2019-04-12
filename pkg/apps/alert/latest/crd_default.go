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

package alert

import (
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
)

// GetDefault creates an Alert crd configuration object with defaults
func (hc *Creater) GetDefault(alt *alertapi.AlertSpec) *alertapi.AlertSpec {
	port := 8443
	standAlone := true

	return &alertapi.AlertSpec{
		Namespace:            "alert-test",
		Version:              "3.1.0",
		AlertImage:           "docker.io/blackducksoftware/blackduck-alert:3.1.0",
		CfsslImage:           "docker.io/blackducksoftware/blackduck-cfssl:1.0.0",
		ExposeService:        "NODEPORT",
		Port:                 &port,
		EncryptionPassword:   "",
		EncryptionGlobalSalt: "",
		PersistentStorage:    true,
		PVCName:              "alert-pvc",
		StandAlone:           &standAlone,
		PVCSize:              "5G",
		PVCStorageClass:      "",
		AlertMemory:          "2560M",
		CfsslMemory:          "640M",
		Environs: []string{
			"ALERT_SERVER_PORT:8443",
			"PUBLIC_HUB_WEBSERVER_HOST:localhost",
			"PUBLIC_HUB_WEBSERVER_PORT:443",
		},
	}
}
