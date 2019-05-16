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

package postgres

import (
	"fmt"
	"strconv"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

const (
	// postgresName will be Postgres container name
	postgresName = "postgres"
	// postgresDataMountPath will be Postgres data mount path
	postgresDataMountPath = "/var/lib/pgsql/data"
	// postgresDataVolumeName will be Postgres data volume name
	postgresDataVolumeName = "postgres-data-volume"
)

// Postgres will provide the postgres container configuration
type Postgres struct {
	Namespace                     string
	PVCName                       string
	Port                          int32
	Image                         string
	MinCPU                        string
	MaxCPU                        string
	MinMemory                     string
	MaxMemory                     string
	Database                      string
	User                          string
	PasswordSecretName            string
	UserPasswordSecretKey         string
	AdminPasswordSecretKey        string
	MaxConnections                int
	SharedBufferInMB              int
	EnvConfigMapRefs              []string
	TerminationGracePeriodSeconds int64
	Labels                        map[string]string
}

// GetPostgresReplicationController will return the postgres replication controller
func (p *Postgres) GetPostgresReplicationController() (*components.ReplicationController, error) {
	postgresEnvs := p.getPostgresEnvconfigs()
	postgresVolumes := p.getPostgresVolumes()
	postgresVolumeMounts := p.getPostgresVolumeMounts()

	postgresExternalContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       postgresName,
			Image:      p.Image,
			PullPolicy: horizonapi.PullIfNotPresent,
			MinMem:     p.MinMemory,
			MaxMem:     p.MaxMemory,
			MinCPU:     p.MinCPU,
			MaxCPU:     p.MaxCPU,
		},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   []*horizonapi.PortConfig{{ContainerPort: p.Port, Protocol: horizonapi.ProtocolTCP}},
		PreStopConfig: &horizonapi.ActionConfig{
			Type:    horizonapi.ActionTypeCommand,
			Command: []string{"sh", "-c", "LD_LIBRARY_PATH=/opt/rh/rh-postgresql96/root/usr/lib64 /opt/rh/rh-postgresql96/root/usr/bin/pg_ctl -D /var/lib/pgsql/data/userdata -l logfile stop"},
		},
		ReadinessProbeConfigs: []*horizonapi.ProbeConfig{{
			ActionConfig: horizonapi.ActionConfig{
				Type: horizonapi.ActionTypeCommand,
				Command: []string{
					"/bin/bash",
					"-c",
					"/opt/rh/rh-postgresql96/root/usr/bin/pg_isready -h localhost"},
			},
			Delay:           5,
			Interval:        10,
			Timeout:         5,
			MinCountFailure: 10,
		}},
	}
	var initContainers []*util.Container
	if len(p.PVCName) > 0 {
		postgresInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine",
				Command: []string{"sh", "-c", fmt.Sprintf("chmod -cR 777 %s", postgresDataMountPath)}},
			VolumeMounts: postgresVolumeMounts,
			PortConfig:   []*horizonapi.PortConfig{{ContainerPort: 3001, Protocol: horizonapi.ProtocolTCP}},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}

	pod, err := util.CreatePod(
		&util.PodConfig{
			Name:           postgresName,
			Volumes:        postgresVolumes,
			Containers:     []*util.Container{postgresExternalContainerConfig},
			InitContainers: initContainers,
			Labels:         p.Labels,
		})

	if err != nil {
		return nil, fmt.Errorf("%+v", err)
	}
	// increase TerminationGracePeriod to better handle pg shutdown
	pod.Spec.TerminationGracePeriodSeconds = &p.TerminationGracePeriodSeconds

	postgres := util.CreateReplicationController(&horizonapi.ReplicationControllerConfig{Namespace: p.Namespace,
		Name: postgresName, Replicas: util.IntToInt32(1)}, pod, p.Labels, p.Labels)

	return postgres, nil
}

// GetPostgresService will return the postgres service
func (p *Postgres) GetPostgresService() *components.Service {
	return util.CreateService(postgresName, p.Labels, p.Namespace, p.Port, p.Port, horizonapi.ServiceTypeServiceIP, p.Labels)
}

// getPostgresVolumes will return the postgres volumes
func (p *Postgres) getPostgresVolumes() []*components.Volume {
	var postgresVolumes []*components.Volume
	var postgresDataVolume *components.Volume
	if len(p.PVCName) > 0 {
		postgresDataVolume, _ = util.CreatePersistentVolumeClaimVolume(postgresDataVolumeName, p.PVCName)
	} else {
		postgresDataVolume, _ = util.CreateEmptyDirVolumeWithoutSizeLimit(postgresDataVolumeName)
	}

	postgresVolumes = append(postgresVolumes, postgresDataVolume)
	return postgresVolumes
}

// getPostgresVolumeMounts will return the postgres volume mount configurations
func (p *Postgres) getPostgresVolumeMounts() []*horizonapi.VolumeMountConfig {
	var postgresVolumeMounts []*horizonapi.VolumeMountConfig
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: postgresDataVolumeName, MountPath: postgresDataMountPath})
	return postgresVolumeMounts
}

// getPostgresEnvconfigs will return the postgres environment variable configurations
func (p *Postgres) getPostgresEnvconfigs() []*horizonapi.EnvConfig {
	postgresEnvs := []*horizonapi.EnvConfig{}
	if p.MaxConnections > 0 {
		postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_MAX_CONNECTIONS", KeyOrVal: strconv.Itoa(p.MaxConnections)})
	}
	if p.SharedBufferInMB > 0 {
		postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_SHARED_BUFFERS", KeyOrVal: fmt.Sprintf("%dMB", p.SharedBufferInMB)})
	}
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: p.Database})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: p.User})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: p.UserPasswordSecretKey, FromName: p.PasswordSecretName})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: p.AdminPasswordSecretKey, FromName: p.PasswordSecretName})
	for _, EnvConfigMapRef := range p.EnvConfigMapRefs {
		postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, FromName: EnvConfigMapRef})
	}
	return postgresEnvs
}
