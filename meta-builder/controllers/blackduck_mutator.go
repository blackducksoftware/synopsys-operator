/*
 * Copyright (C) $year Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/meta-builder/api/v1"
	"github.com/blackducksoftware/synopsys-operator/meta-builder/controllers/controllers_utils"
)

func patchBlackduck(client client.Client, blackduck *synopsysv1.Blackduck, objects map[string]runtime.Object) map[string]runtime.Object {
	patcher := BlackduckPatcher{
		Client:    client,
		blackduck: blackduck,
		objects:   objects,
	}
	return patcher.patch()
}

type BlackduckPatcher struct {
	client.Client
	blackduck *synopsysv1.Blackduck
	objects   map[string]runtime.Object
}

func (p *BlackduckPatcher) patch() map[string]runtime.Object {
	// TODO JD: Patching this way is costly. Consider iterating over the objects only once
	// and apply the necessary changes
	p.patchNamespace()
	p.patchStorage()
	p.patchLiveness()
	p.patchEnvirons()
	p.patchWebserverCertificates()
	p.patchPostgresConfig()
	p.patchImages()
	p.patchAuthCert()
	p.patchProxyCert()
	p.patchExposeService()

	p.patchWithSize()
	p.patchReplicas()
	// TODO - Patch SEAL_KEY + BDBA
	return p.objects
}

func (p *BlackduckPatcher) patchExposeService() error {
	// TODO use contansts
	id := fmt.Sprintf("Service.%s-webserver-exposed", p.blackduck.Name)
	runtimeObject, ok := p.objects[id]
	if !ok {
		return nil
	}

	switch strings.ToUpper(p.blackduck.Spec.ExposeService) {
	case "LOADBALANCER":
		runtimeObject.(*v1.Service).Spec.Type = v1.ServiceTypeLoadBalancer
	case "NODEPORT":
		runtimeObject.(*v1.Service).Spec.Type = v1.ServiceTypeNodePort
	default:
		delete(p.objects, id)
	}

	return nil
}

func (p *BlackduckPatcher) patchAuthCert() error {
	if len(p.blackduck.Spec.AuthCustomCA) == 0 {
		for _, v := range p.objects {
			switch v.(type) {
			case *v1.ReplicationController:
				removeVolumeAndVolumeMountFromRC(v.(*v1.ReplicationController), fmt.Sprintf("%sauth-custom-ca", p.blackduck.Name))
			}
		}
	} else {
		secret, ok := p.objects[fmt.Sprintf("Secret.%s-auth-custom-ca", p.blackduck.Name)]
		if !ok {
			return nil
		}

		if secret.(*v1.Secret).Data == nil {
			secret.(*v1.Secret).Data = make(map[string][]byte)
		}

		secret.(*v1.Secret).Data["AUTH_CUSTOM_CA"] = []byte(p.blackduck.Spec.AuthCustomCA)
	}
	return nil
}

func (p *BlackduckPatcher) patchProxyCert() error {
	if len(p.blackduck.Spec.ProxyCertificate) == 0 {
		for _, v := range p.objects {
			switch v.(type) {
			case *v1.ReplicationController:
				removeVolumeAndVolumeMountFromRC(v.(*v1.ReplicationController), fmt.Sprintf("%s-proxy-certificate", p.blackduck.Name))
			}
		}
	} else {
		secret, ok := p.objects[fmt.Sprintf("Secret.%s-proxy-certificate", p.blackduck.Name)]
		if !ok {
			return nil
		}

		if secret.(*v1.Secret).Data == nil {
			secret.(*v1.Secret).Data = make(map[string][]byte)
		}

		secret.(*v1.Secret).Data["HUB_PROXY_CERT_FILE"] = []byte(p.blackduck.Spec.ProxyCertificate)
	}
	return nil
}

func (p *BlackduckPatcher) patchWithSize() error {
	var size synopsysv1.Size
	if len(p.blackduck.Spec.Size) > 0 {
		if err := p.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: p.blackduck.Namespace,
			Name:      strings.ToLower(p.blackduck.Spec.Size),
		}, &size); err != nil {

			if !apierrs.IsNotFound(err) {
				return err
			}
			if apierrs.IsNotFound(err) {
				return fmt.Errorf("blackduck instance [%s] is configured to use a Size [%s] that doesn't exist", p.blackduck.Namespace, p.blackduck.Spec.Size)
			}
		}
	}

	for _, v := range p.objects {
		switch v.(type) {
		case *v1.ReplicationController:
			componentName, ok := v.(*v1.ReplicationController).GetLabels()["component"]
			if !ok {
				return fmt.Errorf("component name is missing in %s", v.(*v1.ReplicationController).Name)
			}

			sizeConf, ok := size.Spec.PodResources[componentName]
			if !ok {
				return fmt.Errorf("blackduck instance [%s] is configured to use a Size [%s] but the size doesn't contain an entry for [%s]", p.blackduck.Namespace, p.blackduck.Spec.Size, v.(*v1.ReplicationController).Name)
			}
			v.(*v1.ReplicationController).Spec.Replicas = func(i int) *int32 { j := int32(i); return &j }(sizeConf.Replica)
			for containerIndex, container := range v.(*v1.ReplicationController).Spec.Template.Spec.Containers {
				containerConf, ok := sizeConf.ContainerLimit[container.Name]
				if !ok {
					return fmt.Errorf("blackduck instance [%s] is configured to use a Size [%s]. The size oesn't contain an entry for pod [%s] container [%s]", p.blackduck.Namespace, p.blackduck.Spec.Size, v.(*v1.ReplicationController).Name, container.Name)
				}
				resourceRequirements, err := controllers_utils.GenResourceRequirementsFromContainerSize(containerConf)
				if err != nil {
					return err
				}
				fmt.Println(resourceRequirements.Limits.Memory().String())
				v.(*v1.ReplicationController).Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements
			}

		}
	}
	return nil
}

func (p *BlackduckPatcher) patchReplicas() error {
	for _, v := range p.objects {
		switch v.(type) {
		case *v1.ReplicationController:
			switch p.blackduck.Spec.DesiredState {
			case "STOP":
				v.(*v1.ReplicationController).Spec.Replicas = func(i int32) *int32 { return &i }(0)
			case "DBMIGRATE":
			// TODO
			default:
				// TODO apply replica from flavor configuration
			}
		}
	}
	return nil
}

func (p *BlackduckPatcher) patchImages() error {
	if len(p.blackduck.Spec.RegistryConfiguration.Registry) > 0 || len(p.blackduck.Spec.ImageRegistries) > 0 {
		for _, v := range p.objects {
			switch v.(type) {
			case *v1.ReplicationController:
				for i := range v.(*v1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*v1.ReplicationController).Spec.Template.Spec.Containers[i].Image = controllers_utils.GenerateImageTag(v.(*v1.ReplicationController).Spec.Template.Spec.Containers[i].Image, p.blackduck.Spec.ImageRegistries, p.blackduck.Spec.RegistryConfiguration)
				}
			}
		}
	}
	return nil
}

func (p *BlackduckPatcher) patchPostgresConfig() error {
	cmConf, ok := p.objects[fmt.Sprintf("ConfigMap.%s-db-config", p.blackduck.Name)]
	if !ok {
		return nil
	}

	secretConf, ok := p.objects[fmt.Sprintf("Secret.%s-db-creds", p.blackduck.Name)]
	if !ok {
		return nil
	}

	if cmConf.(*v1.ConfigMap).Data == nil {
		cmConf.(*v1.ConfigMap).Data = make(map[string]string)
	}

	if secretConf.(*v1.Secret).Data == nil {
		secretConf.(*v1.Secret).Data = make(map[string][]byte)
	}

	if p.blackduck.Spec.ExternalPostgres != nil {
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_ADMIN"] = p.blackduck.Spec.ExternalPostgres.PostgresAdmin
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_ENABLE_SSL"] = strconv.FormatBool(p.blackduck.Spec.ExternalPostgres.PostgresSsl)
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_HOST"] = p.blackduck.Spec.ExternalPostgres.PostgresHost
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_PORT"] = strconv.Itoa(p.blackduck.Spec.ExternalPostgres.PostgresPort)
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_USER"] = p.blackduck.Spec.ExternalPostgres.PostgresUser

		secretConf.(*v1.Secret).Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"] = []byte(p.blackduck.Spec.ExternalPostgres.PostgresAdminPassword)
		secretConf.(*v1.Secret).Data["HUB_POSTGRES_USER_PASSWORD_FILE"] = []byte(p.blackduck.Spec.ExternalPostgres.PostgresUserPassword)

		// Delete the component required when deploying internal postgres
		delete(p.objects, fmt.Sprintf("PersistentVolumeClaim.%s-postgres", p.blackduck.Name))
		delete(p.objects, fmt.Sprintf("Job.%s-init-postgres", p.blackduck.Name))
		delete(p.objects, fmt.Sprintf("ConfigMap.%s-postgres-init-config", p.blackduck.Name))
		delete(p.objects, fmt.Sprintf("Service.%s-postgres", p.blackduck.Name))
		delete(p.objects, fmt.Sprintf("ReplicationController.%s-postgres", p.blackduck.Name))
	} else {
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_ADMIN"] = "blackduck"
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_ENABLE_SSL"] = "false"
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_HOST"] = fmt.Sprintf("%s-postgres", p.blackduck.Name)
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_PORT"] = "5432"
		cmConf.(*v1.ConfigMap).Data["HUB_POSTGRES_USER"] = "blackduck_user"

		secretConf.(*v1.Secret).Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"] = []byte(p.blackduck.Spec.AdminPassword)
		secretConf.(*v1.Secret).Data["HUB_POSTGRES_USER_PASSWORD_FILE"] = []byte(p.blackduck.Spec.UserPassword)
		secretConf.(*v1.Secret).Data["HUB_POSTGRES_POSTGRES_PASSWORD_FILE"] = []byte(p.blackduck.Spec.PostgresPassword)
	}

	return nil
}

func (p *BlackduckPatcher) patchWebserverCertificates() error {

	if len(p.blackduck.Spec.Certificate) > 0 && len(p.blackduck.Spec.CertificateKey) > 0 {
		runtimeObject, ok := p.objects[fmt.Sprintf("Secret.%s-webserver-certificate", p.blackduck.Name)]
		if !ok {
			return nil
		}
		runtimeObject.(*v1.Secret).Data["WEBSERVER_CUSTOM_CERT_FILE"] = []byte(p.blackduck.Spec.Certificate)
		runtimeObject.(*v1.Secret).Data["WEBSERVER_CUSTOM_KEY_FILE"] = []byte(p.blackduck.Spec.CertificateKey)

	}

	return nil
}

func (p *BlackduckPatcher) patchEnvirons() error {
	configMapRuntimeObject, ok := p.objects[fmt.Sprintf("ConfigMap.%s-config", p.blackduck.Name)]
	if !ok {
		return nil
	}
	configMap := configMapRuntimeObject.(*v1.ConfigMap)
	for _, e := range p.blackduck.Spec.Environs {
		vals := strings.Split(e, ":") // TODO - doesn't handle multiple colons
		if len(vals) != 2 {
			fmt.Printf("Could not split environ '%s' on ':'\n", e) // TODO change to log
			continue
		}
		environKey := strings.TrimSpace(vals[0])
		environVal := strings.TrimSpace(vals[1])
		configMap.Data[environKey] = environVal
	}
	return nil
}
func (p *BlackduckPatcher) patchNamespace() error {
	accessor := meta.NewAccessor()
	for _, runtimeObject := range p.objects {
		accessor.SetNamespace(runtimeObject, p.blackduck.Spec.Namespace)
	}
	return nil
}

func (p *BlackduckPatcher) patchLiveness() error {
	// Removes liveness probes if Spec.LivenessProbes is set to false
	for _, v := range p.objects {
		switch v.(type) {
		case *v1.ReplicationController:
			if !p.blackduck.Spec.LivenessProbes {
				for i := range v.(*v1.ReplicationController).Spec.Template.Spec.Containers {
					v.(*v1.ReplicationController).Spec.Template.Spec.Containers[i].LivenessProbe = nil
				}
			}
		}
	}
	return nil
}

func (p *BlackduckPatcher) patchStorage() error {
	for k, v := range p.objects {
		switch v.(type) {
		case *v1.PersistentVolumeClaim:
			if !p.blackduck.Spec.PersistentStorage {
				delete(p.objects, k)
			} else {
				if len(p.blackduck.Spec.PVCStorageClass) > 0 {
					v.(*v1.PersistentVolumeClaim).Spec.StorageClassName = &p.blackduck.Spec.PVCStorageClass
				}
				for _, pvc := range p.blackduck.Spec.PVC {
					if strings.EqualFold(pvc.Name, v.(*v1.PersistentVolumeClaim).Name) {
						v.(*v1.PersistentVolumeClaim).Spec.VolumeName = pvc.VolumeName
						v.(*v1.PersistentVolumeClaim).Spec.StorageClassName = &pvc.StorageClass
						if quantity, err := resource.ParseQuantity(pvc.Size); err == nil {
							v.(*v1.PersistentVolumeClaim).Spec.Resources.Requests[v1.ResourceStorage] = quantity
						}
					}
				}
			}
		case *v1.ReplicationController:
			if !p.blackduck.Spec.PersistentStorage {
				for i := range v.(*v1.ReplicationController).Spec.Template.Spec.Volumes {
					// If PersistentVolumeClaim then we change it to emptyDir
					if v.(*v1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource.PersistentVolumeClaim != nil {
						v.(*v1.ReplicationController).Spec.Template.Spec.Volumes[i].VolumeSource = v1.VolumeSource{
							EmptyDir: &v1.EmptyDirVolumeSource{
								Medium:    v1.StorageMediumDefault,
								SizeLimit: nil,
							},
						}
					}
				}
			}
		}

	}
	return nil
}

func removeVolumeAndVolumeMountFromRC(rc *v1.ReplicationController, volumeName string) *v1.ReplicationController {
	for volumeNb, volume := range rc.Spec.Template.Spec.Volumes {
		if volume.Secret != nil && strings.Compare(volume.Secret.SecretName, volumeName) == 0 {
			rc.Spec.Template.Spec.Volumes = append(rc.Spec.Template.Spec.Volumes[:volumeNb], rc.Spec.Template.Spec.Volumes[volumeNb+1:]...)
			for containerNb, container := range rc.Spec.Template.Spec.Containers {
				for volumeMountNb, volumeMount := range container.VolumeMounts {
					if strings.Compare(volumeMount.Name, volume.Name) == 0 {
						rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts = append(rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts[:volumeMountNb], rc.Spec.Template.Spec.Containers[containerNb].VolumeMounts[volumeMountNb+1:]...)
					}
				}
			}
		}
	}
	return rc
}
