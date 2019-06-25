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
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	util "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	routeClient := operatorutil.GetRouteClient(restConfig, kubeClient, metav1.NamespaceAll) // kube doesn't have a route client but openshift does
	if routeClient != nil {
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

func getOperatorNamespace(namespace string) (string, error) {
	var err error
	if len(namespace) == 0 {
		isClusterScoped := util.GetClusterScope(apiExtensionClient)
		if isClusterScoped {
			namespace = metav1.NamespaceAll
		}
	}

	log.Debugf("getting Synopsys Operator's namespace")
	namespace, err = util.GetOperatorNamespace(kubeClient, namespace)
	if err != nil {
		return "", err
	}

	if len(namespace) == 0 {
		return "", fmt.Errorf("Synopsys Operator's namespace not found")
	}
	return namespace, nil
}

// getInstanceInfo will provide the name, namespace and crd scope to create each CRD instance
func getInstanceInfo(cmd *cobra.Command, name string, crdType string, crdName string, inputNamespace string) (string, string, apiextensions.ResourceScope, error) {
	crdScope := apiextensions.ClusterScoped
	if cmd.Flags().Lookup("mock") != nil && !cmd.Flags().Lookup("mock").Changed && !cmd.Flags().Lookup("mock-kube").Changed {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdType)
		if err != nil {
			return "", "", "", fmt.Errorf("unable to get Custom Resource Definition '%s' in your cluster due to %+v", crdType, err)
		}
		crdScope = crd.Spec.Scope
	}

	// Check Number of Arguments
	if crdScope != apiextensions.ClusterScoped && len(namespace) == 0 {
		return "", "", "", fmt.Errorf("namespace to create an '%s' instance needs to be provided", inputNamespace)
	}

	var namespace string
	if crdScope == apiextensions.ClusterScoped {
		if len(inputNamespace) == 0 {
			if len(crdName) > 0 {
				ns, err := util.ListNamespaces(kubeClient, fmt.Sprintf("synopsys.com/%s.%s", crdName, name))
				if err != nil {
					return "", "", "", fmt.Errorf("unable to list the '%s' instance '%s' in namespace '%s' due to %+v", crdName, name, namespace, err)
				}
				if len(ns.Items) > 0 {
					namespace = ns.Items[0].Name
				} else {
					return "", "", "", fmt.Errorf("unable to find the namespace of the '%s' instance '%s'", crdName, name)
				}
			}
			namespace = name
		} else {
			namespace = inputNamespace
		}
	} else {
		namespace = inputNamespace
	}
	return name, namespace, crdScope, nil
}

// checkOperatorIsRunning returns true if there is a usable operator for the resource
// based on the cluster scope
func checkOperatorIsRunning(clusterScope apiextensions.ResourceScope, resourceNamespace string) error {
	// Check if Synopsys Operator is running
	var sOperatorNamespace string
	if clusterScope == apiextensions.ClusterScoped {
		sOperatorNamespace = metav1.NamespaceAll
	} else {
		sOperatorNamespace = resourceNamespace
	}
	var exists bool
	if exists = util.IsOperatorExist(kubeClient, sOperatorNamespace); !exists {
		return fmt.Errorf("Synopsys Operator must be running in namespace '%s'", sOperatorNamespace)
	}
	return nil
}
