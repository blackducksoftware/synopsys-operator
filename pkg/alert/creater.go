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

package alert

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetAlertCustomCertificateSecret ...
func GetAlertCustomCertificateSecret(namespace, secretName, customCertificate, customCertificateKey string) (map[string]runtime.Object, error) {
	mapOfUniqueIDToBaseRuntimeObject := make(map[string]runtime.Object, 0)

	mapOfUniqueIDToBaseRuntimeObject["Secret.customCertificate"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"WEBSERVER_CUSTOM_CERT_FILE": []byte(customCertificate),
			"WEBSERVER_CUSTOM_KEY_FILE":  []byte(customCertificateKey),
		},
		Type: corev1.SecretTypeOpaque,
	}

	return mapOfUniqueIDToBaseRuntimeObject, nil
}

// GetAlertJavaKeystoreSecret ...
func GetAlertJavaKeystoreSecret(namespace, secretName, javaKeystore string) (map[string]runtime.Object, error) {
	mapOfUniqueIDToBaseRuntimeObject := make(map[string]runtime.Object, 0)

	mapOfUniqueIDToBaseRuntimeObject["Secret.customCertificate"] = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"cacerts": []byte(javaKeystore),
		},
		Type: corev1.SecretTypeOpaque,
	}

	return mapOfUniqueIDToBaseRuntimeObject, nil
}
