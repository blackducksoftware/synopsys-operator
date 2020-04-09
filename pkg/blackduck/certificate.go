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

package blackduck

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

// CreateSelfSignedCert will create a random self signed certificate
func CreateSelfSignedCert() (string, string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	//Max random value, a 130-bits integer, i.e 2^130 - 1
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))
	template := x509.Certificate{
		SerialNumber: max,
		Subject: pkix.Name{
			CommonName:         "Black Duck",
			OrganizationalUnit: []string{"Cloud Native"},
			Organization:       []string{"Black Duck By Synopsys"},
			Locality:           []string{"Burlington"},
			StreetAddress:      []string{"800 District Avenue"},
			Province:           []string{"Massachusetts"},
			Country:            []string{"US"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 3),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	/*
	   hosts := strings.Split(*host, ",")
	   for _, h := range hosts {
	   	if ip := net.ParseIP(h); ip != nil {
	   		template.IPAddresses = append(template.IPAddresses, ip)
	   	} else {
	   		template.DNSNames = append(template.DNSNames, h)
	   	}
	   }
	   if *isCA {
	   	template.IsCA = true
	   	template.KeyUsage |= x509.KeyUsageCertSign
	   }
	*/

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}
	certificate := &bytes.Buffer{}
	key := &bytes.Buffer{}
	pem.Encode(certificate, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pem.Encode(key, pemBlockForKey(priv))
	return certificate.String(), key.String()
}

func GetCertificateSecretFromFile(secretName, namespace, certPath, keyPath string) (*corev1.Secret, error) {
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	return GetCertificateSecret(secretName, namespace, cert, key)
}

func GetCertificateSecret(secretName string, namespace string, cert []byte, key []byte) (*corev1.Secret, error) {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "blackduck",
				"component": "secret",
				"name":      secretName,
			},
		},
		Data: map[string][]byte{
			"WEBSERVER_CUSTOM_CERT_FILE": cert,
			"WEBSERVER_CUSTOM_KEY_FILE":  key,
		},
		Type: corev1.SecretTypeOpaque,
	}, nil
}

func GetProxyCertificateSecret(secretName string, namespace string, cert []byte) (*corev1.Secret, error) {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "blackduck",
				"component": "secret",
				"name":      secretName,
			},
		},
		Data: map[string][]byte{
			"HUB_PROXY_CERT_FILE": cert,
		},
		Type: corev1.SecretTypeOpaque,
	}, nil
}

func GetAuthCertificateSecret(secretName string, namespace string, cert []byte) (*corev1.Secret, error) {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "blackduck",
				"component": "secret",
				"name":      secretName,
			},
		},
		Data: map[string][]byte{
			"AUTH_CUSTOM_CA": cert,
		},
		Type: corev1.SecretTypeOpaque,
	}, nil
}
