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

package blackduck

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
)

func OperatorAffinityTok8sAffinity(opAffinity []v1.NodeAffinity) corev1.Affinity {
	var hardTerms, softTerms []corev1.NodeSelectorTerm
	for _, aValue := range opAffinity {
		if strings.EqualFold(aValue.AffinityType, "hard") {
			hardTerms = append(hardTerms, corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      aValue.Key,
						Values:   aValue.Values,
						Operator: corev1.NodeSelectorOperator(aValue.Op),
					},
				},
			})
		} else if strings.EqualFold(aValue.AffinityType, "soft") {
			softTerms = append(softTerms, corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      aValue.Key,
						Values:   aValue.Values,
						Operator: corev1.NodeSelectorOperator(aValue.Op),
					},
				},
			})
		}
	}

	var af corev1.Affinity
	if len(hardTerms) > 0 || len(softTerms) > 0 {
		af = corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}
		if len(hardTerms) > 0 {
			af.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{NodeSelectorTerms: hardTerms}
		}
		if len(softTerms) > 0 {
			for _, s := range softTerms {
				af.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(af.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, corev1.PreferredSchedulingTerm{
					Weight:     100,
					Preference: s,
				})
			}
		}
	}

	return af
}

func OperatorSecurityContextTok8sAffinity(opSecurityContext api.SecurityContext) corev1.PodSecurityContext {
	return corev1.PodSecurityContext{
		FSGroup:    opSecurityContext.FsGroup,
		RunAsUser:  opSecurityContext.RunAsUser,
		RunAsGroup: opSecurityContext.RunAsGroup,
	}
}

func GetCertsFromFlagsAndSetHelmValue(name string, namespace string, flagset *pflag.FlagSet, helmVal map[string]interface{}) ([]corev1.Secret, error) {
	var objects []corev1.Secret
	if flagset.Lookup("certificate-file-path").Changed && flagset.Lookup("certificate-key-file-path").Changed {
		certPath := flagset.Lookup("certificate-file-path").Value.String()
		keyPath := flagset.Lookup("certificate-key-file-path").Value.String()
		secretName := fmt.Sprintf("%s-blackduck-webserver-certificate", name)

		secret, err := GetCertificateSecretFromFile(secretName, namespace, certPath, keyPath)
		if err != nil {
			return nil, err
		}
		util.SetHelmValueInMap(helmVal, []string{"tlsCertSecretName"}, secretName)
		objects = append(objects, *secret)
	}

	if flagset.Lookup("proxy-certificate-file-path").Changed {
		certPath := flagset.Lookup("proxy-certificate-file-path").Value.String()
		secretName := fmt.Sprintf("%s-blackduck-proxy-certificate", name)

		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return nil, err
		}

		secret, err := GetProxyCertificateSecret(secretName, namespace, cert)
		if err != nil {
			return nil, err
		}
		util.SetHelmValueInMap(helmVal, []string{"proxyCertSecretName"}, secretName)
		objects = append(objects, *secret)
	}

	if flagset.Lookup("auth-custom-ca-file-path").Changed {
		certPath := flagset.Lookup("auth-custom-ca-file-path").Value.String()
		secretName := fmt.Sprintf("%s-blackduck-auth-custom-ca", name)

		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return nil, err
		}

		secret, err := GetAuthCertificateSecret(secretName, namespace, cert)
		if err != nil {
			return nil, err
		}
		util.SetHelmValueInMap(helmVal, []string{"certAuthCACertSecretName"}, secretName)
		objects = append(objects, *secret)
	}

	return objects, nil
}
