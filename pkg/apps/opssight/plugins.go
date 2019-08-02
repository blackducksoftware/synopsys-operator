/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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
	"time"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var logger *log.Entry

func init() {
	logger = log.WithField("subsystem", "opssight-plugins")
}

// This is a controller that updates the secret in perceptor periodically.
// It is assumed that the secret in perceptor will roll over any time this is updated, and
// if not, that there is a problem in the orchestration environment.

// Updater stores the opssight updater configuration
type Updater struct {
	config          *protoform.Config
	kubeClient      *kubernetes.Clientset
	blackDuckClient *blackduckclientset.Clientset
	opssightClient  *opssightclientset.Clientset
}

// NewUpdater returns the opssight updater configuration
func NewUpdater(config *protoform.Config, kubeClient *kubernetes.Clientset, blackDuckClient *blackduckclientset.Clientset, opssightClient *opssightclientset.Clientset) *Updater {
	return &Updater{
		config:          config,
		kubeClient:      kubeClient,
		blackDuckClient: blackDuckClient,
		opssightClient:  opssightClient,
	}
}

// Run watches for Black Duck and OpsSight events and update the internal Black Duck hosts in Perceptor secret and
// then patch the corresponding replication controller
func (p *Updater) Run(ch <-chan struct{}) {
	logger.Infof("Starting controller for blackduck<->opssight-core updates... this blocks, so running in a go func.")

	go func() {
		for {
			select {
			case <-ch:
				// stop
				return
			default:
				syncFunc := func() {
					err := p.updateAllBlackDucks()
					if len(err) > 0 {
						logger.Errorf("unable to update Black Ducks because %+v", err)
					}
				}

				// watch for Black Duck events to update an OpsSight internal host only if Black Duck crd is enabled
				if strings.Contains(p.config.CrdNames, util.BlackDuckCRDName) {
					log.Debugf("watch for Black Duck events to update an OpsSight internal hosts")
					syncFunc()

					blackDuckListWatch := &cache.ListWatch{
						ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
							return p.blackDuckClient.SynopsysV1().Blackducks(p.config.Namespace).List(options)
						},
						WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
							return p.blackDuckClient.SynopsysV1().Blackducks(p.config.Namespace).Watch(options)
						},
					}
					_, blackDuckController := cache.NewInformer(blackDuckListWatch,
						&blackduckapi.Blackduck{},
						2*time.Second,
						cache.ResourceEventHandlerFuncs{
							// TODO kinda dumb, we just do a complete re-list of all Black Ducks,
							// every time an event happens... But thats all we need to do, so its good enough.
							DeleteFunc: func(obj interface{}) {
								logger.Debugf("updater - Black Duck deleted event ! %v ", obj)
								syncFunc()
							},

							AddFunc: func(obj interface{}) {
								logger.Debugf("updater - Black Duck added event! %v ", obj)
								running := p.isBlackDuckRunning(obj)
								if !running {
									syncFunc()
								}
							},
						},
					)

					// make sure this is called from a go func -- it blocks!
					go blackDuckController.Run(ch)
					<-ch
				} else {
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()
}

// isBlackDuckRunning return whether the Black Duck instance is in running state
func (p *Updater) isBlackDuckRunning(obj interface{}) bool {
	blackduck, _ := obj.(*blackduckapi.Blackduck)
	if strings.EqualFold(blackduck.Status.State, "Running") {
		return true
	}
	return false
}

// updateAllBlackDucks will update the Black Duck instances in opssight resources
func (p *Updater) updateAllBlackDucks() []error {
	opssights, err := util.ListOpsSights(p.opssightClient, p.config.Namespace)
	if err != nil {
		return []error{errors.Annotatef(err, "unable to list opssight in namespace %s", p.config.Namespace)}
	}

	if len(opssights.Items) == 0 {
		return nil
	}

	errList := []error{}
	for _, opssight := range opssights.Items {
		err = p.updateOpsSight(&opssight)
		if err != nil {
			errList = append(errList, errors.Annotate(err, "unable to update opssight"))
		}
	}
	return errList
}

// updateOpsSight will update the opssight resource with latest Black Duck instances
func (p *Updater) updateOpsSight(opssight *opssightapi.OpsSight) error {
	var err error
	if !strings.EqualFold(opssight.Status.State, "stopped") && !strings.EqualFold(opssight.Status.State, "error") {
		for j := 0; j < 20; j++ {
			if strings.EqualFold(opssight.Status.State, "running") {
				break
			}
			logger.Debugf("waiting for opssight %s to be up.....", opssight.Name)
			time.Sleep(10 * time.Second)

			opssight, err = util.GetOpsSight(p.opssightClient, p.config.Namespace, opssight.Name)
			if err != nil {
				return fmt.Errorf("unable to get opssight %s due to %+v", opssight.Name, err)
			}
		}
		err = p.update(opssight)
	}
	return err
}

// update will list all Black Ducks in the cluster, and send them to opssight as scan targets.
func (p *Updater) update(opssight *opssightapi.OpsSight) error {
	blackDuckType := opssight.Spec.Blackduck.BlackduckSpec.Type

	blackduckPassword, err := util.Base64Decode(opssight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return errors.Annotate(err, "unable to decode blackduckPassword")
	}

	allHubs := p.GetAllBlackDucks(blackDuckType, blackduckPassword)

	err = p.updateOpsSightCRD(opssight, allHubs)
	if err != nil {
		return errors.Annotate(err, "unable to update opssight CRD")
	}
	return nil
}

// GetAllBlackDucks get only the internal Black Duck instances from the cluster
func (p *Updater) GetAllBlackDucks(blackDuckType string, blackduckPassword string) []*opssightapi.Host {
	hosts := []*opssightapi.Host{}
	blackDuckList, err := util.ListBlackDucks(p.blackDuckClient, p.config.Namespace)
	if err != nil {
		log.Errorf("unable to list blackducks due to %+v", err)
	}
	for _, blackDuck := range blackDuckList.Items {
		if strings.EqualFold(blackDuck.Spec.Type, blackDuckType) {
			var concurrentScanLimit int
			switch strings.ToUpper(blackDuck.Spec.Size) {
			case "MEDIUM":
				concurrentScanLimit = 3
			case "LARGE":
				concurrentScanLimit = 4
			case "X-LARGE":
				concurrentScanLimit = 6
			default:
				concurrentScanLimit = 2
			}
			host := &opssightapi.Host{
				Domain:              fmt.Sprintf("%s.%s.svc", utils.GetResourceName(blackDuck.Name, util.BlackDuckName, "webserver"), blackDuck.Spec.Namespace),
				ConcurrentScanLimit: concurrentScanLimit,
				Scheme:              "https",
				User:                "sysadmin",
				Port:                443,
				Password:            blackduckPassword,
			}
			hosts = append(hosts, host)
		}
	}
	log.Debugf("total no of Black Duck's for type %s is %d", blackDuckType, len(hosts))
	return hosts
}

// updateOpsSightCRD will update the opssight CRD
func (p *Updater) updateOpsSightCRD(opsSight *opssightapi.OpsSight, blackDucks []*opssightapi.Host) error {
	opssightName := opsSight.Name
	opsSightNamespace := opsSight.Spec.Namespace
	logger.WithField("opssight", opssightName).Info("update opssight: looking for opssight")
	opssight, err := p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).Get(opssightName, metav1.GetOptions{})
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight %s in %s namespace", opssightName, opsSightNamespace)
	}

	opssight.Status.InternalHosts = p.AppendBlackDuckHosts(opssight.Status.InternalHosts, blackDucks)

	_, err = p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).Update(opssight)
	if err != nil {
		return errors.Annotatef(err, "unable to update opssight %s in %s", opssightName, opsSightNamespace)
	}
	return nil
}

// AppendBlackDuckHosts will append the old and new internal Black Duck hosts
func (p *Updater) AppendBlackDuckHosts(oldBlackDucks []*opssightapi.Host, newBlackDucks []*opssightapi.Host) []*opssightapi.Host {
	existingBlackDucks := make(map[string]*opssightapi.Host)
	for _, oldBlackDuck := range oldBlackDucks {
		existingBlackDucks[oldBlackDuck.Domain] = oldBlackDuck
	}

	finalBlackDucks := []*opssightapi.Host{}
	for _, newBlackDuck := range newBlackDucks {
		if existingBlackduck, ok := existingBlackDucks[newBlackDuck.Domain]; ok {
			// add the existing internal Black Duck from the final Black Duck list
			finalBlackDucks = append(finalBlackDucks, existingBlackduck)
		} else {
			// add the new internal Black Duck to the final Black Duck list
			finalBlackDucks = append(finalBlackDucks, newBlackDuck)
		}
	}

	return finalBlackDucks
}

// AppendBlackDuckSecrets will append the secrets of external and internal Black Duck
func (p *Updater) AppendBlackDuckSecrets(existingExternalBlackDucks map[string]*opssightapi.Host, oldInternalBlackDucks []*opssightapi.Host, newInternalBlackDucks []*opssightapi.Host) map[string]*opssightapi.Host {
	existingInternalBlackducks := make(map[string]*opssightapi.Host)
	for _, oldInternalBlackDuck := range oldInternalBlackDucks {
		existingInternalBlackducks[oldInternalBlackDuck.Domain] = oldInternalBlackDuck
	}

	currentInternalBlackducks := make(map[string]*opssightapi.Host)
	for _, newInternalBlackDuck := range newInternalBlackDucks {
		currentInternalBlackducks[newInternalBlackDuck.Domain] = newInternalBlackDuck
	}

	for _, currentInternalBlackduck := range currentInternalBlackducks {
		// check if external host contains the internal host
		if _, ok := existingExternalBlackDucks[currentInternalBlackduck.Domain]; ok {
			// if internal host contains an external host, then check whether it is already part of status,
			// if yes replace it with existing internal host else with new internal host
			if existingInternalBlackduck, ok1 := existingInternalBlackducks[currentInternalBlackduck.Domain]; ok1 {
				existingExternalBlackDucks[currentInternalBlackduck.Domain] = existingInternalBlackduck
			} else {
				existingExternalBlackDucks[currentInternalBlackduck.Domain] = currentInternalBlackduck
			}
		} else {
			// add new internal Black Duck
			existingExternalBlackDucks[currentInternalBlackduck.Domain] = currentInternalBlackduck
		}
	}

	return existingExternalBlackDucks
}
