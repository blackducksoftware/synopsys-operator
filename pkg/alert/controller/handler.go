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

	"github.com/blackducksoftware/perceptor-protoform/pkg/alert"
	alertclientset "github.com/blackducksoftware/perceptor-protoform/pkg/alert/client/clientset/versioned"
	alert_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
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

// AlertHandler will store the configuration that is required to initiantiate the informers callback
type AlertHandler struct {
	Config         *rest.Config
	Clientset      *kubernetes.Clientset
	AlertClientset *alertclientset.Clientset
	Namespace      string
	CmMutex        chan bool
}

// ObjectCreated will be called for create alert events
func (h *AlertHandler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	alertv1 := obj.(*alert_v1.Alert)
	if strings.EqualFold(alertv1.Spec.State, "") {
		// Update status
		alertv1.Spec.State = "pending"
		alertv1.Status.State = "creating"
		_, err := h.updateHubObject(alertv1)
		if err != nil {
			log.Errorf("Couldn't update Alert object: %s", err.Error())
		}

		alertCreator := alert.NewCreater(h.Config, h.Clientset, h.AlertClientset)
		if err != nil {
			log.Errorf("unable to create the new hub creater for %s due to %+v", alertv1.Name, err)
		}
		err = alertCreator.CreateAlert(alertv1)

		if err != nil {
			//Set spec/state  and status/state to started
			alertv1.Spec.State = "error"
			alertv1.Status.State = "error"
		} else {
			alertv1.Spec.State = "running"
			alertv1.Status.State = "running"
		}
	}
}

// ObjectDeleted will be called for delete alert events
func (h *AlertHandler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)
	alertCreator := alert.NewCreater(h.Config, h.Clientset, h.AlertClientset)
	alertCreator.DeleteAlert(name)
}

// ObjectUpdated will be called for update alert events
func (h *AlertHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *AlertHandler) updateHubObject(obj *alert_v1.Alert) (*alert_v1.Alert, error) {
	return h.AlertClientset.SynopsysV1().Alerts(h.Namespace).Update(obj)
}
