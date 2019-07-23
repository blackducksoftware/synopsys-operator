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

package alert

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// Creater stores the configuration and clients to create specific versions of Alerts
type Creater struct {
	protoformDeployer *protoform.Deployer
}

// NewCreater returns this Alert Creater
func NewCreater(protoformDeployer *protoform.Deployer) *Creater {
	return &Creater{protoformDeployer: protoformDeployer}
}

// GetComponents returns the resource components for an Alert
func (ac *Creater) GetComponents(alert *alertapi.Alert) (*api.ComponentList, error) {
	specConfig := NewSpecConfig(alert, ac.protoformDeployer.Config.IsClusterScoped)
	return specConfig.GetComponents()
}

// GetPVC returns the Persistent Volume Claims for an Alert
func (ac *Creater) GetPVC(alert *alertapi.Alert) ([]*components.PersistentVolumeClaim, error) {
	specConfig := NewSpecConfig(alert, ac.protoformDeployer.Config.IsClusterScoped)
	pvc, err := specConfig.getAlertPersistentVolumeClaim()
	return []*components.PersistentVolumeClaim{pvc}, err
}

// Versions is an Interface function that returns the versions supported by this Creater
func (ac *Creater) Versions() []string {
	return GetVersions()
}

// Ensure is an Interface function that will make sure the instance is correctly deployed or deploy it if needed
func (ac *Creater) Ensure(alert *alertapi.Alert) error {
	// Get Kubernetes Components for the Alert
	specConfig := NewSpecConfig(alert, ac.protoformDeployer.Config.IsClusterScoped)
	cpList, err := specConfig.GetComponents()
	if err != nil {
		return err
	}
	if strings.EqualFold(alert.Spec.DesiredState, "STOP") {
		commonConfig := crdupdater.NewCRUDComponents(ac.protoformDeployer.KubeConfig, ac.protoformDeployer.KubeClient, ac.protoformDeployer.Config.DryRun, false, alert.Spec.Namespace, alert.Spec.Version,
			&api.ComponentList{PersistentVolumeClaims: cpList.PersistentVolumeClaims}, fmt.Sprintf("app=%s,name=%s", util.AlertName, alert.Name), false)
		_, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("unable to stop Alert: %+v", errors)
		}
	} else {
		// Update components in cluster
		commonConfig := crdupdater.NewCRUDComponents(ac.protoformDeployer.KubeConfig, ac.protoformDeployer.KubeClient, ac.protoformDeployer.Config.DryRun, false, alert.Spec.Namespace, alert.Spec.Version,
			cpList, fmt.Sprintf("app=%s,name=%s", util.AlertName, alert.Name), false)
		_, errors := commonConfig.CRUDComponents()
		if len(errors) > 0 {
			return fmt.Errorf("unable to update Alert components due to %+v", errors)
		}
	}
	return nil
}
