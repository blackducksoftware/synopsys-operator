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
	"time"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface contains the methods that are required
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// State contains the state of the Alert
type State string

// DesiredState contains the desired state of the Alert
type DesiredState string

const (
	// Running is used when the instance is running
	Running State = "Running"
	// Stopped is used when the instance is about to stop
	Stopped State = "Stopped"
	// Updating is used when the instance is about to update
	Updating State = "Updating"
	// Error is used when the instance deployment errored out
	Error State = "Error"

	// Start is used when the instance to be created or updated
	Start DesiredState = ""
	// Stop is used when the instance to be stopped
	Stop DesiredState = "Stop"
)

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
	h.ObjectUpdated(nil, obj)
}

// ObjectDeleted will be called for delete alert events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)

	// if cluster scope, then check whether the alert CRD exist. If not exist, then don't delete the instance
	if h.config.IsClusterScoped {
		apiClientset, err := clientset.NewForConfig(h.kubeConfig)
		crd, err := apiClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(util.AlertCRDName, metav1.GetOptions{})
		if err != nil || crd.DeletionTimestamp != nil {
			// We do not delete the Alert instance if the CRD doesn't exist or that it is in the process of being deleted
			log.Warnf("Ignoring request to delete %s because the %s CRD doesn't exist or is being deleted", name, util.AlertCRDName)
			return
		}
	}

	app := apps.NewApp(h.config, h.kubeConfig)
	err := app.Alert().Delete(name)
	if err != nil {
		log.Error(err)
	}
}

// ObjectUpdated will be called for update alert events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("updating Object to %+v", objNew)
	// Verify the object is an Alert
	alert, ok := objNew.(*alertapi.Alert)
	if !ok {
		log.Error("unable to cast to Alert object")
		return
	}

	var err error
	if _, ok = alert.Annotations["synopsys.com/created.by"]; !ok {
		alert.Annotations = util.InitAnnotations(alert.Annotations)
		alert.Annotations["synopsys.com/created.by"] = h.config.Version
		alert, err = util.UpdateAlert(h.alertClient, h.config.CrdNamespace, alert)
		if err != nil {
			log.Errorf("couldn't update the annotation for %s Alert instance in %s namespace due to %+v", alert.Name, alert.Spec.Namespace, err)
			return
		}
	}

	// Get Default fields for Alert
	newSpec := alert.Spec
	alertDefaultSpec := h.defaults
	err = mergo.Merge(&newSpec, alertDefaultSpec)
	if err != nil {
		log.Errorf("unable to merge the Alert structs for %s due to %+v", alert.Name, err)
		alert, err = h.updateState(Error, fmt.Sprintf("unable to merge the Alert structs for %s due to %+v", alert.Name, err), alert)
		if err != nil {
			log.Errorf("couldn't update Alert state: %v", err)
		}
		return
	}
	alert.Spec = newSpec

	// An error occurred. We wait for one minute before we try to ensure again
	if strings.EqualFold(alert.Status.State, string(Error)) {
		time.Sleep(time.Minute * 1)
	}

	// Update the Alert
	app := apps.NewApp(h.config, h.kubeConfig)
	err = app.Alert().Ensure(alert)
	if err != nil {
		log.Errorf("unable to ensure the Alert %s due to %+v", alert.Name, err)
		alert, err = h.updateState(Error, fmt.Sprintf("%+v", err), alert)
		if err != nil {
			log.Errorf("couldn't update Alert state: %v", err)
		}
		return
	}

	if strings.EqualFold(alert.Spec.DesiredState, string(Stop)) {
		if !strings.EqualFold(alert.Status.State, string(Stopped)) {
			_, err = h.updateState(Stopped, "", alert)
			if err != nil {
				log.Errorf("couldn't update Alert state: %v", err)
			}
		}
	} else {
		if !strings.EqualFold(alert.Status.State, string(Running)) {
			_, err = h.updateState(Running, "", alert)
			if err != nil {
				log.Errorf("couldn't update Alert state: %v", err)
			}
		}
	}
}

func (h *Handler) updateState(statusState State, errorMessage string, alert *alertapi.Alert) (*alertapi.Alert, error) {
	alert.Status.State = string(statusState)
	alert.Status.ErrorMessage = errorMessage
	alert, err := h.updateAlertObject(alert)
	if err != nil {
		log.Errorf("couldn't update the state of alert object: %s", err.Error())
	}
	return alert, err
}

func (h *Handler) updateAlertObject(obj *alertapi.Alert) (*alertapi.Alert, error) {
	return util.UpdateAlert(h.alertClient, h.config.CrdNamespace, obj)
}
