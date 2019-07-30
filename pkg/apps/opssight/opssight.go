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

package opssight

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sizeclientset "github.com/blackducksoftware/synopsys-operator/pkg/size/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// OpsSight is used to handle OpsSight in the cluster
type OpsSight struct {
	config          *protoform.Config
	kubeConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	opsSightClient  *opssightclientset.Clientset
	blackDuckClient *blackduckclientset.Clientset
	sizeClient      *sizeclientset.Clientset
	routeClient     *routeclient.RouteV1Client
	securityClient  *securityclient.SecurityV1Client
}

// NewOpsSight will return an OpsSight type
func NewOpsSight(protoformDeployer *protoform.Deployer) *OpsSight {
	// Initialize the OpsSight client
	opsSightClient, err := opssightclientset.NewForConfig(protoformDeployer.KubeConfig)
	if err != nil {
		return nil
	}

	// Initialize Black Duck Client
	blackDuckClient, err := blackduckclientset.NewForConfig(protoformDeployer.KubeConfig)
	if err != nil {
		return nil
	}

	// Initialize Size Client
	sizeClient, err := sizeclientset.NewForConfig(protoformDeployer.KubeConfig)
	if err != nil {
		return nil
	}

	return &OpsSight{
		config:          protoformDeployer.Config,
		kubeConfig:      protoformDeployer.KubeConfig,
		kubeClient:      protoformDeployer.KubeClient,
		opsSightClient:  opsSightClient,
		blackDuckClient: blackDuckClient,
		sizeClient:      sizeClient,
		routeClient:     protoformDeployer.RouteClient,
		securityClient:  protoformDeployer.SecurityClient,
	}
}

func (o *OpsSight) ensureVersion(opsSight *opssightapi.OpsSight) error {
	versions := o.Versions()
	// If the version is not provided, then we set it to be the latest
	if len(opsSight.Spec.Version) == 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		opsSight.Spec.Version = versions[0]
	} else {
		// If the verion is provided, check that it's supported
		for _, v := range versions {
			if strings.Compare(v, opsSight.Spec.Version) == 0 {
				return nil
			}
		}
		return fmt.Errorf("version '%s' is not supported.  Supported versions: %s", opsSight.Spec.Version, strings.Join(versions, ", "))
	}
	return nil
}

// Versions returns the versions that the operator supports
func (o *OpsSight) Versions() []string {
	var versions []string
	for v := range publicVersions {
		versions = append(versions, v)
	}
	return versions
}

// Ensure will get the necessary Creater and make sure the instance
// is correctly deployed or deploy it if needed
func (o *OpsSight) Ensure(opsSight *opssightapi.OpsSight) error {
	// If the version is not specified then we set it to be the latest.
	if err := o.ensureVersion(opsSight); err != nil {
		return err
	}

	if strings.EqualFold(opsSight.Spec.DesiredState, "STOP") {
		err := o.stop(opsSight)
		if err != nil {
			return err
		}
	} else {
		version, ok := publicVersions[opsSight.Spec.Version]
		if !ok {
			return fmt.Errorf("version %s is not supported", opsSight.Spec.Version)
		}

		// get the registry auth credentials for default OpenShift internal docker registries
		if !o.config.DryRun {
			o.addRegistryAuth(&opsSight.Spec)
		}

		components, err := store.GetComponents(version, o.config, o.kubeClient, o.sizeClient, opsSight)
		if err != nil {
			return err
		}

		if o.config.DryRun {
			// add secret data
			for _, secret := range components.Secrets {
				if utils.GetResourceName(opsSight.Name, util.OpsSightName, opsSight.Spec.SecretName) == secret.GetName() {
					err = o.addSecretData(opsSight, secret)
					if err != nil {
						return err
					}
				}
			}

			// call the CRUD updater to create or update opssight
			commonConfig := crdupdater.NewCRUDComponents(o.kubeConfig, o.kubeClient, o.config.DryRun, false, opsSight.Spec.Namespace, "2.2.4",
				components, fmt.Sprintf("app=%s,name=%s", util.OpsSightName, opsSight.Name), true)
			_, errs := commonConfig.CRUDComponents()
			if len(errs) > 0 {
				return fmt.Errorf("update components errors: %+v", errs)
			}

			err = o.postDeploy(opsSight)
			if err != nil {
				return errors.Annotatef(err, "post deploy")
			}

			err = o.deployBlackDuck(&opsSight.Spec)
			if err != nil {
				return errors.Annotatef(err, "deploy Black Duck")
			}
		}
	}

	return nil
}

// Stop will stop the Black Duck OpsSight
func (o *OpsSight) stop(ops *opssightapi.OpsSight) error {
	log.Debugf("stop OpsSight %s", ops.Spec.Namespace)
	rcs, err := util.ListReplicationControllers(o.kubeClient, ops.Spec.Namespace, "app=opssight")
	for _, rc := range rcs.Items {
		if util.Int32ToInt(rc.Spec.Replicas) > 0 {
			_, err := util.PatchReplicationControllerForReplicas(o.kubeClient, &rc, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s replication controller with replicas %d in %s namespace because %+v", rc.Name, 0, ops.Spec.Namespace, err)
			}
		}
	}

	deployments, err := util.ListDeployments(o.kubeClient, ops.Spec.Namespace, "app=opssight")
	for _, deployment := range deployments.Items {
		if util.Int32ToInt(deployment.Spec.Replicas) > 0 {
			_, err := util.PatchDeploymentForReplicas(o.kubeClient, &deployment, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s deployment with replicas %d in %s namespace because %+v", deployment.Name, 0, ops.Spec.Namespace, err)
			}
		}
	}
	return err
}

// Delete will delete the Black Duck OpsSight
func (o *OpsSight) Delete(name string) error {
	log.Infof("deleting a %s OpsSight instance", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("invalid name to delete the OpsSight instance")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
		ns, err := util.ListNamespaces(o.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.OpsSightName, name))
		if err != nil {
			log.Errorf("unable to list %s OpsSight instance namespaces %s due to %+v", name, namespace, err)
		}
		if len(ns.Items) > 0 {
			namespace = ns.Items[0].Name
		} else {
			return fmt.Errorf("unable to find %s OpsSight instance namespace", name)
		}
	} else {
		name = values[1]
		namespace = values[0]
	}

	// delete the Black Duck instance
	commonConfig := crdupdater.NewCRUDComponents(o.kubeConfig, o.kubeClient, o.config.DryRun, false, namespace, "",
		&api.ComponentList{}, fmt.Sprintf("app=%s,name=%s", util.OpsSightName, name), true)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		return fmt.Errorf("unable to delete the %s OpsSight instance in %s namespace due to %+v", name, namespace, crudErrors)
	}

	// delete namespace and if other apps are running, remove the Synopsys app label from the namespace
	var delErr error
	// if cluster scope, if no other instance running in Synopsys Operator namespace, delete the namespace or delete the Synopsys labels in the namespace
	if o.config.IsClusterScoped {
		delErr = util.DeleteResourceNamespace(o.kubeClient, util.OpsSightName, namespace, name, false)
	} else {
		// if namespace scope, delete the label from the namespace
		_, delErr = util.CheckAndUpdateNamespace(o.kubeClient, util.OpsSightName, namespace, name, "", true)
	}
	if delErr != nil {
		return delErr
	}

	return nil
}

// GetComponents gets the necessary creater and returns the OpsSight's components
func (o *OpsSight) GetComponents(opsSight *opssightapi.OpsSight) (*api.ComponentList, error) {
	//If the version is not specified then we set it to be the latest.
	if err := o.ensureVersion(opsSight); err != nil {
		return nil, err
	}

	version, ok := publicVersions[opsSight.Spec.Version]
	if !ok {
		return nil, fmt.Errorf("version %s is not supported", opsSight.Spec.Version)
	}

	cp, err := store.GetComponents(version, o.config, o.kubeClient, o.sizeClient, opsSight)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

func (o *OpsSight) addRegistryAuth(opsSightSpec *opssightapi.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if o.routeClient == nil {
		return
	}

	internalRegistries := []*string{}

	// Adding default image registry routes
	routes := map[string]string{"default": "docker-registry", "openshift-image-registry": "image-registry"}
	for namespace, name := range routes {
		route, err := util.GetRoute(o.routeClient, namespace, name)
		if err != nil {
			continue
		}
		internalRegistries = append(internalRegistries, &route.Spec.Host)
		routeHostPort := fmt.Sprintf("%s:443", route.Spec.Host)
		internalRegistries = append(internalRegistries, &routeHostPort)
	}

	// Adding default OpenShift internal Docker/image registry service
	labelSelectors := []string{"docker-registry=default", "router in (router,router-default)"}
	for _, labelSelector := range labelSelectors {
		registrySvcs, err := util.ListServices(o.kubeClient, "", labelSelector)
		if err != nil {
			continue
		}
		for _, registrySvc := range registrySvcs.Items {
			if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
				for _, port := range registrySvc.Spec.Ports {
					clusterIPSvc := fmt.Sprintf("%s:%d", registrySvc.Spec.ClusterIP, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvc)
					clusterIPSvcPort := fmt.Sprintf("%s.%s.svc:%d", registrySvc.Name, registrySvc.Namespace, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvcPort)
				}
			}
		}
	}

	file, err := util.ReadFromFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Errorf("unable to read the service account token file due to %+v", err)
	} else {
		for _, internalRegistry := range internalRegistries {
			opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightSpec.ScannerPod.ImageFacade.InternalRegistries, &opssightapi.RegistryAuth{URL: *internalRegistry, User: "admin", Password: string(file)})
		}
	}
}

func (o *OpsSight) addSecretData(opsSight *opssightapi.OpsSight, secret *components.Secret) error {
	blackduckHosts := make(map[string]*opssightapi.Host)
	// adding External Black Duck credentials
	for _, host := range opsSight.Spec.Blackduck.ExternalHosts {
		blackduckHosts[host.Domain] = host
	}

	// adding Internal Black Duck credentials
	secretEditor := NewUpdater(o.config, o.kubeClient, o.blackDuckClient, o.opsSightClient)
	hubType := opsSight.Spec.Blackduck.BlackduckSpec.Type
	blackduckPassword, err := util.Base64Decode(opsSight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return errors.Annotatef(err, "unable to decode blackduckPassword")
	}

	allBlackDucks := secretEditor.GetAllBlackDucks(hubType, blackduckPassword)
	blackduckPasswords := secretEditor.AppendBlackDuckSecrets(blackduckHosts, opsSight.Status.InternalHosts, allBlackDucks)

	// marshal the blackduck credentials to bytes
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal blackduck passwords")
	}
	secret.AddData(map[string][]byte{opsSight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credentials
	securedRegistries := make(map[string]*opssightapi.RegistryAuth)
	for _, internalRegistry := range opsSight.Spec.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = internalRegistry
	}
	// marshal the Secured registries credentials to bytes
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal secured registries")
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	// add internal hosts to status
	opsSight.Status.InternalHosts = secretEditor.AppendBlackDuckHosts(opsSight.Status.InternalHosts, allBlackDucks)
	return nil
}

func (o *OpsSight) postDeploy(opsSight *opssightapi.OpsSight) error {
	// Need to add the perceptor-scanner service account to the privileged scc
	if o.securityClient != nil {
		processorServiceAccountName := utils.GetResourceName(opsSight.Name, util.OpsSightName, opsSight.Spec.Perceiver.ServiceAccount)
		serviceAccounts := []string{fmt.Sprintf("system:serviceaccount:%s:%s", opsSight.Spec.Namespace, processorServiceAccountName)}
		if !strings.EqualFold(opsSight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
			scannerServiceAccountName := utils.GetResourceName(opsSight.Name, util.OpsSightName, opsSight.Spec.ScannerPod.ImageFacade.ServiceAccount)
			serviceAccounts = append(serviceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", opsSight.Spec.Namespace, scannerServiceAccountName))
		}
		return util.UpdateOpenShiftSecurityConstraint(o.securityClient, serviceAccounts, "privileged")
	}
	return nil
}

func (o *OpsSight) deployBlackDuck(opsSight *opssightapi.OpsSightSpec) error {
	if opsSight.Blackduck.InitialCount > opsSight.Blackduck.MaxCount {
		opsSight.Blackduck.InitialCount = opsSight.Blackduck.MaxCount
	}

	blackDuckErrs := map[string]error{}
	for i := 0; i < opsSight.Blackduck.InitialCount; i++ {
		name := fmt.Sprintf("%s-%v", opsSight.Namespace, i)

		_, err := util.GetNamespace(o.kubeClient, name)
		if err == nil {
			continue
		}

		ns, err := util.CreateNamespace(o.kubeClient, name)
		log.Debugf("created namespace: %+v", ns)
		if err != nil {
			log.Errorf("Black Duck[%d]: unable to create the namespace due to %+v", i, err)
			blackDuckErrs[name] = fmt.Errorf("unable to create the namespace due to %+v", err)
		}

		blackDuckSpec := opsSight.Blackduck.BlackduckSpec
		blackDuckSpec.Namespace = name
		createBlackDuck := &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *blackDuckSpec}
		log.Debugf("Black Duck[%d]: %+v", i, createBlackDuck)
		_, err = util.CreateBlackDuck(o.blackDuckClient, name, createBlackDuck)
		if err != nil {
			log.Errorf("Black Duck[%d]: unable to create the Black Duck due to %+v", i, err)
			blackDuckErrs[name] = fmt.Errorf("unable to create the Black Duck due to %+v", err)
		}
	}

	return util.NewMapErrors(blackDuckErrs)
}
