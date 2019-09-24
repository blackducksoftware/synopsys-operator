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
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// getAlertCustomCertSecret creates a Secret Horizon component for the Alert custom certificate
func (a *SpecConfig) getAlertCustomCertSecret() (*components.Secret, error) {
	// create a secret
	certificateSecret := components.NewSecret(horizonapi.SecretConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "certificate"),
		Namespace: a.alert.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	})
	certificateData := make(map[string][]byte, 0)
	if len(a.alert.Spec.Certificate) > 0 && len(a.alert.Spec.CertificateKey) > 0 {
		certificateData["WEBSERVER_CUSTOM_CERT_FILE"] = []byte(a.alert.Spec.Certificate)
		certificateData["WEBSERVER_CUSTOM_KEY_FILE"] = []byte(a.alert.Spec.CertificateKey)
	}

	if len(a.alert.Spec.JavaKeyStore) > 0 {
		certificateData["cacerts"] = []byte(a.alert.Spec.JavaKeyStore)
	}

	certificateSecret.AddData(certificateData)

	certificateSecret.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})
	return certificateSecret, nil
}
