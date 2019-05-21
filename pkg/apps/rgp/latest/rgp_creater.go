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

package rgp

import (
	"database/sql"
	"fmt"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	rgpapi "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	pg "github.com/blackducksoftware/synopsys-operator/pkg/apps/database/postgres"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	rgpclientset "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	v1_batch "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater stores the configuration and clients to create specific versions of Rgp
type Creater struct {
	Config      *protoform.Config
	KubeConfig  *rest.Config
	KubeClient  *kubernetes.Clientset
	RgpClient   *rgpclientset.Clientset
	RouteClient *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, rgpClient *rgpclientset.Clientset, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, RgpClient: rgpClient, RouteClient: routeClient}
}

// Versions is an Interface function that returns the versions supported by this Creater
func (c *Creater) Versions() []string {
	return GetVersions()
}

// Ensure is an Interface function that will make sure the instance is correctly deployed or deploy it if needed
func (c *Creater) Ensure(rgp *rgpapi.Rgp) error {
	// Get Kubernetes Components for the Rgp
	specConfig := NewSpecConfig(&rgp.Spec)
	componentList := &api.ComponentList{}

	log.Debugf("Create Rgp details for %s: %+v", rgp.Spec.Namespace, rgp.Spec)
	_, err := util.GetNamespace(c.KubeClient, rgp.Spec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", rgp.Spec.Namespace)
		util.CreateNamespace(c.KubeClient, rgp.Spec.Namespace)
	}

	// Mongo
	mongoClaim, err := util.CreatePersistentVolumeClaim("mongodb", rgp.Spec.Namespace, "20Gi", rgp.Spec.StorageClass, horizonapi.ReadWriteOnce)
	if err != nil {
		return fmt.Errorf("unable to create mongo persistent volume claim due to %+v", err)
	}
	mongo := Mongo{
		Namespace: rgp.Spec.Namespace,
		PVCName:   "mongodb",
		Image:     "gcr.io/snps-swip-staging/swip_mongodb:0.0.6",
		MinCPU:    "250m",
		MinMemory: "8Gi",
		Port:      27017,
		Labels:    map[string]string{"app": "rgp", "component": "mongo"},
	}

	mongoDeployer, _ := deployer.NewDeployer(c.KubeConfig)
	mongorc, _ := mongo.GetMongoReplicationController()
	mongosvc := mongo.GetMongoService()
	mongoDeployer.AddComponent(horizonapi.ReplicationControllerComponent, mongorc)
	mongoDeployer.AddComponent(horizonapi.ServiceComponent, mongosvc)
	mongoDeployer.AddComponent(horizonapi.PersistentVolumeClaimComponent, mongoClaim)
	err = mongoDeployer.Run()
	if err != nil {
		return err
	}
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, mongorc)
	componentList.Services = append(componentList.Services, mongosvc)
	componentList.PersistentVolumeClaims = append(componentList.PersistentVolumeClaims, mongoClaim)

	// Postgres
	pw, err := util.RandomString(12)
	if err != nil {
		return err
	}

	c.KubeClient.CoreV1().Secrets(rgp.Spec.Namespace).Create(&v12.Secret{
		ObjectMeta: v13.ObjectMeta{
			Name: "db-creds",
		},
		StringData: map[string]string{
			"POSTGRES_PASSWORD": pw,
		},
		Type: v12.SecretTypeOpaque,
	})

	postgresClaim, _ := util.CreatePersistentVolumeClaim("postgres", rgp.Spec.Namespace, "20Gi", rgp.Spec.StorageClass, horizonapi.ReadWriteOnce)
	postgres := pg.Postgres{
		Namespace:              rgp.Spec.Namespace,
		PVCName:                "postgres",
		Port:                   5432,
		Image:                  "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
		MinCPU:                 "",
		MaxCPU:                 "",
		MinMemory:              "",
		MaxMemory:              "",
		Database:               "test",
		User:                   "postgres",
		PasswordSecretName:     "db-creds",
		UserPasswordSecretKey:  "POSTGRES_PASSWORD",
		AdminPasswordSecretKey: "POSTGRES_PASSWORD",
		EnvConfigMapRefs:       []string{},
		Labels:                 map[string]string{"app": "rgp", "component": "postgres"},
	}

	postgresDeployer, _ := deployer.NewDeployer(c.KubeConfig)
	postgresrc, _ := postgres.GetPostgresReplicationController()
	postgressvc := postgres.GetPostgresService()
	postgresDeployer.AddComponent(horizonapi.ReplicationControllerComponent, postgresrc)
	postgresDeployer.AddComponent(horizonapi.ServiceComponent, postgressvc)
	postgresDeployer.AddComponent(horizonapi.PersistentVolumeClaimComponent, postgresClaim)
	err = postgresDeployer.Run()
	if err != nil {
		log.Error(err)
		return err
	}
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, postgresrc)
	componentList.Services = append(componentList.Services, postgressvc)
	componentList.PersistentVolumeClaims = append(componentList.PersistentVolumeClaims, postgresClaim)

	// Validate postgres pod is cloned/backed up
	err = util.WaitForServiceEndpointReady(c.KubeClient, rgp.Spec.Namespace, "postgres")
	if err != nil {
		return err
	}

	err = util.ValidatePodsAreRunningInNamespace(c.KubeClient, rgp.Spec.Namespace, 600)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.dbInit(rgp.Spec.Namespace, pw)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.init(&rgp.Spec, componentList)
	if err != nil {
		log.Error(err)
		return err
	}

	// Add rgp specific components
	specConfig.AddComponents(componentList)

	// Update components in cluster
	commonConfig := crdupdater.NewCRUDComponents(c.KubeConfig, c.KubeClient, c.Config.DryRun, false, rgp.Spec.Namespace, componentList, "app=rgp")
	_, errors := commonConfig.CRUDComponents()
	if len(errors) > 0 {
		return fmt.Errorf("unable to update Rgp components due to %+v", errors)
	}

	err = c.createIngress(&rgp.Spec)
	if err != nil {
		return err
	}

	return nil
}

// createIngress creates the ingress
func (c *Creater) createIngress(spec *rgpapi.RgpSpec) error {
	_, err := c.KubeClient.ExtensionsV1beta1().Ingresses(spec.Namespace).Create(&v1beta1.Ingress{
		ObjectMeta: v13.ObjectMeta{
			Name: "rgp",
			Annotations: map[string]string{
				"ingress.kubernetes.io/rewrite-target":       "/",
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
				"kubernetes.io/ingress.class":                spec.IngressClass,
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts: []string{spec.IngressHost},
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: spec.IngressHost,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{
							{
								Path: "/api/auth/v0",
								Backend: v1beta1.IngressBackend{
									ServiceName: "auth-server",
									ServicePort: intstr.FromInt(8080),
								},
							},
							{
								Path: "/api/auth",
								Backend: v1beta1.IngressBackend{
									ServiceName: "auth-server",
									ServicePort: intstr.FromInt(8080),
								},
							},
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
									ServiceName: "report-service",
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

// dbInit create the the databases
func (c *Creater) dbInit(namespace string, pw string) error {
	databaseName := "postgres"
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", namespace)

	postgresDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", pw, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}

	for {
		log.Debug("executing SELECT 1")
		_, err := postgresDB.Exec("SELECT 1;")
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
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

func (c *Creater) startJobAndWaitUntilCompletion(namespace string, timeoutValue time.Duration, job *v1_batch.Job) error {
	job, err := c.KubeClient.BatchV1().Jobs(namespace).Create(job)
	if err != nil {
		return err
	}
	timeout := time.After(timeoutValue)
	tick := time.NewTicker(10 * time.Second)

L:
	for {
		select {
		case <-timeout:
			tick.Stop()
			return fmt.Errorf("job failed")

		case <-tick.C:
			job, err = c.KubeClient.BatchV1().Jobs(job.Namespace).Get(job.Name, v13.GetOptions{})
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
	return nil
}
