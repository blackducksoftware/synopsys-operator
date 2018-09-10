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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/blackducksoftware/perceptor-protoform/cmd/protoform-bootstrapper/app/options"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/hubfederator"
	"github.com/blackducksoftware/perceptor-protoform/pkg/apps/perceptor"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"

	"k8s.io/client-go/tools/clientcmd"
)

// Bootstrapper defines how to bootstrap protoform
type Bootstrapper struct {
	*deployer.Deployer
	opts *options.BootstrapperOptions
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
func (b *Bootstrapper) setup() error {
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
	config := api.ProtoformConfig{
		ViperSecret:     "protoform",
		DefaultLogLevel: b.opts.LogLevel,
		Apps:            &api.ProtoformApps{},
	}

	// Perceptor Config
	if (b.opts.AnnotateImages != nil && *b.opts.AnnotateImages) ||
		(b.opts.AnnotatePods != nil && *b.opts.AnnotatePods) {
		config.Apps.PerceptorConfig = &perceptor.AppConfig{
			HubHost: b.opts.HubHost,
			HubUser: b.opts.HubUser,
			HubPort: b.opts.HubPort,
			HubClientTimeoutPerceptorMilliseconds: b.opts.HubClientTimeoutPerceptorMilliseconds,
			HubClientTimeoutScannerSeconds:        b.opts.HubClientTimeoutScannerSeconds,
			ConcurrentScanLimit:                   b.opts.ConcurrentScanLimit,
			Namespace:                             b.opts.PerceptorNamespace,
			ImagePerceiver:                        b.opts.AnnotateImages,
			PodPerceiver:                          b.opts.AnnotatePods,
			PerceptorSkyfire:                      b.opts.EnableSkyfire,
			Metrics:                               b.opts.EnableMetrics,
			InternalRegistries:                    b.opts.InternalRegistries,
			DefaultCPU:                            b.opts.DefaultCPU,
			DefaultMem:                            b.opts.DefaultMem,
			Registry:                              b.opts.DefaultRegistry,
			ImagePath:                             b.opts.DefaultImagePath,
			DefaultVersion:                        b.opts.DefaultImageVersion,
			PerceptorImageName:                    b.opts.PerceptorImage,
			ScannerImageName:                      b.opts.ScannerImage,
			PodPerceiverImageName:                 b.opts.PodPerceiverImage,
			ImagePerceiverImageName:               b.opts.ImagePerceiverImage,
			ImageFacadeImageName:                  b.opts.ImageFacadeImage,
			SkyfireImageName:                      b.opts.SkyfireImage,
			PerceptorImageVersion:                 b.opts.PerceptorImageVersion,
			ScannerImageVersion:                   b.opts.ScannerImageVersion,
			PerceiverImageVersion:                 b.opts.PerceiverImageVersion,
			ImageFacadeImageVersion:               b.opts.ImageFacadeImageVersion,
			SkyfireImageVersion:                   b.opts.SkyfireImageVersion,
			LogLevel:                              b.opts.LogLevel,
			SecretName:                            "protoform",
		}
	}

	// Alert Config
	if b.opts.AlertEnabled != nil && *b.opts.AlertEnabled {
		config.Apps.AlertConfig = &alert.AppConfig{
			HubHost:           b.opts.HubHost,
			HubUser:           b.opts.HubUser,
			HubPort:           b.opts.HubPort,
			Namespace:         b.opts.AlertNamespace,
			Registry:          b.opts.AlertRegistry,
			ImagePath:         b.opts.AlertImagePath,
			AlertImageName:    b.opts.AlertImageName,
			AlertImageVersion: b.opts.AlertImageVersion,
			CfsslImageName:    b.opts.CfsslImageName,
			CfsslImageVersion: b.opts.CfsslImageVersion,
		}
	}

	// Hub Federator Config
	if b.opts.HubFederatorEnabled != nil && *b.opts.HubFederatorEnabled {
		config.Apps.HubFederatorConfig = &hubfederator.AppConfig{
			Registry:        b.opts.HubFederatorRegistry,
			ImagePath:       b.opts.HubFederatorImagePath,
			ImageName:       b.opts.HubFederatorImageName,
			ImageVersion:    b.opts.HubFederatorImageVersion,
			Namespace:       b.opts.HubFederatorNamespace,
			RegistrationKey: b.opts.HubFederatorRegistrationKey,
			Port:            b.opts.HubFederatorPort,
			LogLevel:        b.opts.LogLevel,
		}
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal perceptor config: %v", err)
	}
	yamlOutput, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert json: %v", err)
	}

	cmConfig := horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		Name:       "protoform",
		Namespace:  b.opts.Namespace,
	}
	configMap := components.NewConfigMap(cmConfig)
	configMap.AddData(map[string]string{"protoform.yaml": string(yamlOutput)})
	b.AddConfigMap(configMap)

	b.createDeps()

	return nil
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
