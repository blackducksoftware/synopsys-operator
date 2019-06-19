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

package containers

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
)

// GetSecrets wil return the secrets
func (c *Creater) GetSecrets(cert string, key string) []*components.Secret {
	var secrets []*components.Secret

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webserver-certificate"), Type: horizonapi.SecretTypeOpaque})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(cert), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})
	certificateSecret.AddLabels(c.GetVersionLabel("secret"))
	secrets = append(secrets, certificateSecret)

	if len(c.blackDuck.Spec.ProxyCertificate) > 0 {
		cert, err := c.stringToCertificate(c.blackDuck.Spec.ProxyCertificate)
		if err != nil {
			logrus.Warnf("The proxy certificate provided is invalid")
		} else {
			logrus.Debugf("Adding Proxy certificate with SN: %x", cert.SerialNumber)
			proxyCertificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "proxy-certificate"), Type: horizonapi.SecretTypeOpaque})
			proxyCertificateSecret.AddData(map[string][]byte{"HUB_PROXY_CERT_FILE": []byte(c.blackDuck.Spec.ProxyCertificate)})
			proxyCertificateSecret.AddLabels(c.GetVersionLabel("secret"))
			secrets = append(secrets, proxyCertificateSecret)
		}
	}

	if len(c.blackDuck.Spec.AuthCustomCA) > 0 {
		cert, err := c.stringToCertificate(c.blackDuck.Spec.AuthCustomCA)
		if err != nil {
			logrus.Warnf("The Auth Custom CA provided is invalid")
		} else {
			logrus.Debugf("Adding The Auth Custom CA with SN: %x", cert.SerialNumber)
			authCustomCASecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "auth-custom-ca"), Type: horizonapi.SecretTypeOpaque})
			authCustomCASecret.AddData(map[string][]byte{"AUTH_CUSTOM_CA": []byte(c.blackDuck.Spec.AuthCustomCA)})
			authCustomCASecret.AddLabels(c.GetVersionLabel("secret"))
			secrets = append(secrets, authCustomCASecret)
		}
	}

	return secrets
}

func (c *Creater) stringToCertificate(certificate string) (*x509.Certificate, error) {
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
