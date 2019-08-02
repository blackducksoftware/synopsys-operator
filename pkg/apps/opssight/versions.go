/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
)

const latestOpsSightVersion = "2.2.4"

var publicVersions = map[string]types.PublicVersion{
	"2.2.3":  getPublicVersion("2.2.3"),
	"2.2.4":  getPublicVersion("2.2.4"),
	"":       getPublicVersion(latestOpsSightVersion),
	"latest": getPublicVersion(latestOpsSightVersion),
	"upstream": {
		Size: types.OpsSightSizeV1,
		ClusterRoles: []types.ComponentName{
			types.OpsSightImageProcessorClusterRoleV1,
			types.OpsSightPodProcessorClusterRoleV1,
			types.SkyfireClusterRoleV1,
		},
		ClusterRoleBindings: []types.ComponentName{
			types.OpsSightImageProcessorClusterRoleBindingV1,
			types.OpsSightPodProcessorClusterRoleBindingV1,
			types.OpsSightScannerClusterRoleBindingV1,
			types.SkyfireClusterRoleBindingV1,
		},
		ConfigMaps: []types.ComponentName{
			types.OpsSightConfigMapV1,
			types.OpsSightMetricsConfigMapV1,
		},
		Deployments: map[string]types.PublicPodResource{
			"prometheus": {
				Identifier: types.OpsSightMetricsDeploymentV1,
				Container: map[types.ContainerName]string{
					types.OpsSightMetricsContainerName: "docker.io/prom/prometheus:v2.1.0",
				},
			},
		},
		RCs: map[string]types.PublicPodResource{
			"perceptor": {
				Identifier: types.OpsSightCoreRCV1,
				Container: map[types.ContainerName]string{
					types.PerceptorContainerName: "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
				},
			},
			"pod-perceiver": {
				Identifier: types.OpsSightPodProcessorRCV1,
				Container: map[types.ContainerName]string{
					types.PodPerceiverContainerName: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
				},
			},
			"image-perceiver": {
				Identifier: types.OpsSightImageProcessorRCV1,
				Container: map[types.ContainerName]string{
					types.ImagePerceiverContainerName: "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
				},
			},
			"perceptor-scanner": {
				Identifier: types.OpsSightScannerRCV1,
				Container: map[types.ContainerName]string{
					types.PerceptorImageFacadeContainerName: "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
					types.PerceptorScannerContainerName:     "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
				},
			},
			"skyfire": {
				Identifier: types.SkyfireRCV1,
				Container: map[types.ContainerName]string{
					types.SkyfireContainerName: "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
				},
			},
		},
		Routes: []types.ComponentName{
			types.OpsSightCoreRouteV1,
			types.OpsSightMetricsRouteV1,
		},
		Secrets: []types.ComponentName{
			types.OpsSightSecretV1,
		},
		Services: []types.ComponentName{
			types.OpsSightCoreServiceV1,
			types.OpsSightExposeCoreServiceV1,
			types.OpsSightExposeMetricsServiceV1,
			types.OpsSightImageGetterServiceV1,
			types.OpsSightImageProcessorServiceV1,
			types.OpsSightMetricsServiceV1,
			types.OpsSightPodProcessorServiceV1,
			types.OpsSightScannerServiceV1,
			types.SkyfireServiceV1,
		},
		ServiceAccounts: []types.ComponentName{
			types.OpsSightImageProcessorServiceAccountV1,
			types.OpsSightPodProcessorServiceAccountV1,
			types.OpsSightScannerServiceAccountV1,
			types.SkyfireServiceAccountV1,
		},
	},
}

func getPublicVersion(version string) types.PublicVersion {
	return types.PublicVersion{
		Size: types.OpsSightSizeV1,
		ClusterRoles: []types.ComponentName{
			types.OpsSightImageProcessorClusterRoleV1,
			types.OpsSightPodProcessorClusterRoleV1,
			types.SkyfireClusterRoleV1,
		},
		ClusterRoleBindings: []types.ComponentName{
			types.OpsSightImageProcessorClusterRoleBindingV1,
			types.OpsSightPodProcessorClusterRoleBindingV1,
			types.OpsSightScannerClusterRoleBindingV1,
			types.SkyfireClusterRoleBindingV1,
		},
		ConfigMaps: []types.ComponentName{
			types.OpsSightConfigMapV1,
			types.OpsSightMetricsConfigMapV1,
		},
		Deployments: map[string]types.PublicPodResource{
			"prometheus": {
				Identifier: types.OpsSightMetricsDeploymentV1,
				Container: map[types.ContainerName]string{
					types.OpsSightMetricsContainerName: "docker.io/prom/prometheus:v2.1.0",
				},
			},
		},
		RCs: map[string]types.PublicPodResource{
			"opssight-core": {
				Identifier: types.OpsSightCoreRCV1,
				Container: map[types.ContainerName]string{
					types.OpsSightCoreContainerName: fmt.Sprintf("docker.io/blackducksoftware/opssight-core:%s", version),
				},
			},
			"opssight-image-processor": {
				Identifier: types.OpsSightImageProcessorRCV1,
				Container: map[types.ContainerName]string{
					types.OpsSightImageProcessorContainerName: fmt.Sprintf("docker.io/blackducksoftware/opssight-image-processor:%s", version),
				},
			},
			"opssight-pod-processor": {
				Identifier: types.OpsSightPodProcessorRCV1,
				Container: map[types.ContainerName]string{
					types.OpsSightPodProcessorContainerName: fmt.Sprintf("docker.io/blackducksoftware/opssight-pod-processor:%s", version),
				},
			},
			"opssight-scanner": {
				Identifier: types.OpsSightScannerRCV1,
				Container: map[types.ContainerName]string{
					types.OpsSightImageGetterContainerName: fmt.Sprintf("docker.io/blackducksoftware/opssight-image-getter:%s", version),
					types.OpsSightScannerContainerName:     fmt.Sprintf("docker.io/blackducksoftware/opssight-scanner:%s", version),
				},
			},
			"skyfire": {
				Identifier: types.SkyfireRCV1,
				Container: map[types.ContainerName]string{
					types.SkyfireContainerName: "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
				},
			},
		},
		Routes: []types.ComponentName{
			types.OpsSightCoreRouteV1,
			types.OpsSightMetricsRouteV1,
		},
		Secrets: []types.ComponentName{
			types.OpsSightSecretV1,
		},
		Services: []types.ComponentName{
			types.OpsSightCoreServiceV1,
			types.OpsSightExposeCoreServiceV1,
			types.OpsSightExposeMetricsServiceV1,
			types.OpsSightImageGetterServiceV1,
			types.OpsSightImageProcessorServiceV1,
			types.OpsSightMetricsServiceV1,
			types.OpsSightPodProcessorServiceV1,
			types.OpsSightScannerServiceV1,
			types.SkyfireServiceV1,
		},
		ServiceAccounts: []types.ComponentName{
			types.OpsSightImageProcessorServiceAccountV1,
			types.OpsSightPodProcessorServiceAccountV1,
			types.OpsSightScannerServiceAccountV1,
			types.SkyfireServiceAccountV1,
		},
	}
}
