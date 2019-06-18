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
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// getAlertSecret creates a Secret Horizon component for the Alert
func (a *SpecConfig) getAlertSecret() (*components.Secret, error) {
	// Check Secret Values
	encryptPassLength := len(a.alert.Spec.EncryptionPassword)
	if encryptPassLength > 0 && encryptPassLength < 16 {
		return nil, fmt.Errorf("encryption password is %d characters, it must be 16 or more", encryptPassLength)
	}
	encryptGlobalSaltLength := len(a.alert.Spec.EncryptionGlobalSalt)
	if encryptGlobalSaltLength > 0 && encryptGlobalSaltLength < 16 {
		return nil, fmt.Errorf("encryption global salt is %d characters, it must be 16 or more", encryptGlobalSaltLength)
	}

	// create a secret
	alertSecret := components.NewSecret(horizonapi.SecretConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "secret"),
		Namespace: a.alert.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	})
	alertSecret.AddData(map[string][]byte{
		"ALERT_ENCRYPTION_PASSWORD":    []byte(a.alert.Spec.EncryptionPassword),
		"ALERT_ENCRYPTION_GLOBAL_SALT": []byte(a.alert.Spec.EncryptionGlobalSalt),
	})

	alertSecret.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	return alertSecret, nil

}
