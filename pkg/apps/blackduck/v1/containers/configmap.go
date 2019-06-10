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

package containers

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetConfigmaps will create  hub configMaps
func (c *Creater) GetConfigmaps() []*components.ConfigMap {

	var configMaps []*components.ConfigMap

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.hubSpec.Namespace, Name: util.GetResourceName(c.name, "blackduck-config", c.isClusterScope)})

	hubData := map[string]string{
		"RUN_SECRETS_DIR": "/tmp/secrets",
		"HUB_VERSION":     c.hubSpec.Version,
	}

	if !c.isClusterScope {
		blackduckServiceData := map[string]string{
			// TODO: commented the below 2 environs until the HUB-20412 is fixed. once it if fixed, uncomment them
			// "HUB_AUTHENTICATION_HOST": util.GetResourceName(c.name, "authentication", c.isClusterScope),
			"CLIENT_CERT_CN":        util.GetResourceName(c.name, "binaryscanner", c.isClusterScope),
			"CFSSL":                 fmt.Sprintf("%s:8888", util.GetResourceName(c.name, "cfssl", c.isClusterScope)),
			"HUB_CFSSL_HOST":        util.GetResourceName(c.name, "cfssl", c.isClusterScope),
			"BLACKDUCK_CFSSL_HOST":  util.GetResourceName(c.name, "cfssl", c.isClusterScope),
			"HUB_DOC_HOST":          util.GetResourceName(c.name, "documentation", c.isClusterScope),
			"HUB_JOBRUNNER_HOST":    util.GetResourceName(c.name, "jobrunner", c.isClusterScope),
			"HUB_LOGSTASH_HOST":     util.GetResourceName(c.name, "logstash", c.isClusterScope),
			"RABBIT_MQ_HOST":        util.GetResourceName(c.name, "rabbitmq", c.isClusterScope),
			"HUB_REGISTRATION_HOST": util.GetResourceName(c.name, "registration", c.isClusterScope),
			"HUB_SCAN_HOST":         util.GetResourceName(c.name, "scan", c.isClusterScope),
			"HUB_SOLR_HOST":         util.GetResourceName(c.name, "solr", c.isClusterScope),
			// TODO: commented the below 2 environs until the HUB-20412 is fixed. once it if fixed, uncomment them
			// "BLACKDUCK_UPLOAD_CACHE_HOST": util.GetResourceName(c.name, "uploadcache", c.isClusterScope),
			// "HUB_UPLOAD_CACHE_HOST":       util.GetResourceName(c.name, "uploadcache", c.isClusterScope),
			"HUB_WEBAPP_HOST":    util.GetResourceName(c.name, "webapp", c.isClusterScope),
			"HUB_WEBSERVER_HOST": util.GetResourceName(c.name, "webserver", c.isClusterScope),
			"HUB_ZOOKEEPER_HOST": util.GetResourceName(c.name, "zookeeper", c.isClusterScope),
		}
		hubData = util.MergeEnvMaps(blackduckServiceData, hubData)
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

	environs := GetHubKnobs()
	hubData = util.MergeEnvMaps(hubData, environs)
	hubConfig.AddData(hubData)
	hubConfig.AddLabels(c.GetVersionLabel("configmap"))
	configMaps = append(configMaps, hubConfig)

	return configMaps
}
