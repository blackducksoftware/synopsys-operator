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

package opssight

import (
	"fmt"
	"strings"

	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/imdario/mergo"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface contains the methods that are required
// ... not really sure why we have this type
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// State contains the state of the OpsSight
type State string

// DesiredState contains the desired state of the OpsSight
type DesiredState string

const (
	// Running is used when OpsSight is running
	Running State = "Running"
	// Stopped is used when OpsSight to be stopped
	Stopped State = "Stopped"
	// Error is used when OpsSight deployment errored out
	Error State = "Error"

	// Start is used when OpsSight deployment to be created or updated
	Start DesiredState = ""
	// Stop is used when OpsSight deployment to be stopped
	Stop DesiredState = "Stop"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	Config           *protoform.Config
	KubeConfig       *rest.Config
	KubeClient       *kubernetes.Clientset
	OpsSightClient   *opssightclientset.Clientset
	Defaults         *opssightapi.OpsSightSpec
	Namespace        string
	OSSecurityClient *securityclient.SecurityV1Client
	RouteClient      *routeclient.RouteV1Client
	HubClient        *hubclientset.Clientset
}

// ObjectCreated will be called for create opssight events
func (h *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	recordEvent("objectCreated")
	h.ObjectUpdated(nil, obj)
}

// ObjectDeleted will be called for delete opssight events
func (h *Handler) ObjectDeleted(name string) {
	recordEvent("objectDeleted")
	log.Debugf("objectDeleted: %+v", name)
	app := apps.NewApp(h.Config, h.KubeConfig)
	app.OpsSight().Delete(name)
}

// ObjectUpdated will be called for update opssight events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	recordEvent("objectUpdated")
	log.Debugf("objectUpdated: %+v", objNew)
	opssight, ok := objNew.(*opssightapi.OpsSight)
	if !ok {
		recordError("unable to cast opssight object")
		log.Error("Unable to cast OpsSight object")
		return
	}

	newSpec := opssight.Spec
	defaultSpec := h.Defaults
	err := mergo.Merge(&newSpec, defaultSpec)
	if err != nil {
		recordError("unable to merge default and new objects")
		h.updateState(Error, err.Error(), opssight)
		log.Errorf("unable to merge default and new objects due to %+v", err)
		return
	}

	opssight.Spec = newSpec

	switch strings.ToUpper(opssight.Spec.DesiredState) {
	case "STOP":
		app := apps.NewApp(h.Config, h.KubeConfig)
		err = app.OpsSight().Stop(opssight)
		if err != nil {
			recordError("unable to stop opssight")
			log.Errorf("handle object stop: %s", err.Error())
			h.updateState(Error, err.Error(), opssight)
			return
		}

		_, err = h.updateState(Stopped, "", opssight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotate(err, "unable to update stopped state"))
			return
		}
	case "":
		app := apps.NewApp(h.Config, h.KubeConfig)
		err = app.OpsSight().Ensure(opssight)
		if err != nil {
			recordError("unable to update opssight")
			log.Errorf("handle object update: %s", err.Error())
			_, err = h.updateState(Error, err.Error(), opssight)
			if err != nil {
				recordError("unable to update state")
				log.Error(errors.Annotate(err, "unable to update running state"))
				return
			}
			return
		}

		if !strings.EqualFold(opssight.Status.State, string(Running)) {
			_, err = h.updateState(Running, "", opssight)
			if err != nil {
				recordError("unable to update state")
				log.Error(errors.Annotate(err, "unable to update running state"))
				return
			}
		}
	default:
		recordError("unable to find the desired state value")
		log.Errorf("unable to handle object update due to %+v", fmt.Errorf("desired state value is not expected"))
		return
	}
}

func (h *Handler) updateState(state State, errorMessage string, opssight *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	opssight.Status.State = string(state)
	opssight.Status.ErrorMessage = errorMessage
	opssight, err := h.OpsSightClient.SynopsysV1().OpsSights(h.Namespace).Update(opssight)
	if err != nil {
		return nil, errors.Annotate(err, "unable to update the state of opssight object")
	}
	return opssight, nil
}
