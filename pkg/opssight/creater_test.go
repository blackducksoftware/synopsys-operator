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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/google/go-cmp/cmp"
	"github.com/koki/short/types"
	"github.com/koki/short/util/floatstr"
	"github.com/koki/short/util/intbool"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestUpstreamPerceptor will test the upstream deployment
func TestUpstreamPerceptor(t *testing.T) {
	defaultValues := getOpsSightDefaultValue()
	opssight := NewSpecConfig(nil, defaultValues, true)

	components, err := opssight.GetComponents()

	if err != nil {
		t.Errorf("unable to get the opssight components due to %+v", err)
	}

	fmt.Printf("TODO: reenable tests for %+v with ginkgo, %+v", components, err)
	// TODO reenable these with ginkgo
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

	opssight := NewSpecConfig(nil, defaultValues, true)

	components, err := opssight.GetComponents()
	fmt.Printf("tests of %+v temporarily disabled -- reenable using ginkgo (err: %+v)", components, err)

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

func validateClusterRoleBindings(t *testing.T, clusterRoleBindings []*components.ClusterRoleBinding, opssightSpec *opssightapi.OpsSightSpec) {
	if len(clusterRoleBindings) != 3 {
		t.Errorf("cluster role binding length not equal to 3, actual: %d", len(clusterRoleBindings))
	}
	perceptorScanner := opssightSpec.ScannerPod.Scanner.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name
	perceiver := opssightSpec.Perceiver.ServiceAccount

	expectedClusterRoleBindings := map[string]*types.ClusterRoleBinding{
		perceptorScanner: {Version: "rbac.authorization.k8s.io/v1", Name: perceptorScanner, Subjects: []types.Subject{{Name: types.Name(perceptorScanner), Kind: "ServiceAccount"}}, RoleRef: types.RoleRef{Name: "cluster-admin", Kind: "ClusterRole"}},
		podPerceiver:     {Version: "rbac.authorization.k8s.io/v1", Name: podPerceiver, Subjects: []types.Subject{{Name: types.Name(perceiver), Kind: "ServiceAccount"}}, RoleRef: types.RoleRef{Name: types.Name(podPerceiver), Kind: "ClusterRole"}},
		imagePerceiver:   {Version: "rbac.authorization.k8s.io/v1", Name: imagePerceiver, Subjects: []types.Subject{{Name: types.Name(perceiver), Kind: "ServiceAccount"}}, RoleRef: types.RoleRef{Name: types.Name(imagePerceiver), Kind: "ClusterRole"}},
	}

	for _, cb := range clusterRoleBindings {
		if !cmp.Equal(cb.GetObj(), expectedClusterRoleBindings[cb.GetName()]) {
			t.Errorf("cluster role bindings is not equal for %s. Diff: %+v", cb.GetName(), cmp.Diff(cb.GetObj(), expectedClusterRoleBindings[cb.GetName()]))
		}
	}
}

func validateClusterRoles(t *testing.T, clusterRoles []*components.ClusterRole, opssightSpec *opssightapi.OpsSightSpec) {
	if len(clusterRoles) != 2 {
		t.Errorf("cluster role length not equal to 2, actual: %d", len(clusterRoles))
	}

	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name

	expectedClusterRoles := map[string]*types.ClusterRole{
		podPerceiver:   {Version: "rbac.authorization.k8s.io/v1", Name: podPerceiver, Rules: []types.PolicyRule{{Verbs: []string{"get", "watch", "list", "update"}, APIGroups: []string{"*"}, Resources: []string{"pods"}}}},
		imagePerceiver: {Version: "rbac.authorization.k8s.io/v1", Name: imagePerceiver, Rules: []types.PolicyRule{{Verbs: []string{"get", "watch", "list", "update"}, APIGroups: []string{"*"}, Resources: []string{"images"}}}},
	}

	for _, cr := range clusterRoles {
		if !cmp.Equal(cr.GetObj(), expectedClusterRoles[cr.GetName()]) {
			t.Errorf("cluster role is not equal for %s. Diff: %+v", cr.GetName(), cmp.Diff(cr.GetObj(), expectedClusterRoles[cr.GetName()]))
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
		if !strings.EqualFold(cm.GetObj().Name, actualConfigMap.name) {
			t.Errorf("config map name is not equal. Expected: %s, Actual: %s", actualConfigMap.name, cm.GetObj().Data)
		}

		if !strings.EqualFold(cm.GetName(), prometheus) {
			if _, ok := cm.GetObj().Data[actualConfigMap.fileName]; !ok {
				t.Errorf("config map file name is not equal. Expected: %s, Actual: %s", actualConfigMap.fileName, cm.GetObj().Data[actualConfigMap.fileName])
			}
		}
	}
}

func validateDeployments(t *testing.T, deployments []*components.Deployment, opssightSpec *opssightapi.OpsSightSpec) {
	if len(deployments) != 1 {
		t.Errorf("deployments length not equal to 1, actual: %d", len(deployments))
	}

	replica := int32(1)
	expectedDeployment := map[string]*types.Deployment{
		"prometheus": {
			Name:             "prometheus",
			Replicas:         &replica,
			Selector:         &types.RSSelector{Labels: map[string]string{"app": "prometheus"}},
			TemplateMetadata: &types.PodTemplateMeta{Name: "prometheus", Labels: map[string]string{"app": "prometheus"}},
			PodTemplate: types.PodTemplate{
				Volumes: map[string]types.Volume{
					"data":       {EmptyDir: &types.EmptyDirVolume{}},
					"prometheus": {ConfigMap: &types.ConfigMapVolume{Name: "prometheus", Items: map[string]types.KeyAndMode{}}},
				},
				Containers: []types.Container{
					{
						Name:                 "prometheus",
						Pull:                 types.PullAlways,
						Image:                "prom/prometheus:v2.1.0",
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						Expose:               []types.Port{{Name: "web", ContainerPort: "9090", Protocol: types.ProtocolTCP}},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: "--log.level=debug"},
							{Type: floatstr.String, StringVal: "--config.file=/etc/prometheus/prometheus.yml"},
							{Type: floatstr.String, StringVal: "--storage.tsdb.path=/tmp/data/"},
							{Type: floatstr.String, StringVal: "--storage.tsdb.retention=120d"},
						},
						VolumeMounts: []types.VolumeMount{
							{MountPath: "/data", Store: "data"},
							{MountPath: "/etc/prometheus", Store: "prometheus"},
						},
					},
				},
				RestartPolicy: types.RestartPolicyAlways,
				DNSPolicy:     types.DNSClusterFirstWithHostNet,
			},
			DeploymentStatus: types.DeploymentStatus{Replicas: types.DeploymentReplicasStatus{}},
		},
	}

	for _, d := range deployments {
		if !cmp.Equal(d.GetObj(), expectedDeployment[d.GetName()]) {
			t.Errorf("deployment is not equal for %s. Diff: %+v", d.GetName(), cmp.Diff(d.GetObj(), expectedDeployment[d.GetName()]))
		}
	}
}

func validateReplicationControllers(t *testing.T, replicationControllers []*components.ReplicationController, opssightSpec *opssightapi.OpsSightSpec) {
	if len(replicationControllers) != 4 {
		t.Errorf("replication controllers length not equal to 5, actual: %d", len(replicationControllers))
	}

	perceptor := opssightSpec.Perceptor.Name
	perceptorScanner := opssightSpec.ScannerPod.Scanner.Name
	perceptorImageFacade := opssightSpec.ScannerPod.ImageFacade.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name
	perceiver := opssightSpec.Perceiver.ServiceAccount

	replica := int32(1)
	envRequired := true
	priviledgedFalse := false
	priviledgedTrue := true
	expectedReplicationController := map[string]*types.ReplicationController{
		perceptor: {
			Name:             perceptor,
			Replicas:         &replica,
			Selector:         map[string]string{"name": perceptor},
			TemplateMetadata: &types.PodTemplateMeta{Name: perceptor, Labels: map[string]string{"name": perceptor}},
			PodTemplate: types.PodTemplate{
				Volumes: map[string]types.Volume{
					perceptor: {ConfigMap: &types.ConfigMapVolume{Name: perceptor, Items: map[string]types.KeyAndMode{}}},
				},
				Containers: []types.Container{
					{
						Command: []string{fmt.Sprintf("./%s", perceptor)},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: fmt.Sprintf("/etc/%s/%s.yaml", perceptor, perceptor)},
						},
						Env: []types.Env{
							{From: &types.EnvFrom{Key: "PCP_HUBUSERPASSWORD", From: fmt.Sprintf("secret:%s:HubUserPassword", opssightSpec.SecretName), Required: &envRequired}, Type: types.EnvFromEnvType},
						},
						Image:                "TODO -- fill in",
						Pull:                 types.PullAlways,
						CPU:                  &types.CPU{Min: "300m"},
						Mem:                  &types.Mem{Min: "1300Mi"},
						Name:                 perceptor,
						Expose:               []types.Port{{ContainerPort: "3001", Protocol: types.ProtocolTCP}},
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						VolumeMounts: []types.VolumeMount{
							{MountPath: fmt.Sprintf("/etc/%s", perceptor), Store: perceptor},
						},
					},
				},
				RestartPolicy: types.RestartPolicyAlways,
				DNSPolicy:     types.DNSClusterFirstWithHostNet,
			},
		},
		perceptorScanner: {
			Name:             perceptorScanner,
			Replicas:         &replica,
			Selector:         map[string]string{"name": perceptorScanner},
			TemplateMetadata: &types.PodTemplateMeta{Name: perceptorScanner, Labels: map[string]string{"name": perceptorScanner}},
			PodTemplate: types.PodTemplate{
				Volumes: map[string]types.Volume{
					perceptorScanner:     {ConfigMap: &types.ConfigMapVolume{Name: perceptorScanner, Items: map[string]types.KeyAndMode{}}},
					perceptorImageFacade: {ConfigMap: &types.ConfigMapVolume{Name: perceptorImageFacade, Items: map[string]types.KeyAndMode{}}},
					"var-images":         {EmptyDir: &types.EmptyDirVolume{}},
					"dir-docker-socket":  {HostPath: &types.HostPathVolume{Path: "/var/run/docker.sock"}},
				},
				Containers: []types.Container{
					{
						Command: []string{fmt.Sprintf("./%s", perceptorScanner)},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: fmt.Sprintf("/etc/%s/%s.yaml", perceptorScanner, perceptorScanner)},
						},
						Env: []types.Env{
							{From: &types.EnvFrom{Key: "PCP_HUBUSERPASSWORD", From: fmt.Sprintf("secret:%s:HubUserPassword", opssightSpec.SecretName), Required: &envRequired}, Type: types.EnvFromEnvType},
						},
						Image:                "TODO -- fill in",
						Pull:                 types.PullAlways,
						CPU:                  &types.CPU{Min: "300m"},
						Mem:                  &types.Mem{Min: "1300Mi"},
						Name:                 perceptorScanner,
						Privileged:           &priviledgedFalse,
						Expose:               []types.Port{{ContainerPort: "3003", Protocol: types.ProtocolTCP}},
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						VolumeMounts: []types.VolumeMount{
							{MountPath: fmt.Sprintf("/etc/%s", perceptorScanner), Store: perceptorScanner},
							{MountPath: "/var/images", Store: "var-images"},
						},
					},
					{
						Command: []string{fmt.Sprintf("./%s", perceptorImageFacade)},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: fmt.Sprintf("/etc/%s/%s.json", perceptorImageFacade, perceptorImageFacade)},
						},
						Image:                "TODO -- fill in",
						Pull:                 types.PullAlways,
						CPU:                  &types.CPU{Min: "300m"},
						Mem:                  &types.Mem{Min: "1300Mi"},
						Name:                 perceptorImageFacade,
						Privileged:           &priviledgedTrue,
						Expose:               []types.Port{{ContainerPort: "3004", Protocol: types.ProtocolTCP}},
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						VolumeMounts: []types.VolumeMount{
							{MountPath: fmt.Sprintf("/etc/%s", perceptorImageFacade), Store: perceptorImageFacade},
							{MountPath: "/var/images", Store: "var-images"},
							{MountPath: "/var/run/docker.sock", Store: "dir-docker-socket"},
						},
					},
				},
				RestartPolicy: types.RestartPolicyAlways,
				DNSPolicy:     types.DNSClusterFirstWithHostNet,
				Account:       perceptorScanner,
			},
		},
		podPerceiver: {
			Name:             podPerceiver,
			Replicas:         &replica,
			Selector:         map[string]string{"name": podPerceiver},
			TemplateMetadata: &types.PodTemplateMeta{Name: podPerceiver, Labels: map[string]string{"name": podPerceiver}},
			PodTemplate: types.PodTemplate{
				Volumes: map[string]types.Volume{
					perceiver: {ConfigMap: &types.ConfigMapVolume{Name: perceiver, Items: map[string]types.KeyAndMode{}}},
					"logs":    {EmptyDir: &types.EmptyDirVolume{}},
				},
				Containers: []types.Container{
					{
						Command: []string{fmt.Sprintf("./%s", podPerceiver)},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: fmt.Sprintf("/etc/%s/%s.yaml", perceiver, perceiver)},
						},
						Image:                "TODO -- fill in",
						Pull:                 types.PullAlways,
						CPU:                  &types.CPU{Min: "300m"},
						Mem:                  &types.Mem{Min: "1300Mi"},
						Name:                 podPerceiver,
						Expose:               []types.Port{{ContainerPort: "3002", Protocol: types.ProtocolTCP}},
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						VolumeMounts: []types.VolumeMount{
							{MountPath: fmt.Sprintf("/etc/%s", perceiver), Store: perceiver},
							{MountPath: "/tmp", Store: "logs"},
						},
					},
				},
				RestartPolicy: types.RestartPolicyAlways,
				DNSPolicy:     types.DNSClusterFirstWithHostNet,
				Account:       perceiver,
			},
		},
		imagePerceiver: {
			Name:             imagePerceiver,
			Replicas:         &replica,
			Selector:         map[string]string{"name": imagePerceiver},
			TemplateMetadata: &types.PodTemplateMeta{Name: imagePerceiver, Labels: map[string]string{"name": imagePerceiver}},
			PodTemplate: types.PodTemplate{
				Volumes: map[string]types.Volume{
					perceiver: {ConfigMap: &types.ConfigMapVolume{Name: perceiver, Items: map[string]types.KeyAndMode{}}},
					"logs":    {EmptyDir: &types.EmptyDirVolume{}},
				},
				Containers: []types.Container{
					{
						Command: []string{fmt.Sprintf("./%s", imagePerceiver)},
						Args: []floatstr.FloatOrString{
							{Type: floatstr.String, StringVal: fmt.Sprintf("/etc/%s/%s.yaml", perceiver, perceiver)},
						},
						Image:                "TODO -- fill in",
						Pull:                 types.PullAlways,
						CPU:                  &types.CPU{Min: "300m"},
						Mem:                  &types.Mem{Min: "1300Mi"},
						Name:                 imagePerceiver,
						Expose:               []types.Port{{ContainerPort: "3002", Protocol: types.ProtocolTCP}},
						TerminationMsgPolicy: types.TerminationMessageReadFile,
						VolumeMounts: []types.VolumeMount{
							{MountPath: fmt.Sprintf("/etc/%s", perceiver), Store: perceiver},
							{MountPath: "/tmp", Store: "logs"},
						},
					},
				},
				RestartPolicy: types.RestartPolicyAlways,
				DNSPolicy:     types.DNSClusterFirstWithHostNet,
				Account:       perceiver,
			},
		},
	}

	for _, rc := range replicationControllers {
		if !cmp.Equal(rc.GetObj(), expectedReplicationController[rc.GetName()]) {
			t.Errorf("replication controller is not equal for %s. Diff: %+v", rc.GetName(), cmp.Diff(rc.GetObj(), expectedReplicationController[rc.GetName()]))
		}
	}
}

func validateSecrets(t *testing.T, secrets []*components.Secret, opssightSpec *opssightapi.OpsSightSpec) {
	if len(secrets) != 1 {
		t.Errorf("secrets length not equal to 1, actual: %d", len(secrets))
	}

	expectedSecrets := map[string]*types.Secret{
		opssightSpec.SecretName: {Name: opssightSpec.SecretName, SecretType: types.SecretTypeOpaque},
	}

	for _, secret := range secrets {
		if !cmp.Equal(secret.GetObj(), expectedSecrets[secret.GetName()]) {
			t.Errorf("secret is not equal for %s. Diff: %+v", secret.GetName(), cmp.Diff(secret.GetObj(), expectedSecrets[secret.GetName()]))
		}
	}
}

func validateServiceAccounts(t *testing.T, serviceAccounts []*components.ServiceAccount, opssightSpec *opssightapi.OpsSightSpec) {
	if len(serviceAccounts) != 3 {
		t.Errorf("service account length not equal to 3, actual: %d", len(serviceAccounts))
	}

	perceptorScanner := opssightSpec.ScannerPod.Scanner.Name
	perceiver := opssightSpec.Perceiver.ServiceAccount

	expectedServiceAccounts := map[string]*types.ServiceAccount{
		perceptorScanner: {Name: perceptorScanner},
		perceiver:        {Name: perceiver},
	}

	for _, serviceAccount := range serviceAccounts {
		if !cmp.Equal(serviceAccount.GetObj(), expectedServiceAccounts[serviceAccount.GetName()]) {
			t.Errorf("service account is not equal for %s. Diff: %+v", serviceAccount.GetName(), cmp.Diff(serviceAccount.GetObj(), expectedServiceAccounts[serviceAccount.GetName()]))
		}
	}
}

func validateServices(t *testing.T, services []*components.Service, opssightSpec *opssightapi.OpsSightSpec) {
	if len(services) != 6 {
		t.Errorf("services length not equal to 6, actual: %d", len(services))
	}

	perceptor := opssightSpec.Perceptor.Name
	perceptorScanner := opssightSpec.ScannerPod.Scanner.Name
	perceptorImageFacade := opssightSpec.ScannerPod.ImageFacade.Name
	podPerceiver := opssightSpec.Perceiver.PodPerceiver.Name
	imagePerceiver := opssightSpec.Perceiver.ImagePerceiver.Name

	expectedServices := map[string]*types.Service{
		perceptor: {
			Name:             perceptor,
			Type:             types.ClusterIPServiceTypeDefault,
			Selector:         map[string]string{"name": perceptor},
			Port:             &types.ServicePort{Expose: int32(3001), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 3001}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
		perceptorScanner: {
			Name:             perceptorScanner,
			Type:             types.ClusterIPServiceTypeDefault,
			Selector:         map[string]string{"name": perceptorScanner},
			Port:             &types.ServicePort{Expose: int32(3003), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 3003}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
		perceptorImageFacade: {
			Name:             perceptorImageFacade,
			Type:             types.ClusterIPServiceTypeDefault,
			Selector:         map[string]string{"name": perceptorScanner},
			Port:             &types.ServicePort{Expose: int32(3004), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 3004}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
		podPerceiver: {
			Name:             podPerceiver,
			Type:             types.ClusterIPServiceTypeDefault,
			Selector:         map[string]string{"name": podPerceiver},
			Port:             &types.ServicePort{Expose: int32(3002), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 3002}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
		imagePerceiver: {
			Name:             imagePerceiver,
			Type:             types.ClusterIPServiceTypeDefault,
			Selector:         map[string]string{"name": imagePerceiver},
			Port:             &types.ServicePort{Expose: int32(3002), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 3002}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
		"prometheus": {
			Name:             "prometheus",
			Labels:           map[string]string{"name": "prometheus"},
			Annotations:      map[string]string{"prometheus.io/scrape": "true"},
			Type:             types.ClusterIPServiceTypeNodePort,
			Selector:         map[string]string{"app": "prometheus"},
			Port:             &types.ServicePort{Expose: int32(9090), Protocol: types.ProtocolTCP, PodPort: &intstr.IntOrString{Type: intstr.Int, IntVal: 9090}},
			ClientIPAffinity: &intbool.IntOrBool{Type: intbool.Bool},
		},
	}

	for _, service := range services {
		if !cmp.Equal(service.GetObj(), expectedServices[service.GetName()]) {
			t.Errorf("service is not equal for %s. Diff: %+v", service.GetName(), cmp.Diff(service.GetObj(), expectedServices[service.GetName()]))
		}
	}
}

func prettyPrintObj(components *api.ComponentList) {
	for _, cb := range components.ClusterRoleBindings {
		util.PrettyPrint(cb.GetObj())
	}

	for _, cr := range components.ClusterRoles {
		util.PrettyPrint(cr.GetObj())
	}

	for _, cm := range components.ConfigMaps {
		util.PrettyPrint(cm.GetObj())
	}

	for _, d := range components.Deployments {
		util.PrettyPrint(d.GetObj())
	}

	for _, rc := range components.ReplicationControllers {
		util.PrettyPrint(rc.GetObj())
	}

	for _, s := range components.Secrets {
		util.PrettyPrint(s.GetObj())
	}

	for _, sa := range components.ServiceAccounts {
		util.PrettyPrint(sa.GetObj())
	}

	for _, svc := range components.Services {
		util.PrettyPrint(svc.GetObj())
	}
}

// GetOpsSightDefaultValue creates a perceptor crd configuration object with defaults
func getOpsSightDefaultValue() *opssightapi.OpsSightSpec {
	return &opssightapi.OpsSightSpec{
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
			DeleteBlackDuckThresholdPercentage: 50,
			BlackduckSpec:                      nil,
		},
		EnableMetrics: true,
		EnableSkyfire: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "perceptor",
	}
}
