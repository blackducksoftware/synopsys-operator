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

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get Command flag for -output functionality
var getOutputFormat string

// Get Command flag for -selector functionality
var getSelector string

// getCmd lists resources in the cluster
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display Synopsys resources from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not a valid command")
	},
}

// getAlertCmd display one or many Alerts
var getAlertCmd = &cobra.Command{
	Use:     "alert [NAME]...",
	Example: "synopsysctl get alerts\nsynopsysctl get alert <name>\nsynopsysctl get alerts <name1> <name2>\nsynopsysctl get alerts -n <namespace>\nsynopsysctl get alert <name> -n <namespace>\nsynopsysctl get alerts <name1> <name2> -n <namespace>",
	Aliases: []string{"alerts"},
	Short:   "Display one or many Alerts",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting Alerts...")
		kubectlCmd := []string{"get", "alerts"}
		if len(namespace) > 0 {
			kubectlCmd = append(kubectlCmd, "-n", namespace)
		}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("output").Changed {
			kubectlCmd = append(kubectlCmd, "-o")
			kubectlCmd = append(kubectlCmd, getOutputFormat)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, getSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error getting Alerts due to %+v - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getBlackDuckCmd Display one or many Black Ducks
var getBlackDuckCmd = &cobra.Command{
	Use:     "blackduck [NAME]...",
	Example: "synopsysctl get blackducks\nsynopsysctl get blackduck <name>\nsynopsysctl get blackducks <name1> <name2>\nsynopsysctl get blackducks -n <namespace>\nsynopsysctl get blackduck <name> -n <namespace>\nsynopsysctl get blackducks <name1> <name2> -n <namespace>",
	Aliases: []string{"blackducks"},
	Short:   "Display one or many Black Ducks",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting Black Ducks...")
		kubectlCmd := []string{"get", "blackducks"}
		if len(namespace) > 0 {
			kubectlCmd = append(kubectlCmd, "-n", namespace)
		}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("output").Changed {
			kubectlCmd = append(kubectlCmd, "-o")
			kubectlCmd = append(kubectlCmd, getOutputFormat)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, getSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error getting Black Ducks due to %+v - %s", out, err)
			return nil
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getBlackDuckRootKeyCmd get Black Duck root key for source code upload in the cluster
var getBlackDuckRootKeyCmd = &cobra.Command{
	Use:     "rootkey NAMESPACE FILE_PATH",
	Example: "synopsysctl get blackduck rootkey bdnamespace ~/home/tmp/key",
	Short:   "Get the root key of Black Duck for source code upload",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("this command takes 2 arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := args[0]
		filePath := args[1]

		log.Debugf("getting Black Duck %s Root Key...", namespace)

		_, err := util.GetHub(blackduckClient, namespace, namespace)
		if err != nil {
			log.Errorf("unable to find Black Duck %s instance due to %+v", namespace, err)
			return nil
		}

		// check for any CRD cluster scope, if so, use default namespace or use the blackduck namespace to find the operator's namespace
		operatorNamespace := namespace
		isClusterScoped := util.GetClusterScope(apiExtensionClient)
		if isClusterScoped {
			operatorNamespace = metav1.NamespaceAll
		}

		log.Debugf("getting synopsys operator's secret")
		operatorNamespace, err = util.GetOperatorNamespace(kubeClient, operatorNamespace)
		if err != nil || len(operatorNamespace) == 0 {
			log.Errorf("unable to find the synopsys operator instance due to %+v", err)
			return nil
		}

		// getting the operator secret to retrieve the seal key
		secret, err := util.GetSecret(kubeClient, operatorNamespace, "blackduck-secret")
		if err != nil {
			log.Errorf("unable to find Synopsys Operator blackduck-secret in %s namespace due to %+v", operatorNamespace, err)
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
		log.Infof("successfully wrote Root Key to %s", fileName)
		return nil
	},
}

// getOpsSightCmd Display one or many OpsSights
var getOpsSightCmd = &cobra.Command{
	Use:     "opssight [NAME]...",
	Example: "synopsysctl get opssights\nsynopsysctl get opssight <name>\nsynopsysctl get opssights <name1> <name2>",
	Aliases: []string{"opssights"},
	Short:   "Display one or many OpsSights",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting OpsSights...")
		kubectlCmd := []string{"get", "opssights"}
		if len(args) > 0 {
			kubectlCmd = append(kubectlCmd, args...)
		}
		if cmd.LocalFlags().Lookup("output").Changed {
			kubectlCmd = append(kubectlCmd, "-o")
			kubectlCmd = append(kubectlCmd, getOutputFormat)
		}
		if cmd.LocalFlags().Lookup("selector").Changed {
			kubectlCmd = append(kubectlCmd, "-l")
			kubectlCmd = append(kubectlCmd, getSelector)
		}
		out, err := RunKubeCmd(restconfig, kubectlCmd...)
		if err != nil {
			log.Errorf("error getting OpsSights due to %+v - %s", out, err)
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
	getAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to get the resource(s)")
	getAlertCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format [json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...]")
	getAlertCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getCmd.AddCommand(getAlertCmd)

	getBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace of the synopsys operator to get the resource(s)")
	getBlackDuckCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format [json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...]")
	getBlackDuckCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getBlackDuckCmd.AddCommand(getBlackDuckRootKeyCmd)
	getCmd.AddCommand(getBlackDuckCmd)

	getOpsSightCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format [json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...]")
	getOpsSightCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getCmd.AddCommand(getOpsSightCmd)
}
