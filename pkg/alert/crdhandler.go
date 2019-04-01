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

package alert

import (
	"fmt"
	"strings"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
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
	config      *protoform.Config
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	alertClient *alertclientset.Clientset
	defaults    *alertapi.AlertSpec
	routeClient *routeclient.RouteV1Client
}

// NewHandler will create the handler
func NewHandler(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, alertClient *alertclientset.Clientset, routeClient *routeclient.RouteV1Client, defaults *alertapi.AlertSpec) *Handler {
	return &Handler{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, alertClient: alertClient, routeClient: routeClient, defaults: defaults}
}

// ObjectCreated will be called for create alert events
func (h *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	alertv1, ok := obj.(*alertapi.Alert)
	if !ok {
		log.Error("Unable to cast to Alert object")
		return
	}
	if strings.EqualFold(alertv1.Status.State, "") {
		// merge with default values
		newSpec := alertv1.Spec
		alertDefaultSpec := h.defaults
		err := mergo.Merge(&newSpec, alertDefaultSpec)
		log.Debugf("merged alert details %+v", newSpec)
		if err != nil {
			log.Errorf("unable to merge the alert structs for %s due to %+v", alertv1.Name, err)
			//Set spec/state  and status/state to started
			h.updateState("error", fmt.Sprintf("unable to merge the alert structs for %s due to %+v", alertv1.Name, err), alertv1)
		} else {
			alertv1.Spec = newSpec
			// update status
			alertv1, err := h.updateState("creating", "", alertv1)

			if err == nil {
				alertCreator := NewCreater(h.kubeConfig, h.kubeClient, h.alertClient, h.routeClient)

				// create alert instance
				err = alertCreator.CreateAlert(&alertv1.Spec)

				if err != nil {
					h.updateState("error", fmt.Sprintf("%+v", err), alertv1)
				} else {
					h.updateState("running", "", alertv1)
				}
			}
		}
	}
}

// ObjectDeleted will be called for delete alert events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)
	alertCreator := NewCreater(h.kubeConfig, h.kubeClient, h.alertClient, h.routeClient)
	alertCreator.DeleteAlert(name)
}

// ObjectUpdated will be called for update alert events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *Handler) updateState(statusState string, errorMessage string, alert *alertapi.Alert) (*alertapi.Alert, error) {
	alert.Status.State = statusState
	alert.Status.ErrorMessage = errorMessage
	alert, err := h.updateAlertObject(alert)
	if err != nil {
		log.Errorf("couldn't update the state of alert object: %s", err.Error())
	}
	return alert, err
}

func (h *Handler) updateAlertObject(obj *alertapi.Alert) (*alertapi.Alert, error) {
	return h.alertClient.SynopsysV1().Alerts(h.config.Namespace).Update(obj)
}
