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

package hub

import (
	"fmt"
	"strings"

	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	log "github.com/sirupsen/logrus"
)

func (hc *Creater) init(deployer *horizon.Deployer, createHub *v1.Hub, hubContainerFlavor *ContainerFlavor, allConfigEnv []*kapi.EnvConfig) error {

	// Create a namespaces
	_, err := GetNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", createHub.Spec.Namespace)
		deployer.AddNamespace(components.NewNamespace(kapi.NamespaceConfig{Name: createHub.Spec.Namespace}))
	}

	// Create a secret
	secrets := hc.createHubSecrets(createHub)

	for _, secret := range secrets {
		deployer.AddSecret(secret)
	}

	// Create ConfigMaps
	configMaps := hc.createHubConfig(createHub, hubContainerFlavor)

	for _, configMap := range configMaps {
		deployer.AddConfigMap(configMap)
	}

	var storageClass string
	if strings.EqualFold(createHub.Spec.PVCStorageClass, "none") {
		storageClass = ""
	} else {
		storageClass = createHub.Spec.PVCStorageClass
	}

	if strings.EqualFold(createHub.Spec.PVCStorageClass, "none") {
		// Postgres PV
		if strings.EqualFold(createHub.Spec.NFSServer, "") {
			return fmt.Errorf("unable to create the PV %s due to missing NFS server path", createHub.Name)
		}

		_, err = CreatePersistentVolume(hc.KubeClient, createHub.Name, storageClass, createHub.Spec.PVCClaimSize, "/data/bds/backup", createHub.Spec.NFSServer)

		if err != nil {
			return fmt.Errorf("unable to create the PV %s due to %+v", createHub.Name, err)
		}
	}

	postgresEnvs := allConfigEnv
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: "HUB_POSTGRES_USER_PASSWORD_FILE", FromName: "db-creds"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", FromName: "db-creds"})

	postgresVolumes := []*components.Volume{}
	postgresEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("postgres-persistent-vol")
	postgresInitConfigVol, _ := CreateConfigMapVolume("postgres-init-vol", "postgres-init", 0777)
	postgresBootstrapConfigVol, _ := CreateConfigMapVolume("postgres-bootstrap-vol", "postgres-bootstrap", 0777)
	postgresVolumes = append(postgresVolumes, postgresEmptyDir, postgresInitConfigVol, postgresBootstrapConfigVol)

	postgresVolumeMounts := []*kapi.VolumeMountConfig{}
	postgresVolumeMounts = append(postgresVolumeMounts, &kapi.VolumeMountConfig{Name: "postgres-persistent-vol", MountPath: "/var/lib/pgsql/data", Propagation: kapi.MountPropagationNone})
	postgresVolumeMounts = append(postgresVolumeMounts, &kapi.VolumeMountConfig{Name: "postgres-bootstrap-vol:pgbootstrap.sh", MountPath: "/usr/share/container-scripts/postgresql/pgbootstrap.sh", Propagation: kapi.MountPropagationNone})
	postgresVolumeMounts = append(postgresVolumeMounts, &kapi.VolumeMountConfig{Name: "postgres-init-vol:pginit.sh", MountPath: "/usr/share/container-scripts/postgresql/pginit.sh", Propagation: kapi.MountPropagationNone})

	if strings.EqualFold(createHub.Spec.BackupSupport, "Yes") || !strings.EqualFold(createHub.Spec.PVCStorageClass, "") {
		// Postgres PVC
		postgresPVC, err := CreatePersistentVolumeClaim(createHub.Name, createHub.Name, createHub.Spec.PVCClaimSize, storageClass, kapi.ReadWriteOnce)
		if err != nil {
			return fmt.Errorf("failed to create the postgres PVC for %s due to %+v", createHub.Name, err)
		}
		deployer.AddPVC(postgresPVC)

		postgresBackupDir, _ := CreatePersistentVolumeClaimVolume("postgres-backup-vol", createHub.Name)
		postgresVolumes = append(postgresVolumes, postgresBackupDir)
		postgresVolumeMounts = append(postgresVolumeMounts, &kapi.VolumeMountConfig{Name: "postgres-backup-vol", MountPath: "/data/bds/backup", Propagation: kapi.MountPropagationNone})
	}

	postgresExternalContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "postgres", Image: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1", PullPolicy: kapi.PullAlways,
			MinMem: hubContainerFlavor.PostgresMemoryLimit, MaxMem: "", MinCPU: hubContainerFlavor.PostgresCPULimit, MaxCPU: "",
			Command: []string{"/usr/share/container-scripts/postgresql/pginit.sh"}},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   &kapi.PortConfig{ContainerPort: postgresPort, Protocol: kapi.ProtocolTCP},
	}
	initContainers := []*api.Container{}
	// If the PV storage is other than NFS or if the backup is enabled and PV storage is other than NFS, add the init container
	if !strings.EqualFold(createHub.Spec.PVCStorageClass, "") && !strings.EqualFold(createHub.Spec.PVCStorageClass, "none") {
		postgresInitContainerConfig := &api.Container{
			ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /data/bds/backup"}},
			VolumeMounts: []*kapi.VolumeMountConfig{
				{Name: "postgres-backup-vol", MountPath: "/data/bds/backup", Propagation: kapi.MountPropagationNone},
			},
			PortConfig: &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}

	postgres := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "postgres", Replicas: IntToInt32(1)},
		[]*api.Container{postgresExternalContainerConfig}, postgresVolumes, initContainers, []kapi.AffinityConfig{})
	// log.Infof("postgres : %+v\n", postgres.GetObj())
	deployer.AddDeployment(postgres)
	deployer.AddService(CreateService("postgres", "postgres", createHub.Spec.Namespace, postgresPort, postgresPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("postgres-exposed", "postgres", createHub.Spec.Namespace, postgresPort, postgresPort, kapi.ClusterIPServiceTypeLoadBalancer))
	return nil
}
