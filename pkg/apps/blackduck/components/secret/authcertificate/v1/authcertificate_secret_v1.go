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

package v1

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// BdSecret holds the Black Duck secret configuration
type BdSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.BlackDuckAuthCertificateSecretV1, NewBdSecret)
}

// NewBdSecret returns the Black Duck secret configuration
func NewBdSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.SecretInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdSecret{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetSecret returns the secret
func (b *BdSecret) GetSecret() (*components.Secret, error) {
	if len(b.blackDuck.Spec.AuthCustomCA) > 0 {
		cert, err := stringToCertificate(b.blackDuck.Spec.AuthCustomCA)
		if err != nil {
			return nil, fmt.Errorf("unable to convert the string to auth certificate due to %+v", err)
		}

		logrus.Debugf("Adding The Auth Custom CA with SN: %x", cert.SerialNumber)
		authCustomCASecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "auth-custom-ca"), Type: horizonapi.SecretTypeOpaque})
		authCustomCASecret.AddData(map[string][]byte{"AUTH_CUSTOM_CA": []byte(b.blackDuck.Spec.AuthCustomCA)})
		authCustomCASecret.AddLabels(apputils.GetVersionLabel("secret", b.blackDuck.Name, b.blackDuck.Spec.Version))
		return authCustomCASecret, nil
	}
	return nil, nil
}

func stringToCertificate(certificate string) (*x509.Certificate, error) {
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
