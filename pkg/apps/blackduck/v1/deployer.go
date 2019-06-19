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

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// getPostgresComponents returns the blackduck postgres component list
func (hc *Creater) getPostgresComponents(blackduck *blackduckapi.Blackduck) (*api.ComponentList, error) {
	componentList := &api.ComponentList{}

	// Get Containers Flavor
	hubContainerFlavor, err := hc.getContainersFlavor(blackduck)
	if err != nil {
		return nil, err
	}

	containerCreater := containers.NewCreater(hc.config, hc.kubeConfig, hc.kubeClient, blackduck, hubContainerFlavor, false)
	// Get Db creds
	var adminPassword, userPassword, postgresPassword string
	if blackduck.Spec.ExternalPostgres != nil {

		adminPassword, err = util.Base64Decode(blackduck.Spec.ExternalPostgres.PostgresAdminPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: unable to decode external Postgres adminPassword due to: %+v", blackduck.Spec.Namespace, err)
		}

		userPassword, err = util.Base64Decode(blackduck.Spec.ExternalPostgres.PostgresUserPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: unable to decode external Postgres userPassword due to: %+v", blackduck.Spec.Namespace, err)
		}

	} else {

		adminPassword, err = util.Base64Decode(blackduck.Spec.AdminPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: unable to decode adminPassword due to: %+v", blackduck.Spec.Namespace, err)
		}

		userPassword, err = util.Base64Decode(blackduck.Spec.UserPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: unable to decode userPassword due to: %+v", blackduck.Spec.Namespace, err)
		}

		postgresPassword, err = util.Base64Decode(blackduck.Spec.PostgresPassword)
		if err != nil {
			return nil, fmt.Errorf("%v: unable to decode postgresPassword due to: %+v", blackduck.Spec.Namespace, err)
		}

	}

	postgres := containerCreater.GetPostgres()
	if blackduck.Spec.ExternalPostgres == nil {
		postgresRc, err := postgres.GetPostgresReplicationController()
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, postgresRc)
		componentList.Services = append(componentList.Services, postgres.GetPostgresService())
	}
	componentList.ConfigMaps = append(componentList.ConfigMaps, containerCreater.GetPostgresConfigmap())
	componentList.Secrets = append(componentList.Secrets, containerCreater.GetPostgresSecret(adminPassword, userPassword, postgresPassword))

	return componentList, nil
}

// GetComponents returns the components
func (hc *Creater) GetComponents(blackduck *blackduckapi.Blackduck) (*api.ComponentList, error) {

	componentList := &api.ComponentList{}

	// Get the flavor
	flavor, err := hc.getContainersFlavor(blackduck)
	if err != nil {
		return nil, err
	}

	containerCreater := containers.NewCreater(hc.config, hc.kubeConfig, hc.kubeClient, blackduck, flavor, false)

	// Configmap
	componentList.ConfigMaps = append(componentList.ConfigMaps, containerCreater.GetConfigmaps()...)

	//Secrets
	// nginx certificate
	cert, key, err := hc.getTLSCertKeyOrCreate(blackduck)
	if err != nil {
		return nil, err
	}

	componentList.Secrets = append(componentList.Secrets, containerCreater.GetSecrets(cert, key)...)

	// cfssl
	imageName := containerCreater.GetImageTag("blackduck-cfssl")
	if len(imageName) > 0 {
		cfsslRc, err := containerCreater.GetCfsslDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, cfsslRc)
		componentList.Services = append(componentList.Services, containerCreater.GetCfsslService())
	}

	// nginx
	imageName = containerCreater.GetImageTag("blackduck-nginx")
	if len(imageName) > 0 {
		nginxRc, err := containerCreater.GetWebserverDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, nginxRc)
		componentList.Services = append(componentList.Services, containerCreater.GetWebServerService())
	}

	// documentation
	imageName = containerCreater.GetImageTag("blackduck-documentation")
	if len(imageName) > 0 {
		documentationRc, err := containerCreater.GetDocumentationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, documentationRc)
		componentList.Services = append(componentList.Services, containerCreater.GetDocumentationService())
	}

	// solr
	imageName = containerCreater.GetImageTag("blackduck-solr")
	if len(imageName) > 0 {
		solrRc, err := containerCreater.GetSolrDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, solrRc)
		componentList.Services = append(componentList.Services, containerCreater.GetSolrService())
	}

	// registration
	imageName = containerCreater.GetImageTag("blackduck-registration")
	if len(imageName) > 0 {
		registrationRc, err := containerCreater.GetRegistrationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, registrationRc)
		componentList.Services = append(componentList.Services, containerCreater.GetRegistrationService())
	}

	// zookeeper
	imageName = containerCreater.GetImageTag("blackduck-zookeeper")
	if len(imageName) > 0 {
		zookeeperRc, err := containerCreater.GetZookeeperDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, zookeeperRc)
		componentList.Services = append(componentList.Services, containerCreater.GetZookeeperService())
	}

	// jobRunner
	imageName = containerCreater.GetImageTag("blackduck-jobrunner")
	if len(imageName) > 0 {
		jobRunnerRc, err := containerCreater.GetJobRunnerDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, jobRunnerRc)
	}

	// hub-scan
	imageName = containerCreater.GetImageTag("blackduck-scan")
	if len(imageName) > 0 {
		scanRc, err := containerCreater.GetScanDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, scanRc)
		componentList.Services = append(componentList.Services, containerCreater.GetScanService())
	}

	// hub-authentication
	imageName = containerCreater.GetImageTag("blackduck-authentication")
	if len(imageName) > 0 {
		authRc, err := containerCreater.GetAuthenticationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, authRc)
		componentList.Services = append(componentList.Services, containerCreater.GetAuthenticationService())
	}

	// webapp-logstash
	imageName = containerCreater.GetImageTag("blackduck-webapp")
	if len(imageName) > 0 {
		webappLogstashRc, err := containerCreater.GetWebappLogstashDeployment(imageName, containerCreater.GetImageTag("blackduck-logstash"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.ReplicationControllers = append(componentList.ReplicationControllers, webappLogstashRc)
		componentList.Services = append(componentList.Services, containerCreater.GetWebAppService())
		componentList.Services = append(componentList.Services, containerCreater.GetLogStashService())
	}

	if hc.isBinaryAnalysisEnabled(&blackduck.Spec) {
		// Upload cache
		imageName := containerCreater.GetImageTag("blackduck-upload-cache")
		if len(imageName) > 0 {
			uploadCacheRc, err := containerCreater.GetUploadCacheDeployment(imageName)
			if err != nil {
				return nil, errors.Trace(err)
			}
			componentList.ReplicationControllers = append(componentList.ReplicationControllers, uploadCacheRc)
			componentList.Services = append(componentList.Services, containerCreater.GetUploadCacheService())
		}

		// Binary Scanner
		imageName = containerCreater.GetImageTag("appcheck-worker")
		if len(imageName) > 0 {
			binaryScannerRc, err := containerCreater.GetBinaryScannerDeployment(imageName)
			if err != nil {
				return nil, errors.Trace(err)
			}
			componentList.ReplicationControllers = append(componentList.ReplicationControllers, binaryScannerRc)
		}

		// Rabbitmq
		imageName = containerCreater.GetImageTag("rabbitmq")
		if len(imageName) > 0 {
			rabbitmqRc, err := containerCreater.GetRabbitmqDeployment(imageName)
			if err != nil {
				return nil, errors.Trace(err)
			}
			componentList.ReplicationControllers = append(componentList.ReplicationControllers, rabbitmqRc)
			componentList.Services = append(componentList.Services, containerCreater.GetRabbitmqService())
		}
	}

	// Add Expose service
	if svc := hc.getExposeService(blackduck); svc != nil {
		componentList.Services = append(componentList.Services, svc)
	}

	// Add OpenShift routes
	if util.OPENSHIFT == strings.ToUpper(blackduck.Spec.ExposeService) {
		route := containerCreater.GetOpenShiftRoute()
		if route != nil {
			componentList.Routes = []*api.Route{route}
		}
	}
	return componentList, nil
}

func (hc *Creater) getExposeService(bd *blackduckapi.Blackduck) *components.Service {
	containerCreater := containers.NewCreater(hc.config, hc.kubeConfig, hc.kubeClient, bd, nil, false)
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

// GetPVC returns the PVC
func (hc *Creater) GetPVC(blackduck *blackduckapi.Blackduck) []*components.PersistentVolumeClaim {
	containerCreater := containers.NewCreater(hc.config, hc.kubeConfig, hc.kubeClient, blackduck, nil, hc.isBinaryAnalysisEnabled(&blackduck.Spec))
	return containerCreater.GetPVCs()
}

func (hc *Creater) getTLSCertKeyOrCreate(blackduck *blackduckapi.Blackduck) (string, string, error) {
	if len(blackduck.Spec.Certificate) > 0 && len(blackduck.Spec.CertificateKey) > 0 {
		return blackduck.Spec.Certificate, blackduck.Spec.CertificateKey, nil
	}

	// Cert copy
	if len(blackduck.Spec.CertificateName) > 0 && !strings.EqualFold(blackduck.Spec.CertificateName, "default") {
		secret, err := util.GetSecret(hc.kubeClient, blackduck.Spec.CertificateName, util.GetResourceName(blackduck.Name, util.BlackDuckName, "webserver-certificate"))
		if err == nil {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if certok && keyok {
				return string(cert), string(key), nil
			}
		}
	}

	// default cert
	secret, err := util.GetSecret(hc.kubeClient, hc.config.Namespace, "blackduck-certificate")
	if err == nil {
		data := secret.Data
		if len(data) >= 2 {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if !certok || !keyok {
				util.DeleteSecret(hc.kubeClient, blackduck.Spec.Namespace, util.GetResourceName(blackduck.Name, util.BlackDuckName, "webserver-certificate"))
			} else {
				return string(cert), string(key), nil
			}
		}
	}

	// Default
	return CreateSelfSignedCert()
}
