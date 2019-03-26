/*
Copyright (C) 2018 Synopsys, Inc.

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

package util

import (
	"fmt"
	"os"
	"os/exec"

	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"

	"flag"
	"path/filepath"

	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir" //_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// DetermineClusterClients returns bool values for which client
// to use. They will never both be true
func DetermineClusterClients() (kube, openshift bool) {
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
	restConfig, err := GetKubeConfig()
	if err != nil {
		log.Errorf("Error getting Kube Rest Config: %s", err)
	}
	routeClient, err := routeclient.NewForConfig(restConfig) // kube doesn't have a routeclient
	if routeClient != nil && err == nil {                    // openshift: have a routeClient and no error
		openshiftTest = true
	} else if err != nil { // Kube or Error
		openshiftTest = false
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
func RunKubeCmd(args ...string) (string, error) {
	kube, openshift := DetermineClusterClients()

	var cmd2 *exec.Cmd

	// cluster-info in kube doesnt seem to be in
	// some versions of oc, but status is.
	// double check this.
	if args[0] == "cluster-info" && openshift {
		args[0] = "status"
	}
	if openshift {
		log.Debugf("Using oc client for KubeCmd")
		cmd2 = exec.Command("oc", args...)
	} else if kube {
		log.Debugf("Using kubectl client for KubeCmd")
		cmd2 = exec.Command("kubectl", args...)
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
func RunKubeEditorCmd(args ...string) error {
	kube, openshift := DetermineClusterClients()

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

// GetKubeConfig  will return the kube config
func GetKubeConfig() (*rest.Config, error) {
	log.Debugf("Getting Kube Rest Config\n")
	var err error
	var kubeConfig *rest.Config
	// creates the in-cluster config
	kubeConfig, err = rest.InClusterConfig()
	if err != nil {
		log.Errorf("error getting in cluster config. Fallback to native config. Error message: %+v", err)
		// Determine Config Paths
		var kubeconfigpath = ""
		if home := homeDir(); home != "" {
			kubeconfigpath = filepath.Join(home, ".kube", "config")
		}

		kubeConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigpath,
			},
			&clientcmd.ConfigOverrides{
				ClusterInfo: clientcmdapi.Cluster{
					Server: "",
				},
			}).ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	return kubeConfig, nil
}

// homeDir determines the user's home directory path
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// GetKubeClientSet will return the kube clientset
func GetKubeClientSet(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(kubeConfig)
}

func newKubeClientFromOutsideCluster() (*rest.Config, error) {
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	return config, errors.Annotate(err, "error creating default client config")
}
