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

package blackduck

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	bdutil "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"

	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest/containers"
)

// GetComponents returns the blackduck components
func (hc *Creater) GetComponents(blackduck *v1.Blackduck) (*api.ComponentList, error) {

	componentList := &api.ComponentList{}

	// Get the flavor
	flavor, err := hc.getContainersFlavor(blackduck)
	if err != nil {
		return nil, err
	}

	// Get Db creds
	var adminPassword, userPassword string
	if blackduck.Spec.ExternalPostgres != nil {
		adminPassword = blackduck.Spec.ExternalPostgres.PostgresAdminPassword
		userPassword = blackduck.Spec.ExternalPostgres.PostgresAdminPassword
	} else {
		adminPassword, userPassword, _, err = bdutil.GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)
		if err != nil {
			return nil, err
		}
	}

	containerCreater := containers.NewCreater(hc.Config, &blackduck.Spec, flavor)

	// Configmap
	componentList.ConfigMaps = append(componentList.ConfigMaps, containerCreater.GetConfigmaps()...)

	//Secrets
	// nginx certificatea
	cert, key, err := hc.getTLSCertKeyOrCreate(blackduck)
	if err != nil {
		return nil, err
	}

	componentList.Secrets = append(componentList.Secrets, containerCreater.GetSecrets(adminPassword, userPassword, cert, key)...)

	// cfssl
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetCfsslDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetCfsslService())

	// nginx
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetWebserverDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetWebServerService())

	// documentation
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetDocumentationDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetDocumentationService())

	// solr
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetSolrDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetSolrService())

	// registration
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetRegistrationDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetRegistrationService())

	// zookeeper
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetZookeeperDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetZookeeperService())

	// jobRunner
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetJobRunnerDeployment())

	// hub-scan
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetScanDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetScanService())

	// hub-authentication
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetAuthenticationDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetAuthenticationService())

	// webapp-logstash
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetWebappLogstashDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetWebAppService())
	componentList.Services = append(componentList.Services, containerCreater.GetLogStashService())

	//Upload cache
	//As part of Black Duck 2019.4.0, upload cache is part of Black Duck
	componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetUploadCacheDeployment())
	componentList.Services = append(componentList.Services, containerCreater.GetUploadCacheService())

	// Service account - https://github.com/blackducksoftware/synopsys-operator/issues/95
	//componentList.ServiceAccounts= append(componentList.ServiceAccounts, containerCreater.GetServiceAccount())
	//componentList.ClusterRoleBindings = append(componentList.ClusterRoleBindings, containerCreater.GetClusterRoleBinding())

	if hc.isBinaryAnalysisEnabled(&blackduck.Spec) {
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetUploadCacheDeployment())
		componentList.Services = append(componentList.Services, containerCreater.GetUploadCacheService())

		// Binary Scanner
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetBinaryScannerDeployment())

		// Rabbitmq
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, containerCreater.GetRabbitmqDeployment())
		componentList.Services = append(componentList.Services, containerCreater.GetRabbitmqService())
	}

	// Add Expose service
	if svc := hc.getExposeService(blackduck); svc != nil {
		componentList.Services = append(componentList.Services, svc)
	}
	return componentList, nil
}

func (hc *Creater) getExposeService(bd *v1.Blackduck) *components.Service {
	containerCreater := containers.NewCreater(hc.Config, &bd.Spec, nil)
	var svc *components.Service

	switch strings.ToUpper(bd.Spec.ExposeService) {
	case "NODEPORT":
		svc = containerCreater.GetWebServerNodePortService()
		break
	case "LOADBALANCER":
		svc = containerCreater.GetWebServerLoadBalancerService()
		break
	default:
	}
	return svc
}

// GetPVC returns the PVCs
func (hc *Creater) GetPVC(blackduck *v1.Blackduck) []*components.PersistentVolumeClaim {
	containerCreater := containers.NewCreater(hc.Config, &blackduck.Spec, nil)
	return containerCreater.GetPVCs()
}

func (hc *Creater) getTLSCertKeyOrCreate(blackduck *v1.Blackduck) (string, string, error) {
	if strings.EqualFold(blackduck.Spec.CertificateName, "manual") {
		return blackduck.Spec.Certificate, blackduck.Spec.CertificateKey, nil
	}

	secret, err := util.GetSecret(hc.KubeClient, blackduck.Spec.Namespace, "blackduck-certificate")
	if err == nil {
		data := secret.Data
		if len(data) >= 2 {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if !certok || !keyok {
				util.DeleteSecret(hc.KubeClient, blackduck.Spec.Namespace, "blackduck-certificate")
			} else {
				return string(cert), string(key), nil
			}
		}
	}

	// Cert copy
	if !strings.EqualFold(blackduck.Spec.CertificateName, "default") {
		secret, err := util.GetSecret(hc.KubeClient, blackduck.Spec.CertificateName, "blackduck-certificate")
		if err == nil {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if certok && keyok {
				return string(cert), string(key), nil
			}
		}
	}

	// Default
	return CreateSelfSignedCert()
}

// addAnyUIDToServiceAccount adds the capability to run as 1000 for nginx or other special IDs.  For example, the binaryscanner
// needs to run as root and we plan to add that into protoform in 2.1 / 3.0.
func (hc *Creater) addAnyUIDToServiceAccount(createHub *v1.BlackduckSpec) error {
	if hc.osSecurityClient != nil {
		log.Debugf("Adding anyuid securitycontextconstraint to the service account %s", createHub.Namespace)
		scc, err := util.GetOpenShiftSecurityConstraint(hc.osSecurityClient, "anyuid")
		if err != nil {
			return fmt.Errorf("failed to get scc anyuid: %v", err)
		}

		serviceAccount := createHub.Namespace

		// Only add the service account if it isn't already in the list of users for the privileged scc
		exists := false
		for _, user := range scc.Users {
			if strings.Compare(user, serviceAccount) == 0 {
				exists = true
				break
			}
		}

		if !exists {
			scc.Users = append(scc.Users, serviceAccount)

			_, err = hc.osSecurityClient.SecurityContextConstraints().Update(scc)
			if err != nil {
				return fmt.Errorf("failed to update scc anyuid: %v", err)
			}
		}
	}
	return nil
}

// AddExposeServices add the nodeport / LB services
func (hc *Creater) AddExposeServices(deployer *horizon.Deployer, createHub *v1.BlackduckSpec) {
	containerCreater := containers.NewCreater(hc.Config, createHub, nil)
	deployer.AddService(containerCreater.GetWebServerNodePortService())
	deployer.AddService(containerCreater.GetWebServerLoadBalancerService())
}
