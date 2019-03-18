/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.opssight. The ASF licenses this file
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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (p *SpecConfig) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(p.opssight.ScannerPod.ReplicaCount)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.opssight.ScannerPod.Name,
		Namespace: p.opssight.Namespace,
	})

	rc.AddLabelSelectors(map[string]string{"name": p.opssight.ScannerPod.Name, "app": "opssight"})

	pod, err := p.scannerPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.opssight.ScannerPod.Name,
		ServiceAccount: p.opssight.ScannerPod.ImageFacade.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"name": p.opssight.ScannerPod.Name, "app": "opssight"})

	cont, err := p.scannerContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(cont)

	facadecont, err := p.imageFacadeContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(facadecont)

	vols, err := p.scannerVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating scanner volumes")
	}

	newVols, err := p.imageFacadeVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating image facade volumes")
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) scannerContainer() (*components.Container, error) {
	priv := false
	name := p.opssight.ScannerPod.Scanner.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.opssight.ScannerPod.Scanner.Image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.ConfigMapName)},
		MinCPU:     p.opssight.ScannerCPU,
		MinMem:     p.opssight.ScannerMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.opssight.ScannerPod.Scanner.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.SecretName})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func (p *SpecConfig) imageFacadeContainer() (*components.Container, error) {
	priv := true
	name := p.opssight.ScannerPod.ImageFacade.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.opssight.ScannerPod.ImageFacade.Image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.ConfigMapName)},
		MinCPU:     p.opssight.ScannerCPU,
		MinMem:     p.opssight.ScannerMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.opssight.ScannerPod.ImageFacade.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !strings.EqualFold(p.opssight.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name:      "dir-docker-socket",
			MountPath: "/var/run/docker.sock",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	err = container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.SecretName})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func (p *SpecConfig) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.opssight.ScannerPod.Scanner.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

func (p *SpecConfig) imageFacadeVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.opssight.ScannerPod.ImageFacade.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	if !strings.EqualFold(p.opssight.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
			VolumeName: "dir-docker-socket",
			Path:       "/var/run/docker.sock",
		}))
	}

	return vols, nil
}

// ScannerService creates a service for perceptor scanner
func (p *SpecConfig) ScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.ScannerPod.Scanner.Name,
		Namespace: p.opssight.Namespace,
	})
	service.AddLabels(map[string]string{"name": p.opssight.ScannerPod.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.ScannerPod.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.ScannerPod.Scanner.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.ScannerPod.Scanner.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (p *SpecConfig) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.ScannerPod.ImageFacade.Name,
		Namespace: p.opssight.Namespace,
	})
	// TODO verify that this hits the *perceptor-scanner pod* !!!
	service.AddLabels(map[string]string{"name": p.opssight.ScannerPod.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.ScannerPod.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.ScannerPod.ImageFacade.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.ScannerPod.ImageFacade.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (p *SpecConfig) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      p.opssight.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.opssight.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"name": p.opssight.ScannerPod.ImageFacade.ServiceAccount, "app": "opssight"})
	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *SpecConfig) ScannerClusterRoleBinding() *components.ClusterRoleBinding {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.opssight.ScannerPod.Name, // TODO is this right?  or should it be .ImageFacade.Name ?
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.opssight.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.opssight.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})
	scannerCRB.AddLabels(map[string]string{"name": p.opssight.ScannerPod.Name, "app": "opssight"})

	return scannerCRB
}
