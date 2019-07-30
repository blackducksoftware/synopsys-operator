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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	blackduckutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HandlerInterface interface contains the methods that are required
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// State contains the state of the Black Duck
type State string

// DesiredState contains the desired state of the Black Duck
type DesiredState string

const (
	// Running is used when the instance is running
	Running State = "Running"
	// Starting is used when the instance is starting
	Starting State = "Starting"
	// Stopped is used when the instance is about to stop
	Stopped State = "Stopped"
	// Error is used when the instance deployment errored out
	Error State = "Error"
	// DbMigration is used when the instance is about to be in the migrated state
	DbMigration DesiredState = "DbMigration"

	// Start is used when the instance to be created or updated
	Start DesiredState = ""
	// Stop is used when the instance to be stopped
	Stop DesiredState = "Stop"
	// DbMigrate is used when the instance is migrated
	DbMigrate DesiredState = "DbMigrate"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	protoformDeployer *protoform.Deployer
	blackduckClient   *blackduckclientset.Clientset
	defaults          *blackduckapi.BlackduckSpec
}

// NewHandler will create the handler
func NewHandler(protoformDeployer *protoform.Deployer, blackDuckClient *blackduckclientset.Clientset, defaults *blackduckapi.BlackduckSpec) *Handler {
	return &Handler{protoformDeployer: protoformDeployer, blackduckClient: blackDuckClient, defaults: defaults}
}

// APISetHubsRequest to set the Black Duck urls for Perceptor
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

	// if cluster scope, then check whether the Black Duck CRD exist. If not exist, then don't delete the instance
	if h.protoformDeployer.Config.IsClusterScoped {
		crd, err := h.protoformDeployer.APIExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(util.BlackDuckCRDName, metav1.GetOptions{})
		if err != nil || crd.DeletionTimestamp != nil {
			// We do not delete the Black Duck instance if the CRD doesn't exist or that it is in the process of being deleted
			log.Warnf("Ignoring request to delete %s because the %s CRD doesn't exist or is being deleted", name, util.BlackDuckCRDName)
			return
		}
	}

	// Voluntary deletion. The CRD still exists but the Black Duck resource has been deleted
	app := apps.NewApp(h.protoformDeployer)
	err := app.Blackduck().Delete(name)
	if err != nil {
		log.Error(err)
	}
}

// ObjectUpdated will be called for update black duck events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	var err error
	bd, ok := objNew.(*blackduckapi.Blackduck)
	if !ok {
		log.Error("unable to cast Black Duck object")
		return
	}

	if _, ok = bd.Annotations["synopsys.com/created.by"]; !ok {
		bd.Annotations = util.InitAnnotations(bd.Annotations)
		bd.Annotations["synopsys.com/created.by"] = h.protoformDeployer.Config.Version
		bd, err = util.UpdateBlackduck(h.blackduckClient, h.protoformDeployer.Config.Namespace, bd)
		if err != nil {
			log.Errorf("couldn't update the annotation for %s Black Duck instance in %s namespace due to %+v", bd.Name, bd.Spec.Namespace, err)
			return
		}
	}

	newSpec := bd.Spec
	blackDuckDefaultSpec := h.defaults
	err = mergo.Merge(&newSpec, blackDuckDefaultSpec)
	if err != nil {
		log.Errorf("unable to merge the Black Duck structs for %s due to %+v", bd.Name, err)
		bd, err = blackduckutils.UpdateState(h.blackduckClient, bd.Name, bd.Spec.Namespace, string(Error), err)
		if err != nil {
			log.Errorf("couldn't update the Black Duck state: %v", err)
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
	app := apps.NewApp(h.protoformDeployer)
	err = app.Blackduck().Ensure(bd)
	if err != nil {
		log.Error(err)
		bd, err = blackduckutils.UpdateState(h.blackduckClient, bd.Name, bd.Spec.Namespace, string(Error), err)
		if err != nil {
			log.Errorf("couldn't update the state for %s Black Duck instance in %s namespace due to %+v", bd.Name, bd.Spec.Namespace, err)
		}
		return
	}

	if strings.EqualFold(bd.Spec.DesiredState, string(Stop)) { // Stop State
		if !strings.EqualFold(bd.Status.State, string(Stopped)) {
			bd, err = blackduckutils.UpdateState(h.blackduckClient, bd.Name, bd.Spec.Namespace, string(Stopped), nil)
			if err != nil {
				log.Errorf("couldn't update the state for %s Black Duck instance in %s namespace due to %+v", bd.Name, bd.Spec.Namespace, err)
			}
		}
	} else if strings.EqualFold(bd.Spec.DesiredState, string(DbMigrate)) { // DbMigrate State
		if !strings.EqualFold(bd.Status.State, string(DbMigration)) {
			bd, err = blackduckutils.UpdateState(h.blackduckClient, bd.Name, bd.Spec.Namespace, string(DbMigration), nil)
			if err != nil {
				log.Errorf("couldn't update the state for %s Black Duck instance in %s namespace due to %+v", bd.Name, bd.Spec.Namespace, err)
			}
		}
	} else { // Start, Running, and Error States
		if !strings.EqualFold(bd.Status.State, string(Running)) {
			// Verify that we can access the Black Duck
			blackDuckURL := fmt.Sprintf("%s.%s.svc", utils.GetResourceName(bd.Name, util.BlackDuckName, "webserver"), bd.Spec.Namespace)
			status := h.verifyBlackDuck(blackDuckURL, bd.Spec.Namespace, bd.Name)

			if status { // Set state to Running if we can access the Black Duck
				bd, err = blackduckutils.UpdateState(h.blackduckClient, bd.Name, bd.Spec.Namespace, string(Running), nil)
				if err != nil {
					log.Errorf("couldn't update the state for %s Black Duck instance in %s namespace due to %+v", bd.Name, bd.Spec.Namespace, err)
				}
			}
		}
	}
}

func (h *Handler) verifyBlackDuck(blackDuckURL string, namespace string, name string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	for i := 0; i < 10; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", blackDuckURL))
		if err != nil {
			log.Debugf("unable to talk with the Black Duck %s", blackDuckURL)
			time.Sleep(10 * time.Second)
			_, err := util.GetBlackDuck(h.blackduckClient, namespace, name)
			if err != nil {
				return false
			}
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Debugf("response status for %s Black Duck instance in %s namespace is %v", name, namespace, resp.Status)

		if resp.StatusCode == 200 {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	return false
}

func (h *Handler) isBinaryAnalysisEnabled(envs []string) bool {
	for _, value := range envs {
		if strings.Contains(value, "USE_BINARY_UPLOADS") {
			values := strings.SplitN(value, ":", 2)
			if len(values) == 2 {
				mapValue := strings.TrimSpace(values[1])
				if strings.EqualFold(mapValue, "1") {
					return true
				}
			}
			return false
		}
	}
	return false
}
