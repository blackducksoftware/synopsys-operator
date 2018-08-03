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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const perceptorSetHubsURL = "http://hub-federator:3016/sethubs"

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
	namespace    string
}

// APISetHubsRequest to set the Hub urls for Perceptor
type APISetHubsRequest struct {
	HubURLs []string
}

// ObjectCreated will be called for create hub events
func (h *HubHandler) ObjectCreated(obj *hubv1.Hub) {
	log.Debugf("ObjectCreated: %+v", obj)
	if strings.EqualFold(obj.Spec.State, "") {
		// Update status
		obj.Spec.State = "pending"
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
		h.callPerceptor()
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
	// h.callPerceptor()

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
	return h.hubClientset.SynopsysV1().Hubs(h.namespace).Update(obj)
}

func (h *HubHandler) callPerceptor() {
	hubUrls, err := h.getHubUrls()
	log.Debugf("hubUrls: %+v", hubUrls)
	if err != nil {
		log.Errorf("unable to get the hub urls due to %+v", err)
		return
	}
	err = h.addPerceptorEvents(perceptorSetHubsURL, hubUrls)
	if err != nil {
		log.Errorf("unable to update the hub urls in perceptor due to %+v", err)
		return
	}
}

// HubNamespaces will list the hub namespaces
func (h *HubHandler) getHubUrls() (*APISetHubsRequest, error) {
	// 1. get Hub CDR list from default ns
	hubList, err := h.hubClientset.SynopsysV1().Hubs(h.namespace).List(metav1.ListOptions{})
	if err != nil {
		return &APISetHubsRequest{}, err
	}

	// 2. extract the namespaces
	hubURLs := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 {
			hubURL := fmt.Sprintf("webserver.%s.svc", hub.Spec.Namespace)
			h.verifyHub(hubURL)
			hubURLs = append(hubURLs, hubURL)
		}
	}

	return &APISetHubsRequest{HubURLs: hubURLs}, nil
}

func (h *HubHandler) verifyHub(hubURL string) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", hubURL))
		if err != nil {
			log.Debugf("unable to talk with the hub %s", hubURL)
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Debugf("hub response status for %s is %v", hubURL, resp.Status)

		if resp.StatusCode == 200 {
			break
		}
		time.Sleep(10 * time.Second)
	}
}

func (h *HubHandler) addPerceptorEvents(dest string, obj interface{}) error {
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
