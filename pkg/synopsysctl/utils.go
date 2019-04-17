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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	operatorutil "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// These vars set by setResourceClients() in root command's init()
var restconfig *rest.Config
var kubeClient *kubernetes.Clientset
var blackduckClient *blackduckclientset.Clientset
var opssightClient *opssightclientset.Clientset
var alertClient *alertclientset.Clientset

// These vars used by KubeCmd
var openshift bool
var kube bool

// setResourceClients sets the global variables for the kuberentes rest config
// and the resource clients
func setResourceClients() {
	var err error
	restconfig, err = protoform.GetKubeConfig(kubeconfig, insecureSkipTLSVerify)
	if err != nil {
		log.Errorf("error getting Kube Rest Config: %s", err)
	}
	kubeClient, err = getKubeClient(restconfig)
	if err != nil {
		log.Errorf("error getting Kube Client: %s", err)
	}
	blackduckClient, err = blackduckclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Blackduck Clientset: %s", err)
	}
	opssightClient, err = opssightclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the OpsSight Clientset: %s", err)
	}
	alertClient, err = alertclientset.NewForConfig(restconfig)
	if err != nil {
		log.Errorf("error creating the Alert Clientset: %s", err)
	}
	kube, openshift = DetermineClusterClients(restconfig)
}

// getKubeClient gets the kubernetes client
func getKubeClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// DeployCRDNamespace creates an empty Horizon namespace
func DeployCRDNamespace(restconfig *rest.Config, namespace string) error {
	namespaceDeployer, err := deployer.NewDeployer(restconfig)
	ns := horizoncomponents.NewNamespace(horizonapi.NamespaceConfig{
		Name:      namespace,
		Namespace: namespace,
	})
	namespaceDeployer.AddNamespace(ns)
	err = namespaceDeployer.Run()
	if err != nil {
		return fmt.Errorf("error in creating the namespace due to %+v", err)
	}
	return nil
}

// GetOperatorNamespace returns the namespace of the Synopsys-Operator by
// looking at its cluster role binding
func GetOperatorNamespace() (string, error) {
	// check if operator is already installed
	crb, err := operatorutil.GetClusterRoleBinding(kubeClient, "synopsys-operator-admin")
	if err != nil || len(crb.Subjects) == 0 {
		return "", fmt.Errorf("failed to get synopsys operator namespace due to %+v", err)
	}
	return crb.Subjects[0].Namespace, nil
}

// DetermineClusterClients returns bool values for which client
// to use. They will never both be true
func DetermineClusterClients(restConfig *rest.Config) (kube, openshift bool) {
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
	routeClient := operatorutil.GetRouteClient(restConfig) // kube doesn't have a route client but openshift does
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
func RunKubeCmd(restConfig *rest.Config, kube bool, openshift bool, args ...string) (string, error) {
	var cmd2 *exec.Cmd

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
		return "", fmt.Errorf("Could not determine if openshift or kube")
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
func RunKubeEditorCmd(restConfig *rest.Config, kube bool, openshift bool, args ...string) error {
	var cmd *exec.Cmd

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
