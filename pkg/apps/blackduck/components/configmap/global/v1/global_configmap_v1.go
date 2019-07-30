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

package v1

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// BdConfigmap holds the Black Duck config map configuration
type BdConfigmap struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

func init() {
	store.Register(types.BlackDuckGlobalConfigmapV1, NewBdConfigmap)
}

// NewBdConfigmap returns the Black Duck config map configuration
func NewBdConfigmap(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.ConfigMapInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdConfigmap{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

// GetCM returns the config map
func (b *BdConfigmap) GetCM() (*components.ConfigMap, error) {

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: b.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "config")})

	hubData := map[string]string{
		"RUN_SECRETS_DIR": "/tmp/secrets",
		"HUB_VERSION":     b.blackDuck.Spec.Version,
	}

	blackduckServiceData := map[string]string{
		// TODO: commented the below 2 environs until the HUB-20482 is fixed. once it if fixed, uncomment them
		//"HUB_AUTHENTICATION_HOST": util.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "authentication"),
		//"AUTHENTICATION_HOST":     fmt.Sprintf("%s:%d", util.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "authentication"), int32(8443)),
		"CLIENT_CERT_CN":        apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "binaryscanner"),
		"CFSSL":                 fmt.Sprintf("%s:8888", apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "cfssl")),
		"HUB_CFSSL_HOST":        apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_HOST":  apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "cfssl"),
		"BLACKDUCK_CFSSL_PORT":  "8888",
		"HUB_DOC_HOST":          apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "documentation"),
		"HUB_JOBRUNNER_HOST":    apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "jobrunner"),
		"HUB_LOGSTASH_HOST":     apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "logstash"),
		"RABBIT_MQ_HOST":        apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "rabbitmq"),
		"BROKER_URL":            fmt.Sprintf("amqps://%s/protecodesc", apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "rabbitmq")),
		"HUB_REGISTRATION_HOST": apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "registration"),
		"HUB_SCAN_HOST":         apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "scan"),
		"HUB_SOLR_HOST":         apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "solr"),
		// TODO: commented the below 2 environs until the HUB-20412 is fixed. once it if fixed, uncomment them
		// "BLACKDUCK_UPLOAD_CACHE_HOST": util.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// "HUB_UPLOAD_CACHE_HOST":       util.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "uploadcache"),
		// TODO: commented the below environs until the HUB-20462 is fixed. once it if fixed, uncomment them
		// "HUB_WEBAPP_HOST":    util.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "webapp"),
		"HUB_WEBSERVER_HOST": apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "webserver"),
		"HUB_ZOOKEEPER_HOST": apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "zookeeper"),
	}
	hubData = util.MergeEnvMaps(blackduckServiceData, hubData)

	for _, value := range b.blackDuck.Spec.Environs {
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
	environs := GetBlackDuckKnobs()
	hubData = util.MergeEnvMaps(hubData, environs)

	hubConfig.AddData(hubData)
	hubConfig.AddLabels(apputils.GetVersionLabel("configmap", b.blackDuck.Name, b.blackDuck.Spec.Version))

	return hubConfig, nil
}

// GetBlackDuckKnobs returns the default Black Duck knobs
func GetBlackDuckKnobs() map[string]string {
	return map[string]string{
		"IPV4_ONLY":                         "0",
		"USE_ALERT":                         "0",
		"USE_BINARY_UPLOADS":                "0",
		"RABBIT_MQ_PORT":                    "5671",
		"BROKER_USE_SSL":                    "yes",
		"SCANNER_CONCURRENCY":               "1",
		"HTTPS_VERIFY_CERTS":                "yes",
		"RABBITMQ_DEFAULT_VHOST":            "protecodesc",
		"RABBITMQ_SSL_FAIL_IF_NO_PEER_CERT": "false",
		"ENABLE_SOURCE_UPLOADS":             "false",
		"DATA_RETENTION_IN_DAYS":            "180",
		"MAX_TOTAL_SOURCE_SIZE_MB":          "4000",
	}
}
