/*
Copyright (C) 2020 Synopsys, Inc.

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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// GeneratePemSelfSignedCertificateAndKey returns a self-signed certificate and its key
func GeneratePemSelfSignedCertificateAndKey(name pkix.Name) (string, string, error) {

	if len(name.CommonName) == 0 {
		return "", "", fmt.Errorf("the CommonName cannot be empty")
	}

	// Generate the serial number for the certificate
	sn, err := genx509SerialNumber()
	if err != nil {
		return "", "", err
	}

	template := &x509.Certificate{
		SerialNumber:          sn,
		Subject:               name,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Generate KEY
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	pub := &priv.PublicKey

	// Create the certificate
	cert, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		return "", "", err
	}

	// Encode certificate
	certOut := &bytes.Buffer{}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	// Encode private key
	keyOut := &bytes.Buffer{}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certOut.String(), keyOut.String(), nil
}

func genx509SerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
}
