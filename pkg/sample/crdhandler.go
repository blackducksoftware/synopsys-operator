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
	"fmt"
	"strings"

	samplev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sampleclientset "github.com/blackducksoftware/synopsys-operator/pkg/sample/client/clientset/versioned"
	"github.com/imdario/mergo"
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
	defaults     *samplev1.SampleSpec
}

// ObjectCreated will be called for create sample events.
// It casts the received object to a Sample and attempts
// to create it with a Creater type.
func (handler *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("Handler's ObjectCreated received: %+v", obj)
	sampleObject, ok := obj.(*samplev1.Sample)
	if !ok {
		log.Error("Handler is unable to cast the object to a Sample")
		return
	}
	log.Debugf("Sample Spec State: '%v'", sampleObject.Spec.State)
	if strings.EqualFold(sampleObject.Spec.State, "") {
		// Merge the Default Spec into the Sample Spec
		newSpec := sampleObject.Spec
		sampleDefaultSpec := handler.defaults
		err := mergo.Merge(&newSpec, sampleDefaultSpec)
		log.Debugf("merged sample details %+v", newSpec)
		if err != nil {
			log.Errorf("unable to merge the sample structs for %s due to %+v", sampleObject.Name, err)
			handler.updateState("error", "error", fmt.Sprintf("unable to merge the sample structs for %s due to %+v", sampleObject.Name, err), sampleObject)
			return
		}
		sampleObject.Spec = newSpec

		// Update the status
		sampleObject, err := handler.updateState("pending", "creating", "", sampleObject)
		if err != nil {
			return
		}

		// Create a Sample instance
		sampleCreator := NewSampleCreater(handler.kubeConfig, handler.kubeClient, handler.sampleClient)
		err = sampleCreator.CreateSample(&sampleObject.Spec)
		if err != nil {
			handler.updateState("error", "error", fmt.Sprintf("%+v", err), sampleObject)
			return
		}
		handler.updateState("running", "running", "", sampleObject)
	}
}

// ObjectDeleted will be called for delete sample events
func (handler *Handler) ObjectDeleted(name string) {
	log.Debugf("Handler's ObjectDeleted received: %+v", name)
	sampleCreator := NewSampleCreater(handler.kubeConfig, handler.kubeClient, handler.sampleClient)
	sampleCreator.DeleteSample(name)
}

// ObjectUpdated will be called for update sample events
func (handler *Handler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("Hanlder's ObjectUpdated received: %+v", objNew)
}

// updateState changes the state of the Sample object
func (handler *Handler) updateState(specState string, statusState string, errorMessage string, sample *samplev1.Sample) (*samplev1.Sample, error) {
	sample.Status.State = statusState
	sample.Status.ErrorMessage = errorMessage
	sample, err := handler.updateSampleObject(sample)
	if err != nil {
		log.Errorf("Couldn't update the state of the Sample object: %s", err.Error())
	}
	return sample, err
}

func (handler *Handler) updateSampleObject(obj *samplev1.Sample) (*samplev1.Sample, error) {
	return handler.sampleClient.SynopsysV1().Samples(handler.config.Namespace).Update(obj)
}
