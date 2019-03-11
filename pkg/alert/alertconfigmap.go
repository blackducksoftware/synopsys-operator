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

package alert

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// alertConfigMap creates a config map for alert
func (a *SpecConfig) alertConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "alert",
		Namespace: a.config.Namespace,
	})

	configMap.AddData(map[string]string{
		"ALERT_SERVER_PORT":         fmt.Sprintf("%d", *a.config.Port),
		"PUBLIC_HUB_WEBSERVER_HOST": a.config.BlackduckHost,
		"PUBLIC_HUB_WEBSERVER_PORT": fmt.Sprintf("%d", *a.config.BlackduckPort),
	})

	return configMap
}
