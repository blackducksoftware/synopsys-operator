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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"

	soperator "github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getCmd lists resources in the cluster
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "List Synopsys Resources in your cluster",
	//(PassCmd) PreRunE: func(cmd *cobra.Command, args []string) error {
	//(PassCmd) 	// Display synopsysctl's Help instead of sending to oc/kubectl
	//(PassCmd) 	if len(args) == 1 && args[0] == "--help" {
	//(PassCmd) 		return fmt.Errorf("Help Called")
	//(PassCmd) 	}
	//(PassCmd) 	return nil
	//(PassCmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		//(PassCmd) log.Debugf("Getting a Non-Synopsys Resource\n")
		//(PassCmd) kubeCmdArgs := append([]string{"get"}, args...)
		//(PassCmd) out, err := util.RunKubeCmd(restconfig, kube, openshift, kubeCmdArgs...)
		//(PassCmd) if err != nil {
		//(PassCmd) 	log.Errorf("Error Getting the Resource: %s", out)
		//(PassCmd) 	return nil
		//(PassCmd) }
		//(PassCmd) fmt.Printf("%+v", out)
		//(PassCmd) return nil
		return fmt.Errorf("Not a Valid Command")
	},
}

// getBlackduckCmd lists Blackducks in the cluster
var getBlackduckCmd = &cobra.Command{
	Use:     "blackduck",
	Aliases: []string{"blackducks"},
	Short:   "Get a list of Blackducks in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting Blackducks\n")
		out, err := util.RunKubeCmd(restconfig, kube, openshift, "get", "blackducks")
		if err != nil {
			log.Errorf("Error getting Blackducks: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getBlackduckRootKeyCmd get the Black Duck root key for source code upload in the cluster
var getBlackduckRootKeyCmd = &cobra.Command{
	Use:   "rootKey BLACK_DUCK_NAME FILE_PATH",
	Short: "Get the root key of Black Duck for source code upload",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Black Duck name or file path to store the master key is missing")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting Blackduck Root Key\n")
		namespace := args[0]
		filePath := args[1]
		_, err := util.GetHub(blackduckClient, metav1.NamespaceDefault, namespace)
		if err != nil {
			log.Errorf("unable to find Black Duck %s instance due to %+v", namespace, err)
			return nil
		}
		operatorNamespace, err := soperator.GetOperatorNamespace(restconfig)
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator instance due to %+v", err)
			return nil
		}
		secret, err := util.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find the Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
			return nil
		}
		sealKey := string(secret.Data["SEAL_KEY"])
		// Filter the upload cache pod to get the root key using the seal key
		uploadCachePod, err := util.FilterPodByNamePrefixInNamespace(kubeClient, namespace, "uploadcache")
		if err != nil {
			log.Errorf("unable to filter the upload cache pod of %s due to %+v", namespace, err)
			return nil
		}

		// Create the exec into kubernetes pod request
		req := util.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
		stdout, err := util.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -f --header "X-SEAL-KEY: %s" https://uploadcache:9444/api/internal/master-key --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(sealKey)))})
		if err != nil {
			log.Errorf("unable to exec into upload cache pod in %s because %+v", namespace, err)
			return nil
		}

		fileName := filepath.Join(filePath, fmt.Sprintf("%s.key", namespace))
		err = ioutil.WriteFile(fileName, []byte(stdout), 0777)
		if err != nil {
			log.Errorf("error writing to %s because %+v", fileName, err)
			return nil
		}
		return nil
	},
}

// getOpsSightCmd lists OpsSights in the cluster
var getOpsSightCmd = &cobra.Command{
	Use:     "opssight",
	Aliases: []string{"opssights"},
	Short:   "Get a list of OpsSights in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting OpsSights\n")
		out, err := util.RunKubeCmd(restconfig, kube, openshift, "get", "opssights")
		if err != nil {
			log.Errorf("Error getting OpsSights: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getAlertCmd lists Alerts in the cluster
var getAlertCmd = &cobra.Command{
	Use:     "alert",
	Aliases: []string{"alerts"},
	Short:   "Get a list of Alerts in the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("This command accepts 0 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("Getting Alerts\n")
		out, err := util.RunKubeCmd(restconfig, kube, openshift, "get", "alerts")
		if err != nil {
			log.Errorf("Error getting Alerts with KubeCmd: %s", out)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

func init() {
	//(PassCmd) getCmd.DisableFlagParsing = true // lets getCmd pass flags to kube/oc
	rootCmd.AddCommand(getCmd)

	// Add Commands
	getCmd.AddCommand(getBlackduckCmd)
	getBlackduckCmd.AddCommand(getBlackduckRootKeyCmd)

	getCmd.AddCommand(getOpsSightCmd)
	getCmd.AddCommand(getAlertCmd)
}
