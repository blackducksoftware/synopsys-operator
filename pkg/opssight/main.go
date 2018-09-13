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

package opssight

import (
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/imdario/mergo"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Creater will store the configuration to create OpsSight
type Creater struct {
	kubeConfig     *rest.Config
	kubeClient     *kubernetes.Clientset
	opssightClient *opssightclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset) *Creater {
	return &Creater{kubeConfig: kubeConfig, kubeClient: kubeClient, opssightClient: opssightClient}
}

// NewAppDefaults creates a perceptor app configuration object
// with defaults
func NewAppDefaults() *v1.OpsSightSpec {
	defaultPerceptorPort := 3001
	defaultPerceiverPort := 3002
	defaultScannerPort := 3003
	defaultIFPort := 3004
	defaultSkyfirePort := 3005
	defaultAnnotationInterval := 30
	defaultDumpInterval := 30
	defaultHubPort := 443
	defaultPerceptorHubClientTimeout := 100000
	defaultScannerHubClientTimeout := 600
	defaultScanLimit := 7
	defaultTotalScanLimit := 1000
	defaultCheckForStalledScansPauseHours := 999999
	defaultStalledScanClientTimeoutHours := 999999
	defaultModelMetricsPauseSeconds := 15
	defaultUnknownImagePauseMilliseconds := 15000
	defaultPodPerceiverEnabled := true
	defaultImagePerceiverEnabled := false
	defaultMetricsEnabled := false
	defaultPerceptorSkyfire := false
	defaultUseMockMode := false

	return &v1.OpsSightSpec{
		PerceptorPort:             &defaultPerceptorPort,
		PerceiverPort:             &defaultPerceiverPort,
		ScannerPort:               &defaultScannerPort,
		ImageFacadePort:           &defaultIFPort,
		SkyfirePort:               &defaultSkyfirePort,
		InternalRegistries:        []v1.RegistryAuth{},
		AnnotationIntervalSeconds: &defaultAnnotationInterval,
		DumpIntervalMinutes:       &defaultDumpInterval,
		HubUser:                   "sysadmin",
		HubPort:                   &defaultHubPort,
		HubClientTimeoutPerceptorMilliseconds: &defaultPerceptorHubClientTimeout,
		HubClientTimeoutScannerSeconds:        &defaultScannerHubClientTimeout,
		ConcurrentScanLimit:                   &defaultScanLimit,
		TotalScanLimit:                        &defaultTotalScanLimit,
		CheckForStalledScansPauseHours:        &defaultCheckForStalledScansPauseHours,
		StalledScanClientTimeoutHours:         &defaultStalledScanClientTimeoutHours,
		ModelMetricsPauseSeconds:              &defaultModelMetricsPauseSeconds,
		UnknownImagePauseMilliseconds:         &defaultUnknownImagePauseMilliseconds,
		DefaultVersion:                        "master",
		Registry:                              "docker.io",
		ImagePath:                             "blackducksoftware",
		PerceptorImageName:                    "opssight-core",
		ScannerImageName:                      "opssight-scanner",
		ImagePerceiverImageName:               "opssight-image-processor",
		PodPerceiverImageName:                 "opssight-pod-processor",
		ImageFacadeImageName:                  "opssight-image-getter",
		SkyfireImageName:                      "skyfire",
		PodPerceiver:                          &defaultPodPerceiverEnabled,
		ImagePerceiver:                        &defaultImagePerceiverEnabled,
		Metrics:                               &defaultMetricsEnabled,
		PerceptorSkyfire:                      &defaultPerceptorSkyfire,
		DefaultCPU:                            "300m",
		DefaultMem:                            "1300Mi",
		LogLevel:                              "debug",
		HubUserPasswordEnvVar:                 "PCP_HUBUSERPASSWORD",
		SecretName:                            "perceptor",
		UseMockMode:                           &defaultUseMockMode,
	}
}

// DeleteOpsSight will delete the Black Duck OpsSight
func (ac *Creater) DeleteOpsSight(namespace string) {
	log.Debugf("Delete OpsSight details for %s", namespace)
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(ac.kubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(ac.kubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(ac.kubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(createOpsSight *v1.OpsSight) error {
	log.Debugf("Create OpsSight details for %s: %+v", createOpsSight.Spec.Namespace, createOpsSight)
	newSpec := createOpsSight.Spec
	opssightSpec := NewAppDefaults()
	err := mergo.Merge(&newSpec, opssightSpec)
	if err != nil {
		log.Errorf("unable to merge the opssight structs for %s due to %+v", createOpsSight.Name, err)
		return err
	}

	opssight := NewOpsSight(&newSpec)
	components, err := opssight.GetComponents()
	if err != nil {
		log.Errorf("unable to get opssight components for %s due to %+v", createOpsSight.Name, err)
		return err
	}
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		log.Errorf("unable to get deployer object for %s due to %+v", createOpsSight.Name, err)
		return err
	}
	// Note: controllers that need to continually run to update your app
	// should be added in PreDeploy().
	deployer.PreDeploy(components, createOpsSight.Name)

	// Any new, pluggable maintainance stuff should go in here...
	// deployer.AddController("perceptor_configmap_controller", &plugins.PerceptorConfigMap{})

	err = deployer.Run()

	if err != nil {
		log.Errorf("unable to deploy opssight app due to %+v", err)
	}
	deployer.StartControllers()
	return nil
}
