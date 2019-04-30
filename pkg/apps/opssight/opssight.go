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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	latestopssight "github.com/blackducksoftware/synopsys-operator/pkg/apps/opssight/latest"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
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

// OpsSight is used to handle OpsSight in the cluster
type OpsSight struct {
	config         *protoform.Config
	kubeConfig     *rest.Config
	kubeClient     *kubernetes.Clientset
	opsSightClient *opssightclientset.Clientset
	routeClient    *routeclient.RouteV1Client
	creaters       []Creater
}

// NewOpsSight will return an OpsSight type
func NewOpsSight(config *protoform.Config, kubeConfig *rest.Config) *OpsSight {
	// Initialiase the clienset
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize the OpsSight client
	opsSightClient, err := opssightclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize Security Client
	osSecurityClient, err := securityclient.NewForConfig(kubeConfig)
	if err != nil {
		osSecurityClient = nil
	} else {
		_, err := util.GetOpenShiftSecurityConstraint(osSecurityClient, "anyuid")
		if err != nil {
			osSecurityClient = nil
		}
	}
	// Initialize the Route Client for Openshift routes
	routeClient := util.GetRouteClient(kubeConfig)

	// Initialize Black Duck Client
	blackduckClient, err := blackduckclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil
	}
	// Initialize creaters for different versions of OpsSight (each Creater can support differernt versions)
	creaters := []Creater{
		latestopssight.NewCreater(config, kubeConfig, kubeClient, opsSightClient, osSecurityClient, routeClient, blackduckClient),
	}

	return &OpsSight{
		config:         config,
		kubeConfig:     kubeConfig,
		kubeClient:     kubeClient,
		opsSightClient: opsSightClient,
		routeClient:    routeClient,
		creaters:       creaters,
	}
}

// getCreater loops through each Creater and returns the one
// that supports the specified version
func (o OpsSight) getCreater(version string) (Creater, error) {
	for _, c := range o.creaters {
		for _, v := range c.Versions() {
			if strings.Compare(v, version) == 0 {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("version %s is not supported", version)
}

// Versions returns the versions that the operator supports for OpsSight
func (o OpsSight) Versions() []string {
	var versions []string
	// Get versions that each Creater supports
	for _, c := range o.creaters {
		for _, v := range c.Versions() {
			versions = append(versions, v)
		}
	}
	return versions
}

// Ensure will get the necessary Creater and make sure the instance
// is correctly deployed or deploy it if needed
func (o OpsSight) Ensure(ops *opssightapi.OpsSight) error {
	// If the version is not specified then we set it to be the latest.
	log.Debugf("ensure OpsSight %s", ops.Spec.Namespace)
	if len(ops.Spec.Version) == 0 {
		versions := o.Versions()
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		ops.Spec.Version = versions[0]
		log.Debugf("setting OpsSight version to %s", ops.Spec.Version)
	}

	creater, err := o.getCreater(ops.Spec.Version) // get Creater for the OpsSight Version
	if err != nil {
		return fmt.Errorf("failed to get OpsSight creater for version %s: %s", ops.Spec.Version, err)
	}
	return creater.Ensure(ops) // Ensure the OpsSight
}

// Delete will delete the Black Duck OpsSight
func (o OpsSight) Delete(namespace string) error {
	log.Debugf("delete OpsSight %s", namespace)
	// Verify that the namespace exists
	_, err := util.GetNamespace(o.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to find namespace %s", namespace)
	}

	// get all replication controller for the namespace
	rcs, err := util.ListReplicationControllers(o.kubeClient, namespace, "")
	if err != nil {
		return errors.Annotatef(err, "unable to list the replication controller in %s", namespace)
	}

	// get only opssight related replication controllers for the namespace
	opssightRCs, err := util.ListReplicationControllers(o.kubeClient, namespace, "app=opssight")
	if err != nil {
		return errors.Annotatef(err, "unable to list the opssight's replication controllers in %s", namespace)
	}

	// get only opssight related deployments for the namespace (for Prometheus)
	opssightDeployments, err := util.ListDeployments(o.kubeClient, namespace, "app=opssight")
	if err != nil {
		return errors.Annotatef(err, "unable to list the opssight's deployments in %s", namespace)
	}

	// if both the length same, then delete the namespace, if different, delete only the replication controllers and deployments
	if len(rcs.Items) == len(opssightRCs.Items) {
		// Delete the namespace
		err = util.DeleteNamespace(o.kubeClient, namespace)
		if err != nil {
			return errors.Annotatef(err, "unable to delete namespace %s", namespace)
		}

		for {
			// Verify whether the namespace was deleted
			ns, err := util.GetNamespace(o.kubeClient, namespace)
			log.Infof("namespace: %v, status: %v", namespace, ns.Status)
			if err != nil {
				log.Infof("deleted the namespace %+v", namespace)
				break
			}
			time.Sleep(10 * time.Second)
		}
	} else {
		// delete the replication controllers
		for _, opssightRC := range opssightRCs.Items {
			err = util.DeleteReplicationController(o.kubeClient, namespace, opssightRC.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s replication controller in %s namespace", opssightRC.GetName(), namespace)
			}
		}
		// delete the deployments
		for _, opssightD := range opssightDeployments.Items {
			err = util.DeleteDeployment(o.kubeClient, namespace, opssightD.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s deployment in %s namespace", opssightD.GetName(), namespace)
			}
		}
	}

	clusterRoleBindings, err := util.ListClusterRoleBindings(o.kubeClient, "app=opssight")

	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		if len(clusterRoleBinding.Subjects) == 1 {
			if !strings.EqualFold(clusterRoleBinding.RoleRef.Name, "synopsys-operator-admin") {
				log.Debugf("deleting cluster role %s", clusterRoleBinding.RoleRef.Name)
				err = util.DeleteClusterRole(o.kubeClient, clusterRoleBinding.RoleRef.Name)
				if err != nil {
					log.Errorf("unable to delete the cluster role for %+v", clusterRoleBinding.RoleRef.Name)
				}
			}

			log.Debugf("deleting cluster role binding %s", clusterRoleBinding.GetName())
			err = util.DeleteClusterRoleBinding(o.kubeClient, clusterRoleBinding.GetName())
			if err != nil {
				log.Errorf("unable to delete the cluster role binding for %+v", clusterRoleBinding.GetName())
			}
		} else {
			log.Debugf("updating cluster role binding %s", clusterRoleBinding.GetName())
			clusterRoleBinding.Subjects = removeSubjects(clusterRoleBinding.Subjects, namespace)
			_, err = util.UpdateClusterRoleBinding(o.kubeClient, &clusterRoleBinding)
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

// Stop will stop the Black Duck OpsSight
func (o OpsSight) Stop(ops *opssightapi.OpsSight) error {
	log.Debugf("stop OpsSight %s", ops.Spec.Namespace)
	rcl, err := util.ListReplicationControllers(o.kubeClient, ops.Spec.Namespace, "app=opssight")
	for _, rc := range rcl.Items {
		if util.Int32ToInt(rc.Spec.Replicas) > 0 {
			_, err := util.PatchReplicationControllerForReplicas(o.kubeClient, &rc, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s replication controller with replicas %d in %s namespace because %+v", rc.Name, 0, ops.Spec.Namespace, err)
			}
		}
	}

	dpl, err := util.ListDeployments(o.kubeClient, ops.Spec.Namespace, "app=opssight")
	for _, dp := range dpl.Items {
		if util.Int32ToInt(dp.Spec.Replicas) > 0 {
			_, err := util.PatchDeploymentForReplicas(o.kubeClient, &dp, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s deployment with replicas %d in %s namespace because %+v", dp.Name, 0, ops.Spec.Namespace, err)
			}
		}
	}
	return err
}

// GetComponents gets the necessary creater and returns the OpsSight's components
func (o *OpsSight) GetComponents(opsSight *opssightapi.OpsSight) (*api.ComponentList, error) {
	creater, err := o.getCreater(opsSight.Spec.Version) // get Creater for the OpsSight Version
	if err != nil {
		return nil, err
	}
	return creater.GetComponents(opsSight)
}
