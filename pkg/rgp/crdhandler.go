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

package rgp

import (
	"fmt"
	"strings"
	"time"

	rgpapi "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	rgpclientset "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface contains the methods that are required
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// State contains the state of the OpsSight
type State string

const (
	// Creating is used when it is installing or deploying
	Creating State = "Creating"
	// Running is used when it is running
	Running State = "Running"
	// Stopping is used when it is about to stop
	Stopping State = "Stopping"
	// Stopped is used when it is stopped
	Stopped State = "Stopped"
	// Updating is used when it is about to update
	Updating State = "Updating"
	// Error is used when the deployment errored out
	Error State = "Error"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	config      *protoform.Config
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	rgpClient   *rgpclientset.Clientset
	defaults    *rgpapi.RgpSpec
	routeClient *routeclient.RouteV1Client
}

// NewHandler will create the handler
func NewHandler(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, rgpClient *rgpclientset.Clientset, routeClient *routeclient.RouteV1Client, defaults *rgpapi.RgpSpec) *Handler {
	return &Handler{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, rgpClient: rgpClient, routeClient: routeClient, defaults: defaults}
}

// ObjectCreated will be called for create rgp events
func (h *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	h.ObjectUpdated(nil, obj)
}

// ObjectDeleted will be called for delete alert events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)

	apiClientset, err := clientset.NewForConfig(h.kubeConfig)
	crd, err := apiClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("rgps.synopsys.com", v1.GetOptions{})
	if err != nil || crd.DeletionTimestamp != nil {
		// We do not delete the Rgp instance if the CRD doesn't exist or that it is in the process of being deleted
		log.Warnf("Ignoring request to delete %s because the CRD doesn't exist or is being deleted", name)
		return
	}

	app := apps.NewApp(h.config, h.kubeConfig)
	app.Rgp().Delete(name)
}

// ObjectUpdated will be called for update Rgp events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("updating Object to %+v", objNew)
	// Verify the object is an Rgp
	rgp, ok := objNew.(*rgpapi.Rgp)
	if !ok {
		log.Error("unable to cast to Rgp object")
		return
	}

	// Get Default fields for Rgp
	newSpec := rgp.Spec
	rgpDefaultSpec := h.defaults
	err := mergo.Merge(&newSpec, rgpDefaultSpec)
	if err != nil {
		log.Errorf("unable to merge the Rgo structs for %s due to %+v", rgp.Name, err)
		rgp, err = h.updateState(Error, fmt.Sprintf("unable to merge the Rgp structs for %s due to %+v", rgp.Name, err), rgp)
		if err != nil {
			log.Errorf("couldn't update Rgp state: %v", err)
		}
		return
	}
	rgp.Spec = newSpec

	// An error occurred. We wait for one minute before we try to ensure again
	if strings.EqualFold(rgp.Status.State, string(Error)) {
		time.Sleep(time.Minute * 1)
	}

	// Update the Rgp
	app := apps.NewApp(h.config, h.kubeConfig)
	err = app.Rgp().Ensure(rgp)
	if err != nil {
		log.Errorf("unable to ensure the Rgp %s due to %+v", rgp.Name, err)
		rgp, err = h.updateState(Error, fmt.Sprintf("%+v", err), rgp)
		if err != nil {
			log.Errorf("couldn't update Rgp state: %v", err)
		}
		return
	}
	if !strings.EqualFold(rgp.Status.State, string(Running)) {
		_, err = h.updateState(Running, "", rgp)
		if err != nil {
			log.Errorf("couldn't update Rgp state: %v", err)
		}
	}
}

func (h *Handler) updateState(statusState State, errorMessage string, rgp *rgpapi.Rgp) (*rgpapi.Rgp, error) {
	rgp.Status.State = string(statusState)
	rgp.Status.ErrorMessage = errorMessage
	rgp, err := h.updateRgpObject(rgp)
	if err != nil {
		log.Errorf("couldn't update the state of alert object: %s", err.Error())
	}
	return rgp, err
}

func (h *Handler) updateRgpObject(obj *rgpapi.Rgp) (*rgpapi.Rgp, error) {
	return h.rgpClient.SynopsysV1().Rgps(h.config.Namespace).Update(obj)
}
