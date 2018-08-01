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

	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Handler will have the methods related to infromers callback
type Handler interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// HubHandler will store the configuration that is required to initiantiate the informers callback
type HubHandler struct {
	config       *rest.Config
	clientset    *kubernetes.Clientset
	hubClientset *hubclientset.Clientset
	crdNamespace string
}

// ObjectCreated will be called for create hub events
func (h *HubHandler) ObjectCreated(obj *hubv1.Hub) {
	log.Debugf("ObjectCreated: %+v", obj)
	if strings.EqualFold(obj.Spec.State, "pending") {
		// Update status
		obj.Status.State = "creating"
		obj, err := h.updateHubObject(obj)
		if err != nil {
			log.Errorf("Couldn't update Hub object: %s", err.Error())
		}

		hubCreator, err := hub.NewCreater(h.config, h.clientset, h.hubClientset)
		if err != nil {
			log.Errorf("unable to create the new hub creater for %s due to %+v", obj.Name, err)
		}
		ip, err := hubCreator.CreateHub(obj)

		if err != nil {
			//Set spec/state  and status/state to started
			obj.Spec.State = "error"
			obj.Status.State = "error"
		} else {
			obj.Spec.State = "running"
			obj.Status.State = "running"
		}
		obj.Status.IP = ip
		obj, err = h.updateHubObject(obj)
		if err != nil {
			log.Errorf("Couldn't update Hub object: %s", err.Error())
		}
	}
}

// ObjectDeleted will be called for delete hub events
func (h *HubHandler) ObjectDeleted(obj *hubv1.Hub) {
	log.Debugf("ObjectDeleted: %+v", obj)

	hubCreator, err := hub.NewCreater(h.config, h.clientset, h.hubClientset)
	if err != nil {
		log.Errorf("unable to create the new hub creater for %s due to %+v", obj.Name, err)
	}
	hubCreator.DeleteHub(obj.Name)

	//Set spec/state  and status/state to started
	// obj.Spec.State = "deleted"
	// obj.Status.State = "deleted"
	// obj, err = h.updateHubObject(obj)
	// if err != nil {
	// 	log.Errorf("Couldn't update Hub object: %s", err.Error())
	// }
}

// ObjectUpdated will be called for update hub events
func (h *HubHandler) ObjectUpdated(objOld *hubv1.Hub, objNew *hubv1.Hub) {
	//if strings.Compare(objOld.Spec.State, objNew.Spec.State) != 0 {
	//	log.Infof("%s - Changing state [%s] -> [%s] | Current: [%s]", objNew.Name, objOld.Spec.State, objNew.Spec.State, objNew.Status.State )
	//	// TO DO
	//	objNew.Status.State = objNew.Spec.State
	//	h.hubClientset.SynopsysV1().Hubs(objNew.Namespace).Update(objNew)
	//}
}

func (h *HubHandler) updateHubObject(obj *hubv1.Hub) (*hubv1.Hub, error) {
	return h.hubClientset.SynopsysV1().Hubs(h.crdNamespace).Update(obj)
}
