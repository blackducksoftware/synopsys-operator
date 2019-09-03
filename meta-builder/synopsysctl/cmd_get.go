/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package synopsysctl

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Get Command flag for -output functionality
var getOutputFormat string

// Get Command flag for -selector functionality
var getSelector string

// Get Command flag for --all-namespaces functionality
var getAllNamespaces bool

func generateKubectlGetCommand(resourceName string, args []string) []string {
	kubectlCmd := []string{"get", resourceName}
	if len(namespace) > 0 {
		kubectlCmd = append(kubectlCmd, "-n", namespace)
	}
	if len(args) > 0 {
		kubectlCmd = append(kubectlCmd, args...)
	}
	if getOutputFormat != "" {
		kubectlCmd = append(kubectlCmd, "-o")
		kubectlCmd = append(kubectlCmd, getOutputFormat)
	}
	if getSelector != "" {
		kubectlCmd = append(kubectlCmd, "-l")
		kubectlCmd = append(kubectlCmd, getSelector)
	}
	if getAllNamespaces {
		kubectlCmd = append(kubectlCmd, allNamespacesFlag)
	}
	return kubectlCmd
}

// getCmd lists resources in the cluster
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display Synopsys resources from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// getAlertCmd display one or many Alert instances
var getAlertCmd = &cobra.Command{
	Use:           "alert [NAME...]",
	Example:       "synopsysctl get alerts\nsynopsysctl get alert <name>\nsynopsysctl get alerts <name1> <name2>\nsynopsysctl get alerts -n <namespace>\nsynopsysctl get alert <name> -n <namespace>\nsynopsysctl get alerts <name1> <name2> -n <namespace>",
	Aliases:       []string{"alerts"},
	Short:         "Display one or many Alert instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting Alert instances...")
		out, err := RunKubeCmd(restconfig, kubeClient, generateKubectlGetCommand("alerts", args)...)
		if err != nil {
			return fmt.Errorf("error getting Alert instances due to %+v - %s", out, err)
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getBlackDuckCmd display one or many Black Duck instances
var getBlackDuckCmd = &cobra.Command{
	Use:           "blackduck [NAME...]",
	Example:       "synopsysctl get blackducks\nsynopsysctl get blackduck <name>\nsynopsysctl get blackducks <name1> <name2>\nsynopsysctl get blackducks -n <namespace>\nsynopsysctl get blackduck <name> -n <namespace>\nsynopsysctl get blackducks <name1> <name2> -n <namespace>",
	Aliases:       []string{"blackducks", "bds", "bd"},
	Short:         "Display one or many Black Duck instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting Black Duck instances...")
		out, err := RunKubeCmd(restconfig, kubeClient, generateKubectlGetCommand("blackducks", args)...)
		if err != nil {
			return fmt.Errorf("error getting Black Duck instances due to %+v - %s", out, err)
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

// getBlackDuckRootKeyCmd get Black Duck master key for source code upload in the cluster
//var getBlackDuckRootKeyCmd = &cobra.Command{
//	Use:           "masterkey BLACK_DUCK_NAME FILE_PATH_TO_STORE_MASTER_KEY",
//	Example:       "synopsysctl get blackduck masterkey <name> <file path to store the master key>\nsynopsysctl get blackduck masterkey <name> <file path to store the master key> -n <namespace>",
//	Short:         "Get the master key of the Black Duck instance that is used for source code upload and store it in the host",
//	SilenceUsage:  true,
//	SilenceErrors: true,
//	Args: func(cmd *cobra.Command, args []string) error {
//		if len(args) != 2 {
//			cmd.Help()
//			return fmt.Errorf("this command takes 2 arguments")
//		}
//		return nil
//	},
//	RunE: func(cmd *cobra.Command, args []string) error {
//		blackDuckName, blackDuckNamespace, crdScope, err := getInstanceInfo(false, BlackDuckCRDName, BlackDuckName, namespace, args[0])
//		if err != nil {
//			return err
//		}
//
//		// TODO
//		//operatorNamespace, err := util.GetOperatorNamespaceByCRDScope(kubeClient, BlackDuckCRDName, crdScope, blackDuckNamespace)
//		//if err != nil {
//		//	return fmt.Errorf("unable to find the Synopsys Operator instance due to %+v", err)
//		//}
//
//		operatorNamespace := "synopsys-operator"
//
//		filePath := args[1]
//		log.Infof("getting Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
//
//		_, err = utils.GetBlackduck(restClient, blackDuckNamespace, blackDuckName, v1.GetOptions{})
//		if err != nil {
//			return fmt.Errorf("error getting Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
//		}
//
//		// getting the operator secret to retrieve the seal key
//		secret, err := kubeClient.CoreV1().Secrets(operatorNamespace).Get("blackduck-secret",  v1.GetOptions{})
//		if err != nil {
//			return fmt.Errorf("unable to find Synopsys Operator's blackduck-secret in namespace '%s' due to %+v", operatorNamespace, err)
//		}
//
//		sealKey := string(secret.Data["SEAL_KEY"])
//		// Filter the upload cache pod to get the master key using the seal key
//		uploadCachePod, err := util.FilterPodByNamePrefixInNamespace(kubeClient, namespace, utils.GetResourceName(blackDuckName, util.BlackDuckName, "uploadcache"))
//		if err != nil {
//			return fmt.Errorf("unable to filter the upload cache pod in namespace '%s' due to %+v", namespace, err)
//		}
//
//		// Create the exec into Kubernetes pod request
//		req := util.CreateExecContainerRequest(kubeClient, uploadCachePod, "/bin/sh")
//		// TODO: changed the upload cache service name to authentication until the HUB-20412 is fixed. once it if fixed, changed the name to use GetResource method
//		stdout, err := util.ExecContainer(restconfig, req, []string{fmt.Sprintf(`curl -f --header "X-SEAL-KEY: %s" https://uploadcache:9444/api/internal/master-key --cert /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.crt --key /opt/blackduck/hub/blackduck-upload-cache/security/blackduck-upload-cache-server.key --cacert /opt/blackduck/hub/blackduck-upload-cache/security/root.crt`, base64.StdEncoding.EncodeToString([]byte(sealKey)))})
//		if err != nil {
//			return fmt.Errorf("unable to exec into upload cache pod in namespace '%s' due to %+v", namespace, err)
//		}
//
//		fileName := filepath.Join(filePath, fmt.Sprintf("%s-%s.key", blackDuckNamespace, blackDuckName))
//		err = ioutil.WriteFile(fileName, []byte(stdout), 0777)
//		if err != nil {
//			return fmt.Errorf("error writing to file '%s' due to %+v", fileName, err)
//		}
//		log.Infof("successfully created the master key in file '%s' for Black Duck '%s' in namespace '%s'", fileName, blackDuckName, blackDuckNamespace)
//		return nil
//	},
//}

// getOpsSightCmd display one or many OpsSight instances
var getOpsSightCmd = &cobra.Command{
	Use:           "opssight [NAME...]",
	Example:       "synopsysctl get opssights\nsynopsysctl get opssight <name>\nsynopsysctl get opssights <name1> <name2>",
	Aliases:       []string{"opssights", "ops"},
	Short:         "Display one or many OpsSight instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debugf("getting OpsSight instances...")
		out, err := RunKubeCmd(restconfig, kubeClient, generateKubectlGetCommand("opssights", args)...)
		if err != nil {
			return fmt.Errorf("error getting OpsSight instances due to %+v - %s", out, err)
		}
		fmt.Printf("%+v", out)
		return nil
	},
}

func init() {
	//(PassCmd) getCmd.DisableFlagParsing = true // lets getCmd pass flags to kube/oc
	rootCmd.AddCommand(getCmd)

	// Add Commands
	getAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	getAlertCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format (json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...)")
	getAlertCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getAlertCmd.Flags().BoolVar(&getAllNamespaces, "all-namespaces", getAllNamespaces, "If present, list the requested object(s) across all namespaces")
	getCmd.AddCommand(getAlertCmd)

	getBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	getBlackDuckCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format (json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...)")
	getBlackDuckCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getBlackDuckCmd.Flags().BoolVar(&getAllNamespaces, "all-namespaces", getAllNamespaces, "If present, list the requested object(s) across all namespaces")

	//getBlackDuckRootKeyCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	//getBlackDuckCmd.AddCommand(getBlackDuckRootKeyCmd)
	getCmd.AddCommand(getBlackDuckCmd)

	getOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	getOpsSightCmd.Flags().StringVarP(&getOutputFormat, "output", "o", getOutputFormat, "Output format (json,yaml,wide,name,custom-columns=...,custom-columns-file=...,go-template=...,go-template-file=...,jsonpath=...,jsonpath-file=...)")
	getOpsSightCmd.Flags().StringVarP(&getSelector, "selector", "l", getSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	getOpsSightCmd.Flags().BoolVar(&getAllNamespaces, "all-namespaces", getAllNamespaces, "If present, list the requested object(s) across all namespaces")
	getCmd.AddCommand(getOpsSightCmd)
}
