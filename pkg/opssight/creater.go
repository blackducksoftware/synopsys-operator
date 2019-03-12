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

package opssight

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create OpsSight
type Creater struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	opssightClient   *opssightclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	hubClient        *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client, hubClient *hubclientset.Clientset) *Creater {
	return &Creater{
		config:           config,
		kubeConfig:       kubeConfig,
		kubeClient:       kubeClient,
		opssightClient:   opssightClient,
		osSecurityClient: osSecurityClient,
		routeClient:      routeClient,
		hubClient:        hubClient,
	}
}

// DeleteOpsSight will delete the Black Duck OpsSight
func (ac *Creater) DeleteOpsSight(namespace string) error {
	log.Debugf("delete OpsSight details for %s", namespace)
	// Verify that the namespace exists
	_, err := util.GetNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to find namespace %s", namespace)
	}

	// get all replication controller for the namespace
	rcs, err := util.GetReplicationControllerList(ac.kubeClient, namespace, "")
	if err != nil {
		return errors.Annotatef(err, "unable to list the replication controller in %s", namespace)
	}

	// get only opssight related replication controller for the namespace
	opssightRCs, err := util.GetReplicationControllerList(ac.kubeClient, namespace, "app=opssight")
	if err != nil {
		return errors.Annotatef(err, "unable to list the opssight's replication controller in %s", namespace)
	}

	// if both the length same, then delete the namespace, if different, delete only the replication controller
	if len(rcs.Items) == len(opssightRCs.Items) {
		// Delete the namespace
		err = util.DeleteNamespace(ac.kubeClient, namespace)
		if err != nil {
			return errors.Annotatef(err, "unable to delete namespace %s", namespace)
		}

		for {
			// Verify whether the namespace was deleted
			ns, err := util.GetNamespace(ac.kubeClient, namespace)
			log.Infof("namespace: %v, status: %v", namespace, ns.Status)
			if err != nil {
				log.Infof("deleted the namespace %+v", namespace)
				break
			}
			time.Sleep(10 * time.Second)
		}
	} else {
		// delete the replication controller
		for _, opssightRC := range opssightRCs.Items {
			err = util.DeleteReplicationController(ac.kubeClient, namespace, opssightRC.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s replication controller in %s namespace", opssightRC.GetName(), namespace)
			}
		}
	}

	clusterRoleBindings, err := util.ListClusterRoleBinding(ac.kubeClient, "app=opssight")

	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		if len(clusterRoleBinding.Subjects) == 1 {
			if !strings.EqualFold(clusterRoleBinding.RoleRef.Name, "cluster-admin") {
				log.Debugf("deleting cluster role %s", clusterRoleBinding.RoleRef.Name)
				err = util.DeleteClusterRole(ac.kubeClient, clusterRoleBinding.RoleRef.Name)
				if err != nil {
					log.Errorf("unable to delete the cluster role for %+v", clusterRoleBinding.RoleRef.Name)
				}
			}

			log.Debugf("deleting cluster role binding %s", clusterRoleBinding.GetName())
			err = util.DeleteClusterRoleBinding(ac.kubeClient, clusterRoleBinding.GetName())
			if err != nil {
				log.Errorf("unable to delete the cluster role binding for %+v", clusterRoleBinding.GetName())
			}
		} else {
			log.Debugf("updating cluster role binding %s", clusterRoleBinding.GetName())
			clusterRoleBinding.Subjects = removeSubjects(clusterRoleBinding.Subjects, namespace)
			_, err = util.UpdateClusterRoleBinding(ac.kubeClient, &clusterRoleBinding)
			if err != nil {
				log.Errorf("unable to update the cluster role binding for %+v", clusterRoleBinding.GetName())
			}
		}
	}

	return nil
}

func removeSubjects(subjects []rbacv1.Subject, namespace string) []rbacv1.Subject {
	newSubjects := []rbacv1.Subject{}
	for _, subject := range subjects {
		if !strings.EqualFold(subject.Namespace, namespace) {
			newSubjects = append(newSubjects, subject)
		}
	}
	return newSubjects
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(opsSight *opssightapi.OpsSightSpec) error {
	log.Debugf("create OpsSight details for %s: %+v", opsSight.Namespace, opsSight)

	// get the registry auth credentials for default OpenShift internal docker registries
	if !ac.config.DryRun {
		ac.addRegistryAuth(opsSight)
	}

	opssight := NewSpecConfig(ac.kubeClient, opsSight, ac.config.DryRun)

	components, err := opssight.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opsSight.Namespace)
	}

	// setting up blackduck password in perceptor secret
	if !ac.config.DryRun {
		for _, secret := range components.Secrets {
			if strings.EqualFold(secret.GetName(), opsSight.SecretName) {
				ac.addSecretData(opsSight, secret)
				break
			}
		}
	}

	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", opsSight.Namespace)
	}
	// Note: controllers that need to continually run to update your app
	// should be added in PreDeploy().
	deployer.PreDeploy(components, opsSight.Namespace)

	if !ac.config.DryRun {
		err = deployer.Run()
		if err != nil {
			log.Errorf("unable to deploy opssight %s due to %+v", opsSight.Namespace, err)
		}
		deployer.StartControllers()
		// if OpenShift, add a privileged role to scanner account
		err = ac.postDeploy(opssight, opsSight.Namespace)
		if err != nil {
			return errors.Trace(err)
		}

		err = ac.deployHub(opsSight)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// StopOpsSight will stop the Black Duck OpsSight
func (ac *Creater) StopOpsSight(opssight *opssightapi.OpsSightSpec) error {
	rcl, err := util.GetReplicationControllerList(ac.kubeClient, opssight.Namespace, "app=opssight")
	for _, rc := range rcl.Items {
		if util.Int32ToInt(rc.Spec.Replicas) > 0 {
			err := util.PatchReplicationControllerForReplicas(ac.kubeClient, rc, 0)
			if err != nil {
				return fmt.Errorf("unable to patch %s replication controller with replicas %d in %s namespace because %+v", rc.Name, 0, opssight.Namespace, err)
			}
		}
	}
	return err
}

// ReplicationControllerComparator used to compare Replication controller attributes
type ReplicationControllerComparator struct {
	Image    string
	Replicas *int32
	MinCPU   *resource.Quantity
	MaxCPU   *resource.Quantity
	MinMem   *resource.Quantity
	MaxMem   *resource.Quantity
}

// UpdateOpsSight will update the Black Duck OpsSight
func (ac *Creater) UpdateOpsSight(opssight *opssightapi.OpsSightSpec) error {
	newConfigMapConfig := NewSpecConfig(ac.kubeClient, opssight, ac.config.DryRun)
	// check whether the configmap is changed, if so update the configmap
	isConfigMapUpdated, err := ac.updateConfigMap(opssight, newConfigMapConfig)
	if err != nil {
		return errors.Annotate(err, "update configmap:")
	}

	// check whether the secret is changed, if so update the secret
	isSecretUpdated, err := ac.updateSecret(opssight, newConfigMapConfig)
	if err != nil {
		return errors.Annotate(err, "update secret:")
	}

	// check whether any replication controller, configmap, secret, services, cluster role or cluster role binding is updated, if so create/patch the replication controller
	err = ac.update(opssight, newConfigMapConfig, isConfigMapUpdated, isSecretUpdated)
	if err != nil {
		return errors.Annotate(err, "update replication controller:")
	}

	return nil
}

func (ac *Creater) updateConfigMap(opssight *opssightapi.OpsSightSpec, newConfigMapConfig *SpecConfig) (bool, error) {
	configMapName := fmt.Sprintf("%s.json", opssight.ConfigMapName)
	// build new configmap data for comparing
	newConfig, err := newConfigMapConfig.configMap.horizonConfigMap(opssight.ConfigMapName, opssight.Namespace, configMapName)
	if err != nil {
		return false, errors.Annotatef(err, "unable to create horizon configmap %s in opssight namespace %s", opssight.ConfigMapName, opssight.Namespace)
	}
	newConfigMapKube, err := newConfig.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert configmap %s to kube in opssight namespace %s", opssight.ConfigMapName, opssight.Namespace)
	}
	newConfigMap := newConfigMapKube.(*corev1.ConfigMap)
	newConfigMapData := newConfigMap.Data

	// getting old configmap data
	oldConfigMap, err := util.GetConfigMap(ac.kubeClient, opssight.Namespace, opssight.ConfigMapName)
	if err != nil {
		return false, errors.Annotatef(err, "unable to find the configmap %s in opssight namespace %s", opssight.ConfigMapName, opssight.Namespace)
	}
	oldConfigMapData := oldConfigMap.Data

	// compare for difference between old and new configmap data, if changed update the configmap
	if !reflect.DeepEqual(newConfigMapData, oldConfigMapData) {
		oldConfigMap.Data = newConfigMapData
		err = util.UpdateConfigMap(ac.kubeClient, opssight.Namespace, oldConfigMap)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the configmap %s in namespace %s", opssight.ConfigMapName, opssight.Namespace)
		}
		return true, nil
	}
	return false, nil
}

func (ac *Creater) updateSecret(opssight *opssightapi.OpsSightSpec, newConfigMapConfig *SpecConfig) (bool, error) {
	secretName := opssight.SecretName
	// build new secret data for comparing
	secret := newConfigMapConfig.PerceptorSecret()
	err := ac.addSecretData(opssight, secret)
	if err != nil {
		return false, errors.Annotate(err, fmt.Sprintf("unable to add secret data to %s secret in %s namespace", secretName, opssight.Namespace))
	}
	newSecretKube, err := secret.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert secret %s to kube in opssight namespace %s", secretName, opssight.Namespace)
	}
	newSecret := newSecretKube.(*corev1.Secret)
	newSecretData := newSecret.Data

	// getting old secret data
	oldSecret, err := util.GetSecret(ac.kubeClient, opssight.Namespace, secretName)
	if err != nil {
		return false, errors.Annotatef(err, "unable to find the secret %s in namespace %s", secretName, opssight.Namespace)
	}
	oldSecretData := oldSecret.Data

	// compare for difference between old and new secret data, if changed update the secret
	if !reflect.DeepEqual(newSecretData, oldSecretData) {
		oldSecret.Data = newSecretData
		err = util.UpdateSecret(ac.kubeClient, opssight.Namespace, oldSecret)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the secret %s in namespace %s", secretName, opssight.Namespace)
		}
		return true, nil
	}
	return false, nil
}

func (ac *Creater) update(opssight *opssightapi.OpsSightSpec, newConfigMapConfig *SpecConfig, isConfigMapUpdated bool, isSecretUpdated bool) error {
	// get new components build from the latest updates
	components, err := newConfigMapConfig.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opssight.Namespace)
	}

	// add the missing services
	err = ac.addServices(opssight.Namespace, components.Services)
	if err != nil {
		return errors.Annotatef(err, "unable to add opssight services for %s", opssight.Namespace)
	}

	// add the missing cluster roles
	err = ac.addClusterRoles(opssight.Namespace, components.ClusterRoles)
	if err != nil {
		return errors.Annotatef(err, "unable to add opssight cluster roles for %s", opssight.Namespace)
	}

	// add the missing cluster role bindings
	err = ac.addClusterRoleBindings(opssight.Namespace, components.ClusterRoleBindings)
	if err != nil {
		return errors.Annotatef(err, "unable to add opssight cluster role bindings for %s", opssight.Namespace)
	}

	// update the replication controller
	err = ac.updateReplicationController(opssight, components.ReplicationControllers, isConfigMapUpdated, isSecretUpdated)
	if err != nil {
		return errors.Annotatef(err, "opssight replication controller")
	}

	return nil
}

func (ac *Creater) updateReplicationController(opssight *opssightapi.OpsSightSpec, replicationControllers []*components.ReplicationController, isConfigMapUpdated bool, isSecretUpdated bool) error {
	// get old replication controller
	rcl, err := util.GetReplicationControllerList(ac.kubeClient, opssight.Namespace, "app=opssight")
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight replication controllers for %s", opssight.Namespace)
	}

	oldRCs := make(map[string]corev1.ReplicationController)
	for _, rc := range rcl.Items {
		oldRCs[rc.GetName()] = rc
	}

	// iterate through the replication controller list for any changes
	for _, component := range replicationControllers {
		newRCKube, err := component.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert rc %s to kube in opssight namespace %s", component.GetName(), opssight.Namespace)
		}

		newRC := newRCKube.(*corev1.ReplicationController)
		oldRC := oldRCs[newRC.GetName()]

		// if the replication controller is not found in the cluster, create it
		if _, ok := oldRCs[newRC.GetName()]; !ok {
			deployer, err := util.NewDeployer(ac.kubeConfig)
			if err != nil {
				return errors.Annotatef(err, "unable to get deployer object for %s", opssight.Namespace)
			}
			deployer.Deployer.AddReplicationController(component)
			deployer.Deployer.Run()
		}

		// if config map or secret is updated, patch the replication controller
		if isConfigMapUpdated || isSecretUpdated {
			err = util.PatchReplicationController(ac.kubeClient, oldRC, *newRC)
			if err != nil {
				return errors.Annotatef(err, "unable to patch rc %s to kube in opssight namespace %s", component.GetName(), opssight.Namespace)
			}
			continue
		}

		// check whether the replication controller or its container got changed
		isChanged := false
		for _, oldContainer := range oldRC.Spec.Template.Spec.Containers {
			for _, newContainer := range newRC.Spec.Template.Spec.Containers {
				if strings.EqualFold(oldContainer.Name, newContainer.Name) &&
					!reflect.DeepEqual(
						ReplicationControllerComparator{
							Image:    oldContainer.Image,
							Replicas: oldRC.Spec.Replicas,
							MinCPU:   oldContainer.Resources.Requests.Cpu(),
							MaxCPU:   oldContainer.Resources.Limits.Cpu(),
							MinMem:   oldContainer.Resources.Requests.Memory(),
							MaxMem:   oldContainer.Resources.Limits.Memory(),
						},
						ReplicationControllerComparator{
							Image:    newContainer.Image,
							Replicas: newRC.Spec.Replicas,
							MinCPU:   newContainer.Resources.Requests.Cpu(),
							MaxCPU:   newContainer.Resources.Limits.Cpu(),
							MinMem:   newContainer.Resources.Requests.Memory(),
							MaxMem:   newContainer.Resources.Limits.Memory(),
						}) {
					isChanged = true
				}
			}
		}

		// if changed from the above step, patch the replication controller
		if isChanged {
			err = util.PatchReplicationController(ac.kubeClient, oldRC, *newRC)
			if err != nil {
				return errors.Annotatef(err, "unable to patch rc %s to kube in opssight namespace %s", component.GetName(), opssight.Namespace)
			}
		}
	}
	return nil
}

func (ac *Creater) addServices(namespace string, services []*components.Service) error {
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	isRun := false
	for _, service := range services {
		_, err := util.GetService(ac.kubeClient, namespace, service.GetName())
		if err != nil {
			deployer.Deployer.AddService(service)
			isRun = true
		}
	}

	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy service object due to %+v", err)
		}
	}
	return nil
}

func (ac *Creater) addClusterRoles(namespace string, clusterRoles []*components.ClusterRole) error {
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	isRun := false
	for _, clusterRole := range clusterRoles {
		_, err := util.GetClusterRole(ac.kubeClient, clusterRole.GetName())
		if err != nil {
			deployer.Deployer.AddClusterRole(clusterRole)
			isRun = true
		}
	}

	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy cluster role object due to %+v", err)
		}
	}
	return nil
}

func (ac *Creater) addClusterRoleBindings(namespace string, clusterRoleBindings []*components.ClusterRoleBinding) error {
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", namespace)
	}
	isRun := false
	for _, clusterRoleBinding := range clusterRoleBindings {
		_, err := util.GetClusterRoleBinding(ac.kubeClient, clusterRoleBinding.GetName())
		if err != nil {
			deployer.Deployer.AddClusterRoleBinding(clusterRoleBinding)
			isRun = true
		}
	}

	if isRun {
		err = deployer.Deployer.Run()
		if err != nil {
			log.Debugf("unable to deploy cluster role binding object due to %+v", err)
		}
	}
	return nil
}

func (ac *Creater) addSecretData(opsSight *opssightapi.OpsSightSpec, secret *components.Secret) error {
	blackduckPasswords := make(map[string]interface{})
	// adding External Black Duck passwords
	for _, host := range opsSight.Blackduck.ExternalHosts {
		blackduckPasswords[host.Domain] = &host
	}
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return errors.Trace(err)
	}
	secret.AddData(map[string][]byte{opsSight.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credential
	securedRegistries := make(map[string]interface{})
	for _, internalRegistry := range opsSight.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = &internalRegistry
	}
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return errors.Trace(err)
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})
	return nil
}

// GetDefaultPasswords returns admin,user,postgres passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (hubPassword string, err error) {
	blackduckSecret, err := util.GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		return "", errors.Annotate(err, "You need to first create a 'blackduck-secret' in this namespace with HUB_PASSWORD")
	}
	hubPassword = string(blackduckSecret.Data["HUB_PASSWORD"])

	// default named return
	return hubPassword, nil
}

func (ac *Creater) addRegistryAuth(opsSightSpec *opssightapi.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if ac.routeClient == nil {
		return
	}

	internalRegistries := []string{}
	route, err := util.GetOpenShiftRoutes(ac.routeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry router in default namespace due to %+v", err)
	} else {
		internalRegistries = append(internalRegistries, route.Spec.Host)
		internalRegistries = append(internalRegistries, fmt.Sprintf("%s:443", route.Spec.Host))
	}

	registrySvc, err := util.GetService(ac.kubeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry service in default namespace due to %+v", err)
	} else {
		if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
			for _, port := range registrySvc.Spec.Ports {
				internalRegistries = append(internalRegistries, fmt.Sprintf("%s:%s", registrySvc.Spec.ClusterIP, strconv.Itoa(int(port.Port))))
				internalRegistries = append(internalRegistries, fmt.Sprintf("%s:%s", "docker-registry.default.svc", strconv.Itoa(int(port.Port))))
			}
		}
	}

	file, err := util.ReadFromFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Errorf("unable to read the service account token file due to %+v", err)
	} else {
		for _, internalRegistry := range internalRegistries {
			registryAuth := &opssightapi.RegistryAuth{URL: internalRegistry, User: "admin", Password: string(file)}
			opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightSpec.ScannerPod.ImageFacade.InternalRegistries, registryAuth)
		}
	}
}

func (ac *Creater) postDeploy(opssight *SpecConfig, namespace string) error {
	// Need to add the perceptor-scanner service account to the privileged scc
	if ac.osSecurityClient != nil {
		scannerServiceAccount := opssight.ScannerServiceAccount()
		perceiverServiceAccount := opssight.PodPerceiverServiceAccount()
		serviceAccounts := []string{fmt.Sprintf("system:serviceaccount:%s:%s", namespace, perceiverServiceAccount.GetName())}
		if !strings.EqualFold(opssight.config.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
			serviceAccounts = append(serviceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, scannerServiceAccount.GetName()))
		}
		return util.UpdateOpenShiftSecurityConstraint(ac.osSecurityClient, serviceAccounts, "privileged")
	}
	return nil
}

func (ac *Creater) deployHub(createOpsSight *opssightapi.OpsSightSpec) error {
	if createOpsSight.Blackduck.InitialCount > createOpsSight.Blackduck.MaxCount {
		createOpsSight.Blackduck.InitialCount = createOpsSight.Blackduck.MaxCount
	}

	hubErrs := map[string]error{}
	for i := 0; i < createOpsSight.Blackduck.InitialCount; i++ {
		name := fmt.Sprintf("%s-%v", createOpsSight.Namespace, i)

		ns, err := util.CreateNamespace(ac.kubeClient, name)
		log.Debugf("created namespace: %+v", ns)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the namespace due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the namespace due to %+v", err)
		}

		hubSpec := createOpsSight.Blackduck.BlackduckSpec
		hubSpec.Namespace = name
		createHub := &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *hubSpec}
		log.Debugf("hub[%d]: %+v", i, createHub)
		_, err = util.CreateHub(ac.hubClient, name, createHub)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the hub due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the hub due to %+v", err)
		}
	}

	return util.NewMapErrors(hubErrs)
}
