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

package blackduck

import (
	"fmt"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/blackduck/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// AddToDeployer will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
// TODO doc what 'allConfigEnv' actually is ???
func (hc *Creater) AddToDeployer(deployer *horizon.Deployer, createHub *v1.BlackduckSpec, hubContainerFlavor *containers.ContainerFlavor, allConfigEnv []*horizonapi.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
	}

	dbSecretVolume := components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]horizonapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: util.IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})

	// dbEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("cloudsql")

	var proxySecretVolume *components.Volume

	proxyCertSecret, err := util.GetSecret(hc.KubeClient, createHub.Namespace, "blackduck-proxy-certificate")
	if err == nil {
		if _, err := hc.stringToCertificate(string(proxyCertSecret.Data["HUB_PROXY_CERT_FILE"])); err == nil {
			proxySecretVolume = components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
				VolumeName:      "blackduck-proxy-certificate",
				MapOrSecretName: "blackduck-proxy-certificate",
				Items: map[string]horizonapi.KeyAndMode{
					"HUB_PROXY_CERT_FILE": {KeyOrPath: "HUB_PROXY_CERT_FILE", Mode: util.IntToInt32(420)},
				},
				DefaultMode: util.IntToInt32(420),
			})
		}
	}

	containerCreater := containers.NewCreater(hc.Config, createHub, hubContainerFlavor, hubConfigEnv, allConfigEnv, dbSecretVolume, proxySecretVolume)

	// cfssl
	deployer.AddReplicationController(containerCreater.GetCfsslDeployment())
	deployer.AddService(containerCreater.GetCfsslService())

	// nginx certificate
	for {
		secret, err := util.GetSecret(hc.KubeClient, createHub.Namespace, "blackduck-certificate")
		if err != nil {
			log.Errorf("unable to get the secret in %s due to %+v", createHub.Namespace, err)
			break
		}
		data := secret.Data
		if len(data) > 0 {
			break
		}
		time.Sleep(10 * time.Second)
	}

	// nginx
	deployer.AddReplicationController(containerCreater.GetWebserverDeployment())
	deployer.AddService(containerCreater.GetWebServerService())
	deployer.AddService(containerCreater.GetWebServerNodePortService())
	deployer.AddService(containerCreater.GetWebServerLoadBalancerService())

	// documentation
	deployer.AddReplicationController(containerCreater.GetDocumentationDeployment())
	deployer.AddService(containerCreater.GetDocumentationService())

	// solr
	deployer.AddReplicationController(containerCreater.GetSolrDeployment())
	deployer.AddService(containerCreater.GetSolrService())

	// registration
	deployer.AddReplicationController(containerCreater.GetRegistrationDeployment())
	deployer.AddService(containerCreater.GetRegistrationService())

	// zookeeper
	deployer.AddReplicationController(containerCreater.GetZookeeperDeployment())
	deployer.AddService(containerCreater.GetZookeeperService())

	// jobRunner
	deployer.AddReplicationController(containerCreater.GetJobRunnerDeployment())

	// hub-scan
	deployer.AddReplicationController(containerCreater.GetScanDeployment())
	deployer.AddService(containerCreater.GetScanService())

	// hub-authentication
	deployer.AddReplicationController(containerCreater.GetAuthenticationDeployment())
	deployer.AddService(containerCreater.GetAuthenticationService())

	// webapp-logstash
	deployer.AddReplicationController(containerCreater.GetWebappLogstashDeployment())
	deployer.AddService(containerCreater.GetWebAppService())
	deployer.AddService(containerCreater.GetLogStashService())
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
