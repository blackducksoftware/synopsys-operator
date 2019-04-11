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

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	hubutils "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// PostEditContainer ...
func (c *Creater) PostEditContainer(cc *util.Container) {
	if c.getUID(cc.ContainerConfig.Name) != nil {
		cc.ContainerConfig.UID = c.getUID(cc.ContainerConfig.Name)
		log.Infof("Image UID %v was tag modded to %v", cc.ContainerConfig.Name, cc.ContainerConfig.UID)
	}
}

// Creater will store the configuration to create the hub containers
type Creater struct {
	config             *protoform.Config
	hubSpec            *v1.BlackduckSpec
	hubContainerFlavor *ContainerFlavor
}

// NewCreater will return a creater
func NewCreater(config *protoform.Config, hubSpec *v1.BlackduckSpec, hubContainerFlavor *ContainerFlavor) *Creater {
	return &Creater{config: config, hubSpec: hubSpec, hubContainerFlavor: hubContainerFlavor}
}

// GetFullContainerNameFromImageRegistryConf returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) GetFullContainerNameFromImageRegistryConf(baseContainer string) string {
	//blackduckVersion := hubutils.GetHubVersion(c.hubSpec.Environs)
	for _, reg := range c.hubSpec.ImageRegistries {
		// normal case: we expect registries
		if strings.Contains(reg, baseContainer) {
			_, err := hubutils.ParseImageString(reg)
			if err != nil {
				log.Error(err)
				break
			}
			return reg
		}
	}

	return ""
}

// getTag returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) getUID(baseContainer string) *int64 {
	if tag, ok := c.hubSpec.ImageUIDMap[baseContainer]; ok {
		return &tag
	}
	return nil
}
