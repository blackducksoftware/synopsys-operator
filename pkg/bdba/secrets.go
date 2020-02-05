/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package bdba

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	math_rand "math/rand"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GenerateRandomString will generate a string of given length
func GenerateRandomString(n int) (string, error) {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return "", err
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	var letterRunes = []rune("_-!0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	randomString := make([]rune, n)
	for i := range randomString {
		r := math_rand.Intn(len(letterRunes))
		randomString[i] = letterRunes[r]
	}
	return string(randomString), nil
}

// GenerateRandomAWSAccessKey will generate a random AWS key
func GenerateRandomAWSAccessKey() (string, error) {
	const n = 14
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		return "", err
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	var letterRunes = []rune("234567ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var fifthRunes = []rune("IJ")
	var lastRunes = []rune("AQ")
	randomString := make([]rune, n)
	for i := range randomString {
		r := math_rand.Intn(len(letterRunes))
		randomString[i] = letterRunes[r]
	}
	var randomFifth = fifthRunes[math_rand.Intn(len(fifthRunes))]
	var randomLast = lastRunes[math_rand.Intn(len(lastRunes))]
	return "AKIA" + string(randomFifth) + string(randomString) + string(randomLast), nil
}

// GetExistingSecret will get an existing deployed secret with a given name
func GetExistingSecret(clientset *kubernetes.Clientset, nameSpace string, secretName string) *corev1.Secret {
	if clientset == nil {
		return nil
	}

	secrets, _ := clientset.CoreV1().Secrets(nameSpace).List(v1.ListOptions{})
	for _, v := range secrets.Items {
		if v.Name == secretName {
			return &v
		}
	}
	return nil
}

// GetSecret will get secret with a given name from RuntimeObjectPatcher
func (p *RuntimeObjectPatcher) GetSecret(secretName string) *corev1.Secret {
	for k, v := range p.mapOfUniqueIDToRuntimeObject {
		switch v.(type) {
		case *corev1.Secret:
			var secret = p.mapOfUniqueIDToRuntimeObject[k].(*corev1.Secret)
			if strings.Contains(secret.Name, secretName) {
				return secret
			}
		}
	}
	return nil
}

// CreateKubeClient creates a kubernetes client from given config file
func CreateKubeClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, err
}
