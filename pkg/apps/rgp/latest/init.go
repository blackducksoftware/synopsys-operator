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
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	deployer2 "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	v14 "k8s.io/api/core/v1"
	v12 "k8s.io/api/rbac/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// init deploys  minio, vault and consul
func (c *Creater) init(spec *v1.RgpSpec, componentList *api.ComponentList) error {
	const vaultConfig = `{"listener":{"tcp":{"address":"[::]:8200","cluster_address":"[::]:8201","tls_cert_file":"/vault/tls/tls.crt","tls_cipher_suites":"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA","tls_disable":false,"tls_key_file":"/vault/tls/tls.key","tls_prefer_server_cipher_suites":true}},"storage":{"consul":{"address":"consul:8500","path":"vault"}}}`

	err := c.eventStoreInit(spec, componentList)

	// Minio
	minioclaim, _ := util.CreatePersistentVolumeClaim("minio", spec.Namespace, "1Gi", spec.StorageClass, horizonapi.ReadWriteOnce)
	minioDeployer, _ := deployer2.NewDeployer(c.KubeConfig)

	// TODO generate random password
	minioCreater := NewMinio(spec.Namespace, "minio", "aaaa2wdadwdawdawd", "b2112r43rfefefbbb")
	minioSecret := minioCreater.GetSecret()
	minioService := minioCreater.GetServices()
	minioDeployment := minioCreater.GetDeployment()
	minioDeployer.AddComponent(horizonapi.SecretComponent, minioSecret)
	minioDeployer.AddComponent(horizonapi.ServiceComponent, minioService)
	minioDeployer.AddComponent(horizonapi.DeploymentComponent, minioDeployment)
	minioDeployer.AddComponent(horizonapi.PersistentVolumeClaimComponent, minioclaim)
	err = minioDeployer.Run()
	if err != nil {
		return err
	}
	componentList.Secrets = append(componentList.Secrets, minioSecret)
	componentList.Services = append(componentList.Services, minioService)
	componentList.Deployments = append(componentList.Deployments, minioDeployment)
	componentList.PersistentVolumeClaims = append(componentList.PersistentVolumeClaims, minioclaim)

	// Consul
	consulDeployer, _ := deployer2.NewDeployer(c.KubeConfig)
	consulCreater := NewConsul(spec.Namespace, spec.StorageClass)
	consulServices := consulCreater.GetConsulServices()
	consulStatefulSet := consulCreater.GetConsulStatefulSet()
	consulSecrets := consulCreater.GetConsulSecrets()

	consulDeployer.AddComponent(horizonapi.ServiceComponent, consulServices)
	consulDeployer.AddComponent(horizonapi.StatefulSetComponent, consulStatefulSet)
	consulDeployer.AddComponent(horizonapi.SecretComponent, consulSecrets)

	err = consulDeployer.Run()
	if err != nil {
		return err
	}
	componentList.Secrets = append(componentList.Secrets, consulSecrets)
	componentList.Services = append(componentList.Services, consulServices)
	componentList.StatefulSets = append(componentList.StatefulSets, consulStatefulSet)

	time.Sleep(60 * time.Second)

	// Vault Init - Generate Root CA and auth certs
	// This will create the following secrets :
	// - auth-client-tls-certificate
	// - auth-server-tls-certificate
	// - vault-ca-certificate
	// - vault-tls-certificate
	// - vault-init-secret
	err = c.vaultInit(spec.Namespace, componentList)
	if err != nil {
		return err
	}

	// Vault
	vaultDeployer, _ := deployer2.NewDeployer(c.KubeConfig)

	vaultCreater := NewVault(spec.Namespace, vaultConfig, map[string]string{
		"vault-tls-certificate": "/vault/tls",
	}, "/vault/tls/ca.crt")
	vaultServices := vaultCreater.GetVaultServices()
	vaultDeployer.AddComponent(horizonapi.ServiceComponent, vaultServices)

	// Inject auto-unseal sidecar
	vaultInit := VaultSideCar{spec.Namespace}
	vaultPod := vaultCreater.GetPod()
	vaultPod.AddVolume(components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-init-secret",
		MapOrSecretName: "vault-init-secret",
	}))

	sidecarUnsealContainer, _ := vaultInit.GetSidecarUnsealContainer()
	vaultPod.AddContainer(sidecarUnsealContainer)
	vaultDeployment := util.CreateDeployment(&horizonapi.DeploymentConfig{
		Name:      "vault",
		Namespace: spec.Namespace,
		Replicas:  util.IntToInt32(3),
	}, vaultPod, map[string]string{
		"app":       "rgp",
		"component": "vault",
	})
	vaultConfigMap := vaultCreater.GetVaultConfigConfigMap()

	vaultDeployer.AddComponent(horizonapi.DeploymentComponent, vaultDeployment)
	vaultDeployer.AddComponent(horizonapi.ConfigMapComponent, vaultConfigMap)

	err = vaultDeployer.Run()
	if err != nil {
		return err
	}
	componentList.Services = append(componentList.Services, vaultServices)
	componentList.Deployments = append(componentList.Deployments, vaultDeployment)
	componentList.ConfigMaps = append(componentList.ConfigMaps, vaultConfigMap)

	time.Sleep(30 * time.Second)

	return err
}

func (c *Creater) eventStoreInit(spec *v1.RgpSpec, componentList *api.ComponentList) error {
	eventStore := NewEventstore(spec.Namespace, spec.StorageClass, 100)

	// eventstore
	eventStoreDeployer, _ := deployer2.NewDeployer(c.KubeConfig)
	eventStoreService := eventStore.GetEventStoreService()
	eventStoreStatefulSet := eventStore.GetEventStoreStatefulSet()
	eventStoreDeployer.AddComponent(horizonapi.StatefulSetComponent, eventStoreStatefulSet)
	eventStoreDeployer.AddComponent(horizonapi.ServiceComponent, eventStoreService)
	err := eventStoreDeployer.Run()
	if err != nil {
		return err
	}
	componentList.Services = append(componentList.Services, eventStoreService)
	componentList.StatefulSets = append(componentList.StatefulSets, eventStoreStatefulSet)

	// Create service account
	_, err = c.KubeClient.CoreV1().ServiceAccounts(spec.Namespace).Create(&v14.ServiceAccount{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
		},
	})
	if err != nil {
		return err
	}

	// Create role
	_, err = c.KubeClient.RbacV1().Roles(spec.Namespace).Create(&v12.Role{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
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

	// Bind role to service account
	_, err = c.KubeClient.RbacV1().RoleBindings(spec.Namespace).Create(&v12.RoleBinding{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
		},
		Subjects: []v12.Subject{
			{
				Kind: "ServiceAccount",
				Name: "eventstore-init",
			},
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "eventstore-init",
		},
	})
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	// Init Job
	err = c.startJobAndWaitUntilCompletion(spec.Namespace, 30*time.Minute, eventStore.GetInitJob())
	if err != nil {
		return err
	}

	return nil
}

// vaultInit start the vault initialization job
func (c *Creater) vaultInit(namespace string, componentList *api.ComponentList) error {
	// Init
	serviceAccount, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Create(&v14.ServiceAccount{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
	})
	if err != nil {
		return err
	}
	componentList.ServiceAccounts = append(componentList.ServiceAccounts,
		&components.ServiceAccount{ServiceAccount: serviceAccount})

	_, err = c.KubeClient.RbacV1().Roles(namespace).Create(&v12.Role{
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
	_, err = c.KubeClient.RbacV1().RoleBindings(namespace).Create(&v12.RoleBinding{
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
	vaultInit := VaultSideCar{namespace}
	err = c.startJobAndWaitUntilCompletion(namespace, 30*time.Minute, vaultInit.GetJob())
	if err != nil {
		log.Print(err)
		return err
	}

	vaultInitDeploy, _ := deployer2.NewDeployer(c.KubeConfig)
	vaultInitConfigMap := vaultInit.GetConfigmap()
	vaultInitDeployment := vaultInit.GetDeployment()
	vaultInitDeploy.AddComponent(horizonapi.ConfigMapComponent, vaultInitConfigMap)
	vaultInitDeploy.AddComponent(horizonapi.DeploymentComponent, vaultInitDeployment)
	err = vaultInitDeploy.Run()
	if err != nil {
		log.Print(err)
		return err
	}
	componentList.ConfigMaps = append(componentList.ConfigMaps, vaultInitConfigMap)
	componentList.Deployments = append(componentList.Deployments, vaultInitDeployment)

	return nil
}
