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
	"strings"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the hub containers
type Creater struct {
	config                  *protoform.Config
	kubeConfig              *rest.Config
	kubeClient              *kubernetes.Clientset
	blackDuck               *blackduckapi.Blackduck
	hubContainerFlavor      *ContainerFlavor
	isBinaryAnalysisEnabled bool
}

// NewCreater will return a creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, blackDuck *blackduckapi.Blackduck,
	hubContainerFlavor *ContainerFlavor, isBinaryAnalysisEnabled bool) *Creater {
	return &Creater{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, blackDuck: blackDuck, hubContainerFlavor: hubContainerFlavor,
		isBinaryAnalysisEnabled: isBinaryAnalysisEnabled}
}

// GetFullContainerNameFromImageRegistryConf returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) GetFullContainerNameFromImageRegistryConf(baseContainer string) string {
	//blackduckVersion := hubutils.GetHubVersion(c.blackDuck.Spec.Environs)
	for _, reg := range c.blackDuck.Spec.ImageRegistries {
		// normal case: we expect registries
		if strings.Contains(reg, baseContainer) {
			log.Infof("Image %v found inside of the [ %v ] tag map. Returning %v as the container name for %v.", reg, c.blackDuck.Spec.ImageRegistries, reg, baseContainer)
			_, err := util.ValidateImageString(reg)
			if err != nil {
				log.Error(err)
				break
			}
			return reg
		}
	}

	//ignoredContainers := []string{"postgres", "appcheck", "rabbitmq", "upload"}
	//for _, ignoredContainer := range ignoredContainers {
	//	if strings.EqualFold(baseContainer, ignoredContainer) {
	//		return ""
	//	}
	//}
	//
	//if strings.EqualFold(baseContainer, "solr") && strings.HasPrefix(blackduckVersion, "20") {
	//	return ""
	//}
	//
	//img := fmt.Sprintf("docker.io/blackducksoftware/hub-%v:%v", baseContainer, blackduckVersion)
	//log.Warnf("Couldn't get container name for : %v, set it manually in the deployment, returning a reasonable default instead %v.", baseContainer, img)
	//log.Warn("In the future, you should provide fully qualified images for every single container when running the blackduck operator.")
	//return img
	return ""
}
