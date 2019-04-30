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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// GetScannerReplicationController returns a Replication Controller for OpsSight's Scanner
func (p *SpecConfig) GetScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(p.opssight.Spec.ScannerPod.ReplicaCount)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.opssight.Spec.ScannerPod.Name,
		Namespace: p.opssight.Spec.Namespace,
	})

	rc.AddSelectors(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})
	rc.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})
	pod, err := p.getScannerPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// getScannerPod returns a Pod for OpsSight's Scanner
func (p *SpecConfig) getScannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.opssight.Spec.ScannerPod.Name,
		ServiceAccount: p.opssight.Spec.ScannerPod.ImageFacade.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})

	cont, err := p.getScannerContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(cont)

	facadecont, err := p.getImageFacadeContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(facadecont)

	vols, err := p.getScannerVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating scanner volumes")
	}

	newVols, err := p.getImageFacadeVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating image facade volumes")
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

// getScannerContainer returns a Cotnainer for the scanner of OpsSight's Scanner
func (p *SpecConfig) getScannerContainer() (*components.Container, error) {
	priv := false
	name := p.opssight.Spec.ScannerPod.Scanner.Name
	image := p.opssight.Spec.ScannerPod.Scanner.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "opssight-scanner")
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.Spec.ConfigMapName)},
		MinCPU:     p.opssight.Spec.ScannerCPU,
		MinMem:     p.opssight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(p.opssight.Spec.ScannerPod.Scanner.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
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

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.Spec.SecretName})

	return container, nil
}

// getImageFacadeContainer returns a Container for the Image Facade of OpsSight's Scanner
func (p *SpecConfig) getImageFacadeContainer() (*components.Container, error) {
	priv := false
	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		priv = true
	}

	name := p.opssight.Spec.ScannerPod.ImageFacade.Name
	image := p.opssight.Spec.ScannerPod.ImageFacade.Image
	if image == "" {
		image = GetImageTag(p.opssight.Spec.Version, "opssight-image-getter")
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.Spec.ConfigMapName)},
		MinCPU:     p.opssight.Spec.ScannerCPU,
		MinMem:     p.opssight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(p.opssight.Spec.ScannerPod.ImageFacade.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
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

	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name:      "dir-docker-socket",
			MountPath: "/var/run/docker.sock",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.Spec.SecretName})

	return container, nil
}

// getScannerVolumes returns a list of Volumes for OpsSight's Scanner
func (p *SpecConfig) getScannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.opssight.Spec.ScannerPod.Scanner.Name)}

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

// getImageFacadeVolumes returns a list of Volumes for OpsSight's Scanner
func (p *SpecConfig) getImageFacadeVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.opssight.Spec.ScannerPod.ImageFacade.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
			VolumeName: "dir-docker-socket",
			Path:       "/var/run/docker.sock",
		}))
	}

	return vols, nil
}

// GetScannerService returns a Service for the Scanner of OpsSight's Scanner
func (p *SpecConfig) GetScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.Spec.ScannerPod.Scanner.Name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.ScannerPod.Scanner.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.ScannerPod.Scanner.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.opssight.Spec.ScannerPod.Scanner.Name),
	})

	return service
}

// GetImageFacadeService returns a Service for the Image Facade of OpsSight's Scanner
func (p *SpecConfig) GetImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.Spec.ScannerPod.ImageFacade.Name,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	// TODO verify that this hits the *perceptor-scanner pod* !!!
	service.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.ScannerPod.ImageFacade.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.ScannerPod.ImageFacade.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.opssight.Spec.ScannerPod.ImageFacade.Name),
	})

	return service
}

// GetScannerServiceAccount returns a Service Account for OpsSight's Scanner
func (p *SpecConfig) GetScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      p.opssight.Spec.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.ImageFacade.ServiceAccount, "app": "opssight"})
	return serviceAccount
}

// GetScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *SpecConfig) GetScannerClusterRoleBinding() (*components.ClusterRoleBinding, error) {
	clusterRole := "synopsys-operator-admin"
	var err error
	if !p.config.DryRun {
		clusterRole, err = util.GetOperatorClusterRole(p.kubeClient)
		if err != nil {
			return nil, err
		}
	}

	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.opssight.Spec.ScannerPod.Name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.opssight.Spec.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.opssight.Spec.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole,
	})
	scannerCRB.AddLabels(map[string]string{"name": p.opssight.Spec.ScannerPod.Name, "app": "opssight"})

	return scannerCRB, nil
}
