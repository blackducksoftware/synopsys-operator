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
package webhook

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	v1beta12 "k8s.io/api/authentication/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ow *OperatorWebhook) serveCustomResource(w http.ResponseWriter, r *http.Request) {
	ow.serve(w, r, ow.blackduckCustomResource)
}

func (ow *OperatorWebhook) blackduckCustomResource(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	bd := v1.Blackduck{}
	err := json.Unmarshal(ar.Request.Object.Raw, &bd)
	if err != nil {
		logrus.Error(err)
		return ow.returnError(err.Error())
	}

	// Get TokenReview for current user
	tok, err := ow.kubeClient.AuthenticationV1beta1().TokenReviews().Create(&v1beta12.TokenReview{Spec: v1beta12.TokenReviewSpec{Token: ow.kubeConfig.BearerToken}})
	if err != nil {
		return ow.returnError(err.Error())
	}

	// Log the request
	logrus.Infof("Resource: %s, Operation: %s\nUsername: %s\n\n", ar.Request.Name, ar.Request.Operation, ar.Request.UserInfo.Username)

	reviewResponse := v1beta1.AdmissionResponse{}

	// Only if the resource is being updated and that request was made by a different user
	if ar.Request.Operation == v1beta1.Update && !strings.EqualFold(tok.Status.User.Username, ar.Request.UserInfo.Username) {
		current, err := ow.blackduckClient.SynopsysV1().Blackducks(corev1.NamespaceDefault).Get(bd.Name, metav1.GetOptions{})
		if err != nil {
			logrus.Error(err)
			return ow.returnError(err.Error())
		}

		if !reflect.DeepEqual(current.Status, bd.Status) {
			logrus.Error(err)
			return ow.returnError("Status cannot be modified")
		}

		if strings.Compare(current.Spec.Namespace, bd.Spec.Namespace) != 0 {
			return ow.returnError("Namespace cannot be modified")
		}

		if bd.Spec.ExternalPostgres.PostgresHost == "" {
			// if external postgres host is not configured, it means we are using internal and should require all internal passwords
			if bd.Spec.AdminPassword == "" || bd.Spec.UserPassword == "" || bd.Spec.PostgresPassword == "" {
				return ow.returnError("For Postgres, adminPassword, userPassword and postgressPassword are required")
			}
		} else {
			// otherwise require all corresponding external postgress fields with the exception of Ssl, as it's set to false by default
			if bd.Spec.ExternalPostgres.PostgresPort == 0 || bd.Spec.ExternalPostgres.PostgresAdmin == "" || bd.Spec.ExternalPostgres.PostgresUser == "" || bd.Spec.ExternalPostgres.PostgresAdminPassword == "" ||
				bd.Spec.ExternalPostgres.PostgresUserPassword == "" {
				return ow.returnError("For external Postgres, host, port, admin, user, adminPassword, and userPassword are required")
			}
		}
	}

	reviewResponse.Allowed = true
	return &reviewResponse
}
