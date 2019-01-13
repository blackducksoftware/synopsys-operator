/*
Copyright (C) 2018 Synopsys, Inc.

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

package hub

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/hub/v2"
	"github.com/sirupsen/logrus"
)

func (hc *Creater) createHubSecrets(createHub *v2.HubSpec, adminPassword string, userPassword string) []*components.Secret {
	var secrets []*components.Secret
	hubSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: createHub.Namespace, Name: "db-creds", Type: horizonapi.SecretTypeOpaque})

	if createHub.ExternalPostgres != (v2.PostgresExternalDBConfig{}) {
		hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": createHub.ExternalPostgres.PostgresAdminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": createHub.ExternalPostgres.PostgresUserPassword})
	} else {
		hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": adminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": userPassword})
	}
	secrets = append(secrets, hubSecret)

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: createHub.Namespace, Name: "blackduck-certificate", Type: horizonapi.SecretTypeOpaque})
	if strings.EqualFold(createHub.CertificateName, "manual") {
		certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(createHub.Certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(createHub.CertificateKey)})
	}
	secrets = append(secrets, certificateSecret)

	if len(createHub.ProxyCertificate) > 0 {
		cert, err := hc.stringToCertificate(createHub.ProxyCertificate)
		if err != nil {
			logrus.Warnf("The proxy certificate provided is invalid")
		} else {
			logrus.Debugf("Adding Proxy certificate with SN: %x", cert.SerialNumber)
			proxyCertificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: createHub.Namespace, Name: "blackduck-proxy-certificate", Type: horizonapi.SecretTypeOpaque})
			proxyCertificateSecret.AddData(map[string][]byte{"HUB_PROXY_CERT_FILE": []byte(createHub.ProxyCertificate)})
			secrets = append(secrets, proxyCertificateSecret)
		}
	}
	return secrets
}

func (hc *Creater) stringToCertificate(certificate string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
