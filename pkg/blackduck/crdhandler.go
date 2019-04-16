/*
Copyright (C) 2019 Synopsys, Inc.

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

package blackduck

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	hubutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface interface contains the methods that are required
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
	// Running is used when the instance is running
	Running State = "Running"
	// Stopped is used when the instance is about to stop
	Stopped State = "Stopped"
	// Error is used when the instance deployment errored out
	Error State = "Error"

	// Start is used when the instance  to be created or updated
	Start DesiredState = "Start"
	// Stop is used when the instance  to be stopped
	Stop DesiredState = "Stop"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	blackduckClient  *blackduckclientset.Clientset
	defaults         *blackduckv1.BlackduckSpec
	federatorBaseURL string
	cmMutex          chan bool
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewHandler will create the handler
func NewHandler(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *blackduckclientset.Clientset, defaults *blackduckv1.BlackduckSpec,
	federatorBaseURL string, cmMutex chan bool, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Handler {
	return &Handler{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, blackduckClient: hubClient, defaults: defaults,
		federatorBaseURL: federatorBaseURL, cmMutex: cmMutex, osSecurityClient: osSecurityClient, routeClient: routeClient}
}

// APISetHubsRequest to set the Blackduck urls for Perceptor
type APISetHubsRequest struct {
	HubURLs []string
}

// ObjectCreated will be called for create hub events
func (h *Handler) ObjectCreated(obj interface{}) {
	h.ObjectUpdated(nil, obj)

}

// ObjectDeleted will be called for delete hub events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %+v", name)

	apiClientset, err := clientset.NewForConfig(h.kubeConfig)
	crd, err := apiClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("blackducks.synopsys.com", v1.GetOptions{})
	if err != nil || crd.DeletionTimestamp != nil {
		// We do not delete the Blackduck instance if the CRD doesn't exist or that it is in the process of being deleted
		log.Warnf("Ignoring request to delete %s because the CRD doesn't exist or is being deleted", name)
		return
	}

	// Voluntary deletion. The CRD still exists but the Blackduck resource has been deleted
	app := apps.NewApp(h.config, h.kubeConfig)
	app.Blackduck().Delete(name)

}

// ObjectUpdated will be called for update hub events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	bd, ok := objNew.(*blackduckv1.Blackduck)
	if !ok {
		log.Error("Unable to cast Blackduck object")
		return
	}

	newSpec := bd.Spec
	hubDefaultSpec := h.defaults
	err := mergo.Merge(&newSpec, hubDefaultSpec)
	if err != nil {
		log.Errorf("unable to merge the hub structs for %s due to %+v", bd.Name, err)
		bd, err = hubutils.UpdateState(h.blackduckClient, bd.Name, h.config.Namespace, string(Error), err)
		if err != nil {
			log.Errorf("Couldn't update the blackduck state: %v", err)
		}
		return
	}
	bd.Spec = newSpec

	// An error occurred. We wait for one minute before we try to ensure again
	if strings.EqualFold(bd.Status.State, string(Error)) {
		time.Sleep(time.Minute * 1)
	}

	log.Debugf("ObjectUpdated: %s", bd.Name)

	// Ensure
	app := apps.NewApp(h.config, h.kubeConfig)
	err = app.Blackduck().Ensure(bd)
	if err != nil {
		log.Error(err)
		bd, err = hubutils.UpdateState(h.blackduckClient, bd.Name, h.config.Namespace, string(Error), err)
		if err != nil {
			log.Errorf("Couldn't update the blackduck state: %v", err)
		}
		return
	}

	// Verify that we can access the Hub
	hubURL := fmt.Sprintf("webserver.%s.svc", bd.Spec.Namespace)
	h.verifyHub(hubURL, bd.Spec.Namespace)

	if !strings.EqualFold(bd.Status.State, string(Running)) {
		bd, err = hubutils.UpdateState(h.blackduckClient, bd.Name, h.config.Namespace, string(Running), nil)
		if err != nil {
			log.Errorf("Couldn't update the blackduck state: %v", err)
		}
	}
}

func (h *Handler) callHubFederator() {
	// IMPORTANT ! This will block.
	h.cmMutex <- true
	defer func() {
		<-h.cmMutex
	}()
	hubUrls, err := h.getHubUrls()
	log.Debugf("hubUrls: %+v", hubUrls)
	if err != nil {
		log.Errorf("unable to get the hub urls due to %+v", err)
		return
	}
	err = h.addHubFederatorEvents(fmt.Sprintf("%s/sethubs", h.federatorBaseURL), hubUrls)
	if err != nil {
		log.Errorf("unable to update the hub urls in perceptor due to %+v", err)
		return
	}
}

// HubNamespaces will list the hub namespaces
func (h *Handler) getHubUrls() (*APISetHubsRequest, error) {
	// 1. get Blackduck CDR list from default ns
	hubList, err := util.ListHubs(h.blackduckClient, h.config.Namespace)
	if err != nil {
		return &APISetHubsRequest{}, err
	}

	// 2. extract the namespaces
	hubURLs := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 && strings.EqualFold(hub.Spec.DesiredState, "running") {
			hubURL := fmt.Sprintf("webserver.%s.svc", hub.Spec.Namespace)
			status := h.verifyHub(hubURL, hub.Spec.Namespace)
			if status {
				hubURLs = append(hubURLs, hubURL)
			}
		}
	}

	return &APISetHubsRequest{HubURLs: hubURLs}, nil
}

func (h *Handler) verifyHub(hubURL string, name string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	for i := 0; i < 120; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", hubURL))
		if err != nil {
			log.Debugf("unable to talk with the hub %s", hubURL)
			time.Sleep(10 * time.Second)
			_, err := util.GetHub(h.blackduckClient, h.config.Namespace, name)
			if err != nil {
				return false
			}
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Debugf("hub response status for %s is %v", hubURL, resp.Status)

		if resp.StatusCode == 200 {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	return false
}

func (h *Handler) addHubFederatorEvents(dest string, obj interface{}) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to serialize %v: %v", obj, err)
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, dest, bytes.NewBuffer(jsonBytes))
	log.Debugf("hub req: %+v", req)
	if err != nil {
		return fmt.Errorf("unable to create the request due to %v", err)
	}
	resp, err := client.Do(req)
	log.Debugf("hub resp: %+v", resp)
	if err != nil {
		return fmt.Errorf("unable to POST to %s: %v", dest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http POST request to %s failed with status code %d", dest, resp.StatusCode)
	}
	return nil
}

func (h *Handler) isBinaryAnalysisEnabled(envs []string) bool {
	for _, value := range envs {
		if strings.Contains(value, "USE_BINARY_UPLOADS") {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				mapValue := strings.Trim(values[1], " ")
				if strings.EqualFold(mapValue, "1") {
					return true
				}
			}
			return false
		}
	}
	return false
}
