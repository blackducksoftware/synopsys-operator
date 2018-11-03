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

package hub

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	hub_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub/hubutils"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface interface contains the methods that are required
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	hubClient        *hubclientset.Clientset
	defaults         *hub_v1.HubSpec
	federatorBaseURL string
	cmMutex          chan bool
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewHandler will create the handler
func NewHandler(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset, defaults *hub_v1.HubSpec,
	federatorBaseURL string, cmMutex chan bool, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Handler {
	return &Handler{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, hubClient: hubClient, defaults: defaults,
		federatorBaseURL: federatorBaseURL, cmMutex: cmMutex, osSecurityClient: osSecurityClient, routeClient: routeClient}
}

// APISetHubsRequest to set the Hub urls for Perceptor
type APISetHubsRequest struct {
	HubURLs []string
}

// ObjectCreated will be called for create hub events
func (h *Handler) ObjectCreated(obj interface{}) {
	log.Debugf("ObjectCreated: %+v", obj)
	hubv1 := obj.(*hub_v1.Hub)
	if strings.EqualFold(hubv1.Spec.State, "") {
		newSpec := hubv1.Spec
		hubDefaultSpec := h.defaults
		err := mergo.Merge(&newSpec, hubDefaultSpec)
		log.Debugf("merged hub details %+v", newSpec)
		if err != nil {
			log.Errorf("unable to merge the hub structs for %s due to %+v", hubv1.Name, err)
			hubutils.UpdateState(h.hubClient, "error", "error", err, hubv1)
		} else {
			hubv1.Spec = newSpec
			// Update status
			hubv1, err := hubutils.UpdateState(h.hubClient, "pending", "creating", nil, hubv1)

			if err == nil {
				hubCreator := NewCreater(h.config, h.kubeConfig, h.kubeClient, h.hubClient, h.osSecurityClient, h.routeClient)
				ip, pvc, updateError, err := hubCreator.CreateHub(&hubv1.Spec)
				if err != nil {
					log.Errorf("unable to create hub for %s due to %+v", hubv1.Name, err)
				}
				hubv1.Status.IP = ip
				hubv1.Status.PVCVolumeName = pvc
				if updateError {
					hubutils.UpdateState(h.hubClient, "error", "error", err, hubv1)
				} else {
					hubutils.UpdateState(h.hubClient, "running", "running", err, hubv1)
					hubURL := fmt.Sprintf("webserver.%s.svc", hubv1.Spec.Namespace)
					h.verifyHub(hubURL, hubv1.Spec.Namespace)
					h.autoRegisterHub(&hubv1.Spec)
					h.callHubFederator()
				}
			}
		}
	}

	log.Infof("Done w/ install, starting post-install nanny monitors...")

}

// ObjectDeleted will be called for delete hub events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %+v", name)

	hubCreator := NewCreater(h.config, h.kubeConfig, h.kubeClient, h.hubClient, h.osSecurityClient, h.routeClient)
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
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	//if strings.Compare(objOld.Spec.State, objNew.Spec.State) != 0 {
	//	log.Infof("%s - Changing state [%s] -> [%s] | Current: [%s]", objNew.Name, objOld.Spec.State, objNew.Spec.State, objNew.Status.State )
	//	// TO DO
	//	objNew.Status.State = objNew.Spec.State
	//	h.hubClient.SynopsysV1().Hubs(objNew.Namespace).Update(objNew)
	//}
}

func (h *Handler) autoRegisterHub(createHub *hub_v1.HubSpec) error {
	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod, err := util.FilterPodByNamePrefixInNamespace(h.kubeClient, createHub.Namespace, "registration")
	log.Debugf("registration pod: %+v", registrationPod)
	if err != nil {
		log.Errorf("unable to filter the registration pod in %s because %+v", createHub.Namespace, err)
		return err
	}
	registrationKey := os.Getenv("REGISTRATION_KEY")

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for i := 0; i < 20; i++ {
			registrationPod, err := util.GetPods(h.kubeClient, createHub.Namespace, registrationPod.Name)
			if err != nil {
				log.Errorf("unable to find the registration pod in %s because %+v", createHub.Namespace, err)
				return err
			}

			// Create the exec into kubernetes pod request
			req := util.CreateExecContainerRequest(h.kubeClient, registrationPod)
			// Exec into the kubernetes pod and execute the commands
			if strings.HasPrefix(createHub.HubVersion, "4.") {
				err = util.ExecContainer(h.kubeConfig, req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate"`, registrationKey)})
			} else {
				err = util.ExecContainer(h.kubeConfig, req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate" -k --cert /opt/blackduck/hub/hub-registration/security/blackduck_system.crt --key /opt/blackduck/hub/hub-registration/security/blackduck_system.key`, registrationKey)})
			}

			if err == nil {
				log.Infof("hub %s is created and auto registered. Exit!!!!", createHub.Namespace)
				return nil
			}
			log.Infof("error in Stream: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
	log.Errorf("unable to register the hub for %s.... please manually auto register the hub", createHub.Namespace)
	return fmt.Errorf("unable to register the hub for %s.... please manually auto register the hub", createHub.Namespace)
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
	// 1. get Hub CDR list from default ns
	hubList, err := util.ListHubs(h.hubClient, h.config.Namespace)
	if err != nil {
		return &APISetHubsRequest{}, err
	}

	// 2. extract the namespaces
	hubURLs := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 && strings.EqualFold(hub.Spec.State, "running") {
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
			_, err := util.GetHub(h.hubClient, h.config.Namespace, name)
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
