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
	"fmt"

	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

// Destroy Command Defaults
var destroyNamespace = "synopsys-operator"

// destroyCmd removes the Synopsys-Operator from the cluster
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Removes the Synopsys-Operator and CRDs from Cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) != 0 {
			return fmt.Errorf("this command takes 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the namespace of the Synopsys-Operator
		destroyNamespace, err := GetOperatorNamespace()
		if err != nil {
			log.Warnf("error finding synopsys operator due to %+v", err)
		}
		log.Infof("Destroying the Synopsys-Operator '%s'...", destroyNamespace)

		// delete  namespace
		log.Debugf("Deleting namespace %s", destroyNamespace)
		err = util.DeleteNamespace(kubeClient, destroyNamespace)
		if err != nil {
			log.Warnf("Unable to delete the %s namespace because %+v", destroyNamespace, err)
		}

		// delete crds
		log.Debugf("Deleting CRDs")
		apiExtensionClient, err := apiextensionsclient.NewForConfig(restconfig)
		if err != nil {
			log.Errorf("error creating the api extension client due to %+v", err)
		}

		crds := []string{"alerts.synopsys.com", "blackducks.synopsys.com", "hubs.synopsys.com", "opssights.synopsys.com"}

		for _, crd := range crds {
			err = util.DeleteCustomResourceDefinition(apiExtensionClient, crd)
			if err != nil {
				log.Warnf("Unable to delete the %s crd because %+v", crd, err)
			}
		}

		// delete cluster role bindings
		log.Debugf("Deleting ClusterRoleBinding")
		err = util.DeleteClusterRoleBinding(kubeClient, "synopsys-operator-admin")
		if err != nil {
			log.Warnf("Unable to delete the synopsys-operator-admin cluster role binding because %+v", err)
		}

		// delete cluster roles
		log.Debugf("Deleting ClusterRoles")
		err = util.DeleteClusterRole(kubeClient, "synopsys-operator-admin")
		if err != nil {
			log.Warnf("Unable to delete the synopsys-operator-admin cluster role because %+v", err)
		}

		log.Infof("Finished destroying synopsys-operator: '%s'", destroyNamespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
