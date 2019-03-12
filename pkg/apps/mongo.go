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

package apps

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

const (
	// mongoName will be Mongo container name
	mongoName = "mongodb"
	// mongoDataMountPath will be Mongo data mount path
	mongoDataMountPath = "/data/db/"
	// mongoDataVolumeName will be Mongo data volume name
	mongoDataVolumeName = "mongodb-data"
)

// Mongo will provide the Mongo container configuration
type Mongo struct {
	Namespace              string
	PVCName                string
	Port                   string
	Image                  string
	MinCPU                 string
	MaxCPU                 string
	MinMemory              string
	MaxMemory              string
	Database               string
	User                   string
	PasswordSecretName     string
	UserPasswordSecretKey  string
	AdminPasswordSecretKey string
	EnvConfigMapRefs       []string
}

// GetMongoReplicationController will return the Mongo replication controller
func (p *Mongo) GetMongoReplicationController() *components.ReplicationController {
	mongoEnvs := p.getMongoEnvconfigs()
	mongoVolumes := p.getMongoVolumes()
	mongoVolumeMounts := p.getMongoVolumeMounts()

	mongoExternalContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       mongoName,
			Image:      p.Image,
			PullPolicy: horizonapi.PullIfNotPresent,
			MinMem:     p.MinMemory,
			MaxMem:     p.MaxMemory,
			MinCPU:     p.MinCPU,
			MaxCPU:     p.MaxCPU,
		},
		EnvConfigs:   mongoEnvs,
		VolumeMounts: mongoVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{Name: mongoName, ContainerPort: p.Port, Protocol: horizonapi.ProtocolTCP}},
	}
	var initContainers []*util.Container
	if len(p.PVCName) > 0 {
		mongoInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine",
				Command: []string{"sh", "-c", fmt.Sprintf("chmod -cR 777 %s", mongoDataMountPath)}},
			VolumeMounts: mongoVolumeMounts,
			PortConfig:   []*horizonapi.PortConfig{{ContainerPort: "3001", Protocol: horizonapi.ProtocolTCP}},
		}
		initContainers = append(initContainers, mongoInitContainerConfig)
	}

	mongo := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: p.Namespace,
		Name: mongoName, Replicas: util.IntToInt32(1)}, "", []*util.Container{mongoExternalContainerConfig},
		mongoVolumes, initContainers, []horizonapi.AffinityConfig{})

	return mongo
}

// GetMongoService will return the Mongo service
func (p *Mongo) GetMongoService() *components.Service {
	return util.CreateService(mongoName, mongoName, p.Namespace, p.Port, p.Port, horizonapi.ClusterIPServiceTypeDefault)
}

// getMongoVolumes will return the Mongo volumes
func (p *Mongo) getMongoVolumes() []*components.Volume {
	var mongoVolumes []*components.Volume
	var mongoDataVolume *components.Volume
	if len(p.PVCName) > 0 {
		mongoDataVolume, _ = util.CreatePersistentVolumeClaimVolume(mongoDataVolumeName, p.PVCName)
	} else {
		mongoDataVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit(mongoDataVolumeName)
	}

	mongoVolumes = append(mongoVolumes, mongoDataVolume)
	return mongoVolumes
}

// getMongoVolumeMounts will return the Mongo volume mount configurations
func (p *Mongo) getMongoVolumeMounts() []*horizonapi.VolumeMountConfig {
	var mongoVolumeMounts []*horizonapi.VolumeMountConfig
	mongoVolumeMounts = append(mongoVolumeMounts, &horizonapi.VolumeMountConfig{Name: mongoDataVolumeName, MountPath: mongoDataMountPath})
	return mongoVolumeMounts
}

// getMongoEnvconfigs will return the Mongo environment variable configurations
func (p *Mongo) getMongoEnvconfigs() []*horizonapi.EnvConfig {
	mongoEnvs := []*horizonapi.EnvConfig{}
	if len(p.Database) > 0 {
		mongoEnvs = append(mongoEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_DATABASE", KeyOrVal: p.Database})
	}

	if len(p.User) > 0 {
		mongoEnvs = append(mongoEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "MONGODB_USER", KeyOrVal: p.User})
	}

	if len(p.UserPasswordSecretKey) > 0 && len(p.PasswordSecretName) > 0 {
		mongoEnvs = append(mongoEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MONGODB_PASSWORD", KeyOrVal: p.UserPasswordSecretKey, FromName: p.PasswordSecretName})
	}

	if len(p.AdminPasswordSecretKey) > 0 && len(p.PasswordSecretName) > 0 {
		mongoEnvs = append(mongoEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "MONGODB_ADMIN_PASSWORD", KeyOrVal: p.AdminPasswordSecretKey, FromName: p.PasswordSecretName})

	}

	for _, EnvConfigMapRef := range p.EnvConfigMapRefs {
		mongoEnvs = append(mongoEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: EnvConfigMapRef})
	}
	return mongoEnvs
}
