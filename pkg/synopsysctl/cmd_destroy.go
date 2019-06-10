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

package synopsysctl

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// destroyCmd removes Synopsys Operator from the cluster
var destroyCmd = &cobra.Command{
	Use:     "destroy [NAMESPACES...]",
	Short:   "Removes one or more Synopsys Operator and its associated CRD's on your cluster",
	Example: "synopsysctl destroy",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		isClusterScoped := util.GetClusterScope(apiExtensionClient)
		// Read Commandline Parameters
		if len(args) > 0 {
			for _, namespace := range args {
				destroyNamespace(namespace, isClusterScoped)
			}
		} else {
			namespace := DefaultDeployNamespace
			var err error
			if isClusterScoped {
				namespace, err = util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
				if err != nil {
					log.Error(err)
				}
			}
			destroyNamespace(namespace, isClusterScoped)
		}
		return nil
	},
}

func destroyNamespace(namespace string, isClusterScoped bool) {
	if metav1.NamespaceAll != namespace {
		log.Infof("destroying the synopsys operator in '%s' namespace...", namespace)
		_, err := util.GetNamespace(kubeClient, namespace)
		if err != nil {
			log.Warnf("unable to find the synopsys operator in '%s' namespace due to %+v", namespace, err)
		} else {
			// delete namespace
			isDeleteOperatorNamespace := util.IsDeleteOperatorNamespace(kubeClient, namespace)
			if isDeleteOperatorNamespace {
				log.Debugf("deleting namespace %s", namespace)
				err = util.DeleteNamespace(kubeClient, namespace)
				if err != nil {
					log.Errorf("unable to delete the %s namespace because %+v", namespace, err)
				}
			} else {
				log.Warnf("synopsys operator in '%s' namespace will not be deleted because other instances are still running in the namespace", namespace)
			}
		}
	}

	// delete crds
	crds := []string{util.AlertCRDName, util.BlackDuckCRDName, util.OpsSightCRDName, util.PrmCRDName}
	for _, crd := range crds {
		log.Infof("deleting %s custom resource definitions", crd)
		err := util.DeleteCustomResourceDefinition(apiExtensionClient, crd)
		if err != nil {
			log.Errorf("unable to delete the %s custom resource definitions because %+v", crd, err)
		}
	}

	// delete cluster roles/ roles
	clusterRoles, roles, err := util.GetOperatorRoles(kubeClient, namespace)
	if err != nil {
		log.Errorf("error getting the role or cluster role due to %+v", err)
	}

	for _, clusterRole := range clusterRoles {
		log.Infof("deleting %s cluster role ", clusterRole)
		err := util.DeleteClusterRole(kubeClient, clusterRole)
		if err != nil {
			log.Errorf("unable to delete the %s cluster role because %+v", clusterRole, err)
		}
	}

	for _, role := range roles {
		log.Infof("deleting %s role ", role)
		err := util.DeleteRole(kubeClient, namespace, role)
		if err != nil {
			log.Errorf("unable to delete the %s role because %+v", role, err)
		}
	}

	// delete cluster role/role bindings
	clusterRoleBindings, roleBindings, err := util.GetOperatorRoleBindings(kubeClient, namespace)
	if err != nil {
		log.Errorf("error getting the role binding or cluster role binding due to %+v", err)
	}

	for _, clusterRoleBinding := range clusterRoleBindings {
		log.Infof("deleting %s cluster role binding", clusterRoleBinding)
		err := util.DeleteClusterRoleBinding(kubeClient, clusterRoleBinding)
		if err != nil {
			log.Errorf("unable to delete the %s cluster role binding because %+v", clusterRoleBinding, err)
		}
	}

	for _, roleBinding := range roleBindings {
		log.Infof("deleting %s role binding", roleBinding)
		err = util.DeleteRoleBinding(kubeClient, namespace, roleBinding)

		if err != nil {
			log.Errorf("unable to delete the %s role binding  because %+v", roleBinding, err)
		}
	}

	if metav1.NamespaceAll != namespace {
		log.Infof("finished destroying synopsys operator in '%s' namespace", namespace)
	}
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
