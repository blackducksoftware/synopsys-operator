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

import (
	"fmt"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DeployOperatorTestSuite TODO
type DeployOperatorTestSuite struct {
	Tests []TestInterface
	RC    *rest.Config
	KC    *kubernetes.Clientset
}

// NewDeployOperatorTestSuite TODO
func NewDeployOperatorTestSuite() *DeployOperatorTestSuite {
	rc, _ := GetRestConfig()
	kc, _ := GetKubeClient(rc)

	tests := []TestInterface{}
	tests = append(tests, NewDeployOperatorTest("cluster scope test", NewSynopsysctl(), rc, kc))

	return &DeployOperatorTestSuite{
		Tests: tests,
		RC:    rc,
		KC:    kc,
	}
}

// GetTests TODO
func (ts *DeployOperatorTestSuite) GetTests() []TestInterface {
	return ts.Tests
}

// Setup TODO
func (ts *DeployOperatorTestSuite) Setup() {

}

// Cleanup TODO
func (ts *DeployOperatorTestSuite) Cleanup() {

}

// DeployOperatorInClusterScopeTest TODO
type DeployOperatorInClusterScopeTest struct {
	Name            string
	TestSynopsysCtl *Synopsysctl
	RC              *rest.Config
	KC              *kubernetes.Clientset
}

// GetName TODO
func (t *DeployOperatorInClusterScopeTest) GetName() string {
	return t.Name
}

// GetTestSynopsysCtl TODO
func (t *DeployOperatorInClusterScopeTest) GetTestSynopsysCtl() *Synopsysctl {
	return t.TestSynopsysCtl
}

// NewDeployOperatorTest TODO
func NewDeployOperatorTest(name string, testSynopsysCtl *Synopsysctl, rc *rest.Config, kc *kubernetes.Clientset) *DeployOperatorInClusterScopeTest {
	return &DeployOperatorInClusterScopeTest{
		Name:            name,
		TestSynopsysCtl: testSynopsysCtl,
		RC:              rc,
		KC:              kc,
	}
}

// TestToRun TODO
func (t *DeployOperatorInClusterScopeTest) TestToRun() error {
	// deploy in cluster scope
	out, err := t.TestSynopsysCtl.Exec("deploy", "--cluster-scoped", "--enable-alert", "--enable-blackduck", "--enable-opssight", "-i=gcr.io/saas-hub-stg/blackducksoftware/synopsys-operator:release-2019.6.x")
	if err != nil {
		return fmt.Errorf("Out: %s Error: %v", out, err)
	}

	label := labels.NewSelector()
	r, _ := labels.NewRequirement("app", selection.Equals, []string{"synopsys-operator"})
	label.Add(*r)

	log.Infof("Started WaitForPodsWithLabelRunningReady")
	_, err = WaitForPodsWithLabelRunningReady(t.KC, "synopsys-operator", label, 2, time.Duration(60*time.Second))
	if err != nil {
		// failed
		return fmt.Errorf("Pods failed to come up: %v", err)
	}

	// get Crd
	apiExtensionClient, err := apiextensionsclient.NewForConfig(t.RC)
	if err != nil {
		return fmt.Errorf("error creating the api extension client due to %+v", err)
	}
	log.Infof("Started BlockUntilCrdIsAdded")
	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "alerts.synopsys.com", 10)
	if err != nil {
		return fmt.Errorf("alert crd was not added: %v", err)
	}
	log.Infof("Started BlockUntilCrdIsAdded")
	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "blackducks.synopsys.com", 10)
	if err != nil {
		return fmt.Errorf("bd crd was not added: %v", err)
	}
	log.Infof("Started BlockUntilCrdIsAdded")
	err = util.BlockUntilCrdIsAdded(apiExtensionClient, "opssights.synopsys.com", 10)
	if err != nil {
		return fmt.Errorf("ops crd was not added: %v", err)
	}
	return nil
}

// Cleanup TODO
func (t *DeployOperatorInClusterScopeTest) Cleanup() []error {
	errs := []error{}
	log.Infof("Started DeleteNamespace")
	err := util.DeleteNamespace(t.KC, "synopsys-operator")
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to delete namespace: %+v : %+v", "synopsys-operator", err))
	}
	// waitForNamespaceToBeDeleted()
	apiExtensionClient, err := apiextensionsclient.NewForConfig(t.RC)
	if err != nil {
		errs = append(errs, fmt.Errorf("error creating the api extension client: %+v", err))
	}
	log.Infof("Started DeleteCustomResourceDefinition")
	err = util.DeleteCustomResourceDefinition(apiExtensionClient, "alerts.synopsys.com")
	if err != nil {
		errs = append(errs, fmt.Errorf("error deleting alerts crd: %+v", err))
	}
	log.Infof("Started DeleteCustomResourceDefinition")
	err = util.DeleteCustomResourceDefinition(apiExtensionClient, "blackducks.synopsys.com")
	if err != nil {
		errs = append(errs, fmt.Errorf("error deleting blackducks crd: %+v", err))
	}
	log.Infof("Started DeleteCustomResourceDefinition")
	err = util.DeleteCustomResourceDefinition(apiExtensionClient, "opssights.synopsys.com")
	if err != nil {
		errs = append(errs, fmt.Errorf("error deleting opssights crd: %+v", err))
	}

	return errs
}
