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
	"reflect"

	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertapi "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
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
var blackduckClient *blackduckclientset.Clientset
var opssightClient *opssightclientset.Clientset
var alertClient *alertclientset.Clientset
var blackDuckClient *blackduckclientset.Clientset
var opsSightClient *opssightclientset.Clientset

// setResourceClients sets the global variables for the kuberentes rest config
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
	routeClient := operatorutil.GetRouteClient(restConfig, metav1.NamespaceAll) // kube doesn't have a route client but openshift does
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
func RunKubeCmd(restConfig *rest.Config, args ...string) (string, error) {
	var cmd2 *exec.Cmd
	kube, openshift := DetermineClusterClients(restconfig)

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
func RunKubeEditorCmd(restConfig *rest.Config, args ...string) error {
	var cmd *exec.Cmd
	kube, openshift := DetermineClusterClients(restconfig)

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

	log.Debugf("getting synopsys operator's namespace")
	namespace, err = util.GetOperatorNamespace(kubeClient, namespace)
	if err != nil {
		return "", err
	}

	if len(namespace) == 0 {
		return "", fmt.Errorf("synopsys operator namespace not found")
	}
	return namespace, nil
}

func ctlUpdateResource(resource interface{}, mock bool, mockFormat string, kubeMock bool, mockKubeFormat string) error {
	if mock {
		log.Debugf("running mock mode")
		err := PrintResource(resource, mockFormat, false)
		if err != nil {
			return fmt.Errorf("failed to print resource: %s", err)
		}
	} else if kubeMock {
		log.Debugf("running kube mock mode")
		err := PrintResource(resource, mockKubeFormat, true)
		if err != nil {
			return fmt.Errorf("failed to print resource: %s", err)
		}
	} else {
		switch reflect.TypeOf(resource) {
		case reflect.TypeOf(alertapi.Alert{}):
			alt := resource.(alertapi.Alert)
			_, err := operatorutil.UpdateAlert(alertClient, alt.Spec.Namespace, &alt)
			if err != nil {
				log.Errorf("error updating the %s Alert instance due to %+v", alt.Name, err)
				return nil
			}
		case reflect.TypeOf(blackduckapi.Blackduck{}):
			bd := resource.(blackduckapi.Blackduck)
			_, err := operatorutil.UpdateBlackduck(blackDuckClient, bd.Spec.Namespace, &bd)
			if err != nil {
				log.Errorf("error updating the %s Black Duck instance due to %+v", bd.Name, err)
				return nil
			}
		case reflect.TypeOf(opssightapi.OpsSight{}):
			ops := resource.(opssightapi.OpsSight)
			_, err := operatorutil.UpdateOpsSight(opsSightClient, ops.Spec.Namespace, &ops)
			if err != nil {
				log.Errorf("error updating the %s OpsSight instance due to %+v", ops.Name, err)
				return nil
			}
		default:
			return fmt.Errorf("type %+v is not supported for updating", reflect.TypeOf(resource))
		}
	}
	return nil
}

// getInstanceInfo will provide the name, namespace and crd scope to create each CRD instance
func getInstanceInfo(cmd *cobra.Command, name string, crdType string, crdName string, inputNamespace string) (string, string, apiextensions.ResourceScope, error) {
	crdScope := apiextensions.ClusterScoped
	if !cmd.LocalFlags().Lookup("mock").Changed && !cmd.LocalFlags().Lookup("mock-kube").Changed {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdType)
		if err != nil {
			return "", "", "", fmt.Errorf("unable to get the %s custom resource definition in your cluster due to %+v", crdType, err)
		}
		crdScope = crd.Spec.Scope
	}

	// Check Number of Arguments
	if crdScope != apiextensions.ClusterScoped && len(namespace) == 0 {
		return "", "", "", fmt.Errorf("namespace to create an %s instance need to be provided", inputNamespace)
	}

	var namespace string
	if crdScope == apiextensions.ClusterScoped {
		if len(inputNamespace) == 0 {
			if len(crdName) > 0 {
				ns, err := util.ListNamespaces(kubeClient, fmt.Sprintf("synopsys.com.%s.%s", crdName, name))
				if err != nil {
					return "", "", "", fmt.Errorf("unable to list %s %s instance namespaces %s due to %+v", name, crdName, namespace, err)
				}
				if len(ns.Items) > 0 {
					namespace = ns.Items[0].Name
				} else {
					return "", "", "", fmt.Errorf("unable to find %s %s instance namespace", name, crdName)
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
