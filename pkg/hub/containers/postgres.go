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
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
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
		ContainerConfig: &horizonapi.ContainerConfig{Name: "postgres", Image: postgresImage, PullPolicy: horizonapi.PullAlways,
			MinMem: c.hubContainerFlavor.PostgresMemoryLimit, MaxMem: "", MinCPU: c.hubContainerFlavor.PostgresCPULimit, MaxCPU: "", Command: []string{"/usr/share/container-scripts/postgresql/pginit.sh"},
		},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   &horizonapi.PortConfig{ContainerPort: postgresPort, Protocol: horizonapi.ProtocolTCP},
		ReadinessProbeConfigs: []*horizonapi.ProbeConfig{{
			ActionConfig:    horizonapi.ActionConfig{Command: []string{"/bin/sh", "-c", `if [[ -f /tmp/BLACKDUCK_MIGRATING ]] ; then cat /tmp/BLACKDUCK_MIGRATING1; fi`}},
			Delay:           1,
			Interval:        5,
			MinCountFailure: 100000,
			Timeout:         100000,
		}},
	}
	initContainers := []*util.Container{}
	// If the PV storage is other than NFS or if the backup is enabled and PV storage is other than NFS, add the init container
	if !strings.EqualFold(c.hubSpec.PVCStorageClass, "") && !strings.EqualFold(c.hubSpec.PVCStorageClass, "none") {
		postgresInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", fmt.Sprintf("chmod -cR 777 %s", c.config.NFSPath)}},
			VolumeMounts: []*horizonapi.VolumeMountConfig{
				{Name: "postgres-backup-vol", MountPath: c.config.NFSPath},
			},
			PortConfig: &horizonapi.PortConfig{ContainerPort: "3001", Protocol: horizonapi.ProtocolTCP},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}
	// c.PostEditContainer(postgresExternalContainerConfig)

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
	postgresVolumes := []*components.Volume{}
	postgresEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("postgres-persistent-vol")
	postgresInitConfigVol, _ := util.CreateConfigMapVolume("postgres-init-vol", "postgres-init", 0777)
	postgresBootstrapConfigVol, _ := util.CreateConfigMapVolume("postgres-bootstrap-vol", "postgres-bootstrap", 0777)
	postgresVolumes = append(postgresVolumes, postgresEmptyDir, postgresInitConfigVol, postgresBootstrapConfigVol)

	if strings.EqualFold(c.hubSpec.BackupSupport, "Yes") || !strings.EqualFold(c.hubSpec.DbPrototype, "empty") {
		postgresBackupDir, _ := util.CreatePersistentVolumeClaimVolume("postgres-backup-vol", c.hubSpec.Namespace)
		postgresVolumes = append(postgresVolumes, postgresBackupDir)
	}
	return postgresVolumes
}

// getPostgresVolumeMounts will return the postgres volume mounts
func (c *Creater) getPostgresVolumeMounts() []*horizonapi.VolumeMountConfig {
	postgresVolumeMounts := []*horizonapi.VolumeMountConfig{}
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-persistent-vol", MountPath: "/var/lib/pgsql/data"})
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-bootstrap-vol:pgbootstrap.sh", MountPath: "/usr/share/container-scripts/postgresql/pgbootstrap.sh"})
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-init-vol:pginit.sh", MountPath: "/usr/share/container-scripts/postgresql/pginit.sh"})

	if strings.EqualFold(c.hubSpec.BackupSupport, "Yes") || !strings.EqualFold(c.hubSpec.DbPrototype, "empty") {
		postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-backup-vol", MountPath: c.config.NFSPath})
	}
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
