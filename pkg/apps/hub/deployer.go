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
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	//	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	log "github.com/sirupsen/logrus"
)

// createDeployer will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
func (hc *Creater) createDeployer(deployer *horizon.Deployer, createHub *v1.Hub, hubContainerFlavor *ContainerFlavor, allConfigEnv []*horizonapi.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
	}

	dbSecretVolume := components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]horizonapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: IntToInt32(420)},
		},
		DefaultMode: IntToInt32(420),
	})

	dbEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("cloudsql")

	// cfssl
	// cfsslGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-cfssl", fmt.Sprintf("%s-%s", "cfssl-disk", createHub.Spec.Namespace), "ext4")
	cfsslEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	cfsslContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: fmt.Sprintf("%s/%s/hub-cfssl:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.CfsslMemoryLimit, MaxMem: hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: cfsslPort, Protocol: horizonapi.ProtocolTCP},
	}
	cfssl := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "cfssl", Replicas: IntToInt32(1)},
		[]*Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("cfssl : %v\n", cfssl.GetObj())
	deployer.AddDeployment(cfssl)

	// webserver
	// webServerGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webserver", fmt.Sprintf("%s-%s", "webserver-disk", createHub.Spec.Namespace), "ext4")
	for {
		secret, err := GetSecret(hc.KubeClient, createHub.Name, "hub-certificate")
		if err != nil {
			log.Errorf("unable to get the secret in %s due to %+v", createHub.Name, err)
			break
		}
		data := secret.Data
		if len(data) > 0 {
			break
		}
		time.Sleep(10 * time.Second)
	}
	webServerEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-webserver")
	webServerSecretVol, _ := CreateSecretVolume("certificate", "hub-certificate", 0777)
	webServerContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/hub-nginx:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.WebserverMemoryLimit, MaxMem: hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: "", UID: IntToInt64(1000)},
		EnvConfigs: hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: horizonapi.MountPropagationNone},
			{Name: "certificate", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
		},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webserverPort, Protocol: horizonapi.ProtocolTCP},
	}
	webserver := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webserver", Replicas: IntToInt32(1)},
		[]*Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir, webServerSecretVol}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("webserver : %v\n", webserver.GetObj())
	deployer.AddDeployment(webserver)

	// documentation
	documentationContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: fmt.Sprintf("%s/%s/hub-documentation:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.DocumentationMemoryLimit, MaxMem: hubContainerFlavor.DocumentationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: hubConfigEnv,
		PortConfig: &horizonapi.PortConfig{ContainerPort: documentationPort, Protocol: horizonapi.ProtocolTCP},
	}
	documentation := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "documentation", Replicas: IntToInt32(1)},
		[]*Container{documentationContainerConfig}, []*components.Volume{}, []*Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("documentation : %v\n", documentation.GetObj())
	deployer.AddDeployment(documentation)

	// solr
	// solrGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-solr", fmt.Sprintf("%s-%s", "solr-disk", createHub.Spec.Namespace), "ext4")
	solrEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	solrContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: fmt.Sprintf("%s/%s/hub-solr:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.SolrMemoryLimit, MaxMem: hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: solrPort, Protocol: horizonapi.ProtocolTCP},
	}
	solr := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "solr", Replicas: IntToInt32(1)},
		[]*Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("solr : %v\n", solr.GetObj())
	deployer.AddDeployment(solr)

	// registration
	// registrationGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-registration", fmt.Sprintf("%s-%s", "registration-disk", createHub.Spec.Namespace), "ext4")
	registrationEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	registrationContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/hub-registration:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.RegistrationMemoryLimit, MaxMem: hubContainerFlavor.RegistrationMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: registrationPort, Protocol: horizonapi.ProtocolTCP},
	}
	registration := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "registration", Replicas: IntToInt32(1)},
		[]*Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("registration : %v\n", registration.GetObj())
	deployer.AddDeployment(registration)

	// zookeeper
	zookeeperEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper")
	zookeeperContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/hub-zookeeper:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: zookeeperPort, Protocol: horizonapi.ProtocolTCP},
	}
	zookeeper := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "zookeeper", Replicas: IntToInt32(1)},
		[]*Container{zookeeperContainerConfig}, []*components.Volume{zookeeperEmptyDir}, []*Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("zookeeper : %v\n", zookeeper.GetObj())
	deployer.AddDeployment(zookeeper)

	// jobRunner
	jobRunnerEnvs := allConfigEnv
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "jobrunner-mem", FromName: "hub-config-resources"})
	jobRunnerContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "jobrunner", Image: fmt.Sprintf("%s/%s/hub-jobrunner:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: "1", MaxCPU: "1"},
		EnvConfigs:   jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: jobRunnerPort, Protocol: horizonapi.ProtocolTCP},
	}

	jobRunner := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "jobrunner", Replicas: hubContainerFlavor.JobRunnerReplicas},
		[]*Container{jobRunnerContainerConfig}, []*components.Volume{dbSecretVolume, dbEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("jobRunner : %v\n", jobRunner.GetObj())
	deployer.AddDeployment(jobRunner)

	// hub-scan
	scannerEnvs := allConfigEnv
	scannerEnvs = append(scannerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "scan-mem", FromName: "hub-config-resources"})
	hubScanEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-scan", Image: fmt.Sprintf("%s/%s/hub-scan:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.ScanMemoryLimit, MaxMem: hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-scan", MountPath: "/opt/blackduck/hub/hub-scan/security", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: scannerPort, Protocol: horizonapi.ProtocolTCP},
	}
	hubScan := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "hub-scan", Replicas: hubContainerFlavor.ScanReplicas},
		[]*Container{hubScanContainerConfig}, []*components.Volume{hubScanEmptyDir, dbSecretVolume, dbEmptyDir}, []*Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("hubScan : %v\n", hubScan.GetObj())
	deployer.AddDeployment(hubScan)

	// hub-authentication
	authEnvs := allConfigEnv
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: hubContainerFlavor.AuthenticationHubMaxMemory})
	// hubAuthGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-authentication", fmt.Sprintf("%s-%s", "authentication-disk", createHub.Spec.Namespace), "ext4")
	hubAuthEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	hubAuthContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/hub-authentication:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: authenticationPort, Protocol: horizonapi.ProtocolTCP},
	}
	hubAuth := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "hub-authentication", Replicas: IntToInt32(1)},
		[]*Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, dbSecretVolume, dbEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("hubAuth : %v\n", hubAuthc.GetObj())
	deployer.AddDeployment(hubAuth)

	// webapp-logstash
	webappEnvs := allConfigEnv
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", createHub.Spec.Namespace), "ext4")
	webappEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	webappContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/hub-webapp:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.WebappMemoryLimit, MaxMem: hubContainerFlavor.WebappMemoryLimit, MinCPU: hubContainerFlavor.WebappCPULimit,
			MaxCPU: hubContainerFlavor.WebappCPULimit},
		EnvConfigs: webappEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webappPort, Protocol: horizonapi.ProtocolTCP},
	}
	logstashEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	logstashContainerConfig := &Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/hub-logstash:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.LogstashMemoryLimit, MaxMem: hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: logstashPort, Protocol: horizonapi.ProtocolTCP},
	}
	webappLogstash := CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webapp-logstash", Replicas: IntToInt32(1)}, []*Container{webappContainerConfig, logstashContainerConfig},
		[]*components.Volume{webappEmptyDir, logstashEmptyDir, dbSecretVolume, dbEmptyDir}, []*Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("webappLogstash : %v\n", webappLogstashc.GetObj())
	deployer.AddDeployment(webappLogstash)

	deployer.AddService(CreateService("cfssl", "cfssl", createHub.Spec.Namespace, cfsslPort, cfsslPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("zookeeper", "zookeeper", createHub.Spec.Namespace, zookeeperPort, zookeeperPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("webserver", "webserver", createHub.Spec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("webserver-np", "webserver", createHub.Spec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(CreateService("webserver-lb", "webserver", createHub.Spec.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeLoadBalancer))
	deployer.AddService(CreateService("webapp", "webapp-logstash", createHub.Spec.Namespace, webappPort, webappPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("logstash", "webapp-logstash", createHub.Spec.Namespace, logstashPort, logstashPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("solr", "solr", createHub.Spec.Namespace, solrPort, solrPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("documentation", "documentation", createHub.Spec.Namespace, documentationPort, documentationPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("scan", "hub-scan", createHub.Spec.Namespace, scannerPort, scannerPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("authentication", "hub-authentication", createHub.Spec.Namespace, authenticationPort, authenticationPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("registration", "registration", createHub.Spec.Namespace, registrationPort, registrationPort, horizonapi.ClusterIPServiceTypeDefault))
}
