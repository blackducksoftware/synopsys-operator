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
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetConfigmaps will return the configMaps
func (c *Creater) GetConfigmaps() []*components.ConfigMap {

	var configMaps []*components.ConfigMap

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.blackDuck.Spec.Namespace, Name: util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "config")})

	hubData := map[string]string{
		"RUN_SECRETS_DIR": "/tmp/secrets",
		"HUB_VERSION":     c.blackDuck.Spec.Version,
	}

	blackduckServiceData := map[string]string{
		"HUB_AUTHENTICATION_HOST": util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "authentication"),
		"AUTHENTICATION_HOST":     fmt.Sprintf("%s:%d", util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "authentication"), authenticationPort),
		"CLIENT_CERT_CN":          util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "binaryscanner"),
		"CFSSL":                   fmt.Sprintf("%s:8888", util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl")),
		"HUB_CFSSL_HOST":          util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_HOST":    util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_PORT":    "8888",
		"HUB_DOC_HOST":            util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "documentation"),
		"HUB_JOBRUNNER_HOST":      util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "jobrunner"),
		"HUB_LOGSTASH_HOST":       util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "logstash"),
		"RABBIT_MQ_HOST":          util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "rabbitmq"),
		"BROKER_URL":              fmt.Sprintf("amqps://%s/protecodesc", util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "rabbitmq")),
		"HUB_REGISTRATION_HOST":   util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "registration"),
		"HUB_SCAN_HOST":           util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "scan"),
		"HUB_SOLR_HOST":           util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "solr"),
		// TODO: commented the below 2 environs until the HUB-20412 is fixed. once it if fixed, uncomment them
		// "BLACKDUCK_UPLOAD_CACHE_HOST": util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// "HUB_UPLOAD_CACHE_HOST":       util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// TODO: commented the below environs until the HUB-20462 is fixed. once it if fixed, uncomment them
		// "HUB_WEBAPP_HOST":    util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webapp"),
		"HUB_WEBSERVER_HOST": util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "webserver"),
		"HUB_ZOOKEEPER_HOST": util.GetResourceName(c.blackDuck.Name, util.BlackDuckName, "zookeeper"),
	}
	hubData = util.MergeEnvMaps(blackduckServiceData, hubData)

	for _, value := range c.blackDuck.Spec.Environs {
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
	environs := GetHubKnobs()
	hubData = util.MergeEnvMaps(hubData, environs)

	hubConfig.AddData(hubData)
	hubConfig.AddLabels(c.GetVersionLabel("configmap"))
	configMaps = append(configMaps, hubConfig)

	return configMaps
}
