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
	"reflect"
	"strings"
	"time"

	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HandlerInterface contains the methods that are required
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
	protoformDeployer       *protoform.Deployer
	opsSightClient          *opssightclientset.Clientset
	isBlackDuckClusterScope bool
	defaults                *opssightapi.OpsSightSpec
	blackDuckClient         *blackduckclientset.Clientset
}

// ObjectCreated will be called for create opssight events
func (h *Handler) ObjectCreated(obj interface{}) {
	recordEvent("objectCreated")
	h.ObjectUpdated(nil, obj)
}

// ObjectDeleted will be called for delete opssight events
func (h *Handler) ObjectDeleted(name string) {
	recordEvent("objectDeleted")
	log.Debugf("objectDeleted: %+v", name)

	// if cluster scope, then check whether the OpsSight CRD exist. If not exist, then don't delete the instance
	if h.protoformDeployer.Config.IsClusterScoped {
		apiClientset, err := clientset.NewForConfig(h.protoformDeployer.KubeConfig)
		crd, err := apiClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(util.OpsSightCRDName, metav1.GetOptions{})
		if err != nil || crd.DeletionTimestamp != nil {
			// We do not delete the OpsSight instance if the CRD doesn't exist or that it is in the process of being deleted
			log.Warnf("Ignoring request to delete %s because the %s CRD doesn't exist or is being deleted", name, util.OpsSightCRDName)
			return
		}
	}

	// Voluntary deletion. The CRD still exists but the OpsSight resource has been deleted
	app := apps.NewApp(h.protoformDeployer)
	err := app.OpsSight().Delete(name)
	if err != nil {
		log.Error(err)
	}
}

// ObjectUpdated will be called for update opssight events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	recordEvent("objectUpdated")
	// log.Debugf("objectUpdated: %+v", objNew)
	opsSight, ok := objNew.(*opssightapi.OpsSight)
	if !ok {
		recordError("unable to cast opssight object")
		log.Error("Unable to cast OpsSight object")
		return
	}

	var err error
	if _, ok = opsSight.Annotations["synopsys.com/created.by"]; !ok {
		opsSight.Annotations = util.InitAnnotations(opsSight.Annotations)
		opsSight.Annotations["synopsys.com/created.by"] = h.protoformDeployer.Config.Version
		opsSight, err = util.UpdateOpsSight(h.opsSightClient, h.protoformDeployer.Config.Namespace, opsSight)
		if err != nil {
			log.Errorf("couldn't update the annotation for %s OpsSight instance in %s namespace due to %+v", opsSight.Name, opsSight.Spec.Namespace, err)
			return
		}
	}

	newSpec := opsSight.Spec
	defaultSpec := h.defaults
	err = mergo.Merge(&newSpec, defaultSpec)
	if err != nil {
		recordError("unable to merge default and new objects")
		h.updateState(Error, err.Error(), opsSight)
		log.Errorf("unable to merge default and new objects due to %+v", err)
		return
	}

	opsSight.Spec = newSpec

	// An error occurred. We wait for one minute before we try to ensure again
	if strings.EqualFold(opsSight.Status.State, string(Error)) {
		time.Sleep(time.Minute * 1)
	}

	log.Debugf("ObjectUpdated: %s", opsSight.Name)

	// Ensure
	app := apps.NewApp(h.protoformDeployer)
	err = app.OpsSight().Ensure(opsSight)
	if err != nil {
		log.Error(err)
		_, err = h.updateState(Error, err.Error(), opsSight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotatef(err, "couldn't update the state for %s OpsSight instance in %s namespace", opsSight.Name, opsSight.Spec.Namespace))
		}
		return
	}

	switch strings.ToUpper(opsSight.Spec.DesiredState) {
	case "STOP":
		_, err = h.updateState(Stopped, "", opsSight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotatef(err, "couldn't update the stopped state for %s OpsSight instance in %s namespace", opsSight.Name, opsSight.Spec.Namespace))
			return
		}
	case "":
		if !strings.EqualFold(opsSight.Status.State, string(Running)) {
			_, err = h.updateState(Running, "", opsSight)
			if err != nil {
				recordError("unable to update state")
				log.Error(errors.Annotatef(err, "couldn't update the running state for %s OpsSight instance in %s namespace", opsSight.Name, opsSight.Spec.Namespace))
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
	newOpssight, err := util.GetOpsSight(h.opsSightClient, opssight.Spec.Namespace, opssight.Name)
	if err != nil {
		return nil, errors.Annotate(err, "unable to get the opssigh to update the state of opssight object")
	}

	if !reflect.DeepEqual(newOpssight.Status.State, opssight.Status.State) || !reflect.DeepEqual(newOpssight.Status.ErrorMessage, opssight.Status.ErrorMessage) {
		newOpssight.Spec = opssight.Spec
		newOpssight.Status.State = string(state)
		newOpssight.Status.ErrorMessage = errorMessage
		newOpssight, err = h.updateOpsSightObject(newOpssight)
		if err != nil {
			return nil, errors.Annotate(err, "unable to update the state of opssight object")
		}
	}
	return newOpssight, nil
}

func (h *Handler) updateOpsSightObject(obj *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	return h.opsSightClient.SynopsysV1().OpsSights(h.protoformDeployer.Config.Namespace).Update(obj)
}
