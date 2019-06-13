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
	"fmt"
	"strings"
	"testing"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/google/go-cmp/cmp"
)

// TestUpstreamPerceptor will test the upstream deployment
func TestUpstreamPerceptor(t *testing.T) {
	defaultValues := getOpsSightDefaultValue()
	opssight := NewSpecConfig(&protoform.Config{DryRun: true}, nil, nil, nil, defaultValues, true)

	// TODO "components, err := opssight.GetComponents()" [add components back with ginko tests]
	_, err := opssight.GetComponents()
	if err != nil {
		t.Errorf("unable to get the opssight components due to %+v", err)
	}

	fmt.Printf("TODO: reenable tests with ginkgo\n")
	// TODO reenable these with ginkgo [add components back above]
	// validateClusterRoleBindings(t, components.ClusterRoleBindings, defaultValues)
	// validateClusterRoles(t, components.ClusterRoles, defaultValues)
	// validateConfigMaps(t, components.ConfigMaps, defaultValues)
	// validateDeployments(t, components.Deployments, defaultValues)
	// validateReplicationControllers(t, components.ReplicationControllers, defaultValues)
	// validateSecrets(t, components.Secrets, defaultValues)
	// validateServiceAccounts(t, components.ServiceAccounts, defaultValues)
	// validateServices(t, components.Services, defaultValues)
}

// TestDownstreamPerceptor will test the downstream deployment
func TestDownstreamPerceptor(t *testing.T) {
	defaultValues := getOpsSightDefaultValue()

	opssight := NewSpecConfig(&protoform.Config{DryRun: true}, nil, nil, nil, defaultValues, true)

	_, err := opssight.GetComponents() // TODO add components back ", err := opssight.GetComponents()"
	fmt.Printf("TODO tests DownStreamPerceptor Components are temporarily disabled -- reenable using ginkgo\n")

	if err != nil {
		t.Errorf("unable to get the opssight components due to %+v", err)
	}

	// TODO convert to Ginkgo
	// validateClusterRoleBindings(t, components.ClusterRoleBindings, defaultValues)
	// validateClusterRoles(t, components.ClusterRoles, defaultValues)
	// validateConfigMaps(t, components.ConfigMaps, defaultValues)
	// validateDeployments(t, components.Deployments, defaultValues)
	// validateReplicationControllers(t, components.ReplicationControllers, defaultValues)
	// validateSecrets(t, components.Secrets, defaultValues)
	// validateServiceAccounts(t, components.ServiceAccounts, defaultValues)
	// validateServices(t, components.Services, defaultValues)
}

func validateClusterRoleBindings(t *testing.T, clusterRoleBindings []*components.ClusterRoleBinding, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(clusterRoleBindings) != 3 {
		t.Errorf("cluster role binding length not equal to 3, actual: %d", len(clusterRoleBindings))
	}
	scanner := opssightSpec.ScannerPod.Scanner.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name

	scannerClusterRoleBinding, _ := opssightSpecConfig.ScannerClusterRoleBinding()
	expectedClusterRoleBindings := map[string]*components.ClusterRoleBinding{
		scanner:        scannerClusterRoleBinding,
		podPerceiver:   opssightSpecConfig.PodPerceiverClusterRoleBinding(opssightSpecConfig.PodPerceiverClusterRole()),
		imagePerceiver: opssightSpecConfig.ImagePerceiverClusterRoleBinding(opssightSpecConfig.ImagePerceiverClusterRole()),
	}

	for _, cb := range clusterRoleBindings {
		if !cmp.Equal(cb.ClusterRoleBinding, expectedClusterRoleBindings[cb.GetName()]) {
			t.Errorf("cluster role bindings is not equal for %s. Diff: %+v", cb.GetName(), cmp.Diff(cb.ClusterRoleBinding, expectedClusterRoleBindings[cb.GetName()]))
		}
	}
}

func validateClusterRoles(t *testing.T, clusterRoles []*components.ClusterRole, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(clusterRoles) != 2 {
		t.Errorf("cluster role length not equal to 2, actual: %d", len(clusterRoles))
	}

	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name
	expectedClusterRoles := map[string]*components.ClusterRole{
		podPerceiver:   opssightSpecConfig.PodPerceiverClusterRole(),
		imagePerceiver: opssightSpecConfig.ImagePerceiverClusterRole(),
	}

	for _, cr := range clusterRoles {
		if !cmp.Equal(cr.ClusterRole, expectedClusterRoles[cr.GetName()]) {
			t.Errorf("cluster role is not equal for %s. Diff: %+v", cr.GetName(), cmp.Diff(cr.ClusterRole, expectedClusterRoles[cr.GetName()]))
		}
	}
}

func validateConfigMaps(t *testing.T, configMaps []*components.ConfigMap, opssightSpec *opssightapi.OpsSightSpec) {
	if len(configMaps) != 6 {
		t.Errorf("config maps length not equal to 6, actual: %d", len(configMaps))
	}

	perceptor := opssightSpec.Perceptor.Name
	perceptorScanner := opssightSpec.ScannerPod.Scanner.Name
	perceptorImageFacade := opssightSpec.ScannerPod.ImageFacade.Name
	perceiver := opssightSpec.Perceiver.ServiceAccount
	prometheus := "prometheus"

	type configMap struct {
		name     string
		fileName string
	}

	expectedConfigMaps := map[string]*configMap{
		perceptor:            {name: perceptor, fileName: fmt.Sprintf("%s.yaml", perceptor)},
		perceptorScanner:     {name: perceptorScanner, fileName: fmt.Sprintf("%s.yaml", perceptorScanner)},
		perceptorImageFacade: {name: perceptorImageFacade, fileName: fmt.Sprintf("%s.json", perceptorImageFacade)},
		perceiver:            {name: perceiver, fileName: fmt.Sprintf("%s.yaml", perceiver)},
		prometheus:           {name: prometheus},
	}

	for _, cm := range configMaps {
		actualConfigMap := expectedConfigMaps[cm.GetName()]
		if !strings.EqualFold(cm.Name, actualConfigMap.name) {
			t.Errorf("config map name is not equal. Expected: %s, Actual: %s", actualConfigMap.name, cm.Data)
		}

		if !strings.EqualFold(cm.GetName(), prometheus) {
			if _, ok := cm.Data[actualConfigMap.fileName]; !ok {
				t.Errorf("config map file name is not equal. Expected: %s, Actual: %s", actualConfigMap.fileName, cm.Data[actualConfigMap.fileName])
			}
		}
	}
}

func validateDeployments(t *testing.T, deployments []*components.Deployment, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(deployments) != 1 {
		t.Errorf("deployments length not equal to 1, actual: %d", len(deployments))
	}

	prometheusDeployment, _ := opssightSpecConfig.PerceptorMetricsDeployment()

	expectedDeployment := map[string]*components.Deployment{
		"prometheus": prometheusDeployment,
	}

	for _, d := range deployments {
		if !cmp.Equal(d.Deployment, expectedDeployment[d.GetName()]) {
			t.Errorf("deployment is not equal for %s. Diff: %+v", d.GetName(), cmp.Diff(d.Deployment, expectedDeployment[d.GetName()]))
		}
	}
}

func validateReplicationControllers(t *testing.T, replicationControllers []*components.ReplicationController, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(replicationControllers) != 4 {
		t.Errorf("replication controllers length not equal to 5, actual: %d", len(replicationControllers))
	}

	perceptor := opssightSpec.Perceptor.Name
	scanner := opssightSpec.ScannerPod.Scanner.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name

	perceptorRc, _ := opssightSpecConfig.PerceptorReplicationController()
	scannerRc, _ := opssightSpecConfig.ScannerReplicationController()
	podPerceiverRc, _ := opssightSpecConfig.PodPerceiverReplicationController()
	imagePerceiverRc, _ := opssightSpecConfig.ImagePerceiverReplicationController()
	expectedReplicationController := map[string]*components.ReplicationController{
		perceptor:      perceptorRc,
		scanner:        scannerRc,
		podPerceiver:   podPerceiverRc,
		imagePerceiver: imagePerceiverRc,
	}

	for _, rc := range replicationControllers {
		if !cmp.Equal(rc.ReplicationController, expectedReplicationController[rc.GetName()]) {
			t.Errorf("replication controller is not equal for %s. Diff: %+v", rc.GetName(), cmp.Diff(rc.ReplicationController, expectedReplicationController[rc.GetName()]))
		}
	}
}

func validateSecrets(t *testing.T, secrets []*components.Secret, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(secrets) != 1 {
		t.Errorf("secrets length not equal to 1, actual: %d", len(secrets))
	}

	perceptorSecret, err := opssightSpecConfig.PerceptorSecret()
	t.Errorf("invalid secret: %s", err)
	expectedSecrets := map[string]*components.Secret{
		opssightSpec.SecretName: perceptorSecret,
	}

	for _, secret := range secrets {
		if !cmp.Equal(secret.Secret, expectedSecrets[secret.GetName()]) {
			t.Errorf("secret is not equal for %s. Diff: %+v", secret.GetName(), cmp.Diff(secret.Secret, expectedSecrets[secret.GetName()]))
		}
	}
}

func validateServiceAccounts(t *testing.T, serviceAccounts []*components.ServiceAccount, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(serviceAccounts) != 3 {
		t.Errorf("service account length not equal to 3, actual: %d", len(serviceAccounts))
	}

	scanner := opssightSpec.ScannerPod.Scanner.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name
	expectedServiceAccounts := map[string]*components.ServiceAccount{
		scanner:        opssightSpecConfig.ScannerServiceAccount(),
		imagePerceiver: opssightSpecConfig.ImagePerceiverServiceAccount(),
		podPerceiver:   opssightSpecConfig.PodPerceiverServiceAccount(),
	}

	for _, serviceAccount := range serviceAccounts {
		if !cmp.Equal(serviceAccount.ServiceAccount, expectedServiceAccounts[serviceAccount.GetName()]) {
			t.Errorf("service account is not equal for %s. Diff: %+v", serviceAccount.GetName(), cmp.Diff(serviceAccount.ServiceAccount, expectedServiceAccounts[serviceAccount.GetName()]))
		}
	}
}

func validateServices(t *testing.T, services []*components.Service, opssightSpec *opssightapi.OpsSightSpec, opssightSpecConfig *SpecConfig) {
	if len(services) != 6 {
		t.Errorf("services length not equal to 6, actual: %d", len(services))
	}

	// perceptor
	perceptor := opssightSpec.Perceptor.Name
	perceptorService, _ := opssightSpecConfig.PerceptorService()
	scanner := opssightSpec.ScannerPod.Scanner.Name
	imageFacade := opssightSpec.ScannerPod.ImageFacade.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name

	// prometheus
	prometheusService, _ := opssightSpecConfig.PerceptorMetricsService()

	expectedServices := map[string]*components.Service{
		perceptor:      perceptorService,
		scanner:        opssightSpecConfig.ScannerService(),
		imageFacade:    opssightSpecConfig.ImageFacadeService(),
		podPerceiver:   opssightSpecConfig.PodPerceiverService(),
		imagePerceiver: opssightSpecConfig.ImageFacadeService(),
		"prometheus":   prometheusService,
	}

	for _, service := range services {
		if !cmp.Equal(service.Service, expectedServices[service.GetName()]) {
			t.Errorf("service is not equal for %s. Diff: %+v", service.GetName(), cmp.Diff(service.Service, expectedServices[service.GetName()]))
		}
	}
}

func prettyPrintObj(components *api.ComponentList) {
	for _, cb := range components.ClusterRoleBindings {
		util.PrettyPrint(cb.ClusterRoleBinding)
	}

	for _, cr := range components.ClusterRoles {
		util.PrettyPrint(cr.ClusterRole)
	}

	for _, cm := range components.ConfigMaps {
		util.PrettyPrint(cm.ConfigMap)
	}

	for _, d := range components.Deployments {
		util.PrettyPrint(d.Deployment)
	}

	for _, rc := range components.ReplicationControllers {
		util.PrettyPrint(rc.ReplicationController)
	}

	for _, s := range components.Secrets {
		util.PrettyPrint(s.Secret)
	}

	for _, sa := range components.ServiceAccounts {
		util.PrettyPrint(sa.ServiceAccount)
	}

	for _, svc := range components.Services {
		util.PrettyPrint(svc.Service)
	}
}

// GetOpsSightDefaultValue creates a perceptor crd configuration object with defaults
func getOpsSightDefaultValue() *opssightapi.OpsSight {
	return &opssightapi.OpsSight{
		Spec: opssightapi.OpsSightSpec{
			Perceptor: &opssightapi.Perceptor{
				Name:                           "perceptor",
				Port:                           3001,
				Image:                          "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
				CheckForStalledScansPauseHours: 999999,
				StalledScanClientTimeoutHours:  999999,
				ModelMetricsPauseSeconds:       15,
				UnknownImagePauseMilliseconds:  15000,
				ClientTimeoutMilliseconds:      100000,
			},
			Perceiver: &opssightapi.Perceiver{
				EnableImagePerceiver: false,
				EnablePodPerceiver:   true,
				Port:                 3002,
				ImagePerceiver: &opssightapi.ImagePerceiver{
					Name:  "image-perceiver",
					Image: "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
				},
				PodPerceiver: &opssightapi.PodPerceiver{
					Name:  "pod-perceiver",
					Image: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
				},
				ServiceAccount:            "perceiver",
				AnnotationIntervalSeconds: 30,
				DumpIntervalMinutes:       30,
			},
			ScannerPod: &opssightapi.ScannerPod{
				ImageFacade: &opssightapi.ImageFacade{
					Port:               3004,
					InternalRegistries: []*opssightapi.RegistryAuth{},
					Image:              "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
					ServiceAccount:     "perceptor-scanner",
					Name:               "perceptor-imagefacade",
				},
				Scanner: &opssightapi.Scanner{
					Name:                 "perceptor-scanner",
					Port:                 3003,
					Image:                "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
					ClientTimeoutSeconds: 600,
				},
				ReplicaCount: 1,
			},
			Prometheus: &opssightapi.Prometheus{
				Name:  "prometheus",
				Image: "docker.io/prom/prometheus:v2.1.0",
				Port:  9090,
			},
			Skyfire: &opssightapi.Skyfire{
				Image:          "gcr.io/saas-hub-stg/blackducksoftware/skyfire:master",
				Name:           "skyfire",
				Port:           3005,
				ServiceAccount: "skyfire",
			},
			Blackduck: &opssightapi.Blackduck{
				InitialCount:                       1,
				MaxCount:                           1,
				DeleteBlackduckThresholdPercentage: 50,
				BlackduckSpec:                      nil,
			},
			EnableMetrics: true,
			EnableSkyfire: false,
			DefaultCPU:    "300m",
			DefaultMem:    "1300Mi",
			LogLevel:      "debug",
			SecretName:    "perceptor",
		},
	}
}
