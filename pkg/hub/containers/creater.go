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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/sirupsen/logrus"
)

// TagGetterInterface ...
type TagGetterInterface interface {
	getTag(s string) string
	getUID(s string) *int64
}

// PostEditContainer edits to containers (like tags, uids) go here!
func (c *Creater) PostEditContainer(cc *util.Container) {
	PostEdit(cc, c)
}

// PostEdit ...
func PostEdit(cc *util.Container, c TagGetterInterface) {
	// Replace the tag with any tag maps to individual containers.
	// This is the "joe gamache wants to try a rogue jobrunner" feature.
	if c.getTag(cc.ContainerConfig.Name) != "" {
		fields := strings.Split(cc.ContainerConfig.Image, ":")
		if c.getTag(cc.ContainerConfig.Name) != "" {
			tagIndex := len(fields) - 1
			tagValue := c.getTag(cc.ContainerConfig.Name)
			fields[tagIndex] = tagValue
			//rejoin the split tags
			image := strings.Join(fields, ":")
			cc.ContainerConfig.Image = image
			logrus.Infof("Image for %v was tag modded to %v", cc.ContainerConfig.Name, cc.ContainerConfig.Image)
		}
	}
	if c.getUID(cc.ContainerConfig.Name) != nil {
		cc.ContainerConfig.UID = c.getUID(cc.ContainerConfig.Name)
		logrus.Infof("Image UID %v was tag modded to %v", cc.ContainerConfig.Name, cc.ContainerConfig.UID)
	}
}

// Creater will store the configuration to create the hub containers
type Creater struct {
	config             *protoform.Config
	hubSpec            *v1.HubSpec
	hubContainerFlavor *ContainerFlavor
	hubConfigEnv       []*horizonapi.EnvConfig
	allConfigEnv       []*horizonapi.EnvConfig
	dbSecretVolume     *components.Volume
	dbEmptyDir         *components.Volume
	proxySecretVolume  *components.Volume
	containerTags      map[string]string
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, hubSpec *v1.HubSpec, hubContainerFlavor *ContainerFlavor, hubConfigEnv []*horizonapi.EnvConfig, allConfigEnv []*horizonapi.EnvConfig,
	dbSecretVolume *components.Volume, dbEmptyDir *components.Volume, proxySecretVolume *components.Volume) *Creater {
	containerTags := hubSpec.ImageTagMap
	imageTags := map[string]string{}
	for _, containerTag := range containerTags {
		tags := strings.SplitN(containerTag, ":", 2)
		if len(tags) == 2 {
			imageTags[strings.Trim(tags[0], " ")] = strings.Trim(tags[1], " ")
		}
	}
	return &Creater{
		config:             config,
		hubSpec:            hubSpec,
		hubContainerFlavor: hubContainerFlavor,
		hubConfigEnv:       hubConfigEnv,
		allConfigEnv:       allConfigEnv,
		dbSecretVolume:     dbSecretVolume,
		dbEmptyDir:         dbEmptyDir,
		proxySecretVolume:  proxySecretVolume,
		containerTags:      imageTags,
	}
}

// getTag returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) getTag(baseContainer string) string {
	if tag, ok := c.containerTags[baseContainer]; ok {
		return tag
	}
	return c.hubSpec.HubVersion
}

// getTag returns the tag that is specified for a container by trying to look in the custom tags provided,
// if those arent filled, it uses the "HubVersion" as a default, which works for blackduck < 5.1.0.
func (c *Creater) getUID(baseContainer string) *int64 {
	if tag, ok := c.hubSpec.ImageUIDMap[baseContainer]; ok {
		return &tag
	}
	return nil
}
