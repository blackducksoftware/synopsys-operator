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
	// name will be Postgres container name
	name = "postgres"
	// dataMountPath will be Postgres data mount path
	dataMountPath = "/var/lib/pgsql/data"
	// dataVolumeName will be Postgres data volume name
	dataVolumeName = "postgres-data-volume"
)

// Postgres will provide the postgres container configuration
type Postgres struct {
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

// GetPostgresReplicationController will return the postgres replication controller
func (p *Postgres) GetPostgresReplicationController() *components.ReplicationController {
	postgresEnvs := p.getPostgresEnvconfigs()
	postgresVolumes := p.getPostgresVolumes()
	postgresVolumeMounts := p.getPostgresVolumeMounts()

	postgresExternalContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       name,
			Image:      p.Image,
			PullPolicy: horizonapi.PullAlways,
			MinMem:     p.MinMemory,
			MaxMem:     p.MaxMemory,
			MinCPU:     p.MinCPU,
			MaxCPU:     p.MaxCPU,
		},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: p.Port, Protocol: horizonapi.ProtocolTCP}},
	}
	var initContainers []*util.Container
	if len(p.PVCName) > 0 {
		postgresInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine",
				Command: []string{"sh", "-c", fmt.Sprintf("chmod -cR 777 %s", dataMountPath)}},
			VolumeMounts: postgresVolumeMounts,
			PortConfig:   []*horizonapi.PortConfig{{ContainerPort: "3001", Protocol: horizonapi.ProtocolTCP}},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}

	postgres := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: p.Namespace,
		Name: name, Replicas: util.IntToInt32(1)}, "", []*util.Container{postgresExternalContainerConfig},
		postgresVolumes, initContainers, []horizonapi.AffinityConfig{})

	return postgres
}

// GetPostgresService will return the postgres service
func (p *Postgres) GetPostgresService() *components.Service {
	return util.CreateService(name, name, p.Namespace, p.Port, p.Port, horizonapi.ClusterIPServiceTypeDefault)
}

// getPostgresVolumes will return the postgres volumes
func (p *Postgres) getPostgresVolumes() []*components.Volume {
	var postgresVolumes []*components.Volume
	var postgresDataVolume *components.Volume
	if len(p.PVCName) > 0 {
		postgresDataVolume, _ = util.CreatePersistentVolumeClaimVolume(dataVolumeName, p.PVCName)
	} else {
		postgresDataVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit(dataVolumeName)
	}

	postgresVolumes = append(postgresVolumes, postgresDataVolume)
	return postgresVolumes
}

// getPostgresVolumeMounts will return the postgres volume mount configurations
func (p *Postgres) getPostgresVolumeMounts() []*horizonapi.VolumeMountConfig {
	var postgresVolumeMounts []*horizonapi.VolumeMountConfig
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: dataVolumeName, MountPath: dataMountPath})
	return postgresVolumeMounts
}

// getPostgresEnvconfigs will return the postgres environment variable configurations
func (p *Postgres) getPostgresEnvconfigs() []*horizonapi.EnvConfig {
	postgresEnvs := []*horizonapi.EnvConfig{}
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: p.Database})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: p.User})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: p.UserPasswordSecretKey, FromName: p.PasswordSecretName})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: p.AdminPasswordSecretKey, FromName: p.PasswordSecretName})
	for _, EnvConfigMapRef := range p.EnvConfigMapRefs {
		postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: EnvConfigMapRef})
	}
	return postgresEnvs
}
