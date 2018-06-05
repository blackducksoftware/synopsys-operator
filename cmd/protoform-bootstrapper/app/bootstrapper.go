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

package app

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/perceptor-protoform/cmd/protoform-bootstrapper/app/options"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"
	"github.com/koki/short/util/floatstr"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Bootstrapper defines how to bootstrap protoform
type Bootstrapper struct {
	opts   *options.BootstrapperOptions
	client *kubernetes.Clientset
}

// NewBootstrapper creats a new Bootstrapper object
func NewBootstrapper(opts *options.BootstrapperOptions) (*Bootstrapper, error) {
	config, err := clientcmd.BuildConfigFromFlags("", opts.ClusterConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %v", err)
	}
	return &Bootstrapper{opts: opts, client: c}, nil
}

// Run starts the bootstrapping process
func (b *Bootstrapper) Run() error {
	secretEnv, err := types.NewEnvFromSecret("PCP_HUBUSERPASSWORD", "protoform", "HubUserPassword")
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	// Create the protoform pod
	pod := types.Pod{
		Version: "v1",
		PodTemplateMeta: types.PodTemplateMeta{
			Name: "protoform",
		},
		PodTemplate: types.PodTemplate{
			RestartPolicy: types.RestartPolicyNever,
			Account:       "protoform",
			Volumes: map[string]types.Volume{
				"protoform": {
					ConfigMap: &types.ConfigMapVolume{
						Name: "protoform",
					},
				},
			},
			Containers: []types.Container{
				{
					Name:  "protoform",
					Image: fmt.Sprintf("%s:%s", b.generateImageName(), b.opts.ProtoformImageVersion),
					Env: []types.Env{
						secretEnv,
					},
					Pull:    types.PullAlways,
					Command: []string{"./protoform"},
					Args: []floatstr.FloatOrString{
						{
							Type:      floatstr.String,
							StringVal: "/etc/protoform/protoform.yaml",
						},
					},
					Expose: []types.Port{
						{
							ContainerPort: "3001", // TODO: Use config
							Protocol:      types.ProtocolTCP,
						},
					},
					VolumeMounts: []types.VolumeMount{
						{
							Store:     "protoform",
							MountPath: "/etc/protoform",
						},
					},
				},
			},
		},
	}

	// Create the secret
	secret := types.Secret{
		Version:    "v1",
		Name:       "protoform",
		Data:       map[string][]byte{"HubUserPassword": []byte(b.opts.HubUserPassword)},
		SecretType: types.SecretTypeOpaque,
	}

	// Create protoform's configuration
	protoformConfig := []string{fmt.Sprintf("%s: %s", "DockerPasswordOrToken", b.opts.DockerPasswordOrToken),
		fmt.Sprintf("%s: %s", "HubHost", b.opts.HubHost),
		fmt.Sprintf("%s: %s", "HubUser", b.opts.HubUser),
		fmt.Sprintf("%s: %d", "HubPort", b.opts.HubPort),
		fmt.Sprintf("%s: %d", "HubClientTimeoutPerceptorMilliseconds", b.opts.HubClientTimeoutPerceptorMilliseconds),
		fmt.Sprintf("%s: %d", "HubClientTimeoutScannerSeconds", b.opts.HubClientTimeoutScannerSeconds),
		fmt.Sprintf("%s: %d", "ConcurrentScanLimit", b.opts.ConcurrentScanLimit),
		fmt.Sprintf("%s: %s", "DockerUsername", "admin"),
		fmt.Sprintf("%s: %s", "Namespace", b.opts.Namespace),
		fmt.Sprintf("%s: %t", "ImagePerceiver", b.opts.AnnotateImages),
		fmt.Sprintf("%s: %t", "PodPerceiver", b.opts.AnnotatePods),
		fmt.Sprintf("%s: %v", "InternalDockerRegistries", b.opts.InternalDockerRegistries),
		fmt.Sprintf("%s: %s", "DefaultCPU", b.opts.DefaultCPU),
		fmt.Sprintf("%s: %s", "DefaultMem", b.opts.DefaultMem),
		fmt.Sprintf("%s: %s", "Registry", b.opts.DefaultRegistry),
		fmt.Sprintf("%s: %s", "ImagePath", b.opts.DefaultImagePath),
		fmt.Sprintf("%s: %s", "Defaultversion", b.opts.DefaultImageVersion),
		fmt.Sprintf("%s: %s", "PerceptorImageName", b.opts.PerceptorImage),
		fmt.Sprintf("%s: %s", "ScannerImageName", b.opts.ScannerImage),
		fmt.Sprintf("%s: %s", "PodPerceiverImageName", b.opts.PodPerceiverImage),
		fmt.Sprintf("%s: %s", "ImagePerceiverImageName", b.opts.ImagePerceiverImage),
		fmt.Sprintf("%s: %s", "ImageFacadeImageName", b.opts.ImageFacadeImage),
		fmt.Sprintf("%s: %s", "SkyfireImageName", b.opts.SkyfireImage),
		fmt.Sprintf("%s: %s", "PerceptorImageVersion", b.opts.PerceptorImageVersion),
		fmt.Sprintf("%s: %s", "ScannerImageVersion", b.opts.ScannerImageVersion),
		fmt.Sprintf("%s: %s", "PerceiverImageVersion", b.opts.PerceiverImageVersion),
		fmt.Sprintf("%s: %s", "ImageFacadeImageVersion", b.opts.ImageFacadeImageVersion),
		fmt.Sprintf("%s: %s", "SkyfireImageVersion", b.opts.SkyfireImageVersion),
		fmt.Sprintf("%s: %s", "LogLevel", b.opts.LogLevel),
		fmt.Sprintf("%s: %t", "Metrics", b.opts.EnableMetrics),
		fmt.Sprintf("%s: %s", "Namespace", b.opts.Namespace),
		fmt.Sprintf("%s: %s", "ViperSecret", "protoform"),
		fmt.Sprintf("%s: %t", "PerceptorSkyfire", b.opts.EnableSkyfire),
		fmt.Sprintf("%s: %s", "LogLevel", b.opts.LogLevel),
	}

	configMap := types.ConfigMap{
		Version: "v1",
		Name:    "protoform",
		Data:    map[string]string{"protoform.yaml": strings.Join(protoformConfig, "\n")},
	}

	b.createDeps()

	// Deploy the configmap to the cluster
	cmWrapper := &types.ConfigMapWrapper{ConfigMap: configMap}
	config, err := converters.Convert_Koki_ConfigMap_to_Kube_v1_ConfigMap(cmWrapper)
	if err != nil {
		return fmt.Errorf("failed to convert configmap: %v", err)
	}
	_, err = b.client.Core().ConfigMaps(b.opts.Namespace).Create(config)
	if err != nil {
		return fmt.Errorf("failed to submit configmap: %v", err)
	}

	// Deploy the secret to the cluster
	sWrapper := &types.SecretWrapper{Secret: secret}
	s, err := converters.Convert_Koki_Secret_to_Kube_v1_Secret(sWrapper)
	if err != nil {
		return fmt.Errorf("failed to convert secret: %v", err)
	}
	_, err = b.client.Core().Secrets(b.opts.Namespace).Create(s)
	if err != nil {
		return fmt.Errorf("failed to submit secret: %v", err)
	}

	// Finally deploy the pod to the cluster
	podWrapper := &types.PodWrapper{Pod: pod}
	protoformPod, err := converters.Convert_Koki_Pod_to_Kube_v1_Pod(podWrapper)
	if err != nil {
		return fmt.Errorf("failed to convert pod: %v", err)
	}
	_, err = b.client.Core().Pods(b.opts.Namespace).Create(protoformPod)
	if err != nil {
		return fmt.Errorf("failed to submit pod: %v", err)
	}

	return nil
}

func (b *Bootstrapper) createDeps() error {
	protoformServiceAccount := types.ServiceAccount{
		Name:      "protoform",
		Namespace: b.opts.Namespace,
	}

	protoformNameSpace := types.Namespace{
		Name: b.opts.Namespace,
	}

	protoformSARoleBinding := types.ClusterRoleBinding{
		Name:    "protoform",
		Version: "rbac.authorization.k8s.io/v1",
		Subjects: []types.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "protoform",
				Namespace: b.opts.Namespace,
			},
		},
		RoleRef: types.RoleRef{
			APIGroup: "",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	nsWrapper := &types.NamespaceWrapper{Namespace: protoformNameSpace}
	nsObj, err := converters.Convert_Koki_Namespace_to_Kube_Namespace(nsWrapper)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}

	b.client.Core().Namespaces().Create(nsObj)
	/* TODO: Print a warning?
	if err != nil {

	}
	*/

	saWrapper := &types.ServiceAccountWrapper{ServiceAccount: protoformServiceAccount}
	saObj, err := converters.Convert_Koki_ServiceAccount_to_Kube_ServiceAccount(saWrapper)
	if err != nil {
		return fmt.Errorf("failed to convert protoform service account: %v", err)
	}
	_, err = b.client.Core().ServiceAccounts(b.opts.Namespace).Create(saObj)
	if err != nil {
		return fmt.Errorf("failed to create protoform service account: %v", err)
	}

	rbWrapper := &types.ClusterRoleBindingWrapper{ClusterRoleBinding: protoformSARoleBinding}
	rbObject, err := converters.Convert_Koki_ClusterRoleBinding_to_Kube(rbWrapper)
	if err != nil {
		return fmt.Errorf("failed to convert protoform cluster role binding: %v", err)
	}
	_, err = b.client.Rbac().ClusterRoleBindings().Create(rbObject)
	if err != nil {
		return fmt.Errorf("failed to create protoform cluster role binding: %v", err)
	}
	return nil
}

func (b *Bootstrapper) generateImageName() string {
	registry := b.opts.DefaultRegistry
	path := b.opts.DefaultImagePath
	image := b.opts.ProtoformImage
	registryIndex := strings.Index(b.opts.ProtoformImage, "/")
	imageIndex := strings.LastIndex(b.opts.ProtoformImage, "/")

	// Check if a registry was provided
	if registryIndex >= 0 {
		substr := b.opts.ProtoformImage[0:registryIndex]
		if strings.Contains(substr, ".") || strings.Contains(substr, ":") {
			// This is a registry
			registry = substr
			if registryIndex != imageIndex {
				path = b.opts.ProtoformImage[registryIndex+1 : imageIndex]
			} else {
				path = ""
			}
		} else {
			registry = b.opts.DefaultRegistry
			path = b.opts.ProtoformImage[0:imageIndex]
		}
	}

	// Find the image name if there is a slash
	if imageIndex > 0 {
		image = b.opts.ProtoformImage[imageIndex+1:]
	}

	return strings.Join([]string{registry, path, image}, "/")
}
