/*
Copyright (C) 2019 Synopsys, Inc.

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

package alert

import (
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// getAlertConfigMap returns a new ConfigMap for an Alert
func (a *SpecConfig) getAlertConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "blackduck-config"),
		Namespace: a.alert.Spec.Namespace,
	})

	// Add Environs
	configMapData := map[string]string{}
	for _, environ := range a.alert.Spec.Environs {
		vals := strings.Split(environ, ":")
		if len(vals) != 2 {
			log.Errorf("Could not split environ '%s' on ':'", environ)
			continue
		}
		environKey := strings.TrimSpace(vals[0])
		environVal := strings.TrimSpace(vals[1])
		log.Debugf("Adding Environ %s", environKey)
		configMapData[environKey] = environVal
	}

	// Add data to the ConfigMap
	configMap.AddData(configMapData)

	configMap.AddLabels(map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"})

	return configMap
}
