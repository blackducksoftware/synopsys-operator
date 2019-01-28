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

package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	blackduckClient "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetHubVersion will return the Hub version from the list of Hub environment variables
func GetHubVersion(environs []string) string {
	for _, value := range environs {
		if strings.Contains(value, "HUB_VERSION") {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				return strings.Trim(values[1], " ")
			}
			break
		}
	}

	return ""
}

// GetDefaultPasswords returns admin,user,postgres passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (adminPassword string, userPassword string, postgresPassword string, err error) {
	blackduckSecret, err := util.GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		logrus.Infof("warning: You need to first create a 'blackduck-secret' in this namespace with ADMIN_PASSWORD, USER_PASSWORD, POSTGRES_PASSWORD")
		return "", "", "", err
	}
	adminPassword = string(blackduckSecret.Data["ADMIN_PASSWORD"])
	userPassword = string(blackduckSecret.Data["USER_PASSWORD"])
	postgresPassword = string(blackduckSecret.Data["POSTGRES_PASSWORD"])

	// default named return
	return adminPassword, userPassword, postgresPassword, err
}

func updateBlackduckObject(h *blackduckClient.Clientset, namespace string, obj *v1.Blackduck) (*v1.Blackduck, error) {
	return h.SynopsysV1().Blackducks(namespace).Update(obj)
}

// UpdateState will be used to update the hub object
func UpdateState(h *blackduckClient.Clientset, namespace string, specState string, statusState string, err error, blackduck *v1.Blackduck) (*v1.Blackduck, error) {
	blackduck.Spec.State = specState
	blackduck.Status.State = statusState
	if err != nil {
		blackduck.Status.ErrorMessage = fmt.Sprintf("%+v", err)
	}
	blackduck, err = updateBlackduckObject(h, namespace, blackduck)
	if err != nil {
		logrus.Errorf("couldn't update the state of blackduck object: %s", err.Error())
	}
	return blackduck, err
}

// GetHubDBPassword will retrieve the blackduck and blackduck_user db password
func GetHubDBPassword(kubeClient *kubernetes.Clientset, namespace string) (string, string, error) {
	var userPw, adminPw string

	secret, err := util.GetSecret(kubeClient, namespace, "db-creds")
	if err != nil {
		return userPw, adminPw, err
	}

	s, ok := secret.Data["HUB_POSTGRES_USER_PASSWORD_FILE"]
	if !ok {
		return "", "", fmt.Errorf("HUB_POSTGRES_USER_PASSWORD_FILE is missing")
	}
	userPw = string(s)

	s, ok = secret.Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"]
	if !ok {
		return "", "", fmt.Errorf("HUB_POSTGRES_ADMIN_PASSWORD_FILE is missing")
	}
	adminPw = string(s)
	return userPw, adminPw, nil
}

// CloneJob create a Kube job to clone a postgres instance
func CloneJob(clientset *kubernetes.Clientset, namespace string, from string, to string, password string) error {
	command := fmt.Sprintf("pg_dumpall -h postgres.%s.svc.cluster.local -U postgres | psql -h postgres.%s.svc.cluster.local -U postgres", from, to)

	cloneJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("clone-job-%s", to),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "clone",
							Image:   "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
							Command: []string{"/bin/bash"},
							Args: []string{
								"-c",
								command,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PGPASSWORD",
									Value: password,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	job, err := clientset.BatchV1().Jobs(namespace).Create(cloneJob)
	if err != nil {
		return err
	}

	timeout := time.After(30 * time.Minute)
	tick := time.Tick(10 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("the clone operation timed out")

		case <-tick:
			job, err = clientset.BatchV1().Jobs(job.Namespace).Get(job.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if job.Status.Succeeded > 0 {
				//clientset.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{})
				return nil
			}
		}
	}
}
