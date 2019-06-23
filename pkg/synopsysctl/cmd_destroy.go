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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var isForceDestroy bool

// destroyCmd removes Synopsys Operator from the cluster
var destroyCmd = &cobra.Command{
	Use:           "destroy [NAMESPACE...]",
	Example:       "synopsysctl destroy\nsynopsysctl destroy --force\nsynopsysctl destroy <namespace>\nsynopsysctl destroy <namespace> --force\nsynopsysctl destroy <namespace1> <namespace2>",
	Short:         "Remove one or many Synopsys Operator instances and their associated CRDs",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read Commandline Parameters
		if len(args) > 0 {
			for _, operatorNamespace := range args {
				return destroy(operatorNamespace)
			}
		} else {
			operatorNamespace := DefaultOperatorNamespace
			var err error
			isClusterScoped := util.GetClusterScope(apiExtensionClient)
			if isClusterScoped {
				namespace, err = util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
				if err != nil {
					return err
				}
				if metav1.NamespaceAll != namespace {
					operatorNamespace = namespace
				}
			}
			return destroy(operatorNamespace)
		}
		return nil
	},
}

func destroy(namespace string) error {
	log.Infof("destroying Synopsys Operator in namespace '%s'...", namespace)
	crds := []string{}
	cm, err := util.GetConfigMap(kubeClient, namespace, "synopsys-operator")
	if err != nil {
		log.Errorf("unable to find the 'synopsy-operator' config map in namespace '%s' due to %+v", namespace, err)
	} else {
		data := cm.Data["config.json"]
		var cmData map[string]interface{}
		err = json.Unmarshal([]byte(data), &cmData)
		if err != nil {
			log.Errorf("unable to unmarshal config map data due to %+v", err)
		}
		if crdNames, ok := cmData["CrdNames"]; ok {
			crds = strings.Split(crdNames.(string), ",")
		}
	}

	// delete namespace
	checkErr := util.CheckResourceNamespace(restconfig, kubeClient, strings.Join(crds, ","), namespace, true)
	if err != nil && isForceDestroy {
		log.Warnf("%s. namespace cannot be deleted", err.Error())
	} else if err != nil && !isForceDestroy {
		return fmt.Errorf("%s. It is not recommended to destroy the Synopsys Operator so these resources can continue to be managed. If you are sure you want to delete the Synopsys Operator anyway then you can use the 'force' option which will keep all the instances and delete only the Synopsys Operator", err.Error())
	}

	log.Infof("deleting the Synopsys Operator in namespace '%s'", namespace)
	// delete Synopsys Operator instance
	commonConfig := crdupdater.NewCRUDComponents(restconfig, kubeClient, false, false, namespace, &api.ComponentList{}, "app=synopsys-operator", false)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		log.Errorf("unable to delete the Synopsys Operator in namespace '%s' due to %+v", namespace, err)
	}

	if checkErr == nil {
		err = util.DeleteNamespace(kubeClient, namespace)
		if err != nil {
			log.Errorf("unable to delete the Synopsys Operator namespace due to %+v", err)
		}
	}

	// delete crds
	deleteCrds(crds, namespace)

	// delete cluster role bindings
	clusterRoleBindings, roleBindings, err := util.GetOperatorRoleBindings(kubeClient, namespace)
	if err != nil {
		log.Errorf("error getting role binding or cluster role binding due to %+v", err)
	}

	for _, clusterRoleBinding := range clusterRoleBindings {
		crb, err := util.GetClusterRoleBinding(kubeClient, clusterRoleBinding)
		if err != nil {
			log.Errorf("unable to get cluster role binding '%s' due to %+v", clusterRoleBinding, err)
		}
		// check whether any subject present for other namespace before deleting them
		newSubjects := []rbacv1.Subject{}
		for _, subject := range crb.Subjects {
			isExist := util.IsSubjectExistForOtherNamespace(subject, namespace)
			if isExist {
				newSubjects = append(newSubjects, subject)
			}
		}
		if len(newSubjects) > 0 {
			crb.Subjects = newSubjects
			// update the cluster role binding to remove the old cluster role binding subject
			log.Infof("updating cluster role binding '%s'", clusterRoleBinding)
			_, err = util.UpdateClusterRoleBinding(kubeClient, crb)
			if err != nil {
				log.Errorf("unable to update cluster role binding '%s' due to %+v", clusterRoleBinding, err)
			}
		} else {
			log.Infof("deleting cluster role binding '%s'", clusterRoleBinding)
			err := util.DeleteClusterRoleBinding(kubeClient, clusterRoleBinding)
			if err != nil {
				log.Errorf("unable to delete cluster role binding '%s' due to %+v", clusterRoleBinding, err)
			}
		}
	}

	// delete role bindings
	for _, roleBinding := range roleBindings {
		log.Infof("deleting role binding '%s'", roleBinding)
		err = util.DeleteRoleBinding(kubeClient, namespace, roleBinding)

		if err != nil {
			log.Errorf("unable to delete role binding '%s' due to %+v", roleBinding, err)
		}
	}

	// delete cluster roles
	clusterRoles, roles, err := util.GetOperatorRoles(kubeClient, namespace)
	if err != nil {
		log.Errorf("error getting role or cluster role due to %+v", err)
	}

	crbs, err := util.ListClusterRoleBindings(kubeClient, "app in (synopsys-operator,opssight)")

	for _, clusterRole := range clusterRoles {
		isExist := false
		// check whether the cluster role is referenced in any cluster role binding
		for _, crb := range crbs.Items {
			if util.IsClusterRoleRefExistForOtherNamespace(crb.RoleRef, clusterRole, namespace, crb.Subjects) {
				isExist = true
			}
		}
		if !isExist {
			log.Infof("deleting cluster role '%s'", clusterRole)
			err := util.DeleteClusterRole(kubeClient, clusterRole)
			if err != nil {
				log.Errorf("unable to delete cluster role '%s' due to %+v", clusterRole, err)
			}
		}
	}

	// delete roles
	for _, role := range roles {
		log.Infof("deleting role '%s'", role)
		err := util.DeleteRole(kubeClient, namespace, role)
		if err != nil {
			log.Errorf("unable to delete role '%s' due to %+v", role, err)
		}
	}

	log.Infof("successfully submitted destroy Synopsys Operator in namespace '%s'", namespace)
	return nil
}

// isOtherNamespaceExistInCRDLabel return whether any other namespace exist in the Synopsys CRD namespace label
func isOtherNamespaceExistInCRDLabel(crd *apiextensions.CustomResourceDefinition, namespace string) (bool, error) {
	for key, value := range crd.Labels {
		if strings.HasPrefix(key, "synopsys.com/operator.") {
			if value != namespace {
				delete(crd.Labels, fmt.Sprintf("synopsys.com/operator.%s", namespace))
				_, err := util.UpdateCustomResourceDefinition(apiExtensionClient, crd)
				if err != nil {
					return true, fmt.Errorf("unable to update the labels for %s custom resource definition due to %+v", crd, err)
				}
				return true, fmt.Errorf("%s custom resource definition is already in use by other namespaces and hence removed the namespace operator label from the CRD", crd.Name)
			}
		}
	}
	return false, nil
}

func isDeleteCrd(crd string, namespace string) error {
	switch crd {
	case util.AlertCRDName:
		// check custom resource definition exist
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.AlertCRDName)
		if err != nil {
			return fmt.Errorf("unable to get %s custom resource definition due to %+v", util.AlertCRDName, err)
		}

		// check whether any other namespace is using the CRD
		if isExist, err := isOtherNamespaceExistInCRDLabel(crd, namespace); isExist {
			return err
		}

		// check whether any alert instance is running in the namespace
		alerts, err := util.ListAlerts(alertClient, namespace)
		if err != nil {
			return fmt.Errorf("unable to list Alert instances due to %+v", err)
		}

		if len(alerts.Items) > 0 {
			return fmt.Errorf("already Alert instances exist in other namespaces. Please delete them before deleting the custom resources")
		}
	case util.BlackDuckCRDName:
		// check custom resource definition exist
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.BlackDuckCRDName)
		if err != nil {
			return fmt.Errorf("unable to get %s custom resource definition due to %+v", util.BlackDuckCRDName, err)
		}

		// check whether any other namespace is using the CRD
		if isExist, err := isOtherNamespaceExistInCRDLabel(crd, namespace); isExist {
			return err
		}

		// check whether any alert instance is running in the namespace
		blackDucks, err := util.ListHubs(blackDuckClient, namespace)
		if err != nil {
			return fmt.Errorf("unable to list Black Duck instances due to %+v", err)
		}

		if len(blackDucks.Items) > 0 {
			return fmt.Errorf("already Black Duck instances exist in other namespaces. Please delete them before deleting the custom resources")
		}
	case util.OpsSightCRDName:
		// check custom resource definition exist
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, util.OpsSightCRDName)
		if err != nil {
			return fmt.Errorf("unable to get %s custom resource definition due to %+v", util.OpsSightCRDName, err)
		}

		// check whether any other namespace is using the CRD
		if isExist, err := isOtherNamespaceExistInCRDLabel(crd, namespace); isExist {
			return err
		}

		// check whether any alert instance is running in the namespace
		opsSights, err := util.ListOpsSights(opsSightClient, namespace)
		if err != nil {
			return fmt.Errorf("unable to list OpsSight instances due to %+v", err)
		}

		if len(opsSights.Items) > 0 {
			return fmt.Errorf("already OpsSight instances exist in other namespaces. Please delete them before deleting the custom resources")
		}
	}
	return nil
}

// deleteCrds will check and delete multiple custom resource definition
func deleteCrds(crds []string, namespace string) {
	for _, crd := range crds {
		deleteCrd(strings.TrimSpace(crd), namespace)
	}
}

// deleteCrd will check and delete the custom resource definition
func deleteCrd(crd string, namespace string) {
	err := isDeleteCrd(crd, namespace)
	if err != nil {
		log.Warn(err)
	} else {
		log.Infof("deleting Custom Resource Definition '%s'", crd)
		err := util.DeleteCustomResourceDefinition(apiExtensionClient, crd)
		if err != nil {
			log.Errorf("unable to delete Custom Resource Definition '%s' due to %+v", crd, err)
		}
	}
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&isForceDestroy, "force", "f", isForceDestroy, "Forcefully destroy the Synopsys Operator in your cluster")
}
