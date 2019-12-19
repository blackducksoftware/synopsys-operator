/*
Copyright (C) 2019 Synopsys, Inc.

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

package alert

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	appsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// getAlertDeployment returns a new deployment for an Alert
func (a *SpecConfig) getAlertDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deploymentConfig := horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(a.alert.Name, util.AlertName, "alert"),
		Namespace: a.alert.Spec.Namespace,
	}
	labels := map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"}

	pod, err := a.getAlertPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Pod: %s", err)
	}

	return util.CreateDeployment(&deploymentConfig, pod, pod.GetLabels(), labels), nil
}

// getAlertPod returns a new Pod for an Alert
func (a *SpecConfig) getAlertPod() (*components.Pod, error) {
	podConfig := &util.PodConfig{
		Name:           util.GetResourceName(a.alert.Name, util.AlertName, "alert"),
		Labels:         map[string]string{"app": util.AlertName, "name": a.alert.Name, "component": "alert"},
		ServiceAccount: util.GetResourceName(a.alert.Name, util.AlertName, "service-account"),
	}

	appsutil.ConfigurePodConfigSecurityContext(podConfig, a.alert.Spec.SecurityContexts, "alert", 1000, a.isOpenshift)

	pod, err := util.CreatePod(podConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Pod: %+v", err)
	}

	container, err := a.getAlertContainer()
	if err != nil {
		return nil, err
	}
	pod.AddContainer(container)

	if a.alert.Spec.PersistentStorage {
		log.Debugf("Adding a PersistentVolumeClaim Volume to the Alert's Pod")
		pod.AddVolume(a.getAlertPVCVolume())
	} else {
		log.Debugf("Adding an EmptyDir Volume to the Alert's Pod")
		vol, err := a.getAlertEmptyDirVolume()
		if err != nil {
			return nil, fmt.Errorf("failed to Add Volume to Alert Pod: %s", err)
		}
		pod.AddVolume(vol)
	}

	// add cert volume
	if len(a.alert.Spec.JavaKeyStore) > 0 || (len(a.alert.Spec.Certificate) > 0 && len(a.alert.Spec.CertificateKey) > 0) {
		certVolume, err := a.getAlertCustomCertSecretVolume()
		if err != nil {
			return nil, fmt.Errorf("failed to Add Cert Volume to Alert Pod: %s", err)
		}
		pod.AddVolume(certVolume)
	}

	if a.alert.Spec.RegistryConfiguration != nil && len(a.alert.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(a.alert.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

// getAlertContainer returns a new Container for an Alert
func (a *SpecConfig) getAlertContainer() (*components.Container, error) {
	image := appsutil.GenerateImageTag(GetImageTag(a.alert.Spec.Version, "blackduck-alert"), a.alert.Spec.ImageRegistries, a.alert.Spec.RegistryConfiguration)
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       "alert",
		Image:      image,
		PullPolicy: horizonapi.PullAlways,
		MinMem:     a.alert.Spec.AlertMemory,
		MaxMem:     a.alert.Spec.AlertMemory,
	})

	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: *a.alert.Spec.Port,
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-alert",
		MountPath: "/opt/blackduck/alert/alert-config",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(a.alert.Spec.JavaKeyStore) > 0 || (len(a.alert.Spec.Certificate) > 0 && len(a.alert.Spec.CertificateKey) > 0) {
		container.AddEnv(horizonapi.EnvConfig{
			Type:         horizonapi.EnvFromName,
			NameOrPrefix: "RUN_SECRETS_DIR",
			KeyOrVal:     "/tmp/secrets",
		})
	}

	if len(a.alert.Spec.Certificate) > 0 && len(a.alert.Spec.CertificateKey) > 0 {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_CERT_FILE", SubPath: "WEBSERVER_CUSTOM_CERT_FILE",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}

		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name: "certificate", MountPath: "/tmp/secrets/WEBSERVER_CUSTOM_KEY_FILE", SubPath: "WEBSERVER_CUSTOM_KEY_FILE",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	if len(a.alert.Spec.JavaKeyStore) > 0 {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name: "certificate", MountPath: "/tmp/secrets/cacerts", SubPath: "cacerts",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: util.GetResourceName(a.alert.Name, util.AlertName, "blackduck-config"),
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromSecret,
		FromName: util.GetResourceName(a.alert.Name, util.AlertName, "secret"),
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Type:    horizonapi.ActionTypeCommand,
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/alert/api/about"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 5,
	})

	return container, nil
}

// getAlertEmptyDirVolume returns a new EmptyDirVolume for an Alert
func (a *SpecConfig) getAlertEmptyDirVolume() (*components.Volume, error) {
	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "dir-alert",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert EmptyDir: %s", err)
	}

	return vol, err
}

// getAlertPVCVolume returns a new PVCVolume for an Alert
func (a *SpecConfig) getAlertPVCVolume() *components.Volume {
	name := util.GetResourceName(a.alert.Name, util.AlertName, a.alert.Spec.PVCName)
	if a.alert.Annotations["synopsys.com/created.by"] == "pre-2019.6.0" {
		name = a.alert.Spec.PVCName
	}

	vol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		VolumeName: "dir-alert",
		PVCName:    name,
		ReadOnly:   false,
	})

	return vol
}

func (a *SpecConfig) getAlertCustomCertSecretVolume() (*components.Volume, error) {
	// Custom cert
	customCertVolume, err := util.CreateSecretVolume("certificate", util.GetResourceName(a.alert.Name, util.AlertName, "certificate"), 0444)
	if err != nil {
		return nil, fmt.Errorf("failed to create Alert Custom certificate secret volume: %s", err)
	}
	return customCertVolume, nil
}
