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

	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	cr "github.com/blackducksoftware/synopsys-operator/pkg/apps/rgp/latest"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	rgpclientset "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
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
	config     *protoform.Config
	kubeConfig *rest.Config
	kubeClient *kubernetes.Clientset
	rgpClient  *rgpclientset.Clientset
}

// NewHandler ...
func NewHandler(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, rgpClient *rgpclientset.Clientset) *Handler {
	return &Handler{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, rgpClient: rgpClient}
}

// ObjectCreated will be called for create events
func (h *Handler) ObjectCreated(obj interface{}) {
	var err error
	log.Debugf("ObjectCreated: %+v", obj)
	gr, ok := obj.(*v1.Rgp)
	if !ok {
		log.Error("Unable to cast object")
		return
	}

	if len(gr.Status.State) > 0 {
		return
	}

	if strings.EqualFold(gr.Status.State, string(Running)) || strings.EqualFold(gr.Status.State, string(Stopped)) {
		h.ObjectUpdated(nil, gr)
	}

	log.Info(gr.Name)

	gr.Status.State = string(Creating)
	gr, err = h.rgpClient.SynopsysV1().Rgps(h.config.Namespace).Update(gr)
	if err != nil {
		log.Error(err.Error())
		return
	}

	creater := cr.NewCreater(h.kubeConfig, h.kubeClient, h.rgpClient)
	err = creater.Create(&gr.Spec)
	if err != nil {
		log.Error(err.Error())
		gr.Status.ErrorMessage = err.Error()
		gr.Status.State = string(Error)
	} else {
		gr.Status.Fqdn = fmt.Sprintf("%s/reporting", gr.Spec.IngressHost)
		gr.Status.State = string(Running)
	}

	_, err = h.rgpClient.SynopsysV1().Rgps(h.config.Namespace).Update(gr)
	if err != nil {
		log.Error(err.Error())
	}

}

// ObjectDeleted will be called for delete events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %s", name)
	err := util.DeleteNamespace(h.kubeClient, name)
	if err != nil {
		log.Error(err.Error())
	}
}

// ObjectUpdated will be called for update events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	gr, ok := objNew.(*v1.Rgp)
	if !ok {
		log.Error("Unable to cast object")
		return
	}
	if strings.EqualFold(gr.Status.State, string(Running)) || strings.EqualFold(gr.Status.State, string(Stopped)) {
		log.Debugf("Update: %s", gr.Name)
		creater := cr.NewCreater(h.kubeConfig, h.kubeClient, h.rgpClient)
		err := creater.Update(&gr.Spec)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
