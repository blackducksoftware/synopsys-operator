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

package gr

import (
	"database/sql"
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	deployer2 "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	"github.com/blackducksoftware/synopsys-operator/pkg/gr/containers"
	v14 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	v12 "k8s.io/api/rbac/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/gr/v1"
	grclientset "github.com/blackducksoftware/synopsys-operator/pkg/gr/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create the Blackduck
type Creater struct {
	kubeConfig *rest.Config
	kubeClient *kubernetes.Clientset
	grClient   *grclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, grClient *grclientset.Clientset) *Creater {
	return &Creater{kubeConfig: kubeConfig, kubeClient: kubeClient, grClient: grClient}
}

// Delete will delete
func (c *Creater) Delete(namespace string) {
	// TODO
}

// Create will create
func (c *Creater) Create(spec *v1.GrSpec) error {
	log.Debugf("Create Gr details for %s: %+v", spec.Namespace, spec)

	_, err := util.GetNamespace(c.kubeClient, spec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", spec.Namespace)
		util.CreateNamespace(c.kubeClient, spec.Namespace)
	}

	// Postgres
	c.kubeClient.CoreV1().Secrets(spec.Namespace).Create(&v14.Secret{
		ObjectMeta: v13.ObjectMeta{
			Name: "db-creds",
		},
		StringData: map[string]string{
			"POSTGRES_USER_PASSWORD_FILE":  "test",
			"POSTGRES_ADMIN_PASSWORD_FILE": "test",
		},
		Type: v14.SecretTypeOpaque,
	})
	//
	postgres := apps.Postgres{
		Namespace: spec.Namespace,
		//PVCName:                "blackduck-postgres",
		Port:                   "5432",
		Image:                  "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
		MinCPU:                 "",
		MaxCPU:                 "",
		MinMemory:              "",
		MaxMemory:              "",
		Database:               "postgres",
		User:                   "admin",
		PasswordSecretName:     "db-creds",
		UserPasswordSecretKey:  "POSTGRES_USER_PASSWORD_FILE",
		AdminPasswordSecretKey: "POSTGRES_ADMIN_PASSWORD_FILE",
		EnvConfigMapRefs:       []string{},
	}

	postgresDeployer, _ := deployer2.NewDeployer(c.kubeConfig)
	postgresDeployer.AddReplicationController(postgres.GetPostgresReplicationController())
	postgresDeployer.AddService(postgres.GetPostgresService())
	err = postgresDeployer.Run()
	if err != nil {
		log.Error(err)
		return err
	}

	err = util.ValidatePodsAreRunningInNamespace(c.kubeClient, spec.Namespace, 600)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.dbInit(spec.Namespace)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.init(spec)
	if err != nil {
		log.Error(err)
		return err
	}

	//Deploy
	gr := containers.NewGrDeployer(spec, c.kubeConfig, spec)
	deployer, err := gr.GetDeployer()
	if err != nil {
		log.Errorf("unable to get gr components for %s due to %+v", spec.Namespace, err)
		return err
	}


	err = deployer.Run()
	if err != nil {
		return err
	}

	return c.createIngress(spec)
}

// init deploys  minio, vault and consul
func (c *Creater) init(spec *v1.GrSpec) error {
	const vaultConfig = `{"listener":{"tcp":{"address":"[::]:8200","cluster_address":"[::]:8201","tls_cert_file":"/vault/tls/tls.crt","tls_cipher_suites":"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA","tls_disable":false,"tls_key_file":"/vault/tls/tls.key","tls_prefer_server_cipher_suites":true}},"storage":{"consul":{"address":"consul:8500","path":"vault"}}}`

	// Minio
	minioclaim, _ := util.CreatePersistentVolumeClaim("minio", spec.Namespace, "1Gi", "", horizonapi.ReadWriteOnce)
	minioDeployer, _ := deployer2.NewDeployer(c.kubeConfig)

	// TODO generate random password
	minioCreater := apps.NewMinio(spec.Namespace, "minio", "aaaa2wdadwdawdawd", "b2112r43rfefefbbb")
	minioDeployer.AddSecret(minioCreater.GetSecret())
	minioDeployer.AddService(minioCreater.GetServices())
	minioDeployer.AddDeployment(minioCreater.GetDeployment())
	minioDeployer.AddPVC(minioclaim)
	err := minioDeployer.Run()
	if err != nil {
		return err
	}

	// Consul
	consulDeployer, _ := deployer2.NewDeployer(c.kubeConfig)
	consulCreater := apps.NewConsul(spec.Namespace, spec.StorageClass)
	consulDeployer.AddService(consulCreater.GetConsulServices())
	consulDeployer.AddStatefulSet(consulCreater.GetConsulStatefulSet())
	consulDeployer.AddSecret(consulCreater.GetConsulSecrets())

	err = consulDeployer.Run()
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	// Vault Init - Generate Root CA and auth certs
	// This will create the following secrets :
	// - auth-client-tls-certificate
	// - auth-server-tls-certificate
	// - vault-ca-certificate
	// - vault-tls-certificate
	err = c.vaultInit(spec.Namespace)
	if err != nil {
		return err
	}

	// Vault
	vaultDeployer, _ := deployer2.NewDeployer(c.kubeConfig)

	vaultCreater := apps.NewVault(spec.Namespace, vaultConfig, map[string]string{
		"vault-tls-certificate" : "/vault/tls",
	}, "/vault/tls/ca.crt")
	vaultDeployer.AddService(vaultCreater.GetVaultServices())

	// Inject auto-unseal sidecar
	vaultInit := Vault{spec.Namespace}
	vaultPod := vaultCreater.GetPod()
	vaultPod.AddVolume(components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-init-secret",
		MapOrSecretName: "vault-init-secret",
	}))

	vaultPod.AddContainer(vaultInit.GetSidecarUnsealContainer())

	vaultDeployer.AddDeployment(util.CreateDeployment(&horizonapi.DeploymentConfig{
		Name:      "vault",
		Namespace: spec.Namespace,
		Replicas:  util.IntToInt32(3),
	}, vaultPod))
	vaultDeployer.AddConfigMap(vaultCreater.GetVaultConfigConfigMap())
	err = vaultDeployer.Run()
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	return err
}

// vaultInit start the vault initialization job
func (c *Creater) vaultInit(namespace string) error {
	// Init
	_, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Create(&v14.ServiceAccount{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
	})
	if err != nil {
		return err
	}
	_, err = c.kubeClient.RbacV1().Roles(namespace).Create(&v12.Role{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
		Rules: []v12.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs: []string{
					"get",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = c.kubeClient.RbacV1().RoleBindings(namespace).Create(&v12.RoleBinding{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
		Subjects: []v12.Subject{
			{
				Kind: "ServiceAccount",
				Name: "vault-init",
			},
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "vault-init",
		},
	})
	if err != nil {
		return err
	}

	// Start job and create CM
	vaultInit := Vault{namespace}
	job, err := c.kubeClient.BatchV1().Jobs(namespace).Create(vaultInit.GetJob())
	if err != nil {
		log.Print(err)
		return err
	}
	timeout := time.After(30 * time.Minute)
	tick := time.NewTicker(10 * time.Second)

L:
	for {
		select {
		case <-timeout:
			tick.Stop()
			return fmt.Errorf("vault-init job failed")

		case <-tick.C:
			job, err = c.kubeClient.BatchV1().Jobs(job.Namespace).Get(job.Name, v13.GetOptions{})
			if err != nil {
				tick.Stop()
				return err
			}
			if job.Status.Succeeded > 0 {
				tick.Stop()
				break L
			}
		}
	}

	vaultInitDeploy, _ := deployer2.NewDeployer(c.kubeConfig)
	vaultInitDeploy.AddConfigMap(vaultInit.GetConfigmap())
	vaultInitDeploy.AddDeployment(vaultInit.GetDeployment())
	err = vaultInitDeploy.Run()
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// dbInit create the the databases
func (c *Creater) dbInit(namespace string) error {
	databaseName := "postgres"
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", namespace)
	postgresDB, err := OpenDatabaseConnection(hostName, databaseName, "admin", "test", "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}

	_, err = postgresDB.Exec("CREATE DATABASE \"tools-portfolio\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"rp-portfolio\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"report-service\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"issue-manager\";")
	if err != nil {
		return err
	}
	return nil
}

// createIngress creates the ingress
func (c *Creater) createIngress(spec *v1.GrSpec) error {
	_, err := c.kubeClient.ExtensionsV1beta1().Ingresses(spec.Namespace).Create(&v1beta1.Ingress{
		ObjectMeta: v13.ObjectMeta{
			Name: "rgp",
			Annotations: map[string]string{
				"ingress.kubernetes.io/rewrite-target": "/",
				"kubernetes.io/ingress.class": spec.IngressClass,
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: spec.IngressHost,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							[]v1beta1.HTTPIngressPath{
								{
									Path: "/reporting",
									Backend: v1beta1.IngressBackend{
										ServiceName: "frontend-service",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/reporting/tps",
									Backend: v1beta1.IngressBackend{
										ServiceName: "tools-portfolio-service",
										ServicePort: intstr.FromInt(60281),
									},
								},
								{
									Path: "/reporting/im",
									Backend: v1beta1.IngressBackend{
										ServiceName: "rp-issue-manager",
										ServicePort: intstr.FromInt(6888),
									},
								},
								{
									Path: "/reporting/rpps",
									Backend: v1beta1.IngressBackend{
										ServiceName: "rp-portfolio-service",
										ServicePort: intstr.FromInt(60289),
									},
								},
								{
									Path: "/reporting/rs",
									Backend: v1beta1.IngressBackend{
										ServiceName: "rp-issue-manager",
										ServicePort: intstr.FromInt(7979),
									},
								},
							},
						},
					},
				},
			},
		},
	})
	return err
}

// OpenDatabaseConnection open a connection to the database
func OpenDatabaseConnection(hostName string, dbName string, user string, password string, sqlType string) (*sql.DB, error) {
	// Note that sslmode=disable is required it does not mean that the connection
	// is unencrypted. All connections via the proxy are completely encrypted.
	log.Debug("attempting to open database connection")
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable connect_timeout=10", hostName, dbName, user, password)
	db, err := sql.Open(sqlType, dsn)
	//defer db.Close()
	if err == nil {
		log.Debug("connected to database ")
	}
	return db, err
}
