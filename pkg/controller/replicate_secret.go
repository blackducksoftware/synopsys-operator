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

package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	// hub_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	log "github.com/sirupsen/logrus"
)

type SecretReplicator struct {
	client     *kubernetes.Clientset
	hubClient  *hubclientset.Clientset
	namespace  string
	store      cache.Store
	controller cache.Controller

	dependencyMap map[string][]string
}

// Annotations that are used to control this controller's behaviour
const (
	ReplicateFromAnnotation         = "replicator.v1.mittwald.de/replicate-from"
	ReplicatedAtAnnotation          = "replicator.v1.mittwald.de/replicated-at"
	ReplicatedFromVersionAnnotation = "replicator.v1.mittwald.de/replicated-from-version"
)

// JSONPatchOperation is a struct that defines PATCH operations on
// a JSON structure.
type JSONPatchOperation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value,omitempty"`
}

// NewSecretReplicator creates a new secret replicator
func NewSecretReplicator(client *kubernetes.Clientset, hubClient *hubclientset.Clientset, namespace string, resyncPeriod time.Duration) *SecretReplicator {
	repl := SecretReplicator{
		client:        client,
		hubClient:     hubClient,
		namespace:     namespace,
		dependencyMap: make(map[string][]string),
	}

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets(v1.NamespaceAll).List(opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Secrets(v1.NamespaceAll).Watch(opts)
			},
		},
		&v1.Secret{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    repl.secretAdded,
			UpdateFunc: func(old interface{}, new interface{}) { repl.secretAdded(new) },
			DeleteFunc: repl.secretDeleted,
		},
	)

	repl.store = store
	repl.controller = controller

	return &repl
}

func (r *SecretReplicator) Run() {
	log.Printf("running secret controller")
	r.controller.Run(wait.NeverStop)
}

func (r *SecretReplicator) secretAdded(obj interface{}) {
	secret := obj.(*v1.Secret)

	hubList, err := hub.ListHubs(r.hubClient, r.namespace)
	if err != nil {
		log.Errorf("unable to list the hubs due to %+v", err)
	}

	if strings.EqualFold(secret.Name, "hub-certificate") {
		for _, hub := range hubList.Items {
			if strings.EqualFold(secret.Namespace, hub.Name) {
				if !strings.EqualFold(hub.Spec.CertificateName, "Custom") {
					replicas := r.dependencyMap[hub.Spec.CertificateName]
					replicas = append(replicas, secret.Namespace)
					r.dependencyMap[hub.Spec.CertificateName] = replicas
				} else {
					return
				}
			}
		}
	} else {
		return
	}

	replicas, ok := r.dependencyMap["secretKey"]
	if ok {
		log.Printf("secret %s has %d dependents", "secretKey", len(replicas))
		r.updateDependents(secret, replicas)
	}

	val, ok := secret.Annotations[ReplicateFromAnnotation]
	if !ok {
		return
	}

	log.Printf("secret %s/%s is replicated from %s", secret.Namespace, secret.Name, val)
	v := strings.SplitN(val, "/", 2)

	if len(v) < 2 {
		return
	}

	sourceObject, exists, err := r.store.GetByKey(val)
	if err != nil {
		log.Printf("could not get secret %s: %s", val, err)
		return
	} else if !exists {
		log.Printf("could not get secret %s: does not exist", val)
		return
	}

	if _, ok := r.dependencyMap[val]; !ok {
		r.dependencyMap[val] = make([]string, 0, 1)
	}

	r.dependencyMap[val] = append(r.dependencyMap[val], "secretKey")

	sourceSecret := sourceObject.(*v1.Secret)

	r.replicateSecret(secret, sourceSecret)
}

func (r *SecretReplicator) replicateSecret(secret *v1.Secret, sourceSecret *v1.Secret) error {
	targetVersion, ok := secret.Annotations[ReplicatedFromVersionAnnotation]
	sourceVersion := sourceSecret.ResourceVersion

	if ok && targetVersion == sourceVersion {
		log.Printf("secret %s/%s is already up-to-date", secret.Namespace, secret.Name)
		return nil
	}

	secretCopy := secret.DeepCopy()

	if secretCopy.Data == nil {
		secretCopy.Data = make(map[string][]byte)
	}

	for key, value := range sourceSecret.Data {
		newValue := make([]byte, len(value))
		copy(newValue, value)
		secretCopy.Data[key] = newValue
	}

	log.Printf("updating secret %s/%s", secret.Namespace, secret.Name)

	secretCopy.Annotations[ReplicatedAtAnnotation] = time.Now().Format(time.RFC3339)
	secretCopy.Annotations[ReplicatedFromVersionAnnotation] = sourceSecret.ResourceVersion

	s, err := r.client.CoreV1().Secrets(secret.Namespace).Update(secretCopy)
	if err != nil {
		return err
	}

	r.store.Update(s)
	return nil
}

func (r *SecretReplicator) secretFromStore(key string) (*v1.Secret, error) {
	obj, exists, err := r.store.GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("could not get secret %s: %s", key, err)
	}

	if !exists {
		return nil, fmt.Errorf("could not get secret %s: does not exist", key)
	}

	secret, ok := obj.(*v1.Secret)
	if !ok {
		return nil, fmt.Errorf("bad type returned from store: %T", obj)
	}

	return secret, nil
}

func (r *SecretReplicator) updateDependents(secret *v1.Secret, dependents []string) error {
	for _, dependentKey := range dependents {
		log.Printf("updating dependent secret %s/%s -> %s", secret.Namespace, secret.Name, dependentKey)

		targetObject, exists, err := r.store.GetByKey(dependentKey)
		if err != nil {
			log.Printf("could not get dependent secret %s: %s", dependentKey, err)
			continue
		} else if !exists {
			log.Printf("could not get dependent secret %s: does not exist", dependentKey)
			continue
		}

		targetSecret := targetObject.(*v1.Secret)

		r.replicateSecret(targetSecret, secret)
	}

	return nil
}

func (r *SecretReplicator) secretDeleted(obj interface{}) {
	secret := obj.(*v1.Secret)
	secretKey := fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)

	replicas, ok := r.dependencyMap[secretKey]
	if !ok {
		log.Printf("secret %s has no dependents and can be deleted without issues", secretKey)
		return
	}

	for _, dependentKey := range replicas {
		targetSecret, err := r.secretFromStore(dependentKey)
		if err != nil {
			log.Printf("could not load dependent secret: %s", err)
			continue
		}

		patch := []JSONPatchOperation{{Operation: "remove", Path: "/data"}}
		patchBody, err := json.Marshal(&patch)

		if err != nil {
			log.Printf("error while building patch body for secret %s: %s", dependentKey, err)
			continue
		}

		log.Printf("clearing dependent secret %s", dependentKey)
		log.Printf("patch body: %s", string(patchBody))

		s, err := r.client.CoreV1().Secrets(targetSecret.Namespace).Patch(targetSecret.Name, types.JSONPatchType, patchBody)
		if err != nil {
			log.Printf("error while patching secret %s: %s", dependentKey, err)
			continue
		}

		r.store.Update(s)
	}
}
