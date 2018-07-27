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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"

	"k8s.io/client-go/tools/clientcmd"
)

// Bootstrapper defines how to bootstrap protoform
type Bootstrapper struct {
	*deployer.Deployer
	opts *options.BootstrapperOptions
	//	client *kubernetes.Clientset
}

// NewBootstrapper creats a new Bootstrapper object
func NewBootstrapper(opts *options.BootstrapperOptions) (*Bootstrapper, error) {
	config, err := clientcmd.BuildConfigFromFlags("", opts.ClusterConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	d, err := deployer.NewDeployer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %v", err)
	}

	b := Bootstrapper{d, opts}
	b.setup()
	return &b, nil
}

// Start begins the bootstrapping process
func (b *Bootstrapper) setup() {
	// Create the protoform container
	containerConfig := horizonapi.ContainerConfig{
		Name:       "protoform",
		Image:      fmt.Sprintf("%s:%s", b.generateImageName(), b.opts.ProtoformImageVersion),
		PullPolicy: horizonapi.PullAlways,
		Command:    []string{"./protoform"},
		Args:       []string{"/etc/protoform/protoform.yaml"},
	}
	container := components.NewContainer(containerConfig)
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: "3001", // TODO: Use config
		Protocol:      horizonapi.ProtocolTCP,
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "protoform",
		MountPath: "/etc/protoform",
	})
	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "PCP_HUBUSERPASSWORD",
		FromName:     "protoform",
		KeyOrVal:     "HubUserPassword",
		Type:         horizonapi.EnvFromSecret,
	})

	// Create volumes for the pod
	volConfig := horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "protoform",
		MapOrSecretName: "protoform",
	}
	vol := components.NewConfigMapVolume(volConfig)

	// Create the pod
	pc := horizonapi.PodConfig{
		APIVersion:     "v1",
		Name:           "protoform",
		Namespace:      b.opts.Namespace,
		RestartPolicy:  horizonapi.RestartPolicyNever,
		ServiceAccount: "protoform",
	}
	pod := components.NewPod(pc)
	pod.AddContainer(container)
	pod.AddVolume(vol)
	b.AddPod(pod)

	// Create the secret
	secretConfig := horizonapi.SecretConfig{
		APIVersion: "v1",
		Name:       "protoform",
		Namespace:  b.opts.Namespace,
		Type:       horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)
	secret.AddData(map[string][]byte{"HubUserPassword": []byte(b.opts.HubUserPassword)})
	b.AddSecret(secret)

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

	cmConfig := horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		Name:       "protoform",
		Namespace:  b.opts.Namespace,
	}
	configMap := components.NewConfigMap(cmConfig)
	configMap.AddData(map[string]string{"protoform.yaml": strings.Join(protoformConfig, "\n")})
	b.AddConfigMap(configMap)

	b.createDeps()
}

func (b *Bootstrapper) createDeps() {
	protoformServiceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "protoform",
		Namespace: b.opts.Namespace,
	})
	b.AddServiceAccount(protoformServiceAccount)

	protoformNameSpace := components.NewNamespace(horizonapi.NamespaceConfig{
		Name: b.opts.Namespace,
	})
	b.AddNamespace(protoformNameSpace)

	protoformSARoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "protoform",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	protoformSARoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "protoform",
		Namespace: b.opts.Namespace,
	})
	protoformSARoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})
	b.AddClusterRoleBinding(protoformSARoleBinding)
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
