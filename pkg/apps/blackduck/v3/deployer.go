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
	"time"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	containers "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v3/containers"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// GetPostgresComponents returns the blackduck postgres component list
func (hc *Creater) GetPostgresComponents(blackduck *blackduckapi.Blackduck) (*api.ComponentList, error) {
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
		postgresDeployment, err := postgres.GetPostgresDeployment()
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, postgresDeployment)
		componentList.Services = append(componentList.Services, postgres.GetPostgresService())
	}
	componentList.ConfigMaps = append(componentList.ConfigMaps, containerCreater.GetPostgresConfigmap())
	componentList.Secrets = append(componentList.Secrets, containerCreater.GetPostgresSecret(adminPassword, userPassword, postgresPassword))

	return componentList, nil
}

// GetComponents returns the blackduck components
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
	cert, key, _ := hc.getTLSCertKeyOrCreate(blackduck)
	if !hc.config.DryRun {
		secret, err := util.GetSecret(hc.kubeClient, hc.config.Namespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find Synopsys Operator blackduck-secret in %s namespace due to %+v", hc.config.Namespace, err)
			return nil, err
		}

		// if Black Duck instance level Seal Key is provided, then use it else use the operator level seal key
		sealKey := secret.Data["SEAL_KEY"]
		if len(blackduck.Spec.SealKey) > 0 {
			sealKeyStr, err := util.Base64Decode(blackduck.Spec.SealKey)
			if err != nil {
				return nil, fmt.Errorf("%v: unable to decode seal key due to: %+v", blackduck.Spec.Namespace, err)
			}
			sealKey = []byte(sealKeyStr)
		}
		componentList.Secrets = append(componentList.Secrets, containerCreater.GetSecrets(cert, key, sealKey)...)
	} else {
		componentList.Secrets = append(componentList.Secrets, containerCreater.GetSecrets(cert, key, []byte{})...)
	}

	// cfssl
	imageName := containerCreater.GetImageTag("blackduck-cfssl")
	if len(imageName) > 0 {
		cfsslDeployment, err := containerCreater.GetCfsslDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, cfsslDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetCfsslService())
	}

	// nginx
	imageName = containerCreater.GetImageTag("blackduck-nginx")
	if len(imageName) > 0 {
		nginxDeployment, err := containerCreater.GetWebserverDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, nginxDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetWebServerService())
	}

	// documentation
	imageName = containerCreater.GetImageTag("blackduck-documentation")
	if len(imageName) > 0 {
		documentationDeployment, err := containerCreater.GetDocumentationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, documentationDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetDocumentationService())
	}

	// TODO: solr is not supported in latest (leaving here in case we consolidate the deployers)
	imageName = containerCreater.GetImageTag("blackduck-solr")
	if len(imageName) > 0 {
		solrDeployment, err := containerCreater.GetSolrDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, solrDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetSolrService())
	}

	// registration
	imageName = containerCreater.GetImageTag("blackduck-registration")
	if len(imageName) > 0 {
		registrationDeployment, err := containerCreater.GetRegistrationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, registrationDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetRegistrationService())
	}

	// zookeeper
	imageName = containerCreater.GetImageTag("blackduck-zookeeper")
	if len(imageName) > 0 {
		zookeeperDeployment, err := containerCreater.GetZookeeperDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, zookeeperDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetZookeeperService())
	}

	// jobRunner
	imageName = containerCreater.GetImageTag("blackduck-jobrunner")
	if len(imageName) > 0 {
		jobRunnerDeployment, err := containerCreater.GetJobRunnerDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, jobRunnerDeployment)
	}

	// hub-scan
	imageName = containerCreater.GetImageTag("blackduck-scan")
	if len(imageName) > 0 {
		scanDeployment, err := containerCreater.GetScanDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, scanDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetScanService())
	}

	// hub-authentication
	imageName = containerCreater.GetImageTag("blackduck-authentication")
	if len(imageName) > 0 {
		authDeployment, err := containerCreater.GetAuthenticationDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, authDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetAuthenticationService())
	}

	// webapp-logstash
	imageName = containerCreater.GetImageTag("blackduck-webapp")
	if len(imageName) > 0 {
		webappLogstashDeployment, err := containerCreater.GetWebappLogstashDeployment(imageName, containerCreater.GetImageTag("blackduck-logstash"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, webappLogstashDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetWebAppService())
		componentList.Services = append(componentList.Services, containerCreater.GetLogStashService())
	}

	//Upload cache
	//As part of Black Duck 2019.4.0, upload cache is part of Black Duck
	imageName = containerCreater.GetImageTag("blackduck-upload-cache")
	if len(imageName) > 0 {
		uploadCacheDeployment, err := containerCreater.GetUploadCacheDeployment(imageName)
		if err != nil {
			return nil, errors.Trace(err)
		}
		componentList.Deployments = append(componentList.Deployments, uploadCacheDeployment)
		componentList.Services = append(componentList.Services, containerCreater.GetUploadCacheService())
	}

	if hc.isBinaryAnalysisEnabled(&blackduck.Spec) {
		// Binary Scanner
		var imageName string
		// If the Black Duck version is greater than or equal to 2019.10.0, use docker.io/sigsynopsys else use docker.io/blackducksoftware
		if isVersionGreaterThanOrEqualTo, _ := util.IsVersionGreaterThanOrEqualTo(blackduck.Spec.Version, 2019, time.October, 0); isVersionGreaterThanOrEqualTo {
			imageName = containerCreater.GetImageTag("appcheck-worker", "docker.io/sigsynopsys")
		} else {
			imageName = containerCreater.GetImageTag("appcheck-worker")
		}

		if len(imageName) > 0 {
			binaryScannerDeployment, err := containerCreater.GetBinaryScannerDeployment(imageName)
			if err != nil {
				return nil, errors.Trace(err)
			}
			componentList.Deployments = append(componentList.Deployments, binaryScannerDeployment)
		}

		// Rabbitmq
		imageName = containerCreater.GetImageTag("rabbitmq")
		if len(imageName) > 0 {
			rabbitmqDeployment, err := containerCreater.GetRabbitmqDeployment(imageName)
			if err != nil {
				return nil, errors.Trace(err)
			}
			componentList.Deployments = append(componentList.Deployments, rabbitmqDeployment)
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
	case util.NODEPORT:
		svc = containerCreater.GetWebServerNodePortService()
		break
	case util.LOADBALANCER:
		svc = containerCreater.GetWebServerLoadBalancerService()
		break
	default:
	}
	return svc
}

// GetPVC returns the PVCs
func (hc *Creater) GetPVC(blackduck *blackduckapi.Blackduck) ([]*components.PersistentVolumeClaim, error) {
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
