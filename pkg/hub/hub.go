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

	kapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	Config     *rest.Config
	KubeClient *kubernetes.Clientset
	HubClient  *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset) (*Creater, error) {
	return &Creater{Config: config, KubeClient: kubeClient, HubClient: hubClient}, nil
}

// DeleteHub will delete the Black Duck Hub
func (hc *Creater) DeleteHub(namespace string) {
	var err error
	// Verify whether the namespace exist
	_, err = GetNamespace(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = DeleteNamespace(hc.KubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := GetNamespace(hc.KubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}
}

// CreateHub will create the Black Duck Hub
func (hc *Creater) CreateHub(createHub *v1.Hub) (string, error) {
	log.Debugf("Create Hub details for %s: %+v", createHub.Spec.Namespace, createHub)
	// Create a horizon deployer for each hub
	deployer, err := horizon.NewDeployer(hc.Config)
	if err != nil {
		return "", fmt.Errorf("unable to create the horizon deployer due to %+v", err)
	}

	// Get Containers Flavor
	hubContainerFlavor := GetContainersFlavor(createHub.Spec.Flavor)
	log.Debugf("Hub Container Flavor: %+v", hubContainerFlavor)

	if hubContainerFlavor == nil {
		return "", fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) OpsSight, Actual: %s", createHub.Spec.Flavor)
	}

	// All ConfigMap environment variables
	allConfigEnv := []*kapi.EnvConfig{
		{Type: kapi.EnvFromConfigMap, FromName: "hub-config"},
		{Type: kapi.EnvFromConfigMap, FromName: "hub-db-config"},
		{Type: kapi.EnvFromConfigMap, FromName: "hub-db-config-granular"},
	}

	if createHub.Spec.IsRandomPassword {
		createHub.Spec.AdminPassword, _ = RandomString(12)
		createHub.Spec.UserPassword, _ = RandomString(12)
	} else {
		createHub.Spec.AdminPassword = createHub.Spec.PostgresPassword
		createHub.Spec.UserPassword = createHub.Spec.PostgresPassword
	}
	log.Debugf("Before init: %+v", createHub)
	// Create the config-maps, secrets and postgres container
	hc.init(deployer, createHub, hubContainerFlavor, allConfigEnv)
	// Deploy config-maps, secrets and postgres container
	err = deployer.Run()
	if err != nil {
		log.Errorf("Deployments failed because %+v", err)
	}
	// time.Sleep(20 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err := GetAllPodsForNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		return "", fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Spec.Namespace, err)
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.KubeClient, pods)
	// Initialize the hub database
	InitDatabase(createHub)

	// Create all hub deployments
	deployer, err = horizon.NewDeployer(hc.Config)
	hc.createHubDeployer(deployer, createHub, hubContainerFlavor, allConfigEnv)
	log.Debugf("%+v", deployer)
	// Deploy all hub containers
	err = deployer.Run()
	if err != nil {
		log.Errorf("Deployments failed because %+v", err)
	}
	time.Sleep(10 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err = GetAllPodsForNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		return "", fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Spec.Namespace, err)
	}
	// Validate all pods are in running state
	ValidatePodsAreRunning(hc.KubeClient, pods)

	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod := FilterPodByNamePrefix(pods)
	log.Debugf("registration pod: %+v", registrationPod)
	registrationKey := os.Getenv("REGISTRATION_KEY")
	log.Debugf("registration key: %s", registrationKey)

	if registrationPod != nil {
		for {
			// Create the exec into kubernetes pod request
			req := CreateExecContainerRequest(hc.KubeClient, registrationPod)
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

	ipAddress, err := hc.getLoadBalancerIPAddress(createHub.Spec.Namespace, "webserver-exp")
	if err != nil {
		return "", err
	}
	log.Infof("hub Ip address: %s", ipAddress)
	return ipAddress, nil
}

func (hc *Creater) execContainer(request *rest.Request, command []string) error {
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

	log.Debugf("stdout: %s, stderr: %s", stdout.String(), stderr.String())
	return err
}

// CreateHubConfig will create the hub configMaps
func (hc *Creater) createHubConfig(namespace string, hubversion string, hubContainerFlavor *ContainerFlavor) map[string]*components.ConfigMap {
	configMaps := make(map[string]*components.ConfigMap)

	hubConfig := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: namespace, Name: "hub-config"})
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

	hubDbConfig := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: namespace, Name: "hub-db-config"})
	hubDbConfig.AddData(map[string]string{
		"HUB_POSTGRES_ADMIN": "blackduck",
		"HUB_POSTGRES_USER":  "blackduck_user",
		"HUB_POSTGRES_PORT":  "5432",
		"HUB_POSTGRES_HOST":  "postgres",
	})

	configMaps["hub-db-config"] = hubDbConfig

	hubConfigResources := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: namespace, Name: "hub-config-resources"})
	hubConfigResources.AddData(map[string]string{
		"webapp-mem":    hubContainerFlavor.WebappHubMaxMemory,
		"jobrunner-mem": hubContainerFlavor.JobRunnerHubMaxMemory,
		"scan-mem":      hubContainerFlavor.ScanHubMaxMemory,
	})

	configMaps["hub-config-resources"] = hubConfigResources

	hubDbConfigGranular := components.NewConfigMap(kapi.ConfigMapConfig{Namespace: namespace, Name: "hub-db-config-granular"})
	hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})

	configMaps["hub-db-config-granular"] = hubDbConfigGranular
	return configMaps
}

func (hc *Creater) createHubSecrets(namespace string, adminPassword string, userPassword string) []*components.Secret {
	var secrets []*components.Secret
	// file, err := ReadFromFile(GCLOUD_AUTH_FILE_PATH)
	//
	// if err != nil {
	// 	log.Errorf("Unable to read the file %s due to error: +%v", GCLOUD_AUTH_FILE_PATH, err)
	// } else {
	//
	// 	gcloudAuthSecret := components.NewSecret(kapi.SecretConfig{
	// 		Namespace: namespace,
	// 		Name:      "hub-postgres-gcloud-instance-creds",
	// 		Type:      kapi.SecretTypeOpaque,
	// 	})
	// 	gcloudAuthSecret.AddStringData(map[string]string{"credentials.json": string(file)})
	// 	secrets = append(secrets, gcloudAuthSecret)
	// }

	hubSecret := components.NewSecret(kapi.SecretConfig{Namespace: namespace, Name: "db-creds", Type: kapi.SecretTypeOpaque})
	hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": adminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": userPassword})
	secrets = append(secrets, hubSecret)
	return secrets
}

func (hc *Creater) init(deployer *horizon.Deployer, createHub *v1.Hub, hubContainerFlavor *ContainerFlavor, allConfigEnv []*kapi.EnvConfig) {

	// Create a namespaces
	_, err := GetNamespace(hc.KubeClient, createHub.Spec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", createHub.Spec.Namespace)
		deployer.AddNamespace(components.NewNamespace(kapi.NamespaceConfig{Name: createHub.Spec.Namespace}))
	}

	// Create a secret
	secrets := hc.createHubSecrets(createHub.Spec.Namespace, createHub.Spec.AdminPassword, createHub.Spec.UserPassword)

	for _, secret := range secrets {
		deployer.AddSecret(secret)
	}

	// Create ConfigMaps
	configMaps := hc.createHubConfig(createHub.Spec.Namespace, createHub.Spec.HubVersion, hubContainerFlavor)

	for _, configMap := range configMaps {
		deployer.AddConfigMap(configMap)
	}

	//Postgres
	postgresEnvs := allConfigEnv
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: "HUB_POSTGRES_USER_PASSWORD_FILE", FromName: "db-creds"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &kapi.EnvConfig{Type: kapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", FromName: "db-creds"})
	postgresEmptyDir, _ := CreateEmptyDirVolume("postgres-persistent-vol", "1G")
	postgresExternalContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "postgres", Image: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1", PullPolicy: kapi.PullAlways,
			MinMem: hubContainerFlavor.PostgresMemoryLimit, MaxMem: "", MinCPU: hubContainerFlavor.PostgresCPULimit, MaxCPU: ""},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "postgres-persistent-vol", MountPath: "/var/lib/pgsql/data", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: postgresPort, Protocol: kapi.ProtocolTCP},
	}
	postgres := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "postgres", Replicas: IntToInt32(1)}, []*api.Container{postgresExternalContainerConfig},
		[]*components.Volume{postgresEmptyDir}, []*api.Container{}, []kapi.AffinityConfig{})
	// log.Infof("postgres : %+v\n", postgres.GetObj())
	deployer.AddDeployment(postgres)
	deployer.AddService(CreateService("postgres", "postgres", createHub.Spec.Namespace, postgresPort, postgresPort, false))
}

// CreateHub will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
func (hc *Creater) createHubDeployer(deployer *horizon.Deployer, createHub *v1.Hub, hubContainerFlavor *ContainerFlavor, allConfigEnv []*kapi.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*kapi.EnvConfig{
		{Type: kapi.EnvFromConfigMap, FromName: "hub-config"},
	}

	// Common DB secret volume
	// hubPostgresSecretVolume := components.NewSecretVolume(kapi.ConfigMapOrSecretVolumeConfig{
	// 	VolumeName:      "hub-postgres-gcloud-instance-creds",
	// 	MapOrSecretName: "hub-postgres-gcloud-instance-creds",
	// 	DefaultMode:     IntToInt32(420),
	// })

	dbSecretVolume := components.NewSecretVolume(kapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]kapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: IntToInt32(420)},
		},
		DefaultMode: IntToInt32(420),
	})

	dbEmptyDir, _ := CreateEmptyDirVolume("cloudsql", "1G")

	// cfssl
	// cfsslGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-cfssl", fmt.Sprintf("%s-%s", "cfssl-disk", createHub.Spec.Namespace), "ext4")
	cfsslEmptyDir, _ := CreateEmptyDirVolume("dir-cfssl", "1G")
	cfsslContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "cfssl", Image: fmt.Sprintf("%s/%s/hub-cfssl:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.CfsslMemoryLimit, MaxMem: hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: cfsslPort, Protocol: kapi.ProtocolTCP},
	}
	// cfsslInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /etc/cfssl"}},
	// 	VolumeMounts:    []*kapi.VolumeMountConfig{&kapi.VolumeMountConfig{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig:      &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	cfssl := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "cfssl", Replicas: IntToInt32(1)},
		[]*api.Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("cfssl : %v\n", cfssl.GetObj())
	deployer.AddDeployment(cfssl)

	// webserver
	// webServerGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webserver", fmt.Sprintf("%s-%s", "webserver-disk", createHub.Spec.Namespace), "ext4")
	webServerEmptyDir, _ := CreateEmptyDirVolume("dir-webserver", "1G")
	webServerContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/hub-nginx:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.WebserverMemoryLimit, MaxMem: hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: webserverPort, Protocol: kapi.ProtocolTCP},
	}
	// webserverInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /opt/blackduck/hub/"}},
	// 	VolumeMounts:    []*kapi.VolumeMountConfig{&kapi.VolumeMountConfig{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig:      &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	webserver := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webserver", Replicas: IntToInt32(1)},
		[]*api.Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir}, []*api.Container{},
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
	solrEmptyDir, _ := CreateEmptyDirVolume("dir-solr", "1G")
	solrContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "solr", Image: fmt.Sprintf("%s/%s/hub-solr:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.SolrMemoryLimit, MaxMem: hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: solrPort, Protocol: kapi.ProtocolTCP},
	}
	// solrInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/solr/cores.data"}},
	// 	VolumeMounts:    []*kapi.VolumeMountConfig{&kapi.VolumeMountConfig{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig:      &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	solr := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "solr", Replicas: IntToInt32(1)},
		[]*api.Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("solr : %v\n", solr.GetObj())
	deployer.AddDeployment(solr)

	// registration
	// registrationGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-registration", fmt.Sprintf("%s-%s", "registration-disk", createHub.Spec.Namespace), "ext4")
	registrationEmptyDir, _ := CreateEmptyDirVolume("dir-registration", "1G")
	registrationContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/hub-registration:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.RegistrationMemoryLimit, MaxMem: hubContainerFlavor.RegistrationMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: registrationPort, Protocol: kapi.ProtocolTCP},
	}
	// registrationInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/hub-registration/config"}},
	// 	VolumeMounts:    []*kapi.VolumeMountConfig{&kapi.VolumeMountConfig{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig:      &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	registration := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "registration", Replicas: IntToInt32(1)},
		[]*api.Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("registration : %v\n", registration.GetObj())
	deployer.AddDeployment(registration)

	// zookeeper
	zookeeperEmptyDir, _ := CreateEmptyDirVolume("dir-zookeeper", "1G")
	zookeeperContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/hub-zookeeper:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs", Propagation: kapi.MountPropagationBidirectional}},
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
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: jobRunnerPort, Protocol: kapi.ProtocolTCP},
	}

	// cloudProxyContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "cloudsql-proxy", Image: "gcr.io/cloudsql-docker/gce-proxy:1.11", PullPolicy: kapi.PullAlways},
	// 	VolumeMounts: []*kapi.VolumeMountConfig{
	// 		&kapi.VolumeMountConfig{Name: "hub-postgres-gcloud-instance-creds", MountPath: "/secrets/cloudsql", Propagation: kapi.MountPropagationBidirectional},
	// 		&kapi.VolumeMountConfig{Name: "cloudsql", MountPath: "/cloudsql", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig: &kapi.PortConfig{ContainerPort: POSTGRES_PORT, Protocol: kapi.ProtocolTCP},
	// }

	jobRunner := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "jobrunner", Replicas: hubContainerFlavor.JobRunnerReplicas},
		[]*api.Container{jobRunnerContainerConfig}, []*components.Volume{dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("jobRunner : %v\n", jobRunner.GetObj())
	deployer.AddDeployment(jobRunner)

	// hub-scan
	scannerEnvs := allConfigEnv
	scannerEnvs = append(scannerEnvs, &kapi.EnvConfig{Type: kapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "scan-mem", FromName: "hub-config-resources"})
	hubScanEmptyDir, _ := CreateEmptyDirVolume("dir-scan", "1G")
	hubScanContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-scan", Image: fmt.Sprintf("%s/%s/hub-scan:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.ScanMemoryLimit, MaxMem: hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationBidirectional},
			{Name: "dir-scan", MountPath: "/opt/blackduck/hub/hub-scan/security", Propagation: kapi.MountPropagationBidirectional}},
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
	hubAuthEmptyDir, _ := CreateEmptyDirVolume("dir-authentication", "1G")
	hubAuthContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/hub-authentication:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationBidirectional},
			{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig: &kapi.PortConfig{ContainerPort: authenticationPort, Protocol: kapi.ProtocolTCP},
	}
	// hubAuthInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"chmod", "-cR", "777", "/opt/blackduck/hub/hub-authentication/security"}},
	// 	VolumeMounts:    []*kapi.VolumeMountConfig{&kapi.VolumeMountConfig{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig:      &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	hubAuth := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "hub-authentication", Replicas: IntToInt32(1)},
		[]*api.Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("hubAuth : %v\n", hubAuthc.GetObj())
	deployer.AddDeployment(hubAuth)

	// webapp-logstash
	webappEnvs := allConfigEnv
	webappEnvs = append(webappEnvs, &kapi.EnvConfig{Type: kapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", createHub.Spec.Namespace), "ext4")
	webappEmptyDir, _ := CreateEmptyDirVolume("dir-webapp", "1G")
	webappContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/hub-webapp:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.WebappMemoryLimit, MaxMem: hubContainerFlavor.WebappMemoryLimit, MinCPU: hubContainerFlavor.WebappCPULimit,
			MaxCPU: hubContainerFlavor.WebappCPULimit},
		EnvConfigs: webappEnvs,
		VolumeMounts: []*kapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: kapi.MountPropagationBidirectional},
			{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: kapi.MountPropagationBidirectional},
			{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig: &kapi.PortConfig{ContainerPort: webappPort, Protocol: kapi.ProtocolTCP},
	}
	// webappLogStashInitContainerConfig := &api.Container{
	// 	ContainerConfig: &kapi.ContainerConfig{Name: "alpine", Image: "alpine", PullPolicy: kapi.PullAlways, Command: []string{"sh", "-c", "chmod -cR 777 /var/lib/logstash/data && chmod -cR 777 /opt/blackduck/hub/"}},
	// 	VolumeMounts: []*kapi.VolumeMountConfig{
	// 		&kapi.VolumeMountConfig{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: kapi.MountPropagationBidirectional},
	// 		&kapi.VolumeMountConfig{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: kapi.MountPropagationBidirectional}},
	// 	PortConfig: &kapi.PortConfig{ContainerPort: "3001", Protocol: kapi.ProtocolTCP},
	// }
	// logStashGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-logstash", fmt.Sprintf("%s-%s", "logstash-disk", createHub.Spec.Namespace), "ext4")
	logstashEmptyDir, _ := CreateEmptyDirVolume("dir-logstash", "1G")
	logstashContainerConfig := &api.Container{
		ContainerConfig: &kapi.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/hub-logstash:%s", createHub.Spec.DockerRegistry, createHub.Spec.DockerRepo, createHub.Spec.HubVersion),
			PullPolicy: kapi.PullAlways, MinMem: hubContainerFlavor.LogstashMemoryLimit, MaxMem: hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*kapi.VolumeMountConfig{{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: kapi.MountPropagationBidirectional}},
		PortConfig:   &kapi.PortConfig{ContainerPort: logstashPort, Protocol: kapi.ProtocolTCP},
	}
	webappLogstash := CreateDeploymentFromContainer(&kapi.DeploymentConfig{Namespace: createHub.Spec.Namespace, Name: "webapp-logstash", Replicas: IntToInt32(1)}, []*api.Container{webappContainerConfig, logstashContainerConfig},
		[]*components.Volume{webappEmptyDir, logstashEmptyDir, dbSecretVolume, dbEmptyDir}, []*api.Container{},
		[]kapi.AffinityConfig{})
	// log.Infof("webappLogstash : %v\n", webappLogstashc.GetObj())
	deployer.AddDeployment(webappLogstash)

	deployer.AddService(CreateService("cfssl", "cfssl", createHub.Spec.Namespace, cfsslPort, cfsslPort, false))
	deployer.AddService(CreateService("zookeeper", "zookeeper", createHub.Spec.Namespace, zookeeperPort, zookeeperPort, false))
	deployer.AddService(CreateService("webserver", "webserver", createHub.Spec.Namespace, "443", webserverPort, false))
	deployer.AddService(CreateService("webserver-exp", "webserver", createHub.Spec.Namespace, "443", webserverPort, true))
	deployer.AddService(CreateService("webapp", "webapp-logstash", createHub.Spec.Namespace, webappPort, webappPort, false))
	deployer.AddService(CreateService("logstash", "webapp-logstash", createHub.Spec.Namespace, logstashPort, logstashPort, false))
	deployer.AddService(CreateService("solr", "solr", createHub.Spec.Namespace, solrPort, solrPort, false))
	deployer.AddService(CreateService("documentation", "documentation", createHub.Spec.Namespace, documentationPort, documentationPort, false))
	deployer.AddService(CreateService("scan", "hub-scan", createHub.Spec.Namespace, scannerPort, scannerPort, false))
	deployer.AddService(CreateService("authentication", "hub-authentication", createHub.Spec.Namespace, authenticationPort, authenticationPort, false))
	deployer.AddService(CreateService("registration", "registration", createHub.Spec.Namespace, registrationPort, registrationPort, false))
}

func (hc *Creater) getLoadBalancerIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		service, err := GetService(hc.KubeClient, namespace, serviceName)
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
