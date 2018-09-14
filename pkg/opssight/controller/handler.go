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
	"strings"

	opssight_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/opssight"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Handler interface contains the methods that are required
type Handler interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// OpsSightHandler will store the configuration that is required to initiantiate the informers callback
type OpsSightHandler struct {
	Config            *model.Config
	KubeConfig        *rest.Config
	Clientset         *kubernetes.Clientset
	OpsSightClientset *opssightclientset.Clientset
	Namespace         string
	CmMutex           chan bool
}

// ObjectCreated will be called for create opssight events
func (h *OpsSightHandler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	opssightv1 := obj.(*opssight_v1.OpsSight)
	if strings.EqualFold(opssightv1.Spec.State, "") {
		// Update status
		opssightv1.Spec.State = "pending"
		opssightv1.Status.State = "creating"
		_, err := h.updateHubObject(opssightv1)
		if err != nil {
			log.Errorf("Couldn't update Alert object: %s", err.Error())
		}

		opssightCreator := opssight.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset)
		if err != nil {
			log.Errorf("unable to create the new hub creater for %s due to %+v", opssightv1.Name, err)
		}
		err = opssightCreator.CreateOpsSight(opssightv1)

		if err != nil {
			//Set spec/state  and status/state to started
			opssightv1.Spec.State = "error"
			opssightv1.Status.State = "error"
		} else {
			opssightv1.Spec.State = "running"
			opssightv1.Status.State = "running"
		}
	}
}

// ObjectDeleted will be called for delete opssight events
func (h *OpsSightHandler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)
	opssightCreator := opssight.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset)
	opssightCreator.DeleteOpsSight(name)
}

// ObjectUpdated will be called for update opssight events
func (h *OpsSightHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *OpsSightHandler) updateHubObject(obj *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	return h.OpsSightClientset.SynopsysV1().OpsSights(h.Namespace).Update(obj)
}
