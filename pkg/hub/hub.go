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
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type HubCreater struct {
	Config *rest.Config
	Client *kubernetes.Clientset
}

func NewHubCreater() *HubCreater {
	config, err := GetKubeConfig()

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("unable to get the kubernetes client due to %v", err)
	}
	return &HubCreater{Config: config, Client: client}
}

func (hc *HubCreater) DeleteHub(deleteHub *model.DeleteHubRequest) {
	var err error
	// Verify whether the namespace exist
	_, err = GetNamespace(hc.Client, deleteHub.Namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", deleteHub.Namespace, err)
	} else {
		// Delete a namespace
		err = DeleteNamespace(hc.Client, deleteHub.Namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", deleteHub.Namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := GetNamespace(hc.Client, deleteHub.Namespace)
			log.Infof("Namespace: %v, status: %v", deleteHub.Namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", deleteHub.Namespace)
				break
			}
		}
	}
}

func (hc *HubCreater) CreateHub(createHub *model.CreateHub) {
	// Create a horizon deployer for each hub
	deployer, err := horizon.NewDeployer(hc.Config)
	if err != nil {
		log.Errorf("unable to create the horizon deployer due to %+v", err)
	}

	// Get Containers Flavor
	hubContainerFlavor := GetHubContainersFlavor(createHub.Flavor)
	log.Debugf("Hub Container Flavor: %+v", hubContainerFlavor)

	// All ConfigMap environment variables
	allConfigEnv := []*api.EnvConfig{
		&api.EnvConfig{Type: api.EnvFromConfigMap, FromName: "hub-config"},
		&api.EnvConfig{Type: api.EnvFromConfigMap, FromName: "hub-db-config"},
		&api.EnvConfig{Type: api.EnvFromConfigMap, FromName: "hub-db-config-granular"},
	}

	// Create the config-maps, secrets and postgres container
	hc.init(deployer, createHub, hubContainerFlavor, allConfigEnv)
	// Deploy config-maps, secrets and postgres container
	err = deployer.Run()
	time.Sleep(20 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err := GetAllPodsForNamespace(hc.Client, createHub.Namespace)
	if err != nil {
		log.Errorf("unable to list the pods in namespace %s due to %+v")
		return
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.Client, pods)
	// Initialize the hub database
	InitDatabase(createHub.Namespace)

	// Create all hub deployments
	hc.createHubDeployer(deployer, createHub, hubContainerFlavor, allConfigEnv)
	log.Debugf("%+v", deployer)
	// Deploy all hub containers
	err = deployer.Run()
	if err != nil {
		log.Errorf("Deployments failed because %+v", err)
	}
	time.Sleep(10 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err = GetAllPodsForNamespace(hc.Client, createHub.Namespace)
	if err != nil {
		log.Errorf("unable to list the pods in namespace %s due to %+v")
		return
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.Client, pods)

	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod := FilterPodByNamePrefix(pods)
	log.Infof("Registration pod: %+v", registrationPod)
	registrationKey := os.Getenv("REGISTRATION_KEY")
	log.Infof("Registration key: %s", registrationKey)

	if registrationPod != nil {
		for {
			// Create the exec into kubernetes pod request
			req := CreateExecContainerRequest(hc.Client, registrationPod)
			// Exec into the kubernetes pod and execute the commands
			err = hc.execContainer(req, []string{fmt.Sprintf("curl -k -X POST https://127.0.0.1:8443/registration/HubRegistration?action=activate\\&registrationid=%s", registrationKey)})
			if err != nil {
				log.Infof("error in Stream: %v", err)
			} else {
				// Hub created and auto registered. Exit!!!!
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	ipAddress, err := hc.getLoadBalancerIpAddress(createHub.Namespace, "webserver-exp")
	if err != nil {
		log.Error(err)
	}
	log.Infof("Hub Ip address: %s", ipAddress)
}

func (hc *HubCreater) execContainer(request *rest.Request, command []string) error {
	var stdin io.Reader
	stdin = NewStringReader(command)

	log.Debugf("Request URL: %+v, request: %+v", request.URL().String(), request)

	exec, err := remotecommand.NewSPDYExecutor(hc.Config, "POST", request.URL())
	log.Debugf("exec: %+v, error: %+v", exec, err)
	if err != nil {
		log.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	log.Infof("stdout: %s, stderr: %s", stdout.String(), stderr.String())
	return err
}

// CreateHubConfig will create the hub configMaps
func (hc *HubCreater) createHubConfig(namespace string, hubversion string, hubContainerFlavor *HubContainerFlavor) map[string]*components.ConfigMap {
	configMaps := make(map[string]*components.ConfigMap)

	hubConfig := components.NewConfigMap(api.ConfigMapConfig{Namespace: namespace, Name: "hub-config"})
	hubConfig.AddData(map[string]string{
		"PUBLIC_HUB_WEBSERVER_HOST": "localhost",
		"PUBLIC_HUB_WEBSERVER_PORT": "443",
		"HUB_WEBSERVER_PORT":        "8443",
		"IPV4_ONLY":                 "0",
		"RUN_SECRETS_DIR":           "/tmp/secrets",
		"HUB_VERSION":               hubversion,
		"HUB_PROXY_NON_PROXY_HOSTS": "solr",
	})

	configMaps["hub-config"] = hubConfig

	hubDbConfig := components.NewConfigMap(api.ConfigMapConfig{Namespace: namespace, Name: "hub-db-config"})
	hubDbConfig.AddData(map[string]string{
		"HUB_POSTGRES_ADMIN": "blackduck",
		"HUB_POSTGRES_USER":  "blackduck_user",
		"HUB_POSTGRES_PORT":  "5432",
		"HUB_POSTGRES_HOST":  "postgres",
	})

	configMaps["hub-db-config"] = hubDbConfig

	hubConfigResources := components.NewConfigMap(api.ConfigMapConfig{Namespace: namespace, Name: "hub-config-resources"})
	hubConfigResources.AddData(map[string]string{
		"webapp-mem":    hubContainerFlavor.WebappHubMaxMemory,
		"jobrunner-mem": hubContainerFlavor.JobRunnerHubMaxMemory,
		"scan-mem":      hubContainerFlavor.ScanHubMaxMemory,
	})

	configMaps["hub-config-resources"] = hubConfigResources

	hubDbConfigGranular := components.NewConfigMap(api.ConfigMapConfig{Namespace: namespace, Name: "hub-db-config-granular"})
	hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})

	configMaps["hub-db-config-granular"] = hubDbConfigGranular
	return configMaps
}

func (hc *HubCreater) createHubSecrets(namespace string, adminPassword string, userPassword string) []*components.Secret {
	var secrets []*components.Secret
	// file, err := ReadFromFile(GCLOUD_AUTH_FILE_PATH)
	//
	// if err != nil {
	// 	log.Errorf("Unable to read the file %s due to error: +%v", GCLOUD_AUTH_FILE_PATH, err)
	// } else {
	//
	// 	gcloudAuthSecret := components.NewSecret(api.SecretConfig{
	// 		Namespace: namespace,
	// 		Name:      "hub-postgres-gcloud-instance-creds",
	// 		Type:      api.SecretTypeOpaque,
	// 	})
	// 	gcloudAuthSecret.AddStringData(map[string]string{"credentials.json": string(file)})
	// 	secrets = append(secrets, gcloudAuthSecret)
	// }

	hubSecret := components.NewSecret(api.SecretConfig{Namespace: namespace, Name: "db-creds", Type: api.SecretTypeOpaque})
	hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": adminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": userPassword})
	secrets = append(secrets, hubSecret)
	return secrets
}

func (hc *HubCreater) init(deployer *horizon.Deployer, createHub *model.CreateHub, hubContainerFlavor *HubContainerFlavor, allConfigEnv []*api.EnvConfig) {

	// Create a namespaces
	deployer.AddNamespace(components.NewNamespace(api.NamespaceConfig{Name: createHub.Namespace}))

	// Create a secret
	secrets := hc.createHubSecrets(createHub.Namespace, createHub.AdminPassword, createHub.UserPassword)

	for _, secret := range secrets {
		deployer.AddSecret(secret)
	}

	// Create ConfigMaps
	configMaps := hc.createHubConfig(createHub.Namespace, createHub.HubVersion, hubContainerFlavor)

	for _, configMap := range configMaps {
		deployer.AddConfigMap(configMap)
	}

	//Postgres
	postgresEnvs := allConfigEnv
	postgresEnvs = append(postgresEnvs, &api.EnvConfig{Type: api.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &api.EnvConfig{Type: api.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: "HUB_POSTGRES_USER_PASSWORD_FILE", FromName: "db-creds"})
	postgresEnvs = append(postgresEnvs, &api.EnvConfig{Type: api.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &api.EnvConfig{Type: api.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", FromName: "db-creds"})
	postgresEmptyDir, _ := CreateEmptyDirVolume("postgres-persistent-vol", "1G")
	postgresExternalContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "postgres", Image: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1", PullPolicy: api.PullAlways,
			MinMem: hubContainerFlavor.PostgresMemoryLimit, MaxMem: "", MinCPU: hubContainerFlavor.PostgresCpuLimit, MaxCPU: ""},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "postgres-persistent-vol", MountPath: "/var/lib/pgsql/data", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: POSTGRES_PORT, Protocol: api.ProtocolTCP},
	}
	postgres := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "postgres", Replicas: 1}, []*Container{postgresExternalContainerConfig},
		[]*components.Volume{postgresEmptyDir}, []*Container{}, []api.AffinityConfig{})
	// log.Infof("postgres : %+v\n", postgres.GetObj())
	deployer.AddDeployment(postgres)
	deployer.AddService(CreateService("postgres", "postgres", createHub.Namespace, POSTGRES_PORT, POSTGRES_PORT, false))
}

// CreateHub will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
func (hc *HubCreater) createHubDeployer(deployer *horizon.Deployer, createHub *model.CreateHub, hubContainerFlavor *HubContainerFlavor, allConfigEnv []*api.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*api.EnvConfig{
		&api.EnvConfig{Type: api.EnvFromConfigMap, FromName: "hub-config"},
	}

	// Common DB secret volume
	hubPostgresSecretVolume := components.NewSecretVolume(api.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "hub-postgres-gcloud-instance-creds",
		MapOrSecretName: "hub-postgres-gcloud-instance-creds",
		DefaultMode:     420,
	})

	dbSecretVolume := components.NewSecretVolume(api.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]api.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": api.KeyAndMode{KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: 420},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  api.KeyAndMode{KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: 420},
		},
		DefaultMode: 420,
	})

	dbEmptyDir, _ := CreateEmptyDirVolume("cloudsql", "1G")

	// cfssl
	// cfsslGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-cfssl", fmt.Sprintf("%s-%s", "cfssl-disk", createHub.Namespace), "ext4")
	cfsslEmptyDir, _ := CreateEmptyDirVolume("dir-cfssl", "1G")
	cfsslContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "cfssl", Image: fmt.Sprintf("%s/%s/hub-cfssl:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.CfsslMemoryLimit, MaxMem: hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: CFSSL_PORT, Protocol: api.ProtocolTCP},
	}
	// cfsslInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /etc/cfssl"}},
	// 	VolumeMounts:    []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig:      &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	cfssl := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "cfssl", Replicas: 1},
		[]*Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("cfssl : %v\n", cfssl.GetObj())
	deployer.AddDeployment(cfssl)

	// webserver
	// webServerGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webserver", fmt.Sprintf("%s-%s", "webserver-disk", createHub.Namespace), "ext4")
	webServerEmptyDir, _ := CreateEmptyDirVolume("dir-webserver", "1G")
	webServerContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/hub-nginx:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.WebserverMemoryLimit, MaxMem: hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: WEBSERVER_PORT, Protocol: api.ProtocolTCP},
	}
	// webserverInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/"}},
	// 	VolumeMounts:    []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig:      &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	webserver := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "webserver", Replicas: 1},
		[]*Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("webserver : %v\n", webserver.GetObj())
	deployer.AddDeployment(webserver)

	// documentation
	documentationContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "documentation", Image: fmt.Sprintf("%s/%s/hub-documentation:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.DocumentationMemoryLimit, MaxMem: hubContainerFlavor.DocumentationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: hubConfigEnv,
		PortConfig: &api.PortConfig{ContainerPort: DOCUMENTATION_PORT, Protocol: api.ProtocolTCP},
	}
	documentation := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "documentation", Replicas: 1},
		[]*Container{documentationContainerConfig}, []*components.Volume{}, []*Container{}, []api.AffinityConfig{})
	// log.Infof("documentation : %v\n", documentation.GetObj())
	deployer.AddDeployment(documentation)

	// solr
	// solrGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-solr", fmt.Sprintf("%s-%s", "solr-disk", createHub.Namespace), "ext4")
	solrEmptyDir, _ := CreateEmptyDirVolume("dir-solr", "1G")
	solrContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "solr", Image: fmt.Sprintf("%s/%s/hub-solr:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.SolrMemoryLimit, MaxMem: hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: SOLR_PORT, Protocol: api.ProtocolTCP},
	}
	// solrInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/solr/cores.data"}},
	// 	VolumeMounts:    []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig:      &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	solr := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "solr", Replicas: 1},
		[]*Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("solr : %v\n", solr.GetObj())
	deployer.AddDeployment(solr)

	// registration
	// registrationGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-registration", fmt.Sprintf("%s-%s", "registration-disk", createHub.Namespace), "ext4")
	registrationEmptyDir, _ := CreateEmptyDirVolume("dir-registration", "1G")
	registrationContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/hub-registration:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.RegistrationMemoryLimit, MaxMem: hubContainerFlavor.RegistrationMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: REGISTRATION_PORT, Protocol: api.ProtocolTCP},
	}
	// registrationInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/hub-registration/config"}},
	// 	VolumeMounts:    []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig:      &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	registration := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "registration", Replicas: 1},
		[]*Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("registration : %v\n", registration.GetObj())
	deployer.AddDeployment(registration)

	// zookeeper
	zookeeperEmptyDir, _ := CreateEmptyDirVolume("dir-zookeeper", "1G")
	zookeeperContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/hub-zookeeper:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: ZOOKEEPER_PORT, Protocol: api.ProtocolTCP},
	}
	zookeeper := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "zookeeper", Replicas: 1},
		[]*Container{zookeeperContainerConfig}, []*components.Volume{zookeeperEmptyDir}, []*Container{}, []api.AffinityConfig{})
	// log.Infof("zookeeper : %v\n", zookeeper.GetObj())
	deployer.AddDeployment(zookeeper)

	// jobRunner
	jobRunnerEnvs := allConfigEnv
	jobRunnerEnvs = append(jobRunnerEnvs, &api.EnvConfig{Type: api.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "jobrunner-mem", FromName: "hub-config-resources"})
	jobRunnerContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "jobrunner", Image: fmt.Sprintf("%s/%s/hub-jobrunner:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: "1", MaxCPU: "1"},
		EnvConfigs:   jobRunnerEnvs,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: JOBRUNNER_PORT, Protocol: api.ProtocolTCP},
	}

	// cloudProxyContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "cloudsql-proxy", Image: "gcr.io/cloudsql-docker/gce-proxy:1.11", PullPolicy: api.PullAlways},
	// 	VolumeMounts: []*api.VolumeMountConfig{
	// 		&api.VolumeMountConfig{Name: "hub-postgres-gcloud-instance-creds", MountPath: "/secrets/cloudsql", Propagation: api.MountPropagationBidirectional},
	// 		&api.VolumeMountConfig{Name: "cloudsql", MountPath: "/cloudsql", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig: &api.PortConfig{ContainerPort: POSTGRES_PORT, Protocol: api.ProtocolTCP},
	// }

	jobRunner := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "jobrunner", Replicas: hubContainerFlavor.JobRunnerReplicas},
		[]*Container{jobRunnerContainerConfig}, []*components.Volume{dbSecretVolume, hubPostgresSecretVolume, dbEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("jobRunner : %v\n", jobRunner.GetObj())
	deployer.AddDeployment(jobRunner)

	// hub-scan
	scannerEnvs := allConfigEnv
	scannerEnvs = append(scannerEnvs, &api.EnvConfig{Type: api.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "scan-mem", FromName: "hub-config-resources"})
	hubScanEmptyDir, _ := CreateEmptyDirVolume("dir-scan", "1G")
	hubScanContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "hub-scan", Image: fmt.Sprintf("%s/%s/hub-scan:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.ScanMemoryLimit, MaxMem: hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*api.VolumeMountConfig{
			&api.VolumeMountConfig{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: api.MountPropagationBidirectional},
			&api.VolumeMountConfig{Name: "dir-scan", MountPath: "/opt/blackduck/hub/hub-scan/security", Propagation: api.MountPropagationBidirectional}},
		PortConfig: &api.PortConfig{ContainerPort: SCANNER_PORT, Protocol: api.ProtocolTCP},
	}
	hubScan := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "hub-scan", Replicas: hubContainerFlavor.ScanReplicas},
		[]*Container{hubScanContainerConfig}, []*components.Volume{hubScanEmptyDir, dbSecretVolume, hubPostgresSecretVolume, dbEmptyDir}, []*Container{}, []api.AffinityConfig{})
	// log.Infof("hubScan : %v\n", hubScan.GetObj())
	deployer.AddDeployment(hubScan)

	// hub-authentication
	authEnvs := allConfigEnv
	authEnvs = append(authEnvs, &api.EnvConfig{Type: api.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: hubContainerFlavor.AuthenticationHubMaxMemory})
	// hubAuthGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-authentication", fmt.Sprintf("%s-%s", "authentication-disk", createHub.Namespace), "ext4")
	hubAuthEmptyDir, _ := CreateEmptyDirVolume("dir-authentication", "1G")
	hubAuthContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/hub-authentication:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*api.VolumeMountConfig{
			&api.VolumeMountConfig{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: api.MountPropagationBidirectional},
			&api.VolumeMountConfig{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: api.MountPropagationBidirectional}},
		PortConfig: &api.PortConfig{ContainerPort: AUTHENTICATION_PORT, Protocol: api.ProtocolTCP},
	}
	// hubAuthInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/hub-authentication/security"}},
	// 	VolumeMounts:    []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig:      &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	hubAuth := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "hub-authentication", Replicas: 1},
		[]*Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, dbSecretVolume, hubPostgresSecretVolume, dbEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("hubAuth : %v\n", hubAuthc.GetObj())
	deployer.AddDeployment(hubAuth)

	// webapp-logstash
	webappEnvs := allConfigEnv
	webappEnvs = append(webappEnvs, &api.EnvConfig{Type: api.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", createHub.Namespace), "ext4")
	webappEmptyDir, _ := CreateEmptyDirVolume("dir-webapp", "1G")
	webappContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/hub-webapp:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.WebappMemoryLimit, MaxMem: hubContainerFlavor.WebappMemoryLimit, MinCPU: hubContainerFlavor.WebappCpuLimit,
			MaxCPU: hubContainerFlavor.WebappCpuLimit},
		EnvConfigs: webappEnvs,
		VolumeMounts: []*api.VolumeMountConfig{
			&api.VolumeMountConfig{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: api.MountPropagationBidirectional},
			&api.VolumeMountConfig{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: api.MountPropagationBidirectional},
			&api.VolumeMountConfig{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs", Propagation: api.MountPropagationBidirectional}},
		PortConfig: &api.PortConfig{ContainerPort: WEBAPP_PORT, Protocol: api.ProtocolTCP},
	}
	// webappLogStashInitContainerConfig := &Container{
	// 	ContainerConfig: &api.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: api.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /var/lib/logstash/data && chmod -cR 777 /opt/blackduck/hub/"}},
	// 	VolumeMounts: []*api.VolumeMountConfig{
	// 		&api.VolumeMountConfig{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: api.MountPropagationBidirectional},
	// 		&api.VolumeMountConfig{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: api.MountPropagationBidirectional}},
	// 	PortConfig: &api.PortConfig{ContainerPort: "3001", Protocol: api.ProtocolTCP},
	// }
	// logStashGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-logstash", fmt.Sprintf("%s-%s", "logstash-disk", createHub.Namespace), "ext4")
	logstashEmptyDir, _ := CreateEmptyDirVolume("dir-logstash", "1G")
	logstashContainerConfig := &Container{
		ContainerConfig: &api.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/hub-logstash:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: api.PullAlways, MinMem: hubContainerFlavor.LogstashMemoryLimit, MaxMem: hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*api.VolumeMountConfig{&api.VolumeMountConfig{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: api.MountPropagationBidirectional}},
		PortConfig:   &api.PortConfig{ContainerPort: LOGSTASH_PORT, Protocol: api.ProtocolTCP},
	}
	webappLogstash := CreateDeploymentFromContainer(&api.DeploymentConfig{Namespace: createHub.Namespace, Name: "webapp-logstash", Replicas: 1}, []*Container{webappContainerConfig, logstashContainerConfig},
		[]*components.Volume{webappEmptyDir, logstashEmptyDir, dbSecretVolume, hubPostgresSecretVolume, dbEmptyDir}, []*Container{},
		[]api.AffinityConfig{})
	// log.Infof("webappLogstash : %v\n", webappLogstashc.GetObj())
	deployer.AddDeployment(webappLogstash)

	deployer.AddService(CreateService("cfssl", "cfssl", createHub.Namespace, CFSSL_PORT, CFSSL_PORT, false))
	deployer.AddService(CreateService("zookeeper", "zookeeper", createHub.Namespace, ZOOKEEPER_PORT, ZOOKEEPER_PORT, false))
	deployer.AddService(CreateService("webserver", "webserver", createHub.Namespace, "443", WEBSERVER_PORT, false))
	deployer.AddService(CreateService("webserver-exp", "webserver", createHub.Namespace, "443", WEBSERVER_PORT, true))
	deployer.AddService(CreateService("webapp", "webapp-logstash", createHub.Namespace, WEBAPP_PORT, WEBAPP_PORT, false))
	deployer.AddService(CreateService("logstash", "webapp-logstash", createHub.Namespace, LOGSTASH_PORT, LOGSTASH_PORT, false))
	deployer.AddService(CreateService("solr", "solr", createHub.Namespace, SOLR_PORT, SOLR_PORT, false))
	deployer.AddService(CreateService("documentation", "documentation", createHub.Namespace, DOCUMENTATION_PORT, DOCUMENTATION_PORT, false))
	deployer.AddService(CreateService("scan", "hub-scan", createHub.Namespace, SCANNER_PORT, SCANNER_PORT, false))
	deployer.AddService(CreateService("authentication", "hub-authentication", createHub.Namespace, AUTHENTICATION_PORT, AUTHENTICATION_PORT, false))
	deployer.AddService(CreateService("registration", "registration", createHub.Namespace, REGISTRATION_PORT, REGISTRATION_PORT, false))
}

func (hc *HubCreater) getLoadBalancerIpAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		service, err := GetService(hc.Client, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace due to %s", serviceName, namespace, err.Error())
		}

		log.Debugf("Service: %v", service)

		if len(service.Status.LoadBalancer.Ingress) > 0 {
			ipAddress := service.Status.LoadBalancer.Ingress[0].IP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}
