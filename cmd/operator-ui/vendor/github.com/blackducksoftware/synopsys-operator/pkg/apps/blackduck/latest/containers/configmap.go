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

package containers

import (
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetConfigmaps will return the configMaps
func (c *Creater) GetConfigmaps() []*components.ConfigMap {

	var configMaps []*components.ConfigMap

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.hubSpec.Namespace, Name: "blackduck-config"})
	hubData := map[string]string{
		"RUN_SECRETS_DIR": "/tmp/secrets",
		"HUB_VERSION":     c.hubSpec.Version,
	}

	for _, value := range c.hubSpec.Environs {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				hubData[mapKey] = mapValue
			}
		}
	}

	// merge default and input environs
	environs, _ := GetHubKnobs()
	hubData = util.MergeEnvMaps(hubData, environs)

	hubConfig.AddData(hubData)
	hubConfig.AddLabels(c.GetVersionLabel("configmap"))
	configMaps = append(configMaps, hubConfig)

	return configMaps
}
