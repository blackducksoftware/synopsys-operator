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

package sample

import (
	sample_v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sampleclientset "github.com/blackducksoftware/synopsys-operator/pkg/sample/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface contains the methods that are required
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	config       *protoform.Config
	kubeConfig   *rest.Config
	kubeClient   *kubernetes.Clientset
	sampleClient *sampleclientset.Clientset
	defaults     *sample_v1.SampleSpec
}

// ObjectCreated will be called for create sample events
func (h *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("A Sample is being Created: %+v", obj)
}

// ObjectDeleted will be called for delete sample events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("A Sample is being Deleted: %+v", name)
}

// ObjectUpdated will be called for update sample events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("A Sample is being Updated: %+v", objNew)
}
