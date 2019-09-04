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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/util"

	"k8s.io/apimachinery/pkg/types"
)

func (p *OpsSightBlackDuckReconciler) sync() {
	err := p.updateAllBlackDucks()
	if len(err) > 0 {
		p.Log.Error(nil, "unable to update Black Ducks", "errors", err)
	}
}

// isBlackDuckRunning return whether the Black Duck instance is in running state
func (p *OpsSightBlackDuckReconciler) isBlackDuckRunning(blackDuck *synopsysv1.Blackduck) bool {
	if strings.EqualFold(blackDuck.Status.State, "Running") {
		return true
	}
	return false
}

// updateAllBlackDucks will update the Black Duck instances in opssight resources
func (p *OpsSightBlackDuckReconciler) updateAllBlackDucks() []error {
	opsSights := synopsysv1.OpsSightList{}
	if err := p.Client.List(context.TODO(), &opsSights); err != nil {
		return []error{fmt.Errorf("unable to list OpsSight instances due to %+v", err)}
	}

	if len(opsSights.Items) == 0 {
		return nil
	}

	errList := []error{}
	for _, opsSight := range opsSights.Items {
		if err := p.updateOpsSight(&opsSight); err != nil {
			errList = append(errList, fmt.Errorf("unable to update %s OpsSight instance due to %+v", opsSight.Name, err))
		}
	}
	return errList
}

// updateOpsSight will update the opssight resource with latest Black Duck instances
func (p *OpsSightBlackDuckReconciler) updateOpsSight(opsSight *synopsysv1.OpsSight) error {
	var err error
	if !strings.EqualFold(opsSight.Status.State, "stopped") && !strings.EqualFold(opsSight.Status.State, "error") {
		for j := 0; j < 20; j++ {
			if strings.EqualFold(opsSight.Status.State, "running") {
				break
			}
			p.Log.V(1).Info("waiting for opssight to be up.....", "name", opsSight.Name)
			time.Sleep(10 * time.Second)

			if err := p.Client.Get(context.TODO(), types.NamespacedName{Name: opsSight.Name, Namespace: opsSight.Namespace}, opsSight); err != nil {
				return fmt.Errorf("unable to get opssight %s due to %+v", opsSight.Name, err)
			}
		}
		err = p.update(opsSight)
	}
	return err
}

// update will list all Black Ducks in the cluster, and send them to opssight as scan targets.
func (p *OpsSightBlackDuckReconciler) update(opsSight *synopsysv1.OpsSight) error {
	blackDuckType := opsSight.Spec.Blackduck.BlackduckSpec.Type

	blackDuckPassword, err := controllers_utils.Base64Decode(opsSight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return fmt.Errorf("unable to decode blackDuckPassword for %s OpsSight instance due to %+v", opsSight.Name, err)
	}

	allHubs := p.GetAllBlackDucks(blackDuckType, blackDuckPassword)

	err = p.updateOpsSightCRD(opsSight, allHubs)
	if err != nil {
		return err
	}
	return nil
}

// GetAllBlackDucks get only the internal Black Duck instances from the cluster
func (p *OpsSightBlackDuckReconciler) GetAllBlackDucks(blackDuckType string, blackDuckPassword string) []*synopsysv1.Host {
	hosts := []*synopsysv1.Host{}
	blackDucks := synopsysv1.BlackduckList{}
	if err := p.Client.List(context.TODO(), &blackDucks); err != nil {
		p.Log.Error(err, "unable to list blackduck instances")
	}

	for _, blackDuck := range blackDucks.Items {
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
			host := &synopsysv1.Host{
				Domain:              fmt.Sprintf("%s.%s.svc", controllers_utils.GetResourceName(blackDuck.Name, controllers_utils.BLACKDUCK, "webserver"), blackDuck.Spec.Namespace),
				ConcurrentScanLimit: concurrentScanLimit,
				Scheme:              "https",
				User:                "sysadmin",
				Port:                443,
				Password:            blackDuckPassword,
			}
			hosts = append(hosts, host)
		}
	}
	p.Log.V(1).Info("total no of Black Duck's", "blackDuckType", blackDuckType, "numbers", len(hosts))
	return hosts
}

// updateOpsSightCRD will update the opssight CRD
func (p *OpsSightBlackDuckReconciler) updateOpsSightCRD(opsSight *synopsysv1.OpsSight, blackDucks []*synopsysv1.Host) error {
	opssightName := opsSight.Name
	opsSightNamespace := opsSight.Namespace

	if err := p.Client.Get(context.TODO(), types.NamespacedName{Name: opssightName, Namespace: opsSightNamespace}, opsSight); err != nil {
		return fmt.Errorf("unable to get OpsSight instance %s due to %+v", opssightName, err)
	}

	opsSight.Status.InternalHosts = p.AppendBlackDuckHosts(opsSight.Status.InternalHosts, blackDucks)

	if err := p.Client.Update(context.TODO(), opsSight); err != nil {
		return fmt.Errorf("unable to update OpsSight instance %s due to %+v", opssightName, err)
	}

	return nil
}

// AppendBlackDuckHosts will append the old and new internal Black Duck hosts
func (p *OpsSightBlackDuckReconciler) AppendBlackDuckHosts(oldBlackDucks []*synopsysv1.Host, newBlackDucks []*synopsysv1.Host) []*synopsysv1.Host {
	existingBlackDucks := make(map[string]*synopsysv1.Host)
	for _, oldBlackDuck := range oldBlackDucks {
		existingBlackDucks[oldBlackDuck.Domain] = oldBlackDuck
	}

	finalBlackDucks := []*synopsysv1.Host{}
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
func (p *OpsSightBlackDuckReconciler) AppendBlackDuckSecrets(existingExternalBlackDucks map[string]*synopsysv1.Host, oldInternalBlackDucks []*synopsysv1.Host, newInternalBlackDucks []*synopsysv1.Host) map[string]*synopsysv1.Host {
	existingInternalBlackducks := make(map[string]*synopsysv1.Host)
	for _, oldInternalBlackDuck := range oldInternalBlackDucks {
		existingInternalBlackducks[oldInternalBlackDuck.Domain] = oldInternalBlackDuck
	}

	currentInternalBlackducks := make(map[string]*synopsysv1.Host)
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
