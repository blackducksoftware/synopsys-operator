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

	hub_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
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

// HubHandler will store the configuration that is required to initiantiate the informers callback
type HubHandler struct {
	Config           *rest.Config
	Clientset        *kubernetes.Clientset
	HubClientset     *hubclientset.Clientset
	Namespace        string
	FederatorBaseURL string
	CmMutex          chan bool
}

// APISetHubsRequest to set the Hub urls for Perceptor
type APISetHubsRequest struct {
	HubURLs []string
}

// ObjectCreated will be called for create hub events
func (h *HubHandler) ObjectCreated(obj interface{}) {
	log.Debugf("ObjectCreated: %+v", obj)
	hubv1 := obj.(*hub_v1.Hub)
	if strings.EqualFold(hubv1.Spec.State, "") {
		// Update status
		hubv1.Spec.State = "pending"
		hubv1.Status.State = "creating"
		hubv1, err := h.updateHubObject(hubv1)
		if err != nil {
			log.Errorf("Couldn't update Hub object: %s", err.Error())
		}

		hubCreator := hub.NewCreater(h.Config, h.Clientset, h.HubClientset)
		if err != nil {
			log.Errorf("unable to create the new hub creater for %s due to %+v", hubv1.Name, err)
		}
		ip, pvc, updateError, err := hubCreator.CreateHub(hubv1)

		if updateError {
			//Set spec/state  and status/state to started
			hubv1.Spec.State = "error"
			hubv1.Status.State = "error"
		} else {
			hubv1.Spec.State = "running"
			hubv1.Status.State = "running"
		}
		hubv1.Status.IP = ip
		hubv1.Status.PVCVolumeName = pvc
		hubv1, err = h.updateHubObject(hubv1)
		if err != nil {
			log.Errorf("Couldn't update Hub object: %s", err.Error())
		}
		h.callHubFederator()
	}
}

// ObjectDeleted will be called for delete hub events
func (h *HubHandler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %+v", name)

	hubCreator := hub.NewCreater(h.Config, h.Clientset, h.HubClientset)
	hubCreator.DeleteHub(name)
	h.callHubFederator()

	//Set spec/state  and status/state to started
	// obj.Spec.State = "deleted"
	// obj.Status.State = "deleted"
	// obj, err = h.updateHubObject(obj)
	// if err != nil {
	// 	log.Errorf("Couldn't update Hub object: %s", err.Error())
	// }
}

// ObjectUpdated will be called for update hub events
func (h *HubHandler) ObjectUpdated(objOld, objNew interface{}) {
	//if strings.Compare(objOld.Spec.State, objNew.Spec.State) != 0 {
	//	log.Infof("%s - Changing state [%s] -> [%s] | Current: [%s]", objNew.Name, objOld.Spec.State, objNew.Spec.State, objNew.Status.State )
	//	// TO DO
	//	objNew.Status.State = objNew.Spec.State
	//	h.hubClientset.SynopsysV1().Hubs(objNew.Namespace).Update(objNew)
	//}
}

func (h *HubHandler) updateHubObject(obj *hub_v1.Hub) (*hub_v1.Hub, error) {
	return h.HubClientset.SynopsysV1().Hubs(h.Namespace).Update(obj)
}

func (h *HubHandler) callHubFederator() {
	// IMPORTANT ! This will block.
	h.CmMutex <- true
	defer func() {
		<-h.CmMutex
	}()
	hubUrls, err := h.getHubUrls()
	log.Debugf("hubUrls: %+v", hubUrls)
	if err != nil {
		log.Errorf("unable to get the hub urls due to %+v", err)
		return
	}
	err = h.addHubFederatorEvents(fmt.Sprintf("%s/sethubs", h.FederatorBaseURL), hubUrls)
	if err != nil {
		log.Errorf("unable to update the hub urls in perceptor due to %+v", err)
		return
	}
}

// HubNamespaces will list the hub namespaces
func (h *HubHandler) getHubUrls() (*APISetHubsRequest, error) {
	// 1. get Hub CDR list from default ns
	hubList, err := hub.ListHubs(h.HubClientset, h.Namespace)
	if err != nil {
		return &APISetHubsRequest{}, err
	}

	// 2. extract the namespaces
	hubURLs := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 && strings.EqualFold(hub.Spec.State, "running") {
			hubURL := fmt.Sprintf("webserver.%s.svc", hub.Spec.Namespace)
			status := h.verifyHub(hubURL)
			if status {
				hubURLs = append(hubURLs, hubURL)
			}
		}
	}

	return &APISetHubsRequest{HubURLs: hubURLs}, nil
}

func (h *HubHandler) verifyHub(hubURL string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", hubURL))
		if err != nil {
			log.Debugf("unable to talk with the hub %s", hubURL)
			time.Sleep(10 * time.Second)
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

func (h *HubHandler) addHubFederatorEvents(dest string, obj interface{}) error {
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
