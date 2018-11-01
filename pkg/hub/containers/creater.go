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
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
)

// Creater will store the configuration to create the hub containers
type Creater struct {
	config             *protoform.Config
	hubSpec            *v1.HubSpec
	hubContainerFlavor *ContainerFlavor
	hubConfigEnv       []*horizonapi.EnvConfig
	allConfigEnv       []*horizonapi.EnvConfig
	dbSecretVolume     *components.Volume
	dbEmptyDir         *components.Volume
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, hubSpec *v1.HubSpec, hubContainerFlavor *ContainerFlavor, hubConfigEnv []*horizonapi.EnvConfig, allConfigEnv []*horizonapi.EnvConfig,
	dbSecretVolume *components.Volume, dbEmptyDir *components.Volume) *Creater {
	return &Creater{config: config, hubSpec: hubSpec, hubContainerFlavor: hubContainerFlavor, hubConfigEnv: hubConfigEnv, allConfigEnv: allConfigEnv, dbSecretVolume: dbSecretVolume,
		dbEmptyDir: dbEmptyDir}
}

// getTag returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) getTag(baseContainer string) string {
	if tag, ok := c.hubSpec.ImageTagMap[baseContainer]; ok {
		return tag
	}
	return c.hubSpec.HubVersion
}
