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
	"os"
	"os/exec"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	sizeclientset "github.com/blackducksoftware/synopsys-operator/pkg/size/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// These vars set by setResourceClients() in root command's init()
var restconfig *rest.Config
var kubeClient *kubernetes.Clientset
var apiExtensionClient *apiextensionsclient.Clientset
var alertClient *alertclientset.Clientset
var blackDuckClient *blackduckclientset.Clientset
var opsSightClient *opssightclientset.Clientset
var sizeClient *sizeclientset.Clientset

func parseLogLevelAndKubeConfig(cmd *cobra.Command) error {
	// Set the Log Level
	lvl, err := log.ParseLevel(logLevelCtl)
	if err != nil {
		log.Errorf("ctl-log-Level '%s' is not a valid level: %s", logLevelCtl, err)
		return err
	}
	log.SetLevel(lvl)
	if !cmd.Flags().Lookup("kubeconfig").Changed { // if kubeconfig wasn't set, check the environ
		if kubeconfigEnvVal, exists := os.LookupEnv("KUBECONFIG"); exists { // set kubeconfig if environ is set
			kubeconfig = kubeconfigEnvVal
		}
	}
	return nil
}

func callSetResourceClients() {
	// Sets kubeconfig and initializes resource client libraries
	if err := setResourceClients(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// setResourceClients sets the global variables for the Kuberentes rest config
// and the resource clients
func setResourceClients() error {
	var err error
	restconfig, err = protoform.GetKubeConfig(kubeconfig, insecureSkipTLSVerify)
	log.Debugf("rest config: %+v", restconfig)
	if err != nil {
		return err
	}
	kubeClient, err = getKubeClient(restconfig)
	if err != nil {
		return err
	}
	apiExtensionClient, err = apiextensionsclient.NewForConfig(restconfig)
	if err != nil {
		return err
	}
	alertClient, err = alertclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating Alert Clientset: %s", err)
	}
	blackDuckClient, err = blackduckclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating Black Duck Clientset: %s", err)
	}
	opsSightClient, err = opssightclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating OpsSight Clientset: %s", err)
	}
	sizeClient, err = sizeclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating Size Clientset: %s", err)
	}
	return nil
}

// getKubeClient gets the kubernetes client
func getKubeClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// DetermineClusterClients returns bool values for which client
// to use. They will never both be true
func DetermineClusterClients(restConfig *rest.Config, kubeClient *kubernetes.Clientset) (kube, openshift bool) {
	openshift = false
	kube = false

	kubectlPath := false
	ocPath := false
	_, exists := exec.LookPath("kubectl")
	if exists == nil {
		kubectlPath = true
	}
	_, ocexists := exec.LookPath("oc")
	if ocexists == nil {
		ocPath = true
	}

	// Add Openshift rules
	openshiftTest := false
	if util.IsOpenshift(kubeClient) {
		openshiftTest = true
	}

	if ocPath && openshiftTest { // if oc exists and the cluster is openshift
		log.Debugf("oc exists and the cluster is openshift")
		return false, true
	}
	if kubectlPath && !openshiftTest { // if kubectl exists and it isn't openshift
		log.Debugf("kubectl exists and it isn't openshift")
		return true, false
	}
	if kubectlPath && !ocPath && openshiftTest { // if kubectl exists, oc doesn't exist, and it is openshift
		log.Debugf("kubectl exists, oc doesn't exist, and it is openshift")
		return true, false
	}
	if ocPath && !kubectlPath && !openshiftTest { // If oc exists, kubectl doesn't exist, and it isn't openshift
		log.Debugf("oc exists, kubectl doesn't exist, and it isn't openshift")
		return false, true
	}
	return false, false // neither client exists
}

// RunKubeCmd is a simple wrapper to oc/kubectl exec that captures output.
// TODO consider replacing w/ go api but not crucial for now.
func RunKubeCmd(restconfig *rest.Config, kubeClient *kubernetes.Clientset, args ...string) (string, error) {
	var cmd2 *exec.Cmd
	kube, openshift := DetermineClusterClients(restconfig, kubeClient)

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
	}
	// add global flags: insecure-skip-tls-verify and --kubeconfig
	if insecureSkipTLSVerify == true {
		args = append([]string{fmt.Sprintf("--insecure-skip-tls-verify=%t", insecureSkipTLSVerify)}, args...)
	}
	if kubeconfig != "" {
		args = append([]string{fmt.Sprintf("--kubeconfig=%s", kubeconfig)}, args...)
	}
	if openshift {
		cmd2 = exec.Command("oc", args...)
		log.Debugf("Command: %+v", cmd2.Args)
	} else if kube {
		cmd2 = exec.Command("kubectl", args...)
		log.Debugf("Command: %+v", cmd2.Args)
	} else {
		return "", fmt.Errorf("couldn't determine if running in Openshift or Kubernetes")
	}
	stdoutErr, err := cmd2.CombinedOutput()
	if err != nil {
		return string(stdoutErr), err
	}
	//time.Sleep(1 * time.Second) TODO why did Jay put this here???
	return string(stdoutErr), nil
}

// RunKubeEditorCmd is a wrapper for oc/kubectl but redirects
// input/output to the user - ex: let user control text editor
func RunKubeEditorCmd(restConfig *rest.Config, kubeClient *kubernetes.Clientset, args ...string) error {
	var cmd *exec.Cmd
	kube, openshift := DetermineClusterClients(restconfig, kubeClient)

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
	}
	// add global flags: insecure-skip-tls-verify and --kubeconfig
	if insecureSkipTLSVerify == true {
		args = append([]string{fmt.Sprintf("--insecure-skip-tls-verify=%t", insecureSkipTLSVerify)}, args...)
	}
	if kubeconfig != "" {
		args = append([]string{fmt.Sprintf("--kubeconfig=%s", kubeconfig)}, args...)
	}
	if openshift {
		cmd = exec.Command("oc", args...)
	} else if kube {
		cmd = exec.Command("kubectl", args...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	//time.Sleep(1 * time.Second) TODO why did Jay put this here???
	return nil
}

// getInstanceInfo provides the name, namespace and crd scope of the request custom resource instance
func getInstanceInfo(mock bool, crdName string, appName string, namespace string, name string) (string, string, apiextensions.ResourceScope, error) {
	crdScope := apiextensions.ClusterScoped
	if !mock {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdName)
		if err != nil {
			return "", "", "", fmt.Errorf("unable to get Custom Resource Definition '%s' in your cluster due to %+v", crdName, err)
		}
		crdScope = crd.Spec.Scope
	}

	// if the CRD scope is namespaced scope, then the user need to provide the namespace
	if crdScope != apiextensions.ClusterScoped && len(namespace) == 0 {
		return "", "", "", fmt.Errorf("namespace needs to be provided. please use the 'namespace' option to set it")
	}

	if crdScope == apiextensions.ClusterScoped {
		if len(namespace) == 0 {
			namespace = name
			// update scenrio to find out the namespace in case of cluster scope
			if len(appName) > 0 {
				ns, err := util.ListNamespaces(kubeClient, fmt.Sprintf("synopsys.com/%s.%s", appName, name))
				if err != nil {
					return "", "", "", fmt.Errorf("unable to list the '%s' instance '%s' in namespace '%s' due to %+v", appName, name, namespace, err)
				}
				if len(ns.Items) > 0 {
					namespace = ns.Items[0].Name
				} else {
					return "", "", "", fmt.Errorf("unable to find the namespace of the '%s' instance '%s'", appName, name)
				}
			}
		}
	}

	return name, namespace, crdScope, nil
}
