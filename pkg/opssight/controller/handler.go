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
	Config            *rest.Config
	Clientset         *kubernetes.Clientset
	OpsSightClientset *opssightclientset.Clientset
	Namespace         string
	CmMutex           chan bool
}

// ObjectCreated will be called for create opssight events
func (h *OpsSightHandler) ObjectCreated(obj interface{}) {
	log.Debugf("ObjectCreated: %+v", obj)
	opsSightv1 := obj.(*opssight_v1.OpsSight)
	if strings.EqualFold(opsSightv1.Spec.State, "") {
		// Update status
		opsSightv1.Spec.State = "pending"
		opsSightv1.Status.State = "creating"
		_, err := h.updateHubObject(opsSightv1)
		if err != nil {
			log.Errorf("Couldn't update Alert object: %s", err.Error())
		}
	}
}

// ObjectDeleted will be called for delete opssight events
func (h *OpsSightHandler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %+v", name)

	//Set spec/state  and status/state to started
	// obj.Spec.State = "deleted"
	// obj.Status.State = "deleted"
	// obj, err = h.updateHubObject(obj)
	// if err != nil {
	// 	log.Errorf("Couldn't update Hub object: %s", err.Error())
	// }
}

// ObjectUpdated will be called for update opssight events
func (h *OpsSightHandler) ObjectUpdated(objOld, objNew interface{}) {
	//if strings.Compare(objOld.Spec.State, objNew.Spec.State) != 0 {
	//	log.Infof("%s - Changing state [%s] -> [%s] | Current: [%s]", objNew.Name, objOld.Spec.State, objNew.Spec.State, objNew.Status.State )
	//	// TO DO
	//	objNew.Status.State = objNew.Spec.State
	//	h.hubClientset.SynopsysV1().Hubs(objNew.Namespace).Update(objNew)
	//}
}

func (h *OpsSightHandler) updateHubObject(obj *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	return h.OpsSightClientset.SynopsysV1().OpsSights(h.Namespace).Update(obj)
}
