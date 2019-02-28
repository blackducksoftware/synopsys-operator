// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	alertclientset "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned"
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	blackduckclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// These vars set by setResourceClients() in root command's init()
var restconfig *rest.Config
var blackduckClient *blackduckclientset.Clientset
var opssightClient *opssightclientset.Clientset
var alertClient *alertclientset.Clientset

// These vars used by KubeCmd
var openshift bool
var kube bool

// getBlackduckSpecFromCluster returns the CRD for a blackduck in namespace
func getBlackduckSpecFromCluster(namespace string) (*blackduckv1.Blackduck, error) {
	blackduck, err := blackduckClient.SynopsysV1().Blackducks(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return blackduck, fmt.Errorf("Error Editing Blackduck: %+v", err)
	}
	return blackduck, nil
}

// updateBlackduckSpecInCluster updates the CRD for a blackduck
func updateBlackduckSpecInCluster(namespace string, crd *blackduckv1.Blackduck) error {
	_, err := blackduckClient.SynopsysV1().Blackducks(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("Error Editing Blackduck: %+v", err)
	}
	return nil
}

// getOpsSightSpecFromCluster returns the CRD for an OpsSight in namespace
func getOpsSightSpecFromCluster(namespace string) (*opssightv1.OpsSight, error) {
	opssight, err := opssightClient.SynopsysV1().OpsSights(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return opssight, fmt.Errorf("Error Editing OpsSight: %+v", err)
	}
	return opssight, nil
}

// updateOpsSightSpecInCluster updates the CRD for an OpsSight
func updateOpsSightSpecInCluster(namespace string, crd *opssightv1.OpsSight) error {
	_, err := opssightClient.SynopsysV1().OpsSights(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("Error Editing OpsSight: %+v", err)
	}
	return nil
}

// getAlertSpecFromCluster returns the CRD for an Alert in namespace
func getAlertSpecFromCluster(namespace string) (*alertv1.Alert, error) {
	alert, err := alertClient.SynopsysV1().Alerts(namespace).Get(namespace, metav1.GetOptions{})
	if err != nil {
		return alert, fmt.Errorf("Error Editing Alert: %+v", err)
	}
	return alert, nil
}

// updateAlertSpecInCluster updates the CRD for an Alert
func updateAlertSpecInCluster(namespace string, crd *alertv1.Alert) error {
	_, err := alertClient.SynopsysV1().Alerts(namespace).Update(crd)
	if err != nil {
		return fmt.Errorf("Error Editing Alert: %+v", err)
	}
	return nil
}

// determineClusterClients sets bool values to true
// if it can find kube/oc in the path
func determineClusterClients() {
	_, exists := exec.LookPath("kubectl")
	if exists == nil {
		kube = true
	}
	_, ocexists := exec.LookPath("oc")
	if ocexists == nil {
		openshift = true
	}
}

// GetOperatorNamespace returns the namespace of the Synopsys-Operator by
// looking at its cluster role binding
func GetOperatorNamespace() (string, error) {
	namespace, err := RunKubeCmd("get", "clusterrolebindings", "synopsys-operator-admin", "-o", "go-template='{{range .subjects}}{{.namespace}}{{end}}'")
	if err != nil {
		return "", fmt.Errorf("%s", namespace)
	}
	return destroyNamespace, nil
}

// RunKubeCmd is a simple wrapper to oc/kubectl exec that captures output.
// TODO consider replacing w/ go api but not crucial for now.
func RunKubeCmd(args ...string) (string, error) {
	determineClusterClients()

	var cmd2 *exec.Cmd

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
	}
	if openshift {
		cmd2 = exec.Command("oc", args...)
	} else if kube {
		cmd2 = exec.Command("kubectl", args...)
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
func RunKubeEditorCmd(args ...string) error {
	determineClusterClients()

	var cmd *exec.Cmd

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
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

// setResourceClients sets the global variables for the kuberentes rest config
// and the resource clients
func setResourceClients() {
	restconfig = getKubeRestConfig()
	bClient, err := blackduckclientset.NewForConfig(restconfig)
	if err != nil {
		panic(fmt.Errorf("Error creating the Blackduck Clientset: %s", err))
	}
	blackduckClient = bClient
	oClient, err := opssightclientset.NewForConfig(restconfig)
	if err != nil {
		panic(fmt.Errorf("Error creating the OpsSight Clientset: %s", err))
	}
	opssightClient = oClient
	aClient, err := alertclientset.NewForConfig(restconfig)
	if err != nil {
		panic(fmt.Errorf("Error creating the Alert Clientset: %s", err))
	}
	alertClient = aClient
}

// getKubeRestConfig gets the user's kubeconfig from their system
func getKubeRestConfig() *rest.Config {
	log.Debugf("Getting Kube Rest Config\n")
	// Determine Config Paths
	var masterURL = ""
	var kubeconfigpath = ""
	if home := homeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfigpath = ""
	}
	// Get Rest Config using Paths
	var restconfig *rest.Config
	var err error
	if masterURL == "" && kubeconfigpath == "" {
		restconfig, err = rest.InClusterConfig() // if no paths then use in-cluster config
	} else {
		restconfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigpath,
			},
			&clientcmd.ConfigOverrides{
				ClusterInfo: clientcmdapi.Cluster{
					Server: masterURL,
				},
			}).ClientConfig()
	}
	if err != nil {
		panic(err)
	}
	return restconfig
}

// homeDir determines the user's home directory path
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
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
		return fmt.Errorf("Error deploying namespace with Horizon : %s", err)
	}
	return nil
}
