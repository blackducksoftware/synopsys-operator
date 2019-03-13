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

package rgp

import (
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	grclientset "github.com/blackducksoftware/synopsys-operator/pkg/rgp/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/rgp/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
func (c *Creater) Create(spec *v1.RgpSpec) error {
	log.Debugf("Create Gr details for %s: %+v", spec.Namespace, spec)

	_, err := util.GetNamespace(c.kubeClient, spec.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", spec.Namespace)
		util.CreateNamespace(c.kubeClient, spec.Namespace)
	}

	// Mongo

	mongoClaim, _ := util.CreatePersistentVolumeClaim("mongodb", spec.Namespace, "20Gi", spec.StorageClass, horizonapi.ReadWriteOnce)
	mongo := apps.Mongo{
		Namespace: spec.Namespace,
		PVCName:   "mongodb",
		Image:     "gcr.io/snps-swip-staging/swip_mongodb:latest",
		MinCPU:    "250m",
		//MinMemory: "8Gi",
		Port: "27017",
	}

	mongoDeployer, _ := deployer.NewDeployer(c.kubeConfig)
	mongoDeployer.AddReplicationController(mongo.GetMongoReplicationController())
	mongoDeployer.AddService(mongo.GetMongoService())
	mongoDeployer.AddPVC(mongoClaim)
	err = mongoDeployer.Run()
	if err != nil {
		return err
	}

	// Postgres
	pw, err := util.RandomString(12)
	if err != nil {
		return err
	}

	c.kubeClient.CoreV1().Secrets(spec.Namespace).Create(&v12.Secret{
		ObjectMeta: v13.ObjectMeta{
			Name: "db-creds",
		},
		StringData: map[string]string{
			"POSTGRES_PASSWORD": pw,
		},
		Type: v12.SecretTypeOpaque,
	})
	//

	postgresClaim, _ := util.CreatePersistentVolumeClaim("postgres", spec.Namespace, "20Gi", spec.StorageClass, horizonapi.ReadWriteOnce)
	postgres := apps.Postgres{
		Namespace:              spec.Namespace,
		PVCName:                "postgres",
		Port:                   "5432",
		Image:                  "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1",
		MinCPU:                 "",
		MaxCPU:                 "",
		MinMemory:              "",
		MaxMemory:              "",
		Database:               "postgres",
		User:                   "postgres",
		PasswordSecretName:     "db-creds",
		UserPasswordSecretKey:  "POSTGRES_PASSWORD",
		AdminPasswordSecretKey: "POSTGRES_PASSWORD",
		EnvConfigMapRefs:       []string{},
	}

	postgresDeployer, _ := deployer.NewDeployer(c.kubeConfig)
	postgresDeployer.AddReplicationController(postgres.GetPostgresReplicationController())
	postgresDeployer.AddService(postgres.GetPostgresService())
	postgresDeployer.AddPVC(postgresClaim)
	err = postgresDeployer.Run()
	if err != nil {
		log.Error(err)
		return err
	}

	// Validate postgres pod is cloned/backed up
	err = util.WaitForServiceEndpointReady(c.kubeClient, spec.Namespace, "postgres")
	if err != nil {
		return err
	}

	err = util.ValidatePodsAreRunningInNamespace(c.kubeClient, spec.Namespace, 600)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.dbInit(spec.Namespace, pw)
	if err != nil {
		log.Error(err)
		return err
	}

	err = c.init(spec)
	if err != nil {
		log.Error(err)
		return err
	}
	//
	//Deploy
	gr := containers.NewRgpDeployer(spec, c.kubeConfig, spec)
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

// createIngress creates the ingress
func (c *Creater) createIngress(spec *v1.RgpSpec) error {
	_, err := c.kubeClient.ExtensionsV1beta1().Ingresses(spec.Namespace).Create(&v1beta1.Ingress{
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
