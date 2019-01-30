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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetPostgresDeployment will return the postgres deployment
func (c *Creater) GetPostgresDeployment() *components.ReplicationController {
	postgresEnvs := c.getPostgresEnvConfigs()
	postgresVolumes := c.getPostgresVolumes()
	postgresVolumeMounts := c.getPostgresVolumeMounts()

	postgresImage := c.getFullContainerName("postgres")
	if len(postgresImage) == 0 {
		postgresImage = "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1"
	}

	postgresExternalContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "postgres",
			Image:      postgresImage,
			PullPolicy: horizonapi.PullAlways,
			MinMem:     c.hubContainerFlavor.PostgresMemoryLimit,
			MaxMem:     "",
			MinCPU:     c.hubContainerFlavor.PostgresCPULimit,
			MaxCPU:     "",
		},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: postgresPort, Protocol: horizonapi.ProtocolTCP}},
	}
	var initContainers []*util.Container
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-postgres") {
		postgresInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /var/lib/pgsql/data"}},
			VolumeMounts:    postgresVolumeMounts,
			PortConfig:      []*horizonapi.PortConfig{{ContainerPort: "3001", Protocol: horizonapi.ProtocolTCP}},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}

	postgres := util.CreateReplicationControllerFromContainer(&horizonapi.ReplicationControllerConfig{Namespace: c.hubSpec.Namespace, Name: "postgres", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{postgresExternalContainerConfig}, postgresVolumes, initContainers, []horizonapi.AffinityConfig{})

	return postgres
}

// GetPostgresService will return the postgres service
func (c *Creater) GetPostgresService() *components.Service {
	return util.CreateService("postgres", "postgres", c.hubSpec.Namespace, postgresPort, postgresPort, horizonapi.ClusterIPServiceTypeDefault)
}

// getPostgresVolumes will return the postgres volumes
func (c *Creater) getPostgresVolumes() []*components.Volume {
	var postgresVolumes []*components.Volume
	var postgresDataVolume *components.Volume
	if c.hubSpec.PersistentStorage && c.hasPVC("blackduck-postgres") {
		postgresDataVolume, _ = util.CreatePersistentVolumeClaimVolume("postgres-data-volume", "blackduck-postgres")
	} else {
		postgresDataVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit("postgres-data-volume")
	}

	postgresVolumes = append(postgresVolumes, postgresDataVolume)
	return postgresVolumes
}

// getPostgresVolumeMounts will return the postgres volume mounts
func (c *Creater) getPostgresVolumeMounts() []*horizonapi.VolumeMountConfig {
	var postgresVolumeMounts []*horizonapi.VolumeMountConfig
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-data-volume", MountPath: "/var/lib/pgsql/data"})
	return postgresVolumeMounts
}

// getPostgresEnvConfigs will return the postgres environment config maps
func (c *Creater) getPostgresEnvConfigs() []*horizonapi.EnvConfig {
	postgresEnvs := c.allConfigEnv
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: "HUB_POSTGRES_USER_PASSWORD_FILE", FromName: "db-creds"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", FromName: "db-creds"})
	return postgresEnvs
}
