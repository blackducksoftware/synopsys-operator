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
	"sort"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	latestblackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/latest"
	v1blackduck "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Blackduck is used for the Blackduck deployment
type Blackduck struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	blackduckClient  *blackduckclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	creaters         []Creater
}

// NewBlackduck will return a Blackduck
func NewBlackduck(config *protoform.Config, kubeConfig *rest.Config) *Blackduck {
	// Initialiase the clienset using kubeConfig
	kubeclient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}

	osClient, err := securityclient.NewForConfig(kubeConfig)
	if err != nil {
		osClient = nil
	} else {
		_, err := util.GetOpenShiftSecurityConstraint(osClient, "anyuid")
		if err != nil {
			osClient = nil
		}
	}

	routeClient := util.GetRouteClient(kubeConfig, config.Namespace)

	creaters := []Creater{
		v1blackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, osClient, routeClient),
		latestblackduck.NewCreater(config, kubeConfig, kubeclient, blackduckClient, osClient, routeClient),
	}

	return &Blackduck{
		kubeConfig:       kubeConfig,
		kubeClient:       kubeclient,
		blackduckClient:  blackduckClient,
		osSecurityClient: osClient,
		routeClient:      routeClient,
		creaters:         creaters,
	}
}

func (b *Blackduck) getCreater(version string) (Creater, error) {
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

// Delete will be used to delete a blackduck instance
func (b *Blackduck) Delete(name string) error {
	log.Infof("deleting %s", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("unable to find the Black Duck namespace")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
	} else {
		name = values[1]
		namespace = values[0]
	}

	// Verify whether the namespace exist
	_, err := util.GetNamespace(b.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to find the namespace %+v", namespace)
	}

	// get all replication controller for the namespace
	rcs, err := util.ListReplicationControllers(b.kubeClient, namespace, fmt.Sprintf("app!=blackduck,name!=%s", name))
	if err != nil {
		return errors.Annotatef(err, "unable to list the replication controllers in %s namespace", namespace)
	}

	// get all deployment for the namespace
	deployments, err := util.ListDeployments(b.kubeClient, namespace, fmt.Sprintf("app!=blackduck,name!=%s", name))
	if err != nil {
		return errors.Annotatef(err, "unable to list the deployments in %s namespace", namespace)
	}

	// get only Black Duck related replication controller for the namespace
	blackduckRCs, err := util.ListReplicationControllers(b.kubeClient, namespace, fmt.Sprintf("app=blackduck,name=%s", name))
	if err != nil {
		return errors.Annotatef(err, "unable to list the Black Duck's replication controller in %s namespace", namespace)
	}

	// if both the length same, then delete the namespace, if different, delete only the replication controller
	if (len(rcs.Items) == 0 && len(deployments.Items) == 0) || (len(rcs.Items) == len(blackduckRCs.Items)) {
		// Delete the namespace
		err = util.DeleteNamespace(b.kubeClient, namespace)
		if err != nil {
			return errors.Annotatef(err, "unable to delete %s namespace", namespace)
		}

		// Verify whether the namespace deleted
		var attempts = 30
		var retryWait time.Duration = 10
		for i := 0; i <= attempts; i++ {
			_, err := util.GetNamespace(b.kubeClient, namespace)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
			if i >= 10 {
				return fmt.Errorf("unable to delete the namespace %+v after %f minutes", namespace, float64(attempts)*retryWait.Seconds()/60)
			}
			time.Sleep(retryWait * time.Second)
		}
	} else {
		// delete the replication controller
		for _, blackduckRC := range blackduckRCs.Items {
			err = util.DeleteReplicationController(b.kubeClient, namespace, blackduckRC.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s replication controller in %s namespace", blackduckRC.GetName(), namespace)
			}
		}

		// get only Black Duck related services for the namespace
		services, err := util.ListServices(b.kubeClient, namespace, fmt.Sprintf("app=blackduck,name=%s", name))
		if err != nil {
			return errors.Annotatef(err, "unable to list the Black Duck's service in %s", namespace)
		}

		// delete the service
		for _, service := range services.Items {
			err = util.DeleteService(b.kubeClient, namespace, service.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s service in %s namespace", service.GetName(), namespace)
			}
		}

		// get only Black Duck related pvcs for the namespace
		pvcs, err := util.ListPVCs(b.kubeClient, namespace, fmt.Sprintf("app=blackduck,name=%s", name))
		if err != nil {
			return errors.Annotatef(err, "unable to list the Black Duck's pvc in %s", namespace)
		}

		// delete the pvc
		for _, pvc := range pvcs.Items {
			err = util.DeletePVC(b.kubeClient, namespace, pvc.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s pvc in %s namespace", pvc.GetName(), namespace)
			}
		}

		// get only Black Duck related configmaps for the namespace
		cms, err := util.ListConfigMaps(b.kubeClient, namespace, fmt.Sprintf("app=blackduck,name=%s", name))
		if err != nil {
			return errors.Annotatef(err, "unable to list the Black Duck's config map in %s", namespace)
		}

		// delete the config map
		for _, cm := range cms.Items {
			err = util.DeleteConfigMap(b.kubeClient, namespace, cm.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s config map in %s namespace", cm.GetName(), namespace)
			}
		}

		// get only Black Duck related secrets for the namespace
		secrets, err := util.ListSecrets(b.kubeClient, name, fmt.Sprintf("app=blackduck,name=%s", name))
		if err != nil {
			return errors.Annotatef(err, "unable to list the Black Duck's secret in %s", name)
		}

		// delete the config map
		for _, secret := range secrets.Items {
			err = util.DeleteSecret(b.kubeClient, name, secret.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s secret in %s namespace", secret.GetName(), name)
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

// Versions returns the versions that the operator supports
func (b *Blackduck) Versions() []string {
	var versions []string
	for _, c := range b.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will make sure the instance is correctly deployed or deploy it if needed
func (b *Blackduck) Ensure(bd *v1.Blackduck) error {
	// If the version is not specified then we set it to be the latest.
	if len(bd.Spec.Version) == 0 {
		versions := b.Versions()
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		bd.Spec.Version = versions[0]
	}

	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return err
	}

	return creater.Ensure(bd)
}

// GetComponents gets the BlackDuck's creater and returns the components
func (b Blackduck) GetComponents(bd *blackduckapi.Blackduck) (*api.ComponentList, error) {
	creater, err := b.getCreater(bd.Spec.Version)
	if err != nil {
		return nil, err
	}
	return creater.GetComponents(bd)
}
