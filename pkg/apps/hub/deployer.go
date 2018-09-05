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

	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	log "github.com/sirupsen/logrus"
)

// createDeployer will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
func (hc *Creater) createDeployer(deployer *horizon.Deployer, createHub *v1.Hub, hubContainerFlavor *ContainerFlavor, allConfigEnv []*kapi.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*kapi.EnvConfig{
		{Type: kapi.EnvFromConfigMap, FromName: "hub-config"},
	}

	dbSecretVolume := components.NewSecretVolume(kapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]kapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: IntToInt32(420)},
		},
		DefaultMode: IntToInt32(420),
	})

	dbEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("cloudsql")

	// cfssl
	// cfsslGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-cfssl", fmt.Sprintf("%s-%s", "cfssl-disk", createHub.Spec.Namespace), "ext4")
	cfsslEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	cfsslContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "cfssl", Image: fmt.Sprintf("%s/%s/hub-cfssl:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.CfsslMemoryLimit, MaxMem: hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: cfsslPort, Protocol: kapi.ProtocolTCP},
	}
	cfssl := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "cfssl", Replicas: IntToInt32(1)},
		[]*api.Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
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
	webServerContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/hub-nginx:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.WebserverMemoryLimit, MaxMem: hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: "", UID: IntToInt64(1000)},
		EnvConfigs: hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: kapi.MountPropagationNone},
			{Name: "certificate", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationNone},
		},
		PortConfig: &kapi.PortConfig{ContainerPort: webserverPort, Protocol: kapi.ProtocolTCP},
	}
	webserver := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webserver", Replicas: IntToInt32(1)},
		[]*api.Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir, webServerSecretVol}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("webserver : %v\n", webserver.GetObj())
	deployer.AddDeployment(webserver)

	// documentation
	documentationContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "documentation", Image: fmt.Sprintf("%s/%s/hub-documentation:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.DocumentationMemoryLimit, MaxMem: hubContainerFlavor.DocumentationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: hubConfigEnv,
		PortConfig: &kapi.PortConfig{ContainerPort: documentationPort, Protocol: kapi.ProtocolTCP},
	}
	documentation := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "documentation", Replicas: IntToInt32(1)},
		[]*api.Container{documentationContainerConfig}, []*components.Volume{}, []*api.Container{}, []kapi.AffinityConfig{})
	// log.Infof("documentation : %v\n", documentation.GetObj())
	deployer.AddDeployment(documentation)

	// solr
	// solrGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-solr", fmt.Sprintf("%s-%s", "solr-disk", createHub.Spec.Namespace), "ext4")
	solrEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	solrContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "solr", Image: fmt.Sprintf("%s/%s/hub-solr:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.SolrMemoryLimit, MaxMem: hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: solrPort, Protocol: kapi.ProtocolTCP},
	}
	solr := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "solr", Replicas: IntToInt32(1)},
		[]*api.Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("solr : %v\n", solr.GetObj())
	deployer.AddDeployment(solr)

	// registration
	// registrationGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-registration", fmt.Sprintf("%s-%s", "registration-disk", createHub.Spec.Namespace), "ext4")
	registrationEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	registrationContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/hub-registration:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.RegistrationMemoryLimit, MaxMem: hubContainerFlavor.RegistrationMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: registrationPort, Protocol: kapi.ProtocolTCP},
	}
	registration := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "registration", Replicas: IntToInt32(1)},
		[]*api.Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("registration : %v\n", registration.GetObj())
	deployer.AddDeployment(registration)

	// zookeeper
	zookeeperEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper")
	zookeeperContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/hub-zookeeper:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: zookeeperPort, Protocol: kapi.ProtocolTCP},
	}
	zookeeper := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "zookeeper", Replicas: IntToInt32(1)},
		[]*api.Container{zookeeperContainerConfig}, []*components.Volume{zookeeperEmptyDir}, []*api.Container{}, []kapi.AffinityConfig{})
	// log.Infof("zookeeper : %v\n", zookeeper.GetObj())
	deployer.AddDeployment(zookeeper)

	// jobRunner
	jobRunnerEnvs := allConfigEnv
	jobRunnerEnvs = append(jobRunnerEnvs, &kapi.EnvConfig{Type: kapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "jobrunner-mem", FromName: "hub-config-resources"})
	jobRunnerContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "jobrunner", Image: fmt.Sprintf("%s/%s/hub-jobrunner:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: "1", MaxCPU: "1"},
		EnvConfigs:   jobRunnerEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: jobRunnerPort, Protocol: kapi.ProtocolTCP},
	}

	jobRunner := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "jobrunner", Replicas: hubContainerFlavor.JobRunnerReplicas},
		[]*api.Container{jobRunnerContainerConfig}, []*components.Volume{dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("jobRunner : %v\n", jobRunner.GetObj())
	deployer.AddDeployment(jobRunner)

	// hub-scan
	scannerEnvs := allConfigEnv
	scannerEnvs = append(scannerEnvs, &kapi.EnvConfig{Type: kapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "scan-mem", FromName: "hub-config-resources"})
	hubScanEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-scan", Image: fmt.Sprintf("%s/%s/hub-scan:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.ScanMemoryLimit, MaxMem: hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationNone},
			{Name: "dir-scan", MountPath: "/opt/blackduck/hub/hub-scan/security", Propagation: kapi.MountPropagationNone}},
		PortConfig: &kapi.PortConfig{ContainerPort: scannerPort, Protocol: kapi.ProtocolTCP},
	}
	hubScan := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "hub-scan", Replicas: hubContainerFlavor.ScanReplicas},
		[]*api.Container{hubScanContainerConfig}, []*components.Volume{hubScanEmptyDir, dbSecretVolume, dbEmptyDir}, []*api.Container{}, []kapi.AffinityConfig{})
	// log.Infof("hubScan : %v\n", hubScan.GetObj())
	deployer.AddDeployment(hubScan)

	// hub-authentication
	authEnvs := allConfigEnv
	authEnvs = append(authEnvs, &kapi.EnvConfig{Type: kapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: hubContainerFlavor.AuthenticationHubMaxMemory})
	// hubAuthGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-authentication", fmt.Sprintf("%s-%s", "authentication-disk", createHub.Spec.Namespace), "ext4")
	hubAuthEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	hubAuthContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/hub-authentication:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationNone},
			{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: kapi.MountPropagationNone}},
		PortConfig: &kapi.PortConfig{ContainerPort: authenticationPort, Protocol: kapi.ProtocolTCP},
	}
	hubAuth := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "hub-authentication", Replicas: IntToInt32(1)},
		[]*api.Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("hubAuth : %v\n", hubAuthc.GetObj())
	deployer.AddDeployment(hubAuth)

	// webapp-logstash
	webappEnvs := allConfigEnv
	webappEnvs = append(webappEnvs, &kapi.EnvConfig{Type: kapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", createHub.Spec.Namespace), "ext4")
	webappEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	webappContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/hub-webapp:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.WebappMemoryLimit, MaxMem: hubContainerFlavor.WebappMemoryLimit, MinCPU: hubContainerFlavor.WebappCPULimit,
			MaxCPU: hubContainerFlavor.WebappCPULimit},
		EnvConfigs: webappEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationNone},
			{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: kapi.MountPropagationNone},
			{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs", Propagation: kapi.MountPropagationNone}},
		PortConfig: &kapi.PortConfig{ContainerPort: webappPort, Protocol: kapi.ProtocolTCP},
	}
	logstashEmptyDir, _ := CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	logstashContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/hub-logstash:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.LogstashMemoryLimit, MaxMem: hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: kapi.MountPropagationNone}},
		PortConfig:   &kapi.PortConfig{ContainerPort: logstashPort, Protocol: kapi.ProtocolTCP},
	}
	webappLogstash := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webapp-logstash", Replicas: IntToInt32(1)}, []*api.Container{webappContainerConfig, logstashContainerConfig},
		[]*components.Volume{webappEmptyDir, logstashEmptyDir, dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("webappLogstash : %v\n", webappLogstashc.GetObj())
	deployer.AddDeployment(webappLogstash)

	deployer.AddService(CreateService("cfssl", "cfssl", createHub.Spec.Namespace, cfsslPort, cfsslPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("zookeeper", "zookeeper", createHub.Spec.Namespace, zookeeperPort, zookeeperPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("webserver", "webserver", createHub.Spec.Namespace, "443", webserverPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("webserver-np", "webserver", createHub.Spec.Namespace, "443", webserverPort, kapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(CreateService("webserver-lb", "webserver", createHub.Spec.Namespace, "443", webserverPort, kapi.ClusterIPServiceTypeLoadBalancer))
	deployer.AddService(CreateService("webapp", "webapp-logstash", createHub.Spec.Namespace, webappPort, webappPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("logstash", "webapp-logstash", createHub.Spec.Namespace, logstashPort, logstashPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("solr", "solr", createHub.Spec.Namespace, solrPort, solrPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("documentation", "documentation", createHub.Spec.Namespace, documentationPort, documentationPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("scan", "hub-scan", createHub.Spec.Namespace, scannerPort, scannerPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("authentication", "hub-authentication", createHub.Spec.Namespace, authenticationPort, authenticationPort, kapi.ClusterIPServiceTypeDefault))
	deployer.AddService(CreateService("registration", "registration", createHub.Spec.Namespace, registrationPort, registrationPort, kapi.ClusterIPServiceTypeDefault))
}
