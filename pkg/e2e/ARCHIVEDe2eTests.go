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

package e2e

// import (
// 	"fmt"
// 	"time"

// 	log "github.com/sirupsen/logrus"

// 	"github.com/blackducksoftware/synopsys-operator/pkg/util"
// 	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/apimachinery/pkg/selection"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/rest"
// )

// // SynopsysOperator TODO
// type SynopsysOperator struct {
// 	namespace   string
// 	synopsysctl *Synopsysctl
// }

// func NewSynopsysOperator() *SynopsysOperator {
// 	sc := NewSynopsysctl()
// 	// "app=synopsys-operator"
// 	label := labels.NewSelector()
// 	r, _ := labels.NewRequirement("app", selection.Equals, []string{"synopsys-operator"})
// 	label.Add(*r)
// 	return &SynopsysOperator{
// 		labels:      label,
// 		namespace:   "synopsys-operator",
// 		synopsysctl: sc,
// 	}
// }

// // Deploy TODO
// func (so *SynopsysOperator) Deploy(args ...string) (string, error) {
// 	log.Printf("[Deploy] args: %v \n\n", args)
// 	args = append([]string{"deploy"}, args...)
// 	return so.synopsysctl.Exec(args...)
// }

// // Alert TODO
// type Alert struct {
// 	label string
// }

// func newAlert() *Alert {
// 	return &Alert{label: "app-alert"}
// }

// func getRestConfig() (*rest.Config, error) {
// 	kubeconfig := ""
// 	restconfig, err := GetKubeConfig(kubeconfig, false)
// 	log.Debugf("rest config: %+v", restconfig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("[getRestConfig] restconfig: %v \n\n", restconfig)
// 	return restconfig, nil
// }

// // getKubeClient gets the kubernetes client
// func getKubeClient(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
// 	client, err := kubernetes.NewForConfig(kubeConfig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("[getKubeClient] client: %v \n\n", client)
// 	return client, nil
// }

// // MockTest TODO
// func MockTest() error {
// 	rc, err := getRestConfig()
// 	if err != nil {
// 		return err
// 	}
// 	kc, err := getKubeClient(rc)
// 	if err != nil {
// 		return err
// 	}

// 	so := newSynopsysOperator()
// 	so.Deploy("--cluster-scoped", "--enable-alert", "--enable-blackduck", "--enable-opssight", "-i=docker.io/black:fail")

// 	_, err = WaitForPodsWithLabelRunningReady(kc, so.namespace, so.labels, 2, time.Duration(5*time.Second))
// 	if err != nil {
// 		// failed
// 		fmt.Println("YOU FAILED")
// 		cleanup()
// 		return err
// 	}
// 	// get Crd
// 	apiExtensionClient, err := apiextensionsclient.NewForConfig(rc)
// 	if err != nil {
// 		return fmt.Errorf("error creating the api extension client due to %+v", err)
// 	}
// 	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "alerts.synopsys.com", 10)
// 	if err != nil {
// 		fmt.Printf("alert crd was not added: %v", err)
// 		err = cleanup()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "blackducks.synopsys.com", 10)
// 	if err != nil {
// 		fmt.Printf("bd crd was not added: %v", err)
// 		err = cleanup()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "opssights.synopsys.com", 10)
// 	if err != nil {
// 		fmt.Printf("ops crd was not added: %v", err)
// 		err = cleanup()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	// CLEANUP
// 	err = cleanup()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func cleanup() error {
// 	rc, err := getRestConfig()
// 	if err != nil {
// 		return err
// 	}
// 	kc, err := getKubeClient(rc)
// 	if err != nil {
// 		return err
// 	}
// 	so := newSynopsysOperator()

// 	util.DeleteNamespace(kc, so.namespace)
// 	// waitForNamespaceToBeDeleted()
// 	apiExtensionClient, err := apiextensionsclient.NewForConfig(rc)
// 	if err != nil {
// 		return fmt.Errorf("error creating the api extension client due to %+v", err)
// 	}
// 	util.DeleteCustomResourceDefinition(apiExtensionClient, "alerts.synopsys.com")
// 	util.DeleteCustomResourceDefinition(apiExtensionClient, "blackducks.synopsys.com")
// 	util.DeleteCustomResourceDefinition(apiExtensionClient, "opssights.synopsys.com")
// 	return nil
// }
